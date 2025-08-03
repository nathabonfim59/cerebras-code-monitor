package cmd

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/nathabonfim59/cerebras-code-monitor/internal/cerebras"
	"github.com/spf13/viper"
)

func TestUsageCommand(t *testing.T) {
	// Save original viper state
	origOrgID := viper.GetString("org-id")
	origModel := viper.GetString("model")

	// Reset after test
	defer func() {
		viper.Set("org-id", origOrgID)
		viper.Set("model", origModel)
	}()

	tests := []struct {
		name         string
		orgID        string
		model        string
		args         []string
		apiKey       string
		sessionToken string
		expectError  bool
		errorMsg     string
	}{
		{
			name:         "with API key auth and default config",
			orgID:        "",
			model:        "qwen-3-coder-480b",
			args:         []string{},
			apiKey:       "test-api-key",
			sessionToken: "",
			expectError:  false,
		},
		{
			name:         "with API key auth and custom model",
			orgID:        "",
			model:        "custom-model",
			args:         []string{},
			apiKey:       "test-api-key",
			sessionToken: "",
			expectError:  false,
		},
		{
			name:         "with session token auth and org ID in config",
			orgID:        "test-org-123",
			model:        "qwen-3-coder-480b",
			args:         []string{},
			apiKey:       "",
			sessionToken: "test-session-token",
			expectError:  false,
		},
		{
			name:         "with session token auth and org ID as argument",
			orgID:        "",
			model:        "qwen-3-coder-480b",
			args:         []string{"arg-org-456"},
			apiKey:       "",
			sessionToken: "test-session-token",
			expectError:  false,
		},
		{
			name:         "with session token auth but no org ID",
			orgID:        "",
			model:        "qwen-3-coder-480b",
			args:         []string{},
			apiKey:       "",
			sessionToken: "test-session-token",
			expectError:  true,
			errorMsg:     "organization must be provided",
		},
		{
			name:         "with no authentication",
			orgID:        "",
			model:        "qwen-3-coder-480b",
			args:         []string{},
			apiKey:       "",
			sessionToken: "",
			expectError:  true,
			errorMsg:     "No authentication method configured",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup environment
			if tt.apiKey != "" {
				err := os.Setenv("CEREBRAS_API_KEY", tt.apiKey)
				if err != nil {
					t.Fatalf("Failed to set CEREBRAS_API_KEY: %v", err)
				}
			} else {
				err := os.Unsetenv("CEREBRAS_API_KEY")
				if err != nil {
					t.Fatalf("Failed to unset CEREBRAS_API_KEY: %v", err)
				}
			}

			if tt.sessionToken != "" {
				err := os.Setenv("CEREBRAS_SESSION_TOKEN", tt.sessionToken)
				if err != nil {
					t.Fatalf("Failed to set CEREBRAS_SESSION_TOKEN: %v", err)
				}
			} else {
				err := os.Unsetenv("CEREBRAS_SESSION_TOKEN")
				if err != nil {
					t.Fatalf("Failed to unset CEREBRAS_SESSION_TOKEN: %v", err)
				}
			}

			// Setup viper config
			viper.Set("org-id", tt.orgID)
			viper.Set("model", tt.model)

			// Capture output
			var buf bytes.Buffer

			// Execute the command logic (simulating the Run function)
			organization := ""
			if len(tt.args) > 0 {
				organization = tt.args[0]
			} else {
				organization = viper.GetString("org-id")
			}

			client := cerebras.NewClient()

			// Test authentication check
			if !client.HasAuth() {
				if !tt.expectError {
					t.Errorf("Expected no error, but got authentication error")
				} else if !strings.Contains("No authentication method configured", tt.errorMsg) {
					t.Errorf("Expected error message to contain '%s'", tt.errorMsg)
				}
				return
			}

			// Test organization requirement for session token auth
			if client.SessionToken() != "" && client.APIKey() == "" && organization == "" {
				if !tt.expectError {
					t.Errorf("Expected no error, but got organization requirement error")
				} else if !strings.Contains("organization must be provided", tt.errorMsg) {
					t.Errorf("Expected error message to contain '%s'", tt.errorMsg)
				}
				return
			}

			// If we get here and expect error, the test should fail
			if tt.expectError {
				t.Errorf("Expected error but got none")
			}

			// Test that client is properly configured
			if tt.apiKey != "" && client.APIKey() != tt.apiKey {
				t.Errorf("Expected API key '%s', got '%s'", tt.apiKey, client.APIKey())
			}

			if tt.sessionToken != "" && client.SessionToken() != tt.sessionToken {
				t.Errorf("Expected session token '%s', got '%s'", tt.sessionToken, client.SessionToken())
			}

			// Cleanup
			_ = buf.String() // Use the buffer to avoid unused variable
		})
	}
}

func TestUsageCommandArguments(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		configOrgID   string
		expectedOrgID string
	}{
		{
			name:          "no args, use config org ID",
			args:          []string{},
			configOrgID:   "config-org-123",
			expectedOrgID: "config-org-123",
		},
		{
			name:          "arg provided, use arg org ID",
			args:          []string{"arg-org-456"},
			configOrgID:   "config-org-123",
			expectedOrgID: "arg-org-456",
		},
		{
			name:          "no args, no config, empty org ID",
			args:          []string{},
			configOrgID:   "",
			expectedOrgID: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viper.Set("org-id", tt.configOrgID)

			// Simulate the argument processing logic from getUsageCmd
			organization := ""
			if len(tt.args) > 0 {
				organization = tt.args[0]
			} else {
				organization = viper.GetString("org-id")
			}

			if organization != tt.expectedOrgID {
				t.Errorf("Expected organization ID '%s', got '%s'", tt.expectedOrgID, organization)
			}
		})
	}
}

func TestUsageCommandModelSelection(t *testing.T) {
	tests := []struct {
		name          string
		configModel   string
		expectedModel string
	}{
		{
			name:          "custom model from config",
			configModel:   "custom-model-name",
			expectedModel: "custom-model-name",
		},
		{
			name:          "default model when config empty",
			configModel:   "",
			expectedModel: "qwen-3-coder-480b",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viper.Set("model", tt.configModel)

			// Simulate the model selection logic from getUsageCmd
			model := viper.GetString("model")
			if model == "" {
				model = "qwen-3-coder-480b"
			}

			if model != tt.expectedModel {
				t.Errorf("Expected model '%s', got '%s'", tt.expectedModel, model)
			}
		})
	}
}
