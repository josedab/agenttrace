package domain

import "testing"

func TestAPIKeyContextHasScope(t *testing.T) {
	t.Parallel()

	context := APIKeyContext{Scopes: []string{"traces:read", "scores:*"}}
	if !context.HasScope("traces:read") {
		t.Fatal("expected exact scope to match")
	}
	if !context.HasScope("scores:write") {
		t.Fatal("expected wildcard scope to match")
	}
	if context.HasScope("traces:write") {
		t.Fatal("unexpected scope match")
	}

	admin := APIKeyContext{Scopes: []string{"admin:write"}}
	if !admin.HasScope("datasets:delete") {
		t.Fatal("expected admin:write to permit scoped operations")
	}
}
