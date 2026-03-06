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
			metadataBytes, err := json.Marshal(input.Metadata)
			if err != nil {
				s.logger.Warn("failed to marshal trace metadata in batch, skipping metadata",
					zap.String("trace_id", traceID),
					zap.Error(err),
				)
			} else {
				metadata = string(metadataBytes)
			}
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
			metadataBytes, err := json.Marshal(input.Metadata)
			if err != nil {
				s.logger.Warn("failed to marshal observation metadata in batch, skipping metadata",
					zap.String("observation_id", obsID),
					zap.Error(err),
				)
			} else {
				metadata = string(metadataBytes)
			}
		}
		if input.Input != nil {
			inputBytes, err := json.Marshal(input.Input)
			if err != nil {
				s.logger.Warn("failed to marshal observation input in batch, skipping input",
					zap.String("observation_id", obsID),
					zap.Error(err),
				)
			} else {
				inputStr = string(inputBytes)
			}
		}
		if input.Output != nil {
			outputBytes, err := json.Marshal(input.Output)
			if err != nil {
				s.logger.Warn("failed to marshal observation output in batch, skipping output",
					zap.String("observation_id", obsID),
					zap.Error(err),
				)
			} else {
				outputStr = string(outputBytes)
			}
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
			metadataBytes, err := json.Marshal(input.Metadata)
			if err != nil {
				s.logger.Warn("failed to marshal generation metadata in batch, skipping metadata",
					zap.String("observation_id", obsID),
					zap.Error(err),
				)
			} else {
				metadata = string(metadataBytes)
			}
		}
		if input.Input != nil {
			inputBytes, err := json.Marshal(input.Input)
			if err != nil {
				s.logger.Warn("failed to marshal generation input in batch, skipping input",
					zap.String("observation_id", obsID),
					zap.Error(err),
				)
			} else {
				inputStr = string(inputBytes)
			}
		}
		if input.Output != nil {
			outputBytes, err := json.Marshal(input.Output)
			if err != nil {
				s.logger.Warn("failed to marshal generation output in batch, skipping output",
					zap.String("observation_id", obsID),
					zap.Error(err),
				)
			} else {
				outputStr = string(outputBytes)
			}
		}
		if input.ModelParameters != nil {
			paramsBytes, err := json.Marshal(input.ModelParameters)
			if err != nil {
				s.logger.Warn("failed to marshal model parameters in batch, skipping model params",
					zap.String("observation_id", obsID),
					zap.Error(err),
				)
			} else {
				modelParams = string(paramsBytes)
			}
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
