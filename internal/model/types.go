package model

import (
	"time"
)

// TokenSource indicates where the price data originates.
type TokenSource string

const (
	SourceBinance TokenSource = "BINANCE"
	SourceDEX     TokenSource = "DEX"
)

// TokenType categorizes the cryptocurrency.
type TokenType string

const (
	TypeMajor  TokenType = "MAJOR"
	TypeMeme   TokenType = "MEME"
	TypeStable TokenType = "STABLE"
)

// WatchItem represents a token in the user's watchlist.
type WatchItem struct {
	Symbol        string      `json:"symbol"`         // e.g. "BTCUSDT" or "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v"
	DisplaySymbol string      `json:"display_symbol"` // e.g. "BTC", "SOL", "PEPE", "POPCAT"
	Name          string      `json:"name"`           // e.g. "Bitcoin", "Pepe"
	Source        TokenSource `json:"source"`         // SourceBinance or SourceDEX
	Category      TokenType   `json:"category"`       // TypeMajor, TypeMeme, TypeStable
	Chain         string      `json:"chain"`          // e.g. "binance", "solana", "base", "ethereum", "bsc"
	PairAddress   string      `json:"pair_address"`   // For DEX tokens
	AddedAt       time.Time   `json:"added_at"`
}

// MarketData holds the real-time ticker data for a token.
type MarketData struct {
	Symbol            string    `json:"symbol"`
	DisplaySymbol     string    `json:"display_symbol"`
	Source            TokenSource
	Category          TokenType
	Chain             string
	Price             float64   `json:"price"`
	PriceChange24h    float64   `json:"price_change_24h"` // Percentage (e.g. 5.25 for +5.25%)
	PriceChangeAmount float64   `json:"price_change_amount"`
	High24h           float64   `json:"high_24h"`
	Low24h            float64   `json:"low_24h"`
	Volume24h         float64   `json:"volume_24h"`       // In quote currency (e.g. USD)
	PriceHistory      []float64 `json:"price_history"`    // Up to 20 recent price samples for sparklines
	LastTickDirection int       `json:"last_tick"`        // +1 = up, -1 = down, 0 = equal
	LiquidityUSD      float64   `json:"liquidity_usd"`    // For DEX meme coins
	MarketCapUSD      float64   `json:"market_cap_usd"`   // For DEX meme coins
	LastUpdated       time.Time `json:"last_updated"`
}

// PriceAlert triggers when a token reaches a threshold.
type PriceAlert struct {
	ID          string    `json:"id"`
	Symbol      string    `json:"symbol"`
	TargetPrice float64   `json:"target_price"`
	Direction   string    `json:"direction"` // "ABOVE" or "BELOW"
	Triggered   bool      `json:"triggered"`
	CreatedAt   time.Time `json:"created_at"`
}

// PortfolioItem represents an asset holding from Binance account.
type PortfolioItem struct {
	Asset    string  `json:"asset"`     // e.g. "BTC", "ETH", "USDT"
	Free     float64 `json:"free"`      // Available balance
	Locked   float64 `json:"locked"`    // In open orders
	Total    float64 `json:"total"`     // Free + Locked
	PriceUSD float64 `json:"price_usd"` // Current price in USD
	ValueUSD float64 `json:"value_usd"` // Total * PriceUSD
}

// Config stores persistent settings.
type Config struct {
	BinanceAPIKey      string       `json:"binance_api_key"`
	BinanceAPISecret   string       `json:"binance_api_secret"`
	Currency           string       `json:"currency"`            // "USD" or "IDR"
	USDToIDR           float64      `json:"usd_to_idr"`          // Conversion rate
	RefreshIntervalSec int          `json:"refresh_interval_sec"`// For DEX polling (default 5s)
	SoundAlerts        bool         `json:"sound_alerts"`        // Enable terminal bell on alerts
	Watchlist          []WatchItem  `json:"watchlist"`
	Alerts             []PriceAlert `json:"alerts"`
}

// DexTokenPair represents a pair returned by DexScreener.
type DexTokenPair struct {
	ChainId     string `json:"chainId"`
	DexId       string `json:"dexId"`
	Url         string `json:"url"`
	PairAddress string `json:"pairAddress"`
	BaseToken   struct {
		Address string `json:"address"`
		Name    string `json:"name"`
		Symbol  string `json:"symbol"`
	} `json:"baseToken"`
	QuoteToken struct {
		Address string `json:"address"`
		Name    string `json:"name"`
		Symbol  string `json:"symbol"`
	} `json:"quoteToken"`
	PriceNative string `json:"priceNative"`
	PriceUsd    string `json:"priceUsd"`
	Txns        struct {
		H24 struct {
			Buys  int `json:"buys"`
			Sells int `json:"sells"`
		} `json:"h24"`
	} `json:"txns"`
	Volume struct {
		H24 float64 `json:"h24"`
		H6  float64 `json:"h6"`
		H1  float64 `json:"h1"`
		M5  float64 `json:"m5"`
	} `json:"volume"`
	PriceChange struct {
		M5  float64 `json:"m5"`
		H1  float64 `json:"h1"`
		H6  float64 `json:"h6"`
		H24 float64 `json:"h24"`
	} `json:"priceChange"`
	Liquidity struct {
		Usd   float64 `json:"usd"`
		Base  float64 `json:"base"`
		Quote float64 `json:"quote"`
	} `json:"liquidity"`
	Fdv float64 `json:"fdv"`
}

// DexSearchResponse is the response from DexScreener search API.
type DexSearchResponse struct {
	SchemaVersion string         `json:"schemaVersion"`
	Pairs         []DexTokenPair `json:"pairs"`
}
