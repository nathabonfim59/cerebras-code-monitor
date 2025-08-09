package main

import (
	"fmt"
	"os"

	"github.com/nathabonfim59/cerebras-code-monitor/internal/cmd"
	"github.com/nathabonfim59/cerebras-code-monitor/internal/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

var rootCmd = &cobra.Command{
	Use:   "cerebras-monitor",
	Short: "A tool to monitor Cerebras AI usage",
	Long:  "Real-time monitoring tool for Cerebras AI usage with rate limit tracking. Track your token consumption and request limits with predictions and warnings.",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Migration check removed to prevent output
		// Users should manually run migrations if needed
	},
	Run: func(cmd *cobra.Command, args []string) {
		// Check if version flag is set
		if showVersion, _ := cmd.Flags().GetBool("version"); showVersion {
			fmt.Printf("cerebras-code-monitor %s (commit: %s, built: %s)\n", version, commit, date)
			return
		}

		// Default behavior - start monitoring
		fmt.Println("Starting Cerebras Code - Usage Monitor...")
		// TODO: Implement monitoring logic
	},
}

func init() {
	// Configuration flags
	rootCmd.PersistentFlags().String("session-token", "", "Cerebras session token (can be set via environment variable)")
	rootCmd.PersistentFlags().String("org-id", "", "Organization ID to monitor")
	rootCmd.PersistentFlags().String("model", "qwen-3-coder-480b", "Model to monitor")
	rootCmd.PersistentFlags().Int("refresh-rate", 10, "Data refresh rate in seconds (1-60)")
	rootCmd.PersistentFlags().Float64("refresh-per-second", 0.75, "Display refresh rate in Hz (0.1-20.0)")
	rootCmd.PersistentFlags().String("timezone", config.GetUserTimezone(), "Timezone (auto-detected)")
	rootCmd.PersistentFlags().String("time-format", "auto", "Time format: 12h, 24h, or auto")
	rootCmd.PersistentFlags().String("theme", "auto", "Display theme: light, dark, or auto")
	rootCmd.PersistentFlags().String("log-level", "INFO", "Logging level: DEBUG, INFO, WARNING, ERROR, CRITICAL")
	rootCmd.PersistentFlags().String("log-file", "", "Log file path")
	rootCmd.PersistentFlags().Bool("debug", false, "Enable debug logging")
	rootCmd.PersistentFlags().Bool("clear", false, "Clear saved configuration")
	rootCmd.PersistentFlags().String("icons", "emoji", "Icon set to use: emoji or nerdfont")
	rootCmd.PersistentFlags().BoolP("version", "v", false, "Show version information")

	// Bind flags to viper
	err := viper.BindPFlag("session-token", rootCmd.PersistentFlags().Lookup("session-token"))
	if err != nil {
		fmt.Printf("Error binding session-token flag: %v\n", err)
	}
	err = viper.BindPFlag("org-id", rootCmd.PersistentFlags().Lookup("org-id"))
	if err != nil {
		fmt.Printf("Error binding org-id flag: %v\n", err)
	}
	err = viper.BindPFlag("model", rootCmd.PersistentFlags().Lookup("model"))
	if err != nil {
		fmt.Printf("Error binding model flag: %v\n", err)
	}
	err = viper.BindPFlag("refresh-rate", rootCmd.PersistentFlags().Lookup("refresh-rate"))
	if err != nil {
		fmt.Printf("Error binding refresh-rate flag: %v\n", err)
	}
	err = viper.BindPFlag("refresh-per-second", rootCmd.PersistentFlags().Lookup("refresh-per-second"))
	if err != nil {
		fmt.Printf("Error binding refresh-per-second flag: %v\n", err)
	}
	err = viper.BindPFlag("timezone", rootCmd.PersistentFlags().Lookup("timezone"))
	if err != nil {
		fmt.Printf("Error binding timezone flag: %v\n", err)
	}
	err = viper.BindPFlag("time-format", rootCmd.PersistentFlags().Lookup("time-format"))
	if err != nil {
		fmt.Printf("Error binding time-format flag: %v\n", err)
	}
	err = viper.BindPFlag("theme", rootCmd.PersistentFlags().Lookup("theme"))
	if err != nil {
		fmt.Printf("Error binding theme flag: %v\n", err)
	}
	err = viper.BindPFlag("log-level", rootCmd.PersistentFlags().Lookup("log-level"))
	if err != nil {
		fmt.Printf("Error binding log-level flag: %v\n", err)
	}
	err = viper.BindPFlag("log-file", rootCmd.PersistentFlags().Lookup("log-file"))
	if err != nil {
		fmt.Printf("Error binding log-file flag: %v\n", err)
	}
	err = viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug"))
	if err != nil {
		fmt.Printf("Error binding debug flag: %v\n", err)
	}
	err = viper.BindPFlag("clear", rootCmd.PersistentFlags().Lookup("clear"))
	if err != nil {
		fmt.Printf("Error binding clear flag: %v\n", err)
	}
	err = viper.BindPFlag("icons", rootCmd.PersistentFlags().Lookup("icons"))
	if err != nil {
		fmt.Printf("Error binding icons flag: %v\n", err)
	}
	err = viper.BindPFlag("version", rootCmd.PersistentFlags().Lookup("version"))
	if err != nil {
		fmt.Printf("Error binding version flag: %v\n", err)
	}

	// Set environment variable support
	viper.SetEnvPrefix("CEREBRAS")
	viper.AutomaticEnv()

	// Set config file using XDG conventions
	config.SetupViper()

	// Try to read config file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			fmt.Printf("Error reading config file: %v\n", err)
		}
	}

	// Add subcommands
	rootCmd.AddCommand(cmd.LoginCmd)
	rootCmd.AddCommand(cmd.OrganizationsCmd)
	rootCmd.AddCommand(cmd.QuotasCmd)
	rootCmd.AddCommand(cmd.UsageCmd)
	rootCmd.AddCommand(cmd.MigrationsCmd)
	rootCmd.AddCommand(cmd.TestCmd)
	rootCmd.AddCommand(cmd.DashboardCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
