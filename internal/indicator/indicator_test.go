package indicator_test

import (
	"math"
	"testing"

	"github.com/fchyoga/cryptoterm/internal/indicator"
	"github.com/fchyoga/cryptoterm/internal/model"
)

func TestCalculateRSI(t *testing.T) {
	// 15 ascending prices -> RSI should be high (> 70)
	var ascending []float64
	for i := 1; i <= 20; i++ {
		ascending = append(ascending, float64(i*10))
	}
	rsiHigh := indicator.CalculateRSI(ascending, 14)
	if rsiHigh < 70 {
		t.Errorf("Expected overbought RSI > 70 for pure ascending prices, got %f", rsiHigh)
	}

	// 15 descending prices -> RSI should be low (< 30)
	var descending []float64
	for i := 20; i >= 1; i-- {
		descending = append(descending, float64(i*10))
	}
	rsiLow := indicator.CalculateRSI(descending, 14)
	if rsiLow > 30 {
		t.Errorf("Expected oversold RSI < 30 for pure descending prices, got %f", rsiLow)
	}

	// Insufficient data -> neutral 50
	short := []float64{10, 20}
	rsiNeutral := indicator.CalculateRSI(short, 14)
	if math.Abs(rsiNeutral-50.0) > 0.01 {
		t.Errorf("Expected 50.0 for short data, got %f", rsiNeutral)
	}
}

func TestCalculateEMA(t *testing.T) {
	prices := []float64{10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}
	ema := indicator.CalculateEMA(prices, 5)
	if ema <= 0 {
		t.Errorf("Expected positive EMA, got %f", ema)
	}
	if ema < 10 || ema > 20 {
		t.Errorf("Expected EMA to be within price bounds (10-20), got %f", ema)
	}
}

func TestAnalyzeTechnicals(t *testing.T) {
	var klines []model.Kline
	for i := 1; i <= 30; i++ {
		price := float64(100 + i)
		klines = append(klines, model.Kline{
			Open:   price - 1,
			High:   price + 2,
			Low:    price - 2,
			Close:  price,
			Volume: 1000,
		})
	}

	summary := indicator.AnalyzeTechnicals(135.0, klines)
	if summary.RSI14 <= 0 {
		t.Errorf("Expected positive RSI, got %f", summary.RSI14)
	}
	if summary.Support <= 0 || summary.Resistance <= 0 {
		t.Errorf("Expected valid support and resistance levels")
	}
}
