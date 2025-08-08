package cerebras

import (
	"time"
)

// RateLimitInfo represents comprehensive rate limit information
type RateLimitInfo struct {
	// Limits (prefer GraphQL quotas; fall back to REST headers)
	LimitRequestsMinute int64 `json:"limit_requests_minute,omitempty"`
	LimitRequestsHour   int64 `json:"limit_requests_hour,omitempty"`
	LimitRequestsDay    int64 `json:"limit_requests_day,omitempty"`

	LimitTokensMinute int64 `json:"limit_tokens_minute,omitempty"`
	LimitTokensHour   int64 `json:"limit_tokens_hour,omitempty"`
	LimitTokensDay    int64 `json:"limit_tokens_day,omitempty"`

	// Usage (prefer GraphQL usage; else derive as limit - remaining when available)
	UsageRequestsMinute int64 `json:"usage_requests_minute,omitempty"`
	UsageRequestsHour   int64 `json:"usage_requests_hour,omitempty"`
	UsageRequestsDay    int64 `json:"usage_requests_day,omitempty"`

	UsageTokensMinute int64 `json:"usage_tokens_minute,omitempty"`
	UsageTokensHour   int64 `json:"usage_tokens_hour,omitempty"`
	UsageTokensDay    int64 `json:"usage_tokens_day,omitempty"`

	// Remaining (prefer REST headers when provided; otherwise computed)
	RemainingRequestsMinute int64 `json:"remaining_requests_minute,omitempty"`
	RemainingRequestsHour   int64 `json:"remaining_requests_hour,omitempty"`
	RemainingRequestsDay    int64 `json:"remaining_requests_day,omitempty"`

	RemainingTokensMinute int64 `json:"remaining_tokens_minute,omitempty"`
	RemainingTokensHour   int64 `json:"remaining_tokens_hour,omitempty"`
	RemainingTokensDay    int64 `json:"remaining_tokens_day,omitempty"`

	// Reset (seconds until reset; REST often provides for minute/day)
	ResetRequestsMinute int64 `json:"reset_requests_minute,omitempty"`
	ResetRequestsHour   int64 `json:"reset_requests_hour,omitempty"`
	ResetRequestsDay    int64 `json:"reset_requests_day,omitempty"`

	ResetTokensMinute int64 `json:"reset_tokens_minute,omitempty"`
	ResetTokensHour   int64 `json:"reset_tokens_hour,omitempty"`
	ResetTokensDay    int64 `json:"reset_tokens_day,omitempty"`

	// Metadata (from GraphQL quotas)
	ModelId             string `json:"model_id,omitempty"`
	RegionId            string `json:"region_id,omitempty"`
	MaxSequenceLength   int64  `json:"max_sequence_length,omitempty"`
	MaxCompletionTokens int64  `json:"max_completion_tokens,omitempty"`

	// Backward compatibility fields
	LimitRequestsDayOld      int64 `json:"limit_requests_day,omitempty"`
	LimitTokensMinuteOld     int64 `json:"limit_tokens_minute,omitempty"`
	RemainingRequestsDayOld  int64 `json:"remaining_requests_day,omitempty"`
	RemainingTokensMinuteOld int64 `json:"remaining_tokens_minute,omitempty"`
	ResetRequestsDayOld      int64 `json:"reset_requests_day,omitempty"`
	ResetTokensMinuteOld     int64 `json:"reset_tokens_minute,omitempty"`
}

// ToQuota converts RateLimitInfo to Quota
func (r *RateLimitInfo) ToQuota() *Quota {
	// Converting requests per day quota
	resetTime := "Unknown"
	if r.ResetRequestsDay > 0 {
		// The reset time is a relative timestamp in seconds from now
		resetTime = time.Now().Add(time.Duration(r.ResetRequestsDay) * time.Second).Format(time.RFC3339)
	}

	return &Quota{
		Limit:     r.LimitRequestsDay,
		Remaining: r.RemainingRequestsDay,
		ResetTime: resetTime,
	}
}

// ToUsageMetrics converts RateLimitInfo to UsageMetrics
func (r *RateLimitInfo) ToUsageMetrics(orgID, modelName string) *UsageMetrics {
	// Note: We're making assumptions about which limits correspond to which quotas
	// This is a basic conversion and may not be accurate for all use cases
	return &UsageMetrics{
		OrganizationID: orgID,
		ModelName:      modelName,
		TokensUsed:     r.LimitTokensMinute - r.RemainingTokensMinute,
		TokensLimit:    r.LimitTokensMinute,
		RequestsUsed:   r.LimitRequestsDay - r.RemainingRequestsDay,
		RequestsLimit:  r.LimitRequestsDay,
		Quotas:         []Quota{*r.ToQuota()},
	}
}

// Quota represents rate limit quota information
type Quota struct {
	Limit     int64  `json:"limit,omitempty"`
	Remaining int64  `json:"remaining,omitempty"`
	ResetTime string `json:"reset_time,omitempty"`
}

// UsageQuota represents usage quota information for an organization
type UsageQuota struct {
	ModelId             string `json:"modelId,omitempty"`
	RegionId            string `json:"regionId,omitempty"`
	OrganizationId      string `json:"organizationId,omitempty"`
	RequestsPerMinute   string `json:"requestsPerMinute,omitempty"`
	TokensPerMinute     string `json:"tokensPerMinute,omitempty"`
	RequestsPerHour     string `json:"requestsPerHour,omitempty"`
	TokensPerHour       string `json:"tokensPerHour,omitempty"`
	RequestsPerDay      string `json:"requestsPerDay,omitempty"`
	TokensPerDay        string `json:"tokensPerDay,omitempty"`
	MaxSequenceLength   string `json:"maxSequenceLength,omitempty"`
	MaxCompletionTokens string `json:"maxCompletionTokens,omitempty"`
	Typename            string `json:"__typename,omitempty"`
}

// UsageMetrics represents the usage metrics for an organization
type UsageMetrics struct {
	OrganizationID string  `json:"organization_id,omitempty"`
	ModelName      string  `json:"model_name,omitempty"`
	TokensUsed     int64   `json:"tokens_used,omitempty"`
	TokensLimit    int64   `json:"tokens_limit,omitempty"`
	RequestsUsed   int64   `json:"requests_used,omitempty"`
	RequestsLimit  int64   `json:"requests_limit,omitempty"`
	Quotas         []Quota `json:"quotas,omitempty"`
}

// OrganizationUsage represents usage data for an organization
type OrganizationUsage struct {
	ModelId  string `json:"modelId,omitempty"`  // Model identifier
	RegionId string `json:"regionId,omitempty"` // Region identifier
	RPM      string `json:"rpm,omitempty"`      // Requests Per Minute
	TPM      string `json:"tpm,omitempty"`      // Tokens Per Minute
	RPH      string `json:"rph,omitempty"`      // Requests Per Hour
	TPH      string `json:"tph,omitempty"`      // Tokens Per Hour
	RPD      string `json:"rpd,omitempty"`      // Requests Per Day
	TPD      string `json:"tpd,omitempty"`      // Tokens Per Day
	Typename string `json:"__typename,omitempty"`
}

// Organization represents a Cerebras organization
type Organization struct {
	ID               string `json:"id,omitempty"`
	Name             string `json:"name,omitempty"`
	OrganizationType string `json:"organizationType,omitempty"`
	State            string `json:"state,omitempty"`
	Typename         string `json:"__typename,omitempty"`
}

// OrganizationsResponse represents the response from the organizations endpoint
type OrganizationsResponse struct {
	Organizations []Organization `json:"organizations"`
}
