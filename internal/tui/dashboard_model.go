package tui

import (
	"fmt"
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
	quitting     bool
}

// NewDashboardModel creates a new dashboard model
func NewDashboardModel(client *cerebras.Client, organization, modelName string, refreshRate int) DashboardModel {
	return DashboardModel{
		client:       client,
		organization: organization,
		modelName:    modelName,
		refreshRate:  refreshRate,
		tabs:         []string{"Realtime"},
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
	}

	return m, nil
}

// View renders the model
func (m DashboardModel) View() string {
	if m.quitting {
		return "Quitting dashboard...\n"
	}

	icons := config.GetIcons()

	// Apply primary color styling to the dashboard title
	titleStyle := lipgloss.NewStyle().Foreground(primaryColor).Bold(true)
	s := titleStyle.Render(fmt.Sprintf("%s Cerebras Code Monitor Dashboard\n\n", icons.Dashboard))

	// Render tabs with primary color styling
	tabStyle := lipgloss.NewStyle().Foreground(primaryColor)
	activeTabStyle := lipgloss.NewStyle().Foreground(primaryColor).Bold(true)

	for i, tab := range m.tabs {
		if m.activeTab == i {
			s += activeTabStyle.Render(fmt.Sprintf("> %s <", tab))
		} else {
			s += tabStyle.Render(fmt.Sprintf("  %s  ", tab))
		}
		if i < len(m.tabs)-1 {
			s += " | "
		}
	}
	s += "\n\n"

	// Render metrics if available
	if m.metrics != nil {
		s += fmt.Sprintf("%s Organization: %s\n", icons.Organization, m.organization)
		s += fmt.Sprintf("%s Model: %s\n", icons.Model, m.modelName)
		s += fmt.Sprintf("%s Refresh Rate: %d seconds\n\n", icons.Time, m.refreshRate)

		s += fmt.Sprintf("%s Rate Limits:\n", icons.Request)
		s += fmt.Sprintf("  %s Daily Request Limit: %d\n", icons.Request, m.metrics.LimitRequestsDay)
		s += fmt.Sprintf("  %s Daily Requests Remaining: %d\n", icons.Request, m.metrics.RemainingRequestsDay)
		if m.metrics.ResetRequestsDay > 0 {
			resetDaily := time.Now().Add(time.Duration(m.metrics.ResetRequestsDay) * time.Second)
			hoursUntilReset := int(m.metrics.ResetRequestsDay / 3600)
			minutesUntilReset := int((m.metrics.ResetRequestsDay % 3600) / 60)
			s += fmt.Sprintf("  %s Daily Request Reset: %s (%d hours %d minutes)\n", icons.Time, resetDaily.Format(time.RFC3339), hoursUntilReset, minutesUntilReset)
		} else {
			s += fmt.Sprintf("  %s Daily Request Reset: Unknown\n", icons.Warning)
		}

		s += "\n"
		s += fmt.Sprintf("  %s Minute Token Limit: %d\n", icons.Token, m.metrics.LimitTokensMinute)
		s += fmt.Sprintf("  %s Minute Tokens Remaining: %d\n", icons.Token, m.metrics.RemainingTokensMinute)
		if m.metrics.ResetTokensMinute > 0 {
			resetMinute := time.Now().Add(time.Duration(m.metrics.ResetTokensMinute) * time.Second)
			secondsUntilReset := int(m.metrics.ResetTokensMinute)
			s += fmt.Sprintf("  %s Minute Token Reset: %s (%d seconds)\n", icons.Time, resetMinute.Format(time.RFC3339), secondsUntilReset)
		} else {
			s += fmt.Sprintf("  %s Minute Token Reset: Unknown\n", icons.Warning)
		}

		// Calculate usage percentages
		if m.metrics.LimitRequestsDay > 0 {
			requestsUsed := m.metrics.LimitRequestsDay - m.metrics.RemainingRequestsDay
			requestsPercent := float64(requestsUsed) / float64(m.metrics.LimitRequestsDay) * 100
			statusIcon := icons.Check
			if requestsPercent > 90 {
				statusIcon = icons.Error
			} else if requestsPercent > 75 {
				statusIcon = icons.Warning
			}
			s += fmt.Sprintf("\n%s Daily Requests Usage: %.1f%% (%d/%d)\n", statusIcon, requestsPercent, requestsUsed, m.metrics.LimitRequestsDay)
		}

		if m.metrics.LimitTokensMinute > 0 {
			tokensUsed := m.metrics.LimitTokensMinute - m.metrics.RemainingTokensMinute
			tokensPercent := float64(tokensUsed) / float64(m.metrics.LimitTokensMinute) * 100
			statusIcon := icons.Check
			if tokensPercent > 90 {
				statusIcon = icons.Error
			} else if tokensPercent > 75 {
				statusIcon = icons.Warning
			}
			s += fmt.Sprintf("%s Minute Tokens Usage: %.1f%% (%d/%d)\n", statusIcon, tokensPercent, tokensUsed, m.metrics.LimitTokensMinute)
		}
	} else if m.err != nil {
		s += fmt.Sprintf("%s Error: %v\n", icons.Error, m.err)
	} else {
		s += fmt.Sprintf("%s Loading metrics...\n", icons.Info)
	}

	s += fmt.Sprintf("\n%s Controls:\n", icons.Settings)
	s += fmt.Sprintf("  q/ctrl+c: Quit %s\n", icons.Error)
	s += fmt.Sprintf("  tab: Switch tabs %s\n", icons.Theme)
	s += fmt.Sprintf("  r: Refresh data %s\n", icons.Refresh)

	return s
}
