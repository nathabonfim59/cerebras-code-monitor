package graphql

// ListOrganizationsQuery is the GraphQL query for listing organizations
const ListOrganizationsQuery = `query ListMyOrganizations {
  ListMyOrganizations {
    id
    name
    organizationType
    state
    __typename
  }
}`

// ListOrganizationUsageQuotasQuery is the GraphQL query for listing organization usage quotas
const ListOrganizationUsageQuotasQuery = `query ListOrganizationUsageQuotas($organizationId: ID!, $modelId: ID, $regionId: ID) {
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

// ListOrganizationUsageQuery is the GraphQL query for listing organization usage
const ListOrganizationUsageQuery = `query ListOrganizationUsage($organizationId: ID!) {
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
