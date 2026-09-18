package ui

import (
	"fmt"
	"math"
	"strings"

	"github.com/fchyoga/cryptoterm/internal/model"

	"github.com/charmbracelet/lipgloss"
)

var (
	// Brand colors
	ColorBrandPrimary = lipgloss.Color("#7D56F4") // Neon violet
	ColorBrandAccent  = lipgloss.Color("#00F0FF") // Neon cyan
	ColorBinance      = lipgloss.Color("#F0B90B") // Binance Gold
	ColorDEX          = lipgloss.Color("#38BDF8") // Sky blue

	// Semantic colors
	ColorUp      = lipgloss.Color("#10B981") // Emerald green
	ColorDown    = lipgloss.Color("#EF4444") // Crimson red
	ColorNeutral = lipgloss.Color("#9CA3AF") // Gray
	ColorBgCard  = lipgloss.Color("#181825")
	ColorBorder  = lipgloss.Color("#313244")
	ColorHighlight = lipgloss.Color("#45475A")

	// Base styles
	StyleTitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(ColorBrandPrimary).
			Padding(0, 1)

	StyleSubtitle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#A6ADC8")).
			Italic(true)

	StyleTabActive = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(ColorBrandPrimary).
			Padding(0, 2)

	StyleTabInactive = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6C7086")).
			Padding(0, 2)

	StyleHeaderCell = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#CDD6F4")).
			Background(ColorHighlight).
			Padding(0, 1)

	StyleSelectedRow = lipgloss.NewStyle().
				Background(ColorHighlight).
				Bold(true)

	StyleNormalRow = lipgloss.NewStyle()

	StyleModal = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorBrandAccent).
			Background(ColorBgCard).
			Padding(1, 2)

	StyleHelpKey = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorBrandAccent)

	StyleHelpDesc = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#A6ADC8"))
)

// FormatPrice formats a cryptocurrency price according to its magnitude and currency.
func FormatPrice(priceUSD float64, currency string, usdToIdr float64) string {
	if currency == "IDR" {
		priceIDR := priceUSD * usdToIdr
		if priceIDR >= 1_000_000 {
			return fmt.Sprintf("Rp %.2fM", priceIDR/1_000_000)
		} else if priceIDR >= 1000 {
			return fmt.Sprintf("Rp %.0f", priceIDR)
		} else if priceIDR >= 1 {
			return fmt.Sprintf("Rp %.2f", priceIDR)
		}
		return fmt.Sprintf("Rp %.4f", priceIDR)
	}

	// USD formatting
	if priceUSD >= 1000 {
		return fmt.Sprintf("$%s", addCommas(fmt.Sprintf("%.2f", priceUSD)))
	} else if priceUSD >= 1 {
		return fmt.Sprintf("$%.2f", priceUSD)
	} else if priceUSD >= 0.01 {
		return fmt.Sprintf("$%.4f", priceUSD)
	} else if priceUSD >= 0.0001 {
		return fmt.Sprintf("$%.6f", priceUSD)
	} else if priceUSD > 0 {
		return fmt.Sprintf("$%.8f", priceUSD)
	}
	return "$0.00"
}

// FormatChange formats the 24h percentage change with colored arrows.
func FormatChange(pct float64) string {
	if math.Abs(pct) < 0.005 {
		neutralStyle := lipgloss.NewStyle().Foreground(ColorNeutral)
		return neutralStyle.Render("  0.00% ━")
	}

	if pct > 0 {
		upStyle := lipgloss.NewStyle().Foreground(ColorUp).Bold(true)
		return upStyle.Render(fmt.Sprintf("+%5.2f%% ▲", pct))
	}

	downStyle := lipgloss.NewStyle().Foreground(ColorDown).Bold(true)
	return downStyle.Render(fmt.Sprintf("%6.2f%% ▼", pct))
}

// FormatVolume formats large numbers into B/M/K format.
func FormatVolume(volUSD float64, currency string, usdToIdr float64) string {
	val := volUSD
	prefix := "$"
	if currency == "IDR" {
		val = volUSD * usdToIdr
		prefix = "Rp "
	}

	if val >= 1_000_000_000 {
		return fmt.Sprintf("%s%.2fB", prefix, val/1_000_000_000)
	} else if val >= 1_000_000 {
		return fmt.Sprintf("%s%.2fM", prefix, val/1_000_000)
	} else if val >= 1_000 {
		return fmt.Sprintf("%s%.1fK", prefix, val/1_000)
	} else if val > 0 {
		return fmt.Sprintf("%s%.0f", prefix, val)
	}
	return "-"
}

// FormatSourceBadge returns a stylized source indicator badge.
func FormatSourceBadge(source model.TokenSource) string {
	if source == model.SourceBinance {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("#000000")).
			Background(ColorBinance).
			Bold(true).
			Render(" BNC ")
	}
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#000000")).
		Background(ColorDEX).
		Bold(true).
		Render(" DEX ")
}

// FormatCategoryBadge returns a stylized category badge.
func FormatCategoryBadge(category model.TokenType) string {
	switch category {
	case model.TypeMajor:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#89B4FA")).Render("MAJOR")
	case model.TypeMeme:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#F38BA8")).Bold(true).Render("MEME ")
	case model.TypeStable:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#A6E3A1")).Render("STABL")
	default:
		return "TOKEN"
	}
}

func addCommas(s string) string {
	parts := strings.Split(s, ".")
	intPart := parts[0]
	var result []byte
	n := len(intPart)
	for i, c := range intPart {
		if (n-i)%3 == 0 && i != 0 {
			result = append(result, ',')
		}
		result = append(result, byte(c))
	}
	if len(parts) > 1 {
		return string(result) + "." + parts[1]
	}
	return string(result)
}
