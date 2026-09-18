package ui

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"binance-terminal/internal/model"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// AlertModal handles price alert creation.
type AlertModal struct {
	Symbol       string
	CurrentPrice float64
	PriceInput   textinput.Model
	DirectionIdx int // 0: ABOVE, 1: BELOW
	ActiveField  int // 0: Input, 1: Direction, 2: Save
	StatusMsg    string
}

// NewAlertModal initializes an alert modal for a specific token.
func NewAlertModal(symbol string, currentPrice float64) AlertModal {
	ti := textinput.New()
	ti.Placeholder = fmt.Sprintf("%.4f", currentPrice)
	ti.SetValue(fmt.Sprintf("%.4f", currentPrice))
	ti.Focus()
	ti.CharLimit = 24
	ti.Width = 32

	return AlertModal{
		Symbol:       symbol,
		CurrentPrice: currentPrice,
		PriceInput:   ti,
		DirectionIdx: 0,
		ActiveField:  0,
	}
}

// Update handles key presses in the alert modal.
func (m *AlertModal) Update(msg tea.Msg) (bool, *model.PriceAlert) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab", "down":
			m.ActiveField = (m.ActiveField + 1) % 3
			if m.ActiveField == 0 {
				m.PriceInput.Focus()
			} else {
				m.PriceInput.Blur()
			}
			return false, nil

		case "shift+tab", "up":
			m.ActiveField = (m.ActiveField - 1 + 3) % 3
			if m.ActiveField == 0 {
				m.PriceInput.Focus()
			} else {
				m.PriceInput.Blur()
			}
			return false, nil

		case "left", "right", " ":
			if m.ActiveField == 1 {
				m.DirectionIdx = 1 - m.DirectionIdx
			}
			return false, nil

		case "enter":
			target, err := strconv.ParseFloat(strings.TrimSpace(m.PriceInput.Value()), 64)
			if err != nil || target <= 0 {
				m.StatusMsg = "Please enter a valid price greater than 0"
				return false, nil
			}

			direction := "ABOVE"
			if m.DirectionIdx == 1 {
				direction = "BELOW"
			}

			alert := &model.PriceAlert{
				ID:          fmt.Sprintf("%s-%d", m.Symbol, time.Now().Unix()),
				Symbol:      m.Symbol,
				TargetPrice: target,
				Direction:   direction,
				Triggered:   false,
				CreatedAt:   time.Now(),
			}

			return true, alert
		}
	}

	var cmd tea.Cmd
	if m.ActiveField == 0 {
		m.PriceInput, cmd = m.PriceInput.Update(msg)
		_ = cmd
	}

	return false, nil
}

// View renders the Alert modal.
func (m AlertModal) View() string {
	var b strings.Builder

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(ColorBrandAccent)
	b.WriteString(titleStyle.Render(fmt.Sprintf("🔔 Set Price Alert for %s", m.Symbol)))
	b.WriteString("\n\n")

	currPriceStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#A6ADC8"))
	b.WriteString(currPriceStyle.Render(fmt.Sprintf("Current Price: $%.4f\n\n", m.CurrentPrice)))

	// Field 1: Price Input
	b.WriteString("Target Price (USD):\n")
	inputBorder := lipgloss.NewStyle().Border(lipgloss.RoundedBorder())
	if m.ActiveField == 0 {
		inputBorder = inputBorder.BorderForeground(ColorBrandAccent)
	} else {
		inputBorder = inputBorder.BorderForeground(ColorBorder)
	}
	b.WriteString(inputBorder.Render(m.PriceInput.View()))
	b.WriteString("\n\n")

	// Field 2: Direction
	b.WriteString("Alert when price goes:\n")
	optAbove := " [ ▲ ABOVE Target ] "
	optBelow := " [ ▼ BELOW Target ] "
	if m.DirectionIdx == 0 {
		optAbove = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Background(ColorUp).Bold(true).Render(optAbove)
		optBelow = lipgloss.NewStyle().Foreground(lipgloss.Color("#6C7086")).Render(optBelow)
	} else {
		optAbove = lipgloss.NewStyle().Foreground(lipgloss.Color("#6C7086")).Render(optAbove)
		optBelow = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Background(ColorDown).Bold(true).Render(optBelow)
	}

	if m.ActiveField == 1 {
		b.WriteString(lipgloss.NewStyle().Foreground(ColorBrandAccent).Render("▶ ") + optAbove + "  " + optBelow)
	} else {
		b.WriteString("  " + optAbove + "  " + optBelow)
	}
	b.WriteString("\n\n")

	if m.StatusMsg != "" {
		errStyle := lipgloss.NewStyle().Foreground(ColorDown)
		b.WriteString(errStyle.Render(m.StatusMsg) + "\n\n")
	}

	saveBtn := " [ SET ALERT ] "
	if m.ActiveField == 2 {
		saveBtn = lipgloss.NewStyle().Foreground(lipgloss.Color("#000000")).Background(ColorBrandAccent).Bold(true).Render(saveBtn)
	} else {
		saveBtn = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Background(ColorHighlight).Render(saveBtn)
	}
	b.WriteString(saveBtn)
	b.WriteString("\n\n")

	hintStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#6C7086"))
	b.WriteString(hintStyle.Render("[Tab] Switch   [Space] Toggle Direction   [Enter] Save   [Esc] Cancel"))

	return StyleModal.Render(b.String())
}
