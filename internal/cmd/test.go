package cmd

import (
	"fmt"

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

func init() {
	TestCmd.AddCommand(testExampleCmd)
}
