package graphql

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// MockHTTPClient is a mock implementation of http.Client for testing
type MockHTTPClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	if m.DoFunc != nil {
		return m.DoFunc(req)
	}
	return nil, fmt.Errorf("DoFunc not implemented")
}

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

func TestMakeRequestWithOperationName(t *testing.T) {
	// Test data
	testQuery := `query TestQuery { test }`
	testVariables := map[string]interface{}{"key": "value"}
	testOperationName := "TestQuery"
	testResponseBody := []byte(`{"data": {"test": "result"}}`)

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		// Verify content type
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}

		// Verify authentication cookie
		expectedCookie := "authjs.session-token=test-session-token"
		if r.Header.Get("Cookie") != expectedCookie {
			t.Errorf("Expected Cookie header '%s', got '%s'", expectedCookie, r.Header.Get("Cookie"))
		}

		// Verify request body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("Failed to read request body: %v", err)
		}

		var requestBody map[string]interface{}
		if err := json.Unmarshal(body, &requestBody); err != nil {
			t.Fatalf("Failed to unmarshal request body: %v", err)
		}

		if requestBody["query"] != testQuery {
			t.Errorf("Expected query '%s', got '%s'", testQuery, requestBody["query"])
		}

		if requestBody["operationName"] != testOperationName {
			t.Errorf("Expected operationName '%s', got '%s'", testOperationName, requestBody["operationName"])
		}

		variables, ok := requestBody["variables"].(map[string]interface{})
		if !ok {
			t.Fatalf("Variables is not a map")
		}

		if variables["key"] != "value" {
			t.Errorf("Expected variables[key] to be 'value', got '%s'", variables["key"])
		}

		// Send response
		w.WriteHeader(http.StatusOK)
		_, err = w.Write(testResponseBody)
		if err != nil {
			t.Fatalf("Failed to write response: %v", err)
		}
	}))
	defer server.Close()

	// Create client with mock HTTP client
	client := &Client{
		httpClient:   &http.Client{},
		sessionToken: "test-session-token",
		url:          server.URL,
	}

	// Execute request
	response, err := client.MakeRequestWithOperationName(testOperationName, testQuery, testVariables)
	if err != nil {
		t.Fatalf("MakeRequestWithOperationName failed: %v", err)
	}

	// Verify response
	if !bytes.Equal(response, testResponseBody) {
		t.Errorf("Expected response body '%s', got '%s'", string(testResponseBody), string(response))
	}
}

func TestMakeRequestWithOperationNameUnauthorized(t *testing.T) {
	// Create client without session token
	client := NewClient("")

	// Execute request
	_, err := client.MakeRequestWithOperationName("TestQuery", `query TestQuery { test }`, map[string]interface{}{})
	if err == nil {
		t.Fatal("Expected error for unauthorized request, got nil")
	}

	expectedError := "GraphQL requests require session token authentication"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestMakeRequestWithOperationNameHTTPError(t *testing.T) {
	// Create test server that returns error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}))
	defer server.Close()

	// Create client with mock HTTP client
	client := &Client{
		httpClient:   &http.Client{},
		sessionToken: "test-session-token",
		url:          server.URL,
	}

	// Execute request
	_, err := client.MakeRequestWithOperationName("TestQuery", `query TestQuery { test }`, map[string]interface{}{})
	if err == nil {
		t.Fatal("Expected error for HTTP error, got nil")
	}

	if err.Error() != "GraphQL request failed with status code: 500, body: Internal Server Error\n" {
		t.Errorf("Expected specific HTTP error message, got '%s'", err.Error())
	}
}
