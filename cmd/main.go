package main

import (
	"fmt"
	"os"

	"github.com/nathabonfim59/cerebras-code-monitor/internal/cmd"
	"github.com/nathabonfim59/cerebras-code-monitor/internal/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "cerebras-monitor",
	Short: "A tool to monitor Cerebras AI usage",
	Long:  "Real-time monitoring tool for Cerebras AI usage with rate limit tracking. Track your token consumption and request limits with predictions and warnings.",
	Run: func(cmd *cobra.Command, args []string) {
		// Check if version flag is set
		if version, _ := cmd.Flags().GetBool("version"); version {
			fmt.Println("cerebras-code-monitor v0.1.0")
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
	rootCmd.PersistentFlags().String("model", "llama3.1-8b", "Model to monitor")
	rootCmd.PersistentFlags().Int("refresh-rate", 10, "Data refresh rate in seconds (1-60)")
	rootCmd.PersistentFlags().Float64("refresh-per-second", 0.75, "Display refresh rate in Hz (0.1-20.0)")
	rootCmd.PersistentFlags().String("timezone", "auto", "Timezone (auto-detected)")
	rootCmd.PersistentFlags().String("time-format", "auto", "Time format: 12h, 24h, or auto")
	rootCmd.PersistentFlags().String("theme", "auto", "Display theme: light, dark, or auto")
	rootCmd.PersistentFlags().String("log-level", "INFO", "Logging level: DEBUG, INFO, WARNING, ERROR, CRITICAL")
	rootCmd.PersistentFlags().String("log-file", "", "Log file path")
	rootCmd.PersistentFlags().Bool("debug", false, "Enable debug logging")
	rootCmd.PersistentFlags().Bool("clear", false, "Clear saved configuration")
	rootCmd.PersistentFlags().BoolP("version", "v", false, "Show version information")

	// Bind flags to viper
	viper.BindPFlag("session-token", rootCmd.PersistentFlags().Lookup("session-token"))
	viper.BindPFlag("org-id", rootCmd.PersistentFlags().Lookup("org-id"))
	viper.BindPFlag("model", rootCmd.PersistentFlags().Lookup("model"))
	viper.BindPFlag("refresh-rate", rootCmd.PersistentFlags().Lookup("refresh-rate"))
	viper.BindPFlag("refresh-per-second", rootCmd.PersistentFlags().Lookup("refresh-per-second"))
	viper.BindPFlag("timezone", rootCmd.PersistentFlags().Lookup("timezone"))
	viper.BindPFlag("time-format", rootCmd.PersistentFlags().Lookup("time-format"))
	viper.BindPFlag("theme", rootCmd.PersistentFlags().Lookup("theme"))
	viper.BindPFlag("log-level", rootCmd.PersistentFlags().Lookup("log-level"))
	viper.BindPFlag("log-file", rootCmd.PersistentFlags().Lookup("log-file"))
	viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug"))
	viper.BindPFlag("clear", rootCmd.PersistentFlags().Lookup("clear"))
	viper.BindPFlag("version", rootCmd.PersistentFlags().Lookup("version"))

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
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
