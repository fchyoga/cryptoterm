package api

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"binance-terminal/internal/model"
)

const (
	dexScreenerBaseURL = "https://api.dexscreener.com/latest/dex"
)

// DexClient handles queries to DexScreener API (free, no API key required).
type DexClient struct {
	client *http.Client
}

// NewDexClient creates a new DexClient instance.
func NewDexClient() *DexClient {
	tr := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	return &DexClient{
		client: &http.Client{
			Timeout:   8 * time.Second,
			Transport: tr,
		},
	}
}

// FetchTokens retrieves market data for a comma-separated list of token contract addresses.
func (d *DexClient) FetchTokens(addresses []string) (map[string]model.MarketData, error) {
	result := make(map[string]model.MarketData)
	if len(addresses) == 0 {
		return result, nil
	}

	joined := strings.Join(addresses, ",")
	reqURL := fmt.Sprintf("%s/tokens/%s", dexScreenerBaseURL, joined)

	resp, err := d.client.Get(reqURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("dexscreener returned status %d", resp.StatusCode)
	}

	var searchResp model.DexSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return nil, err
	}

	for _, pair := range searchResp.Pairs {
		// Prefer the pair with highest liquidity for a given token address
		addr := pair.BaseToken.Address
		priceUSD, _ := strconv.ParseFloat(pair.PriceUsd, 64)

		existing, found := result[addr]
		if !found || pair.Liquidity.Usd > existing.LiquidityUSD {
			displaySymbol := pair.BaseToken.Symbol
			if displaySymbol == "" {
				displaySymbol = pair.BaseToken.Name
			}

			result[addr] = model.MarketData{
				Symbol:            addr,
				DisplaySymbol:     displaySymbol,
				Source:            model.SourceDEX,
				Category:          model.TypeMeme,
				Chain:             pair.ChainId,
				Price:             priceUSD,
				PriceChange24h:    pair.PriceChange.H24,
				PriceChangeAmount: 0,
				High24h:           0,
				Low24h:            0,
				Volume24h:         pair.Volume.H24,
				LiquidityUSD:      pair.Liquidity.Usd,
				MarketCapUSD:      pair.Fdv,
				PriceHistory:      []float64{priceUSD},
				LastUpdated:       time.Now(),
			}
		}
	}

	return result, nil
}

// Search searches for tokens or pairs by query string (e.g. "POPCAT", "WIF", or contract address).
func (d *DexClient) Search(query string) ([]model.DexTokenPair, error) {
	trimmed := strings.TrimSpace(query)
	if trimmed == "" {
		return nil, nil
	}

	reqURL := fmt.Sprintf("%s/search?q=%s", dexScreenerBaseURL, url.QueryEscape(trimmed))

	resp, err := d.client.Get(reqURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("dexscreener search returned status %d", resp.StatusCode)
	}

	var searchResp model.DexSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return nil, err
	}

	// Limit to top 15 results
	limit := 15
	if len(searchResp.Pairs) < limit {
		limit = len(searchResp.Pairs)
	}

	return searchResp.Pairs[:limit], nil
}

// FetchTrendingMemes gets a curated list of trending meme tokens.
func (d *DexClient) FetchTrendingMemes() ([]model.MarketData, error) {
	// Search for popular meme queries
	queries := []string{"SOL", "PEPE", "DOGE"}
	var results []model.MarketData
	seen := make(map[string]bool)

	for _, q := range queries {
		pairs, err := d.Search(q)
		if err != nil {
			continue
		}

		for _, pair := range pairs {
			addr := pair.BaseToken.Address
			if seen[addr] {
				continue
			}
			seen[addr] = true

			priceUSD, _ := strconv.ParseFloat(pair.PriceUsd, 64)
			// Filter out very low liquidity tokens
			if pair.Liquidity.Usd < 50000 {
				continue
			}

			results = append(results, model.MarketData{
				Symbol:            addr,
				DisplaySymbol:     pair.BaseToken.Symbol,
				Source:            model.SourceDEX,
				Category:          model.TypeMeme,
				Chain:             pair.ChainId,
				Price:             priceUSD,
				PriceChange24h:    pair.PriceChange.H24,
				Volume24h:         pair.Volume.H24,
				LiquidityUSD:      pair.Liquidity.Usd,
				MarketCapUSD:      pair.Fdv,
				PriceHistory:      []float64{priceUSD},
				LastUpdated:       time.Now(),
			})

			if len(results) >= 20 {
				break
			}
		}
		if len(results) >= 20 {
			break
		}
	}

	return results, nil
}
