package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/pkg/id"
)

// TraceRepository defines trace repository operations
type TraceRepository interface {
	Create(ctx context.Context, trace *domain.Trace) error
	CreateBatch(ctx context.Context, traces []*domain.Trace) error
	GetByID(ctx context.Context, projectID uuid.UUID, traceID string) (*domain.Trace, error)
	Update(ctx context.Context, trace *domain.Trace) error
	UpdateCosts(ctx context.Context, projectID uuid.UUID, traceID string, inputCost, outputCost, totalCost float64) error
	List(ctx context.Context, filter *domain.TraceFilter, limit, offset int) (*domain.TraceList, error)
	SetBookmark(ctx context.Context, projectID uuid.UUID, traceID string, bookmarked bool) error
	GetBySessionID(ctx context.Context, projectID uuid.UUID, sessionID string) ([]domain.Trace, error)
}

// ObservationRepository defines observation repository operations
type ObservationRepository interface {
	Create(ctx context.Context, obs *domain.Observation) error
	CreateBatch(ctx context.Context, observations []*domain.Observation) error
	GetByID(ctx context.Context, projectID uuid.UUID, observationID string) (*domain.Observation, error)
	GetByTraceID(ctx context.Context, projectID uuid.UUID, traceID string) ([]domain.Observation, error)
	Update(ctx context.Context, obs *domain.Observation) error
	UpdateCosts(ctx context.Context, projectID uuid.UUID, observationID string, inputCost, outputCost, totalCost float64) error
	List(ctx context.Context, filter *domain.ObservationFilter, limit, offset int) ([]domain.Observation, int64, error)
	GetGenerationsWithoutCost(ctx context.Context, projectID uuid.UUID, limit int) ([]domain.Observation, error)
	GetTree(ctx context.Context, projectID uuid.UUID, traceID string) (*domain.ObservationTree, error)
}

// SessionRepository defines session repository operations
type SessionRepository interface {
	Upsert(ctx context.Context, session *domain.Session) error
	GetByID(ctx context.Context, projectID uuid.UUID, sessionID string) (*domain.Session, error)
	List(ctx context.Context, filter *domain.SessionFilter, limit, offset int) (*domain.SessionList, error)
}

// IngestionService handles trace and observation ingestion
type IngestionService struct {
	traceRepo       TraceRepository
	observationRepo ObservationRepository
	costService     *CostService
	evalService     *EvalService
}

// NewIngestionService creates a new ingestion service
func NewIngestionService(
	traceRepo TraceRepository,
	observationRepo ObservationRepository,
	costService *CostService,
	evalService *EvalService,
) *IngestionService {
	return &IngestionService{
		traceRepo:       traceRepo,
		observationRepo: observationRepo,
		costService:     costService,
		evalService:     evalService,
	}
}

// IngestTrace ingests a single trace
func (s *IngestionService) IngestTrace(ctx context.Context, projectID uuid.UUID, input *domain.TraceInput) (*domain.Trace, error) {
	now := time.Now()

	// Generate trace ID if not provided
	traceID := input.ID
	if traceID == "" {
		traceID = id.NewTraceID()
	}

	// Marshal metadata
	var metadata string
	if input.Metadata != nil {
		metadataBytes, err := json.Marshal(input.Metadata)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal metadata: %w", err)
		}
		metadata = string(metadataBytes)
	}

	// Set timestamps
	startTime := now
	if input.StartTime != nil {
		startTime = *input.StartTime
	} else if input.Timestamp != nil {
		startTime = *input.Timestamp
	}

	trace := &domain.Trace{
		ID:          traceID,
		ProjectID:   projectID,
		Name:        input.Name,
		UserID:      input.UserID,
		SessionID:   input.SessionID,
		Metadata:    metadata,
		Tags:        input.Tags,
		Release:     input.Release,
		Version:     input.Version,
		Public:      input.Public,
		StartTime:   startTime,
		EndTime:     input.EndTime,
		GitCommitSha: input.GitCommitSha,
		GitBranch:   input.GitBranch,
		GitRepoURL:  input.GitRepoURL,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.traceRepo.Create(ctx, trace); err != nil {
		return nil, fmt.Errorf("failed to create trace: %w", err)
	}

	// Trigger evaluators asynchronously
	if s.evalService != nil {
		go func() {
			_ = s.evalService.TriggerForTrace(context.Background(), projectID, trace)
		}()
	}

	return trace, nil
}

// IngestObservation ingests a single observation (span or generation)
func (s *IngestionService) IngestObservation(ctx context.Context, projectID uuid.UUID, input *domain.ObservationInput) (*domain.Observation, error) {
	now := time.Now()

	// Generate observation ID if not provided
	var obsID string
	if input.ID != nil && *input.ID != "" {
		obsID = *input.ID
	} else {
		obsID = id.NewSpanID()
	}

	// Get trace ID
	var traceID string
	if input.TraceID != nil {
		traceID = *input.TraceID
	}

	// Marshal metadata
	var metadata string
	if input.Metadata != nil {
		metadataBytes, err := json.Marshal(input.Metadata)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal metadata: %w", err)
		}
		metadata = string(metadataBytes)
	}

	// Marshal input/output
	var inputStr, outputStr string
	if input.Input != nil {
		inputBytes, err := json.Marshal(input.Input)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal input: %w", err)
		}
		inputStr = string(inputBytes)
	}
	if input.Output != nil {
		outputBytes, err := json.Marshal(input.Output)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal output: %w", err)
		}
		outputStr = string(outputBytes)
	}

	// Set timestamps
	startTime := now
	if input.StartTime != nil {
		startTime = *input.StartTime
	}

	// Handle optional fields with defaults
	var obsType domain.ObservationType
	if input.Type != nil {
		obsType = *input.Type
	} else {
		obsType = domain.ObservationTypeSpan
	}

	var name string
	if input.Name != nil {
		name = *input.Name
	}

	var level domain.Level
	if input.Level != nil {
		level = *input.Level
	} else {
		level = domain.LevelDefault
	}

	var statusMessage string
	if input.StatusMessage != nil {
		statusMessage = *input.StatusMessage
	}

	var version string
	if input.Version != nil {
		version = *input.Version
	}

	obs := &domain.Observation{
		ID:                  obsID,
		TraceID:             traceID,
		ProjectID:           projectID,
		ParentObservationID: input.ParentObservationID,
		Type:                obsType,
		Name:                name,
		StartTime:           startTime,
		EndTime:             input.EndTime,
		Metadata:            metadata,
		Level:               level,
		StatusMessage:       statusMessage,
		Version:             version,
		Input:               inputStr,
		Output:              outputStr,
		CreatedAt:           now,
		UpdatedAt:           now,
	}

	if err := s.observationRepo.Create(ctx, obs); err != nil {
		return nil, fmt.Errorf("failed to create observation: %w", err)
	}

	return obs, nil
}

// IngestGeneration ingests a generation (LLM call) observation
func (s *IngestionService) IngestGeneration(ctx context.Context, projectID uuid.UUID, input *domain.GenerationInput) (*domain.Observation, error) {
	now := time.Now()

	// Generate observation ID if not provided
	var obsID string
	if input.ID != nil && *input.ID != "" {
		obsID = *input.ID
	} else {
		obsID = id.NewSpanID()
	}

	// Get trace ID
	var traceID string
	if input.TraceID != nil {
		traceID = *input.TraceID
	}

	// Marshal metadata
	var metadata string
	if input.Metadata != nil {
		metadataBytes, err := json.Marshal(input.Metadata)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal metadata: %w", err)
		}
		metadata = string(metadataBytes)
	}

	// Marshal input/output
	var inputStr, outputStr string
	if input.Input != nil {
		inputBytes, err := json.Marshal(input.Input)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal input: %w", err)
		}
		inputStr = string(inputBytes)
	}
	if input.Output != nil {
		outputBytes, err := json.Marshal(input.Output)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal output: %w", err)
		}
		outputStr = string(outputBytes)
	}

	// Marshal model parameters
	var modelParams string
	if input.ModelParameters != nil {
		paramsBytes, err := json.Marshal(input.ModelParameters)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal model parameters: %w", err)
		}
		modelParams = string(paramsBytes)
	}

	// Set timestamps
	startTime := now
	if input.StartTime != nil {
		startTime = *input.StartTime
	}

	// Calculate duration
	var durationMs float64
	if input.EndTime != nil {
		durationMs = float64(input.EndTime.Sub(startTime).Milliseconds())
	}

	// Build usage details
	var usageDetails domain.UsageDetails
	if input.Usage != nil {
		normalized := input.Usage.Normalize()
		usageDetails = normalized
	}

	// Handle optional fields with defaults
	var name string
	if input.Name != nil {
		name = *input.Name
	}

	var level domain.Level
	if input.Level != nil {
		level = *input.Level
	} else {
		level = domain.LevelDefault
	}

	var statusMessage string
	if input.StatusMessage != nil {
		statusMessage = *input.StatusMessage
	}

	var version string
	if input.Version != nil {
		version = *input.Version
	}

	obs := &domain.Observation{
		ID:                  obsID,
		TraceID:             traceID,
		ProjectID:           projectID,
		ParentObservationID: input.ParentObservationID,
		Type:                domain.ObservationTypeGeneration,
		Name:                name,
		StartTime:           startTime,
		EndTime:             input.EndTime,
		Metadata:            metadata,
		Level:               level,
		StatusMessage:       statusMessage,
		Version:             version,
		Input:               inputStr,
		Output:              outputStr,
		Model:               input.Model,
		ModelParameters:     modelParams,
		UsageDetails:        usageDetails,
		PromptName:          input.PromptName,
		DurationMs:          durationMs,
		CreatedAt:           now,
		UpdatedAt:           now,
	}

	// Calculate costs if we have usage data and a model
	if usageDetails.TotalTokens > 0 && input.Model != "" && s.costService != nil {
		cost, err := s.costService.CalculateCost(ctx, projectID, input.Model, int64(usageDetails.InputTokens), int64(usageDetails.OutputTokens))
		if err == nil && cost != nil {
			obs.CostDetails = *cost
		}
	}

	if err := s.observationRepo.Create(ctx, obs); err != nil {
		return nil, fmt.Errorf("failed to create generation: %w", err)
	}

	// Update trace with accumulated costs
	if obs.CostDetails.TotalCost > 0 {
		go func() {
			_ = s.updateTraceCosts(context.Background(), projectID, traceID)
		}()
	}

	// Trigger evaluators asynchronously
	if s.evalService != nil {
		go func() {
			_ = s.evalService.TriggerForObservation(context.Background(), projectID, obs)
		}()
	}

	return obs, nil
}

// IngestBatch ingests multiple traces and observations in a batch
func (s *IngestionService) IngestBatch(ctx context.Context, projectID uuid.UUID, batch *domain.IngestionBatch) error {
	now := time.Now()

	// Process traces
	traces := make([]*domain.Trace, 0, len(batch.Traces))
	for _, input := range batch.Traces {
		traceID := input.ID
		if traceID == "" {
			traceID = id.NewTraceID()
		}

		var metadata string
		if input.Metadata != nil {
			metadataBytes, _ := json.Marshal(input.Metadata)
			metadata = string(metadataBytes)
		}

		startTime := now
		if input.StartTime != nil {
			startTime = *input.StartTime
		} else if input.Timestamp != nil {
			startTime = *input.Timestamp
		}

		traces = append(traces, &domain.Trace{
			ID:           traceID,
			ProjectID:    projectID,
			Name:         input.Name,
			UserID:       input.UserID,
			SessionID:    input.SessionID,
			Metadata:     metadata,
			Tags:         input.Tags,
			Release:      input.Release,
			Version:      input.Version,
			Public:       input.Public,
			StartTime:    startTime,
			EndTime:      input.EndTime,
			GitCommitSha: input.GitCommitSha,
			GitBranch:    input.GitBranch,
			GitRepoURL:   input.GitRepoURL,
			CreatedAt:    now,
			UpdatedAt:    now,
		})
	}

	if len(traces) > 0 {
		if err := s.traceRepo.CreateBatch(ctx, traces); err != nil {
			return fmt.Errorf("failed to batch create traces: %w", err)
		}
	}

	// Process observations
	observations := make([]*domain.Observation, 0, len(batch.Observations)+len(batch.Generations))

	for _, input := range batch.Observations {
		var obsID string
		if input.ID != nil && *input.ID != "" {
			obsID = *input.ID
		} else {
			obsID = id.NewSpanID()
		}

		var traceID string
		if input.TraceID != nil {
			traceID = *input.TraceID
		}

		var metadata, inputStr, outputStr string
		if input.Metadata != nil {
			metadataBytes, _ := json.Marshal(input.Metadata)
			metadata = string(metadataBytes)
		}
		if input.Input != nil {
			inputBytes, _ := json.Marshal(input.Input)
			inputStr = string(inputBytes)
		}
		if input.Output != nil {
			outputBytes, _ := json.Marshal(input.Output)
			outputStr = string(outputBytes)
		}

		startTime := now
		if input.StartTime != nil {
			startTime = *input.StartTime
		}

		// Handle optional fields with defaults
		var obsType domain.ObservationType
		if input.Type != nil {
			obsType = *input.Type
		} else {
			obsType = domain.ObservationTypeSpan
		}

		var name string
		if input.Name != nil {
			name = *input.Name
		}

		var level domain.Level
		if input.Level != nil {
			level = *input.Level
		} else {
			level = domain.LevelDefault
		}

		var statusMessage string
		if input.StatusMessage != nil {
			statusMessage = *input.StatusMessage
		}

		var version string
		if input.Version != nil {
			version = *input.Version
		}

		observations = append(observations, &domain.Observation{
			ID:                  obsID,
			TraceID:             traceID,
			ProjectID:           projectID,
			ParentObservationID: input.ParentObservationID,
			Type:                obsType,
			Name:                name,
			StartTime:           startTime,
			EndTime:             input.EndTime,
			Metadata:            metadata,
			Level:               level,
			StatusMessage:       statusMessage,
			Version:             version,
			Input:               inputStr,
			Output:              outputStr,
			CreatedAt:           now,
			UpdatedAt:           now,
		})
	}

	for _, input := range batch.Generations {
		var obsID string
		if input.ID != nil && *input.ID != "" {
			obsID = *input.ID
		} else {
			obsID = id.NewSpanID()
		}

		var traceID string
		if input.TraceID != nil {
			traceID = *input.TraceID
		}

		var metadata, inputStr, outputStr, modelParams string
		if input.Metadata != nil {
			metadataBytes, _ := json.Marshal(input.Metadata)
			metadata = string(metadataBytes)
		}
		if input.Input != nil {
			inputBytes, _ := json.Marshal(input.Input)
			inputStr = string(inputBytes)
		}
		if input.Output != nil {
			outputBytes, _ := json.Marshal(input.Output)
			outputStr = string(outputBytes)
		}
		if input.ModelParameters != nil {
			paramsBytes, _ := json.Marshal(input.ModelParameters)
			modelParams = string(paramsBytes)
		}

		startTime := now
		if input.StartTime != nil {
			startTime = *input.StartTime
		}

		var durationMs float64
		if input.EndTime != nil {
			durationMs = float64(input.EndTime.Sub(startTime).Milliseconds())
		}

		var usageDetails domain.UsageDetails
		if input.Usage != nil {
			usageDetails = input.Usage.Normalize()
		}

		// Handle optional fields with defaults
		var name string
		if input.Name != nil {
			name = *input.Name
		}

		var level domain.Level
		if input.Level != nil {
			level = *input.Level
		} else {
			level = domain.LevelDefault
		}

		var statusMessage string
		if input.StatusMessage != nil {
			statusMessage = *input.StatusMessage
		}

		var version string
		if input.Version != nil {
			version = *input.Version
		}

		obs := &domain.Observation{
			ID:                  obsID,
			TraceID:             traceID,
			ProjectID:           projectID,
			ParentObservationID: input.ParentObservationID,
			Type:                domain.ObservationTypeGeneration,
			Name:                name,
			StartTime:           startTime,
			EndTime:             input.EndTime,
			Metadata:            metadata,
			Level:               level,
			StatusMessage:       statusMessage,
			Version:             version,
			Input:               inputStr,
			Output:              outputStr,
			Model:               input.Model,
			ModelParameters:     modelParams,
			UsageDetails:        usageDetails,
			PromptName:          input.PromptName,
			DurationMs:          durationMs,
			CreatedAt:           now,
			UpdatedAt:           now,
		}

		// Calculate costs
		if usageDetails.TotalTokens > 0 && input.Model != "" && s.costService != nil {
			cost, err := s.costService.CalculateCost(ctx, projectID, input.Model, int64(usageDetails.InputTokens), int64(usageDetails.OutputTokens))
			if err == nil && cost != nil {
				obs.CostDetails = *cost
			}
		}

		observations = append(observations, obs)
	}

	if len(observations) > 0 {
		if err := s.observationRepo.CreateBatch(ctx, observations); err != nil {
			return fmt.Errorf("failed to batch create observations: %w", err)
		}
	}

	return nil
}

// UpdateTrace updates an existing trace
func (s *IngestionService) UpdateTrace(ctx context.Context, projectID uuid.UUID, traceID string, input *domain.TraceInput) (*domain.Trace, error) {
	trace, err := s.traceRepo.GetByID(ctx, projectID, traceID)
	if err != nil {
		return nil, err
	}

	// Update fields
	if input.Name != "" {
		trace.Name = input.Name
	}
	if input.UserID != "" {
		trace.UserID = input.UserID
	}
	if input.SessionID != "" {
		trace.SessionID = input.SessionID
	}
	if input.Metadata != nil {
		metadataBytes, _ := json.Marshal(input.Metadata)
		trace.Metadata = string(metadataBytes)
	}
	if len(input.Tags) > 0 {
		trace.Tags = input.Tags
	}
	if input.Release != "" {
		trace.Release = input.Release
	}
	if input.Version != "" {
		trace.Version = input.Version
	}

	trace.UpdatedAt = time.Now()

	if err := s.traceRepo.Update(ctx, trace); err != nil {
		return nil, fmt.Errorf("failed to update trace: %w", err)
	}

	return trace, nil
}

// UpdateObservation updates an existing observation
func (s *IngestionService) UpdateObservation(ctx context.Context, projectID uuid.UUID, obsID string, input *domain.ObservationInput) (*domain.Observation, error) {
	obs, err := s.observationRepo.GetByID(ctx, projectID, obsID)
	if err != nil {
		return nil, err
	}

	// Update fields
	if input.Name != nil && *input.Name != "" {
		obs.Name = *input.Name
	}
	if input.EndTime != nil {
		obs.EndTime = input.EndTime
		obs.DurationMs = float64(input.EndTime.Sub(obs.StartTime).Milliseconds())
	}
	if input.Metadata != nil {
		metadataBytes, _ := json.Marshal(input.Metadata)
		obs.Metadata = string(metadataBytes)
	}
	if input.Output != nil {
		outputBytes, _ := json.Marshal(input.Output)
		obs.Output = string(outputBytes)
	}
	if input.Level != nil {
		obs.Level = *input.Level
	}
	if input.StatusMessage != nil && *input.StatusMessage != "" {
		obs.StatusMessage = *input.StatusMessage
	}

	obs.UpdatedAt = time.Now()

	if err := s.observationRepo.Update(ctx, obs); err != nil {
		return nil, fmt.Errorf("failed to update observation: %w", err)
	}

	return obs, nil
}

// updateTraceCosts recalculates and updates trace costs
func (s *IngestionService) updateTraceCosts(ctx context.Context, projectID uuid.UUID, traceID string) error {
	observations, err := s.observationRepo.GetByTraceID(ctx, projectID, traceID)
	if err != nil {
		return err
	}

	var inputCost, outputCost, totalCost float64
	for _, obs := range observations {
		inputCost += obs.CostDetails.InputCost
		outputCost += obs.CostDetails.OutputCost
		totalCost += obs.CostDetails.TotalCost
	}

	return s.traceRepo.UpdateCosts(ctx, projectID, traceID, inputCost, outputCost, totalCost)
}

// IngestionBatch represents a batch of ingestion items
type IngestionBatchInput struct {
	Traces       []*domain.TraceInput      `json:"traces"`
	Observations []*domain.ObservationInput `json:"observations"`
	Generations  []*domain.GenerationInput  `json:"generations"`
}
