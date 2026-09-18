package api

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/fchyoga/cryptoterm/internal/model"
)
const (
	binancePublicRESTURL = "https://data-api.binance.vision"
	binanceAuthRESTURL   = "https://api.binance.com"
)

// BinanceRESTClient handles HTTP requests to Binance.
type BinanceRESTClient struct {
	apiKey    string
	apiSecret string
	client    *http.Client
}

func NewBinanceRESTClient(apiKey, apiSecret string) *BinanceRESTClient {
	tr := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	return &BinanceRESTClient{
		apiKey:    apiKey,
		apiSecret: apiSecret,
		client: &http.Client{
			Timeout:   8 * time.Second,
			Transport: tr,
		},
	}
}

// SetCredentials updates the API key and secret.
func (c *BinanceRESTClient) SetCredentials(apiKey, apiSecret string) {
	c.apiKey = apiKey
	c.apiSecret = apiSecret
}

// HasCredentials returns true if both key and secret are configured.
func (c *BinanceRESTClient) HasCredentials() bool {
	return c.apiKey != "" && c.apiSecret != ""
}

// Ticker24hrResponse is the JSON structure from /api/v3/ticker/24hr
type Ticker24hrResponse struct {
	Symbol             string `json:"symbol"`
	PriceChange        string `json:"priceChange"`
	PriceChangePercent string `json:"priceChangePercent"`
	LastPrice          string `json:"lastPrice"`
	HighPrice          string `json:"highPrice"`
	LowPrice           string `json:"lowPrice"`
	Volume             string `json:"volume"`
	QuoteVolume        string `json:"quoteVolume"`
}

// Fetch24hrTickers retrieves 24hr ticker data for the specified symbols.
func (c *BinanceRESTClient) Fetch24hrTickers(symbols []string) (map[string]model.MarketData, error) {
	result := make(map[string]model.MarketData)
	if len(symbols) == 0 {
		return result, nil
	}

	// Prepare symbols JSON query
	var quoted []string
	for _, s := range symbols {
		quoted = append(quoted, fmt.Sprintf("%q", strings.ToUpper(s)))
	}
	symbolsParam := url.QueryEscape(fmt.Sprintf("[%s]", strings.Join(quoted, ",")))
	reqURL := fmt.Sprintf("%s/api/v3/ticker/24hr?symbols=%s", binancePublicRESTURL, symbolsParam)

	resp, err := c.client.Get(reqURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("binance API returned status %d", resp.StatusCode)
	}

	var tickers []Ticker24hrResponse
	if err := json.NewDecoder(resp.Body).Decode(&tickers); err != nil {
		return nil, err
	}

	for _, t := range tickers {
		price, _ := strconv.ParseFloat(t.LastPrice, 64)
		changePct, _ := strconv.ParseFloat(t.PriceChangePercent, 64)
		changeAmt, _ := strconv.ParseFloat(t.PriceChange, 64)
		high, _ := strconv.ParseFloat(t.HighPrice, 64)
		low, _ := strconv.ParseFloat(t.LowPrice, 64)
		quoteVol, _ := strconv.ParseFloat(t.QuoteVolume, 64)

		displaySymbol := strings.TrimSuffix(t.Symbol, "USDT")
		if displaySymbol == "" {
			displaySymbol = t.Symbol
		}

		result[t.Symbol] = model.MarketData{
			Symbol:            t.Symbol,
			DisplaySymbol:     displaySymbol,
			Source:            model.SourceBinance,
			Chain:             "binance",
			Price:             price,
			PriceChange24h:    changePct,
			PriceChangeAmount: changeAmt,
			High24h:           high,
			Low24h:            low,
			Volume24h:         quoteVol,
			PriceHistory:      []float64{price},
			LastUpdated:       time.Now(),
		}
	}

	return result, nil
}

// ValidateSymbol checks whether a symbol exists on Binance.
func (c *BinanceRESTClient) ValidateSymbol(symbol string) (string, bool) {
	upper := strings.ToUpper(strings.TrimSpace(symbol))
	if !strings.HasSuffix(upper, "USDT") {
		upper = upper + "USDT"
	}

	url := fmt.Sprintf("%s/api/v3/ticker/price?symbol=%s", binancePublicRESTURL, upper)
	resp, err := c.client.Get(url)
	if err != nil || resp.StatusCode != http.StatusOK {
		return "", false
	}
	defer resp.Body.Close()
	return upper, true
}

// AccountBalance represents a holding from Binance /api/v3/account
type AccountBalance struct {
	Asset  string `json:"asset"`
	Free   string `json:"free"`
	Locked string `json:"locked"`
}

type AccountResponse struct {
	AccountType string           `json:"accountType"`
	Balances    []AccountBalance `json:"balances"`
}

// FetchPortfolio queries the user's non-zero balances using the configured API credentials.
func (c *BinanceRESTClient) FetchPortfolio(priceMap map[string]float64) ([]model.PortfolioItem, error) {
	if !c.HasCredentials() {
		return nil, fmt.Errorf("binance API Key & Secret not configured")
	}

	timestamp := time.Now().UnixMilli()
	queryString := fmt.Sprintf("recvWindow=5000&timestamp=%d", timestamp)

	// Generate HMAC-SHA256 signature
	mac := hmac.New(sha256.New, []byte(c.apiSecret))
	mac.Write([]byte(queryString))
	signature := hex.EncodeToString(mac.Sum(nil))

	fullURL := fmt.Sprintf("%s/api/v3/account?%s&signature=%s", binanceAuthRESTURL, queryString, signature)

	req, err := http.NewRequest(http.MethodGet, fullURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-MBX-APIKEY", c.apiKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp struct {
			Code int    `json:"code"`
			Msg  string `json:"msg"`
		}
		_ = json.NewDecoder(resp.Body).Decode(&errResp)
		if errResp.Msg != "" {
			return nil, fmt.Errorf("binance error (%d): %s", errResp.Code, errResp.Msg)
		}
		return nil, fmt.Errorf("failed to fetch account, HTTP status: %d", resp.StatusCode)
	}

	var acc AccountResponse
	if err := json.NewDecoder(resp.Body).Decode(&acc); err != nil {
		return nil, err
	}

	var items []model.PortfolioItem
	for _, b := range acc.Balances {
		free, _ := strconv.ParseFloat(b.Free, 64)
		locked, _ := strconv.ParseFloat(b.Locked, 64)
		total := free + locked

		if total > 0.000001 { // Filter out dust
			price := 0.0
			asset := strings.ToUpper(b.Asset)

			if asset == "USDT" || asset == "USD" || asset == "USDC" || asset == "FDUSD" {
				price = 1.0
			} else {
				// Lookup in price map
				pair := asset + "USDT"
				if p, ok := priceMap[pair]; ok {
					price = p
				} else if p, ok := priceMap[asset]; ok {
					price = p
				}
			}

			value := total * price

			// Only show if value > $0.05 or total significant
			if value > 0.05 || price == 0 {
				items = append(items, model.PortfolioItem{
					Asset:    asset,
					Free:     free,
					Locked:   locked,
					Total:    total,
					PriceUSD: price,
					ValueUSD: value,
				})
			}
		}
	}

	return items, nil
}
