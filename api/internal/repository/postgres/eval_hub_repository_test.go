package postgres

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEvalHubAccessClausePreservesVisibilityBoundaries(t *testing.T) {
	assert.Contains(t, evalHubAccessClause, "owner_project_id = $1")
	assert.Contains(t, evalHubAccessClause, "visibility = 'public'")
	assert.Contains(t, evalHubAccessClause, "visibility = 'organization'")
	assert.Contains(t, evalHubAccessClause, "organization_id = $2")
	assert.NotContains(t, evalHubAccessClause, "visibility = 'private' OR")
}
