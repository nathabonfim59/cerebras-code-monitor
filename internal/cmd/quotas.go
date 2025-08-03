package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var QuotasCmd = &cobra.Command{
	Use:   "quotas",
	Short: "Manage quotas",
	Long:  "Commands to retrieve and display quotas for organizations",
}

var getQuotasCmd = &cobra.Command{
	Use:   "get [organizationID]",
	Short: "Get quotas for an organization",
	Args:  cobra.MaximumNArgs(1), // Make organizationID optional
	Run: func(cmd *cobra.Command, args []string) {
		// If organizationID is not provided, use the one from configuration
		organizationID := ""
		if len(args) > 0 {
			organizationID = args[0]
		} else {
			// TODO: Get organization ID from configuration/viper
			organizationID = viper.GetString("org-id")
		}

		if organizationID == "" {
			fmt.Println("Error: organization ID must be provided either as an argument or via --org-id flag")
			return
		}

		fmt.Printf("Getting quotas for organization %s...\n", organizationID)
		// TODO: Implement quota retrieval logic
		// This would make a request to the Cerebras GraphQL endpoint
		// and extract rate limit information from response headers
	},
}

func init() {
	QuotasCmd.AddCommand(getQuotasCmd)
}
