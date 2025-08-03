package cerebras

// Quota represents rate limit quota information
type Quota struct {
	Limit     int64  `json:"limit"`
	Remaining int64  `json:"remaining"`
	ResetTime string `json:"reset_time"`
}

// UsageQuota represents usage quota information for an organization
type UsageQuota struct {
	ModelId             string `json:"modelId"`
	RegionId            string `json:"regionId"`
	OrganizationId      string `json:"organizationId"`
	RequestsPerMinute   string `json:"requestsPerMinute"`
	TokensPerMinute     string `json:"tokensPerMinute"`
	RequestsPerHour     string `json:"requestsPerHour"`
	TokensPerHour       string `json:"tokensPerHour"`
	RequestsPerDay      string `json:"requestsPerDay"`
	TokensPerDay        string `json:"tokensPerDay"`
	MaxSequenceLength   string `json:"maxSequenceLength"`
	MaxCompletionTokens string `json:"maxCompletionTokens"`
	Typename            string `json:"__typename"`
}

// UsageMetrics represents the usage metrics for an organization
type UsageMetrics struct {
	OrganizationID string  `json:"organization_id"`
	ModelName      string  `json:"model_name"`
	TokensUsed     int64   `json:"tokens_used"`
	TokensLimit    int64   `json:"tokens_limit"`
	RequestsUsed   int64   `json:"requests_used"`
	RequestsLimit  int64   `json:"requests_limit"`
	Quotas         []Quota `json:"quotas"`
}

// OrganizationUsage represents usage data for an organization
type OrganizationUsage struct {
	ModelId  string `json:"modelId"`
	RegionId string `json:"regionId"`
	RPM      string `json:"rpm"`
	TPM      string `json:"tpm"`
	RPH      string `json:"rph"`
	TPH      string `json:"tph"`
	RPD      string `json:"rpd"`
	TPD      string `json:"tpd"`
	Typename string `json:"__typename"`
}

// Organization represents a Cerebras organization
type Organization struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	OrganizationType string `json:"organizationType"`
	State            string `json:"state"`
	Typename         string `json:"__typename"`
}

// OrganizationsResponse represents the response from the organizations endpoint
type OrganizationsResponse struct {
	Organizations []Organization `json:"organizations"`
}
