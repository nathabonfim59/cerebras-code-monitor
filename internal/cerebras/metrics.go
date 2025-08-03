package cerebras

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/spf13/viper"
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
	// Prioritize API key over session token
	if c.apiKey != "" {
		// API key doesn't need organization parameter
		return c.getMetricsWithAPIKey()
	} else if c.sessionToken != "" {
		return c.getMetricsWithSessionToken(organization)
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
	// Make a chat completion request to get rate limit headers
	url := fmt.Sprintf("%s/v1/chat/completions", c.baseURL)

	// Create a minimal request body that should work
	body := `{
		"model": "llama3.1-8b",
		"messages": [{"role": "user", "content": "hello"}],
		"max_completion_tokens": 1
	}`

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
	defer resp.Body.Close()

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
	metrics := &Metrics{}
	if limit := resp.Header.Get("X-Ratelimit-Limit-Requests-Day"); limit != "" {
		if val, err := strconv.ParseInt(limit, 10, 64); err == nil {
			metrics.LimitRequestsDay = val
		}
	}

	if limit := resp.Header.Get("X-Ratelimit-Limit-Tokens-Minute"); limit != "" {
		if val, err := strconv.ParseInt(limit, 10, 64); err == nil {
			metrics.LimitTokensMinute = val
		}
	}

	if remaining := resp.Header.Get("X-Ratelimit-Remaining-Requests-Day"); remaining != "" {
		if val, err := strconv.ParseInt(remaining, 10, 64); err == nil {
			metrics.RemainingRequestsDay = val
		}
	}

	if remaining := resp.Header.Get("X-Ratelimit-Remaining-Tokens-Minute"); remaining != "" {
		if val, err := strconv.ParseInt(remaining, 10, 64); err == nil {
			metrics.RemainingTokensMinute = val
		}
	}

	if reset := resp.Header.Get("X-Ratelimit-Reset-Requests-Day"); reset != "" {
		if val, err := strconv.ParseFloat(reset, 64); err == nil {
			metrics.ResetRequestsDay = int64(val)
		}
	}

	if reset := resp.Header.Get("X-Ratelimit-Reset-Tokens-Minute"); reset != "" {
		if val, err := strconv.ParseFloat(reset, 64); err == nil {
			metrics.ResetTokensMinute = int64(val)
		}
	}
	// If we got rate limit headers, return the metrics even if the request failed
	if metrics.LimitRequestsDay > 0 || metrics.LimitTokensMinute > 0 || metrics.RemainingRequestsDay > 0 || metrics.RemainingTokensMinute > 0 {
		return metrics, nil
	}

	// Otherwise, return an error
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status code: %d", resp.StatusCode)
	}

	return metrics, nil
}
