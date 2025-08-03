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

		// Create Cerebras client
		client := cerebras.NewClient()
		if !client.HasAuth() {
			fmt.Println("Error: No authentication method configured. Please login first.")
			return
		}

		// For session token auth, organization is required
		// Only require organization if we're using session token auth (not API key auth)
		if client.SessionToken() != "" && client.APIKey() == "" && organization == "" {
			fmt.Println("Error: organization must be provided either as an argument or via --org-id flag when using session token authentication")
			return
		}
		if organization != "" {
			fmt.Printf("Getting usage statistics for organization %s...\n", organization)
		} else {
			fmt.Println("Getting usage statistics...")
		}

		metrics, err := client.GetMetrics(organization)
		if err != nil {
			fmt.Printf("Error fetching metrics: %v\n", err)
			return
		}

		// Convert metrics to UsageMetrics and Quota types
		orgID := organization
		if orgID == "" {
			orgID = "unknown"
		}

		usageMetrics := metrics.ToUsageMetrics(orgID, "unknown")
		quota := metrics.ToQuota()

		// Display usage metrics
		fmt.Printf("Usage Metrics:\n")
		fmt.Printf("  Organization ID: %s\n", usageMetrics.OrganizationID)
		fmt.Printf("  Model Name: %s\n", usageMetrics.ModelName)
		fmt.Printf("  Tokens Used: %d\n", usageMetrics.TokensUsed)
		fmt.Printf("  Tokens Limit: %d\n", usageMetrics.TokensLimit)
		fmt.Printf("  Requests Used: %d\n", usageMetrics.RequestsUsed)
		fmt.Printf("  Requests Limit: %d\n", usageMetrics.RequestsLimit)

		// Display quota information
		fmt.Printf("\nQuota Information:\n")
		fmt.Printf("  Limit: %d\n", quota.Limit)
		fmt.Printf("  Remaining: %d\n", quota.Remaining)
		fmt.Printf("  Reset Time: %s\n", quota.ResetTime)
	},
}
var monitorUsageCmd = &cobra.Command{
	Use:   "monitor",
	Short: "Start real-time monitoring of usage",
	Run: func(cmd *cobra.Command, args []string) {
		// Get organization ID from configuration/viper
		organization := viper.GetString("org-id")

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

		// Debug output
		// fmt.Printf("Debug: API Key: '%s', Session Token: '%s'\n", client.APIKey(), client.SessionToken())

		// For session token auth, organization is required
		// Only require organization if we're using session token auth (not API key auth)
		if client.SessionToken() != "" && client.APIKey() == "" && organization == "" {
			fmt.Println("Error: organization ID must be set via --org-id flag or configuration when using session token authentication")
			return
		}

		if organization != "" {
			fmt.Printf("Starting real-time monitoring for organization %s (model: %s, refresh: %ds)...\n", organization, model, refreshRate)
		} else {
			fmt.Printf("Starting real-time monitoring (model: %s, refresh: %ds)...\n", model, refreshRate)
		}

		// TODO: Implement real-time monitoring logic
		// This would continuously make requests to the Cerebras GraphQL endpoint
		// and display usage information with color-coded progress bars and tables
	},
}

func init() {
	UsageCmd.AddCommand(getUsageCmd)
	UsageCmd.AddCommand(monitorUsageCmd)
}
