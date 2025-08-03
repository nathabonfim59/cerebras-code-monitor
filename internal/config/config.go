package config

import (
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
	"github.com/spf13/viper"
)

const (
	AppName    = "cerebras-monitor"
	ConfigFile = "settings.json"
)

// GetConfigPath returns the full path to the config file following XDG conventions
func GetConfigPath() string {
	return filepath.Join(xdg.ConfigHome, AppName, ConfigFile)
}

// GetConfigDir returns the directory where config files are stored following XDG conventions
func GetConfigDir() string {
	return filepath.Join(xdg.ConfigHome, AppName)
}

// SetupViper configures viper to use XDG config directory
func SetupViper() {
	// Ensure config directory exists
	configDir := GetConfigDir()
	if err := os.MkdirAll(configDir, 0755); err != nil {
		// If we can't create the config directory, viper will use current directory
		configDir = "."
	}

	viper.SetConfigName("settings")
	viper.SetConfigType("json")
	viper.AddConfigPath(configDir)

	// Also check current directory for config (for development)
	viper.AddConfigPath(".")
}
