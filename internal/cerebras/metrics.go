package cerebras

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/spf13/viper"
)

// GetMetrics fetches usage metrics from Cerebras servers
func (c *Client) GetMetrics(organization string) (*RateLimitInfo, error) {
	if !c.HasAuth() {
		return nil, fmt.Errorf("no authentication method configured")
	}

	// Determine which authentication method to use
	// Prefer session token (GraphQL) when organization is provided, for live usage/quotas
	if c.sessionToken != "" && organization != "" {
		return c.getMetricsWithSessionToken(organization)
	}

	// Otherwise, fall back to API key (REST) which exposes rate limit headers
	if c.apiKey != "" {
		return c.getMetricsWithAPIKey()
	}

	// As a last resort, if only session token is available but no organization provided
	if c.sessionToken != "" {
		return nil, fmt.Errorf("organization ID is required when using session token authentication")
	}

	return nil, fmt.Errorf("no valid authentication method found")
}

// getMetricsWithSessionToken fetches metrics using GraphQL with session token auth
func (c *Client) getMetricsWithSessionToken(organization string) (*RateLimitInfo, error) {
	// Make GraphQL request to list organization usage quotas
	// We will map the quota limits to our RateLimitInfo structure.
	query := `query ListOrganizationUsageQuotas($organizationId: ID!, $modelId: ID, $regionId: ID) {
  ListOrganizationUsageQuotas(
    organizationId: $organizationId
    modelId: $modelId
    regionId: $regionId
  ) {
    modelId
    regionId
    organizationId
    requestsPerMinute
    tokensPerMinute
    requestsPerHour
    tokensPerHour
    requestsPerDay
    tokensPerDay
    maxSequenceLength
    maxCompletionTokens
    __typename
  }
}`

	variables := map[string]interface{}{
		"organizationId": organization,
	}

	responseBody, err := c.MakeGraphQLRequestWithDebug(query, variables, viper.GetBool("debug"))
	if err != nil {
		return nil, err
	}

	// Parse the GraphQL response
	var response struct {
		Data struct {
			ListOrganizationUsageQuotas []UsageQuota `json:"ListOrganizationUsageQuotas"`
		} `json:"data"`
	}

	if err := json.Unmarshal(responseBody, &response); err != nil {
		return nil, fmt.Errorf("failed to parse GraphQL response: %w", err)
	}

	// If no quotas returned, provide empty metrics
	if len(response.Data.ListOrganizationUsageQuotas) == 0 {
		return &RateLimitInfo{}, nil
	}

	// Try to find quota for configured model, otherwise use the first one
	model := viper.GetString("model")
	selected := response.Data.ListOrganizationUsageQuotas[0]
	if model != "" {
		for _, q := range response.Data.ListOrganizationUsageQuotas {
			if q.ModelId == model {
				selected = q
				break
			}
		}
	}

	// Helper to parse string to int64; returns 0 on error or "-1" sentinel
	parse := func(s string) int64 {
		if s == "" {
			return 0
		}
		v, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return 0
		}
		if v < 0 { // treat -1 (unlimited) as 0 limit (unknown)
			return 0
		}
		return v
	}

	limits := &RateLimitInfo{
		LimitRequestsDay:      parse(selected.RequestsPerDay),
		LimitTokensMinute:     parse(selected.TokensPerMinute),
		RemainingRequestsDay:  0, // will compute below
		RemainingTokensMinute: 0, // will compute below
		ResetRequestsDay:      0,
		ResetTokensMinute:     0,
	}

	// Additionally fetch current usage to compute "remaining" values so UI shows progress
	usageQuery := `query ListOrganizationUsage($organizationId: ID!) {
  ListOrganizationUsage(organizationId: $organizationId) {
    modelId
    regionId
    rpm
    tpm
    rph
    tph
    rpd
    tpd
    __typename
  }
}`

	usageBody, err := c.MakeGraphQLRequestWithDebug(usageQuery, map[string]interface{}{"organizationId": organization}, viper.GetBool("debug"))
	if err == nil { // best-effort; if it fails, we still return limits
		var usageResp struct {
			Data struct {
				ListOrganizationUsage []OrganizationUsage `json:"ListOrganizationUsage"`
			} `json:"data"`
		}
		if err := json.Unmarshal(usageBody, &usageResp); err == nil {
			// Match by model and (if available) region
			var matched *OrganizationUsage
			for i := range usageResp.Data.ListOrganizationUsage {
				u := &usageResp.Data.ListOrganizationUsage[i]
				if u.ModelId == selected.ModelId {
					if selected.RegionId == "" || u.RegionId == selected.RegionId {
						matched = u
						break
					}
				}
			}

			if matched != nil {
				// Parse token/minute usage and requests/day usage
				usedTPM := parse(matched.TPM)
				usedRPD := parse(matched.RPD)

				if limits.LimitTokensMinute > 0 {
					remTPM := limits.LimitTokensMinute - usedTPM
					if remTPM < 0 {
						remTPM = 0
					}
					limits.RemainingTokensMinute = remTPM
				}
				if limits.LimitRequestsDay > 0 {
					remRPD := limits.LimitRequestsDay - usedRPD
					if remRPD < 0 {
						remRPD = 0
					}
					limits.RemainingRequestsDay = remRPD
				}
			} else {
				// Fallback: if usage not found, set remaining equal to limits
				limits.RemainingRequestsDay = limits.LimitRequestsDay
				limits.RemainingTokensMinute = limits.LimitTokensMinute
			}
		} else {
			// Parsing usage failed; set remaining equal to limits
			limits.RemainingRequestsDay = limits.LimitRequestsDay
			limits.RemainingTokensMinute = limits.LimitTokensMinute
		}
	} else {
		// Usage request failed; set remaining equal to limits
		limits.RemainingRequestsDay = limits.LimitRequestsDay
		limits.RemainingTokensMinute = limits.LimitTokensMinute
	}

	return limits, nil
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
