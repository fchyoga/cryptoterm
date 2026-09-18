package ui

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"binance-terminal/internal/api"
	"binance-terminal/internal/config"
	"binance-terminal/internal/model"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// ActiveView represents the current main tab.
type ActiveView int

const (
	ViewWatchlist ActiveView = iota
	ViewMemeRadar
	ViewPortfolio
	ViewDetail
)

// Message types for Bubbletea
type (
	MarketDataMsg      model.MarketData
	BatchMarketDataMsg map[string]model.MarketData
	WSStatusMsg        bool
	DexTickMsg         struct{}
	TrendingMemesMsg   []model.MarketData
	PortfolioMsg    struct {
		Items []model.PortfolioItem
		Err   error
	}
	ToastMsg struct {
		Text   string
		IsBell bool
	}
	ClearToastMsg struct{}
)

// UIModel is the root Bubbletea model.
type UIModel struct {
	Cfg             *model.Config
	MarketData      map[string]model.MarketData
	ActiveTab       ActiveView
	SelectedIdx     int
	MemeSelectedIdx int
	TrendingMemes   []model.MarketData
	PortfolioItems  []model.PortfolioItem
	PortfolioErr    string

	// Modals & Overlays
	ShowAddModal    bool
	AddModal        AddModal
	ShowConfigModal bool
	ConfigModal     ConfigModal
	ShowAlertModal  bool
	AlertModal      AlertModal
	ShowHelp        bool

	// Search / Filter
	IsFiltering bool
	FilterInput textinput.Model
	FilterText  string

	// Sorting: 0: Default, 1: 24h Change Desc, 2: Price Desc, 3: Volume Desc, 4: Name Asc
	SortMode int

	// Notifications
	ToastMessage string
	ToastTimer   int

	// API Clients
	BinanceWS   *api.BinanceWSClient
	BinanceREST *api.BinanceRESTClient
	DexClient   *api.DexClient
	WSConnected bool

	// Terminal Dimensions
	Width  int
	Height int
}

// NewUIModel creates and initializes the main UI model.
func NewUIModel(cfg *model.Config) UIModel {
	// Collect initial Binance symbols for WS
	var binanceSymbols []string
	for _, item := range cfg.Watchlist {
		if item.Source == model.SourceBinance {
			binanceSymbols = append(binanceSymbols, item.Symbol)
		}
	}

	wsClient := api.NewBinanceWSClient(binanceSymbols)
	restClient := api.NewBinanceRESTClient(cfg.BinanceAPIKey, cfg.BinanceAPISecret)
	dexClient := api.NewDexClient()

	fi := textinput.New()
	fi.Placeholder = "Type to filter tokens..."
	fi.CharLimit = 32
	fi.Width = 30

	return UIModel{
		Cfg:          cfg,
		MarketData:   make(map[string]model.MarketData),
		ActiveTab:    ViewWatchlist,
		SortMode:     0,
		FilterInput:  fi,
		BinanceWS:    wsClient,
		BinanceREST:  restClient,
		DexClient:    dexClient,
		AddModal:     NewAddModal(),
		ConfigModal:  NewConfigModal(cfg),
		Width:        100,
		Height:       30,
	}
}

// Init starts background streams and timers.
func (m UIModel) Init() tea.Cmd {
	m.BinanceWS.Start()

	return tea.Batch(
		waitForWSUpdates(m.BinanceWS.Updates()),
		waitForWSStatus(m.BinanceWS.Status()),
		fetchInitialBinanceData(m.BinanceREST, m.Cfg.Watchlist),
		fetchDEXData(m.DexClient, m.Cfg.Watchlist),
		fetchTrendingMemes(m.DexClient),
		scheduleDexTick(),
		m.fetchPortfolioCmd(),
	)
}

// Update handles events and input.
func (m UIModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height

	case BatchMarketDataMsg:
		for _, md := range msg {
			existing, found := m.MarketData[md.Symbol]
			if found {
				history := existing.PriceHistory
				if len(history) >= 20 {
					history = history[1:]
				}
				history = append(history, md.Price)
				md.PriceHistory = history
			} else {
				md.PriceHistory = []float64{md.Price}
			}
			m.MarketData[md.Symbol] = md
		}

	case MarketDataMsg:
		md := model.MarketData(msg)
		existing, found := m.MarketData[md.Symbol]
		if found {
			// Keep price history for sparkline
			history := existing.PriceHistory
			if len(history) >= 20 {
				history = history[1:]
			}
			history = append(history, md.Price)
			md.PriceHistory = history

			// Track tick direction
			if md.Price > existing.Price {
				md.LastTickDirection = 1
			} else if md.Price < existing.Price {
				md.LastTickDirection = -1
			} else {
				md.LastTickDirection = 0
			}
		} else {
			md.PriceHistory = []float64{md.Price}
		}
		m.MarketData[md.Symbol] = md

		// Check price alerts
		for i := range m.Cfg.Alerts {
			alert := &m.Cfg.Alerts[i]
			if !alert.Triggered && alert.Symbol == md.Symbol {
				if (alert.Direction == "ABOVE" && md.Price >= alert.TargetPrice) ||
					(alert.Direction == "BELOW" && md.Price <= alert.TargetPrice) {
					alert.Triggered = true
					_ = config.Save(m.Cfg)
					cmds = append(cmds, showToast(fmt.Sprintf("ALERT: %s hit $%.4f (%s)", alert.Symbol, md.Price, alert.Direction), m.Cfg.SoundAlerts))
				}
			}
		}

		// Keep listening for next WS update
		cmds = append(cmds, waitForWSUpdates(m.BinanceWS.Updates()))

	case WSStatusMsg:
		m.WSConnected = bool(msg)
		cmds = append(cmds, waitForWSStatus(m.BinanceWS.Status()))

	case DexTickMsg:
		cmds = append(cmds, fetchDEXData(m.DexClient, m.Cfg.Watchlist), scheduleDexTick())

	case TrendingMemesMsg:
		m.TrendingMemes = []model.MarketData(msg)

	case PortfolioMsg:
		if msg.Err != nil {
			m.PortfolioErr = msg.Err.Error()
		} else {
			m.PortfolioErr = ""
			m.PortfolioItems = msg.Items
		}

	case ToastMsg:
		m.ToastMessage = msg.Text
		if msg.IsBell {
			fmt.Print("\a") // Terminal bell sound
		}
		cmds = append(cmds, tea.Tick(4*time.Second, func(t time.Time) tea.Msg {
			return ClearToastMsg{}
		}))

	case ClearToastMsg:
		m.ToastMessage = ""

	case tea.KeyMsg:
		// Modal handling takes precedence
		if m.ShowAddModal {
			if msg.String() == "esc" {
				m.ShowAddModal = false
				return m, nil
			}
			done, item := m.AddModal.Update(msg)
			if done && item != nil {
				m.Cfg.Watchlist = append(m.Cfg.Watchlist, *item)
				_ = config.Save(m.Cfg)
				m.ShowAddModal = false

				// Update WS subscriptions
				var bnc []string
				for _, w := range m.Cfg.Watchlist {
					if w.Source == model.SourceBinance {
						bnc = append(bnc, w.Symbol)
					}
				}
				m.BinanceWS.UpdateSymbols(bnc)
				cmds = append(cmds, fetchInitialBinanceData(m.BinanceREST, m.Cfg.Watchlist))
				cmds = append(cmds, fetchDEXData(m.DexClient, m.Cfg.Watchlist))
				cmds = append(cmds, showToast(fmt.Sprintf("Added %s to watchlist", item.DisplaySymbol), false))
			}
			return m, nil
		}

		if m.ShowConfigModal {
			if msg.String() == "esc" {
				m.ShowConfigModal = false
				return m, nil
			}
			saved, _ := m.ConfigModal.Update(msg)
			if saved {
				m.ConfigModal.ApplyToConfig(m.Cfg)
				_ = config.Save(m.Cfg)
				m.BinanceREST.SetCredentials(m.Cfg.BinanceAPIKey, m.Cfg.BinanceAPISecret)
				m.ShowConfigModal = false
				cmds = append(cmds, showToast("Configuration saved!", false))
				cmds = append(cmds, m.fetchPortfolioCmd())
			}
			return m, nil
		}

		if m.ShowAlertModal {
			if msg.String() == "esc" {
				m.ShowAlertModal = false
				return m, nil
			}
			done, alert := m.AlertModal.Update(msg)
			if done && alert != nil {
				m.Cfg.Alerts = append(m.Cfg.Alerts, *alert)
				_ = config.Save(m.Cfg)
				m.ShowAlertModal = false
				cmds = append(cmds, showToast(fmt.Sprintf("Alert set for %s at $%.4f", alert.Symbol, alert.TargetPrice), false))
			}
			return m, nil
		}

		if m.ShowHelp {
			if msg.String() == "esc" || msg.String() == "?" || msg.String() == "q" {
				m.ShowHelp = false
				return m, nil
			}
			return m, nil
		}

		// Filter input mode
		if m.IsFiltering {
			switch msg.String() {
			case "esc", "enter":
				m.IsFiltering = false
				m.FilterInput.Blur()
				return m, nil
			default:
				var cmd tea.Cmd
				m.FilterInput, cmd = m.FilterInput.Update(msg)
				m.FilterText = m.FilterInput.Value()
				return m, cmd
			}
		}

		// Main navigation keys
		switch msg.String() {
		case "q", "ctrl+c":
			m.BinanceWS.Stop()
			return m, tea.Quit

		case "tab":
			m.ActiveTab = (m.ActiveTab + 1) % 4

		case "shift+tab":
			m.ActiveTab = (m.ActiveTab - 1 + 4) % 4

		case "1":
			m.ActiveTab = ViewWatchlist
		case "2":
			m.ActiveTab = ViewMemeRadar
		case "3":
			m.ActiveTab = ViewPortfolio
		case "4":
			m.ActiveTab = ViewDetail

		case "up", "k":
			if m.ActiveTab == ViewWatchlist && m.SelectedIdx > 0 {
				m.SelectedIdx--
			} else if m.ActiveTab == ViewMemeRadar && m.MemeSelectedIdx > 0 {
				m.MemeSelectedIdx--
			}

		case "down", "j":
			filtered := m.getFilteredWatchlist()
			if m.ActiveTab == ViewWatchlist && m.SelectedIdx < len(filtered)-1 {
				m.SelectedIdx++
			} else if m.ActiveTab == ViewMemeRadar && m.MemeSelectedIdx < len(m.TrendingMemes)-1 {
				m.MemeSelectedIdx++
			}

		case "enter":
			if m.ActiveTab == ViewWatchlist || m.ActiveTab == ViewMemeRadar {
				m.ActiveTab = ViewDetail
			}

		case "a":
			m.AddModal = NewAddModal()
			m.ShowAddModal = true

		case "c":
			m.ConfigModal = NewConfigModal(m.Cfg)
			m.ShowConfigModal = true

		case "!":
			// Set price alert for selected token
			filtered := m.getFilteredWatchlist()
			if len(filtered) > 0 && m.SelectedIdx < len(filtered) {
				item := filtered[m.SelectedIdx]
				md := m.MarketData[item.Symbol]
				m.AlertModal = NewAlertModal(item.DisplaySymbol, md.Price)
				m.ShowAlertModal = true
			}

		case "d", "x":
			// Delete token from watchlist
			filtered := m.getFilteredWatchlist()
			if len(filtered) > 0 && m.SelectedIdx < len(filtered) {
				target := filtered[m.SelectedIdx]
				var newWatch []model.WatchItem
				for _, w := range m.Cfg.Watchlist {
					if w.Symbol != target.Symbol {
						newWatch = append(newWatch, w)
					}
				}
				m.Cfg.Watchlist = newWatch
				_ = config.Save(m.Cfg)
				if m.SelectedIdx >= len(newWatch) && m.SelectedIdx > 0 {
					m.SelectedIdx--
				}
				cmds = append(cmds, showToast(fmt.Sprintf("Removed %s", target.DisplaySymbol), false))
			}

		case "s":
			m.SortMode = (m.SortMode + 1) % 5
			sortNames := []string{"Default", "24h % (High to Low)", "Price (High to Low)", "Volume", "Name"}
			cmds = append(cmds, showToast(fmt.Sprintf("Sorted by: %s", sortNames[m.SortMode]), false))

		case "/":
			m.IsFiltering = true
			m.FilterInput.Focus()

		case "r":
			cmds = append(cmds,
				fetchInitialBinanceData(m.BinanceREST, m.Cfg.Watchlist),
				fetchDEXData(m.DexClient, m.Cfg.Watchlist),
				fetchTrendingMemes(m.DexClient),
				m.fetchPortfolioCmd(),
				showToast("Refreshed all data!", false),
			)

		case "?", "h":
			m.ShowHelp = true
		}
	}

	return m, tea.Batch(cmds...)
}

func (m UIModel) getFilteredWatchlist() []model.WatchItem {
	var items []model.WatchItem
	filter := strings.ToUpper(strings.TrimSpace(m.FilterText))

	for _, w := range m.Cfg.Watchlist {
		if filter == "" ||
			strings.Contains(strings.ToUpper(w.DisplaySymbol), filter) ||
			strings.Contains(strings.ToUpper(w.Name), filter) ||
			strings.Contains(strings.ToUpper(string(w.Category)), filter) {
			items = append(items, w)
		}
	}

	// Apply sorting
	switch m.SortMode {
	case 1: // 24h Change Desc
		sort.SliceStable(items, func(i, j int) bool {
			return m.MarketData[items[i].Symbol].PriceChange24h > m.MarketData[items[j].Symbol].PriceChange24h
		})
	case 2: // Price Desc
		sort.SliceStable(items, func(i, j int) bool {
			return m.MarketData[items[i].Symbol].Price > m.MarketData[items[j].Symbol].Price
		})
	case 3: // Volume Desc
		sort.SliceStable(items, func(i, j int) bool {
			return m.MarketData[items[i].Symbol].Volume24h > m.MarketData[items[j].Symbol].Volume24h
		})
	case 4: // Name Asc
		sort.SliceStable(items, func(i, j int) bool {
			return items[i].DisplaySymbol < items[j].DisplaySymbol
		})
	}

	return items
}

func (m UIModel) fetchPortfolioCmd() tea.Cmd {
	return func() tea.Msg {
		if !m.BinanceREST.HasCredentials() {
			return PortfolioMsg{Items: nil, Err: nil}
		}

		priceMap := make(map[string]float64)
		for s, md := range m.MarketData {
			priceMap[s] = md.Price
			priceMap[md.DisplaySymbol] = md.Price
		}

		items, err := m.BinanceREST.FetchPortfolio(priceMap)
		return PortfolioMsg{Items: items, Err: err}
	}
}

// Background command helpers
func waitForWSUpdates(ch <-chan model.MarketData) tea.Cmd {
	return func() tea.Msg {
		md, ok := <-ch
		if !ok {
			return nil
		}
		return MarketDataMsg(md)
	}
}

func waitForWSStatus(ch <-chan bool) tea.Cmd {
	return func() tea.Msg {
		status, ok := <-ch
		if !ok {
			return nil
		}
		return WSStatusMsg(status)
	}
}

func fetchInitialBinanceData(rest *api.BinanceRESTClient, watchlist []model.WatchItem) tea.Cmd {
	return func() tea.Msg {
		var symbols []string
		for _, w := range watchlist {
			if w.Source == model.SourceBinance {
				symbols = append(symbols, w.Symbol)
			}
		}
		dataMap, err := rest.Fetch24hrTickers(symbols)
		if err != nil {
			return nil
		}
		return BatchMarketDataMsg(dataMap)
	}
}

func fetchDEXData(client *api.DexClient, watchlist []model.WatchItem) tea.Cmd {
	return func() tea.Msg {
		var addresses []string
		for _, w := range watchlist {
			if w.Source == model.SourceDEX {
				addresses = append(addresses, w.Symbol)
			}
		}
		dataMap, err := client.FetchTokens(addresses)
		if err != nil {
			return nil
		}
		return BatchMarketDataMsg(dataMap)
	}
}

func fetchTrendingMemes(client *api.DexClient) tea.Cmd {
	return func() tea.Msg {
		memes, err := client.FetchTrendingMemes()
		if err != nil {
			return nil
		}
		return TrendingMemesMsg(memes)
	}
}

func scheduleDexTick() tea.Cmd {
	return tea.Tick(5*time.Second, func(t time.Time) tea.Msg {
		return DexTickMsg{}
	})
}

func showToast(text string, bell bool) tea.Cmd {
	return func() tea.Msg {
		return ToastMsg{Text: text, IsBell: bell}
	}
}
