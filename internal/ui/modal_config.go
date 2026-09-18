package ui

import (
	"strings"

	"github.com/fchyoga/cryptoterm/internal/model"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ConfigModal manages settings and Binance API credentials.
type ConfigModal struct {
	KeyInput      textinput.Model
	SecretInput   textinput.Model
	CurrencyIdx   int // 0: USD, 1: IDR
	SoundAlerts   bool
	ActiveField   int // 0: Key, 1: Secret, 2: Currency, 3: Sound, 4: Save
	StatusMessage string
}

// NewConfigModal creates a new configuration modal with existing config values.
func NewConfigModal(cfg *model.Config) ConfigModal {
	ki := textinput.New()
	ki.Placeholder = "Enter Binance API Key (optional)"
	ki.SetValue(cfg.BinanceAPIKey)
	ki.CharLimit = 128
	ki.Width = 52
	ki.Focus()

	si := textinput.New()
	si.Placeholder = "Enter Binance API Secret (optional)"
	si.SetValue(cfg.BinanceAPISecret)
	si.EchoMode = textinput.EchoPassword
	si.EchoCharacter = '•'
	si.CharLimit = 128
	si.Width = 52

	currencyIdx := 0
	if cfg.Currency == "IDR" {
		currencyIdx = 1
	}

	return ConfigModal{
		KeyInput:    ki,
		SecretInput: si,
		CurrencyIdx: currencyIdx,
		SoundAlerts: cfg.SoundAlerts,
		ActiveField: 0,
	}
}

// Update processes events in the configuration modal.
func (m *ConfigModal) Update(msg tea.Msg) (bool, bool) {
	// Returns: (saved, cancelled)
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab", "down":
			m.ActiveField = (m.ActiveField + 1) % 5
			m.updateFocus()
			return false, false

		case "shift+tab", "up":
			m.ActiveField = (m.ActiveField - 1 + 5) % 5
			m.updateFocus()
			return false, false

		case "left", "right", " ":
			if m.ActiveField == 2 {
				m.CurrencyIdx = 1 - m.CurrencyIdx
			} else if m.ActiveField == 3 {
				m.SoundAlerts = !m.SoundAlerts
			}
			return false, false

		case "enter":
			if m.ActiveField == 4 || m.ActiveField == 0 || m.ActiveField == 1 {
				return true, false
			}
		}
	}

	var cmd tea.Cmd
	if m.ActiveField == 0 {
		m.KeyInput, cmd = m.KeyInput.Update(msg)
		_ = cmd
	} else if m.ActiveField == 1 {
		m.SecretInput, cmd = m.SecretInput.Update(msg)
		_ = cmd
	}

	return false, false
}

func (m *ConfigModal) updateFocus() {
	if m.ActiveField == 0 {
		m.KeyInput.Focus()
		m.SecretInput.Blur()
	} else if m.ActiveField == 1 {
		m.KeyInput.Blur()
		m.SecretInput.Focus()
	} else {
		m.KeyInput.Blur()
		m.SecretInput.Blur()
	}
}

// ApplyToConfig writes the modal values back to the config struct.
func (m *ConfigModal) ApplyToConfig(cfg *model.Config) {
	cfg.BinanceAPIKey = strings.TrimSpace(m.KeyInput.Value())
	cfg.BinanceAPISecret = strings.TrimSpace(m.SecretInput.Value())
	if m.CurrencyIdx == 1 {
		cfg.Currency = "IDR"
	} else {
		cfg.Currency = "USD"
	}
	cfg.SoundAlerts = m.SoundAlerts
}

// View renders the modal.
func (m ConfigModal) View() string {
	var b strings.Builder

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(ColorBrandAccent)
	b.WriteString(titleStyle.Render("⚙️  API Credentials & Terminal Settings"))
	b.WriteString("\n")

	infoStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#A6ADC8")).Italic(true)
	b.WriteString(infoStyle.Render("Note: Public price monitoring works without API keys.\nAPI keys are only used for the Portfolio / Balance tab."))
	b.WriteString("\n\n")

	inputBorder := lipgloss.NewStyle().Border(lipgloss.RoundedBorder())

	// Field 1: API Key
	border1 := inputBorder
	if m.ActiveField == 0 {
		border1 = border1.BorderForeground(ColorBrandAccent)
	} else {
		border1 = border1.BorderForeground(ColorBorder)
	}
	b.WriteString("Binance API Key:\n")
	b.WriteString(border1.Render(m.KeyInput.View()))
	b.WriteString("\n\n")

	// Field 2: API Secret
	border2 := inputBorder
	if m.ActiveField == 1 {
		border2 = border2.BorderForeground(ColorBrandAccent)
	} else {
		border2 = border2.BorderForeground(ColorBorder)
	}
	b.WriteString("Binance API Secret:\n")
	b.WriteString(border2.Render(m.SecretInput.View()))
	b.WriteString("\n\n")

	// Field 3: Currency
	b.WriteString("Display Currency:\n")
	currUSD := " USD ($) "
	currIDR := " IDR (Rp) "
	if m.CurrencyIdx == 0 {
		currUSD = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Background(ColorBrandPrimary).Bold(true).Render(currUSD)
		currIDR = lipgloss.NewStyle().Foreground(lipgloss.Color("#6C7086")).Render(currIDR)
	} else {
		currUSD = lipgloss.NewStyle().Foreground(lipgloss.Color("#6C7086")).Render(currUSD)
		currIDR = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Background(ColorBrandPrimary).Bold(true).Render(currIDR)
	}
	if m.ActiveField == 2 {
		b.WriteString(lipgloss.NewStyle().Foreground(ColorBrandAccent).Render("▶ ") + currUSD + "  " + currIDR)
	} else {
		b.WriteString("  " + currUSD + "  " + currIDR)
	}
	b.WriteString("\n\n")

	// Field 4: Sound Alerts
	b.WriteString("Terminal Sound Alerts:\n")
	soundStatus := " [OFF] "
	if m.SoundAlerts {
		soundStatus = " [ON] "
	}
	if m.ActiveField == 3 {
		b.WriteString(lipgloss.NewStyle().Foreground(ColorBrandAccent).Render("▶ ") +
			lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Background(ColorBrandPrimary).Bold(true).Render(soundStatus) +
			" (Press Space/Enter to toggle)")
	} else {
		b.WriteString("  " + lipgloss.NewStyle().Foreground(lipgloss.Color("#A6ADC8")).Render(soundStatus))
	}
	b.WriteString("\n\n")

	// Field 5: Save button
	saveBtn := " [ SAVE & APPLY ] "
	if m.ActiveField == 4 {
		saveBtn = lipgloss.NewStyle().Foreground(lipgloss.Color("#000000")).Background(ColorUp).Bold(true).Render(saveBtn)
	} else {
		saveBtn = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Background(ColorHighlight).Render(saveBtn)
	}
	b.WriteString(saveBtn)
	b.WriteString("\n\n")

	hintStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#6C7086"))
	b.WriteString(hintStyle.Render("[Tab/↑↓] Navigate   [Space/Enter] Toggle/Save   [Esc] Cancel"))

	return StyleModal.Render(b.String())
}
