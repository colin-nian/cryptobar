package main

import (
	"io"
	"log"
	"os"

	"cryptobar/internal/alert"
	"cryptobar/internal/config"
	"cryptobar/internal/datasource"
	_ "cryptobar/internal/datasource"
	"cryptobar/internal/i18n"
	"cryptobar/internal/icons"
	"cryptobar/internal/menubar"
	"cryptobar/internal/price"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("[Main] failed to load config: %v", err)
	}

	i18n.Set(cfg.Settings.Language)

	setupLogging()
	log.Println("[Main] CryptoBar starting...")
	icons.Init(config.ConfigDir())
	log.Printf("[Main] loaded %d coins, data source: %s", len(cfg.Coins), cfg.Settings.DataSource)

	store := price.NewStore()

	ds, err := datasource.New(cfg.Settings.DataSource, store)
	if err != nil {
		log.Printf("[Main] failed to create data source %s, falling back to binance: %v", cfg.Settings.DataSource, err)
		ds, _ = datasource.New("binance", store)
	}

	go func() {
		coins, err := ds.FetchCoinList(config.ConfigDir())
		if err != nil {
			log.Printf("[Main] failed to load coin list: %v", err)
		} else {
			log.Printf("[Main] loaded %d coins from %s", len(coins), ds.Name())
		}
	}()

	pairs := make([]string, len(cfg.Coins))
	for i, c := range cfg.Coins {
		pairs[i] = c.Pair
	}
	ds.SetPairs(pairs)
	ds.Start()
	log.Printf("[Main] %s data source started", ds.Name())

	_ = alert.NewManager(store)
	log.Println("[Main] alert manager initialized")

	mb := menubar.New(store, ds)
	menubar.SetGlobalMenuBar(mb)
	log.Println("[Main] starting menu bar...")
	mb.Run()
}

func setupLogging() {
	logDir := config.LogDir()
	if err := os.MkdirAll(logDir, 0755); err != nil {
		log.Printf("[Main] failed to create log dir: %v", err)
		return
	}

	logFile, err := os.OpenFile(config.LogFilePath(), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Printf("[Main] failed to open log file: %v", err)
		return
	}

	mw := io.MultiWriter(os.Stderr, logFile)
	log.SetOutput(mw)
	log.SetFlags(log.Ltime | log.Lshortfile)
}
