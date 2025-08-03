package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/nathabonfim59/cerebras-code-monitor/internal/cerebras"
	"github.com/nathabonfim59/cerebras-code-monitor/internal/cerebras/graphql"
	"github.com/spf13/cobra"
)

var OrganizationsCmd = &cobra.Command{
	Use:   "organizations",
	Short: "Manage organizations",
	Long:  "Commands to list and select organizations for monitoring",
}

var listOrganizationsCmd = &cobra.Command{
	Use:   "list",
	Short: "List available organizations",
	Run: func(cmd *cobra.Command, args []string) {
		debug, _ := cmd.Flags().GetBool("debug")
		fmt.Println("Listing organizations...")

		// Create Cerebras client
		client := cerebras.NewClient()
		if !client.HasAuth() {
			fmt.Println("Error: No authentication method configured. Please login first.")
			return
		}

		// Only proceed if we have a session token (GraphQL only works with session token)
		if client.SessionToken() == "" {
			fmt.Println("Error: Organization listing requires session token authentication.")
			return
		}

		// Make GraphQL request to list organizations
		query := graphql.ListOrganizationsQuery

		variables := map[string]interface{}{}

		if debug {
			fmt.Printf("Debug: Making GraphQL request to: %s\n", "https://cloud.cerebras.ai/api/graphql")
			fmt.Printf("Debug: Query: %s\n", query)
			fmt.Printf("Debug: Variables: %+v\n", variables)
			fmt.Printf("Debug: Session Token: %s\n", client.SessionToken())
		}

		// Create GraphQL client
		graphqlClient := graphql.NewClient(client.SessionToken())

		responseBody, err := graphqlClient.MakeRequestWithOperationName("ListMyOrganizations", query, variables)
		if err != nil {
			fmt.Printf("Error fetching organizations: %v\n", err)
			return
		}

		if debug {
			fmt.Printf("Debug: Response Body: %s\n", string(responseBody))
		}

		// Parse the GraphQL response
		var response struct {
			Data struct {
				ListMyOrganizations []cerebras.Organization `json:"ListMyOrganizations"`
			} `json:"data"`
		}

		if err := json.Unmarshal(responseBody, &response); err != nil {
			fmt.Printf("Error parsing response: %v\n", err)
			return
		}

		// Display organizations in a formatted way
		fmt.Printf("Organizations:\n")
		for _, org := range response.Data.ListMyOrganizations {
			fmt.Printf("  ID: %s\n", org.ID)
			fmt.Printf("  Name: %s\n", org.Name)
			fmt.Printf("  Type: %s\n", org.OrganizationType)
			fmt.Printf("  State: %s\n", org.State)
			fmt.Printf("  ---\n")
		}
	},
}

var selectOrganizationCmd = &cobra.Command{
	Use:   "select [organizationID]",
	Short: "Select an organization for monitoring",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		organizationID := args[0]
		fmt.Printf("Selecting organization %s for monitoring...\n", organizationID)
		// TODO: Implement organization selection logic
		// This would save the organization ID to configuration
	},
}

func init() {
	listOrganizationsCmd.Flags().Bool("debug", false, "Enable debug output showing request/response details")
	OrganizationsCmd.AddCommand(listOrganizationsCmd)
	OrganizationsCmd.AddCommand(selectOrganizationCmd)
}
