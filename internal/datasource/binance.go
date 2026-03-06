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
	Register("binance", func(store *price.Store) DataSource {
		return newBinance(store)
	})
	RegisterDefaultURLs("binance", bnWSURLs[0], bnTickerPriceURL)
}

const (
	bnTickerPriceURL = "https://api.binance.com/api/v3/ticker/price"
	bnMaxReconn      = 60 * time.Second
	bnInitReconn     = 1 * time.Second
)

var bnWSURLs = []string{
	"wss://stream.binance.com:9443/stream?streams=",
	"wss://stream.binance.com:443/stream?streams=",
	"wss://data-stream.binance.com/stream?streams=",
}

type binanceSource struct {
	store       *price.Store
	conn        *websocket.Conn
	mu          sync.Mutex
	pairs       map[string]bool
	wsURLIdx    int
	stopCh      chan struct{}
	pingDone    chan struct{}
	reconnDelay time.Duration
	running     bool

	coinsMu  sync.RWMutex
	coins    []CoinInfo
	cacheDir string
}

func newBinance(store *price.Store) *binanceSource {
	return &binanceSource{
		store:       store,
		pairs:       make(map[string]bool),
		stopCh:      make(chan struct{}),
		reconnDelay: bnInitReconn,
	}
}

func (b *binanceSource) Name() string { return "binance" }

func (b *binanceSource) Start() {
	b.mu.Lock()
	if b.running {
		b.mu.Unlock()
		return
	}
	b.running = true
	b.stopCh = make(chan struct{})
	b.mu.Unlock()
	go b.connectLoop()
}

func (b *binanceSource) Stop() {
	b.mu.Lock()
	defer b.mu.Unlock()
	if !b.running {
		return
	}
	b.running = false
	close(b.stopCh)
	if b.conn != nil {
		b.conn.Close()
		b.conn = nil
	}
}

func (b *binanceSource) SetPairs(pairs []string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.pairs = make(map[string]bool)
	for _, p := range pairs {
		b.pairs[strings.ToLower(p)] = true
	}
	if b.running && b.conn != nil {
		b.conn.Close()
		b.conn = nil
	}
}

func (b *binanceSource) connectLoop() {
	failCount := 0
	for {
		select {
		case <-b.stopCh:
			return
		default:
		}
		if err := b.connect(); err != nil {
			log.Printf("[Binance] connection error (endpoint %d): %v", b.wsURLIdx%len(bnWSURLs), err)
			failCount++
			if failCount >= 2 {
				b.wsURLIdx++
				failCount = 0
				log.Printf("[Binance] switching to endpoint %d: %s...", b.wsURLIdx%len(bnWSURLs), bnWSURLs[b.wsURLIdx%len(bnWSURLs)][:40])
			}
			b.backoff()
			continue
		}
		b.reconnDelay = bnInitReconn
		failCount = 0
		b.readLoop()
	}
}

func (b *binanceSource) buildStreamURL() string {
	b.mu.Lock()
	pairs := make([]string, 0, len(b.pairs))
	for p := range b.pairs {
		pairs = append(pairs, p)
	}
	b.mu.Unlock()
	if len(pairs) == 0 {
		return ""
	}
	streams := make([]string, len(pairs))
	for i, p := range pairs {
		streams[i] = p + "@miniTicker"
	}
	suffix := strings.Join(streams, "/")
	baseURL := bnWSURLs[b.wsURLIdx%len(bnWSURLs)]
	return baseURL + suffix
}

func (b *binanceSource) connect() error {
	url := b.buildStreamURL()
	if url == "" {
		time.Sleep(2 * time.Second)
		return fmt.Errorf("no pairs configured")
	}
	log.Printf("[Binance] connecting to WebSocket...")
	dialer := websocket.Dialer{HandshakeTimeout: 15 * time.Second}
	conn, resp, err := dialer.Dial(url, nil)
	if err != nil {
		if resp != nil {
			return fmt.Errorf("dial (status %d): %w", resp.StatusCode, err)
		}
		return fmt.Errorf("dial: %w", err)
	}
	conn.SetPongHandler(func(string) error { return nil })
	b.mu.Lock()
	if b.pingDone != nil {
		close(b.pingDone)
	}
	b.pingDone = make(chan struct{})
	b.conn = conn
	pingDone := b.pingDone
	b.mu.Unlock()
	go b.pingLoop(conn, pingDone)
	log.Printf("[Binance] connected successfully")
	return nil
}

func (b *binanceSource) pingLoop(conn *websocket.Conn, done chan struct{}) {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-b.stopCh:
			return
		case <-done:
			return
		case <-ticker.C:
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Printf("[Binance] ping error: %v", err)
				return
			}
		}
	}
}

type bnCombinedMsg struct {
	Stream string          `json:"stream"`
	Data   json.RawMessage `json:"data"`
}

type bnMiniTicker struct {
	EventType string `json:"e"`
	EventTime int64  `json:"E"`
	Symbol    string `json:"s"`
	Close     string `json:"c"`
	Open      string `json:"o"`
}

func (b *binanceSource) readLoop() {
	for {
		select {
		case <-b.stopCh:
			return
		default:
		}
		b.mu.Lock()
		conn := b.conn
		b.mu.Unlock()
		if conn == nil {
			return
		}
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("[Binance] read error: %v", err)
			return
		}
		var combined bnCombinedMsg
		if err := json.Unmarshal(message, &combined); err != nil {
			continue
		}
		var ticker bnMiniTicker
		if err := json.Unmarshal(combined.Data, &ticker); err != nil {
			continue
		}
		closePrice, err := strconv.ParseFloat(ticker.Close, 64)
		if err != nil {
			continue
		}
		openPrice, _ := strconv.ParseFloat(ticker.Open, 64)
		b.store.Update(strings.ToUpper(ticker.Symbol), closePrice, openPrice)
	}
}

func (b *binanceSource) backoff() {
	select {
	case <-b.stopCh:
		return
	case <-time.After(b.reconnDelay):
	}
	b.reconnDelay = time.Duration(math.Min(float64(b.reconnDelay)*2, float64(bnMaxReconn)))
}

// --- Coin list ---

type bnTickerPrice struct {
	Symbol string `json:"symbol"`
	Price  string `json:"price"`
}

type bnCoinCache struct {
	Coins     []CoinInfo `json:"coins"`
	UpdatedAt time.Time  `json:"updated_at"`
}

func (b *binanceSource) FetchCoinList(cacheDir string) ([]CoinInfo, error) {
	b.cacheDir = cacheDir
	cachePath := filepath.Join(cacheDir, "coins_binance.json")

	if data, err := os.ReadFile(cachePath); err == nil {
		var cache bnCoinCache
		if json.Unmarshal(data, &cache) == nil && time.Since(cache.UpdatedAt) < 24*time.Hour {
			b.coinsMu.Lock()
			b.coins = cache.Coins
			b.coinsMu.Unlock()
			return cache.Coins, nil
		}
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(bnTickerPriceURL)
	if err != nil {
		return nil, fmt.Errorf("fetch ticker prices: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}
	var tickers []bnTickerPrice
	if err := json.Unmarshal(body, &tickers); err != nil {
		return nil, fmt.Errorf("parse tickers: %w", err)
	}

	seen := make(map[string]bool)
	var coins []CoinInfo
	for _, t := range tickers {
		if !strings.HasSuffix(t.Symbol, "USDT") {
			continue
		}
		base := strings.TrimSuffix(t.Symbol, "USDT")
		if base == "" || seen[base] || IsStablecoin(base) || IsLeveraged(base) {
			continue
		}
		seen[base] = true
		coins = append(coins, CoinInfo{
			Symbol:    base,
			BaseAsset: base,
			Name:      HumanName(base),
			Pair:      t.Symbol,
		})
	}

	sort.Slice(coins, func(i, j int) bool {
		pi, pj := CoinPriority(coins[i].Symbol), CoinPriority(coins[j].Symbol)
		if pi != pj {
			return pi < pj
		}
		return coins[i].Symbol < coins[j].Symbol
	})

	b.coinsMu.Lock()
	b.coins = coins
	b.coinsMu.Unlock()

	cache := bnCoinCache{Coins: coins, UpdatedAt: time.Now()}
	if data, err := json.MarshalIndent(cache, "", "  "); err == nil {
		_ = os.WriteFile(cachePath, data, 0644)
	}
	return coins, nil
}

func (b *binanceSource) FindCoinByPair(pair string) (CoinInfo, bool) {
	b.coinsMu.RLock()
	defer b.coinsMu.RUnlock()
	for _, c := range b.coins {
		if c.Pair == pair {
			return c, true
		}
	}
	return CoinInfo{}, false
}

func (b *binanceSource) GetAllCoins() []CoinInfo {
	b.coinsMu.RLock()
	defer b.coinsMu.RUnlock()
	result := make([]CoinInfo, len(b.coins))
	copy(result, b.coins)
	return result
}
