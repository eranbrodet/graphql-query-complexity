// Package complexity calculates the GraphQL query complexity.
package complexity

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/graph-gophers/graphql-go/types"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/graphql-query-complexity/internal/query"
)

// connectionComplexity sets complexity value for fields that are type connection
// ojectComplexity sets complexity value for field that are type object
// mutationComplexity sets complexity for mutation operation
const (
	connectionComplexity = 2
	objectComplexity     = 1
	mutationComplexity   = 10
)

type queryState struct {
	variables      map[string]interface{}
	fieldOverrides map[string]int
	fragUsed       map[string]types.SelectionSet
}

// GetQueryComplexity traverses queries and calculates complexity
func GetQueryComplexity(queryString string, variables map[string]interface{}, fieldOverrides map[string]int) (int, error) {
	complexity := 0
	// fragUsed initializes a map with a key: fragment name value: types.Fragment
	fragUsed := make(map[string]types.SelectionSet)
	// parse and lex the provided query string
	executableDefinition, err := query.Parse(queryString)
	if err != nil {
		return 0, err
	}
	// creates map of fragments with name as key
	for _, f := range executableDefinition.Fragments {
		fragUsed[f.Name.Name] = f.Selections
	}

	state := &queryState{
		variables:      variables,
		fieldOverrides: fieldOverrides,
		fragUsed:       fragUsed,
	}

	// for each operation calculate complexity based on operation type and field types
	for _, op := range executableDefinition.Operations {
		switch op.Type {
		case query.Query:
			c, err := calculateSelectionComplexity(op.Selections, state)
			if err != nil {
				return 0, err
			}
			complexity += c
		case query.Mutation:
			c, err := calculateMutationComplexity(op.Selections, state)
			if err != nil {
				return 0, err
			}
			complexity += c
		case query.Subscription:
			// including incase sub cal is needed in the future
			// return complexity
		}
	}
	return complexity, nil
}

// calculateSelectionComplexity calculates and returns complexity for queries
func calculateSelectionComplexity(sels []types.Selection, state *queryState) (int, error) {
	complexity := 0

	for _, sel := range sels {
		switch sel := sel.(type) {
		case *types.Field:
			fieldName := sel.Name.Name
			if isOverride(fieldName, state.fieldOverrides) {
				overrideVal := state.fieldOverrides[fieldName]
				complexity += overrideVal
			} else if fieldName == "pageInfo" {
				continue
			} else if fieldName == "edges" {
				c, err := calculateSelectionComplexity(sel.SelectionSet, state)
				if err != nil {
					return 0, err
				}
				complexity += c
			} else if isConnection(sel.Arguments) {
				c, err := calculateSelectionComplexity(sel.SelectionSet, state)
				if err != nil {
					return 0, err
				}
				itemCount, err := getConnectionNodeCount(sel.Arguments, state.variables)
				if err != nil {
					return 0, err
				}
				complexity += (itemCount * c) + connectionComplexity
			} else {
				if sel.SelectionSet != nil {
					c, err := calculateSelectionComplexity(sel.SelectionSet, state)
					if err != nil {
						return 0, err
					}
					complexity += (c + objectComplexity)
				}
			}
		case *types.FragmentSpread:
			fieldName := sel.Name.Name
			if fragVal, ok := state.fragUsed[fieldName]; ok {
				c, err := calculateSelectionComplexity(fragVal, state)
				if err != nil {
					return 0, err
				}
				complexity += c
			}
		case *types.InlineFragment:
			c, err := calculateSelectionComplexity(sel.Fragment.Selections, state)
			if err != nil {
				return 0, err
			}
			complexity += c
		}
	}
	return complexity, nil
}

// calculateMutationComplexity calculates complexity recursively for mutations
func calculateMutationComplexity(sels []types.Selection, state *queryState) (int, error) {
	complexity := mutationComplexity

	for _, sel := range sels {
		switch sel := sel.(type) {
		case *types.Field:
			for _, x := range sel.SelectionSet {
				switch y := x.(type) {
				case *types.Field:
					c, err := calculateSelectionComplexity(y.SelectionSet, state)
					if err != nil {
						return 0, err
					}
					complexity += c
				case *types.FragmentSpread:
					if fragVal, ok := state.fragUsed[y.Name.Name]; ok {
						c, err := calculateSelectionComplexity(fragVal, state)
						if err != nil {
							return 0, err
						}
						complexity += c
					}
				}
			}
		}
	}
	return complexity, nil
}

// getConnectionNodeCount returns the item count passed into the connection argument
func getConnectionNodeCount(args types.ArgumentList, variables map[string]interface{}) (int, error) {
	itemCount := 0
	for _, a := range args {
		if a.Name.Name == "first" || a.Name.Name == "last" {
			var itemCount int
			switch arg := a.Value.(type) {
			case *types.PrimitiveValue:
				itemCount, _ = strconv.Atoi(arg.String())
			case *types.Variable:

				variableValue, ok := variables[arg.Name]
				if !ok {
					return 0, fmt.Errorf("Variable %s not defined", a.Name.Name)
				}

				switch i := variableValue.(type) {
				case json.Number:
					value, err := i.Int64()
					if err == nil {
						itemCount = int(value)
					}
				case float64:
					itemCount = int(i)
				case float32:
					itemCount = int(i)
				}
			}

			return itemCount, nil
		}
	}
	return itemCount, nil
}

// isConnection checks if a field is a connection
// only fields with args first/last will be counted as a connection
func isConnection(args types.ArgumentList) bool {
	for _, a := range args {
		if a.Name.Name == "first" || a.Name.Name == "last" {
			return true
		}
	}
	return false
}

// isOverride checks if a field has a custom complexity passed in by user
func isOverride(fieldName string, fieldOverrides map[string]int) bool {
	_, ok := fieldOverrides[fieldName]
	return ok
}
