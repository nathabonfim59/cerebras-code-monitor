package cerebras

import (
	"fmt"
	"net/http"
	"strconv"
)

// Metrics represents usage metrics from Cerebras
type Metrics struct {
	LimitRequestsDay      int64 `json:"limit_requests_day"`
	LimitTokensMinute     int64 `json:"limit_tokens_minute"`
	RemainingRequestsDay  int64 `json:"remaining_requests_day"`
	RemainingTokensMinute int64 `json:"remaining_tokens_minute"`
	ResetRequestsDay      int64 `json:"reset_requests_day"`
	ResetTokensMinute     int64 `json:"reset_tokens_minute"`
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
	// Make a simple API call to get the rate limit headers
	url := fmt.Sprintf("%s/v1/chat/completions", c.baseURL)
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return nil, err
	}

	// Add authentication headers
	headers := c.getAuthHeaders()
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Parse rate limit headers
	metrics := &Metrics{}

	if limit := resp.Header.Get("x-ratelimit-limit-requests-day"); limit != "" {
		if val, err := strconv.ParseInt(limit, 10, 64); err == nil {
			metrics.LimitRequestsDay = val
		}
	}

	if limit := resp.Header.Get("x-ratelimit-limit-tokens-minute"); limit != "" {
		if val, err := strconv.ParseInt(limit, 10, 64); err == nil {
			metrics.LimitTokensMinute = val
		}
	}

	if remaining := resp.Header.Get("x-ratelimit-remaining-requests-day"); remaining != "" {
		if val, err := strconv.ParseInt(remaining, 10, 64); err == nil {
			metrics.RemainingRequestsDay = val
		}
	}

	if remaining := resp.Header.Get("x-ratelimit-remaining-tokens-minute"); remaining != "" {
		if val, err := strconv.ParseInt(remaining, 10, 64); err == nil {
			metrics.RemainingTokensMinute = val
		}
	}

	if reset := resp.Header.Get("x-ratelimit-reset-requests-day"); reset != "" {
		if val, err := strconv.ParseInt(reset, 10, 64); err == nil {
			metrics.ResetRequestsDay = val
		}
	}

	if reset := resp.Header.Get("x-ratelimit-reset-tokens-minute"); reset != "" {
		if val, err := strconv.ParseInt(reset, 10, 64); err == nil {
			metrics.ResetTokensMinute = val
		}
	}

	return metrics, nil
}
