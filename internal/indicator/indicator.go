package indicator

import (
	"github.com/fchyoga/cryptoterm/internal/model"
)

// CalculateRSI computes the relative strength index over a series of closing prices.
func CalculateRSI(closes []float64, period int) float64 {
	if len(closes) <= period {
		return 50.0 // Default neutral if insufficient samples
	}

	var gains, losses float64
	for i := 1; i <= period; i++ {
		change := closes[i] - closes[i-1]
		if change > 0 {
			gains += change
		} else {
			losses -= change
		}
	}

	avgGain := gains / float64(period)
	avgLoss := losses / float64(period)

	for i := period + 1; i < len(closes); i++ {
		change := closes[i] - closes[i-1]
		var currentGain, currentLoss float64
		if change > 0 {
			currentGain = change
		} else {
			currentLoss = -change
		}

		avgGain = (avgGain*float64(period-1) + currentGain) / float64(period)
		avgLoss = (avgLoss*float64(period-1) + currentLoss) / float64(period)
	}

	if avgLoss == 0 {
		return 100.0
	}

	rs := avgGain / avgLoss
	rsi := 100.0 - (100.0 / (1.0 + rs))
	return rsi
}

// CalculateEMA computes the Exponential Moving Average for a series of values.
func CalculateEMA(values []float64, period int) float64 {
	if len(values) == 0 {
		return 0
	}
	if len(values) < period {
		period = len(values)
	}

	// Initial SMA
	var sum float64
	for i := range period {
		sum += values[i]
	}
	ema := sum / float64(period)

	// Smoothing multiplier
	k := 2.0 / float64(period+1)
	for i := period; i < len(values); i++ {
		ema = (values[i] * k) + (ema * (1.0 - k))
	}

	return ema
}

// DetectKeyLevels identifies local support and resistance from recent candlesticks.
func DetectKeyLevels(klines []model.Kline) (support float64, resistance float64) {
	if len(klines) == 0 {
		return 0, 0
	}

	support = klines[0].Low
	resistance = klines[0].High

	for _, k := range klines {
		if k.Low < support {
			support = k.Low
		}
		if k.High > resistance {
			resistance = k.High
		}
	}

	return support, resistance
}

// TechnicalSummary holds calculated indicators for LLM context.
type TechnicalSummary struct {
	CurrentPrice float64
	RSI14        float64
	EMA20        float64
	EMA50        float64
	Support      float64
	Resistance   float64
	Trend        string
}

// AnalyzeTechnicals computes indicators from klines and provides a baseline trend.
func AnalyzeTechnicals(currentPrice float64, klines []model.Kline) TechnicalSummary {
	var closes []float64
	for _, k := range klines {
		closes = append(closes, k.Close)
	}

	rsi := CalculateRSI(closes, 14)
	ema20 := CalculateEMA(closes, 20)
	ema50 := CalculateEMA(closes, 50)
	sup, res := DetectKeyLevels(klines)

	trend := "NEUTRAL / SIDEWAYS"
	if currentPrice > ema20 && ema20 > ema50 {
		trend = "STRONG BULLISH (Price > EMA20 > EMA50)"
	} else if currentPrice > ema20 && ema20 <= ema50 {
		trend = "BULLISH REVERSAL (Price above EMA20)"
	} else if currentPrice < ema20 && ema20 < ema50 {
		trend = "STRONG BEARISH (Price < EMA20 < EMA50)"
	} else if currentPrice < ema20 && ema20 >= ema50 {
		trend = "BEARISH CORRECTION (Price below EMA20)"
	}

	return TechnicalSummary{
		CurrentPrice: currentPrice,
		RSI14:        rsi,
		EMA20:        ema20,
		EMA50:        ema50,
		Support:      sup,
		Resistance:   res,
		Trend:        trend,
	}
}
