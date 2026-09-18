package ui

import (
	"fmt"
	"strings"
	"time"

	"binance-terminal/internal/model"

	"github.com/charmbracelet/lipgloss"
)

// View renders the full terminal UI layout.
func (m UIModel) View() string {
	// If modal is active, display modal centered over the view
	if m.ShowAddModal {
		return m.renderWithModal(m.AddModal.View())
	}
	if m.ShowConfigModal {
		return m.renderWithModal(m.ConfigModal.View())
	}
	if m.ShowAlertModal {
		return m.renderWithModal(m.AlertModal.View())
	}
	if m.ShowHelp {
		return m.renderWithModal(m.renderHelpOverlay())
	}

	var b strings.Builder

	// 1. Header
	b.WriteString(m.renderHeader())
	b.WriteString("\n")

	// 2. Body based on active tab
	switch m.ActiveTab {
	case ViewWatchlist:
		b.WriteString(m.renderWatchlistView())
	case ViewMemeRadar:
		b.WriteString(m.renderMemeRadarView())
	case ViewPortfolio:
		b.WriteString(RenderPortfolio(m.PortfolioItems, m.BinanceREST.HasCredentials(), m.Cfg.Currency, m.Cfg.USDToIDR, m.Width, m.PortfolioErr))
	case ViewDetail:
		b.WriteString(m.renderDetailView())
	}

	// 3. Footer / Status Bar
	b.WriteString("\n")
	b.WriteString(m.renderFooter())

	return b.String()
}

func (m UIModel) renderHeader() string {
	// Brand title
	title := StyleTitle.Render("⚡ CRYPTO TERMINAL")

	// Tabs
	tabs := []string{"[1] Watchlist", "[2] Meme Radar", "[3] Portfolio", "[4] Detail"}
	var renderedTabs []string
	for i, t := range tabs {
		if ActiveView(i) == m.ActiveTab {
			renderedTabs = append(renderedTabs, StyleTabActive.Render(t))
		} else {
			renderedTabs = append(renderedTabs, StyleTabInactive.Render(t))
		}
	}
	tabBar := strings.Join(renderedTabs, " ")

	// Connection & Status indicators
	var wsBadge string
	if m.WSConnected {
		wsBadge = lipgloss.NewStyle().Foreground(ColorUp).Bold(true).Render("● BNC WS")
	} else {
		wsBadge = lipgloss.NewStyle().Foreground(ColorDown).Render("○ BNC WS")
	}

	dexBadge := lipgloss.NewStyle().Foreground(ColorDEX).Bold(true).Render("● DEX")
	currBadge := lipgloss.NewStyle().Foreground(ColorBrandAccent).Render(fmt.Sprintf("[%s]", m.Cfg.Currency))
	clock := lipgloss.NewStyle().Foreground(lipgloss.Color("#6C7086")).Render(time.Now().UTC().Format("15:04:05 UTC"))

	metaBar := fmt.Sprintf("%s  %s  %s  %s", wsBadge, dexBadge, currBadge, clock)

	headerLine := fmt.Sprintf("%s  %s", title, tabBar)
	gap := m.Width - lipgloss.Width(headerLine) - lipgloss.Width(metaBar) - 2
	if gap < 2 {
		gap = 2
	}

	divider := lipgloss.NewStyle().Foreground(ColorBorder).Render(strings.Repeat("─", max(m.Width-2, 40)))

	return fmt.Sprintf("%s%s%s\n%s", headerLine, strings.Repeat(" ", gap), metaBar, divider)
}

func (m UIModel) renderWatchlistView() string {
	var b strings.Builder

	// Filter bar if active or filtering text present
	if m.IsFiltering || m.FilterText != "" {
		filterPrompt := lipgloss.NewStyle().Foreground(ColorBrandAccent).Bold(true).Render("🔍 Filter: ")
		b.WriteString(filterPrompt + m.FilterInput.View() + "\n")
	}

	items := m.getFilteredWatchlist()
	if len(items) == 0 {
		emptyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#A6ADC8")).Padding(2, 4)
		b.WriteString(emptyStyle.Render("No tokens match your current filter. Press [a] to add a token or [/] to clear filter."))
		return b.String()
	}

	// Table column widths
	colTicker := 14
	colSource := 7
	colCat := 7
	colPrice := 15
	colChange := 12
	colHigh := 13
	colLow := 13
	colVol := 12
	colSpark := 12

	// Headers
	hRow := lipgloss.JoinHorizontal(lipgloss.Top,
		StyleHeaderCell.Width(colTicker).Render("TICKER"),
		StyleHeaderCell.Width(colSource).Render("SRC"),
		StyleHeaderCell.Width(colCat).Render("CAT"),
		StyleHeaderCell.Width(colPrice).Render("PRICE"),
		StyleHeaderCell.Width(colChange).Render("24H CHG"),
		StyleHeaderCell.Width(colHigh).Render("24H HIGH"),
		StyleHeaderCell.Width(colLow).Render("24H LOW"),
		StyleHeaderCell.Width(colVol).Render("VOLUME"),
		StyleHeaderCell.Width(colSpark).Render("TREND"),
	)
	b.WriteString(hRow + "\n")

	// Rows
	maxRows := m.Height - 12
	if maxRows < 5 {
		maxRows = 10
	}

	startIdx := 0
	if m.SelectedIdx >= maxRows {
		startIdx = m.SelectedIdx - maxRows + 1
	}
	endIdx := min(startIdx+maxRows, len(items))

	for i := startIdx; i < endIdx; i++ {
		it := items[i]
		md := m.MarketData[it.Symbol]

		isSelected := (i == m.SelectedIdx)

		// Format cells
		tickerStr := it.DisplaySymbol
		if isSelected {
			tickerStr = "▶ " + tickerStr
		} else {
			tickerStr = "  " + tickerStr
		}

		priceStr := FormatPrice(md.Price, m.Cfg.Currency, m.Cfg.USDToIDR)
		changeStr := FormatChange(md.PriceChange24h)
		highStr := FormatPrice(md.High24h, m.Cfg.Currency, m.Cfg.USDToIDR)
		lowStr := FormatPrice(md.Low24h, m.Cfg.Currency, m.Cfg.USDToIDR)
		volStr := FormatVolume(md.Volume24h, m.Cfg.Currency, m.Cfg.USDToIDR)
		sparkStr := GenerateSparkline(md.PriceHistory, colSpark-2)

		rowStyle := StyleNormalRow
		if isSelected {
			rowStyle = StyleSelectedRow
		}

		cTicker := rowStyle.Width(colTicker).Render(tickerStr)
		cSource := rowStyle.Width(colSource).Render(FormatSourceBadge(it.Source))
		cCat := rowStyle.Width(colCat).Render(FormatCategoryBadge(it.Category))
		cPrice := rowStyle.Width(colPrice).Render(priceStr)
		cChange := rowStyle.Width(colChange).Render(changeStr)
		cHigh := rowStyle.Width(colHigh).Render(highStr)
		cLow := rowStyle.Width(colLow).Render(lowStr)
		cVol := rowStyle.Width(colVol).Render(volStr)
		cSpark := rowStyle.Width(colSpark).Render(sparkStr)

		row := lipgloss.JoinHorizontal(lipgloss.Top,
			cTicker, cSource, cCat, cPrice, cChange, cHigh, cLow, cVol, cSpark,
		)
		b.WriteString(row + "\n")
	}

	// Quick snapshot pane for selected coin at the bottom of watchlist
	if len(items) > 0 && m.SelectedIdx < len(items) {
		selectedItem := items[m.SelectedIdx]
		selectedMD := m.MarketData[selectedItem.Symbol]
		b.WriteString("\n")
		b.WriteString(m.renderMiniSummary(selectedItem, selectedMD))
	}

	return b.String()
}

func (m UIModel) renderMiniSummary(item model.WatchItem, md model.MarketData) string {
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorBorder).
		Padding(0, 1)

	symbolPart := lipgloss.NewStyle().Bold(true).Foreground(ColorBrandAccent).Render(item.DisplaySymbol)
	namePart := lipgloss.NewStyle().Foreground(lipgloss.Color("#A6ADC8")).Render(fmt.Sprintf("(%s - %s)", item.Name, item.Chain))

	// 24h High/Low range bar
	var rangeBar string
	if md.High24h > md.Low24h && md.Price >= md.Low24h {
		ratio := (md.Price - md.Low24h) / (md.High24h - md.Low24h)
		totalBars := 20
		filled := int(ratio * float64(totalBars))
		if filled > totalBars {
			filled = totalBars
		}
		if filled < 0 {
			filled = 0
		}
		barFill := lipgloss.NewStyle().Foreground(ColorBrandAccent).Render(strings.Repeat("■", filled))
		barEmpty := lipgloss.NewStyle().Foreground(ColorHighlight).Render(strings.Repeat("░", totalBars-filled))
		rangeBar = fmt.Sprintf("24h Range: L %s [%s%s] H %s",
			FormatPrice(md.Low24h, m.Cfg.Currency, m.Cfg.USDToIDR),
			barFill, barEmpty,
			FormatPrice(md.High24h, m.Cfg.Currency, m.Cfg.USDToIDR),
		)
	}

	var extra string
	if item.Source == model.SourceDEX {
		extra = fmt.Sprintf(" | Liquidity: %s | FDV: %s",
			FormatVolume(md.LiquidityUSD, m.Cfg.Currency, m.Cfg.USDToIDR),
			FormatVolume(md.MarketCapUSD, m.Cfg.Currency, m.Cfg.USDToIDR),
		)
	}

	content := fmt.Sprintf("%s %s  %s%s", symbolPart, namePart, rangeBar, extra)
	return boxStyle.Render(content)
}

func (m UIModel) renderMemeRadarView() string {
	var b strings.Builder

	titleBox := lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorBrandAccent).
		MarginBottom(1).
		Render("🔥 TRENDING DEX MEME COINS (Solana, Base, Ethereum via DexScreener)")
	b.WriteString(titleBox + "\n")

	if len(m.TrendingMemes) == 0 {
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#A6ADC8")).Render("Scanning DEXs for trending meme tokens..."))
		return b.String()
	}

	colToken := 16
	colChain := 10
	colPrice := 16
	colChange := 12
	colLiq := 14
	colMcap := 14
	colVol := 14
	colSpark := 12

	hRow := lipgloss.JoinHorizontal(lipgloss.Top,
		StyleHeaderCell.Width(colToken).Render("TOKEN"),
		StyleHeaderCell.Width(colChain).Render("CHAIN"),
		StyleHeaderCell.Width(colPrice).Render("PRICE"),
		StyleHeaderCell.Width(colChange).Render("24H CHG"),
		StyleHeaderCell.Width(colLiq).Render("LIQUIDITY"),
		StyleHeaderCell.Width(colMcap).Render("FDV / MCAP"),
		StyleHeaderCell.Width(colVol).Render("VOLUME"),
		StyleHeaderCell.Width(colSpark).Render("TREND"),
	)
	b.WriteString(hRow + "\n")

	for i, md := range m.TrendingMemes {
		isSelected := (i == m.MemeSelectedIdx)
		rowStyle := StyleNormalRow
		if isSelected {
			rowStyle = StyleSelectedRow
		}

		tokenStr := md.DisplaySymbol
		if isSelected {
			tokenStr = "▶ " + tokenStr
		} else {
			tokenStr = "  " + tokenStr
		}

		chainBadge := lipgloss.NewStyle().Foreground(ColorDEX).Bold(true).Render(strings.ToUpper(md.Chain))
		priceStr := FormatPrice(md.Price, m.Cfg.Currency, m.Cfg.USDToIDR)
		changeStr := FormatChange(md.PriceChange24h)
		liqStr := FormatVolume(md.LiquidityUSD, m.Cfg.Currency, m.Cfg.USDToIDR)
		mcapStr := FormatVolume(md.MarketCapUSD, m.Cfg.Currency, m.Cfg.USDToIDR)
		volStr := FormatVolume(md.Volume24h, m.Cfg.Currency, m.Cfg.USDToIDR)
		sparkStr := GenerateSparkline(md.PriceHistory, colSpark-2)

		row := lipgloss.JoinHorizontal(lipgloss.Top,
			rowStyle.Width(colToken).Render(tokenStr),
			rowStyle.Width(colChain).Render(chainBadge),
			rowStyle.Width(colPrice).Render(priceStr),
			rowStyle.Width(colChange).Render(changeStr),
			rowStyle.Width(colLiq).Render(liqStr),
			rowStyle.Width(colMcap).Render(mcapStr),
			rowStyle.Width(colVol).Render(volStr),
			rowStyle.Width(colSpark).Render(sparkStr),
		)
		b.WriteString(row + "\n")
	}

	return b.String()
}

func (m UIModel) renderDetailView() string {
	items := m.getFilteredWatchlist()
	if len(items) == 0 || m.SelectedIdx >= len(items) {
		return "No token selected. Press [1] to view Watchlist."
	}

	it := items[m.SelectedIdx]
	md := m.MarketData[it.Symbol]

	cardStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorBrandAccent).
		Background(ColorBgCard).
		Padding(1, 3).
		Width(min(m.Width-4, 80))

	var b strings.Builder

	// Title
	symStyle := lipgloss.NewStyle().Bold(true).Foreground(ColorBrandAccent)
	b.WriteString(fmt.Sprintf("%s - %s  [%s]\n\n",
		symStyle.Render(it.DisplaySymbol),
		it.Name,
		it.Source,
	))

	// Large price & 24h change
	priceFormatted := FormatPrice(md.Price, m.Cfg.Currency, m.Cfg.USDToIDR)
	changeFormatted := FormatChange(md.PriceChange24h)
	b.WriteString(fmt.Sprintf("Current Price: %s   (%s)\n\n",
		lipgloss.NewStyle().Bold(true).Foreground(ColorUp).Render(priceFormatted),
		changeFormatted,
	))

	// 24h Stats
	b.WriteString(fmt.Sprintf("24h High:   %s\n", FormatPrice(md.High24h, m.Cfg.Currency, m.Cfg.USDToIDR)))
	b.WriteString(fmt.Sprintf("24h Low:    %s\n", FormatPrice(md.Low24h, m.Cfg.Currency, m.Cfg.USDToIDR)))
	b.WriteString(fmt.Sprintf("24h Volume: %s\n", FormatVolume(md.Volume24h, m.Cfg.Currency, m.Cfg.USDToIDR)))

	if it.Source == model.SourceDEX {
		b.WriteString(fmt.Sprintf("Liquidity:  %s\n", FormatVolume(md.LiquidityUSD, m.Cfg.Currency, m.Cfg.USDToIDR)))
		b.WriteString(fmt.Sprintf("FDV / MCAP: %s\n", FormatVolume(md.MarketCapUSD, m.Cfg.Currency, m.Cfg.USDToIDR)))
		b.WriteString(fmt.Sprintf("Chain:      %s\n", strings.ToUpper(it.Chain)))
		b.WriteString(fmt.Sprintf("Contract:   %s\n", it.Symbol))
	} else {
		b.WriteString(fmt.Sprintf("Pair:       %s\n", it.Symbol))
	}

	b.WriteString("\nRecent Price Trend:\n")
	sparkline := GenerateSparkline(md.PriceHistory, 30)
	b.WriteString(sparkline + "\n\n")

	hintStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#6C7086"))
	b.WriteString(hintStyle.Render("Press [1] to return to Watchlist   [!] to set price alert"))

	return cardStyle.Render(b.String())
}

func (m UIModel) renderFooter() string {
	divider := lipgloss.NewStyle().Foreground(ColorBorder).Render(strings.Repeat("─", max(m.Width-2, 40)))

	// Toast notification
	var toastBar string
	if m.ToastMessage != "" {
		toastStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#000000")).
			Background(ColorBrandAccent).
			Bold(true).
			Padding(0, 2)
		toastBar = toastStyle.Render(m.ToastMessage) + "  "
	}

	// Shortcuts
	hints := []string{
		StyleHelpKey.Render("[a]") + StyleHelpDesc.Render(" Add"),
		StyleHelpKey.Render("[d]") + StyleHelpDesc.Render(" Remove"),
		StyleHelpKey.Render("[/]") + StyleHelpDesc.Render(" Filter"),
		StyleHelpKey.Render("[s]") + StyleHelpDesc.Render(" Sort"),
		StyleHelpKey.Render("[c]") + StyleHelpDesc.Render(" Config"),
		StyleHelpKey.Render("[!]") + StyleHelpDesc.Render(" Alert"),
		StyleHelpKey.Render("[r]") + StyleHelpDesc.Render(" Refresh"),
		StyleHelpKey.Render("[?]") + StyleHelpDesc.Render(" Help"),
		StyleHelpKey.Render("[q]") + StyleHelpDesc.Render(" Quit"),
	}

	shortcuts := strings.Join(hints, "  ")
	return fmt.Sprintf("%s\n%s%s", divider, toastBar, shortcuts)
}

func (m UIModel) renderHelpOverlay() string {
	var b strings.Builder

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(ColorBrandAccent)
	b.WriteString(titleStyle.Render("📖 Keyboard Shortcuts & Navigation Guide"))
	b.WriteString("\n\n")

	keys := [][]string{
		{"Tab / 1-4", "Switch between Watchlist, Meme Radar, Portfolio, and Detail tabs"},
		{"↑ / ↓ or j / k", "Navigate up and down the token list"},
		{"Enter", "View detailed statistics and price history for selected token"},
		{"a", "Add a new token (supports Binance symbols or DEX contract addresses)"},
		{"d / x", "Delete selected token from your watchlist"},
		{"s", "Cycle sort mode (Default, 24h Change, Price, Volume, Name)"},
		{"/", "Quick filter/search tokens in current watchlist"},
		{"c", "Configure Binance API Key & Secret (for Portfolio view) and currency"},
		{"!", "Set custom price target alert for selected token"},
		{"r", "Force refresh all market data and portfolio balances"},
		{"?", "Toggle this help screen"},
		{"q / Ctrl+C", "Quit application"},
	}

	for _, k := range keys {
		keyStr := StyleHelpKey.Width(16).Render(k[0])
		descStr := StyleHelpDesc.Render(k[1])
		b.WriteString(fmt.Sprintf("%s %s\n", keyStr, descStr))
	}

	b.WriteString("\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("#6C7086")).Render("Press [Esc] or [?] to close this help screen."))

	return StyleModal.Render(b.String())
}

func (m UIModel) renderWithModal(modalView string) string {
	return lipgloss.Place(m.Width, m.Height,
		lipgloss.Center, lipgloss.Center,
		modalView,
	)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
