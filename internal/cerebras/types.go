package cerebras

// Quota represents rate limit quota information
type Quota struct {
	Limit     int64  `json:"limit"`
	Remaining int64  `json:"remaining"`
	ResetTime string `json:"reset_time"`
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

// Organization represents a Cerebras organization
type Organization struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// OrganizationsResponse represents the response from the organizations endpoint
type OrganizationsResponse struct {
	Organizations []Organization `json:"organizations"`
}
