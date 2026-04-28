package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

type outcomeRepositoryStub struct {
	projectID uuid.UUID
	snapshot  *domain.OutcomeSnapshot
	err       error
}

func (r *outcomeRepositoryStub) GetSnapshot(
	_ context.Context,
	projectID uuid.UUID,
	_, _ time.Time,
) (*domain.OutcomeSnapshot, error) {
	r.projectID = projectID
	return r.snapshot, r.err
}

func TestOutcomeServiceCalculatesRealRatios(t *testing.T) {
	projectID := uuid.New()
	repository := &outcomeRepositoryStub{
		snapshot: &domain.OutcomeSnapshot{
			Traces: domain.OutcomeTraceAggregate{
				TotalRuns:      10,
				SuccessfulRuns: 8,
				FailedRuns:     2,
				TotalCost:      4,
			},
			Git: domain.OutcomeGitAggregate{
				LinkedCommits: 5,
				LinkedTraces:  8,
				RevertSignals: 1,
			},
			CI: domain.OutcomeCIAggregate{
				TotalRuns:      4,
				PassedRuns:     3,
				FailedRuns:     1,
				LinkedRuns:     3,
				LinkedFailures: 1,
				LinkedPRs:      2,
			},
			AgentBreakdown: []domain.OutcomeBreakdownAggregate{
				{Name: "copilot", Runs: 5, SuccessfulRuns: 4, TotalCost: 2},
			},
			ModelBreakdown: []domain.OutcomeBreakdownAggregate{
				{Name: "gpt-4.1", Runs: 4, SuccessfulRuns: 3, TotalCost: 1.5},
			},
		},
	}
	service := NewOutcomeService(repository)
	service.clock = func() time.Time { return time.Date(2026, 7, 25, 20, 0, 0, 0, time.UTC) }

	overview, err := service.GetOverview(
		context.Background(),
		projectID,
		time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 7, 26, 0, 0, 0, 0, time.UTC),
	)

	require.NoError(t, err)
	assert.Equal(t, projectID, repository.projectID)
	require.NotNil(t, overview.Runs.SuccessRate.Value)
	assert.InDelta(t, 0.8, *overview.Runs.SuccessRate.Value, 0.0001)
	require.NotNil(t, overview.CI.PassRate.Value)
	assert.InDelta(t, 0.75, *overview.CI.PassRate.Value, 0.0001)
	require.NotNil(t, overview.Cost.CostPerSuccessfulOutcome.Value)
	assert.InDelta(t, 0.5, *overview.Cost.CostPerSuccessfulOutcome.Value, 0.0001)
	assert.Equal(t, uint64(1), overview.SourceControl.RegressionSignals)
	assert.True(t, overview.Availability.AgentAttribution)
	assert.True(t, overview.Availability.ModelAttribution)
}

func TestOutcomeServiceRepresentsUnavailableData(t *testing.T) {
	repository := &outcomeRepositoryStub{snapshot: &domain.OutcomeSnapshot{}}
	service := NewOutcomeService(repository)

	overview, err := service.GetOverview(
		context.Background(),
		uuid.New(),
		time.Now().Add(-24*time.Hour),
		time.Now(),
	)

	require.NoError(t, err)
	assert.False(t, overview.Runs.SuccessRate.Available)
	assert.Nil(t, overview.Runs.SuccessRate.Value)
	assert.False(t, overview.Cost.CostPerSuccessfulOutcome.Available)
	assert.Contains(t, overview.Availability.Unavailable, "CI outcomes")
	assert.Contains(t, overview.Availability.Unavailable, "model attribution")
}

func TestOutcomeServiceRejectsInvalidPeriods(t *testing.T) {
	service := NewOutcomeService(&outcomeRepositoryStub{})
	now := time.Now()

	_, err := service.GetOverview(context.Background(), uuid.New(), now, now)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "from must be before to")

	_, err = service.GetOverview(
		context.Background(),
		uuid.New(),
		now.Add(-400*24*time.Hour),
		now,
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot exceed")
}

func TestOutcomeServiceWrapsRepositoryErrors(t *testing.T) {
	service := NewOutcomeService(&outcomeRepositoryStub{err: errors.New("database unavailable")})

	_, err := service.GetOverview(
		context.Background(),
		uuid.New(),
		time.Now().Add(-time.Hour),
		time.Now(),
	)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "get outcome analytics")
}

func TestOutcomeDigestDoesNotFabricateUnavailableMetrics(t *testing.T) {
	overview := &domain.OutcomeOverview{
		ProjectID: uuid.New(),
		Period: domain.OutcomePeriod{
			From: time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
			To:   time.Date(2026, 7, 8, 0, 0, 0, 0, time.UTC),
		},
		Availability: domain.OutcomeAvailability{
			Unavailable: []string{"CI outcomes"},
		},
	}

	digest := BuildOutcomeDigest(overview)
	rendered := RenderOutcomeDigestMarkdown(digest)

	assert.Contains(t, digest.Summary, "No completed agent outcomes")
	assert.Contains(t, rendered, "CI outcome data is unavailable")
	assert.NotContains(t, rendered, "0.0%")
}
