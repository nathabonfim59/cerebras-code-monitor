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

	// Prefer GraphQL (session token + organization) for richer data
	if c.sessionToken != "" && organization != "" {
		gql, err := c.getMetricsWithSessionToken(organization)
		if err == nil {
			// If we also have an API key, try to enrich with REST (resets/remaining)
			if c.apiKey != "" {
				if rest, rerr := c.getMetricsWithAPIKey(); rerr == nil && rest != nil {
					// Use REST resets when GraphQL lacks them
					if gql.ResetRequestsDay == 0 && rest.ResetRequestsDay > 0 {
						gql.ResetRequestsDay = rest.ResetRequestsDay
					}
					if gql.ResetTokensMinute == 0 && rest.ResetTokensMinute > 0 {
						gql.ResetTokensMinute = rest.ResetTokensMinute
					}
					// If GraphQL didn't compute remainings, take REST values
					if gql.LimitRequestsDay > 0 && gql.RemainingRequestsDay == 0 && rest.RemainingRequestsDay > 0 {
						gql.RemainingRequestsDay = rest.RemainingRequestsDay
					}
					if gql.LimitTokensMinute > 0 && gql.RemainingTokensMinute == 0 && rest.RemainingTokensMinute > 0 {
						gql.RemainingTokensMinute = rest.RemainingTokensMinute
					}
				}
			}
			return gql, nil
		}
		// If GraphQL failed, fall through to REST
	}

	// Fallback to REST headers when available
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
		LimitRequestsMinute:   parse(selected.RequestsPerMinute),
		LimitRequestsHour:     parse(selected.RequestsPerHour),
		LimitRequestsDay:      parse(selected.RequestsPerDay),
		LimitTokensMinute:     parse(selected.TokensPerMinute),
		LimitTokensHour:       parse(selected.TokensPerHour),
		LimitTokensDay:        parse(selected.TokensPerDay),
		ResetRequestsDay:      0,
		ResetTokensMinute:     0,
		ModelId:               selected.ModelId,
		RegionId:              selected.RegionId,
		MaxSequenceLength:     parse(selected.MaxSequenceLength),
		MaxCompletionTokens:   parse(selected.MaxCompletionTokens),
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
				// Parse usage across minute/hour/day for requests and tokens
				usedRPM := parse(matched.RPM)
				usedTPM := parse(matched.TPM)
				usedRPH := parse(matched.RPH)
				usedTPH := parse(matched.TPH)
				usedRPD := parse(matched.RPD)
				usedTPD := parse(matched.TPD)

				limits.UsageRequestsMinute = usedRPM
				limits.UsageTokensMinute = usedTPM
				limits.UsageRequestsHour = usedRPH
				limits.UsageTokensHour = usedTPH
				limits.UsageRequestsDay = usedRPD
				limits.UsageTokensDay = usedTPD

				// Compute remaining when limits are known
				if limits.LimitTokensMinute > 0 {
					rem := limits.LimitTokensMinute - usedTPM
					if rem < 0 { rem = 0 }
					limits.RemainingTokensMinute = rem
				}
				if limits.LimitTokensHour > 0 {
					rem := limits.LimitTokensHour - usedTPH
					if rem < 0 { rem = 0 }
					limits.RemainingTokensHour = rem
				}
				if limits.LimitTokensDay > 0 {
					rem := limits.LimitTokensDay - usedTPD
					if rem < 0 { rem = 0 }
					limits.RemainingTokensDay = rem
				}
				if limits.LimitRequestsMinute > 0 {
					rem := limits.LimitRequestsMinute - usedRPM
					if rem < 0 { rem = 0 }
					limits.RemainingRequestsMinute = rem
				}
				if limits.LimitRequestsHour > 0 {
					rem := limits.LimitRequestsHour - usedRPH
					if rem < 0 { rem = 0 }
					limits.RemainingRequestsHour = rem
				}
				if limits.LimitRequestsDay > 0 {
					rem := limits.LimitRequestsDay - usedRPD
					if rem < 0 { rem = 0 }
					limits.RemainingRequestsDay = rem
				}
			} else {
				// Fallback: if usage not found, set remaining equal to known limits
				limits.RemainingRequestsMinute = limits.LimitRequestsMinute
				limits.RemainingRequestsHour = limits.LimitRequestsHour
				limits.RemainingRequestsDay = limits.LimitRequestsDay
				limits.RemainingTokensMinute = limits.LimitTokensMinute
				limits.RemainingTokensHour = limits.LimitTokensHour
				limits.RemainingTokensDay = limits.LimitTokensDay
			}
		} else {
			// Parsing usage failed; set remaining equal to known limits
			limits.RemainingRequestsMinute = limits.LimitRequestsMinute
			limits.RemainingRequestsHour = limits.LimitRequestsHour
			limits.RemainingRequestsDay = limits.LimitRequestsDay
			limits.RemainingTokensMinute = limits.LimitTokensMinute
			limits.RemainingTokensHour = limits.LimitTokensHour
			limits.RemainingTokensDay = limits.LimitTokensDay
		}
	} else {
		// Usage request failed; set remaining equal to known limits
		limits.RemainingRequestsMinute = limits.LimitRequestsMinute
		limits.RemainingRequestsHour = limits.LimitRequestsHour
		limits.RemainingRequestsDay = limits.LimitRequestsDay
		limits.RemainingTokensMinute = limits.LimitTokensMinute
		limits.RemainingTokensHour = limits.LimitTokensHour
		limits.RemainingTokensDay = limits.LimitTokensDay
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

	// Derive usage from remaining when possible
	if rateLimitInfo.LimitRequestsDay > 0 && rateLimitInfo.RemainingRequestsDay >= 0 {
		rateLimitInfo.UsageRequestsDay = rateLimitInfo.LimitRequestsDay - rateLimitInfo.RemainingRequestsDay
		if rateLimitInfo.UsageRequestsDay < 0 { rateLimitInfo.UsageRequestsDay = 0 }
	}
	if rateLimitInfo.LimitTokensMinute > 0 && rateLimitInfo.RemainingTokensMinute >= 0 {
		rateLimitInfo.UsageTokensMinute = rateLimitInfo.LimitTokensMinute - rateLimitInfo.RemainingTokensMinute
		if rateLimitInfo.UsageTokensMinute < 0 { rateLimitInfo.UsageTokensMinute = 0 }
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
