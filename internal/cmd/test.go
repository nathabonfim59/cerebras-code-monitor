package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/nathabonfim59/cerebras-code-monitor/internal/cerebras"
	"github.com/nathabonfim59/cerebras-code-monitor/internal/cerebras/graphql"
	"github.com/spf13/cobra"
)

var TestCmd = &cobra.Command{
	Use:   "test",
	Short: "Test command for development",
	Long:  "Scaffolding command for developing additional Cerebras API requests",
}

var testExampleCmd = &cobra.Command{
	Use:   "example",
	Short: "Example test subcommand",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("This is a test command scaffolding")
		fmt.Println("Add your test implementations here")
	},
}

var testListOrganizationUsageQuotasCmd = &cobra.Command{
	Use:   "quotas",
	Short: "Test ListOrganizationUsageQuotas GraphQL query",
	Run: func(cmd *cobra.Command, args []string) {
		// Create Cerebras client
		client := cerebras.NewClient()
		if !client.HasAuth() {
			fmt.Println("Error: No authentication method configured. Please login first.")
			return
		}

		// Create GraphQL client and validate authentication
		graphqlClient := graphql.NewClient(client.SessionToken())

		// Check if organization ID is provided
		if len(args) < 1 {
			fmt.Println("Error: Organization ID is required as an argument")
			return
		}
		orgID := args[0]

		// Make GraphQL request to list organization usage quotas
		query := graphql.ListOrganizationUsageQuotasQuery
		variables := map[string]interface{}{
			"organizationId": orgID,
		}
		responseBody, err := graphqlClient.MakeRequestWithOperationName("ListOrganizationUsageQuotas", query, variables)
		if err != nil {
			fmt.Printf("Error fetching organization usage quotas: %v\n", err)
			return
		}

		// Parse the GraphQL response
		var response struct {
			Data struct {
				ListOrganizationUsageQuotas []cerebras.UsageQuota `json:"ListOrganizationUsageQuotas"`
			} `json:"data"`
		}

		if err := json.Unmarshal(responseBody, &response); err != nil {
			fmt.Printf("Error parsing response: %v\n", err)
			return
		}

		// Display usage quotas in a formatted way
		fmt.Printf("Usage Quotas for Organization %s:\n", orgID)
		for _, quota := range response.Data.ListOrganizationUsageQuotas {
			fmt.Printf("  Model ID: %s\n", quota.ModelId)
			fmt.Printf("  Region ID: %s\n", quota.RegionId)
			fmt.Printf("  Requests Per Minute: %s\n", quota.RequestsPerMinute)
			fmt.Printf("  Tokens Per Minute: %s\n", quota.TokensPerMinute)
			fmt.Printf("  Requests Per Hour: %s\n", quota.RequestsPerHour)
			fmt.Printf("  Tokens Per Hour: %s\n", quota.TokensPerHour)
			fmt.Printf("  Requests Per Day: %s\n", quota.RequestsPerDay)
			fmt.Printf("  Tokens Per Day: %s\n", quota.TokensPerDay)
			fmt.Printf("  Max Sequence Length: %s\n", quota.MaxSequenceLength)
			fmt.Printf("  Max Completion Tokens: %s\n", quota.MaxCompletionTokens)
			fmt.Printf("  ---\n")
		}
	},
}

var testListOrganizationUsageCmd = &cobra.Command{
	Use:   "usage",
	Short: "Test ListOrganizationUsage GraphQL query",
	Run: func(cmd *cobra.Command, args []string) {
		// Create Cerebras client
		client := cerebras.NewClient()
		if !client.HasAuth() {
			fmt.Println("Error: No authentication method configured. Please login first.")
			return
		}

		// Create GraphQL client and validate authentication
		graphqlClient := graphql.NewClient(client.SessionToken())

		// Check if organization ID is provided
		if len(args) < 1 {
			fmt.Println("Error: Organization ID is required as an argument")
			return
		}
		orgID := args[0]

		// Make GraphQL request to list organization usage
		query := graphql.ListOrganizationUsageQuery
		variables := map[string]interface{}{
			"organizationId": orgID,
		}
		responseBody, err := graphqlClient.MakeRequestWithOperationName("ListOrganizationUsage", query, variables)
		if err != nil {
			fmt.Printf("Error fetching organization usage: %v\n", err)
			return
		}

		// Parse the GraphQL response
		var response struct {
			Data struct {
				ListOrganizationUsage []cerebras.OrganizationUsage `json:"ListOrganizationUsage"`
			} `json:"data"`
		}

		if err := json.Unmarshal(responseBody, &response); err != nil {
			fmt.Printf("Error parsing response: %v\n", err)
			return
		}

		// Display usage information in a formatted way
		fmt.Printf("Usage Information for Organization %s:\n", orgID)
		for _, usage := range response.Data.ListOrganizationUsage {
			fmt.Printf("  Model ID: %s\n", usage.ModelId)
			fmt.Printf("  Region ID: %s\n", usage.RegionId)
			fmt.Printf("  RPM: %s\n", usage.RPM)
			fmt.Printf("  TPM: %s\n", usage.TPM)
			fmt.Printf("  RPH: %s\n", usage.RPH)
			fmt.Printf("  TPH: %s\n", usage.TPH)
			fmt.Printf("  RPD: %s\n", usage.RPD)
			fmt.Printf("  TPD: %s\n", usage.TPD)
			fmt.Printf("  ---\n")
		}
	},
}

func init() {
	TestCmd.AddCommand(testExampleCmd)
	TestCmd.AddCommand(testListOrganizationUsageQuotasCmd)
	TestCmd.AddCommand(testListOrganizationUsageCmd)
}
