package menubar

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"cryptobar/internal/config"
	"cryptobar/internal/datasource"
	"cryptobar/internal/i18n"
	"cryptobar/internal/icons"
	"cryptobar/internal/price"

	"github.com/caseymrm/menuet"
)

func coinIcon(symbol string) string {
	return icons.URL(symbol)
}

type MenuBar struct {
	store    *price.Store
	mu       sync.Mutex
	throttle *time.Ticker
	ds       datasource.DataSource
}

func New(store *price.Store, ds datasource.DataSource) *MenuBar {
	return &MenuBar{
		store: store,
		ds:    ds,
	}
}

func (mb *MenuBar) SetDataSource(ds datasource.DataSource) {
	mb.mu.Lock()
	defer mb.mu.Unlock()
	mb.ds = ds
}

func (mb *MenuBar) GetDataSource() datasource.DataSource {
	mb.mu.Lock()
	defer mb.mu.Unlock()
	return mb.ds
}

func (mb *MenuBar) Run() {
	mb.throttle = time.NewTicker(1 * time.Second)
	go mb.updateLoop()

	menuet.App().Label = "com.cryptobar.app"
	menuet.App().HideStartup()
	menuet.App().Children = mb.menuItems
	mb.updateTitle()
	menuet.App().RunApplication()
}

func (mb *MenuBar) updateLoop() {
	for range mb.throttle.C {
		mb.updateTitle()
	}
}

func (mb *MenuBar) updateTitle() {
	cfg := config.Get()
	if cfg == nil {
		return
	}
	settings := cfg.Settings

	fontSize := settings.FontSize
	if fontSize == 0 {
		fontSize = 12
	}

	var segments []menuet.TitleSegment
	first := true

	for _, coin := range cfg.Coins {
		if !coin.ShowInBar {
			continue
		}

		if !first {
			segments = append(segments, menuet.TitleSegment{Text: " "})
		}
		first = false

		iconMode := settings.IconMode
		if iconMode == "" {
			iconMode = "logo"
		}
		icon := coinIcon(coin.Symbol)
		gray := settings.LogoColor == "gray"

		switch iconMode {
		case "logo":
			if icon != "" {
				segments = append(segments, menuet.TitleSegment{Image: icon, Grayscale: gray})
			} else {
				segments = append(segments, menuet.TitleSegment{Text: coin.Symbol + " "})
			}
		case "text":
			if settings.CompactName {
				segments = append(segments, menuet.TitleSegment{Text: coin.Symbol + " "})
			} else {
				name := coin.Name
				if name == "" {
					name = coin.Symbol
				}
				segments = append(segments, menuet.TitleSegment{Text: coin.Symbol + " " + name + " "})
			}
		case "both":
			if icon != "" {
				segments = append(segments, menuet.TitleSegment{Image: icon, Grayscale: gray})
			}
			segments = append(segments, menuet.TitleSegment{Text: coin.Symbol + " "})
		default:
			if icon != "" {
				segments = append(segments, menuet.TitleSegment{Image: icon, Grayscale: gray})
			} else {
				segments = append(segments, menuet.TitleSegment{Text: coin.Symbol + " "})
			}
		}

		data, ok := mb.store.Get(coin.Pair)
		if !ok {
			segments = append(segments, menuet.TitleSegment{Text: "?"})
			continue
		}

		priceStr := "$" + data.FormatPrice(coin.Precision)
		if settings.ShowChange {
			change := data.Change24h()
			priceStr += fmt.Sprintf(" %+.1f%%", change)
		}
		segments = append(segments, menuet.TitleSegment{Text: priceStr})
	}

	if len(segments) == 0 {
		segments = append(segments, menuet.TitleSegment{Text: "CryptoBar"})
	}

	menuet.App().SetMenuState(&menuet.MenuState{
		FontSize: fontSize,
		Segments: segments,
	})
}

func (mb *MenuBar) menuItems() []menuet.MenuItem {
	cfg := config.Get()
	if cfg == nil {
		return nil
	}
	settings := cfg.Settings

	fontSize := int(settings.FontSize)
	if fontSize == 0 {
		fontSize = 12
	}

	var items []menuet.MenuItem

	for _, coin := range cfg.Coins {
		c := coin
		data, ok := mb.store.Get(c.Pair)

		var title string
		if ok {
			priceStr := data.FormatPrice(c.Precision)
			change := data.Change24h()
			changeStr := fmt.Sprintf("%+.1f%%", change)
			title = fmt.Sprintf("%-6s  $%-10s  %s", c.Symbol, priceStr, changeStr)
		} else {
			title = fmt.Sprintf("%-6s  ---", c.Symbol)
		}

		showInBar := c.ShowInBar
		items = append(items, menuet.MenuItem{
			Text:       title,
			FontSize:   fontSize,
			FontWeight: menuet.WeightRegular,
			State:      showInBar,
			Image:      coinIcon(c.Symbol),
			Clicked: func() {
				_ = config.SetShowInBar(c.Pair, !showInBar)
				mb.updateTitle()
			},
		})
	}

	items = append(items, menuet.MenuItem{Type: menuet.Separator})

	items = append(items, menuet.MenuItem{
		Text:     i18n.T("select_coins"),
		FontSize: fontSize,
		Clicked: func() {
			mb.OpenCoinSelector()
		},
	})

	items = append(items, menuet.MenuItem{Type: menuet.Separator})

	items = append(items, menuet.MenuItem{
		Text:     i18n.T("settings"),
		FontSize: fontSize,
		Clicked: func() {
			mb.OpenSettings()
		},
	})

	items = append(items, menuet.MenuItem{Type: menuet.Separator})

	items = append(items, menuet.MenuItem{
		Text:     i18n.T("refresh_coin_list"),
		FontSize: fontSize,
		Clicked: func() {
			go func() {
				ds := mb.GetDataSource()
				if _, err := ds.FetchCoinList(config.ConfigDir()); err != nil {
					log.Printf("[Menu] refresh coin list error: %v", err)
				}
			}()
		},
	})

	return items
}

func (mb *MenuBar) settingsMenu() []menuet.MenuItem {
	settings := config.GetSettings()

	fontSize := int(settings.FontSize)
	if fontSize == 0 {
		fontSize = 12
	}

	var items []menuet.MenuItem

	items = append(items, menuet.MenuItem{
		Text:     "Show 24h Change",
		FontSize: fontSize,
		State:    settings.ShowChange,
		Clicked: func() {
			_ = config.ToggleShowChange()
			mb.updateTitle()
		},
	})

	items = append(items, menuet.MenuItem{
		Text:     "Compact Name",
		FontSize: fontSize,
		State:    settings.CompactName,
		Clicked: func() {
			_ = config.ToggleCompactName()
			mb.updateTitle()
		},
	})

	items = append(items, menuet.MenuItem{Type: menuet.Separator})

	items = append(items, menuet.MenuItem{
		Text:     "── Font Size ──",
		FontSize: fontSize,
	})

	fontSizes := []struct {
		label string
		size  float64
	}{
		{"Small (10)", 10},
		{"Medium (12)", 12},
		{"Large (14)", 14},
	}
	for _, fs := range fontSizes {
		s := fs
		items = append(items, menuet.MenuItem{
			Text:     s.label,
			FontSize: fontSize,
			State:    settings.FontSize == s.size,
			Clicked: func() {
				_ = config.SetFontSize(s.size)
				mb.updateTitle()
			},
		})
	}

	items = append(items, menuet.MenuItem{Type: menuet.Separator})

	items = append(items, menuet.MenuItem{
		Text:     "── Data Source ──",
		FontSize: fontSize,
	})

	sources := []string{"binance", "htx", "gateio"}
	for _, src := range sources {
		s := src
		items = append(items, menuet.MenuItem{
			Text:     datasource.SourceDisplayName(s),
			FontSize: fontSize,
			State:    settings.DataSource == s,
			Clicked: func() {
				mb.switchDataSource(s)
			},
		})
	}

	return items
}

func (mb *MenuBar) switchDataSource(name string) {
	currentSettings := config.GetSettings()
	if currentSettings.DataSource == name {
		return
	}

	oldName := currentSettings.DataSource
	oldDS := mb.GetDataSource()
	oldDS.Stop()

	_ = config.SetDataSource(name)

	newDS, err := datasource.New(name, mb.store)
	if err != nil {
		log.Printf("[Menu] failed to create data source %s: %v", name, err)
		_ = config.SetDataSource(oldName)
		oldDS.Start()
		go func() {
			menuet.App().Alert(menuet.Alert{
				MessageText:     i18n.T("switch_failed"),
				InformativeText: fmt.Sprintf(i18n.T("switch_failed_desc"), datasource.SourceDisplayName(name), err),
				Buttons:         []string{"OK"},
			})
		}()
		return
	}

	mb.SetDataSource(newDS)

	go func() {
		if _, err := newDS.FetchCoinList(config.ConfigDir()); err != nil {
			log.Printf("[Menu] failed to fetch coin list for %s: %v", name, err)
		}

		cfg := config.Get()
		if cfg != nil {
			pairs := make([]string, len(cfg.Coins))
			for i, c := range cfg.Coins {
				pairs[i] = c.Pair
			}
			newDS.SetPairs(pairs)
		}
		newDS.Start()
		log.Printf("[Menu] switched data source to %s", name)

		// Wait and check if data arrives; if not, warn user
		time.Sleep(20 * time.Second)
		hasData := false
		if cfg != nil {
			for _, c := range cfg.Coins {
				if _, ok := mb.store.Get(c.Pair); ok {
					hasData = true
					break
				}
			}
		}
		if !hasData {
			log.Printf("[Menu] warning: no data received from %s after 20s", name)
			menuet.App().Alert(menuet.Alert{
				MessageText:     i18n.T("conn_issue"),
				InformativeText: fmt.Sprintf(i18n.T("conn_issue_desc"), datasource.SourceDisplayName(name)),
				Buttons:         []string{"OK"},
			})
		}
	}()
}

type coinSelectorJSON struct {
	Coins         []coinEntry       `json:"coins"`
	SelectedPairs []string          `json:"selected_pairs"`
	Strings       map[string]string `json:"strings,omitempty"`
}

type coinEntry struct {
	Symbol string `json:"symbol"`
	Name   string `json:"name"`
	Pair   string `json:"pair"`
}

func (mb *MenuBar) OpenCoinSelector() {
	ds := mb.GetDataSource()
	allCoins := ds.GetAllCoins()
	cfg := config.Get()

	var entries []coinEntry
	for _, c := range allCoins {
		entries = append(entries, coinEntry{
			Symbol: c.Symbol,
			Name:   c.Name,
			Pair:   c.Pair,
		})
	}

	var selectedPairs []string
	if cfg != nil {
		for _, c := range cfg.Coins {
			selectedPairs = append(selectedPairs, c.Pair)
		}
	}

	data := coinSelectorJSON{
		Coins:         entries,
		SelectedPairs: selectedPairs,
		Strings:       i18n.ObjCStrings(),
	}

	b, err := json.Marshal(data)
	if err != nil {
		log.Printf("[CoinSelector] marshal error: %v", err)
		return
	}

	menuet.ShowCoinSelector(string(b))
}

type coinSelectionResult struct {
	Selected []coinEntry `json:"selected"`
}

func onCoinSelectionChanged(jsonStr string) {
	var result coinSelectionResult
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		log.Printf("[CoinSelector] unmarshal error: %v", err)
		return
	}

	cfg := config.Get()
	if cfg == nil {
		return
	}

	selectedPairs := make(map[string]bool)
	for _, c := range result.Selected {
		selectedPairs[c.Pair] = true
	}

	existingCoins := make(map[string]config.CoinConfig)
	for _, c := range cfg.Coins {
		existingCoins[c.Pair] = c
	}

	var newCoins []config.CoinConfig
	for _, c := range result.Selected {
		if existing, ok := existingCoins[c.Pair]; ok {
			newCoins = append(newCoins, existing)
		} else {
			newCoins = append(newCoins, config.CoinConfig{
				Symbol:    c.Symbol,
				Pair:      c.Pair,
				Name:      c.Name,
				Precision: autoPrec(c.Symbol),
				ShowInBar: false,
			})
		}
	}

	if err := config.UpdateCoins(newCoins); err != nil {
		log.Printf("[CoinSelector] save error: %v", err)
	}

	if mbInstance != nil {
		mbInstance.SyncPairs()
		mbInstance.updateTitle()
	}

	log.Printf("[CoinSelector] updated: %d coins selected", len(newCoins))
}

func (mb *MenuBar) coinSelectMenu(ds datasource.DataSource) []menuet.MenuItem {
	allCoins := ds.GetAllCoins()
	cfg := config.Get()

	selectedPairs := make(map[string]bool)
	if cfg != nil {
		for _, c := range cfg.Coins {
			selectedPairs[c.Pair] = true
		}
	}

	fontSize := int(cfg.Settings.FontSize)
	if fontSize == 0 {
		fontSize = 12
	}

	topSymbols := []string{
		"BTCUSDT", "ETHUSDT", "BNBUSDT", "SOLUSDT", "XRPUSDT",
		"ADAUSDT", "DOGEUSDT", "DOTUSDT", "AVAXUSDT", "LINKUSDT",
		"MATICUSDT", "UNIUSDT", "ATOMUSDT", "LTCUSDT", "ETCUSDT",
		"NEARUSDT", "APTUSDT", "ARBUSDT", "OPUSDT", "SUIUSDT",
	}

	var items []menuet.MenuItem
	items = append(items, menuet.MenuItem{Text: i18n.T("popular"), FontSize: fontSize})

	for _, pair := range topSymbols {
		p := pair
		coin, found := ds.FindCoinByPair(p)
		if !found {
			continue
		}
		selected := selectedPairs[p]
		items = append(items, menuet.MenuItem{
			Text:     fmt.Sprintf("%-8s %s", coin.Symbol, coin.Name),
			FontSize: fontSize,
			State:    selected,
			Image:    coinIcon(coin.Symbol),
			Clicked: func() {
				mb.toggleCoin(p, coin, !selected, ds)
			},
		})
	}

	items = append(items, menuet.MenuItem{Type: menuet.Separator})

	letterGroups := make(map[string][]datasource.CoinInfo)
	for _, c := range allCoins {
		if len(c.Symbol) == 0 {
			continue
		}
		first := strings.ToUpper(string(c.Symbol[0]))
		letterGroups[first] = append(letterGroups[first], c)
	}

	letters := "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	for _, ch := range letters {
		letter := string(ch)
		coins, ok := letterGroups[letter]
		if !ok || len(coins) == 0 {
			continue
		}
		letterCoins := coins
		items = append(items, menuet.MenuItem{
			Text:     letter,
			FontSize: fontSize,
			Children: func() []menuet.MenuItem {
				return mb.letterGroupItems(letterCoins, ds)
			},
		})
	}

	return items
}

func (mb *MenuBar) letterGroupItems(coins []datasource.CoinInfo, ds datasource.DataSource) []menuet.MenuItem {
	cfg := config.Get()
	currentSelected := make(map[string]bool)
	fontSize := 12
	if cfg != nil {
		for _, c := range cfg.Coins {
			currentSelected[c.Pair] = true
		}
		if cfg.Settings.FontSize > 0 {
			fontSize = int(cfg.Settings.FontSize)
		}
	}

	var items []menuet.MenuItem
	for _, coin := range coins {
		c := coin
		selected := currentSelected[c.Pair]
		items = append(items, menuet.MenuItem{
			Text:     fmt.Sprintf("%-8s %s", c.Symbol, c.Name),
			FontSize: fontSize,
			State:    selected,
			Image:    coinIcon(c.Symbol),
			Clicked: func() {
				mb.toggleCoin(c.Pair, c, !selected, ds)
			},
		})
	}
	return items
}

func (mb *MenuBar) toggleCoin(pair string, coin datasource.CoinInfo, selected bool, ds datasource.DataSource) {
	if selected {
		cfg := config.Get()
		found := false
		if cfg != nil {
			for _, c := range cfg.Coins {
				if c.Pair == pair {
					found = true
					break
				}
			}
		}
		if !found {
			newCoin := config.CoinConfig{
				Symbol:    coin.Symbol,
				Pair:      coin.Pair,
				Name:      coin.Name,
				Precision: autoPrec(coin.Symbol),
				ShowInBar: false,
			}
			cfg.Coins = append(cfg.Coins, newCoin)
			_ = config.UpdateCoins(cfg.Coins)
		}
	} else {
		_ = config.SetCoinSelected(pair, false)
	}
	mb.SyncPairs()
	mb.updateTitle()
}

func (mb *MenuBar) SyncPairs() {
	cfg := config.Get()
	if cfg == nil {
		return
	}
	ds := mb.GetDataSource()
	pairs := make([]string, len(cfg.Coins))
	for i, c := range cfg.Coins {
		pairs[i] = c.Pair
	}
	ds.SetPairs(pairs)
}

func autoPrec(symbol string) int {
	switch symbol {
	case "BTC", "ETH", "BNB":
		return 0
	case "SOL":
		return 1
	default:
		return -1
	}
}
