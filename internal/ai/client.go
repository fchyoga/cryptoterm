package ai

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/fchyoga/cryptoterm/internal/indicator"
	"github.com/fchyoga/cryptoterm/internal/model"
)

// Provider defaults and endpoint definitions
const (
	GeminiBaseURL     = "https://generativelanguage.googleapis.com/v1beta/openai"
	GeminiDefaultModel = "gemini-2.0-flash"

	OpenRouterBaseURL     = "https://openrouter.ai/api/v1"
	OpenRouterDefaultModel = "deepseek/deepseek-chat"

	DeepSeekBaseURL     = "https://api.deepseek.com/v1"
	DeepSeekDefaultModel = "deepseek-chat"

	GroqBaseURL     = "https://api.groq.com/openai/v1"
	GroqDefaultModel = "llama-3.3-70b-versatile"

	OpenAIBaseURL     = "https://api.openai.com/v1"
	OpenAIDefaultModel = "gpt-4o-mini"

	AnthropicBaseURL     = "https://api.anthropic.com/v1/messages"
	AnthropicDefaultModel = "claude-3-5-sonnet-20241022"

	OllamaBaseURL     = "http://localhost:11434/v1"
	OllamaDefaultModel = "llama3.2"
)

// ResolveProviderConfig returns the effective base URL and model for the configured provider.
func ResolveProviderConfig(cfg *model.Config) (baseURL, modelName string) {
	provider := strings.ToLower(strings.TrimSpace(cfg.AIProvider))
	customURL := strings.TrimSpace(cfg.AIBaseURL)
	customModel := strings.TrimSpace(cfg.AIModel)

	switch provider {
	case "gemini", "google":
		baseURL = GeminiBaseURL
		modelName = GeminiDefaultModel
	case "openrouter", "9router":
		baseURL = OpenRouterBaseURL
		modelName = OpenRouterDefaultModel
	case "deepseek":
		baseURL = DeepSeekBaseURL
		modelName = DeepSeekDefaultModel
	case "groq":
		baseURL = GroqBaseURL
		modelName = GroqDefaultModel
	case "openai":
		baseURL = OpenAIBaseURL
		modelName = OpenAIDefaultModel
	case "anthropic", "claude":
		baseURL = AnthropicBaseURL
		modelName = AnthropicDefaultModel
	case "ollama":
		baseURL = OllamaBaseURL
		modelName = OllamaDefaultModel
	default:
		baseURL = GeminiBaseURL
		modelName = GeminiDefaultModel
	}

	if customURL != "" {
		baseURL = strings.TrimSuffix(customURL, "/")
	}
	if customModel != "" {
		modelName = customModel
	}

	return baseURL, modelName
}

// OpenAIChatMessage represents a message in OpenAI chat completions.
type OpenAIChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// OpenAIChatRequest represents the payload sent to OpenAI-compatible endpoints.
type OpenAIChatRequest struct {
	Model       string              `json:"model"`
	Messages    []OpenAIChatMessage `json:"messages"`
	Temperature float64             `json:"temperature"`
}

// OpenAIChatResponse represents the response from OpenAI-compatible endpoints.
type OpenAIChatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error,omitempty"`
}

// GenerateSignal queries the configured LLM provider to produce an actionable crypto signal.
func GenerateSignal(
	ctx context.Context,
	cfg *model.Config,
	token model.WatchItem,
	market model.MarketData,
	tech indicator.TechnicalSummary,
) (*model.AISignal, error) {
	if strings.TrimSpace(cfg.AIAPIKey) == "" && strings.ToLower(cfg.AIProvider) != "ollama" {
		return nil, fmt.Errorf("AI API Key not configured. Press [c] to set up your provider API key")
	}

	baseURL, modelName := ResolveProviderConfig(cfg)
	providerName := strings.ToUpper(cfg.AIProvider)
	if providerName == "" {
		providerName = "GEMINI"
	}

	prompt := buildPrompt(token, market, tech, cfg.AITradingStyle)

	startTime := time.Now()
	var rawResponse string
	var err error

	if strings.ToLower(cfg.AIProvider) == "anthropic" || strings.ToLower(cfg.AIProvider) == "claude" {
		rawResponse, err = callAnthropic(ctx, cfg.AIAPIKey, baseURL, modelName, prompt)
	} else {
		rawResponse, err = callOpenAICompatible(ctx, cfg.AIAPIKey, baseURL, modelName, prompt)
	}

	if err != nil {
		return nil, err
	}

	latency := time.Since(startTime).Milliseconds()

	// Parse structured JSON response
	signal, err := parseSignalJSON(rawResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to parse AI signal response: %w", err)
	}

	signal.Symbol = token.Symbol
	signal.DisplaySymbol = token.DisplaySymbol
	signal.ModelUsed = modelName
	signal.ProviderUsed = providerName
	signal.LatencyMs = latency
	signal.CreatedAt = time.Now()

	return signal, nil
}

func buildPrompt(
	token model.WatchItem,
	market model.MarketData,
	tech indicator.TechnicalSummary,
	tradingStyle string,
) string {
	if tradingStyle == "" {
		tradingStyle = "daytrader"
	}

	var sb strings.Builder
	sb.WriteString("You are cryptoterm AI Copilot, an institutional-grade quantitative crypto technical analyst.\n")
	sb.WriteString(fmt.Sprintf("Persona Profile: %s (Focus on high-probability setups and disciplined risk management).\n\n", strings.ToUpper(tradingStyle)))

	sb.WriteString("MARKET CONTEXT:\n")
	sb.WriteString(fmt.Sprintf("• Asset: %s (%s)\n", token.DisplaySymbol, token.Name))
	sb.WriteString(fmt.Sprintf("• Category: %s | Source: %s | Chain: %s\n", token.Category, token.Source, strings.ToUpper(token.Chain)))
	sb.WriteString(fmt.Sprintf("• Current Price: $%.8g\n", market.Price))
	sb.WriteString(fmt.Sprintf("• 24h High: $%.8g | 24h Low: $%.8g\n", market.High24h, market.Low24h))
	sb.WriteString(fmt.Sprintf("• 24h Price Change: %+.2f%%\n", market.PriceChange24h))
	sb.WriteString(fmt.Sprintf("• 24h Volume: $%.2f\n", market.Volume24h))

	if token.Source == model.SourceDEX {
		sb.WriteString(fmt.Sprintf("• On-Chain Liquidity: $%.2f\n", market.LiquidityUSD))
		sb.WriteString(fmt.Sprintf("• FDV / Market Cap: $%.2f\n", market.MarketCapUSD))
	}

	sb.WriteString("\nCALCULATED TECHNICAL INDICATORS:\n")
	sb.WriteString(fmt.Sprintf("• RSI (14): %.2f\n", tech.RSI14))
	sb.WriteString(fmt.Sprintf("• EMA (20): $%.8g\n", tech.EMA20))
	sb.WriteString(fmt.Sprintf("• EMA (50): $%.8g\n", tech.EMA50))
	sb.WriteString(fmt.Sprintf("• Local Key Support: $%.8g\n", tech.Support))
	sb.WriteString(fmt.Sprintf("• Local Key Resistance: $%.8g\n", tech.Resistance))
	sb.WriteString(fmt.Sprintf("• Baseline Trend: %s\n\n", tech.Trend))

	sb.WriteString(`Analyze the technical confluence and return a STRICT JSON object only. Do NOT include markdown code blocks, backticks, or explanatory text before or after the JSON.

Expected JSON schema:
{
  "action": "STRONG_BUY" | "BUY" | "NEUTRAL" | "SELL" | "STRONG_SELL",
  "confidence": 8.5,
  "risk_level": "LOW" | "MEDIUM" | "HIGH",
  "entry_zone": "$105.00 - $106.20",
  "take_profit_1": 110.50,
  "take_profit_2": 115.00,
  "stop_loss": 103.80,
  "risk_reward": "1 : 2.5",
  "reasoning": [
    "RSI(14) at 48.2 indicates healthy accumulation before momentum continuation.",
    "Price is holding comfortably above the 20-period EMA on the 1-hour timeframe.",
    "24-hour volume expanded +18% into local resistance, signaling aggressive buying pressure."
  ],
  "invalidation": "Hourly candle close below the designated stop loss level.",
  "technical_summary": "Bullish momentum continuation setup supported by EMA golden cross and healthy RSI."
}`)

	return sb.String()
}

func callOpenAICompatible(
	ctx context.Context,
	apiKey, baseURL, modelName, prompt string,
) (string, error) {
	reqBody := OpenAIChatRequest{
		Model: modelName,
		Messages: []OpenAIChatMessage{
			{
				Role:    "system",
				Content: "You are a disciplined quantitative crypto analyst. You output raw JSON only.",
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Temperature: 0.2,
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	endpoint := baseURL
	if !strings.HasSuffix(endpoint, "/chat/completions") {
		endpoint = strings.TrimSuffix(endpoint, "/") + "/chat/completions"
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewBuffer(data))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}

	// Custom headers for OpenRouter
	if strings.Contains(baseURL, "openrouter.ai") {
		req.Header.Set("HTTP-Referer", "https://github.com/fchyoga/cryptoterm")
		req.Header.Set("X-Title", "cryptoterm")
	}

	tr := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	client := &http.Client{
		Timeout:   25 * time.Second,
		Transport: tr,
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("AI request failed: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		var errResp OpenAIChatResponse
		_ = json.Unmarshal(bodyBytes, &errResp)
		if errResp.Error != nil && errResp.Error.Message != "" {
			return "", fmt.Errorf("%s API error: %s", resp.Status, errResp.Error.Message)
		}
		return "", fmt.Errorf("AI provider returned HTTP %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var chatResp OpenAIChatResponse
	if err := json.Unmarshal(bodyBytes, &chatResp); err != nil {
		return "", fmt.Errorf("failed to decode AI response: %w", err)
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("AI returned empty choices")
	}

	return chatResp.Choices[0].Message.Content, nil
}

func callAnthropic(
	ctx context.Context,
	apiKey, baseURL, modelName, prompt string,
) (string, error) {
	reqPayload := map[string]interface{}{
		"model":      modelName,
		"max_tokens": 1024,
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": prompt,
			},
		},
	}

	data, err := json.Marshal(reqPayload)
	if err != nil {
		return "", err
	}

	endpoint := baseURL
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewBuffer(data))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	tr := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	client := &http.Client{
		Timeout:   25 * time.Second,
		Transport: tr,
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("Anthropic request failed: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Anthropic HTTP %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var anthropicResp struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.Unmarshal(bodyBytes, &anthropicResp); err != nil {
		return "", err
	}

	if len(anthropicResp.Content) == 0 {
		return "", fmt.Errorf("Anthropic returned empty content")
	}

	return anthropicResp.Content[0].Text, nil
}

// parseSignalJSON cleans up any markdown code fencing and unmarshals into AISignal.
func parseSignalJSON(raw string) (*model.AISignal, error) {
	clean := strings.TrimSpace(raw)

	// Strip ```json ... ``` or ``` ... ``` if present
	re := regexp.MustCompile("(?s)```(?:json)?\\s*(.*?)\\s*```")
	matches := re.FindStringSubmatch(clean)
	if len(matches) > 1 {
		clean = strings.TrimSpace(matches[1])
	}

	var sig model.AISignal
	if err := json.Unmarshal([]byte(clean), &sig); err != nil {
		// Fallback: extract substring between first { and last }
		start := strings.Index(clean, "{")
		end := strings.LastIndex(clean, "}")
		if start != -1 && end != -1 && end > start {
			extracted := clean[start : end+1]
			if errFallback := json.Unmarshal([]byte(extracted), &sig); errFallback == nil {
				return &sig, nil
			}
		}
		return nil, fmt.Errorf("%w (raw response: %s)", err, clean)
	}

	return &sig, nil
}
