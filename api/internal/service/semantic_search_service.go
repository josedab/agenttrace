package service

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

// SemanticSearchService handles semantic trace search logic
type SemanticSearchService struct {
	logger       *zap.Logger
	queryService *QueryService
}

// NewSemanticSearchService creates a new semantic search service
func NewSemanticSearchService(logger *zap.Logger, queryService *QueryService) *SemanticSearchService {
	return &SemanticSearchService{
		logger:       logger,
		queryService: queryService,
	}
}

// Search performs a hybrid keyword/semantic search across traces
func (s *SemanticSearchService) Search(ctx context.Context, projectID uuid.UUID, query domain.SemanticSearchQuery) (*domain.SemanticSearchResult, error) {
	start := time.Now()

	s.logger.Debug("performing semantic search",
		zap.String("projectId", projectID.String()),
		zap.String("query", query.Query),
	)

	limit := query.Limit
	if limit <= 0 {
		limit = 20
	}

	// Simulate hybrid search with mock results
	keywords := strings.Fields(strings.ToLower(query.Query))
	results := s.generateMockResults(keywords, limit)

	elapsed := time.Since(start)

	return &domain.SemanticSearchResult{
		Results:     results,
		TotalCount:  len(results),
		QueryTimeMs: elapsed.Milliseconds(),
	}, nil
}

// GetSuggestions returns autocomplete suggestions for a search prefix
func (s *SemanticSearchService) GetSuggestions(ctx context.Context, projectID uuid.UUID, prefix string) ([]string, error) {
	s.logger.Debug("getting search suggestions",
		zap.String("projectId", projectID.String()),
		zap.String("prefix", prefix),
	)

	allSuggestions := []string{
		"error in authentication flow",
		"high latency traces",
		"failed API calls",
		"token usage anomalies",
		"cost optimization opportunities",
		"model comparison results",
		"agent timeout errors",
		"successful deployments",
		"regression test failures",
		"prompt template changes",
	}

	prefix = strings.ToLower(prefix)
	var filtered []string
	for _, suggestion := range allSuggestions {
		if strings.Contains(strings.ToLower(suggestion), prefix) {
			filtered = append(filtered, suggestion)
		}
	}

	if len(filtered) > 5 {
		filtered = filtered[:5]
	}

	return filtered, nil
}

// generateMockResults generates mock search results based on keywords
func (s *SemanticSearchService) generateMockResults(keywords []string, limit int) []domain.SearchHit {
	mockHits := []domain.SearchHit{
		{TraceID: uuid.New().String(), TraceName: "auth-flow-validation", Score: 0.95, MatchType: "semantic", Highlight: "Authentication flow completed with token refresh", Timestamp: time.Now().Add(-1 * time.Hour), Cost: 0.023, Model: "gpt-4"},
		{TraceID: uuid.New().String(), TraceName: "api-error-handler", Score: 0.89, MatchType: "keyword", Highlight: "API error detected in payment processing", Timestamp: time.Now().Add(-2 * time.Hour), Cost: 0.015, Model: "gpt-4"},
		{TraceID: uuid.New().String(), TraceName: "code-review-analysis", Score: 0.87, MatchType: "semantic", Highlight: "Code review agent analyzed 15 files", Timestamp: time.Now().Add(-3 * time.Hour), Cost: 0.042, Model: "claude-3-opus"},
		{TraceID: uuid.New().String(), TraceName: "test-generation", Score: 0.82, MatchType: "keyword", Highlight: "Generated 24 unit tests for auth module", Timestamp: time.Now().Add(-4 * time.Hour), Cost: 0.031, Model: "gpt-4"},
		{TraceID: uuid.New().String(), TraceName: "doc-generation", Score: 0.78, MatchType: "semantic", Highlight: "Documentation generated for API endpoints", Timestamp: time.Now().Add(-5 * time.Hour), Cost: 0.019, Model: "claude-3-sonnet"},
	}

	if limit < len(mockHits) {
		mockHits = mockHits[:limit]
	}

	return mockHits
}
