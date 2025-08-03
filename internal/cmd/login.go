package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var LoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login to Cerebras platform",
	Long:  "Commands to authenticate with the Cerebras platform using either session cookie or API key",
}

var loginWithCookieCmd = &cobra.Command{
	Use:   "cookie [cookie-value]",
	Short: "Login using session cookie",
	Long: `Login using your Cerebras session cookie. Due to HTTP-only cookie restrictions, 
you'll need to manually copy the 'authjs.session-token' cookie value from your browser's 
Developer Tools > Application > Cookies. This cookie is only used to fetch your usage data 
from the Cerebras platform. The tool is open source and you can inspect the code yourself.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cookieValue := args[0]
		fmt.Printf("Logging in with cookie: %s\n", cookieValue)
		// TODO: Implement login with cookie logic
		// This would validate the cookie and save it to configuration
	},
}

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Logout from Cerebras platform",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Logging out...")
		// TODO: Implement logout logic
		// This would clear the saved cookie from configuration
	},
}

var loginWithApiKeyCmd = &cobra.Command{
	Use:   "apikey [api-key]",
	Short: "Login using API key",
	Long: `Login using your Cerebras API key. This is an alternative to session cookie authentication.
Note that API key authentication has limitations:
- Shows only data for that specific key
- Cannot switch organizations
- Less accurate for token prediction calculations
- Minute-level data is not available

You can either specify the API key as an environment variable CEREBRAS_API_KEY or use this command
to save it to your local database at XDG/.config/cerebras-monitor/settings.json`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		apiKeyValue := args[0]
		fmt.Printf("Logging in with API key: %s\n", apiKeyValue)
		// TODO: Implement login with API key logic
		// This would validate the API key and save it to configuration
	},
}

func init() {
	LoginCmd.AddCommand(loginWithCookieCmd)
	LoginCmd.AddCommand(loginWithApiKeyCmd)
	LoginCmd.AddCommand(logoutCmd)
}
