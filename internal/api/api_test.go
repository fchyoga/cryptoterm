package api_test

import (
	"testing"

	"github.com/fchyoga/cryptoterm/internal/api"
)

func TestBinanceFetch24hrTickers(t *testing.T) {
	client := api.NewBinanceRESTClient("", "")
	symbols := []string{"BTCUSDT", "ETHUSDT"}

	data, err := client.Fetch24hrTickers(symbols)
	if err != nil {
		t.Fatalf("Fetch24hrTickers failed: %v", err)
	}

	if len(data) == 0 {
		t.Fatalf("Expected ticker data, got 0 items")
	}

	btc, ok := data["BTCUSDT"]
	if !ok {
		t.Fatalf("Expected BTCUSDT in response")
	}

	if btc.Price <= 0 {
		t.Errorf("Expected BTC price > 0, got %f", btc.Price)
	}
}

func TestDexScreenerSearch(t *testing.T) {
	client := api.NewDexClient()
	pairs, err := client.Search("POPCAT")
	if err != nil {
		t.Fatalf("DexScreener Search failed: %v", err)
	}

	if len(pairs) == 0 {
		t.Logf("Warning: no pairs returned for POPCAT (might be rate-limited or offline)")
		return
	}

	pair := pairs[0]
	if pair.BaseToken.Symbol == "" {
		t.Errorf("Expected base token symbol to be non-empty")
	}
}
