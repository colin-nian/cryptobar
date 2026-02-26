package datasource

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"cryptobar/internal/price"

	"github.com/gorilla/websocket"
)

func init() {
	Register("htx", func(store *price.Store) DataSource {
		return newHTX(store)
	})
	RegisterDefaultURLs("htx", htxWSURL, htxTickerURL)
}

const (
	htxWSURL      = "wss://api.huobi.pro/ws"
	htxTickerURL  = "https://api.huobi.pro/market/tickers"
	htxMaxReconn  = 60 * time.Second
	htxInitReconn = 1 * time.Second
)

type htxSource struct {
	store       *price.Store
	conn        *websocket.Conn
	mu          sync.Mutex
	pairs       map[string]bool
	stopCh      chan struct{}
	reconnDelay time.Duration
	running     bool

	coinsMu  sync.RWMutex
	coins    []CoinInfo
	cacheDir string
}

func newHTX(store *price.Store) *htxSource {
	return &htxSource{
		store:       store,
		pairs:       make(map[string]bool),
		stopCh:      make(chan struct{}),
		reconnDelay: htxInitReconn,
	}
}

func (h *htxSource) Name() string { return "htx" }

func (h *htxSource) Start() {
	h.mu.Lock()
	if h.running {
		h.mu.Unlock()
		return
	}
	h.running = true
	h.stopCh = make(chan struct{})
	h.mu.Unlock()
	go h.connectLoop()
}

func (h *htxSource) Stop() {
	h.mu.Lock()
	defer h.mu.Unlock()
	if !h.running {
		return
	}
	h.running = false
	close(h.stopCh)
	if h.conn != nil {
		h.conn.Close()
		h.conn = nil
	}
}

func (h *htxSource) SetPairs(pairs []string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.pairs = make(map[string]bool)
	for _, p := range pairs {
		sym := strings.ToLower(strings.ReplaceAll(p, "_", ""))
		h.pairs[sym] = true
	}
	if h.running && h.conn != nil {
		h.conn.Close()
		h.conn = nil
	}
}

func (h *htxSource) connectLoop() {
	for {
		select {
		case <-h.stopCh:
			return
		default:
		}
		if err := h.connect(); err != nil {
			log.Printf("[HTX] connection error: %v", err)
			h.backoff()
			continue
		}
		h.reconnDelay = htxInitReconn
		h.readLoop()
	}
}

func (h *htxSource) connect() error {
	h.mu.Lock()
	pairs := make([]string, 0, len(h.pairs))
	for p := range h.pairs {
		pairs = append(pairs, p)
	}
	h.mu.Unlock()

	if len(pairs) == 0 {
		time.Sleep(2 * time.Second)
		return fmt.Errorf("no pairs configured")
	}

	log.Printf("[HTX] connecting to WebSocket...")
	dialer := websocket.Dialer{HandshakeTimeout: 15 * time.Second}
	conn, _, err := dialer.Dial(htxWSURL, nil)
	if err != nil {
		return fmt.Errorf("dial: %w", err)
	}

	h.mu.Lock()
	h.conn = conn
	h.mu.Unlock()

	for _, p := range pairs {
		sub := map[string]string{
			"sub": fmt.Sprintf("market.%s.ticker", p),
			"id":  p,
		}
		if err := conn.WriteJSON(sub); err != nil {
			return fmt.Errorf("subscribe %s: %w", p, err)
		}
	}

	log.Printf("[HTX] connected and subscribed to %d pairs", len(pairs))
	return nil
}

func (h *htxSource) readLoop() {
	for {
		select {
		case <-h.stopCh:
			return
		default:
		}

		h.mu.Lock()
		conn := h.conn
		h.mu.Unlock()
		if conn == nil {
			return
		}

		_, compressed, err := conn.ReadMessage()
		if err != nil {
			log.Printf("[HTX] read error: %v", err)
			return
		}

		message, err := gzipDecompress(compressed)
		if err != nil {
			continue
		}

		var ping struct {
			Ping int64 `json:"ping"`
		}
		if json.Unmarshal(message, &ping) == nil && ping.Ping > 0 {
			pong := map[string]int64{"pong": ping.Ping}
			h.mu.Lock()
			if h.conn != nil {
				_ = h.conn.WriteJSON(pong)
			}
			h.mu.Unlock()
			continue
		}

		var tick htxTickerMsg
		if err := json.Unmarshal(message, &tick); err != nil {
			continue
		}
		if tick.Ch == "" || tick.Tick.Close == 0 {
			continue
		}

		// ch format: "market.btcusdt.ticker"
		parts := strings.Split(tick.Ch, ".")
		if len(parts) < 2 {
			continue
		}
		symbol := strings.ToUpper(parts[1])
		if strings.HasSuffix(symbol, "USDT") {
			// keep as-is for pair matching
		}
		h.store.Update(symbol, tick.Tick.Close, tick.Tick.Open)
	}
}

type htxTickerMsg struct {
	Ch   string `json:"ch"`
	Tick struct {
		Open  float64 `json:"open"`
		Close float64 `json:"close"`
		High  float64 `json:"high"`
		Low   float64 `json:"low"`
	} `json:"tick"`
}

func gzipDecompress(data []byte) ([]byte, error) {
	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	return io.ReadAll(reader)
}

func (h *htxSource) backoff() {
	select {
	case <-h.stopCh:
		return
	case <-time.After(h.reconnDelay):
	}
	h.reconnDelay = time.Duration(math.Min(float64(h.reconnDelay)*2, float64(htxMaxReconn)))
}

// --- Coin list ---

type htxTickerResp struct {
	Status string `json:"status"`
	Data   []struct {
		Symbol string  `json:"symbol"`
		Open   float64 `json:"open"`
		Close  float64 `json:"close"`
	} `json:"data"`
}

type htxCoinCache struct {
	Coins     []CoinInfo `json:"coins"`
	UpdatedAt time.Time  `json:"updated_at"`
}

func (h *htxSource) FetchCoinList(cacheDir string) ([]CoinInfo, error) {
	h.cacheDir = cacheDir
	cachePath := filepath.Join(cacheDir, "coins_htx.json")

	if data, err := os.ReadFile(cachePath); err == nil {
		var cache htxCoinCache
		if json.Unmarshal(data, &cache) == nil && time.Since(cache.UpdatedAt) < 24*time.Hour {
			h.coinsMu.Lock()
			h.coins = cache.Coins
			h.coinsMu.Unlock()
			return cache.Coins, nil
		}
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(htxTickerURL)
	if err != nil {
		return nil, fmt.Errorf("fetch HTX tickers: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	var tickerResp htxTickerResp
	if err := json.Unmarshal(body, &tickerResp); err != nil {
		return nil, fmt.Errorf("parse tickers: %w", err)
	}

	seen := make(map[string]bool)
	var coins []CoinInfo
	for _, t := range tickerResp.Data {
		sym := t.Symbol
		if !strings.HasSuffix(sym, "usdt") {
			continue
		}
		base := strings.ToUpper(strings.TrimSuffix(sym, "usdt"))
		if base == "" || seen[base] || IsStablecoin(base) || IsLeveraged(base) {
			continue
		}
		seen[base] = true
		pair := strings.ToUpper(sym)
		coins = append(coins, CoinInfo{
			Symbol:    base,
			BaseAsset: base,
			Name:      HumanName(base),
			Pair:      pair,
		})
	}

	sort.Slice(coins, func(i, j int) bool {
		pi, pj := CoinPriority(coins[i].Symbol), CoinPriority(coins[j].Symbol)
		if pi != pj {
			return pi < pj
		}
		return coins[i].Symbol < coins[j].Symbol
	})

	h.coinsMu.Lock()
	h.coins = coins
	h.coinsMu.Unlock()

	cache := htxCoinCache{Coins: coins, UpdatedAt: time.Now()}
	if data, err := json.MarshalIndent(cache, "", "  "); err == nil {
		_ = os.WriteFile(cachePath, data, 0644)
	}
	return coins, nil
}

func (h *htxSource) FindCoinByPair(pair string) (CoinInfo, bool) {
	h.coinsMu.RLock()
	defer h.coinsMu.RUnlock()
	for _, c := range h.coins {
		if c.Pair == pair {
			return c, true
		}
	}
	return CoinInfo{}, false
}

func (h *htxSource) GetAllCoins() []CoinInfo {
	h.coinsMu.RLock()
	defer h.coinsMu.RUnlock()
	result := make([]CoinInfo, len(h.coins))
	copy(result, h.coins)
	return result
}
