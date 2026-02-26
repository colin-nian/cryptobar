package menubar

import (
	"encoding/json"
	"log"
	"path/filepath"

	"cryptobar/internal/config"
	"cryptobar/internal/datasource"
	"cryptobar/internal/i18n"

	"github.com/caseymrm/menuet"
)

var mbInstance *MenuBar

func SetGlobalMenuBar(mb *MenuBar) {
	mbInstance = mb
	menuet.SettingsCallback = onSettingsChanged
	menuet.CoinSelectionCallback = onCoinSelectionChanged
}

type settingsJSON struct {
	ShowChange   bool    `json:"show_change"`
	CompactName  bool    `json:"compact_name"`
	FontSize     float64 `json:"font_size"`
	IconMode     string  `json:"icon_mode"`
	LogoColor    string  `json:"logo_color"`
	Language     string  `json:"language"`
	DataSource   string  `json:"data_source"`
	LogPath      string  `json:"log_path"`
	AssetsPath   string  `json:"assets_path"`
	USDTAddress  string  `json:"usdt_address"`
	StartAtLogin bool    `json:"start_at_login"`

	SourceURLs  map[string]config.SourceURLConfig `json:"source_urls,omitempty"`
	DefaultURLs map[string]config.SourceURLConfig `json:"default_urls,omitempty"`
	Strings     map[string]string                 `json:"strings,omitempty"`
}

func (mb *MenuBar) OpenSettings() {
	cfg := config.Get()
	if cfg == nil {
		return
	}

	s := cfg.Settings
	i18n.Set(s.Language)

	data := settingsJSON{
		ShowChange:   s.ShowChange,
		CompactName:  s.CompactName,
		FontSize:     s.FontSize,
		IconMode:     s.IconMode,
		LogoColor:    s.LogoColor,
		Language:     s.Language,
		DataSource:   s.DataSource,
		LogPath:      config.LogFilePath(),
		AssetsPath:   filepath.Join(config.ConfigDir(), "assets"),
		USDTAddress:  "TV75kwC1n7yA33kMi9yYw7EVgybPua4fvQ",
		StartAtLogin: menuet.App().RunningAtStartup(),
		SourceURLs:   s.SourceURLs,
		DefaultURLs:  datasource.AllDefaultURLs(),
		Strings:      i18n.ObjCStrings(),
	}

	b, err := json.Marshal(data)
	if err != nil {
		log.Printf("[Settings] marshal error: %v", err)
		return
	}

	menuet.ShowSettings(string(b))
}

func parseBool(raw json.RawMessage) bool {
	var b bool
	if json.Unmarshal(raw, &b) == nil {
		return b
	}
	var n float64
	if json.Unmarshal(raw, &n) == nil {
		return n != 0
	}
	return false
}

func onSettingsChanged(jsonStr string) {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal([]byte(jsonStr), &raw); err != nil {
		log.Printf("[Settings] unmarshal callback error: %v", err)
		return
	}

	var s settingsJSON
	s.ShowChange = parseBool(raw["show_change"])
	s.CompactName = parseBool(raw["compact_name"])

	if v, ok := raw["font_size"]; ok {
		json.Unmarshal(v, &s.FontSize)
	}
	if v, ok := raw["icon_mode"]; ok {
		json.Unmarshal(v, &s.IconMode)
	}
	if v, ok := raw["logo_color"]; ok {
		json.Unmarshal(v, &s.LogoColor)
	}
	if v, ok := raw["language"]; ok {
		json.Unmarshal(v, &s.Language)
	}
	if v, ok := raw["data_source"]; ok {
		json.Unmarshal(v, &s.DataSource)
	}
	if v, ok := raw["source_urls"]; ok {
		json.Unmarshal(v, &s.SourceURLs)
	}

	s.StartAtLogin = parseBool(raw["start_at_login"])
	currentStartup := menuet.App().RunningAtStartup()
	if s.StartAtLogin != currentStartup {
		menuet.App().ToggleStartup()
		log.Printf("[Settings] start at login: %v", s.StartAtLogin)
	}

	if s.Language != "" {
		i18n.Set(s.Language)
	}

	log.Printf("[Settings] applied: fontSize=%.0f iconMode=%s logoColor=%s lang=%s showChange=%v compactName=%v ds=%s",
		s.FontSize, s.IconMode, s.LogoColor, s.Language, s.ShowChange, s.CompactName, s.DataSource)

	oldSettings := config.GetSettings()
	needSwitchDS := oldSettings.DataSource != s.DataSource

	newSettings := config.DisplaySettings{
		ShowChange:  s.ShowChange,
		CompactName: s.CompactName,
		FontSize:    s.FontSize,
		IconMode:    s.IconMode,
		LogoColor:   s.LogoColor,
		Language:    s.Language,
		DataSource:  s.DataSource,
		SourceURLs:  s.SourceURLs,
	}
	if err := config.UpdateSettings(newSettings); err != nil {
		log.Printf("[Settings] save error: %v", err)
	}

	if mbInstance != nil {
		mbInstance.updateTitle()
		if needSwitchDS {
			go mbInstance.switchDataSource(s.DataSource)
		}
	}
}
