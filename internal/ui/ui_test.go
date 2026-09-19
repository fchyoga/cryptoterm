package ui_test

import (
	"strings"
	"testing"
	"time"

	"github.com/fchyoga/cryptoterm/internal/config"
	"github.com/fchyoga/cryptoterm/internal/model"
	"github.com/fchyoga/cryptoterm/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
)

func TestUIModelRender(t *testing.T) {
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	app := ui.NewUIModel(cfg)
	app.Width = 120
	app.Height = 35

	// 1. Check initial view rendering
	viewStr := app.View()
	if !strings.Contains(viewStr, "CRYPTO TERMINAL") {
		t.Errorf("Expected view to contain title 'CRYPTO TERMINAL'")
	}
	if !strings.Contains(viewStr, "Watchlist") {
		t.Errorf("Expected view to contain 'Watchlist'")
	}

	// 2. Simulate incoming market data tick for BTC
	tickMsg := ui.MarketDataMsg(model.MarketData{
		Symbol:         "BTCUSDT",
		DisplaySymbol:  "BTC",
		Source:         model.SourceBinance,
		Category:       model.TypeMajor,
		Price:          78500.25,
		PriceChange24h: 3.45,
		High24h:        79000.00,
		Low24h:         76000.00,
		Volume24h:      1250000000.00,
		PriceHistory:   []float64{77000, 77500, 78000, 78500.25},
		LastUpdated:    time.Now(),
	})

	updatedModel, _ := app.Update(tickMsg)
	app = updatedModel.(ui.UIModel)

	// Verify price is reflected in rendered view
	viewAfterTick := app.View()
	if !strings.Contains(viewAfterTick, "78,500.25") {
		t.Errorf("Expected rendered view to contain updated price '78,500.25'")
	}

	// 3. Test Tab Navigation
	// Switch to Meme Radar (Tab 2)
	updatedModel, _ = app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
	app = updatedModel.(ui.UIModel)
	if app.ActiveTab != ui.ViewMemeRadar {
		t.Errorf("Expected active tab to be MemeRadar, got %v", app.ActiveTab)
	}

	// Switch to Portfolio (Tab 3)
	updatedModel, _ = app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'3'}})
	app = updatedModel.(ui.UIModel)
	if app.ActiveTab != ui.ViewPortfolio {
		t.Errorf("Expected active tab to be Portfolio, got %v", app.ActiveTab)
	}
	portfolioView := app.View()
	if !strings.Contains(portfolioView, "Portfolio") {
		t.Errorf("Expected portfolio view to contain 'Portfolio'")
	}

	// 4. Test Sparkline generation
	history := []float64{10, 20, 15, 25, 30, 28, 35}
	spark := ui.GenerateSparkline(history, 10)
	if len(spark) == 0 {
		t.Errorf("Expected non-empty sparkline")
	}

	// 5. Test Price formatting
	p1 := ui.FormatPrice(65432.10, "USD", 16250)
	if p1 != "$65,432.10" {
		t.Errorf("Expected '$65,432.10', got %s", p1)
	}

	pMeme := ui.FormatPrice(0.00004523, "USD", 16250)
	if pMeme != "$0.00004523" {
		t.Errorf("Expected '$0.00004523', got %s", pMeme)
	}
}

func TestAIModalRender(t *testing.T) {
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	// Ensure no AI key for initial test
	cfg.AIAPIKey = ""

	app := ui.NewUIModel(cfg)
	app.Width = 120
	app.Height = 35

	// Seed a token market data
	app.MarketData["BTCUSDT"] = model.MarketData{
		Symbol:        "BTCUSDT",
		DisplaySymbol: "BTC",
		Price:         78000,
	}

	// Press 'x' to trigger AI Copilot
	updatedModel, _ := app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	app = updatedModel.(ui.UIModel)

	if !app.ShowAIModal {
		t.Fatalf("Expected ShowAIModal to be true after pressing 'x'")
	}

	viewNoKey := app.View()
	if !strings.Contains(viewNoKey, "AI MARKET SIGNAL ANALYZER") {
		t.Errorf("Expected view to show AI analyzer title")
	}
	if !strings.Contains(viewNoKey, "Bring Your Own Key") {
		t.Errorf("Expected view to show BYOK guidance when key is missing")
	}

	// Simulate successful AI signal arrival
	signalMsg := ui.AISignalResultMsg{
		Signal: &model.AISignal{
			Symbol:        "BTCUSDT",
			DisplaySymbol: "BTC",
			Action:        "STRONG_BUY",
			Confidence:    9.0,
			RiskLevel:     "LOW",
			EntryZone:     "$77,500 - $78,200",
			TakeProfit1:   82000,
			TakeProfit2:   85000,
			StopLoss:      75500,
			RiskReward:    "1 : 2.8",
			Reasoning:     []string{"Bullish continuation on 1h timeframe", "RSI neutral accumulation"},
			LatencyMs:     350,
		},
		Err: nil,
	}

	updatedModel, _ = app.Update(signalMsg)
	app = updatedModel.(ui.UIModel)

	viewSuccess := app.View()
	if !strings.Contains(viewSuccess, "STRONG BUY") {
		t.Errorf("Expected view to contain 'STRONG BUY'")
	}
	if !strings.Contains(viewSuccess, "77,500") {
		t.Errorf("Expected view to contain entry zone")
	}

	// Press Esc to close
	updatedModel, _ = app.Update(tea.KeyMsg{Type: tea.KeyEsc})
	app = updatedModel.(ui.UIModel)
	if app.ShowAIModal {
		t.Errorf("Expected ShowAIModal to be false after pressing Esc")
	}
}
