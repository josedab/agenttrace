package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/pkg/id"
)

// TraceRepository defines the interface for trace persistence operations.
// Implementations may use ClickHouse, PostgreSQL, or other storage backends.
// All methods must be safe for concurrent use.
type TraceRepository interface {
	// Create persists a new trace to storage.
	Create(ctx context.Context, trace *domain.Trace) error
	// CreateBatch persists multiple traces in a single operation for efficiency.
	CreateBatch(ctx context.Context, traces []*domain.Trace) error
	// GetByID retrieves a trace by its project-scoped ID.
	GetByID(ctx context.Context, projectID uuid.UUID, traceID string) (*domain.Trace, error)
	// Update modifies an existing trace's mutable fields.
	Update(ctx context.Context, trace *domain.Trace) error
	// UpdateCosts updates the aggregated cost fields for a trace.
	UpdateCosts(ctx context.Context, projectID uuid.UUID, traceID string, inputCost, outputCost, totalCost float64) error
	// List returns traces matching the filter with pagination.
	List(ctx context.Context, filter *domain.TraceFilter, limit, offset int) (*domain.TraceList, error)
	// SetBookmark marks or unmarks a trace as bookmarked.
	SetBookmark(ctx context.Context, projectID uuid.UUID, traceID string, bookmarked bool) error
	// GetBySessionID retrieves all traces belonging to a session.
	GetBySessionID(ctx context.Context, projectID uuid.UUID, sessionID string) ([]domain.Trace, error)
	// Delete removes a trace by ID.
	// Note: This is a heavy operation in ClickHouse (ALTER TABLE DELETE).
	Delete(ctx context.Context, projectID uuid.UUID, traceID string) error
}

// ObservationRepository defines the interface for observation persistence operations.
// Observations include spans (generic operations) and generations (LLM calls).
// All methods must be safe for concurrent use.
type ObservationRepository interface {
	// Create persists a new observation to storage.
	Create(ctx context.Context, obs *domain.Observation) error
	// CreateBatch persists multiple observations in a single operation for efficiency.
	CreateBatch(ctx context.Context, observations []*domain.Observation) error
	// GetByID retrieves an observation by its project-scoped ID.
	GetByID(ctx context.Context, projectID uuid.UUID, observationID string) (*domain.Observation, error)
	// GetByTraceID retrieves all observations belonging to a trace.
	GetByTraceID(ctx context.Context, projectID uuid.UUID, traceID string) ([]domain.Observation, error)
	// Update modifies an existing observation's mutable fields.
	Update(ctx context.Context, obs *domain.Observation) error
	// UpdateCosts updates the cost fields for an observation.
	UpdateCosts(ctx context.Context, projectID uuid.UUID, observationID string, inputCost, outputCost, totalCost float64) error
	// List returns observations matching the filter with pagination.
	List(ctx context.Context, filter *domain.ObservationFilter, limit, offset int) ([]domain.Observation, int64, error)
	// GetGenerationsWithoutCost retrieves generations that need cost calculation.
	GetGenerationsWithoutCost(ctx context.Context, projectID uuid.UUID, limit int) ([]domain.Observation, error)
	// GetTree retrieves observations as a hierarchical tree for visualization.
	GetTree(ctx context.Context, projectID uuid.UUID, traceID string) (*domain.ObservationTree, error)
}

// SessionRepository defines the interface for session persistence operations.
// Sessions group related traces together (e.g., a user conversation).
type SessionRepository interface {
	// Upsert creates or updates a session, typically called when traces reference it.
	Upsert(ctx context.Context, session *domain.Session) error
	// GetByID retrieves a session by its project-scoped ID.
	GetByID(ctx context.Context, projectID uuid.UUID, sessionID string) (*domain.Session, error)
	// List returns sessions matching the filter with pagination.
	List(ctx context.Context, filter *domain.SessionFilter, limit, offset int) (*domain.SessionList, error)
}

// IngestionService handles trace and observation ingestion from SDKs and APIs.
//
// This is the core service for receiving telemetry data from instrumented applications.
// It processes incoming traces, spans, and LLM generations, persisting them to storage
// while handling:
//   - ID generation for entities without explicit IDs
//   - Metadata and payload JSON marshaling
//   - Timestamp normalization and duration calculation
//   - Cost calculation for LLM generations (via CostService)
//   - Asynchronous evaluation triggering (via EvalService)
//
// The service is safe for concurrent use and designed for high-throughput ingestion.
type IngestionService struct {
	traceRepo       TraceRepository
	observationRepo ObservationRepository
	costService     *CostService
	evalService     *EvalService
	logger          *zap.Logger
}

// NewIngestionService creates a new IngestionService with the provided dependencies.
//
// Parameters:
//   - logger: Structured logger for observability (required)
//   - traceRepo: Repository for trace persistence (required)
//   - observationRepo: Repository for observation persistence (required)
//   - costService: Service for LLM cost calculation (optional, costs won't be calculated if nil)
//   - evalService: Service for triggering evaluations (optional, evals won't trigger if nil)
//
// Returns a configured IngestionService ready for use.
func NewIngestionService(
	logger *zap.Logger,
	traceRepo TraceRepository,
	observationRepo ObservationRepository,
	costService *CostService,
	evalService *EvalService,
) *IngestionService {
	return &IngestionService{
		logger:          logger.Named("ingestion"),
		traceRepo:       traceRepo,
		observationRepo: observationRepo,
		costService:     costService,
		evalService:     evalService,
	}
}

// IngestTrace ingests a single trace into the system.
//
// A trace represents a complete execution flow (e.g., an API request, agent task).
// This method handles:
//   - Generating a trace ID if not provided in input
//   - Marshaling metadata to JSON for storage
//   - Setting timestamps (uses input.StartTime, falls back to input.Timestamp, then now)
//   - Persisting the trace to storage
//   - Triggering any configured evaluators asynchronously
//
// Parameters:
//   - ctx: Request context for cancellation and deadlines
//   - projectID: The project this trace belongs to (from API key or auth)
//   - input: Trace data including name, metadata, timing, and optional git info
//
// Returns:
//   - *domain.Trace: The created trace with all generated fields populated
//   - error: Returns error on metadata marshaling failure or database error
//
// Side Effects:
//   - Triggers evaluators asynchronously via goroutine (errors are silently ignored)
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
	}

	if err := s.traceRepo.Create(ctx, trace); err != nil {
		return nil, fmt.Errorf("failed to create trace: %w", err)
	}

	// Trigger evaluators asynchronously
	if s.evalService != nil {
		go func() {
			if err := s.evalService.TriggerForTrace(context.Background(), projectID, trace); err != nil {
				s.logger.Error("failed to trigger evaluators for trace",
					zap.String("trace_id", trace.ID),
					zap.String("project_id", projectID.String()),
					zap.Error(err),
				)
			}
		}()
	}

	return trace, nil
}

// IngestObservation ingests a single observation (span or event).
//
// An observation represents a unit of work within a trace. This method is typically
// used for spans (generic operations like function calls, API requests) rather than
// LLM generations (use IngestGeneration for those to get cost calculation).
//
// This method handles:
//   - Generating an observation ID if not provided
//   - Marshaling metadata, input, and output to JSON
//   - Setting timestamps and defaults for optional fields
//   - Persisting the observation to storage
//
// Parameters:
//   - ctx: Request context for cancellation and deadlines
//   - projectID: The project this observation belongs to
//   - input: Observation data including trace ID, name, type, and I/O
//
// Returns:
//   - *domain.Observation: The created observation with generated fields
//   - error: Returns error on JSON marshaling failure or database error
//
// Note: For LLM calls, prefer IngestGeneration which handles cost calculation
// and LLM-specific fields like model, usage, and model parameters.
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

// IngestGeneration ingests an LLM generation (model call) observation.
//
// This is the primary method for recording LLM API calls. It extends basic observation
// handling with LLM-specific features:
//   - Token usage normalization from various provider formats
//   - Cost calculation using configured pricing (via CostService)
//   - Model parameter storage for reproducibility
//   - Duration calculation from start/end times
//   - Prompt name tracking for prompt management integration
//
// This method handles:
//   - Generating an observation ID if not provided
//   - Marshaling all JSON fields (metadata, input, output, model params)
//   - Normalizing token usage from different SDK formats
//   - Calculating costs based on model and token counts
//   - Updating parent trace's aggregated costs asynchronously
//   - Triggering evaluators asynchronously
//
// Parameters:
//   - ctx: Request context for cancellation and deadlines
//   - projectID: The project this generation belongs to
//   - input: Generation data including model, usage, prompt/completion, and timing
//
// Returns:
//   - *domain.Observation: The created generation observation with costs calculated
//   - error: Returns error on JSON marshaling failure or database error
//
// Side Effects:
//   - Updates trace costs asynchronously via goroutine
//   - Triggers evaluators asynchronously via goroutine
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
			if err := s.updateTraceCosts(context.Background(), projectID, traceID); err != nil {
				s.logger.Error("failed to update trace costs",
					zap.String("trace_id", traceID),
					zap.String("observation_id", obsID),
					zap.String("project_id", projectID.String()),
					zap.Error(err),
				)
			}
		}()
	}

	// Trigger evaluators asynchronously
	if s.evalService != nil {
		go func() {
			if err := s.evalService.TriggerForObservation(context.Background(), projectID, obs); err != nil {
				s.logger.Error("failed to trigger evaluators for observation",
					zap.String("observation_id", obs.ID),
					zap.String("trace_id", obs.TraceID),
					zap.String("project_id", projectID.String()),
					zap.Error(err),
				)
			}
		}()
	}

	return obs, nil
}

// IngestBatch ingests multiple traces and observations in a single operation.
//
// This method is optimized for high-throughput ingestion scenarios where SDKs
// buffer telemetry and send it periodically. It processes:
//   - Multiple traces in a single batch insert
//   - Multiple observations (spans) in a single batch insert
//   - Multiple generations with cost calculation in a single batch insert
//
// Unlike the individual Ingest* methods, batch ingestion:
//   - Does NOT trigger evaluators (for performance)
//   - Does NOT update trace costs asynchronously
//   - Silently ignores JSON marshaling errors on individual items
//   - Fails atomically if any database batch operation fails
//
// Parameters:
//   - ctx: Request context for cancellation and deadlines
//   - projectID: The project all items in the batch belong to
//   - batch: Container with arrays of traces, observations, and generations
//
// Returns:
//   - error: Returns error if trace or observation batch insert fails
//
// Performance: Use this method when ingesting multiple items at once.
// The batch operations are significantly more efficient than individual inserts.
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

// UpdateTrace updates an existing trace with new field values.
//
// This method supports partial updates - only non-empty fields in the input
// are applied to the existing trace. Useful for:
//   - Setting end time when a trace completes
//   - Adding metadata discovered during execution
//   - Updating tags or session association
//
// Updatable fields: Name, UserID, SessionID, Metadata, Tags, Release, Version
// Non-updatable fields: ID, ProjectID, StartTime, CreatedAt (immutable)
//
// Parameters:
//   - ctx: Request context for cancellation and deadlines
//   - projectID: The project the trace belongs to
//   - traceID: The trace to update
//   - input: Fields to update (only non-empty values are applied)
//
// Returns:
//   - *domain.Trace: The updated trace with all current field values
//   - error: Returns error if trace not found or database error
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

// UpdateObservation updates an existing observation with new field values.
//
// This method supports partial updates - only non-nil fields in the input
// are applied to the existing observation. Common use cases:
//   - Setting end time and calculating duration when operation completes
//   - Adding output after async operation finishes
//   - Updating status/level based on operation result
//
// Updatable fields: Name, EndTime, Metadata, Output, Level, StatusMessage
// Non-updatable fields: ID, TraceID, ProjectID, Type, StartTime, CreatedAt
//
// When EndTime is updated, DurationMs is automatically recalculated from StartTime.
//
// Parameters:
//   - ctx: Request context for cancellation and deadlines
//   - projectID: The project the observation belongs to
//   - obsID: The observation to update
//   - input: Fields to update (only non-nil values are applied)
//
// Returns:
//   - *domain.Observation: The updated observation with all current field values
//   - error: Returns error if observation not found or database error
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

// updateTraceCosts recalculates and updates the aggregated costs for a trace.
//
// This method fetches all observations for a trace and sums their costs to
// update the trace's aggregate cost fields. Called asynchronously after
// ingesting a generation with costs.
//
// The trace stores aggregated costs (inputCost, outputCost, totalCost) to
// enable efficient cost queries without joining to observations.
//
// Parameters:
//   - ctx: Request context (typically context.Background() when called async)
//   - projectID: The project the trace belongs to
//   - traceID: The trace to update costs for
//
// Returns:
//   - error: Returns error if observations fetch or trace update fails
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

// IngestionBatchInput represents a batch of telemetry items for bulk ingestion.
//
// SDKs typically buffer telemetry locally and send batches periodically to
// reduce network overhead. This struct mirrors the domain.IngestionBatch but
// uses input types for deserialization from API requests.
//
// All arrays are optional - a batch can contain any combination of traces,
// observations, and generations. Items within each array are processed together
// in efficient batch database operations.
//
// Example JSON:
//
//	{
//	  "traces": [{"id": "trace-1", "name": "api-request"}],
//	  "observations": [{"traceId": "trace-1", "name": "db-query"}],
//	  "generations": [{"traceId": "trace-1", "model": "gpt-4", "usage": {...}}]
//	}
type IngestionBatchInput struct {
	// Traces to create (parent containers for observations)
	Traces []*domain.TraceInput `json:"traces"`
	// Observations to create (spans, events, generic operations)
	Observations []*domain.ObservationInput `json:"observations"`
	// Generations to create (LLM calls with model/usage/cost tracking)
	Generations []*domain.GenerationInput `json:"generations"`
}
