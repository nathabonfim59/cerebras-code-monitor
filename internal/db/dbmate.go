package db

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"

	"github.com/amacneil/dbmate/v2/pkg/dbmate"
	_ "github.com/amacneil/dbmate/v2/pkg/driver/sqlite"
	dbfiles "github.com/nathabonfim59/cerebras-code-monitor/db"
)

// GetDBMate creates and configures a dbmate instance
func GetDBMate() (*dbmate.DB, error) {
	// Use XDG data directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory: %w", err)
	}

	dbPath := filepath.Join(homeDir, ".local", "share", "cerebras-code", "database.db")

	// Create directory if it doesn't exist
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	// Create database URL
	dbURL := &url.URL{
		Scheme: "sqlite",
		Path:   dbPath,
	}

	// Create dbmate instance
	db := dbmate.New(dbURL)
	db.FS = dbfiles.MigrationFiles
	db.MigrationsDir = []string{"migrations"}

	// Set schema file path
	schemaDir := filepath.Dir(dbPath)
	db.SchemaFile = filepath.Join(schemaDir, "schema.sql")

	return db, nil
}

// MigrateDatabase applies pending migrations
func MigrateDatabase() error {
	db, err := GetDBMate()
	if err != nil {
		return err
	}

	// Apply migrations
	if err := db.CreateAndMigrate(); err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	return nil
}

// MigrationStatus returns the number of pending migrations
func MigrationStatus() (int, error) {
	db, err := GetDBMate()
	if err != nil {
		return 0, err
	}

	// Get status
	status, err := db.Status(false)
	if err != nil {
		return 0, fmt.Errorf("failed to get migration status: %w", err)
	}

	return status, nil
}
