package ai_test

import (
	"testing"

	"github.com/fchyoga/cryptoterm/internal/ai"
	"github.com/fchyoga/cryptoterm/internal/model"
)

func TestResolveProviderConfig(t *testing.T) {
	tests := []struct {
		provider     string
		customURL    string
		customModel  string
		wantURL      string
		wantModel    string
	}{
		{
			provider:  "gemini",
			wantURL:   ai.GeminiBaseURL,
			wantModel: ai.GeminiDefaultModel,
		},
		{
			provider:  "openrouter",
			wantURL:   ai.OpenRouterBaseURL,
			wantModel: ai.OpenRouterDefaultModel,
		},
		{
			provider:  "deepseek",
			wantURL:   ai.DeepSeekBaseURL,
			wantModel: ai.DeepSeekDefaultModel,
		},
		{
			provider:  "groq",
			wantURL:   ai.GroqBaseURL,
			wantModel: ai.GroqDefaultModel,
		},
		{
			provider:    "custom",
			customURL:   "https://my-router.ai/v1",
			customModel: "my-custom-model",
			wantURL:     "https://my-router.ai/v1",
			wantModel:   "my-custom-model",
		},
	}

	for _, tt := range tests {
		cfg := &model.Config{
			AIProvider: tt.provider,
			AIBaseURL:  tt.customURL,
			AIModel:    tt.customModel,
		}
		gotURL, gotModel := ai.ResolveProviderConfig(cfg)
		if gotURL != tt.wantURL {
			t.Errorf("For provider %s, got URL %s, want %s", tt.provider, gotURL, tt.wantURL)
		}
		if gotModel != tt.wantModel {
			t.Errorf("For provider %s, got Model %s, want %s", tt.provider, gotModel, tt.wantModel)
		}
	}
}
