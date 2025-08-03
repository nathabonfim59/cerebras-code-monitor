package cerebras

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/spf13/viper"
)

// GetMetrics fetches usage metrics from Cerebras servers
func (c *Client) GetMetrics(organization string) (*RateLimitInfo, error) {
	if !c.HasAuth() {
		return nil, fmt.Errorf("no authentication method configured")
	}

	// Determine which authentication method to use
	// API key auth only works with REST API requests
	if c.apiKey != "" && c.sessionToken == "" {
		// API key doesn't need organization parameter
		return c.getMetricsWithAPIKey()
	}

	// Session token auth only works with GraphQL requests
	if c.sessionToken != "" && c.apiKey == "" {
		return c.getMetricsWithSessionToken(organization)
	}

	// When both auth methods are available, prioritize based on intended usage
	// For now, we'll prioritize session token for GraphQL usage
	if c.sessionToken != "" {
		return c.getMetricsWithSessionToken(organization)
	} else if c.apiKey != "" {
		return c.getMetricsWithAPIKey()
	}

	return nil, fmt.Errorf("no valid authentication method found")
}

// getMetricsWithSessionToken fetches metrics using GraphQL with session token auth
func (c *Client) getMetricsWithSessionToken(organization string) (*RateLimitInfo, error) {
	// Make GraphQL request to get organization usage
	query := `query GetOrganizationUsage($organizationId: String!) {
		GetOrganizationUsage(organizationId: $organizationId) {
			rateLimit {
				limitRequestsDay
				limitTokensMinute
				remainingRequestsDay
				remainingTokensMinute
				resetRequestsDay
				resetTokensMinute
			}
		}
	}`

	variables := map[string]interface{}{
		"organizationId": organization,
	}

	responseBody, err := c.MakeGraphQLRequest(query, variables)
	if err != nil {
		return nil, err
	}

	// Parse the GraphQL response
	var response struct {
		Data struct {
			GetOrganizationUsage struct {
				RateLimit *RateLimitInfo `json:"rateLimit"`
			} `json:"GetOrganizationUsage"`
		} `json:"data"`
	}

	if err := json.Unmarshal(responseBody, &response); err != nil {
		return nil, fmt.Errorf("failed to parse GraphQL response: %w", err)
	}

	if response.Data.GetOrganizationUsage.RateLimit == nil {
		return &RateLimitInfo{}, nil
	}

	return response.Data.GetOrganizationUsage.RateLimit, nil
}

// getMetricsWithAPIKey fetches metrics using REST API with API key auth
func (c *Client) getMetricsWithAPIKey() (*RateLimitInfo, error) {
	// Make a chat completion request to get rate limit headers
	url := fmt.Sprintf("%s/v1/chat/completions", c.baseURL)

	// Get model from viper config or use default
	model := viper.GetString("model")
	if model == "" {
		model = "qwen-3-coder-480b"
	}

	// Create a minimal request body that should work
	body := fmt.Sprintf(`{
		"model": "%s",
		"messages": [{"role": "user", "content": "hello"}],
		"max_completion_tokens": 1
	}`, model)

	req, err := http.NewRequest("POST", url, strings.NewReader(body))
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

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			// Log the error or handle it appropriately
			// For now, we'll just print it to stderr
			fmt.Fprintf(viper.Get("stderr").(*strings.Builder), "Error closing response body: %v\n", closeErr)
		}
	}()

	// Debug: print all headers if debug flag is set
	if viper.GetBool("debug") {
		fmt.Printf("Response Headers:\n")
		for key, values := range resp.Header {
			for _, value := range values {
				fmt.Printf("  %s: %s\n", key, value)
			}
		}
	}

	// Parse rate limit headers regardless of status code
	rateLimitInfo := &RateLimitInfo{}
	if limit := resp.Header.Get("X-Ratelimit-Limit-Requests-Day"); limit != "" {
		if _, err := fmt.Sscanf(limit, "%d", &rateLimitInfo.LimitRequestsDay); err != nil {
			rateLimitInfo.LimitRequestsDay = 0
		}
	}

	if limit := resp.Header.Get("X-Ratelimit-Limit-Tokens-Minute"); limit != "" {
		if _, err := fmt.Sscanf(limit, "%d", &rateLimitInfo.LimitTokensMinute); err != nil {
			rateLimitInfo.LimitTokensMinute = 0
		}
	}

	if remaining := resp.Header.Get("X-Ratelimit-Remaining-Requests-Day"); remaining != "" {
		if _, err := fmt.Sscanf(remaining, "%d", &rateLimitInfo.RemainingRequestsDay); err != nil {
			rateLimitInfo.RemainingRequestsDay = 0
		}
	}

	if remaining := resp.Header.Get("X-Ratelimit-Remaining-Tokens-Minute"); remaining != "" {
		if _, err := fmt.Sscanf(remaining, "%d", &rateLimitInfo.RemainingTokensMinute); err != nil {
			rateLimitInfo.RemainingTokensMinute = 0
		}
	}

	if reset := resp.Header.Get("X-Ratelimit-Reset-Requests-Day"); reset != "" {
		var val float64
		if _, err := fmt.Sscanf(reset, "%f", &val); err == nil {
			rateLimitInfo.ResetRequestsDay = int64(val)
		}
	}

	if reset := resp.Header.Get("X-Ratelimit-Reset-Tokens-Minute"); reset != "" {
		var val float64
		if _, err := fmt.Sscanf(reset, "%f", &val); err == nil {
			rateLimitInfo.ResetTokensMinute = int64(val)
		}
	}

	// If we got rate limit headers, return the rateLimitInfo even if the request failed
	if rateLimitInfo.LimitRequestsDay > 0 || rateLimitInfo.LimitTokensMinute > 0 || rateLimitInfo.RemainingRequestsDay > 0 || rateLimitInfo.RemainingTokensMinute > 0 {
		return rateLimitInfo, nil
	}

	// Otherwise, return an error
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status code: %d", resp.StatusCode)
	}

	return rateLimitInfo, nil
}
