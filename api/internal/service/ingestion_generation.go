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
