package cerebras

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/spf13/viper"
)

// Client represents a Cerebras API client
type Client struct {
	httpClient   *http.Client
	apiKey       string
	sessionToken string
	baseURL      string
	graphqlURL   string
}

// NewClient creates a new Cerebras API client
func NewClient() *Client {
	client := &Client{
		httpClient: &http.Client{},
		baseURL:    "https://api.cerebras.ai",
		graphqlURL: "https://cloud.cerebras.ai/api/graphql",
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

// SessionToken returns the session token
func (c *Client) SessionToken() string {
	return c.sessionToken
}

// APIKey returns the API key
func (c *Client) APIKey() string {
	return c.apiKey
}

// getAuthHeaders returns the appropriate headers for authentication
func (c *Client) getAuthHeaders() map[string]string {
	headers := make(map[string]string)

	// Prioritize API key over session token for REST API requests
	if c.apiKey != "" {
		headers["Authorization"] = fmt.Sprintf("Bearer %s", c.apiKey)
	} else if c.sessionToken != "" {
		headers["Cookie"] = fmt.Sprintf("authjs.session-token=%s", c.sessionToken)
	}

	return headers
}

// MakeGraphQLRequest makes a GraphQL request to the Cerebras API
func (c *Client) MakeGraphQLRequest(query string, variables map[string]interface{}) ([]byte, error) {
	return c.MakeGraphQLRequestWithDebug(query, variables, false)
}

// MakeGraphQLRequestWithDebug makes a GraphQL request to the Cerebras API with optional debug output
func (c *Client) MakeGraphQLRequestWithDebug(query string, variables map[string]interface{}, debug bool) ([]byte, error) {
	return c.MakeGraphQLRequestWithOperationName("", query, variables, debug)
}

// MakeGraphQLRequestWithOperationName makes a GraphQL request with an operation name
func (c *Client) MakeGraphQLRequestWithOperationName(operationName, query string, variables map[string]interface{}, debug bool) ([]byte, error) {
	if c.sessionToken == "" {
		return nil, fmt.Errorf("GraphQL requests require session token authentication")
	}

	url := c.graphqlURL

	// Create request body
	requestBody := map[string]interface{}{
		"query":     query,
		"variables": variables,
	}

	if operationName != "" {
		requestBody["operationName"] = operationName
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	if debug {
		fmt.Printf("Debug: Request URL: %s\n", url)
		fmt.Printf("Debug: Request Body: %s\n", string(jsonBody))
	}

	req, err := http.NewRequest("POST", url, strings.NewReader(string(jsonBody)))
	if err != nil {
		return nil, err
	}

	// Add authentication headers
	headers := c.getAuthHeaders()
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// Add content type header
	req.Header.Set("Content-Type", "application/json")

	// Add additional headers to match the curl request
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:141.0) Gecko/20100101 Firefox/141.0")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Origin", "https://cloud.cerebras.ai")
	req.Header.Set("Referer", "https://cloud.cerebras.ai/platform")
	req.Header.Set("Sec-GPC", "1")
	req.Header.Set("Alt-Used", "cloud.cerebras.ai")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Header.Set("Priority", "u=4")
	req.Header.Set("TE", "trailers")

	if debug {
		fmt.Printf("Debug: Request Headers:\n")
		for name, values := range req.Header {
			for _, value := range values {
				// Don't print sensitive cookie values in full
				if name == "Cookie" {
					fmt.Printf("  %s: [REDACTED]\n", name)
				} else {
					fmt.Printf("  %s: %s\n", name, value)
				}
			}
		}
	}

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if debug {
		fmt.Printf("Debug: Response Status: %s\n", resp.Status)
		fmt.Printf("Debug: Response Headers:\n")
		for name, values := range resp.Header {
			for _, value := range values {
				fmt.Printf("  %s: %s\n", name, value)
			}
		}
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if debug {
		fmt.Printf("Debug: Response Body: %s\n", string(body))
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GraphQL request failed with status code: %d, body: %s", resp.StatusCode, string(body))
	}

	return body, nil
}
