package cmd

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/nathabonfim59/cerebras-code-monitor/internal/cerebras"
	"github.com/nathabonfim59/cerebras-code-monitor/internal/cerebras/graphql"
	"github.com/nathabonfim59/cerebras-code-monitor/internal/config"
	"github.com/nathabonfim59/cerebras-code-monitor/internal/tui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var OrganizationsCmd = &cobra.Command{
	Use:   "organizations",
	Short: "Manage organizations",
	Long:  "Commands to list and select organizations for monitoring",
	Run: func(cmd *cobra.Command, args []string) {
		// Check if --id flag is provided
		orgID, _ := cmd.Flags().GetString("id")
		if orgID != "" {
			// Save the organization ID to configuration without TUI
			viper.Set("org-id", orgID)
			if err := viper.WriteConfig(); err != nil {
				// If config file doesn't exist, create it
				if _, ok := err.(viper.ConfigFileNotFoundError); ok {
					configDir := config.GetConfigDir()
					configPath := filepath.Join(configDir, "settings.yaml")
					if err := viper.WriteConfigAs(configPath); err != nil {
						fmt.Printf("Error saving configuration: %v\n", err)
						return
					}
				} else {
					fmt.Printf("Error saving configuration: %v\n", err)
					return
				}
			}

			fmt.Printf("Organization ID %s saved to configuration.\n", orgID)
			return
		}

		// If no --id flag, proceed with interactive selection
		fmt.Println("Fetching organizations...")

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

		// Create GraphQL client
		graphqlClient := graphql.NewClient(client.SessionToken())
		responseBody, err := graphqlClient.MakeRequestWithOperationName("ListMyOrganizations", query, variables)
		if err != nil {
			fmt.Printf("Error fetching organizations: %v\n", err)
			return
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

		// Check if we have any organizations
		if len(response.Data.ListMyOrganizations) == 0 {
			fmt.Println("No organizations found.")
			return
		}

		// Use bubbletea to create an interactive selection interface
		model := tui.NewOrganizationListModel(response.Data.ListMyOrganizations)
		p := tea.NewProgram(model)
		if _, err := p.Run(); err != nil {
			fmt.Printf("Error running selection interface: %v\n", err)
			return
		}
	},
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
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// If organization ID is provided as argument, use it directly
		if len(args) > 0 {
			organizationID := args[0]
			fmt.Printf("Selecting organization %s for monitoring...\n", organizationID)

			// Save the organization ID to configuration
			viper.Set("org-id", organizationID)
			if err := viper.WriteConfig(); err != nil {
				// If config file doesn't exist, create it
				if _, ok := err.(viper.ConfigFileNotFoundError); ok {
					configDir := config.GetConfigDir()
					configPath := filepath.Join(configDir, "settings.yaml")
					if err := viper.WriteConfigAs(configPath); err != nil {
						fmt.Printf("Error saving configuration: %v\n", err)
						return
					}
				} else {
					fmt.Printf("Error saving configuration: %v\n", err)
					return
				}
			}

			fmt.Printf("Organization %s selected and saved to configuration.\n", organizationID)
			return
		}

		// If no organization ID provided, fetch and display organizations for selection
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

		// Create GraphQL client
		graphqlClient := graphql.NewClient(client.SessionToken())
		responseBody, err := graphqlClient.MakeRequestWithOperationName("ListMyOrganizations", query, variables)
		if err != nil {
			fmt.Printf("Error fetching organizations: %v\n", err)
			return
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

		// Check if we have any organizations
		if len(response.Data.ListMyOrganizations) == 0 {
			fmt.Println("No organizations found.")
			return
		}

		// Use bubbletea to create an interactive selection interface
		p := tea.NewProgram(tui.NewOrganizationListModel(response.Data.ListMyOrganizations))
		if _, err := p.Run(); err != nil {
			fmt.Printf("Error running selection interface: %v\n", err)
			return
		}
	},
}

func init() {
	listOrganizationsCmd.Flags().Bool("debug", false, "Enable debug output showing request/response details")
	OrganizationsCmd.Flags().String("id", "", "Organization ID to set for monitoring without TUI")
	OrganizationsCmd.AddCommand(listOrganizationsCmd)
	OrganizationsCmd.AddCommand(selectOrganizationCmd)
}
