package ui

import (
	"strings"

	"github.com/fchyoga/cryptoterm/internal/model"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// AddModal handles user input for adding a new token.
type AddModal struct {
	TextInput   textinput.Model
	SourceIdx   int // 0: Auto-Detect, 1: Binance, 2: DEX (DexScreener)
	CategoryIdx int // 0: Meme, 1: Major, 2: Stable
	ActiveField int // 0: Input, 1: Source, 2: Category, 3: Submit
	StatusMsg   string
	IsLoading   bool
}

var (
	sourceOptions   = []string{"Auto-Detect", "Binance", "DEX (DexScreener)"}
	categoryOptions = []string{"MEME", "MAJOR", "STABLE"}
)

// NewAddModal creates and initializes the Add Token modal.
func NewAddModal() AddModal {
	ti := textinput.New()
	ti.Placeholder = "e.g. SOL, PEPE, DOGE, or Contract Address"
	ti.Focus()
	ti.CharLimit = 64
	ti.Width = 44

	return AddModal{
		TextInput:   ti,
		SourceIdx:   0,
		CategoryIdx: 0,
		ActiveField: 0,
	}
}

// Update processes keyboard input inside the Add Token modal.
func (m *AddModal) Update(msg tea.Msg) (bool, *model.WatchItem) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab", "down":
			m.ActiveField = (m.ActiveField + 1) % 4
			if m.ActiveField == 0 {
				m.TextInput.Focus()
			} else {
				m.TextInput.Blur()
			}
			return false, nil

		case "shift+tab", "up":
			m.ActiveField = (m.ActiveField - 1 + 4) % 4
			if m.ActiveField == 0 {
				m.TextInput.Focus()
			} else {
				m.TextInput.Blur()
			}
			return false, nil

		case "left":
			if m.ActiveField == 1 {
				m.SourceIdx = (m.SourceIdx - 1 + len(sourceOptions)) % len(sourceOptions)
			} else if m.ActiveField == 2 {
				m.CategoryIdx = (m.CategoryIdx - 1 + len(categoryOptions)) % len(categoryOptions)
			}
			return false, nil

		case "right":
			if m.ActiveField == 1 {
				m.SourceIdx = (m.SourceIdx + 1) % len(sourceOptions)
			} else if m.ActiveField == 2 {
				m.CategoryIdx = (m.CategoryIdx + 1) % len(categoryOptions)
			}
			return false, nil

		case "enter":
			val := strings.TrimSpace(m.TextInput.Value())
			if val == "" {
				m.StatusMsg = "Please enter a symbol or contract address"
				return false, nil
			}

			// Determine source
			var source model.TokenSource
			if m.SourceIdx == 1 {
				source = model.SourceBinance
			} else if m.SourceIdx == 2 {
				source = model.SourceDEX
			} else {
				// Auto-detect: if length > 25, probably a Solana/EVM contract address
				if len(val) >= 25 {
					source = model.SourceDEX
				} else {
					source = model.SourceBinance
				}
			}

			// Determine category
			var category model.TokenType
			switch m.CategoryIdx {
			case 1:
				category = model.TypeMajor
			case 2:
				category = model.TypeStable
			default:
				category = model.TypeMeme
			}

			displaySymbol := strings.ToUpper(val)
			symbol := val
			chain := "binance"
			if source == model.SourceBinance {
				if !strings.HasSuffix(displaySymbol, "USDT") {
					symbol = displaySymbol + "USDT"
				} else {
					symbol = displaySymbol
					displaySymbol = strings.TrimSuffix(symbol, "USDT")
				}
			} else {
				chain = "solana" // default DEX chain, will be refined on fetch
			}

			item := &model.WatchItem{
				Symbol:        symbol,
				DisplaySymbol: displaySymbol,
				Name:          displaySymbol,
				Source:        source,
				Category:      category,
				Chain:         chain,
			}

			return true, item
		}
	}

	var cmd tea.Cmd
	if m.ActiveField == 0 {
		m.TextInput, cmd = m.TextInput.Update(msg)
		_ = cmd
	}
	return false, nil
}

// View renders the Add Token modal UI.
func (m AddModal) View() string {
	var b strings.Builder

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(ColorBrandAccent)
	b.WriteString(titleStyle.Render("➕ Add Token / Contract to Watchlist"))
	b.WriteString("\n\n")

	// Field 1: Input
	inputBorder := lipgloss.NewStyle().Border(lipgloss.RoundedBorder())
	if m.ActiveField == 0 {
		inputBorder = inputBorder.BorderForeground(ColorBrandAccent)
	} else {
		inputBorder = inputBorder.BorderForeground(ColorBorder)
	}
	b.WriteString("Token Symbol or Contract Address:\n")
	b.WriteString(inputBorder.Render(m.TextInput.View()))
	b.WriteString("\n\n")

	// Field 2: Source selector
	b.WriteString("Data Source:\n")
	for i, opt := range sourceOptions {
		var optStyle lipgloss.Style
		if i == m.SourceIdx {
			optStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFFFF")).
				Background(ColorBrandPrimary).
				Bold(true).
				Padding(0, 1)
		} else {
			optStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#A6ADC8")).
				Padding(0, 1)
		}
		b.WriteString(optStyle.Render(opt) + " ")
	}
	b.WriteString("\n\n")

	// Field 3: Category selector
	b.WriteString("Category:\n")
	for i, opt := range categoryOptions {
		var optStyle lipgloss.Style
		if i == m.CategoryIdx {
			optStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFFFF")).
				Background(ColorBrandPrimary).
				Bold(true).
				Padding(0, 1)
		} else {
			optStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#A6ADC8")).
				Padding(0, 1)
		}
		b.WriteString(optStyle.Render(opt) + " ")
	}
	b.WriteString("\n\n")

	// Submit hint
	if m.StatusMsg != "" {
		errStyle := lipgloss.NewStyle().Foreground(ColorDown)
		b.WriteString(errStyle.Render(m.StatusMsg) + "\n\n")
	}

	hintStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#6C7086"))
	b.WriteString(hintStyle.Render("[Tab/↑↓] Switch Field   [←/→] Change Option   [Enter] Add Token   [Esc] Cancel"))

	return StyleModal.Render(b.String())
}
