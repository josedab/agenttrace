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
		metadataBytes, err := json.Marshal(input.Metadata)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal metadata: %w", err)
		}
		obs.Metadata = string(metadataBytes)
	}
	if input.Output != nil {
		outputBytes, err := json.Marshal(input.Output)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal output: %w", err)
		}
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

// UpdateObservationCosts updates the cost fields for an observation.
//
// This method is called by the cost worker after asynchronously calculating
// costs for observations that were ingested without cost data (e.g., when
// token counts are available but pricing wasn't calculated at ingestion time).
//
// After updating the observation's costs, this method also triggers an
// asynchronous update of the parent trace's aggregated costs.
//
// Parameters:
//   - ctx: Request context for cancellation and deadlines
//   - projectID: The project the observation belongs to
//   - observationID: The observation to update costs for
//   - traceID: The parent trace (for aggregating costs)
//   - inputCost: Cost for input/prompt tokens
//   - outputCost: Cost for output/completion tokens
//   - totalCost: Total cost (inputCost + outputCost)
//
// Returns:
//   - error: Returns error if observation update or trace cost aggregation fails
func (s *IngestionService) UpdateObservationCosts(
	ctx context.Context,
	projectID uuid.UUID,
	observationID string,
	traceID string,
	inputCost, outputCost, totalCost float64,
) error {
	// Update the observation's cost fields
	if err := s.observationRepo.UpdateCosts(ctx, projectID, observationID, inputCost, outputCost, totalCost); err != nil {
		return fmt.Errorf("failed to update observation costs: %w", err)
	}

	// Update the trace's aggregated costs asynchronously
	if traceID != "" {
		go func() {
			if err := s.updateTraceCosts(context.Background(), projectID, traceID); err != nil {
				s.logger.Error("failed to update trace costs after observation cost update",
					zap.String("trace_id", traceID),
					zap.String("observation_id", observationID),
					zap.String("project_id", projectID.String()),
					zap.Error(err),
				)
			}
		}()
	}

	return nil
}
