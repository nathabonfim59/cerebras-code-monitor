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
