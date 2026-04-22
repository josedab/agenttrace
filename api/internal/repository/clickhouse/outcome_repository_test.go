package clickhouse

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOutcomeQueriesRemainProjectScoped(t *testing.T) {
	queries := []string{
		outcomeTraceAggregateQuery,
		outcomeGitAggregateQuery,
		outcomeCIAggregateQuery,
		outcomeLinkedCIAggregateQuery,
		outcomeAgentBreakdownQuery,
		outcomeModelBreakdownQuery,
		outcomeRecentQuery,
	}

	for _, query := range queries {
		normalized := strings.Join(strings.Fields(query), " ")
		assert.Contains(t, normalized, "project_id = ?")
		assert.Contains(t, normalized, ">= ?")
		assert.Contains(t, normalized, "< ?")
	}
}
