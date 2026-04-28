package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	apperrors "github.com/agenttrace/agenttrace/api/internal/pkg/errors"
)

const maxOutcomePeriod = 366 * 24 * time.Hour

// OutcomeRepository defines the persistence boundary for outcome analytics.
type OutcomeRepository interface {
	GetSnapshot(
		ctx context.Context,
		projectID uuid.UUID,
		from, to time.Time,
	) (*domain.OutcomeSnapshot, error)
}

// OutcomeService owns project outcome analytics use cases.
type OutcomeService struct {
	repository OutcomeRepository
	clock      func() time.Time
}

// NewOutcomeService creates an outcome analytics service.
func NewOutcomeService(repository OutcomeRepository) *OutcomeService {
	return &OutcomeService{
		repository: repository,
		clock:      time.Now,
	}
}

// GetOverview calculates project-scoped outcome metrics for the requested period.
func (s *OutcomeService) GetOverview(
	ctx context.Context,
	projectID uuid.UUID,
	from, to time.Time,
) (*domain.OutcomeOverview, error) {
	if projectID == uuid.Nil {
		return nil, apperrors.Validation("project ID is required")
	}
	if !from.Before(to) {
		return nil, apperrors.Validation("from must be before to")
	}
	if to.Sub(from) > maxOutcomePeriod {
		return nil, apperrors.Validation("outcome analytics period cannot exceed 366 days")
	}

	snapshot, err := s.repository.GetSnapshot(ctx, projectID, from, to)
	if err != nil {
		return nil, fmt.Errorf("get outcome analytics: %w", err)
	}

	overview := &domain.OutcomeOverview{
		ProjectID: projectID,
		Period: domain.OutcomePeriod{
			From: from.UTC(),
			To:   to.UTC(),
		},
		Runs: domain.OutcomeRunMetrics{
			Total:       snapshot.Traces.TotalRuns,
			Successful:  snapshot.Traces.SuccessfulRuns,
			Failed:      snapshot.Traces.FailedRuns,
			InProgress:  snapshot.Traces.InProgressRuns,
			SuccessRate: ratioMetric(snapshot.Traces.SuccessfulRuns, snapshot.Traces.TotalRuns),
		},
		CI: domain.OutcomeCIMetrics{
			Total:      snapshot.CI.TotalRuns,
			Passed:     snapshot.CI.PassedRuns,
			Failed:     snapshot.CI.FailedRuns,
			Canceled:   snapshot.CI.CanceledRuns,
			InProgress: snapshot.CI.InProgressRuns,
			LinkedRuns: snapshot.CI.LinkedRuns,
			PassRate:   ratioMetric(snapshot.CI.PassedRuns, snapshot.CI.TotalRuns),
		},
		SourceControl: domain.OutcomeSCMMetrics{
			LinkedCommits:      snapshot.Git.LinkedCommits,
			LinkedTraces:       snapshot.Git.LinkedTraces,
			LinkedPullRequests: snapshot.CI.LinkedPRs,
			RegressionSignals:  snapshot.CI.LinkedFailures,
			RevertSignals:      snapshot.Git.RevertSignals,
		},
		Cost: domain.OutcomeCostMetrics{
			TotalCost: snapshot.Traces.TotalCost,
			CostPerSuccessfulOutcome: amountPerCountMetric(
				snapshot.Traces.TotalCost,
				snapshot.Traces.SuccessfulRuns,
			),
		},
		ByAgent:        mapOutcomeBreakdowns(snapshot.AgentBreakdown),
		ByModel:        mapOutcomeBreakdowns(snapshot.ModelBreakdown),
		RecentOutcomes: mapLinkedOutcomes(snapshot.RecentOutcomes),
		GeneratedAt:    s.clock().UTC(),
	}

	overview.Availability = outcomeAvailability(snapshot)
	return overview, nil
}

// GetDigest builds a reusable team digest from the canonical overview.
func (s *OutcomeService) GetDigest(
	ctx context.Context,
	projectID uuid.UUID,
	from, to time.Time,
) (*domain.OutcomeDigest, error) {
	overview, err := s.GetOverview(ctx, projectID, from, to)
	if err != nil {
		return nil, err
	}
	return BuildOutcomeDigest(overview), nil
}

func ratioMetric(numerator, denominator uint64) domain.OutcomeRateMetric {
	if denominator == 0 {
		return domain.OutcomeRateMetric{Value: nil, Available: false}
	}
	value := float64(numerator) / float64(denominator)
	return domain.OutcomeRateMetric{Value: &value, Available: true}
}

func amountPerCountMetric(amount float64, count uint64) domain.OutcomeRateMetric {
	if count == 0 {
		return domain.OutcomeRateMetric{Value: nil, Available: false}
	}
	value := amount / float64(count)
	return domain.OutcomeRateMetric{Value: &value, Available: true}
}

func mapOutcomeBreakdowns(items []domain.OutcomeBreakdownAggregate) []domain.OutcomeBreakdown {
	result := make([]domain.OutcomeBreakdown, 0, len(items))
	for _, item := range items {
		result = append(result, domain.OutcomeBreakdown{
			Name:                     item.Name,
			Runs:                     item.Runs,
			SuccessfulRuns:           item.SuccessfulRuns,
			SuccessRate:              ratioMetric(item.SuccessfulRuns, item.Runs),
			TotalCost:                item.TotalCost,
			CostPerSuccessfulOutcome: amountPerCountMetric(item.TotalCost, item.SuccessfulRuns),
		})
	}
	return result
}

func mapLinkedOutcomes(items []domain.LinkedOutcomeAggregate) []domain.LinkedOutcome {
	result := make([]domain.LinkedOutcome, 0, len(items))
	for _, item := range items {
		outcome := domain.LinkedOutcome{
			CommitSHA:     item.CommitSHA,
			CommitMessage: item.CommitMessage,
			Branch:        item.Branch,
			CommittedAt:   item.CommittedAt,
			TraceCount:    item.TraceCount,
		}
		if item.CIStatus != "" {
			status := item.CIStatus
			outcome.CIStatus = &status
		}
		if item.CIConclusion != "" {
			conclusion := item.CIConclusion
			outcome.CIConclusion = &conclusion
		}
		if item.CIProviderURL != "" {
			providerURL := item.CIProviderURL
			outcome.CIProviderURL = &providerURL
		}
		if item.PRNumber > 0 {
			prNumber := item.PRNumber
			outcome.PRNumber = &prNumber
		}
		if item.PRTitle != "" {
			prTitle := item.PRTitle
			outcome.PRTitle = &prTitle
		}
		result = append(result, outcome)
	}
	return result
}

func outcomeAvailability(snapshot *domain.OutcomeSnapshot) domain.OutcomeAvailability {
	availability := domain.OutcomeAvailability{
		TraceData:        snapshot.Traces.TotalRuns > 0,
		GitData:          snapshot.Git.LinkedCommits > 0,
		CIData:           snapshot.CI.TotalRuns > 0,
		PullRequestData:  snapshot.CI.LinkedPRs > 0,
		AgentAttribution: len(snapshot.AgentBreakdown) > 0,
		ModelAttribution: len(snapshot.ModelBreakdown) > 0,
		Unavailable:      []string{},
	}

	if !availability.TraceData {
		availability.Unavailable = append(availability.Unavailable, "trace outcomes")
	}
	if !availability.GitData {
		availability.Unavailable = append(availability.Unavailable, "linked commits")
	}
	if !availability.CIData {
		availability.Unavailable = append(availability.Unavailable, "CI outcomes")
	}
	if !availability.PullRequestData {
		availability.Unavailable = append(availability.Unavailable, "pull request outcomes")
	}
	if !availability.AgentAttribution {
		availability.Unavailable = append(availability.Unavailable, "agent attribution")
	}
	if !availability.ModelAttribution {
		availability.Unavailable = append(availability.Unavailable, "model attribution")
	}
	return availability
}
