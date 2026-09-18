package ui

import (
	"fmt"
	"strings"

	"github.com/fchyoga/cryptoterm/internal/model"

	"github.com/charmbracelet/lipgloss"
)

// RenderPortfolio renders the user's Binance spot balance view.
func RenderPortfolio(items []model.PortfolioItem, hasCredentials bool, currency string, usdToIdr float64, width int, errMsg string) string {
	var b strings.Builder

	if !hasCredentials {
		cardStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorBrandAccent).
			Padding(1, 3).
			Width(min(width-4, 76))

		titleStyle := lipgloss.NewStyle().Bold(true).Foreground(ColorBrandAccent)
		b.WriteString(titleStyle.Render("💼 Binance Portfolio / Balance Tracker"))
		b.WriteString("\n\n")

		b.WriteString("Your Binance API Key and Secret are not configured yet.\n\n")
		b.WriteString("To view your real-time spot portfolio, balances, and asset allocations:\n")
		b.WriteString("1. Press " + StyleHelpKey.Render("[c]") + " to open API Settings.\n")
		b.WriteString("2. Enter your Binance API Key & Secret (Read-Only permission is recommended).\n")
		b.WriteString("3. Save and return here to monitor your live portfolio!\n\n")

		noteStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#A6ADC8")).Italic(true)
		b.WriteString(noteStyle.Render("🔒 Security Notice: Your keys are stored locally on your device in\n~/.config/cryptoterm/config.json and never sent anywhere except directly to Binance."))

		return cardStyle.Render(b.String())
	}

	if errMsg != "" {
		errBox := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorDown).
			Padding(1, 2).
			Render(fmt.Sprintf("❌ Error connecting to Binance Account API:\n%s\n\nPress [c] to check your API keys or [r] to retry.", errMsg))
		return errBox
	}

	// Calculate total portfolio value
	var totalUSD float64
	for _, it := range items {
		totalUSD += it.ValueUSD
	}

	// Header banner
	headerBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorBrandPrimary).
		Background(ColorBgCard).
		Padding(0, 2).
		MarginBottom(1)

	totalFormatted := FormatPrice(totalUSD, currency, usdToIdr)
	summary := fmt.Sprintf("TOTAL PORTFOLIO VALUE: %s   |   HOLDINGS: %d ASSETS",
		lipgloss.NewStyle().Bold(true).Foreground(ColorUp).Render(totalFormatted),
		len(items),
	)
	b.WriteString(headerBox.Render(summary))
	b.WriteString("\n")

	if len(items) == 0 {
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#A6ADC8")).Render("No non-zero balances found in your Binance Spot account."))
		return b.String()
	}

	// Portfolio table headers
	headers := []string{"ASSET", "FREE", "LOCKED", "TOTAL", "PRICE", "VALUE", "ALLOC %"}
	colWidths := []int{8, 14, 12, 14, 12, 14, 10}

	var headerRow strings.Builder
	for i, h := range headers {
		cell := StyleHeaderCell.Width(colWidths[i]).Render(h)
		headerRow.WriteString(cell)
	}
	b.WriteString(headerRow.String() + "\n")

	for _, it := range items {
		var allocPct float64
		if totalUSD > 0 {
			allocPct = (it.ValueUSD / totalUSD) * 100.0
		}

		assetCell := lipgloss.NewStyle().Bold(true).Width(colWidths[0]).Render(it.Asset)
		freeCell := lipgloss.NewStyle().Width(colWidths[1]).Render(fmt.Sprintf("%.6g", it.Free))
		lockedCell := lipgloss.NewStyle().Width(colWidths[2]).Render(fmt.Sprintf("%.6g", it.Locked))
		totalCell := lipgloss.NewStyle().Width(colWidths[3]).Render(fmt.Sprintf("%.6g", it.Total))
		priceCell := lipgloss.NewStyle().Width(colWidths[4]).Render(FormatPrice(it.PriceUSD, currency, usdToIdr))
		valueCell := lipgloss.NewStyle().Bold(true).Foreground(ColorUp).Width(colWidths[5]).Render(FormatPrice(it.ValueUSD, currency, usdToIdr))
		allocCell := lipgloss.NewStyle().Foreground(ColorBrandAccent).Width(colWidths[6]).Render(fmt.Sprintf("%5.1f%%", allocPct))

		row := lipgloss.JoinHorizontal(lipgloss.Top,
			assetCell, freeCell, lockedCell, totalCell, priceCell, valueCell, allocCell,
		)
		b.WriteString(row + "\n")
	}

	return b.String()
}
