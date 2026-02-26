package datasource

import (
	"cryptobar/internal/config"
	"cryptobar/internal/price"
	"fmt"
)

type CoinInfo struct {
	Symbol    string `json:"symbol"`
	BaseAsset string `json:"base_asset"`
	Name      string `json:"name"`
	Pair      string `json:"pair"`
}

type DataSource interface {
	Name() string
	Start()
	Stop()
	SetPairs(pairs []string)
	FetchCoinList(cacheDir string) ([]CoinInfo, error)
	FindCoinByPair(pair string) (CoinInfo, bool)
	GetAllCoins() []CoinInfo
}

type Factory func(store *price.Store) DataSource

var registry = map[string]Factory{}

func Register(name string, factory Factory) {
	registry[name] = factory
}

func New(name string, store *price.Store) (DataSource, error) {
	factory, ok := registry[name]
	if !ok {
		return nil, fmt.Errorf("unknown data source: %s", name)
	}
	return factory(store), nil
}

func Available() []string {
	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}
	return names
}

var knownNames = map[string]string{
	"BTC":   "Bitcoin",
	"ETH":   "Ethereum",
	"BNB":   "BNB",
	"SOL":   "Solana",
	"XRP":   "XRP",
	"ADA":   "Cardano",
	"DOGE":  "Dogecoin",
	"DOT":   "Polkadot",
	"MATIC": "Polygon",
	"AVAX":  "Avalanche",
	"LINK":  "Chainlink",
	"UNI":   "Uniswap",
	"ATOM":  "Cosmos",
	"LTC":   "Litecoin",
	"ETC":   "Ethereum Classic",
	"FIL":   "Filecoin",
	"NEAR":  "NEAR Protocol",
	"APT":   "Aptos",
	"ARB":   "Arbitrum",
	"OP":    "Optimism",
	"SUI":   "Sui",
	"SEI":   "Sei",
	"TIA":   "Celestia",
	"INJ":   "Injective",
	"RUNE":  "THORChain",
	"AAVE":  "Aave",
	"MKR":   "Maker",
	"CRV":   "Curve",
	"SNX":   "Synthetix",
	"COMP":  "Compound",
	"LDO":   "Lido DAO",
	"SHIB":  "Shiba Inu",
	"PEPE":  "Pepe",
	"WIF":   "dogwifhat",
	"BONK":  "Bonk",
	"FLOKI": "Floki",
	"TRX":   "TRON",
	"TON":   "Toncoin",
	"HBAR":  "Hedera",
	"VET":   "VeChain",
	"ALGO":  "Algorand",
	"FTM":   "Fantom",
	"SAND":  "The Sandbox",
	"MANA":  "Decentraland",
	"AXS":   "Axie Infinity",
	"GALA":  "Gala",
	"IMX":   "Immutable",
	"RNDR":  "Render",
	"FET":   "Fetch.ai",
	"AGIX":  "SingularityNET",
	"WLD":   "Worldcoin",
	"JUP":   "Jupiter",
	"PYTH":  "Pyth Network",
	"STX":   "Stacks",
	"ICP":   "Internet Computer",
	"EGLD":  "MultiversX",
	"THETA": "Theta Network",
	"XLM":   "Stellar",
	"XMR":   "Monero",
	"BCH":   "Bitcoin Cash",
	"LEO":   "LEO Token",
	"HYPE":  "Hyperliquid",
}

func HumanName(symbol string) string {
	if name, ok := knownNames[symbol]; ok {
		return name
	}
	return symbol
}

var topCoins = []string{
	"BTC", "ETH", "BNB", "SOL", "XRP", "ADA", "DOGE", "DOT",
	"AVAX", "LINK", "MATIC", "UNI", "ATOM", "LTC", "ETC",
	"NEAR", "APT", "ARB", "OP", "SUI",
}

func CoinPriority(symbol string) int {
	for i, s := range topCoins {
		if s == symbol {
			return i
		}
	}
	return 1000
}

func IsStablecoin(symbol string) bool {
	stables := map[string]bool{
		"USDC": true, "BUSD": true, "DAI": true, "TUSD": true,
		"USDP": true, "GUSD": true, "FRAX": true, "LUSD": true,
		"USDD": true, "FDUSD": true, "PYUSD": true, "USDS": true,
		"USDT": true,
	}
	return stables[symbol]
}

func IsLeveraged(symbol string) bool {
	suffixes := []string{"UP", "DOWN", "BULL", "BEAR"}
	for _, s := range suffixes {
		if len(symbol) > len(s) && symbol[len(symbol)-len(s):] == s {
			return true
		}
	}
	return false
}

type DefaultURL struct {
	WSURL  string
	APIURL string
}

var defaultURLRegistry = map[string]DefaultURL{}

func RegisterDefaultURLs(name string, ws, api string) {
	defaultURLRegistry[name] = DefaultURL{WSURL: ws, APIURL: api}
}

func GetDefaultURLs(name string) DefaultURL {
	return defaultURLRegistry[name]
}

func AllDefaultURLs() map[string]config.SourceURLConfig {
	result := make(map[string]config.SourceURLConfig)
	for name, u := range defaultURLRegistry {
		result[name] = config.SourceURLConfig{WSURL: u.WSURL, APIURL: u.APIURL}
	}
	return result
}

func SourceDisplayName(name string) string {
	switch name {
	case "binance":
		return "Binance"
	case "htx":
		return "HTX (Huobi)"
	case "gateio":
		return "Gate.io"
	default:
		return name
	}
}
