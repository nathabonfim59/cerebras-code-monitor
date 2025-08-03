package cerebras

import (
	"fmt"
	"net/http"
	"os"

	"github.com/spf13/viper"
)

// Client represents a Cerebras API client
type Client struct {
	httpClient   *http.Client
	apiKey       string
	sessionToken string
	baseURL      string
}

// NewClient creates a new Cerebras API client
func NewClient() *Client {
	client := &Client{
		httpClient: &http.Client{},
		baseURL:    "https://api.cerebras.ai",
	}

	// Check for API key in environment variable first
	if apiKey := os.Getenv("CEREBRAS_API_KEY"); apiKey != "" {
		client.apiKey = apiKey
	} else {
		// Fall back to config file
		client.apiKey = viper.GetString("api-key")
	}

	// Check for session token in environment variable first
	if sessionToken := os.Getenv("CEREBRAS_SESSION_TOKEN"); sessionToken != "" {
		client.sessionToken = sessionToken
	} else {
		// Fall back to config file
		client.sessionToken = viper.GetString("session-token")
	}

	return client
}

// HasAuth checks if the client has any authentication method configured
func (c *Client) HasAuth() bool {
	return c.apiKey != "" || c.sessionToken != ""
}

// getAuthHeaders returns the appropriate headers for authentication
func (c *Client) getAuthHeaders() map[string]string {
	headers := make(map[string]string)

	if c.apiKey != "" {
		headers["Authorization"] = fmt.Sprintf("Bearer %s", c.apiKey)
	} else if c.sessionToken != "" {
		headers["Cookie"] = fmt.Sprintf("authjs.session-token=%s", c.sessionToken)
	}

	return headers
}
