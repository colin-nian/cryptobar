package datasource

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"cryptobar/internal/price"

	"github.com/gorilla/websocket"
)

func init() {
	Register("gateio", func(store *price.Store) DataSource {
		return newGateIO(store)
	})
	RegisterDefaultURLs("gateio", gateWSURL, gateTickerURL)
}

const (
	gateWSURL      = "wss://api.gateio.ws/ws/v4/"
	gateTickerURL  = "https://api.gateio.ws/api/v4/spot/tickers"
	gateMaxReconn  = 60 * time.Second
	gateInitReconn = 1 * time.Second
)

type gateIOSource struct {
	store       *price.Store
	conn        *websocket.Conn
	mu          sync.Mutex
	pairs       map[string]bool // "BTC_USDT" format
	stopCh      chan struct{}
	reconnDelay time.Duration
	running     bool

	coinsMu  sync.RWMutex
	coins    []CoinInfo
	cacheDir string
}

func newGateIO(store *price.Store) *gateIOSource {
	return &gateIOSource{
		store:       store,
		pairs:       make(map[string]bool),
		stopCh:      make(chan struct{}),
		reconnDelay: gateInitReconn,
	}
}

func (g *gateIOSource) Name() string { return "gateio" }

func (g *gateIOSource) Start() {
	g.mu.Lock()
	if g.running {
		g.mu.Unlock()
		return
	}
	g.running = true
	g.stopCh = make(chan struct{})
	g.mu.Unlock()
	go g.connectLoop()
}

func (g *gateIOSource) Stop() {
	g.mu.Lock()
	defer g.mu.Unlock()
	if !g.running {
		return
	}
	g.running = false
	close(g.stopCh)
	if g.conn != nil {
		g.conn.Close()
		g.conn = nil
	}
}

// toGatePair converts "BTCUSDT" to "BTC_USDT"
func toGatePair(pair string) string {
	pair = strings.ToUpper(pair)
	if strings.Contains(pair, "_") {
		return pair
	}
	if strings.HasSuffix(pair, "USDT") {
		base := strings.TrimSuffix(pair, "USDT")
		return base + "_USDT"
	}
	return pair
}

// fromGatePair converts "BTC_USDT" to "BTCUSDT"
func fromGatePair(pair string) string {
	return strings.ReplaceAll(pair, "_", "")
}

func (g *gateIOSource) SetPairs(pairs []string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.pairs = make(map[string]bool)
	for _, p := range pairs {
		g.pairs[toGatePair(p)] = true
	}
	if g.running && g.conn != nil {
		g.conn.Close()
		g.conn = nil
	}
}

func (g *gateIOSource) connectLoop() {
	for {
		select {
		case <-g.stopCh:
			return
		default:
		}
		if err := g.connect(); err != nil {
			log.Printf("[Gate.io] connection error: %v", err)
			g.backoff()
			continue
		}
		g.reconnDelay = gateInitReconn
		g.readLoop()
	}
}

func (g *gateIOSource) connect() error {
	g.mu.Lock()
	pairs := make([]string, 0, len(g.pairs))
	for p := range g.pairs {
		pairs = append(pairs, p)
	}
	g.mu.Unlock()

	if len(pairs) == 0 {
		time.Sleep(2 * time.Second)
		return fmt.Errorf("no pairs configured")
	}

	log.Printf("[Gate.io] connecting to WebSocket...")
	dialer := websocket.Dialer{HandshakeTimeout: 15 * time.Second}
	conn, _, err := dialer.Dial(gateWSURL, nil)
	if err != nil {
		return fmt.Errorf("dial: %w", err)
	}

	g.mu.Lock()
	g.conn = conn
	g.mu.Unlock()

	sub := gateWSMsg{
		Time:    time.Now().Unix(),
		Channel: "spot.tickers",
		Event:   "subscribe",
		Payload: pairs,
	}
	if err := conn.WriteJSON(sub); err != nil {
		return fmt.Errorf("subscribe: %w", err)
	}

	go g.pingLoop()
	log.Printf("[Gate.io] connected and subscribed to %d pairs", len(pairs))
	return nil
}

type gateWSMsg struct {
	Time    int64    `json:"time"`
	Channel string   `json:"channel"`
	Event   string   `json:"event"`
	Payload []string `json:"payload"`
}

type gateTickerUpdate struct {
	Time    int64  `json:"time"`
	Channel string `json:"channel"`
	Event   string `json:"event"`
	Result  struct {
		CurrencyPair  string `json:"currency_pair"`
		Last          string `json:"last"`
		ChangePercent string `json:"change_percentage"`
		High24h       string `json:"high_24h"`
		Low24h        string `json:"low_24h"`
	} `json:"result"`
}

func (g *gateIOSource) pingLoop() {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-g.stopCh:
			return
		case <-ticker.C:
			g.mu.Lock()
			conn := g.conn
			g.mu.Unlock()
			if conn != nil {
				ping := gateWSMsg{
					Time:    time.Now().Unix(),
					Channel: "spot.ping",
				}
				if err := conn.WriteJSON(ping); err != nil {
					log.Printf("[Gate.io] ping error: %v", err)
					return
				}
			}
		}
	}
}

func (g *gateIOSource) readLoop() {
	for {
		select {
		case <-g.stopCh:
			return
		default:
		}

		g.mu.Lock()
		conn := g.conn
		g.mu.Unlock()
		if conn == nil {
			return
		}

		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("[Gate.io] read error: %v", err)
			return
		}

		var update gateTickerUpdate
		if err := json.Unmarshal(message, &update); err != nil {
			continue
		}
		if update.Channel != "spot.tickers" || update.Event != "update" {
			continue
		}
		if update.Result.CurrencyPair == "" || update.Result.Last == "" {
			continue
		}

		lastPrice, err := strconv.ParseFloat(update.Result.Last, 64)
		if err != nil {
			continue
		}

		// Calculate open price from change percentage
		var openPrice float64
		if pct, err := strconv.ParseFloat(update.Result.ChangePercent, 64); err == nil && pct != 0 {
			openPrice = lastPrice / (1 + pct/100)
		}

		storePair := fromGatePair(update.Result.CurrencyPair)
		g.store.Update(storePair, lastPrice, openPrice)
	}
}

func (g *gateIOSource) backoff() {
	select {
	case <-g.stopCh:
		return
	case <-time.After(g.reconnDelay):
	}
	g.reconnDelay = time.Duration(math.Min(float64(g.reconnDelay)*2, float64(gateMaxReconn)))
}

// --- Coin list ---

type gateTickerItem struct {
	CurrencyPair string `json:"currency_pair"`
	Last         string `json:"last"`
	ChangePct    string `json:"change_percentage"`
	EtfLeverage  string `json:"etf_leverage,omitempty"`
}

type gateCoinCache struct {
	Coins     []CoinInfo `json:"coins"`
	UpdatedAt time.Time  `json:"updated_at"`
}

func (g *gateIOSource) FetchCoinList(cacheDir string) ([]CoinInfo, error) {
	g.cacheDir = cacheDir
	cachePath := filepath.Join(cacheDir, "coins_gateio.json")

	if data, err := os.ReadFile(cachePath); err == nil {
		var cache gateCoinCache
		if json.Unmarshal(data, &cache) == nil && time.Since(cache.UpdatedAt) < 24*time.Hour {
			g.coinsMu.Lock()
			g.coins = cache.Coins
			g.coinsMu.Unlock()
			return cache.Coins, nil
		}
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(gateTickerURL)
	if err != nil {
		return nil, fmt.Errorf("fetch Gate.io tickers: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	var tickers []gateTickerItem
	if err := json.Unmarshal(body, &tickers); err != nil {
		return nil, fmt.Errorf("parse tickers: %w", err)
	}

	seen := make(map[string]bool)
	var coins []CoinInfo
	for _, t := range tickers {
		if !strings.HasSuffix(t.CurrencyPair, "_USDT") {
			continue
		}
		if t.EtfLeverage != "" {
			continue
		}
		base := strings.TrimSuffix(t.CurrencyPair, "_USDT")
		if base == "" || seen[base] || IsStablecoin(base) || IsLeveraged(base) {
			continue
		}
		seen[base] = true
		// Store pair in unified BTCUSDT format for config compatibility
		unifiedPair := base + "USDT"
		coins = append(coins, CoinInfo{
			Symbol:    base,
			BaseAsset: base,
			Name:      HumanName(base),
			Pair:      unifiedPair,
		})
	}

	sort.Slice(coins, func(i, j int) bool {
		pi, pj := CoinPriority(coins[i].Symbol), CoinPriority(coins[j].Symbol)
		if pi != pj {
			return pi < pj
		}
		return coins[i].Symbol < coins[j].Symbol
	})

	g.coinsMu.Lock()
	g.coins = coins
	g.coinsMu.Unlock()

	cache := gateCoinCache{Coins: coins, UpdatedAt: time.Now()}
	if data, err := json.MarshalIndent(cache, "", "  "); err == nil {
		_ = os.WriteFile(cachePath, data, 0644)
	}
	return coins, nil
}

func (g *gateIOSource) FindCoinByPair(pair string) (CoinInfo, bool) {
	g.coinsMu.RLock()
	defer g.coinsMu.RUnlock()
	for _, c := range g.coins {
		if c.Pair == pair {
			return c, true
		}
	}
	return CoinInfo{}, false
}

func (g *gateIOSource) GetAllCoins() []CoinInfo {
	g.coinsMu.RLock()
	defer g.coinsMu.RUnlock()
	result := make([]CoinInfo, len(g.coins))
	copy(result, g.coins)
	return result
}
