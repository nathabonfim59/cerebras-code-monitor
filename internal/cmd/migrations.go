package cmd

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"

	dbfiles "github.com/nathabonfim59/cerebras-code-monitor/db"
	"github.com/spf13/cobra"

	"github.com/amacneil/dbmate/v2/pkg/dbmate"
	_ "github.com/amacneil/dbmate/v2/pkg/driver/sqlite"
)

var force bool

var MigrationsCmd = &cobra.Command{
	Use:   "migrations",
	Short: "Manage database migrations",
	Long: `Manage database migrations for usage statistics tracking.
	
Available commands:
  migrate     Execute pending migrations 
  status      Show current migration status
`,
	Run: func(cmd *cobra.Command, args []string) {
		// If no subcommand is called, show help
		cmd.Help()
	},
}

var dbm *dbmate.DB

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show migration status",
	Long:  "Show pending database migrations that need to be applied",
	Run: func(cmd *cobra.Command, args []string) {
		initDBMate()

		availableMigrations, err := dbm.FindMigrations()
		if err != nil {
			fmt.Printf("Error listing available migrations: %v\n", err)
			os.Exit(1)
		}

		if len(availableMigrations) == 0 {
			fmt.Println("Migration Status:\n\n! No migrations found")
			return
		}

		// Get migrations status
		pending, err := dbm.Status(false)
		if err != nil {
			fmt.Printf("Error getting migrations status: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("📊 Migration Status: \n \\--> ")

		if pending == 0 {
			fmt.Println("✅ Database is up to date")
		} else {
			fmt.Printf("⚠️ %d pending migrations\n", pending)
			fmt.Println("\n📋 Available migrations:")
			for _, m := range availableMigrations {
				fmt.Printf("  📎 %s\n", m.Version)
			}
		}
	},
}

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Run database migrations",
	Long:  "Apply all pending database migrations to update the schema",
	Run: func(cmd *cobra.Command, args []string) {
		initDBMate()

		// If --force is enabled, try to drop the database first
		if force {
			if err := dbm.Drop(); err != nil {
				fmt.Printf("Error dropping database: %v\n", err)
				os.Exit(1)
			}
			if err := dbm.Create(); err != nil {
				fmt.Printf("Error recreating database: %v\n", err)
				os.Exit(1)
			}
		}

		if err := dbm.CreateAndMigrate(); err != nil {
			fmt.Printf("Error executing migrations: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Migrations executed successfully!")
	},
}

func initDBMate() {
	// Get database path - use XDG directory by default
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("Error getting user home directory: %v\n", err)
		os.Exit(1)
	}

	dbPath := filepath.Join(homeDir, ".local", "share", "cerebras-code", "database.db")

	// Create the database URL
	dbURL := &url.URL{
		Scheme: "sqlite",
		Path:   dbPath,
	}

	// Create dbmate instance
	dbm = dbmate.New(dbURL)
	dbm.FS = dbfiles.MigrationFiles
	dbm.MigrationsDir = []string{"migrations"}

	// Set schema file path
	schemaDir := filepath.Dir(dbPath)
	dbm.SchemaFile = filepath.Join(schemaDir, "schema.sql")
}
func init() {
	MigrationsCmd.AddCommand(statusCmd)
	MigrationsCmd.AddCommand(migrateCmd)

	migrateCmd.Flags().BoolVarP(&force, "force", "f", false, "DANGER: Drop and recreate the database before migrating")
}
