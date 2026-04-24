package service

import (
	"net/http"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/config"
	"github.com/agenttrace/agenttrace/api/internal/domain"
	apperrors "github.com/agenttrace/agenttrace/api/internal/pkg/errors"
)

type countingRoundTripper struct {
	calls atomic.Int64
}

func (t *countingRoundTripper) RoundTrip(*http.Request) (*http.Response, error) {
	t.calls.Add(1)
	return nil, nil
}

func newTestNLQueryService() *NLQueryService {
	logger, _ := zap.NewDevelopment()
	cfg := &config.Config{}
	return NewNLQueryService(cfg, nil, logger)
}

func TestNLQueryService_ConvertFilter(t *testing.T) {
	svc := newTestNLQueryService()
	projectID := uuid.New()

	t.Run("valid ParsedFilter produces correct TraceFilter", func(t *testing.T) {
		name := "test-trace"
		userID := "user-123"
		hasError := true
		minCost := 0.5
		maxCost := 10.0
		search := "error in production"

		parsed := ParsedFilter{
			Name:     &name,
			UserID:   &userID,
			HasError: &hasError,
			MinCost:  &minCost,
			MaxCost:  &maxCost,
			Search:   &search,
			Tags:     []string{"prod", "critical"},
		}

		filter := svc.convertFilter(parsed, projectID)

		assert.Equal(t, projectID, filter.ProjectID)
		assert.Equal(t, &name, filter.Name)
		assert.Equal(t, &userID, filter.UserID)
		assert.Equal(t, &hasError, filter.HasError)
		assert.Equal(t, &minCost, filter.MinCost)
		assert.Equal(t, &maxCost, filter.MaxCost)
		assert.Equal(t, &search, filter.Search)
		assert.Equal(t, []string{"prod", "critical"}, filter.Tags)
	})

	t.Run("empty ParsedFilter produces minimal TraceFilter", func(t *testing.T) {
		parsed := ParsedFilter{}
		filter := svc.convertFilter(parsed, projectID)

		assert.Equal(t, projectID, filter.ProjectID)
		assert.Nil(t, filter.Name)
		assert.Nil(t, filter.UserID)
		assert.Nil(t, filter.HasError)
		assert.Nil(t, filter.FromTime)
		assert.Nil(t, filter.ToTime)
	})
}

func TestNLQueryService_ConvertFilter_TimeParsing(t *testing.T) {
	svc := newTestNLQueryService()
	projectID := uuid.New()

	t.Run("RFC3339 time parsed correctly", func(t *testing.T) {
		fromTime := "2024-01-01T00:00:00Z"
		toTime := "2024-01-02T00:00:00Z"

		parsed := ParsedFilter{
			FromTime: &fromTime,
			ToTime:   &toTime,
		}

		filter := svc.convertFilter(parsed, projectID)

		require.NotNil(t, filter.FromTime)
		require.NotNil(t, filter.ToTime)

		expectedFrom, _ := time.Parse(time.RFC3339, fromTime)
		expectedTo, _ := time.Parse(time.RFC3339, toTime)
		assert.Equal(t, expectedFrom, *filter.FromTime)
		assert.Equal(t, expectedTo, *filter.ToTime)
	})

	t.Run("invalid time format uses graceful fallback - nil", func(t *testing.T) {
		invalidTime := "last 24 hours"

		parsed := ParsedFilter{
			FromTime: &invalidTime,
		}

		filter := svc.convertFilter(parsed, projectID)

		// Invalid time should result in nil (graceful fallback)
		assert.Nil(t, filter.FromTime)
	})
}

func TestNLQueryService_ConvertFilter_Level(t *testing.T) {
	svc := newTestNLQueryService()
	projectID := uuid.New()

	level := "ERROR"
	parsed := ParsedFilter{
		Level: &level,
	}

	filter := svc.convertFilter(parsed, projectID)
	require.NotNil(t, filter.Level)
	assert.Equal(t, domain.Level("ERROR"), *filter.Level)
}

func TestNLQueryService_ConvertFilter_GitFields(t *testing.T) {
	svc := newTestNLQueryService()
	projectID := uuid.New()

	branch := "main"
	commit := "abc123"

	parsed := ParsedFilter{
		GitBranch: &branch,
		GitCommit: &commit,
	}

	filter := svc.convertFilter(parsed, projectID)
	assert.Equal(t, &branch, filter.GitBranch)
	assert.Equal(t, &commit, filter.GitCommitSha)
}

func TestNLQueryService_ParseQuery_NoAPIKey(t *testing.T) {
	svc := newTestNLQueryService()
	ctx := t.Context()

	// parseQuery should fail when no API key is configured
	_, err := svc.parseQuery(ctx, "show me errors")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "API key not configured")
}

func TestNLQueryService_BlocksExternalModelInNoEgressMode(t *testing.T) {
	cfg := &config.Config{}
	cfg.Eval.APIKey = "external-key"
	transport := &countingRoundTripper{}
	svc := NewNLQueryService(
		cfg,
		nil,
		zap.NewNop(),
		NewEgressPolicy(true, true),
	)
	svc.httpClient.Transport = transport

	_, err := svc.callOpenAI(t.Context(), "system", "user")

	require.Error(t, err)
	assert.True(t, apperrors.IsUnprocessable(err))
	assert.Zero(t, transport.calls.Load())
}

func TestNLQueryService_QueryTraces_FallbackOnParseFailure(t *testing.T) {
	// When parseQuery fails, QueryTraces should fall back to text search
	// We can't test the full flow without a QueryService, but we can
	// verify the fallback logic by checking that parseQuery returns error
	// when no API key is set
	svc := newTestNLQueryService()
	ctx := t.Context()

	parsed, err := svc.parseQuery(ctx, "find errors")
	assert.Error(t, err)
	assert.Nil(t, parsed)
}

func TestNLQueryService_GetQueryExamples(t *testing.T) {
	svc := newTestNLQueryService()
	examples := svc.GetQueryExamples()
	assert.NotEmpty(t, examples)
	assert.True(t, len(examples) >= 5, "should have several example queries")
}
