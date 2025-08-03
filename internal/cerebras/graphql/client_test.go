package graphql

import (
	"testing"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name                 string
		sessionToken         string
		expectedSessionToken string
		expectedHasAuth      bool
	}{
		{
			name:                 "session token provided",
			sessionToken:         "test-session-token",
			expectedSessionToken: "test-session-token",
			expectedHasAuth:      true,
		},
		{
			name:                 "no session token provided",
			sessionToken:         "",
			expectedSessionToken: "",
			expectedHasAuth:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create client
			client := NewClient(tt.sessionToken)

			// Verify session token
			if client.SessionToken() != tt.expectedSessionToken {
				t.Errorf("Expected session token '%s', got '%s'", tt.expectedSessionToken, client.SessionToken())
			}

			// Verify HasAuth
			if client.HasAuth() != tt.expectedHasAuth {
				t.Errorf("Expected HasAuth() %v, got %v", tt.expectedHasAuth, client.HasAuth())
			}

			// Verify base URL is set correctly
			if client.url != "https://cloud.cerebras.ai/api/graphql" {
				t.Errorf("Expected base URL 'https://cloud.cerebras.ai/api/graphql', got '%s'", client.url)
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
		sessionToken string
		expected     bool
	}{
		{
			name:         "has session token",
			sessionToken: "test-session-token",
			expected:     true,
		},
		{
			name:         "has no session token",
			sessionToken: "",
			expected:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient(tt.sessionToken)

			result := client.HasAuth()
			if result != tt.expected {
				t.Errorf("Expected HasAuth() %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestClientGetAuthHeaders(t *testing.T) {
	tests := []struct {
		name           string
		sessionToken   string
		expectedCookie string
	}{
		{
			name:           "session token authentication",
			sessionToken:   "test-session-token",
			expectedCookie: "authjs.session-token=test-session-token",
		},
		{
			name:           "no authentication",
			sessionToken:   "",
			expectedCookie: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient(tt.sessionToken)

			headers := client.getAuthHeaders()

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
	sessionToken := "test-session-token"

	client := NewClient(sessionToken)

	if client.SessionToken() != sessionToken {
		t.Errorf("Expected SessionToken() to return '%s', got '%s'", sessionToken, client.SessionToken())
	}
}

func TestClientInitialization(t *testing.T) {
	client := NewClient("test-session-token")

	// Verify client is properly initialized
	if client == nil {
		t.Fatal("Expected client to be non-nil")
	}

	if client.httpClient == nil {
		t.Error("Expected HTTP client to be initialized")
	}

	if client.url != "https://cloud.cerebras.ai/api/graphql" {
		t.Errorf("Expected base URL 'https://cloud.cerebras.ai/api/graphql', got '%s'", client.url)
	}
}
