package cerebras

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

func TestGetMetrics(t *testing.T) {
	tests := []struct {
		name         string
		apiKey       string
		sessionToken string
		organization string
		expectError  bool
		errorMsg     string
	}{
		{
			name:         "with API key authentication",
			apiKey:       "test-api-key",
			sessionToken: "",
			organization: "",
			expectError:  false,
		},
		{
			name:         "with session token authentication",
			apiKey:       "",
			sessionToken: "test-session-token",
			organization: "test-org-123",
			expectError:  false,
		},
		{
			name:         "API key takes priority over session token",
			apiKey:       "test-api-key",
			sessionToken: "test-session-token",
			organization: "",
			expectError:  false,
		},
		{
			name:         "no authentication configured",
			apiKey:       "",
			sessionToken: "",
			organization: "",
			expectError:  true,
			errorMsg:     "no authentication method configured",
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

			client := NewClient()

			// This is a basic test - we can't actually make HTTP requests
			// in unit tests, but we can test the authentication logic
			if !client.HasAuth() && !tt.expectError {
				t.Errorf("Expected client to have auth, but HasAuth() returned false")
			}

			if client.HasAuth() && tt.expectError && tt.errorMsg == "no authentication method configured" {
				t.Errorf("Expected no auth, but client has auth configured")
			}

			// Test auth method selection logic
			if tt.apiKey != "" && client.APIKey() != tt.apiKey {
				t.Errorf("Expected API key '%s', got '%s'", tt.apiKey, client.APIKey())
			}

			if tt.sessionToken != "" && client.SessionToken() != tt.sessionToken {
				t.Errorf("Expected session token '%s', got '%s'", tt.sessionToken, client.SessionToken())
			}

			// Test that API key takes priority
			if tt.apiKey != "" && tt.sessionToken != "" {
				if client.APIKey() == "" {
					t.Error("Expected API key to be set when both API key and session token are provided")
				}
			}
		})
	}
}

func TestGetMetricsWithAPIKey(t *testing.T) {
	// Create a test server that returns rate limit headers
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check authentication header
		auth := r.Header.Get("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Check content type
		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "Invalid content type", http.StatusBadRequest)
			return
		}

		// Set rate limit headers
		w.Header().Set("X-Ratelimit-Limit-Requests-Day", "28800")
		w.Header().Set("X-Ratelimit-Limit-Tokens-Minute", "275000")
		w.Header().Set("X-Ratelimit-Remaining-Requests-Day", "27593")
		w.Header().Set("X-Ratelimit-Remaining-Tokens-Minute", "275000")
		w.Header().Set("X-Ratelimit-Reset-Requests-Day", "62341")
		w.Header().Set("X-Ratelimit-Reset-Tokens-Minute", "30")

		// Return success
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(`{"choices": [{"message": {"content": "test"}}]}`))
		if err != nil {
			t.Fatalf("Failed to write response: %v", err)
		}
	}))
	defer server.Close()

	// Create client with test server
	client := &Client{
		httpClient: &http.Client{},
		apiKey:     "test-api-key",
		baseURL:    server.URL,
	}

	// Set up viper config
	viper.Set("model", "qwen-3-coder-480b")

	rateLimitInfo, err := client.getMetricsWithAPIKey()
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify rate limit info was parsed correctly
	if rateLimitInfo.LimitRequestsDay != 28800 {
		t.Errorf("Expected LimitRequestsDay 28800, got %d", rateLimitInfo.LimitRequestsDay)
	}

	if rateLimitInfo.LimitTokensMinute != 275000 {
		t.Errorf("Expected LimitTokensMinute 275000, got %d", rateLimitInfo.LimitTokensMinute)
	}

	if rateLimitInfo.RemainingRequestsDay != 27593 {
		t.Errorf("Expected RemainingRequestsDay 27593, got %d", rateLimitInfo.RemainingRequestsDay)
	}

	if rateLimitInfo.RemainingTokensMinute != 275000 {
		t.Errorf("Expected RemainingTokensMinute 275000, got %d", rateLimitInfo.RemainingTokensMinute)
	}

	if rateLimitInfo.ResetRequestsDay != 62341 {
		t.Errorf("Expected ResetRequestsDay 62341, got %d", rateLimitInfo.ResetRequestsDay)
	}

	if rateLimitInfo.ResetTokensMinute != 30 {
		t.Errorf("Expected ResetTokensMinute 30, got %d", rateLimitInfo.ResetTokensMinute)
	}
}

func TestGetMetricsWithAPIKeyNoHeaders(t *testing.T) {
	// Create a test server that returns no rate limit headers
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(`{"choices": [{"message": {"content": "test"}}]}`))
		if err != nil {
			t.Fatalf("Failed to write response: %v", err)
		}
	}))
	defer server.Close()

	client := &Client{
		httpClient: &http.Client{},
		apiKey:     "test-api-key",
		baseURL:    server.URL,
	}

	viper.Set("model", "qwen-3-coder-480b")

	rateLimitInfo, err := client.getMetricsWithAPIKey()
	if err != nil {
		t.Errorf("Expected no error when response is OK but no headers, got: %v", err)
	}

	// All values should be zero since no headers were provided
	if rateLimitInfo.LimitRequestsDay != 0 || rateLimitInfo.LimitTokensMinute != 0 {
		t.Error("Expected all rate limit values to be zero when no headers present")
	}
}

func TestGetMetricsWithAPIKeyError(t *testing.T) {
	// Create a test server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "API Error", http.StatusInternalServerError)
	}))
	defer server.Close()

	client := &Client{
		httpClient: &http.Client{},
		apiKey:     "test-api-key",
		baseURL:    server.URL,
	}

	viper.Set("model", "qwen-3-coder-480b")

	_, err := client.getMetricsWithAPIKey()
	if err == nil {
		t.Error("Expected error when API returns error, got nil")
	}
}

func TestRateLimitHeaderParsing(t *testing.T) {
	tests := []struct {
		name        string
		headers     map[string]string
		expected    *RateLimitInfo
		expectError bool
	}{
		{
			name: "all headers present and valid",
			headers: map[string]string{
				"X-Ratelimit-Limit-Requests-Day":      "28800",
				"X-Ratelimit-Limit-Tokens-Minute":     "275000",
				"X-Ratelimit-Remaining-Requests-Day":  "27593",
				"X-Ratelimit-Remaining-Tokens-Minute": "275000",
				"X-Ratelimit-Reset-Requests-Day":      "62341.5",
				"X-Ratelimit-Reset-Tokens-Minute":     "30.2",
			},
			expected: &RateLimitInfo{
				LimitRequestsDay:      28800,
				LimitTokensMinute:     275000,
				RemainingRequestsDay:  27593,
				RemainingTokensMinute: 275000,
				ResetRequestsDay:      62341,
				ResetTokensMinute:     30,
			},
			expectError: false,
		},
		{
			name: "partial headers",
			headers: map[string]string{
				"X-Ratelimit-Limit-Requests-Day":     "28800",
				"X-Ratelimit-Remaining-Requests-Day": "27593",
			},
			expected: &RateLimitInfo{
				LimitRequestsDay:      28800,
				LimitTokensMinute:     0,
				RemainingRequestsDay:  27593,
				RemainingTokensMinute: 0,
				ResetRequestsDay:      0,
				ResetTokensMinute:     0,
			},
			expectError: false,
		},
		{
			name:        "no headers",
			headers:     map[string]string{},
			expected:    &RateLimitInfo{},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test server with the specified headers
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				for key, value := range tt.headers {
					w.Header().Set(key, value)
				}
				w.WriteHeader(http.StatusOK)
				_, err := w.Write([]byte(`{"choices": [{"message": {"content": "test"}}]}`))
				if err != nil {
					t.Fatalf("Failed to write response: %v", err)
				}
			}))
			defer server.Close()

			client := &Client{
				httpClient: &http.Client{},
				apiKey:     "test-api-key",
				baseURL:    server.URL,
			}

			viper.Set("model", "qwen-3-coder-480b")

			rateLimitInfo, err := client.getMetricsWithAPIKey()

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			} else if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}

			if !tt.expectError {
				if rateLimitInfo.LimitRequestsDay != tt.expected.LimitRequestsDay {
					t.Errorf("Expected LimitRequestsDay %d, got %d", tt.expected.LimitRequestsDay, rateLimitInfo.LimitRequestsDay)
				}
				if rateLimitInfo.LimitTokensMinute != tt.expected.LimitTokensMinute {
					t.Errorf("Expected LimitTokensMinute %d, got %d", tt.expected.LimitTokensMinute, rateLimitInfo.LimitTokensMinute)
				}
				if rateLimitInfo.RemainingRequestsDay != tt.expected.RemainingRequestsDay {
					t.Errorf("Expected RemainingRequestsDay %d, got %d", tt.expected.RemainingRequestsDay, rateLimitInfo.RemainingRequestsDay)
				}
				if rateLimitInfo.RemainingTokensMinute != tt.expected.RemainingTokensMinute {
					t.Errorf("Expected RemainingTokensMinute %d, got %d", tt.expected.RemainingTokensMinute, rateLimitInfo.RemainingTokensMinute)
				}
				if rateLimitInfo.ResetRequestsDay != tt.expected.ResetRequestsDay {
					t.Errorf("Expected ResetRequestsDay %d, got %d", tt.expected.ResetRequestsDay, rateLimitInfo.ResetRequestsDay)
				}
				if rateLimitInfo.ResetTokensMinute != tt.expected.ResetTokensMinute {
					t.Errorf("Expected ResetTokensMinute %d, got %d", tt.expected.ResetTokensMinute, rateLimitInfo.ResetTokensMinute)
				}
			}
		})
	}
}

func TestModelConfigurationInMetrics(t *testing.T) {
	tests := []struct {
		name          string
		configModel   string
		expectedModel string
	}{
		{
			name:          "custom model from config",
			configModel:   "custom-model",
			expectedModel: "custom-model",
		},
		{
			name:          "default model when empty",
			configModel:   "",
			expectedModel: "qwen-3-coder-480b",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify the request body contains the expected model
				buf, err := io.ReadAll(r.Body)
				if err != nil {
					t.Fatalf("Failed to read request body: %v", err)
				}
				body := string(buf)

				expectedModelInBody := fmt.Sprintf(`"model": "%s"`, tt.expectedModel)
				if !strings.Contains(body, expectedModelInBody) {
					t.Errorf("Expected request body to contain '%s', got: %s", expectedModelInBody, body)
				}

				w.Header().Set("X-Ratelimit-Limit-Requests-Day", "28800")
				w.WriteHeader(http.StatusOK)
				_, err = w.Write([]byte(`{"choices": [{"message": {"content": "test"}}]}`))
				if err != nil {
					t.Fatalf("Failed to write response: %v", err)
				}
			}))
			defer server.Close()

			client := &Client{
				httpClient: &http.Client{},
				apiKey:     "test-api-key",
				baseURL:    server.URL,
			}

			viper.Set("model", tt.configModel)

			_, err := client.getMetricsWithAPIKey()
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}
