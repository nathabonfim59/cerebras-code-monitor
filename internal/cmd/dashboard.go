package cmd

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/nathabonfim59/cerebras-code-monitor/internal/cerebras"
	"github.com/nathabonfim59/cerebras-code-monitor/internal/tui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var DashboardCmd = &cobra.Command{
	Use:   "dashboard",
	Short: "Open the TUI dashboard",
	Long:  "Open a real-time dashboard with bubbletea TUI to monitor Cerebras AI usage",
	Run: func(cmd *cobra.Command, args []string) {
		// Get organization ID from configuration/viper
		organization := viper.GetString("org-id")

		// Get model from configuration/viper
		modelName := viper.GetString("model")
		if modelName == "" {
			modelName = "qwen-3-coder-480b"
		}

		// Get refresh rate from configuration/viper
		refreshRate := viper.GetInt("refresh-rate")
		if refreshRate < 1 {
			refreshRate = 10 // Default to 10 seconds
		}

		// Create Cerebras client
		client := cerebras.NewClient()
		if !client.HasAuth() {
			fmt.Println("Error: No authentication method configured. Please login first.")
			return
		}

		// For session token auth, organization is required
		// Only require organization if we're using session token auth (not API key auth)
		if client.SessionToken() != "" && client.APIKey() == "" && organization == "" {
			fmt.Println("Error: organization ID must be set via --org-id flag or configuration when using session token authentication")
			return
		}

		// Create and run the dashboard model
		dashboardModel := tui.NewDashboardModel(client, organization, modelName, refreshRate)
		p := tea.NewProgram(dashboardModel)
		if _, err := p.Run(); err != nil {
			fmt.Printf("Error running dashboard: %v\n", err)
			return
		}
	},
}
