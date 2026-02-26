package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

type CoinConfig struct {
	Symbol    string `json:"symbol"`
	Pair      string `json:"pair"`
	Name      string `json:"name"`
	Precision int    `json:"precision"`
	ShowInBar bool   `json:"show_in_bar"`
}

type AlertRule struct {
	Pair    string  `json:"pair"`
	Above   float64 `json:"above,omitempty"`
	Below   float64 `json:"below,omitempty"`
	Enabled bool    `json:"enabled"`
}

type SourceURLConfig struct {
	WSURL  string `json:"ws_url"`
	APIURL string `json:"api_url"`
}

type DisplaySettings struct {
	ShowChange  bool                       `json:"show_change"`
	CompactName bool                       `json:"compact_name"`
	DataSource  string                     `json:"data_source"`
	FontSize    float64                    `json:"font_size"`
	IconMode    string                     `json:"icon_mode"`
	LogoColor   string                     `json:"logo_color"`
	Language    string                     `json:"language"`
	SourceURLs  map[string]SourceURLConfig `json:"source_urls,omitempty"`
}

type AppConfig struct {
	Coins    []CoinConfig    `json:"coins"`
	Alerts   []AlertRule     `json:"alerts"`
	Settings DisplaySettings `json:"settings"`
}

var (
	instance *AppConfig
	mu       sync.RWMutex
	cfgPath  string
)

func defaultConfig() *AppConfig {
	return &AppConfig{
		Coins: []CoinConfig{
			{Symbol: "BTC", Pair: "BTCUSDT", Name: "Bitcoin", Precision: 0, ShowInBar: true},
			{Symbol: "ETH", Pair: "ETHUSDT", Name: "Ethereum", Precision: 0, ShowInBar: true},
			{Symbol: "BNB", Pair: "BNBUSDT", Name: "BNB", Precision: 0, ShowInBar: true},
			{Symbol: "SOL", Pair: "SOLUSDT", Name: "Solana", Precision: 1, ShowInBar: true},
		},
		Alerts: []AlertRule{},
		Settings: DisplaySettings{
			ShowChange:  true,
			CompactName: true,
			DataSource:  "binance",
			FontSize:    12,
			IconMode:    "logo",
			LogoColor:   "color",
			Language:    "en",
		},
	}
}

func configDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".cryptobar")
}

func ConfigDir() string {
	return configDir()
}

func LogDir() string {
	return filepath.Join(configDir(), "logs")
}

func LogFilePath() string {
	return filepath.Join(LogDir(), "cryptobar.log")
}

func Load() (*AppConfig, error) {
	mu.Lock()
	defer mu.Unlock()

	dir := configDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	cfgPath = filepath.Join(dir, "config.json")
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		if os.IsNotExist(err) {
			instance = defaultConfig()
			return instance, save()
		}
		return nil, err
	}

	var cfg AppConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		instance = defaultConfig()
		return instance, save()
	}

	if cfg.Settings.DataSource == "" {
		cfg.Settings.DataSource = "binance"
		cfg.Settings.ShowChange = true
	}
	if cfg.Settings.FontSize == 0 {
		cfg.Settings.FontSize = 12
	}
	if cfg.Settings.IconMode == "" {
		cfg.Settings.IconMode = "logo"
	}
	if cfg.Settings.LogoColor == "" {
		cfg.Settings.LogoColor = "color"
	}
	if cfg.Settings.Language == "" {
		cfg.Settings.Language = "en"
	}

	instance = &cfg
	return instance, nil
}

func save() error {
	data, err := json.MarshalIndent(instance, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(cfgPath, data, 0644)
}

func Save() error {
	mu.Lock()
	defer mu.Unlock()
	return save()
}

func Get() *AppConfig {
	mu.RLock()
	defer mu.RUnlock()
	return instance
}

func UpdateCoins(coins []CoinConfig) error {
	mu.Lock()
	defer mu.Unlock()
	instance.Coins = coins
	return save()
}

func SetCoinSelected(pair string, selected bool) error {
	mu.Lock()
	defer mu.Unlock()

	if selected {
		for _, c := range instance.Coins {
			if c.Pair == pair {
				return nil
			}
		}
		instance.Coins = append(instance.Coins, CoinConfig{
			Pair:      pair,
			ShowInBar: false,
		})
	} else {
		filtered := instance.Coins[:0]
		for _, c := range instance.Coins {
			if c.Pair != pair {
				filtered = append(filtered, c)
			}
		}
		instance.Coins = filtered
	}
	return save()
}

func SetShowInBar(pair string, show bool) error {
	mu.Lock()
	defer mu.Unlock()
	for i := range instance.Coins {
		if instance.Coins[i].Pair == pair {
			instance.Coins[i].ShowInBar = show
			break
		}
	}
	return save()
}

func AddAlert(rule AlertRule) error {
	mu.Lock()
	defer mu.Unlock()
	instance.Alerts = append(instance.Alerts, rule)
	return save()
}

func RemoveAlert(pair string) error {
	mu.Lock()
	defer mu.Unlock()
	filtered := instance.Alerts[:0]
	for _, a := range instance.Alerts {
		if a.Pair != pair {
			filtered = append(filtered, a)
		}
	}
	instance.Alerts = filtered
	return save()
}

func GetAlerts() []AlertRule {
	mu.RLock()
	defer mu.RUnlock()
	result := make([]AlertRule, len(instance.Alerts))
	copy(result, instance.Alerts)
	return result
}

func IsCoinSelected(pair string) bool {
	mu.RLock()
	defer mu.RUnlock()
	for _, c := range instance.Coins {
		if c.Pair == pair {
			return true
		}
	}
	return false
}

func ToggleShowChange() error {
	mu.Lock()
	defer mu.Unlock()
	instance.Settings.ShowChange = !instance.Settings.ShowChange
	return save()
}

func ToggleCompactName() error {
	mu.Lock()
	defer mu.Unlock()
	instance.Settings.CompactName = !instance.Settings.CompactName
	return save()
}

func SetDataSource(source string) error {
	mu.Lock()
	defer mu.Unlock()
	instance.Settings.DataSource = source
	return save()
}

func SetFontSize(size float64) error {
	mu.Lock()
	defer mu.Unlock()
	instance.Settings.FontSize = size
	return save()
}

func SetIconMode(mode string) error {
	mu.Lock()
	defer mu.Unlock()
	instance.Settings.IconMode = mode
	return save()
}

func SetSourceURL(name string, urls SourceURLConfig) error {
	mu.Lock()
	defer mu.Unlock()
	if instance.Settings.SourceURLs == nil {
		instance.Settings.SourceURLs = make(map[string]SourceURLConfig)
	}
	instance.Settings.SourceURLs[name] = urls
	return save()
}

func GetSourceURL(name string) SourceURLConfig {
	mu.RLock()
	defer mu.RUnlock()
	if instance.Settings.SourceURLs == nil {
		return SourceURLConfig{}
	}
	return instance.Settings.SourceURLs[name]
}

func GetSettings() DisplaySettings {
	mu.RLock()
	defer mu.RUnlock()
	return instance.Settings
}

func UpdateSettings(s DisplaySettings) error {
	mu.Lock()
	defer mu.Unlock()
	instance.Settings = s
	return save()
}
