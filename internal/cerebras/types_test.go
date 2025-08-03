package cerebras

import (
	"testing"
	"time"
)

func TestRateLimitInfoToQuota(t *testing.T) {
	// Example data from your API responses
	// Using relative timestamps (seconds from now) as that's what the API actually returns
	rateLimit := &RateLimitInfo{
		LimitRequestsDay:      28800,
		LimitTokensMinute:     275000,
		RemainingRequestsDay:  28182,
		RemainingTokensMinute: 190760,
		ResetRequestsDay:      3600, // 1 hour from now
		ResetTokensMinute:     60,   // 1 minute from now
	}

	quota := rateLimit.ToQuota()

	if quota.Limit != rateLimit.LimitRequestsDay {
		t.Errorf("Expected Limit %d, got %d", rateLimit.LimitRequestsDay, quota.Limit)
	}

	if quota.Remaining != rateLimit.RemainingRequestsDay {
		t.Errorf("Expected Remaining %d, got %d", rateLimit.RemainingRequestsDay, quota.Remaining)
	}

	// Check that reset time is in the future (approximately 1 hour from now)
	expectedResetTime := time.Now().Add(time.Duration(rateLimit.ResetRequestsDay) * time.Second)
	resetTime, err := time.Parse(time.RFC3339, quota.ResetTime)
	if err != nil {
		t.Errorf("Error parsing reset time: %v", err)
	}

	// Check that the parsed time is within 5 seconds of expected time
	diff := resetTime.Sub(expectedResetTime)
	if diff > 5*time.Second || diff < -5*time.Second {
		t.Errorf("Expected ResetTime approximately %s, got %s", expectedResetTime.Format(time.RFC3339), quota.ResetTime)
	}
}
func TestRateLimitInfoToUsageMetrics(t *testing.T) {
	// Example data from your API responses
	rateLimit := &RateLimitInfo{
		LimitRequestsDay:      28800,
		LimitTokensMinute:     275000,
		RemainingRequestsDay:  28182,
		RemainingTokensMinute: 190760,
		ResetRequestsDay:      1722729600, // Unix timestamp for 2025-08-04T00:00:00Z
		ResetTokensMinute:     1722729600, // Unix timestamp for 2025-08-04T00:00:00Z
	}

	orgID := "org_yc4f58xph5d2vndemvddrvww"
	modelName := "qwen-3-coder-480b"

	usageMetrics := rateLimit.ToUsageMetrics(orgID, modelName)

	if usageMetrics.OrganizationID != orgID {
		t.Errorf("Expected OrganizationID %s, got %s", orgID, usageMetrics.OrganizationID)
	}

	if usageMetrics.ModelName != modelName {
		t.Errorf("Expected ModelName %s, got %s", modelName, usageMetrics.ModelName)
	}

	expectedTokensUsed := rateLimit.LimitTokensMinute - rateLimit.RemainingTokensMinute
	if usageMetrics.TokensUsed != expectedTokensUsed {
		t.Errorf("Expected TokensUsed %d, got %d", expectedTokensUsed, usageMetrics.TokensUsed)
	}

	if usageMetrics.TokensLimit != rateLimit.LimitTokensMinute {
		t.Errorf("Expected TokensLimit %d, got %d", rateLimit.LimitTokensMinute, usageMetrics.TokensLimit)
	}

	expectedRequestsUsed := rateLimit.LimitRequestsDay - rateLimit.RemainingRequestsDay
	if usageMetrics.RequestsUsed != expectedRequestsUsed {
		t.Errorf("Expected RequestsUsed %d, got %d", expectedRequestsUsed, usageMetrics.RequestsUsed)
	}

	if usageMetrics.RequestsLimit != rateLimit.LimitRequestsDay {
		t.Errorf("Expected RequestsLimit %d, got %d", rateLimit.LimitRequestsDay, usageMetrics.RequestsLimit)
	}

	if len(usageMetrics.Quotas) != 1 {
		t.Errorf("Expected 1 quota, got %d", len(usageMetrics.Quotas))
	}
}
