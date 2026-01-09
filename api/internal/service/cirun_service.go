package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

// CIRunRepository defines CI run repository operations
type CIRunRepository interface {
	Create(ctx context.Context, ciRun *domain.CIRun) error
	GetByID(ctx context.Context, projectID, ciRunID uuid.UUID) (*domain.CIRun, error)
	GetByProviderRunID(ctx context.Context, projectID uuid.UUID, providerRunID string) (*domain.CIRun, error)
	List(ctx context.Context, filter *domain.CIRunFilter, limit, offset int) (*domain.CIRunList, error)
	Update(ctx context.Context, ciRun *domain.CIRun) error
	GetStats(ctx context.Context, projectID uuid.UUID) (*domain.CIRunStats, error)
}

// CIRunService handles CI/CD run tracking
type CIRunService struct {
	ciRunRepo CIRunRepository
}

// NewCIRunService creates a new CI run service
func NewCIRunService(ciRunRepo CIRunRepository) *CIRunService {
	return &CIRunService{
		ciRunRepo: ciRunRepo,
	}
}

// Create creates a new CI run
func (s *CIRunService) Create(ctx context.Context, projectID uuid.UUID, input *domain.CIRunInput) (*domain.CIRun, error) {
	now := time.Now()

	startedAt := now
	if input.StartedAt != nil {
		startedAt = *input.StartedAt
	}

	status := domain.CIRunStatusPending
	if input.Status != nil {
		status = *input.Status
	}

	var pipelineName, jobName, workflowName string
	var gitCommitSha, gitBranch, gitTag, gitRepoURL, gitRef string
	var prTitle, prSourceBranch, prTargetBranch string
	var conclusion, errorMessage string
	var runnerName, runnerOS, runnerArch string
	var triggeredBy, triggerEvent string

	if input.PipelineName != nil {
		pipelineName = *input.PipelineName
	}
	if input.JobName != nil {
		jobName = *input.JobName
	}
	if input.WorkflowName != nil {
		workflowName = *input.WorkflowName
	}
	if input.GitCommitSha != nil {
		gitCommitSha = *input.GitCommitSha
	}
	if input.GitBranch != nil {
		gitBranch = *input.GitBranch
	}
	if input.GitTag != nil {
		gitTag = *input.GitTag
	}
	if input.GitRepoURL != nil {
		gitRepoURL = *input.GitRepoURL
	}
	if input.GitRef != nil {
		gitRef = *input.GitRef
	}
	if input.PRTitle != nil {
		prTitle = *input.PRTitle
	}
	if input.PRSourceBranch != nil {
		prSourceBranch = *input.PRSourceBranch
	}
	if input.PRTargetBranch != nil {
		prTargetBranch = *input.PRTargetBranch
	}
	if input.Conclusion != nil {
		conclusion = *input.Conclusion
	}
	if input.ErrorMessage != nil {
		errorMessage = *input.ErrorMessage
	}
	if input.RunnerName != nil {
		runnerName = *input.RunnerName
	}
	if input.RunnerOS != nil {
		runnerOS = *input.RunnerOS
	}
	if input.RunnerArch != nil {
		runnerArch = *input.RunnerArch
	}
	if input.TriggeredBy != nil {
		triggeredBy = *input.TriggeredBy
	}
	if input.TriggerEvent != nil {
		triggerEvent = *input.TriggerEvent
	}

	var prNumber uint32
	if input.PRNumber != nil {
		prNumber = *input.PRNumber
	}

	var providerRunURL string
	if input.ProviderRunURL != nil {
		providerRunURL = *input.ProviderRunURL
	}

	ciRun := &domain.CIRun{
		ID:              uuid.New(),
		ProjectID:       projectID,
		Provider:        input.Provider,
		ProviderRunID:   input.ProviderRunID,
		ProviderRunURL:  providerRunURL,
		PipelineName:    pipelineName,
		JobName:         jobName,
		WorkflowName:    workflowName,
		GitCommitSha:    gitCommitSha,
		GitBranch:       gitBranch,
		GitTag:          gitTag,
		GitRepoURL:      gitRepoURL,
		GitRef:          gitRef,
		PRNumber:        prNumber,
		PRTitle:         prTitle,
		PRSourceBranch:  prSourceBranch,
		PRTargetBranch:  prTargetBranch,
		StartedAt:       startedAt,
		Status:          status,
		Conclusion:      conclusion,
		ErrorMessage:    errorMessage,
		TraceIDs:        []string{},
		RunnerName:      runnerName,
		RunnerOS:        runnerOS,
		RunnerArch:      runnerArch,
		TriggeredBy:     triggeredBy,
		TriggerEvent:    triggerEvent,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	if err := s.ciRunRepo.Create(ctx, ciRun); err != nil {
		return nil, fmt.Errorf("failed to create CI run: %w", err)
	}

	return ciRun, nil
}

// Get retrieves a CI run by ID
func (s *CIRunService) Get(ctx context.Context, projectID, ciRunID uuid.UUID) (*domain.CIRun, error) {
	return s.ciRunRepo.GetByID(ctx, projectID, ciRunID)
}

// GetByProviderRunID retrieves a CI run by provider's run ID
func (s *CIRunService) GetByProviderRunID(ctx context.Context, projectID uuid.UUID, providerRunID string) (*domain.CIRun, error) {
	return s.ciRunRepo.GetByProviderRunID(ctx, projectID, providerRunID)
}

// List retrieves CI runs with filtering
func (s *CIRunService) List(ctx context.Context, filter *domain.CIRunFilter, limit, offset int) (*domain.CIRunList, error) {
	return s.ciRunRepo.List(ctx, filter, limit, offset)
}

// Update updates a CI run
func (s *CIRunService) Update(ctx context.Context, projectID, ciRunID uuid.UUID, input *domain.CIRunUpdateInput) (*domain.CIRun, error) {
	ciRun, err := s.ciRunRepo.GetByID(ctx, projectID, ciRunID)
	if err != nil {
		return nil, fmt.Errorf("failed to get CI run: %w", err)
	}

	if input.Status != nil {
		ciRun.Status = *input.Status
	}
	if input.Conclusion != nil {
		ciRun.Conclusion = *input.Conclusion
	}
	if input.ErrorMessage != nil {
		ciRun.ErrorMessage = *input.ErrorMessage
	}
	if input.CompletedAt != nil {
		ciRun.CompletedAt = input.CompletedAt
		if ciRun.StartedAt.Before(*input.CompletedAt) {
			ciRun.DurationMs = uint32(input.CompletedAt.Sub(ciRun.StartedAt).Milliseconds())
		}
	}
	if input.TraceIDs != nil {
		ciRun.TraceIDs = input.TraceIDs
		ciRun.TraceCount = uint32(len(input.TraceIDs))
	}
	if input.TotalCost != nil {
		ciRun.TotalCost = *input.TotalCost
	}
	if input.TotalTokens != nil {
		ciRun.TotalTokens = *input.TotalTokens
	}
	if input.TotalObservations != nil {
		ciRun.TotalObservations = *input.TotalObservations
	}

	if err := s.ciRunRepo.Update(ctx, ciRun); err != nil {
		return nil, fmt.Errorf("failed to update CI run: %w", err)
	}

	return ciRun, nil
}

// AddTrace adds a trace ID to a CI run
func (s *CIRunService) AddTrace(ctx context.Context, projectID, ciRunID uuid.UUID, traceID string) error {
	ciRun, err := s.ciRunRepo.GetByID(ctx, projectID, ciRunID)
	if err != nil {
		return fmt.Errorf("failed to get CI run: %w", err)
	}

	// Check if trace already exists
	for _, id := range ciRun.TraceIDs {
		if id == traceID {
			return nil // Already exists
		}
	}

	ciRun.TraceIDs = append(ciRun.TraceIDs, traceID)
	ciRun.TraceCount = uint32(len(ciRun.TraceIDs))

	if err := s.ciRunRepo.Update(ctx, ciRun); err != nil {
		return fmt.Errorf("failed to update CI run: %w", err)
	}

	return nil
}

// Complete marks a CI run as completed
func (s *CIRunService) Complete(ctx context.Context, projectID, ciRunID uuid.UUID, status domain.CIRunStatus, conclusion string) (*domain.CIRun, error) {
	ciRun, err := s.ciRunRepo.GetByID(ctx, projectID, ciRunID)
	if err != nil {
		return nil, fmt.Errorf("failed to get CI run: %w", err)
	}

	now := time.Now()
	ciRun.Status = status
	ciRun.Conclusion = conclusion
	ciRun.CompletedAt = &now
	ciRun.DurationMs = uint32(now.Sub(ciRun.StartedAt).Milliseconds())

	if err := s.ciRunRepo.Update(ctx, ciRun); err != nil {
		return nil, fmt.Errorf("failed to update CI run: %w", err)
	}

	return ciRun, nil
}

// GetStats retrieves CI run statistics
func (s *CIRunService) GetStats(ctx context.Context, projectID uuid.UUID) (*domain.CIRunStats, error) {
	return s.ciRunRepo.GetStats(ctx, projectID)
}
