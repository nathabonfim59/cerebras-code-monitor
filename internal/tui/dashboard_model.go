package tui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nathabonfim59/cerebras-code-monitor/internal/cerebras"
	"github.com/nathabonfim59/cerebras-code-monitor/internal/config"
)

// DashboardModel represents the model for the dashboard
type DashboardModel struct {
	client       *cerebras.Client
	organization string
	modelName    string
	refreshRate  int
	metrics      *cerebras.RateLimitInfo
	err          error
	tabs         []string
	activeTab    int
	width        int
	height       int
	quitting     bool
}

// NewDashboardModel creates a new dashboard model
func NewDashboardModel(client *cerebras.Client, organization, modelName string, refreshRate int) DashboardModel {
	return DashboardModel{
		client:       client,
		organization: organization,
		modelName:    modelName,
		refreshRate:  refreshRate,
		tabs:         []string{"Dashboard", "Usage", "Quotas", "Settings"},
		activeTab:    0,
	}
}

// Init initializes the model
func (m DashboardModel) Init() tea.Cmd {
	// Start the ticker for refreshing data
	return tea.Batch(
		m.fetchMetrics(),
		tea.Tick(time.Duration(m.refreshRate)*time.Second, func(t time.Time) tea.Msg {
			return tickMsg(t)
		}),
	)
}

// tickMsg represents a tick message
type tickMsg time.Time

// fetchMetrics fetches metrics from the Cerebras API
func (m DashboardModel) fetchMetrics() tea.Cmd {
	return func() tea.Msg {
		metrics, err := m.client.GetMetrics(m.organization)
		if err != nil {
			return errMsg{err}
		}
		return metricsMsg{metrics}
	}
}

// metricsMsg represents a metrics message
type metricsMsg struct {
	metrics *cerebras.RateLimitInfo
}

// errMsg represents an error message
type errMsg struct {
	err error
}

// Update handles messages in the model
func (m DashboardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit
		case "tab":
			m.activeTab = (m.activeTab + 1) % len(m.tabs)
			return m, nil
		case "r":
			// Refresh data immediately
			return m, m.fetchMetrics()
		}
	case tickMsg:
		// Refresh data on tick
		return m, tea.Batch(
			m.fetchMetrics(),
			tea.Tick(time.Duration(m.refreshRate)*time.Second, func(t time.Time) tea.Msg {
				return tickMsg(t)
			}),
		)
	case metricsMsg:
		m.metrics = msg.metrics
		return m, nil
	case errMsg:
		m.err = msg.err
		return m, nil
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	}

	return m, nil
}

// View renders the model
func (m DashboardModel) View() string {
	if m.quitting {
		return "\nQuitting dashboard...\n"
	}

	// If we don't have dimensions yet, return a placeholder
	if m.width == 0 || m.height == 0 {
		return "Initializing..."
	}

	icons := config.GetIcons()

	// Define styles
	titleStyle := lipgloss.NewStyle().
		Foreground(primaryColor).
		Bold(true).
		Align(lipgloss.Center).
		Width(m.width)

	tabStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#7D7D7D")).
		Background(lipgloss.Color("#1a1a1a")).
		Padding(0, 2).
		MarginRight(1)

	activeTabStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("white")).
		Background(primaryColor).
		Padding(0, 2).
		MarginRight(1).
		Bold(true)

	statusBarStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#ffffff")).
		Background(lipgloss.Color("#2a2a2a")).
		Padding(0, 1).
		Width(m.width)

	contentStyle := lipgloss.NewStyle().
		Width(m.width).
		Height(m.height-4).
		Padding(1, 2)

	// Render header with title
	title := titleStyle.Render(fmt.Sprintf("%s Cerebras Code Monitor Dashboard", icons.Dashboard))

	// Render tabs
	var tabs []string
	for i, tab := range m.tabs {
		if m.activeTab == i {
			tabs = append(tabs, activeTabStyle.Render(tab))
		} else {
			tabs = append(tabs, tabStyle.Render(tab))
		}
	}
	tabRow := lipgloss.JoinHorizontal(lipgloss.Top, tabs...)

	// Render content based on active tab
	var content string
	switch m.tabs[m.activeTab] {
	case "Dashboard":
		content = m.renderDashboard()
	case "Usage":
		content = m.renderUsage()
	case "Quotas":
		content = m.renderQuotas()
	case "Settings":
		content = m.renderSettings()
	default:
		content = "Unknown tab"
	}

	// Render status bar
	statusBar := statusBarStyle.Render(
		fmt.Sprintf("%s Organization: %s | %s Model: %s | %s Refresh: %ds",
			icons.Organization, m.organization,
			icons.Model, m.modelName,
			icons.Time, m.refreshRate))

	// Combine all elements
	view := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		tabRow,
		contentStyle.Render(content),
		statusBar,
	)

	return view
}

// renderDashboard renders the main dashboard content
func (m DashboardModel) renderDashboard() string {
	icons := config.GetIcons()

	if m.err != nil {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ff0000")).
			Render(fmt.Sprintf("Error: %v", m.err))
	}

	if m.metrics == nil {
		return fmt.Sprintf("%s Loading metrics...", icons.Info)
	}

	// Calculate usage percentages
	var requestsUsed, requestsPercent, tokensUsed, tokensPercent float64
	if m.metrics.LimitRequestsDay > 0 {
		requestsUsed = float64(m.metrics.LimitRequestsDay - m.metrics.RemainingRequestsDay)
		requestsPercent = requestsUsed / float64(m.metrics.LimitRequestsDay) * 100
	}

	if m.metrics.LimitTokensMinute > 0 {
		tokensUsed = float64(m.metrics.LimitTokensMinute - m.metrics.RemainingTokensMinute)
		tokensPercent = tokensUsed / float64(m.metrics.LimitTokensMinute) * 100
	}

	// Create progress bars
	requestsBar := m.createProgressBar(requestsPercent, 30)
	tokensBar := m.createProgressBar(tokensPercent, 30)

	// Determine status colors
	requestsStatusColor := m.getStatusColor(requestsPercent)
	tokensStatusColor := m.getStatusColor(tokensPercent)

	// Style the metrics display
	metricStyle := lipgloss.NewStyle().MarginBottom(1)
	requestsStyle := lipgloss.NewStyle().Foreground(requestsStatusColor)
	tokensStyle := lipgloss.NewStyle().Foreground(tokensStatusColor)

	// Render dashboard content
	var s strings.Builder
	s.WriteString(lipgloss.NewStyle().Bold(true).Foreground(primaryColor).Render("Rate Limit Overview\n\n"))

	s.WriteString(requestsStyle.Render(fmt.Sprintf(
		"%s Daily Requests: %s %.1f%% (%.0f/%d)",
		icons.Request,
		requestsBar,
		requestsPercent,
		requestsUsed,
		m.metrics.LimitRequestsDay,
	)))
	s.WriteString("\n")

	s.WriteString(tokensStyle.Render(fmt.Sprintf(
		"%s Minute Tokens: %s %.1f%% (%.0f/%d)",
		icons.Token,
		tokensBar,
		tokensPercent,
		tokensUsed,
		m.metrics.LimitTokensMinute,
	)))
	s.WriteString("\n\n")

	// Add detailed metrics
	s.WriteString(metricStyle.Render(fmt.Sprintf("%s Daily Request Limit: %d", icons.Request, m.metrics.LimitRequestsDay)))
	s.WriteString("\n")
	s.WriteString(metricStyle.Render(fmt.Sprintf("%s Daily Requests Remaining: %d", icons.Request, m.metrics.RemainingRequestsDay)))
	s.WriteString("\n")

	if m.metrics.ResetRequestsDay > 0 {
		resetDaily := time.Now().Add(time.Duration(m.metrics.ResetRequestsDay) * time.Second)
		hoursUntilReset := int(m.metrics.ResetRequestsDay / 3600)
		minutesUntilReset := int((m.metrics.ResetRequestsDay % 3600) / 60)
		s.WriteString(metricStyle.Render(fmt.Sprintf(
			"%s Daily Request Reset: %s (%d hours %d minutes)",
			icons.Time,
			resetDaily.Format("15:04:05"),
			hoursUntilReset,
			minutesUntilReset,
		)))
	} else {
		s.WriteString(metricStyle.Render(fmt.Sprintf("%s Daily Request Reset: Unknown", icons.Warning)))
	}
	s.WriteString("\n\n")

	s.WriteString(metricStyle.Render(fmt.Sprintf("%s Minute Token Limit: %d", icons.Token, m.metrics.LimitTokensMinute)))
	s.WriteString("\n")
	s.WriteString(metricStyle.Render(fmt.Sprintf("%s Minute Tokens Remaining: %d", icons.Token, m.metrics.RemainingTokensMinute)))
	s.WriteString("\n")

	if m.metrics.ResetTokensMinute > 0 {
		resetMinute := time.Now().Add(time.Duration(m.metrics.ResetTokensMinute) * time.Second)
		secondsUntilReset := int(m.metrics.ResetTokensMinute)
		s.WriteString(metricStyle.Render(fmt.Sprintf(
			"%s Minute Token Reset: %s (%d seconds)",
			icons.Time,
			resetMinute.Format("15:04:05"),
			secondsUntilReset,
		)))
	} else {
		s.WriteString(metricStyle.Render(fmt.Sprintf("%s Minute Token Reset: Unknown", icons.Warning)))
	}

	return s.String()
}

// renderUsage renders the usage tab content
func (m DashboardModel) renderUsage() string {
	icons := config.GetIcons()

	if m.metrics == nil {
		return fmt.Sprintf("%s Loading usage data...", icons.Info)
	}

	// Calculate usage
	var requestsUsed float64
	if m.metrics.LimitRequestsDay > 0 {
		requestsUsed = float64(m.metrics.LimitRequestsDay - m.metrics.RemainingRequestsDay)
	}

	var tokensUsed float64
	if m.metrics.LimitTokensMinute > 0 {
		tokensUsed = float64(m.metrics.LimitTokensMinute - m.metrics.RemainingTokensMinute)
	}

	// Create a table-like view for usage data
	var s strings.Builder
	s.WriteString(lipgloss.NewStyle().Bold(true).Foreground(primaryColor).Render("Usage Statistics\n\n"))

	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#ffffff")).PaddingRight(2)
	valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#a0a0a0")).PaddingRight(2)

	s.WriteString(headerStyle.Render("Metric") + headerStyle.Render("Used") + headerStyle.Render("Limit") + headerStyle.Render("Reset") + "\n")
	s.WriteString(valueStyle.Render("Daily Requests") +
		valueStyle.Render(fmt.Sprintf("%.0f", requestsUsed)) +
		valueStyle.Render(fmt.Sprintf("%d", m.metrics.LimitRequestsDay)) +
		valueStyle.Render(m.formatResetTime(m.metrics.ResetRequestsDay)) + "\n")
	s.WriteString(valueStyle.Render("Minute Tokens") +
		valueStyle.Render(fmt.Sprintf("%.0f", tokensUsed)) +
		valueStyle.Render(fmt.Sprintf("%d", m.metrics.LimitTokensMinute)) +
		valueStyle.Render(m.formatResetTime(m.metrics.ResetTokensMinute)) + "\n")

	return s.String()
}

// renderQuotas renders the quotas tab content
func (m DashboardModel) renderQuotas() string {
	icons := config.GetIcons()

	var s strings.Builder
	s.WriteString(lipgloss.NewStyle().Bold(true).Foreground(primaryColor).Render("Quotas Information\n\n"))
	s.WriteString(fmt.Sprintf("%s Detailed quota information will be displayed here.\n", icons.Info))
	s.WriteString(fmt.Sprintf("%s Currently showing basic rate limits from the Dashboard tab.\n", icons.Info))

	return s.String()
}

// renderSettings renders the settings tab content
func (m DashboardModel) renderSettings() string {
	icons := config.GetIcons()

	var s strings.Builder
	s.WriteString(lipgloss.NewStyle().Bold(true).Foreground(primaryColor).Render("Settings\n\n"))
	s.WriteString(fmt.Sprintf("%s Refresh Rate: %d seconds\n", icons.Time, m.refreshRate))
	s.WriteString(fmt.Sprintf("%s Organization: %s\n", icons.Organization, m.organization))
	s.WriteString(fmt.Sprintf("%s Model: %s\n\n", icons.Model, m.modelName))
	s.WriteString("Available controls:\n")
	s.WriteString(fmt.Sprintf("  %s q/ctrl+c: Quit\n", icons.Error))
	s.WriteString(fmt.Sprintf("  %s tab: Switch tabs\n", icons.Theme))
	s.WriteString(fmt.Sprintf("  %s r: Refresh data\n", icons.Refresh))

	return s.String()
}

// createProgressBar creates a visual progress bar
func (m DashboardModel) createProgressBar(percent float64, width int) string {
	filled := int(percent / 100 * float64(width))
	empty := width - filled

	barStyle := lipgloss.NewStyle().Background(lipgloss.Color("#444444"))
	filledStyle := lipgloss.NewStyle().Background(m.getStatusColor(percent))

	filledBar := filledStyle.Render(strings.Repeat(" ", filled))
	emptyBar := barStyle.Render(strings.Repeat(" ", empty))

	return "[" + filledBar + emptyBar + "]"
}

// getStatusColor returns a color based on usage percentage
func (m DashboardModel) getStatusColor(percent float64) lipgloss.Color {
	switch {
	case percent > 90:
		return lipgloss.Color("#ff0000") // Red
	case percent > 75:
		return lipgloss.Color("#ff9900") // Orange
	case percent > 50:
		return lipgloss.Color("#ffff00") // Yellow
	default:
		return lipgloss.Color("#00ff00") // Green
	}
}

// formatResetTime formats the reset time in a human-readable way
func (m DashboardModel) formatResetTime(seconds int64) string {
	if seconds <= 0 {
		return "Unknown"
	}

	if seconds < 60 {
		return fmt.Sprintf("%ds", seconds)
	}

	if seconds < 3600 {
		minutes := seconds / 60
		return fmt.Sprintf("%dm", minutes)
	}

	hours := seconds / 3600
	minutes := (seconds % 3600) / 60
	return fmt.Sprintf("%dh%dm", hours, minutes)
}
