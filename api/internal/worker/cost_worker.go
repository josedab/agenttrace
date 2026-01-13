package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

const (
	// TypeCostCalculation is the task type for cost calculation
	TypeCostCalculation = "cost:calculate"
)

// CostCalculationPayload is the payload for cost calculation tasks
type CostCalculationPayload struct {
	ProjectID     string `json:"project_id"`
	TraceID       string `json:"trace_id"`
	ObservationID string `json:"observation_id"`
	Model         string `json:"model"`
	PromptTokens  int    `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
}

// NewCostCalculationTask creates a new cost calculation task
func NewCostCalculationTask(payload *CostCalculationPayload) (*asynq.Task, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal cost calculation payload: %w", err)
	}
	return asynq.NewTask(TypeCostCalculation, data, asynq.MaxRetry(3)), nil
}

// CostWorker handles cost calculation tasks
type CostWorker struct {
	logger       *zap.Logger
	costService  *service.CostService
	queryService *service.QueryService
	ingestionService *service.IngestionService
}

// NewCostWorker creates a new cost worker
func NewCostWorker(
	logger *zap.Logger,
	costService *service.CostService,
	queryService *service.QueryService,
	ingestionService *service.IngestionService,
) *CostWorker {
	return &CostWorker{
		logger:           logger,
		costService:      costService,
		queryService:     queryService,
		ingestionService: ingestionService,
	}
}

// ProcessTask processes a cost calculation task
func (w *CostWorker) ProcessTask(ctx context.Context, t *asynq.Task) error {
	var payload CostCalculationPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal cost calculation payload: %w", err)
	}

	w.logger.Info("processing cost calculation",
		zap.String("project_id", payload.ProjectID),
		zap.String("trace_id", payload.TraceID),
		zap.String("observation_id", payload.ObservationID),
		zap.String("model", payload.Model),
		zap.Int("prompt_tokens", payload.PromptTokens),
		zap.Int("completion_tokens", payload.CompletionTokens),
	)

	// Parse project ID
	projectID, err := uuid.Parse(payload.ProjectID)
	if err != nil {
		return fmt.Errorf("invalid project ID: %w", err)
	}

	// Calculate cost
	cost, err := w.costService.CalculateCost(ctx, projectID, payload.Model, int64(payload.PromptTokens), int64(payload.CompletionTokens))
	if err != nil {
		w.logger.Warn("failed to calculate cost", zap.Error(err))
		return nil // Don't retry for unknown models
	}
	if cost == nil {
		w.logger.Debug("no pricing available for model", zap.String("model", payload.Model))
		return nil
	}

	// Update observation with calculated costs
	if err := w.ingestionService.UpdateObservationCosts(
		ctx,
		projectID,
		payload.ObservationID,
		payload.TraceID,
		cost.InputCost,
		cost.OutputCost,
		cost.TotalCost,
	); err != nil {
		w.logger.Error("failed to update observation costs",
			zap.String("observation_id", payload.ObservationID),
			zap.Error(err),
		)
		return fmt.Errorf("failed to update observation costs: %w", err)
	}

	w.logger.Info("cost calculation completed",
		zap.String("observation_id", payload.ObservationID),
		zap.Float64("input_cost", cost.InputCost),
		zap.Float64("output_cost", cost.OutputCost),
		zap.Float64("total_cost", cost.TotalCost),
	)

	return nil
}

// BatchCostCalculationPayload is the payload for batch cost calculation
type BatchCostCalculationPayload struct {
	ProjectID string   `json:"project_id"`
	TraceID   string   `json:"trace_id"`
}

// TypeBatchCostCalculation is the task type for batch cost calculation
const TypeBatchCostCalculation = "cost:calculate-batch"

// NewBatchCostCalculationTask creates a batch cost calculation task
func NewBatchCostCalculationTask(payload *BatchCostCalculationPayload) (*asynq.Task, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal batch cost calculation payload: %w", err)
	}
	return asynq.NewTask(TypeBatchCostCalculation, data, asynq.MaxRetry(3)), nil
}

// ProcessBatchCostTask processes a batch cost calculation task
func (w *CostWorker) ProcessBatchCostTask(ctx context.Context, t *asynq.Task) error {
	var payload BatchCostCalculationPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal batch cost calculation payload: %w", err)
	}

	w.logger.Info("processing batch cost calculation",
		zap.String("project_id", payload.ProjectID),
		zap.String("trace_id", payload.TraceID),
	)

	// Parse project ID
	projectID, err := uuid.Parse(payload.ProjectID)
	if err != nil {
		return fmt.Errorf("invalid project ID: %w", err)
	}

	// Get all observations for the trace
	observations, err := w.queryService.GetObservationsByTraceID(ctx, projectID, payload.TraceID)
	if err != nil {
		w.logger.Error("failed to get observations for trace",
			zap.String("trace_id", payload.TraceID),
			zap.Error(err),
		)
		return fmt.Errorf("failed to get observations: %w", err)
	}

	var processed, skipped int
	for _, obs := range observations {
		// Skip observations that already have costs calculated
		if obs.CostDetails.TotalCost > 0 {
			skipped++
			continue
		}

		// Skip observations without model or token data
		if obs.Model == "" || (obs.UsageDetails.InputTokens == 0 && obs.UsageDetails.OutputTokens == 0) {
			skipped++
			continue
		}

		// Calculate cost
		cost, err := w.costService.CalculateCost(
			ctx,
			projectID,
			obs.Model,
			int64(obs.UsageDetails.InputTokens),
			int64(obs.UsageDetails.OutputTokens),
		)
		if err != nil {
			w.logger.Warn("failed to calculate cost for observation",
				zap.String("observation_id", obs.ID),
				zap.String("model", obs.Model),
				zap.Error(err),
			)
			continue
		}
		if cost == nil {
			// No pricing available for this model
			skipped++
			continue
		}

		// Update observation with calculated costs
		if err := w.ingestionService.UpdateObservationCosts(
			ctx,
			projectID,
			obs.ID,
			obs.TraceID,
			cost.InputCost,
			cost.OutputCost,
			cost.TotalCost,
		); err != nil {
			w.logger.Error("failed to update observation costs",
				zap.String("observation_id", obs.ID),
				zap.Error(err),
			)
			continue
		}

		processed++
	}

	w.logger.Info("batch cost calculation completed",
		zap.String("trace_id", payload.TraceID),
		zap.Int("processed", processed),
		zap.Int("skipped", skipped),
		zap.Int("total", len(observations)),
	)

	return nil
}

// DailyAggregationPayload is the payload for daily cost aggregation
type DailyAggregationPayload struct {
	ProjectID string `json:"project_id"`
	Date      string `json:"date"` // YYYY-MM-DD format
}

// TypeDailyAggregation is the task type for daily aggregation
const TypeDailyAggregation = "cost:aggregate-daily"

// NewDailyAggregationTask creates a daily aggregation task
func NewDailyAggregationTask(payload *DailyAggregationPayload) (*asynq.Task, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal daily aggregation payload: %w", err)
	}
	return asynq.NewTask(TypeDailyAggregation, data, asynq.MaxRetry(3)), nil
}

// ProcessDailyAggregationTask processes a daily aggregation task
func (w *CostWorker) ProcessDailyAggregationTask(ctx context.Context, t *asynq.Task) error {
	var payload DailyAggregationPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal daily aggregation payload: %w", err)
	}

	w.logger.Info("processing daily aggregation",
		zap.String("project_id", payload.ProjectID),
		zap.String("date", payload.Date),
	)

	// Parse project ID
	projectID, err := uuid.Parse(payload.ProjectID)
	if err != nil {
		return fmt.Errorf("invalid project ID: %w", err)
	}

	// Parse date and create time range for the day
	date, err := parseDate(payload.Date)
	if err != nil {
		return fmt.Errorf("invalid date format: %w", err)
	}

	startOfDay := date
	endOfDay := date.AddDate(0, 0, 1)

	// Query all observations for the day with cost data
	filter := &domain.ObservationFilter{
		ProjectID: projectID,
		FromTime:  &startOfDay,
		ToTime:    &endOfDay,
	}

	observations, _, err := w.queryService.ListObservations(ctx, filter, 100000, 0)
	if err != nil {
		w.logger.Error("failed to query observations for daily aggregation",
			zap.String("date", payload.Date),
			zap.Error(err),
		)
		return fmt.Errorf("failed to query observations: %w", err)
	}

	// Aggregate costs by model
	modelCosts := make(map[string]*modelCostAggregation)
	var totalCost float64
	var observationCount int

	for _, obs := range observations {
		if obs.CostDetails.TotalCost <= 0 {
			continue
		}

		totalCost += obs.CostDetails.TotalCost
		observationCount++

		if obs.Model != "" {
			if mc, ok := modelCosts[obs.Model]; ok {
				mc.Cost += obs.CostDetails.TotalCost
				mc.Count++
			} else {
				modelCosts[obs.Model] = &modelCostAggregation{
					Model: obs.Model,
					Cost:  obs.CostDetails.TotalCost,
					Count: 1,
				}
			}
		}
	}

	w.logger.Info("daily aggregation completed",
		zap.String("project_id", payload.ProjectID),
		zap.String("date", payload.Date),
		zap.Float64("total_cost", totalCost),
		zap.Int("observation_count", observationCount),
		zap.Int("model_count", len(modelCosts)),
	)

	return nil
}

// modelCostAggregation holds aggregated cost data for a model
type modelCostAggregation struct {
	Model string
	Cost  float64
	Count int
}

// parseDate parses a date string in YYYY-MM-DD format
func parseDate(dateStr string) (time.Time, error) {
	return time.Parse("2006-01-02", dateStr)
}
