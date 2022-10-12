package complexity_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	complexity "gitlab.com/infor-cloud/martian-cloud/tharsis/graphql-query-complexity"
)

func TestGetQueryComplexity(t *testing.T) {

	tests := []struct {
		name           string
		query          string
		variables      map[string]interface{}
		fieldOverrides map[string]int
		wantComplexity int
		wantErrMessage string
	}{
		{
			name: "Connection with 5 items",
			query: `query{
				groups(first: 5, sort: FULL_PATH_ASC) {
				  edges {
					node {
					  id
					}
				  }
				}
			  }`,
			wantComplexity: 7,
		},
		{
			name: "Connection with 5 items using a variable",
			query: `query GetGroups($first: Int){
				groups(first: $first, sort: FULL_PATH_ASC) {
				  edges {
					node {
					  id
					}
				  }
				}
			  }`,
			variables:      map[string]interface{}{"first": float64(5)},
			wantComplexity: 7,
		},
		{
			name: "query with nested fragments",
			query: `query{
				me {
					... on User {
					  memberships(first:100) {
						edges {
						  node {
							id
							role
							namespace {
							  ... on Group {
								fullPath
							  }
							  ... on Workspace {
								fullPath
							  }
							}
						  }
						}
					  }
					}
				  }
			  }`,
			wantComplexity: 203,
		},
		{
			name: "Connection with missing variable",
			query: `query GetGroups($first: Int){
				groups(first: $first, sort: FULL_PATH_ASC) {
				  edges {
					node {
					  id
					}
				  }
				}
			  }`,
			variables:      map[string]interface{}{},
			wantErrMessage: "Variable first not defined",
		},
		{
			name: "First connection is 0",
			query: `query{
				groups(first: 0, sort: FULL_PATH_ASC) {
				  pageInfo {
					hasNextPage
					hasPreviousPage
				  }
				  edges {
					node {
					  id
					  name
					  fullPath
					  parent {
						id
						name
					  }
					}
				  }
				}
			  }`,
			fieldOverrides: map[string]int{"node": 2, "parent": 2},
			wantComplexity: 2,
		},
		{
			name: "Single object calculation",
			query: `query {
				group(fullPath: "colonies") {
				  name
				}
			  }`,
			wantComplexity: 1,
		},
		{
			name: "Multiple subconnections, one with 0 arg",
			query: `{
				groups(first: 2) {
				  edges {
					node {
					  id
					  name
					  parent {
						id
					  }
					  decendentGroups(last: 3) {
						edges {
						  node {
							id
							name
							decendentGroups(first: 4) {
							  edges {
								node {
								  id
								  name
								  parent {
									id
									name
								  }
								  decendentGroups(last: 0) {
									edges {
									  node {
										id
										name
										description
									  }
									}
								  }
								}
							  }
							}
						  }
						}
					  }
					}
				  }
				}
			  }`,
			wantComplexity: 124,
		},
		{
			name: "Mutation-Create calculation",
			query: `mutation {
				createGroup(
				  input: {name: "testGroup3", description: "testing create group", parentId: "0d253183-aa18-419b-8bc5-d0535f52d8c9"}
				) {
				  group {
					id
					name
					decendentGroups(first: 1) {
					  edges {
						node {
						  id
						  name
						  parent {
							id
							name
						  }
						}
					  }
					}
				  }
				  problems {
					message
				  }
				}
			  }`,
			wantComplexity: 14,
		},
		{
			name: "Mutation-Create: calculation no parent",
			query: `mutation {
				createGroup(
				  input: {name: "testGroup3", description: "testing create group", parentId: "0d253183-aa18-419b-8bc5-d0535f52d8c9"}
				) {
				  group {
					id
					name
					decendentGroups(first: 3) {
					  edges {
						node {
						  id
						  name
						}
					  }
					}
				  }
				  problems {
					message
				  }
				}
			  }`,
			wantComplexity: 15,
		},
		{
			name: "Mutation-Update Group",
			query: `mutation {
				updateGroup(input: {id: "0d253183-aa18-419b-8bc5-d0535f52d8c9", clientMutationId:"test", description:"testupdate"}) {
				  clientMutationId
				  problems {
					message
				  }
				}
			  }`,
			wantComplexity: 10,
		},
		{
			name: "Fragment Calculation on connection",
			query: `{
				groups(first: 5) {
				  ...connection
				}
			  }

			  fragment connection on GroupConnection {
				pageInfo {
				  endCursor
				  startCursor
				}
				edges {
				  node {
					id
					name
					parent {
					  name
					}
				  }
				}
			  }
			  `,
			wantComplexity: 12,
		},
		{
			name: "Multiple Fragment Calculation",
			query: `{
				groups(first: 5) {
				  ...connection
				  ...test
				}
			  }

			  fragment connection on GroupConnection {
				pageInfo {
				  endCursor
				  startCursor
				}
			  }

			  fragment test on GroupConnection {
				edges {
				  node {
					id
					name
					parent {
					  id
					  name
					}
				  }
				}
			  }`,
			wantComplexity: 12,
		},
		{
			name: "Fragment Calculation with sub connection",
			query: `{
				groups(first: 5) {
				  ...connection
				  ...test
				}
			  }

			  fragment connection on GroupConnection {
				edges {
				  node {
					id
					name
					decendentGroups (first: 1) {
					  edges {
						node {
						  id
						  name
						}
					  }
					}
				  }
				}
			  }

			  fragment test on GroupConnection {
				pageInfo {
					endCursor
					startCursor
				  }
				}`,
			wantComplexity: 22,
		},
		{
			name: "Override",
			query: `{
				groups(first: 2) {
				  edges {
					node {
					  id
					  name
					  parent {
						id
					  }
				    }
				  }
				}
			  }`,
			fieldOverrides: map[string]int{"parent": 2, "node": 2},
			wantComplexity: 6,
		},
		{
			name: "Override-connection",
			query: `{
				groups(first: 2) {
				  edges {
					node {
					  id
					  name
					  parent {
						id
					  }
				    }
				  }
				}
			  }`,
			fieldOverrides: map[string]int{"groups": 4},
			wantComplexity: 4,
		},
		{
			name: "Mutation-Frag Test",
			query: `fragment groupFragment on Group {
				edges {
					node {
					  id
					  name
					  parent {
						  id
						  name
					  }
					}
				}
			}

				mutation {
					createGroup(
					  input: {name: "testGroup3", description: "testing create group", parentId: "0d253183-aa18-419b-8bc5-d0535f52d8c9"}
					) {
					  ...groupFragment
					  problems {
						  message
					  }
					}
				}`,
			wantComplexity: 12,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual, err := complexity.GetQueryComplexity(test.query, test.variables, test.fieldOverrides)
			if test.wantErrMessage != "" {
				assert.EqualError(t, err, test.wantErrMessage)
			} else if err != nil {
				t.Fatal(err)
			} else {
				assert.Equal(t, test.wantComplexity, actual)
			}
		})
	}
}
