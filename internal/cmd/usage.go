package cmd

import (
	"fmt"

	"github.com/nathabonfim59/cerebras-code-monitor/internal/cerebras"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var UsageCmd = &cobra.Command{
	Use:   "usage",
	Short: "Track usage",
	Long:  "Commands to track and display usage statistics for organizations",
}

var getUsageCmd = &cobra.Command{
	Use:   "get [organization]",
	Short: "Get usage statistics for an organization",
	Args:  cobra.MaximumNArgs(1), // Make organization optional
	Run: func(cmd *cobra.Command, args []string) {
		// If organization is not provided, use the one from configuration
		organization := ""
		if len(args) > 0 {
			organization = args[0]
		} else {
			// Get organization ID from configuration/viper
			organization = viper.GetString("org-id")
		}

		if organization == "" {
			fmt.Println("Error: organization must be provided either as an argument or via --org-id flag")
			return
		}

		// Create Cerebras client
		client := cerebras.NewClient()
		if !client.HasAuth() {
			fmt.Println("Error: No authentication method configured. Please login first.")
			return
		}

		fmt.Printf("Getting usage statistics for organization %s...\n", organization)
		metrics, err := client.GetMetrics(organization)
		if err != nil {
			fmt.Printf("Error fetching metrics: %v\n", err)
			return
		}

		// Display metrics
		fmt.Printf("Rate Limit Metrics:\n")
		fmt.Printf("  Daily Request Limit: %d\n", metrics.LimitRequestsDay)
		fmt.Printf("  Daily Requests Remaining: %d\n", metrics.RemainingRequestsDay)
		fmt.Printf("  Daily Request Reset Time: %d seconds\n", metrics.ResetRequestsDay)
		fmt.Printf("  Minute Token Limit: %d\n", metrics.LimitTokensMinute)
		fmt.Printf("  Minute Tokens Remaining: %d\n", metrics.RemainingTokensMinute)
		fmt.Printf("  Minute Token Reset Time: %d seconds\n", metrics.ResetTokensMinute)
	},
}

var monitorUsageCmd = &cobra.Command{
	Use:   "monitor",
	Short: "Start real-time monitoring of usage",
	Run: func(cmd *cobra.Command, args []string) {
		// Get organization ID from configuration/viper
		organization := viper.GetString("org-id")
		if organization == "" {
			fmt.Println("Error: organization ID must be set via --org-id flag or configuration")
			return
		}

		// Get model from configuration/viper
		model := viper.GetString("model")

		// Get refresh rate from configuration/viper
		refreshRate := viper.GetInt("refresh-rate")

		// Create Cerebras client
		client := cerebras.NewClient()
		if !client.HasAuth() {
			fmt.Println("Error: No authentication method configured. Please login first.")
			return
		}

		fmt.Printf("Starting real-time monitoring for organization %s (model: %s, refresh: %ds)...\n", organization, model, refreshRate)
		// TODO: Implement real-time monitoring logic
		// This would continuously make requests to the Cerebras GraphQL endpoint
		// and display usage information with color-coded progress bars and tables
	},
}

func init() {
	UsageCmd.AddCommand(getUsageCmd)
	UsageCmd.AddCommand(monitorUsageCmd)
}
