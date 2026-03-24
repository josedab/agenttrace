package graphql

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

// DepthLimitExtension limits the maximum depth of GraphQL queries to prevent
// deeply nested queries from consuming excessive resources.
type DepthLimitExtension struct {
	MaxDepth int
}

var _ interface {
	graphql.OperationContextMutator
	graphql.HandlerExtension
} = DepthLimitExtension{}

func (d DepthLimitExtension) ExtensionName() string {
	return "DepthLimit"
}

func (d DepthLimitExtension) Validate(_ graphql.ExecutableSchema) error {
	return nil
}

func (d DepthLimitExtension) MutateOperationContext(_ context.Context, rc *graphql.OperationContext) *gqlerror.Error {
	op := rc.Doc.Operations.ForName(rc.OperationName)
	if op == nil {
		return nil
	}
	depth := calculateDepth(op.SelectionSet)
	if depth > d.MaxDepth {
		return gqlerror.Errorf("query depth %d exceeds maximum allowed depth of %d", depth, d.MaxDepth)
	}
	return nil
}

func calculateDepth(selectionSet ast.SelectionSet) int {
	if len(selectionSet) == 0 {
		return 0
	}

	maxDepth := 0
	for _, sel := range selectionSet {
		var childDepth int
		switch s := sel.(type) {
		case *ast.Field:
			childDepth = calculateDepth(s.SelectionSet)
		case *ast.InlineFragment:
			childDepth = calculateDepth(s.SelectionSet)
		case *ast.FragmentSpread:
			if s.Definition != nil {
				childDepth = calculateDepth(s.Definition.SelectionSet)
			}
		}
		if childDepth > maxDepth {
			maxDepth = childDepth
		}
	}
	return maxDepth + 1
}
