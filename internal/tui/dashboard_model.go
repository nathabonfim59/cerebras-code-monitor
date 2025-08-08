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
		tea.ClearScreen,
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
	styles := GetStyles()

	// Apply uniform outer padding around the entire view
	outerPadX, outerPadY := 1, 1
	innerWidth := m.width - (2 * outerPadX)
	if innerWidth < 1 {
		innerWidth = 1
	}

	// Define styles using centralized theme
	titleStyle := styles.Header.Copy().
		Width(innerWidth).
		MaxWidth(innerWidth).
		Align(lipgloss.Left)

	tabStyle := styles.TabInactive

	activeTabStyle := styles.TabActive

	statusBarStyle := styles.StatusBar.Copy().
		Width(innerWidth).
		MaxWidth(innerWidth)

	// Calculate content height to fill the screen within outer padding.
	// 1 line for title, 1 blank spacer line, 1 line for tabs, 1 line for status bar => 4 lines.
	innerHeight := m.height - (2 * outerPadY)
	contentHeight := innerHeight - 4
	if contentHeight < 1 {
		contentHeight = 1
	}

	contentStyle := styles.Content.Copy().
		Width(innerWidth).
		MaxWidth(innerWidth).
		Height(contentHeight).
		Align(lipgloss.Left)

	// Render header with specific coloring: icon + "Cerebras" in primary, rest in subtle.
    // Use Header.Copy() for segments so they keep the same background and do not reset it.
    brandPrimary := styles.Header.Copy().Foreground(styles.Palette.Primary)
    headerSubtle := styles.Header.Copy().Foreground(styles.Palette.Subtle)

    iconText := brandPrimary.Render(fmt.Sprintf(" %s", icons.Dashboard)) // leading space before icon
    brandText := brandPrimary.Render(" Cerebras")
    restText := headerSubtle.Render(" Code Monitor Dashboard")
    headerLine := lipgloss.JoinHorizontal(lipgloss.Left, iconText, brandText, restText)
    title := titleStyle.Render(headerLine)

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

	// Combine all elements (add a blank spacer line after the header)
	spacer := ""
	view := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		spacer,
		tabRow,
		contentStyle.Render(content),
		statusBar,
	)

	// Wrap the entire view with outer padding
	outer := lipgloss.NewStyle().Padding(outerPadY, outerPadX)
	return outer.Render(view)
}

// renderDashboard renders the main dashboard content
func (m DashboardModel) renderDashboard() string {
    icons := config.GetIcons()
    styles := GetStyles()

    if m.err != nil {
        return styles.Error.Render(fmt.Sprintf("Error: %v", m.err))
    }

    if m.metrics == nil {
        return fmt.Sprintf("%s Loading metrics...", icons.Info)
    }

    // Helpers to compute used counts with GraphQL-first fallback strategy
    usedFrom := func(usage, limit, remaining int64) int64 {
        if usage > 0 {
            return usage
        }
        if limit > 0 && remaining >= 0 {
            u := limit - remaining
            if u < 0 {
                u = 0
            }
            return u
        }
        return 0
    }

    // Layout measurements
    available := m.width - 2 // matches outer padding in View()
    if available < 40 {
        available = 40
    }
    gap := 2
    twoCols := available >= 80
    colW := available
    if twoCols {
        colW = (available - gap) / 2
    }

    // Base styles (muted, no backgrounds except progress bars)
    title := styles.SectionTitle
    label := lipgloss.NewStyle().Foreground(styles.Palette.Subtle)
    value := lipgloss.NewStyle().Foreground(styles.Palette.Text).Bold(true)
    dim := lipgloss.NewStyle().Foreground(styles.Palette.Subtle)

    // Progress rows
    // Layout: first line has label; second line has a full-width bar; third line has left-aligned percentage and right-aligned (used/limit).
    lblW := 16
    // Make the progress bar span the entire card column width including brackets
    // createProgressBar wraps content in [ and ], so subtract 2 to avoid wrapping
    barW := colW - 2
    if barW < 10 {
        barW = 10
    }

    // Helper to render a metric block (title, bar, stats) or Unknown when limit is missing
    renderMetric := func(icon, name string, used, limit int64) []string {
        titleRow := label.Width(lblW).Render(fmt.Sprintf("%s %s", icon, name))
        if limit <= 0 {
            return []string{titleRow, dim.Render("Unknown")}
        }
        percent := 0.0
        if limit > 0 {
            percent = float64(used) / float64(limit) * 100
        }
        bar := m.createProgressBar(percent, barW)
        left := value.Render(fmt.Sprintf("%.1f%%", percent))
        right := value.Render(fmt.Sprintf("(%d/%d)", used, limit))
        middle := colW - lipgloss.Width(left) - lipgloss.Width(right)
        if middle < 1 {
            middle = 1
        }
        stats := lipgloss.JoinHorizontal(lipgloss.Top, left, strings.Repeat(" ", middle), right)
        return []string{titleRow, bar, stats}
    }

    // Compute used/limits for all relevant metrics
    rpmLimit := m.metrics.LimitRequestsMinute
    rphLimit := m.metrics.LimitRequestsHour
    rpdLimit := m.metrics.LimitRequestsDay
    tpmLimit := m.metrics.LimitTokensMinute
    tphLimit := m.metrics.LimitTokensHour
    tpdLimit := m.metrics.LimitTokensDay

    rpmUsed := usedFrom(m.metrics.UsageRequestsMinute, rpmLimit, m.metrics.RemainingRequestsMinute)
    rphUsed := usedFrom(m.metrics.UsageRequestsHour, rphLimit, m.metrics.RemainingRequestsHour)
    rpdUsed := usedFrom(m.metrics.UsageRequestsDay, rpdLimit, m.metrics.RemainingRequestsDay)
    tpmUsed := usedFrom(m.metrics.UsageTokensMinute, tpmLimit, m.metrics.RemainingTokensMinute)
    tphUsed := usedFrom(m.metrics.UsageTokensHour, tphLimit, m.metrics.RemainingTokensHour)
    tpdUsed := usedFrom(m.metrics.UsageTokensDay, tpdLimit, m.metrics.RemainingTokensDay)

    // Card 1: Rate Limits (Requests & Tokens across minute/hour/day)
    // Insert a blank line between metrics when in vertical (single column) layout
    cardRows := []string{title.Render("Rate Limits")}
    addMetric := func(rows []string) {
        cardRows = append(cardRows, rows...)
        if !twoCols {
            cardRows = append(cardRows, "")
        }
    }
    addMetric(renderMetric(icons.Request, "Requests / Minute", rpmUsed, rpmLimit))
    addMetric(renderMetric(icons.Request, "Requests / Hour", rphUsed, rphLimit))
    addMetric(renderMetric(icons.Request, "Requests / Day", rpdUsed, rpdLimit))
    addMetric(renderMetric(icons.Token, "Tokens / Minute", tpmUsed, tpmLimit))
    addMetric(renderMetric(icons.Token, "Tokens / Hour", tphUsed, tphLimit))
    addMetric(renderMetric(icons.Token, "Tokens / Day", tpdUsed, tpdLimit))
    // Trim trailing blank in two-column mode already avoided; in vertical it's fine to end with a blank
    card1 := lipgloss.JoinVertical(lipgloss.Left, cardRows...)

    // Card 2: Quotas/Remaining
    // Define rows with label on left and value on right
    lr := func(k string, v string) string {
        return lipgloss.JoinHorizontal(lipgloss.Top,
            label.Width(lblW).Render(k),
            dim.Render(""),
            value.Width(colW-lblW).Align(lipgloss.Right).Render(v),
        )
    }

    // Daily request reset string
    dailyReset := "Unknown"
    if m.metrics.ResetRequestsDay > 0 {
        resetDaily := time.Now().Add(time.Duration(m.metrics.ResetRequestsDay) * time.Second)
        hours := int(m.metrics.ResetRequestsDay / 3600)
        mins := int((m.metrics.ResetRequestsDay % 3600) / 60)
        dailyReset = fmt.Sprintf("%s  (%dh %dm)", resetDaily.Format("15:04"), hours, mins)
    }
    // Minute token reset string
    minuteReset := "Unknown"
    if m.metrics.ResetTokensMinute > 0 {
        resetMinute := time.Now().Add(time.Duration(m.metrics.ResetTokensMinute) * time.Second)
        minuteReset = fmt.Sprintf("%s  (%ds)", resetMinute.Format("15:04:05"), int(m.metrics.ResetTokensMinute))
    }

    card2 := lipgloss.JoinVertical(lipgloss.Left,
        title.Render("Quotas & Remaining"),
        lr("Daily Limit", fmt.Sprintf("%d", m.metrics.LimitRequestsDay)),
        lr("Daily Remaining", fmt.Sprintf("%d", m.metrics.RemainingRequestsDay)),
        lr("Daily Reset", dailyReset),
        "",
        lr("Minute Limit", fmt.Sprintf("%d", m.metrics.LimitTokensMinute)),
        lr("Minute Remaining", fmt.Sprintf("%d", m.metrics.RemainingTokensMinute)),
        lr("Minute Reset", minuteReset),
    )

    var content string
    if twoCols {
        left := lipgloss.NewStyle().Width(colW).Render(card1)
        right := lipgloss.NewStyle().Width(colW).Render(card2)
        content = lipgloss.JoinHorizontal(lipgloss.Top, left, strings.Repeat(" ", gap), right)
    } else {
        content = lipgloss.JoinVertical(lipgloss.Left, card1, "", card2)
    }

    return content
}

// renderUsage renders the usage tab content
func (m DashboardModel) renderUsage() string {
    icons := config.GetIcons()
    styles := GetStyles()

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
	s.WriteString(styles.SectionTitle.Render("Usage Statistics") + "\n\n")

	headerStyle := styles.TableHeader
	valueStyle := styles.TableCell

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
	styles := GetStyles()

	var s strings.Builder
	s.WriteString(styles.SectionTitle.Render("Quotas Information") + "\n\n")
	s.WriteString(fmt.Sprintf("%s Detailed quota information will be displayed here.\n", icons.Info))
	s.WriteString(fmt.Sprintf("%s Currently showing basic rate limits from the Dashboard tab.\n", icons.Info))

	return s.String()
}

// renderSettings renders the settings tab content
func (m DashboardModel) renderSettings() string {
    icons := config.GetIcons()
	styles := GetStyles()

	var s strings.Builder
	s.WriteString(styles.SectionTitle.Render("Settings") + "\n\n")
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
	if width <= 0 {
		return ""
	}
	if percent < 0 {
		percent = 0
	}
	if percent > 100 {
		percent = 100
	}
	filled := int(percent / 100 * float64(width))
	if filled > width {
		filled = width
	}
	empty := width - filled

	filledStr := strings.Repeat("█", filled)
	emptyStr := strings.Repeat("░", empty)

	// color the filled part; prefer cerebras primaryColor for safe ranges under 75, else status color
	fillColor := m.getStatusColor(percent)
	if percent <= 75 {
		fillColor = primaryColor
	}

	styles := GetStyles()
	filledStyled := lipgloss.NewStyle().Foreground(fillColor).Render(filledStr)
	emptyStyled := styles.ProgressEmpty.Render(emptyStr)

	return "[" + filledStyled + emptyStyled + "]"
}

// getStatusColor returns a color based on usage percentage
func (m DashboardModel) getStatusColor(percent float64) lipgloss.Color {
	switch {
	case percent > 90:
		return lipgloss.Color("#ff4d4d") // Red
	case percent > 75:
		return lipgloss.Color("#ff9900") // Orange
	case percent > 50:
		return lipgloss.Color("#ffff00") // Yellow
	default:
		return primaryColor // Use cerebras color for healthy usage
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
