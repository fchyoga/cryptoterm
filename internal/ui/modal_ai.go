package ui

import (
	"fmt"
	"strings"

	"github.com/fchyoga/cryptoterm/internal/ai"
	"github.com/fchyoga/cryptoterm/internal/model"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// AIModalState represents the current display state of the AI signal modal.
type AIModalState int

const (
	AIStateLoading AIModalState = iota
	AIStateSuccess
	AIStateNoKey
	AIStateError
)

// AIModal handles rendering and events for the AI Copilot signal analyzer.
type AIModal struct {
	Token        model.WatchItem
	Market       model.MarketData
	State        AIModalState
	Signal       *model.AISignal
	ErrorMessage string
	SpinnerFrame int
	Provider     string
	ModelName    string
}

var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// NewAIModal initializes the AI modal for a specific asset.
func NewAIModal(cfg *model.Config, token model.WatchItem, market model.MarketData) AIModal {
	_, modelName := ai.ResolveProviderConfig(cfg)
	provider := strings.ToUpper(cfg.AIProvider)
	if provider == "" {
		provider = "GEMINI"
	}

	state := AIStateLoading
	if strings.TrimSpace(cfg.AIAPIKey) == "" && strings.ToLower(cfg.AIProvider) != "ollama" {
		state = AIStateNoKey
	}

	return AIModal{
		Token:     token,
		Market:    market,
		State:     state,
		Provider:  provider,
		ModelName: modelName,
	}
}

// Update processes key events inside the AI signal modal.
func (m *AIModal) Update(msg tea.Msg) (closeModal bool, triggerRetry bool, openConfig bool) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "q":
			return true, false, false
		case "r":
			if m.State == AIStateSuccess || m.State == AIStateError {
				m.State = AIStateLoading
				return false, true, false
			}
		case "c":
			return true, false, true
		}
	}
	return false, false, false
}

// NextSpinner advances the loading animation frame.
func (m *AIModal) NextSpinner() {
	m.SpinnerFrame = (m.SpinnerFrame + 1) % len(spinnerFrames)
}

// View renders the AI signal modal.
func (m AIModal) View(width int) string {
	boxWidth := min(width-6, 84)
	if boxWidth < 50 {
		boxWidth = 50
	}

	cardStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorBrandAccent).
		Background(ColorBgCard).
		Padding(1, 3).
		Width(boxWidth)

	var b strings.Builder

	// Header banner
	headerTitle := lipgloss.NewStyle().Bold(true).Foreground(ColorBrandAccent).Render(
		fmt.Sprintf("🤖 AI MARKET SIGNAL ANALYZER — %s (%s)", m.Token.DisplaySymbol, m.Token.Name),
	)
	metaInfo := lipgloss.NewStyle().Foreground(lipgloss.Color("#6C7086")).Render(
		fmt.Sprintf("Provider: %s | Model: %s", m.Provider, m.ModelName),
	)
	b.WriteString(headerTitle + "\n" + metaInfo + "\n\n")

	switch m.State {
	case AIStateNoKey:
		b.WriteString(m.renderNoKeyView())
	case AIStateLoading:
		b.WriteString(m.renderLoadingView())
	case AIStateSuccess:
		b.WriteString(m.renderSuccessView())
	case AIStateError:
		b.WriteString(m.renderErrorView())
	}

	return cardStyle.Render(b.String())
}

func (m AIModal) renderNoKeyView() string {
	var b strings.Builder

	warnStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FBBF24"))
	b.WriteString(warnStyle.Render("⚡ Bring Your Own Key (BYOK) Required") + "\n\n")

	descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#CBD5E1"))
	b.WriteString(descStyle.Render("cryptoterm connects directly from your local terminal to your chosen AI provider.\nNo middleman, zero data leakage, and zero subscription fees.\n\n"))

	b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(ColorBrandAccent).Render("Supported Providers:") + "\n")
	providers := []struct {
		name string
		desc string
	}{
		{"Google Gemini", "100% Free tier available (gemini-2.0-flash) at aistudio.google.com"},
		{"Groq", "Ultra-fast (<0.5s) free tier (llama-3.3-70b) at console.groq.com"},
		{"OpenRouter / 9router", "Universal multi-model gateway at openrouter.ai"},
		{"DeepSeek", "Lowest-cost reasoning models at platform.deepseek.com"},
		{"OpenAI / Anthropic", "GPT-4o & Claude 3.5 Sonnet support"},
		{"Ollama (Local)", "Completely free, runs locally on localhost:11434 (No key needed!)"},
	}

	for _, p := range providers {
		pName := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#A78BFA")).Width(22).Render("• " + p.name)
		pDesc := lipgloss.NewStyle().Foreground(lipgloss.Color("#94A3B8")).Render(p.desc)
		b.WriteString(fmt.Sprintf("%s %s\n", pName, pDesc))
	}

	b.WriteString("\n")
	hintStyle := lipgloss.NewStyle().Foreground(ColorUp).Bold(true)
	b.WriteString(hintStyle.Render("Press [c] to configure your API key now   [Esc] Close"))

	return b.String()
}

func (m AIModal) renderLoadingView() string {
	var b strings.Builder

	spinner := lipgloss.NewStyle().Bold(true).Foreground(ColorBrandAccent).Render(spinnerFrames[m.SpinnerFrame])
	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#F1F5F9")).Render(
		fmt.Sprintf(" Analyzing %s market structure...", m.Token.DisplaySymbol),
	)
	b.WriteString(fmt.Sprintf("%s%s\n\n", spinner, title))

	stepStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#94A3B8"))
	b.WriteString(stepStyle.Render("1. Fetching recent candlestick history & volume profile...\n"))
	b.WriteString(stepStyle.Render("2. Calculating technical indicators: RSI(14), EMA(20), EMA(50)...\n"))
	b.WriteString(stepStyle.Render("3. Evaluating orderbook liquidity and key price boundaries...\n"))
	b.WriteString(stepStyle.Render(fmt.Sprintf("4. Synthesizing quantitative setup via %s (%s)...\n\n", m.Provider, m.ModelName)))

	b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#6C7086")).Render("Please wait a moment..."))
	return b.String()
}

func (m AIModal) renderSuccessView() string {
	if m.Signal == nil {
		return "No signal generated."
	}

	var b strings.Builder
	sig := m.Signal

	// 1. Signal Badge
	var actionBadge string
	switch sig.Action {
	case "STRONG_BUY":
		actionBadge = lipgloss.NewStyle().Foreground(lipgloss.Color("#000000")).Background(ColorUp).Bold(true).Padding(0, 2).Render("🚀 STRONG BUY (HIGH CONVICTION)")
	case "BUY":
		actionBadge = lipgloss.NewStyle().Foreground(lipgloss.Color("#000000")).Background(ColorUp).Bold(true).Padding(0, 2).Render("▲ BUY (BULLISH SETUP)")
	case "SELL":
		actionBadge = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Background(ColorDown).Bold(true).Padding(0, 2).Render("▼ SELL (BEARISH SETUP)")
	case "STRONG_SELL":
		actionBadge = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Background(ColorDown).Bold(true).Padding(0, 2).Render("⚠️ STRONG SELL (HIGH RISK)")
	default:
		actionBadge = lipgloss.NewStyle().Foreground(lipgloss.Color("#000000")).Background(ColorNeutral).Bold(true).Padding(0, 2).Render("━ NEUTRAL / WAIT FOR CONFIRMATION")
	}

	// Confidence score bar
	confBars := int(sig.Confidence)
	if confBars > 10 {
		confBars = 10
	}
	if confBars < 0 {
		confBars = 0
	}
	barFilled := lipgloss.NewStyle().Foreground(ColorBrandAccent).Render(strings.Repeat("■", confBars))
	barEmpty := lipgloss.NewStyle().Foreground(ColorBorder).Render(strings.Repeat("░", 10-confBars))
	confStr := fmt.Sprintf("Confidence: [%s%s] %.1f/10  |  Risk: %s", barFilled, barEmpty, sig.Confidence, sig.RiskLevel)

	b.WriteString(fmt.Sprintf("%s   %s\n\n", actionBadge, confStr))

	// 2. Trade Setup Box
	setupBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorHighlight).
		Padding(0, 2).
		MarginBottom(1)

	var sbSetup strings.Builder
	titleSetup := lipgloss.NewStyle().Bold(true).Foreground(ColorBrandAccent).Render("📊 ACTIONABLE TRADE SETUP")
	sbSetup.WriteString(titleSetup + "\n")
	sbSetup.WriteString(fmt.Sprintf("• Entry Zone   : %s\n", sig.EntryZone))
	sbSetup.WriteString(fmt.Sprintf("• Take Profit 1: $%.6g\n", sig.TakeProfit1))
	sbSetup.WriteString(fmt.Sprintf("• Take Profit 2: $%.6g\n", sig.TakeProfit2))
	sbSetup.WriteString(fmt.Sprintf("• Stop Loss    : $%.6g\n", sig.StopLoss))
	sbSetup.WriteString(fmt.Sprintf("• Risk / Reward: %s", sig.RiskReward))

	b.WriteString(setupBox.Render(sbSetup.String()) + "\n")

	// 3. Technical Summary
	if sig.TechnicalSummary != "" {
		b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#CBD5E1")).Render("💡 Technical Summary:") + "\n")
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#94A3B8")).Render(sig.TechnicalSummary) + "\n\n")
	}

	// 4. Key Reasoning Points
	if len(sig.Reasoning) > 0 {
		b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#CBD5E1")).Render("🔍 Key Rationale:") + "\n")
		for _, r := range sig.Reasoning {
			b.WriteString(fmt.Sprintf("  • %s\n", r))
		}
		b.WriteString("\n")
	}

	// 5. Invalidation
	if sig.Invalidation != "" {
		invStyle := lipgloss.NewStyle().Foreground(ColorDown)
		b.WriteString(fmt.Sprintf("⚠️ %s: %s\n\n", invStyle.Render("Invalidation"), sig.Invalidation))
	}

	// Footer hints
	hintStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#6C7086"))
	b.WriteString(hintStyle.Render(fmt.Sprintf("[r] Re-Analyze   [c] Settings   [Esc] Close   (Latency: %dms)", sig.LatencyMs)))

	return b.String()
}

func (m AIModal) renderErrorView() string {
	var b strings.Builder

	errStyle := lipgloss.NewStyle().Bold(true).Foreground(ColorDown)
	b.WriteString(errStyle.Render("❌ AI Signal Analysis Failed") + "\n\n")

	b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#CBD5E1")).Render(m.ErrorMessage) + "\n\n")

	tipStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#94A3B8"))
	b.WriteString(tipStyle.Render("Troubleshooting Tips:\n"))
	b.WriteString(tipStyle.Render("• Verify that your API Key is correct and has available quota.\n"))
	b.WriteString(tipStyle.Render("• If using Gemini or OpenRouter, make sure the model name is valid.\n"))
	b.WriteString(tipStyle.Render("• If using Ollama, ensure 'ollama serve' is running locally.\n\n"))

	hintStyle := lipgloss.NewStyle().Foreground(ColorBrandAccent)
	b.WriteString(hintStyle.Render("[r] Retry   [c] Check API Settings   [Esc] Close"))

	return b.String()
}
