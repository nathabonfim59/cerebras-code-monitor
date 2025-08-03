package config

import (
	"os"
	"path/filepath"
	"time"

	"github.com/adrg/xdg"
	"github.com/spf13/viper"
)

const (
	AppName    = "cerebras-monitor"
	ConfigFile = "settings.yaml"
)

// GetConfigPath returns the full path to the config file following XDG conventions
func GetConfigPath() string {
	return filepath.Join(xdg.ConfigHome, AppName, ConfigFile)
}

// GetConfigDir returns the directory where config files are stored following XDG conventions
func GetConfigDir() string {
	return filepath.Join(xdg.ConfigHome, AppName)
}

// GetUserTimezone returns the user's local timezone
func GetUserTimezone() string {
	zone, _ := time.Now().Zone()
	return zone
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
	viper.SetConfigType("yaml")
	viper.AddConfigPath(configDir)

	// Also check current directory for config (for development)
	viper.AddConfigPath(".")
}

// Icons represents the available icon sets
type Icons struct {
	Check        string
	Warning      string
	Error        string
	Info         string
	Refresh      string
	Dashboard    string
	Organization string
	Model        string
	Token        string
	Request      string
	Time         string
	Theme        string
	Settings     string
}

// GetIcons returns the appropriate icon set based on configuration
func GetIcons() Icons {
	iconPreference := viper.GetString("icons")
	if iconPreference == "nerdfont" {
		return Icons{
			Check:        NerdfontCheck,
			Warning:      NerdfontWarning,
			Error:        NerdfontError,
			Info:         NerdfontInfo,
			Refresh:      NerdfontRefresh,
			Dashboard:    NerdfontDashboard,
			Organization: NerdfontOrganization,
			Model:        NerdfontModel,
			Token:        NerdfontToken,
			Request:      NerdfontRequest,
			Time:         NerdfontTime,
			Theme:        NerdfontTheme,
			Settings:     NerdfontSettings,
		}
	}

	// Default to emoji icons
	return Icons{
		Check:        EmojiCheck,
		Warning:      EmojiWarning,
		Error:        EmojiError,
		Info:         EmojiInfo,
		Refresh:      EmojiRefresh,
		Dashboard:    EmojiDashboard,
		Organization: EmojiOrganization,
		Model:        EmojiModel,
		Token:        EmojiToken,
		Request:      EmojiRequest,
		Time:         EmojiTime,
		Theme:        EmojiTheme,
		Settings:     EmojiSettings,
	}
}
