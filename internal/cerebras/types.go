package cerebras

// RateLimitInfo represents comprehensive rate limit information
type RateLimitInfo struct {
	LimitRequestsDay      int64 `json:"limit_requests_day,omitempty"`
	LimitTokensMinute     int64 `json:"limit_tokens_minute,omitempty"`
	RemainingRequestsDay  int64 `json:"remaining_requests_day,omitempty"`
	RemainingTokensMinute int64 `json:"remaining_tokens_minute,omitempty"`
	ResetRequestsDay      int64 `json:"reset_requests_day,omitempty"`
	ResetTokensMinute     int64 `json:"reset_tokens_minute,omitempty"`
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
	ModelId  string `json:"modelId,omitempty"`
	RegionId string `json:"regionId,omitempty"`
	RPM      string `json:"rpm,omitempty"`
	TPM      string `json:"tpm,omitempty"`
	RPH      string `json:"rph,omitempty"`
	TPH      string `json:"tph,omitempty"`
	RPD      string `json:"rpd,omitempty"`
	TPD      string `json:"tpd,omitempty"`
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
