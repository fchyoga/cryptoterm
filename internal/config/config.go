package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"binance-terminal/internal/model"
)

var (
	configMu sync.RWMutex
)

// DefaultWatchlist provides a pre-populated list covering Majors, Stables, and Memes.
func DefaultWatchlist() []model.WatchItem {
	now := time.Now()
	return []model.WatchItem{
		// Majors
		{
			Symbol:        "BTCUSDT",
			DisplaySymbol: "BTC",
			Name:          "Bitcoin",
			Source:        model.SourceBinance,
			Category:      model.TypeMajor,
			Chain:         "binance",
			AddedAt:       now,
		},
		{
			Symbol:        "ETHUSDT",
			DisplaySymbol: "ETH",
			Name:          "Ethereum",
			Source:        model.SourceBinance,
			Category:      model.TypeMajor,
			Chain:         "binance",
			AddedAt:       now,
		},
		{
			Symbol:        "SOLUSDT",
			DisplaySymbol: "SOL",
			Name:          "Solana",
			Source:        model.SourceBinance,
			Category:      model.TypeMajor,
			Chain:         "binance",
			AddedAt:       now,
		},
		{
			Symbol:        "BNBUSDT",
			DisplaySymbol: "BNB",
			Name:          "BNB Chain",
			Source:        model.SourceBinance,
			Category:      model.TypeMajor,
			Chain:         "binance",
			AddedAt:       now,
		},
		// Stables
		{
			Symbol:        "USDCUSDT",
			DisplaySymbol: "USDC",
			Name:          "USD Coin",
			Source:        model.SourceBinance,
			Category:      model.TypeStable,
			Chain:         "binance",
			AddedAt:       now,
		},
		{
			Symbol:        "FDUSDUSDT",
			DisplaySymbol: "FDUSD",
			Name:          "First Digital USD",
			Source:        model.SourceBinance,
			Category:      model.TypeStable,
			Chain:         "binance",
			AddedAt:       now,
		},
		// Memes
		{
			Symbol:        "DOGEUSDT",
			DisplaySymbol: "DOGE",
			Name:          "Dogecoin",
			Source:        model.SourceBinance,
			Category:      model.TypeMeme,
			Chain:         "binance",
			AddedAt:       now,
		},
		{
			Symbol:        "PEPEUSDT",
			DisplaySymbol: "PEPE",
			Name:          "Pepe",
			Source:        model.SourceBinance,
			Category:      model.TypeMeme,
			Chain:         "binance",
			AddedAt:       now,
		},
		{
			Symbol:        "SHIBUSDT",
			DisplaySymbol: "SHIB",
			Name:          "Shiba Inu",
			Source:        model.SourceBinance,
			Category:      model.TypeMeme,
			Chain:         "binance",
			AddedAt:       now,
		},
		{
			Symbol:        "WIFUSDT",
			DisplaySymbol: "WIF",
			Name:          "dogwifhat",
			Source:        model.SourceBinance,
			Category:      model.TypeMeme,
			Chain:         "binance",
			AddedAt:       now,
		},
		{
			Symbol:        "BONKUSDT",
			DisplaySymbol: "BONK",
			Name:          "Bonk",
			Source:        model.SourceBinance,
			Category:      model.TypeMeme,
			Chain:         "binance",
			AddedAt:       now,
		},
		// DEX Meme (Solana POPCAT via DexScreener pair)
		{
			Symbol:        "7GCihgDB8fe6KNjn2MYtkzZcRjQy3t9GHdC8uHYmW2hr",
			DisplaySymbol: "POPCAT",
			Name:          "Popcat (SOL)",
			Source:        model.SourceDEX,
			Category:      model.TypeMeme,
			Chain:         "solana",
			PairAddress:   "FRhB8L7Y9Qq41qZXYLtC2nw8WBWDD6duAE2h279bkqxK",
			AddedAt:       now,
		},
	}
}

// GetConfigDir returns the OS-standard configuration directory.
func GetConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".config", "binance-terminal")
	return dir, nil
}

// GetConfigFile returns the absolute path to config.json.
func GetConfigFile() (string, error) {
	dir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

// Load reads the config file or creates default if not present.
func Load() (*model.Config, error) {
	configMu.Lock()
	defer configMu.Unlock()

	configFile, err := GetConfigFile()
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		cfg := &model.Config{
			Currency:           "USD",
			USDToIDR:           16250.0,
			RefreshIntervalSec: 5,
			SoundAlerts:        true,
			Watchlist:          DefaultWatchlist(),
			Alerts:             []model.PriceAlert{},
		}
		if err := saveInternal(cfg, configFile); err != nil {
			return nil, fmt.Errorf("failed to save initial config: %w", err)
		}
		return cfg, nil
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg model.Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config JSON: %w", err)
	}

	if cfg.Currency == "" {
		cfg.Currency = "USD"
	}
	if cfg.USDToIDR == 0 {
		cfg.USDToIDR = 16250.0
	}
	if cfg.RefreshIntervalSec <= 0 {
		cfg.RefreshIntervalSec = 5
	}
	if len(cfg.Watchlist) == 0 {
		cfg.Watchlist = DefaultWatchlist()
	}

	return &cfg, nil
}

// Save persists the config to disk atomically.
func Save(cfg *model.Config) error {
	configMu.Lock()
	defer configMu.Unlock()

	configFile, err := GetConfigFile()
	if err != nil {
		return err
	}
	return saveInternal(cfg, configFile)
}

func saveInternal(cfg *model.Config, configFile string) error {
	dir := filepath.Dir(configFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	tempFile := configFile + ".tmp"
	if err := os.WriteFile(tempFile, data, 0600); err != nil {
		return err
	}

	return os.Rename(tempFile, configFile)
}
