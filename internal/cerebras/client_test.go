package cerebras

import (
	"os"
	"testing"

	"github.com/spf13/viper"
)

func TestNewClient(t *testing.T) {
	// Save original environment
	origAPIKey := os.Getenv("CEREBRAS_API_KEY")
	origSessionToken := os.Getenv("CEREBRAS_SESSION_TOKEN")
	origViperAPIKey := viper.GetString("api-key")
	origViperSessionToken := viper.GetString("session-token")

	// Clean up after test
	defer func() {
		if origAPIKey != "" {
			_ = os.Setenv("CEREBRAS_API_KEY", origAPIKey)
		} else {
			_ = os.Unsetenv("CEREBRAS_API_KEY")
		}
		if origSessionToken != "" {
			_ = os.Setenv("CEREBRAS_SESSION_TOKEN", origSessionToken)
		} else {
			_ = os.Unsetenv("CEREBRAS_SESSION_TOKEN")
		}
		viper.Set("api-key", origViperAPIKey)
		viper.Set("session-token", origViperSessionToken)
	}()

	tests := []struct {
		name                 string
		envAPIKey            string
		envSessionToken      string
		viperAPIKey          string
		viperSessionToken    string
		expectedAPIKey       string
		expectedSessionToken string
		expectedHasAuth      bool
	}{
		{
			name:                 "API key from environment",
			envAPIKey:            "env-api-key",
			envSessionToken:      "",
			viperAPIKey:          "",
			viperSessionToken:    "",
			expectedAPIKey:       "env-api-key",
			expectedSessionToken: "",
			expectedHasAuth:      true,
		},
		{
			name:                 "session token from environment",
			envAPIKey:            "",
			envSessionToken:      "env-session-token",
			viperAPIKey:          "",
			viperSessionToken:    "",
			expectedAPIKey:       "",
			expectedSessionToken: "env-session-token",
			expectedHasAuth:      true,
		},
		{
			name:                 "API key from viper config",
			envAPIKey:            "",
			envSessionToken:      "",
			viperAPIKey:          "viper-api-key",
			viperSessionToken:    "",
			expectedAPIKey:       "viper-api-key",
			expectedSessionToken: "",
			expectedHasAuth:      true,
		},
		{
			name:                 "session token from viper config",
			envAPIKey:            "",
			envSessionToken:      "",
			viperAPIKey:          "",
			viperSessionToken:    "viper-session-token",
			expectedAPIKey:       "",
			expectedSessionToken: "viper-session-token",
			expectedHasAuth:      true,
		},
		{
			name:                 "environment takes precedence over viper",
			envAPIKey:            "env-api-key",
			envSessionToken:      "env-session-token",
			viperAPIKey:          "viper-api-key",
			viperSessionToken:    "viper-session-token",
			expectedAPIKey:       "env-api-key",
			expectedSessionToken: "env-session-token",
			expectedHasAuth:      true,
		},
		{
			name:                 "no authentication configured",
			envAPIKey:            "",
			envSessionToken:      "",
			viperAPIKey:          "",
			viperSessionToken:    "",
			expectedAPIKey:       "",
			expectedSessionToken: "",
			expectedHasAuth:      false,
		},
		{
			name:                 "both API key and session token available",
			envAPIKey:            "test-api-key",
			envSessionToken:      "test-session-token",
			viperAPIKey:          "",
			viperSessionToken:    "",
			expectedAPIKey:       "test-api-key",
			expectedSessionToken: "test-session-token",
			expectedHasAuth:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup environment variables
			if tt.envAPIKey != "" {
				err := os.Setenv("CEREBRAS_API_KEY", tt.envAPIKey)
				if err != nil {
					t.Fatalf("Failed to set CEREBRAS_API_KEY: %v", err)
				}
			} else {
				err := os.Unsetenv("CEREBRAS_API_KEY")
				if err != nil {
					t.Fatalf("Failed to unset CEREBRAS_API_KEY: %v", err)
				}
			}

			if tt.envSessionToken != "" {
				err := os.Setenv("CEREBRAS_SESSION_TOKEN", tt.envSessionToken)
				if err != nil {
					t.Fatalf("Failed to set CEREBRAS_SESSION_TOKEN: %v", err)
				}
			} else {
				err := os.Unsetenv("CEREBRAS_SESSION_TOKEN")
				if err != nil {
					t.Fatalf("Failed to unset CEREBRAS_SESSION_TOKEN: %v", err)
				}
			}

			// Setup viper configuration
			viper.Set("api-key", tt.viperAPIKey)
			viper.Set("session-token", tt.viperSessionToken)

			// Create client
			client := NewClient()

			// Verify API key
			if client.APIKey() != tt.expectedAPIKey {
				t.Errorf("Expected API key '%s', got '%s'", tt.expectedAPIKey, client.APIKey())
			}

			// Verify session token
			if client.SessionToken() != tt.expectedSessionToken {
				t.Errorf("Expected session token '%s', got '%s'", tt.expectedSessionToken, client.SessionToken())
			}

			// Verify HasAuth
			if client.HasAuth() != tt.expectedHasAuth {
				t.Errorf("Expected HasAuth() %v, got %v", tt.expectedHasAuth, client.HasAuth())
			}

			// Verify base URL is set correctly
			if client.baseURL != "https://api.cerebras.ai" {
				t.Errorf("Expected base URL 'https://api.cerebras.ai', got '%s'", client.baseURL)
			}

			// Verify HTTP client is initialized
			if client.httpClient == nil {
				t.Error("Expected HTTP client to be initialized")
			}
		})
	}
}

func TestClientHasAuth(t *testing.T) {
	tests := []struct {
		name         string
		apiKey       string
		sessionToken string
		expected     bool
	}{
		{
			name:         "has API key",
			apiKey:       "test-api-key",
			sessionToken: "",
			expected:     true,
		},
		{
			name:         "has session token",
			apiKey:       "",
			sessionToken: "test-session-token",
			expected:     true,
		},
		{
			name:         "has both",
			apiKey:       "test-api-key",
			sessionToken: "test-session-token",
			expected:     true,
		},
		{
			name:         "has neither",
			apiKey:       "",
			sessionToken: "",
			expected:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &Client{
				apiKey:       tt.apiKey,
				sessionToken: tt.sessionToken,
			}

			result := client.HasAuth()
			if result != tt.expected {
				t.Errorf("Expected HasAuth() %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestClientGetAuthHeaders(t *testing.T) {
	tests := []struct {
		name               string
		apiKey             string
		sessionToken       string
		expectedAuthHeader string
		expectedCookie     string
	}{
		{
			name:               "API key authentication",
			apiKey:             "test-api-key",
			sessionToken:       "",
			expectedAuthHeader: "Bearer test-api-key",
			expectedCookie:     "",
		},
		{
			name:               "session token authentication",
			apiKey:             "",
			sessionToken:       "test-session-token",
			expectedAuthHeader: "",
			expectedCookie:     "authjs.session-token=test-session-token",
		},
		{
			name:               "API key takes priority",
			apiKey:             "test-api-key",
			sessionToken:       "test-session-token",
			expectedAuthHeader: "Bearer test-api-key",
			expectedCookie:     "",
		},
		{
			name:               "no authentication",
			apiKey:             "",
			sessionToken:       "",
			expectedAuthHeader: "",
			expectedCookie:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &Client{
				apiKey:       tt.apiKey,
				sessionToken: tt.sessionToken,
			}

			headers := client.getAuthHeaders()

			// Check Authorization header
			if authHeader, exists := headers["Authorization"]; exists {
				if authHeader != tt.expectedAuthHeader {
					t.Errorf("Expected Authorization header '%s', got '%s'", tt.expectedAuthHeader, authHeader)
				}
			} else if tt.expectedAuthHeader != "" {
				t.Errorf("Expected Authorization header '%s', but header was not set", tt.expectedAuthHeader)
			}

			// Check Cookie header
			if cookie, exists := headers["Cookie"]; exists {
				if cookie != tt.expectedCookie {
					t.Errorf("Expected Cookie header '%s', got '%s'", tt.expectedCookie, cookie)
				}
			} else if tt.expectedCookie != "" {
				t.Errorf("Expected Cookie header '%s', but header was not set", tt.expectedCookie)
			}

			// Verify that we don't have unexpected headers
			expectedHeaderCount := 0
			if tt.expectedAuthHeader != "" {
				expectedHeaderCount++
			}
			if tt.expectedCookie != "" {
				expectedHeaderCount++
			}

			if len(headers) != expectedHeaderCount {
				t.Errorf("Expected %d headers, got %d", expectedHeaderCount, len(headers))
			}
		})
	}
}

func TestClientAccessors(t *testing.T) {
	apiKey := "test-api-key"
	sessionToken := "test-session-token"

	client := &Client{
		apiKey:       apiKey,
		sessionToken: sessionToken,
	}

	if client.APIKey() != apiKey {
		t.Errorf("Expected APIKey() to return '%s', got '%s'", apiKey, client.APIKey())
	}

	if client.SessionToken() != sessionToken {
		t.Errorf("Expected SessionToken() to return '%s', got '%s'", sessionToken, client.SessionToken())
	}
}

func TestClientInitialization(t *testing.T) {
	client := NewClient()

	// Verify client is properly initialized
	if client == nil {
		t.Fatal("Expected client to be non-nil")
	}

	if client.httpClient == nil {
		t.Error("Expected HTTP client to be initialized")
	}

	if client.baseURL != "https://api.cerebras.ai" {
		t.Errorf("Expected base URL 'https://api.cerebras.ai', got '%s'", client.baseURL)
	}
}
