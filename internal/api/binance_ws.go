package api

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"binance-terminal/internal/model"

	"github.com/gorilla/websocket"
)

const (
	binanceWSBaseURL = "wss://data-stream.binance.vision:9443/ws"
)

// MiniTickerEvent is the Binance WebSocket payload for miniTicker.
type MiniTickerEvent struct {
	EventType string `json:"e"` // "24hrMiniTicker"
	EventTime int64  `json:"E"`
	Symbol    string `json:"s"` // "BTCUSDT"
	Close     string `json:"c"` // Current price
	Open      string `json:"o"` // Open price 24h ago
	High      string `json:"h"` // High price 24h
	Low       string `json:"l"` // Low price 24h
	Volume    string `json:"v"` // Base asset volume
	QuoteVol  string `json:"q"` // Quote asset volume
}

// CombinedStreamPayload handles { "stream": "...", "data": {...} }
type CombinedStreamPayload struct {
	Stream string          `json:"stream"`
	Data   json.RawMessage `json:"data"`
}

// BinanceWSClient manages the WebSocket connection to Binance.
type BinanceWSClient struct {
	mu           sync.RWMutex
	conn         *websocket.Conn
	symbols      map[string]bool
	updatesChan  chan model.MarketData
	statusChan   chan bool
	stopChan     chan struct{}
	isConnected  bool
	reconnecting bool
}

// NewBinanceWSClient creates a new WebSocket client.
func NewBinanceWSClient(symbols []string) *BinanceWSClient {
	symbolMap := make(map[string]bool)
	for _, s := range symbols {
		symbolMap[strings.ToUpper(s)] = true
	}

	return &BinanceWSClient{
		symbols:     symbolMap,
		updatesChan: make(chan model.MarketData, 200),
		statusChan:  make(chan bool, 10),
		stopChan:    make(chan struct{}),
	}
}

// Updates returns the channel delivering real-time market data updates.
func (c *BinanceWSClient) Updates() <-chan model.MarketData {
	return c.updatesChan
}

// Status returns the channel delivering connection status (true = connected, false = disconnected).
func (c *BinanceWSClient) Status() <-chan bool {
	return c.statusChan
}

// Start initiates the WebSocket connection in a background goroutine.
func (c *BinanceWSClient) Start() {
	go c.run()
}

// UpdateSymbols updates the subscribed symbols list and sends SUBSCRIBE / UNSUBSCRIBE.
func (c *BinanceWSClient) UpdateSymbols(symbols []string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	newMap := make(map[string]bool)
	for _, s := range symbols {
		newMap[strings.ToUpper(s)] = true
	}
	c.symbols = newMap

	if c.conn != nil && c.isConnected {
		// Send subscription payload
		var streams []string
		for s := range newMap {
			streams = append(streams, strings.ToLower(s)+"@miniTicker")
		}
		subMsg := map[string]interface{}{
			"method": "SUBSCRIBE",
			"params": streams,
			"id":     time.Now().UnixNano(),
		}
		_ = c.conn.WriteJSON(subMsg)
	}
}

// Stop closes the WebSocket connection gracefully.
func (c *BinanceWSClient) Stop() {
	c.mu.Lock()
	defer c.mu.Unlock()

	select {
	case <-c.stopChan:
		return
	default:
		close(c.stopChan)
	}

	if c.conn != nil {
		_ = c.conn.Close()
	}
}

func (c *BinanceWSClient) run() {
	for {
		select {
		case <-c.stopChan:
			return
		default:
		}

		err := c.connectAndListen()
		if err != nil {
			c.setConnected(false)
			// Backoff before reconnecting
			select {
			case <-c.stopChan:
				return
			case <-time.After(2 * time.Second):
			}
		}
	}
}

func (c *BinanceWSClient) connectAndListen() error {
	c.mu.RLock()
	var streams []string
	for s := range c.symbols {
		streams = append(streams, strings.ToLower(s)+"@miniTicker")
	}
	c.mu.RUnlock()

	var wsURL string
	if len(streams) > 0 {
		wsURL = fmt.Sprintf("wss://data-stream.binance.vision:9443/stream?streams=%s", strings.Join(streams, "/"))
	} else {
		wsURL = binanceWSBaseURL
	}

	dialer := websocket.Dialer{
		HandshakeTimeout: 5 * time.Second,
		Proxy:            http.ProxyFromEnvironment,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	conn, _, err := dialer.Dial(wsURL, nil)
	if err != nil {
		return err
	}

	c.mu.Lock()
	c.conn = conn
	c.mu.Unlock()
	c.setConnected(true)

	defer conn.Close()

	for {
		select {
		case <-c.stopChan:
			return nil
		default:
		}

		_, message, err := conn.ReadMessage()
		if err != nil {
			return err
		}

		c.parseAndEmit(message)
	}
}

func (c *BinanceWSClient) parseAndEmit(msg []byte) {
	// Try combined stream first
	var combined CombinedStreamPayload
	if err := json.Unmarshal(msg, &combined); err == nil && len(combined.Data) > 0 {
		var ticker MiniTickerEvent
		if err := json.Unmarshal(combined.Data, &ticker); err == nil && ticker.Symbol != "" {
			c.emitTicker(ticker)
			return
		}
	}

	// Try direct miniTicker event
	var ticker MiniTickerEvent
	if err := json.Unmarshal(msg, &ticker); err == nil && ticker.Symbol != "" {
		c.emitTicker(ticker)
		return
	}
}

func (c *BinanceWSClient) emitTicker(t MiniTickerEvent) {
	price, _ := strconv.ParseFloat(t.Close, 64)
	open, _ := strconv.ParseFloat(t.Open, 64)
	high, _ := strconv.ParseFloat(t.High, 64)
	low, _ := strconv.ParseFloat(t.Low, 64)
	quoteVol, _ := strconv.ParseFloat(t.QuoteVol, 64)

	var changePct float64
	var changeAmt float64
	if open > 0 {
		changeAmt = price - open
		changePct = (changeAmt / open) * 100.0
	}

	displaySymbol := strings.TrimSuffix(t.Symbol, "USDT")
	if displaySymbol == "" {
		displaySymbol = t.Symbol
	}

	marketData := model.MarketData{
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
		LastUpdated:       time.Now(),
	}

	select {
	case c.updatesChan <- marketData:
	default:
		// Drop frame if receiver is busy to keep latency near zero
	}
}

func (c *BinanceWSClient) setConnected(connected bool) {
	c.mu.Lock()
	c.isConnected = connected
	c.mu.Unlock()

	select {
	case c.statusChan <- connected:
	default:
	}
}
