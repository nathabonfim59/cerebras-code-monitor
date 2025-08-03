package graphql

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Client represents a GraphQL client for Cerebras API
type Client struct {
	httpClient   *http.Client
	sessionToken string
	url          string
}

// NewClient creates a new GraphQL client
func NewClient(sessionToken string) *Client {
	return &Client{
		httpClient:   &http.Client{},
		sessionToken: sessionToken,
		url:          "https://cloud.cerebras.ai/api/graphql",
	}
}

// HasAuth checks if the client has session token authentication configured
func (c *Client) HasAuth() bool {
	return c.sessionToken != ""
}

// SessionToken returns the session token
func (c *Client) SessionToken() string {
	return c.sessionToken
}

// getAuthHeaders returns the appropriate headers for GraphQL authentication
func (c *Client) getAuthHeaders() map[string]string {
	headers := make(map[string]string)

	// For GraphQL requests, always use session token via cookies if available
	if c.sessionToken != "" {
		headers["Cookie"] = fmt.Sprintf("authjs.session-token=%s", c.sessionToken)
	}

	return headers
}

// MakeRequest makes a GraphQL request to the Cerebras API
func (c *Client) MakeRequest(query string, variables map[string]interface{}) ([]byte, error) {
	return c.MakeRequestWithOperationName("", query, variables)
}

// MakeRequestWithOperationName makes a GraphQL request with an operation name
func (c *Client) MakeRequestWithOperationName(operationName, query string, variables map[string]interface{}) ([]byte, error) {
	if c.sessionToken == "" {
		return nil, fmt.Errorf("GraphQL requests require session token authentication")
	}

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

	req, err := http.NewRequest("POST", c.url, strings.NewReader(string(jsonBody)))
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

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GraphQL request failed with status code: %d, body: %s", resp.StatusCode, string(body))
	}

	return body, nil
}
