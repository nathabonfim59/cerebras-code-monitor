package cerebras

import (
	"fmt"
	"net/http"
)

// Metrics represents usage metrics from Cerebras
type Metrics struct {
	// TODO: Define the structure based on Cerebras API response
}

// GetMetrics fetches usage metrics from Cerebras servers
func (c *Client) GetMetrics(organization string) (*Metrics, error) {
	if !c.HasAuth() {
		return nil, fmt.Errorf("no authentication method configured")
	}

	// Determine which authentication method to use
	if c.sessionToken != "" {
		return c.getMetricsWithSessionToken(organization)
	} else if c.apiKey != "" {
		// API key doesn't need organization parameter
		return c.getMetricsWithAPIKey()
	}

	return nil, fmt.Errorf("no valid authentication method found")
}

// getMetricsWithSessionToken fetches metrics using GraphQL with session token auth
func (c *Client) getMetricsWithSessionToken(organization string) (*Metrics, error) {
	// TODO: Implement GraphQL request to fetch metrics
	// Use c.sessionToken for authentication

	url := fmt.Sprintf("%s/graphql", c.baseURL)
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return nil, err
	}

	// Add authentication headers
	headers := c.getAuthHeaders()
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// TODO: Execute request and parse response
	return nil, fmt.Errorf("not implemented")
}

// getMetricsWithAPIKey fetches metrics using REST API with API key auth
func (c *Client) getMetricsWithAPIKey() (*Metrics, error) {
	// TODO: Implement REST API request to fetch metrics
	// Use c.apiKey for authentication

	url := fmt.Sprintf("%s/v1/metrics", c.baseURL)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// Add authentication headers
	headers := c.getAuthHeaders()
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// TODO: Execute request and parse response
	return nil, fmt.Errorf("not implemented")
}
