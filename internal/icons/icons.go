package icons

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const (
	ccAPI      = "https://min-api.cryptocompare.com/data/all/coinlist?summary=true"
	ccBase     = "https://www.cryptocompare.com"
	coincapFmt = "https://assets.coincap.io/assets/icons/%s@2x.png"
)

var (
	mu       sync.RWMutex
	iconMap  map[string]string // symbol -> full URL
	cacheDir string
)

type ccResponse struct {
	Data map[string]ccCoin `json:"Data"`
}

type ccCoin struct {
	ImageUrl string `json:"ImageUrl"`
}

type iconCache struct {
	Icons     map[string]string `json:"icons"`
	UpdatedAt time.Time         `json:"updated_at"`
}

func Init(dir string) {
	cacheDir = dir
	loadCache()
	go refresh()
}

func URL(symbol string) string {
	s := strings.ToUpper(symbol)
	mu.RLock()
	url, ok := iconMap[s]
	mu.RUnlock()
	if ok && url != "" {
		return url
	}
	return fmt.Sprintf(coincapFmt, strings.ToLower(symbol))
}

func loadCache() {
	path := cachePath()
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	var cache iconCache
	if json.Unmarshal(data, &cache) == nil && len(cache.Icons) > 0 {
		mu.Lock()
		iconMap = cache.Icons
		mu.Unlock()
		log.Printf("[Icons] loaded %d icons from cache", len(cache.Icons))

		if time.Since(cache.UpdatedAt) < 7*24*time.Hour {
			return
		}
	}
}

func refresh() {
	icons, err := fetchAll()
	if err != nil {
		log.Printf("[Icons] fetch error: %v", err)
		return
	}
	mu.Lock()
	iconMap = icons
	mu.Unlock()
	log.Printf("[Icons] fetched %d icon URLs from CryptoCompare", len(icons))
	saveCache(icons)
}

func fetchAll() (map[string]string, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(ccAPI)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result ccResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	icons := make(map[string]string, len(result.Data))
	for sym, coin := range result.Data {
		if coin.ImageUrl != "" {
			icons[strings.ToUpper(sym)] = ccBase + coin.ImageUrl
		}
	}
	return icons, nil
}

func saveCache(icons map[string]string) {
	if cacheDir == "" {
		return
	}
	cache := iconCache{Icons: icons, UpdatedAt: time.Now()}
	data, err := json.Marshal(cache)
	if err != nil {
		return
	}
	_ = os.WriteFile(cachePath(), data, 0644)
}

func cachePath() string {
	return filepath.Join(cacheDir, "icon_urls.json")
}
