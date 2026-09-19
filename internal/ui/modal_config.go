package ui

import (
	"strings"

	"github.com/fchyoga/cryptoterm/internal/ai"
	"github.com/fchyoga/cryptoterm/internal/model"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	providerOptions = []string{"Gemini", "OpenRouter", "DeepSeek", "Groq", "OpenAI", "Anthropic", "Ollama", "Custom"}
	styleOptions    = []string{"Day Trader", "Scalper", "Degen Meme"}
)

// ConfigModal manages settings, AI provider credentials, and Binance API keys.
type ConfigModal struct {
	// AI Settings
	ProviderIdx    int
	AIKeyInput     textinput.Model
	AIModelInput   textinput.Model
	StyleIdx       int

	// Binance Settings
	BncKeyInput    textinput.Model
	BncSecretInput textinput.Model

	// Preferences
	CurrencyIdx    int // 0: USD, 1: IDR
	SoundAlerts    bool

	// Active navigation field (0 to 8)
	ActiveField    int
	StatusMessage  string
}

// NewConfigModal creates a new configuration modal with existing config values.
func NewConfigModal(cfg *model.Config) ConfigModal {
	// 1. Provider
	pIdx := 0
	providerLower := strings.ToLower(cfg.AIProvider)
	for i, p := range providerOptions {
		if strings.ToLower(p) == providerLower {
			pIdx = i
			break
		}
	}

	// 2. AI Key
	aik := textinput.New()
	aik.Placeholder = "Enter AI Provider API Key (optional)"
	aik.SetValue(cfg.AIAPIKey)
	aik.EchoMode = textinput.EchoPassword
	aik.EchoCharacter = '•'
	aik.CharLimit = 160
	aik.Width = 48

	// 3. AI Model
	aim := textinput.New()
	_, defaultModel := ai.ResolveProviderConfig(cfg)
	aim.Placeholder = defaultModel
	aim.SetValue(cfg.AIModel)
	aim.CharLimit = 64
	aim.Width = 48

	// 4. Style
	sIdx := 0
	styleLower := strings.ToLower(cfg.AITradingStyle)
	for i, s := range styleOptions {
		if strings.ToLower(strings.ReplaceAll(s, " ", "")) == strings.ReplaceAll(styleLower, " ", "") {
			sIdx = i
			break
		}
	}

	// 5. Binance Key
	bk := textinput.New()
	bk.Placeholder = "Binance API Key (Read-Only, optional)"
	bk.SetValue(cfg.BinanceAPIKey)
	bk.CharLimit = 128
	bk.Width = 48

	// 6. Binance Secret
	bs := textinput.New()
	bs.Placeholder = "Binance API Secret (Read-Only, optional)"
	bs.SetValue(cfg.BinanceAPISecret)
	bs.EchoMode = textinput.EchoPassword
	bs.EchoCharacter = '•'
	bs.CharLimit = 128
	bs.Width = 48

	currencyIdx := 0
	if cfg.Currency == "IDR" {
		currencyIdx = 1
	}

	return ConfigModal{
		ProviderIdx:    pIdx,
		AIKeyInput:     aik,
		AIModelInput:   aim,
		StyleIdx:       sIdx,
		BncKeyInput:    bk,
		BncSecretInput: bs,
		CurrencyIdx:    currencyIdx,
		SoundAlerts:    cfg.SoundAlerts,
		ActiveField:    0,
	}
}

// Update processes events in the configuration modal.
func (m *ConfigModal) Update(msg tea.Msg) (bool, bool) {
	// Total fields: 9 (0: Provider, 1: AIKey, 2: AIModel, 3: Style, 4: BncKey, 5: BncSecret, 6: Currency, 7: Sound, 8: Save)
	totalFields := 9

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab", "down":
			m.ActiveField = (m.ActiveField + 1) % totalFields
			m.updateFocus()
			return false, false

		case "shift+tab", "up":
			m.ActiveField = (m.ActiveField - 1 + totalFields) % totalFields
			m.updateFocus()
			return false, false

		case "left":
			if m.ActiveField == 0 {
				m.ProviderIdx = (m.ProviderIdx - 1 + len(providerOptions)) % len(providerOptions)
				m.syncDefaultModel()
			} else if m.ActiveField == 3 {
				m.StyleIdx = (m.StyleIdx - 1 + len(styleOptions)) % len(styleOptions)
			} else if m.ActiveField == 6 {
				m.CurrencyIdx = 1 - m.CurrencyIdx
			} else if m.ActiveField == 7 {
				m.SoundAlerts = !m.SoundAlerts
			}
			return false, false

		case "right":
			if m.ActiveField == 0 {
				m.ProviderIdx = (m.ProviderIdx + 1) % len(providerOptions)
				m.syncDefaultModel()
			} else if m.ActiveField == 3 {
				m.StyleIdx = (m.StyleIdx + 1) % len(styleOptions)
			} else if m.ActiveField == 6 {
				m.CurrencyIdx = 1 - m.CurrencyIdx
			} else if m.ActiveField == 7 {
				m.SoundAlerts = !m.SoundAlerts
			}
			return false, false

		case " ":
			if m.ActiveField == 6 {
				m.CurrencyIdx = 1 - m.CurrencyIdx
			} else if m.ActiveField == 7 {
				m.SoundAlerts = !m.SoundAlerts
			}
			return false, false

		case "enter":
			if m.ActiveField == 8 {
				return true, false
			}
		}
	}

	var cmd tea.Cmd
	switch m.ActiveField {
	case 1:
		m.AIKeyInput, cmd = m.AIKeyInput.Update(msg)
	case 2:
		m.AIModelInput, cmd = m.AIModelInput.Update(msg)
	case 4:
		m.BncKeyInput, cmd = m.BncKeyInput.Update(msg)
	case 5:
		m.BncSecretInput, cmd = m.BncSecretInput.Update(msg)
	}
	_ = cmd

	return false, false
}

func (m *ConfigModal) syncDefaultModel() {
	tmpCfg := &model.Config{AIProvider: strings.ToLower(providerOptions[m.ProviderIdx])}
	_, defModel := ai.ResolveProviderConfig(tmpCfg)
	m.AIModelInput.Placeholder = defModel
}

func (m *ConfigModal) updateFocus() {
	m.AIKeyInput.Blur()
	m.AIModelInput.Blur()
	m.BncKeyInput.Blur()
	m.BncSecretInput.Blur()

	switch m.ActiveField {
	case 1:
		m.AIKeyInput.Focus()
	case 2:
		m.AIModelInput.Focus()
	case 4:
		m.BncKeyInput.Focus()
	case 5:
		m.BncSecretInput.Focus()
	}
}

// ApplyToConfig writes the modal values back to the config struct.
func (m *ConfigModal) ApplyToConfig(cfg *model.Config) {
	cfg.AIProvider = strings.ToLower(providerOptions[m.ProviderIdx])
	cfg.AIAPIKey = strings.TrimSpace(m.AIKeyInput.Value())
	modelVal := strings.TrimSpace(m.AIModelInput.Value())
	if modelVal == "" {
		tmpCfg := &model.Config{AIProvider: cfg.AIProvider}
		_, defModel := ai.ResolveProviderConfig(tmpCfg)
		modelVal = defModel
	}
	cfg.AIModel = modelVal

	switch m.StyleIdx {
	case 1:
		cfg.AITradingStyle = "scalper"
	case 2:
		cfg.AITradingStyle = "degen"
	default:
		cfg.AITradingStyle = "daytrader"
	}

	cfg.BinanceAPIKey = strings.TrimSpace(m.BncKeyInput.Value())
	cfg.BinanceAPISecret = strings.TrimSpace(m.BncSecretInput.Value())

	if m.CurrencyIdx == 1 {
		cfg.Currency = "IDR"
	} else {
		cfg.Currency = "USD"
	}
	cfg.SoundAlerts = m.SoundAlerts
}

// View renders the enhanced Config modal.
func (m ConfigModal) View() string {
	var b strings.Builder

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(ColorBrandAccent)
	b.WriteString(titleStyle.Render("⚙️  SETTINGS & API CREDENTIALS") + "\n\n")

	inputBorder := lipgloss.NewStyle().Border(lipgloss.RoundedBorder())

	// SECTION 1: AI COPILOT
	secAI := lipgloss.NewStyle().Bold(true).Foreground(ColorBrandPrimary).Render("─── 🤖 AI SIGNAL COPILOT (BYOK) ───────────────────────────")
	b.WriteString(secAI + "\n")

	// Field 0: AI Provider
	b.WriteString("AI Provider:\n")
	for i, p := range providerOptions {
		var pStyle lipgloss.Style
		if i == m.ProviderIdx {
			pStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Background(ColorBrandPrimary).Bold(true).Padding(0, 1)
		} else {
			pStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#A6ADC8")).Padding(0, 1)
		}
		b.WriteString(pStyle.Render(p) + " ")
	}
	if m.ActiveField == 0 {
		b.WriteString(lipgloss.NewStyle().Foreground(ColorBrandAccent).Render(" ◀ [←/→ to select]"))
	}
	b.WriteString("\n\n")

	// Field 1: AI API Key
	borderAIKey := inputBorder
	if m.ActiveField == 1 {
		borderAIKey = borderAIKey.BorderForeground(ColorBrandAccent)
	} else {
		borderAIKey = borderAIKey.BorderForeground(ColorBorder)
	}
	b.WriteString("AI API Key:\n")
	b.WriteString(borderAIKey.Render(m.AIKeyInput.View()) + "\n\n")

	// Field 2: AI Model
	borderAIModel := inputBorder
	if m.ActiveField == 2 {
		borderAIModel = borderAIModel.BorderForeground(ColorBrandAccent)
	} else {
		borderAIModel = borderAIModel.BorderForeground(ColorBorder)
	}
	b.WriteString("AI Model (Optional, defaults automatically):\n")
	b.WriteString(borderAIModel.Render(m.AIModelInput.View()) + "\n\n")

	// Field 3: AI Style
	b.WriteString("Analysis Style:\n")
	for i, s := range styleOptions {
		var sStyle lipgloss.Style
		if i == m.StyleIdx {
			sStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Background(ColorBrandAccent).Bold(true).Padding(0, 1)
		} else {
			sStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#A6ADC8")).Padding(0, 1)
		}
		b.WriteString(sStyle.Render(s) + " ")
	}
	if m.ActiveField == 3 {
		b.WriteString(lipgloss.NewStyle().Foreground(ColorBrandAccent).Render(" ◀ [←/→ to select]"))
	}
	b.WriteString("\n\n")

	// SECTION 2: BINANCE SPOT PORTFOLIO
	secBnc := lipgloss.NewStyle().Bold(true).Foreground(ColorBinance).Render("─── 💼 BINANCE SPOT PORTFOLIO (OPTIONAL) ───────────────────")
	b.WriteString(secBnc + "\n")

	// Field 4: Binance Key
	borderBncKey := inputBorder
	if m.ActiveField == 4 {
		borderBncKey = borderBncKey.BorderForeground(ColorBrandAccent)
	} else {
		borderBncKey = borderBncKey.BorderForeground(ColorBorder)
	}
	b.WriteString("Binance API Key:\n")
	b.WriteString(borderBncKey.Render(m.BncKeyInput.View()) + "\n\n")

	// Field 5: Binance Secret
	borderBncSec := inputBorder
	if m.ActiveField == 5 {
		borderBncSec = borderBncSec.BorderForeground(ColorBrandAccent)
	} else {
		borderBncSec = borderBncSec.BorderForeground(ColorBorder)
	}
	b.WriteString("Binance API Secret:\n")
	b.WriteString(borderBncSec.Render(m.BncSecretInput.View()) + "\n\n")

	// SECTION 3: PREFERENCES
	secPref := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#CBD5E1")).Render("─── ⚙️  PREFERENCES ─────────────────────────────────────────")
	b.WriteString(secPref + "\n")

	// Field 6: Currency
	b.WriteString("Currency: ")
	currUSD := " USD ($) "
	currIDR := " IDR (Rp) "
	if m.CurrencyIdx == 0 {
		currUSD = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Background(ColorBrandPrimary).Bold(true).Render(currUSD)
		currIDR = lipgloss.NewStyle().Foreground(lipgloss.Color("#6C7086")).Render(currIDR)
	} else {
		currUSD = lipgloss.NewStyle().Foreground(lipgloss.Color("#6C7086")).Render(currUSD)
		currIDR = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Background(ColorBrandPrimary).Bold(true).Render(currIDR)
	}
	if m.ActiveField == 6 {
		b.WriteString(lipgloss.NewStyle().Foreground(ColorBrandAccent).Render("▶ ") + currUSD + "  " + currIDR)
	} else {
		b.WriteString("  " + currUSD + "  " + currIDR)
	}
	b.WriteString("\n")

	// Field 7: Sound
	b.WriteString("Sound Alerts: ")
	soundText := " [OFF] "
	if m.SoundAlerts {
		soundText = " [ON] "
	}
	if m.ActiveField == 7 {
		b.WriteString(lipgloss.NewStyle().Foreground(ColorBrandAccent).Render("▶ ") +
			lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Background(ColorBrandPrimary).Bold(true).Render(soundText))
	} else {
		b.WriteString("  " + lipgloss.NewStyle().Foreground(lipgloss.Color("#A6ADC8")).Render(soundText))
	}
	b.WriteString("\n\n")

	// Field 8: Save Button
	saveBtn := " [ SAVE ALL SETTINGS ] "
	if m.ActiveField == 8 {
		saveBtn = lipgloss.NewStyle().Foreground(lipgloss.Color("#000000")).Background(ColorUp).Bold(true).Render(saveBtn)
	} else {
		saveBtn = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Background(ColorHighlight).Render(saveBtn)
	}
	b.WriteString(saveBtn + "\n\n")

	hintStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#6C7086"))
	b.WriteString(hintStyle.Render("[Tab/↑↓] Navigate   [←/→ or Space] Toggle   [Enter] Save   [Esc] Cancel"))

	return StyleModal.Render(b.String())
}
