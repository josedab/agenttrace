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
	metadata := "{}"
	if input.Metadata != nil {
		metadataBytes, err := json.Marshal(input.Metadata)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal metadata: %w", err)
		}
		metadata = string(metadataBytes)
	}

	inputJSON, err := marshalTraceValue(input.Input)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal trace input: %w", err)
	}
	outputJSON, err := marshalTraceValue(input.Output)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal trace output: %w", err)
	}

	level := input.Level
	if level == "" {
		level = domain.LevelDefault
	}

	// Set timestamps
	startTime := now
	if input.StartTime != nil {
		startTime = *input.StartTime
	} else if input.Timestamp != nil {
		startTime = *input.Timestamp
	}

	trace := &domain.Trace{
		ID:            traceID,
		ProjectID:     projectID,
		Name:          input.Name,
		UserID:        input.UserID,
		SessionID:     input.SessionID,
		Metadata:      metadata,
		Tags:          input.Tags,
		Release:       input.Release,
		Version:       input.Version,
		Public:        input.Public,
		Input:         inputJSON,
		Output:        outputJSON,
		Level:         level,
		StatusMessage: input.StatusMessage,
		StartTime:     startTime,
		EndTime:       input.EndTime,
		GitCommitSha:  input.GitCommitSha,
		GitBranch:     input.GitBranch,
		GitRepoURL:    input.GitRepoURL,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := s.traceRepo.Create(ctx, trace); err != nil {
		return nil, fmt.Errorf("failed to create trace: %w", err)
	}

	// Evaluate guardrails asynchronously
	if s.guardrailService != nil {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			result, err := s.guardrailService.Evaluate(ctx, projectID, trace, nil)
			if err != nil {
				s.logger.Error("failed to evaluate guardrails",
					zap.String("trace_id", trace.ID),
					zap.Error(err),
				)
				return
			}
			if !result.Passed {
				s.logger.Warn("guardrail violations detected",
					zap.String("trace_id", trace.ID),
					zap.Int("violations", len(result.Violations)),
				)
			}
		}()
	}

	// Trigger evaluators asynchronously
	if s.evalService != nil {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			if err := s.evalService.TriggerForTrace(ctx, projectID, trace); err != nil {
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

func marshalTraceValue(value any) (string, error) {
	if value == nil {
		return "", nil
	}
	data, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	return string(data), nil
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
		metadataBytes, err := json.Marshal(input.Metadata)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal metadata: %w", err)
		}
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
