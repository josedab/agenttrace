package service

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

// PredictiveCostService manages predictive cost modeling
type PredictiveCostService struct {
	logger      *zap.Logger
	mu          sync.RWMutex
	predictions map[uuid.UUID]*domain.CostPrediction
	approvals   map[uuid.UUID]*domain.BudgetApproval
}

// NewPredictiveCostService creates a new predictive cost service
func NewPredictiveCostService(logger *zap.Logger) *PredictiveCostService {
	return &PredictiveCostService{
		logger:      logger,
		predictions: make(map[uuid.UUID]*domain.CostPrediction),
		approvals:   make(map[uuid.UUID]*domain.BudgetApproval),
	}
}

// PredictCost generates a cost prediction for a task
func (s *PredictiveCostService) PredictCost(ctx context.Context, projectID uuid.UUID, input *domain.PredictionInput) (*domain.CostPrediction, error) {
	model := input.Model
	if model == "" {
		model = s.recommendModel(input.TaskDescription)
	}

	tokens := s.estimateTokens(input.TaskDescription)
	cost := s.estimateCost(model, tokens)
	latency := s.estimateLatency(model, tokens)
	quality := s.estimateQuality(model, input.TaskDescription)

	budgetStatus := "within"
	if input.MaxBudget != nil {
		if cost > *input.MaxBudget {
			budgetStatus = "exceeded"
		} else if cost > *input.MaxBudget*0.8 {
			budgetStatus = "warning"
		}
	}

	prediction := &domain.CostPrediction{
		ID:                 uuid.New(),
		ProjectID:          projectID,
		TaskDescription:    input.TaskDescription,
		PredictedCost:      math.Round(cost*10000) / 10000,
		PredictedLatencyMs: latency,
		PredictedQuality:   math.Round(quality*100) / 100,
		PredictedTokens:    tokens,
		ConfidenceLevel:    0.7 + rand.Float64()*0.25, //nolint:gosec
		RecommendedModel:   model,
		BudgetStatus:       budgetStatus,
		SimilarTraces:      5 + rand.Intn(45), //nolint:gosec
		CreatedAt:          time.Now(),
	}

	s.mu.Lock()
	s.predictions[prediction.ID] = prediction
	s.mu.Unlock()

	s.logger.Info("generated cost prediction",
		zap.String("predictionId", prediction.ID.String()),
		zap.Float64("cost", prediction.PredictedCost),
		zap.Int("tokens", prediction.PredictedTokens),
		zap.String("model", model),
	)

	return prediction, nil
}

// ListPredictions lists all predictions for a project
func (s *PredictiveCostService) ListPredictions(ctx context.Context, projectID uuid.UUID) ([]domain.CostPrediction, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var predictions []domain.CostPrediction
	for _, p := range s.predictions {
		if p.ProjectID == projectID {
			predictions = append(predictions, *p)
		}
	}

	if predictions == nil {
		predictions = []domain.CostPrediction{}
	}
	return predictions, nil
}

// RequestApproval creates a budget approval request for a prediction
func (s *PredictiveCostService) RequestApproval(ctx context.Context, projectID uuid.UUID, predictionID uuid.UUID) (*domain.BudgetApproval, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, exists := s.predictions[predictionID]
	if !exists {
		return nil, fmt.Errorf("prediction not found")
	}

	approval := &domain.BudgetApproval{
		ID:           uuid.New(),
		ProjectID:    projectID,
		PredictionID: predictionID,
		Status:       "pending",
		CreatedAt:    time.Now(),
	}

	s.approvals[approval.ID] = approval

	s.logger.Info("requested budget approval",
		zap.String("approvalId", approval.ID.String()),
		zap.String("predictionId", predictionID.String()),
	)

	return approval, nil
}

// DecideApproval processes an approval decision
func (s *PredictiveCostService) DecideApproval(ctx context.Context, approvalID uuid.UUID, input *domain.ApprovalDecisionInput) (*domain.BudgetApproval, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	approval, exists := s.approvals[approvalID]
	if !exists {
		return nil, fmt.Errorf("approval not found")
	}

	if approval.Status != "pending" {
		return nil, fmt.Errorf("approval already decided")
	}

	now := time.Now()
	approval.Status = input.Status
	approval.Note = input.Note
	approval.DecidedAt = &now

	s.logger.Info("decided budget approval",
		zap.String("approvalId", approvalID.String()),
		zap.String("status", input.Status),
	)

	return approval, nil
}

func (s *PredictiveCostService) recommendModel(task string) string {
	taskLower := strings.ToLower(task)

	// Simple heuristic based on task complexity
	if strings.Contains(taskLower, "complex") || strings.Contains(taskLower, "analysis") ||
		strings.Contains(taskLower, "architect") || strings.Contains(taskLower, "design") {
		return "gpt-4o"
	}
	if strings.Contains(taskLower, "code") || strings.Contains(taskLower, "implement") ||
		strings.Contains(taskLower, "debug") || strings.Contains(taskLower, "refactor") {
		return "claude-3.5-sonnet"
	}
	if strings.Contains(taskLower, "simple") || strings.Contains(taskLower, "format") ||
		strings.Contains(taskLower, "summarize") || strings.Contains(taskLower, "translate") {
		return "gpt-4o-mini"
	}
	return "gpt-4o"
}

func (s *PredictiveCostService) estimateTokens(task string) int {
	// Base tokens from task description length
	baseTokens := len(task) * 4
	if baseTokens < 500 {
		baseTokens = 500
	}

	// Add variation based on expected output
	outputMultiplier := 3.0 + rand.Float64()*2.0 //nolint:gosec
	return int(float64(baseTokens) * outputMultiplier)
}

func (s *PredictiveCostService) estimateCost(model string, tokens int) float64 {
	// Cost per 1K tokens (approximate)
	costPer1K := map[string]float64{
		"gpt-4o":            0.005,
		"gpt-4o-mini":       0.00015,
		"claude-3.5-sonnet": 0.003,
		"claude-3-haiku":    0.00025,
	}

	rate, exists := costPer1K[model]
	if !exists {
		rate = 0.003
	}

	return float64(tokens) / 1000.0 * rate
}

func (s *PredictiveCostService) estimateLatency(model string, tokens int) int64 {
	// Base latency per model (ms)
	baseLatency := map[string]int64{
		"gpt-4o":            2000,
		"gpt-4o-mini":       500,
		"claude-3.5-sonnet": 1500,
		"claude-3-haiku":    300,
	}

	base, exists := baseLatency[model]
	if !exists {
		base = 1500
	}

	// Add token-proportional latency
	tokenLatency := int64(float64(tokens) * 0.5)
	return base + tokenLatency + int64(rand.Intn(500)) //nolint:gosec
}

func (s *PredictiveCostService) estimateQuality(model string, task string) float64 {
	baseQuality := map[string]float64{
		"gpt-4o":            90,
		"gpt-4o-mini":       72,
		"claude-3.5-sonnet": 88,
		"claude-3-haiku":    68,
	}

	quality, exists := baseQuality[model]
	if !exists {
		quality = 80
	}

	// Adjust based on task complexity
	if len(task) > 200 {
		quality -= 3
	}
	quality += (rand.Float64() - 0.5) * 6 //nolint:gosec

	if quality > 100 {
		quality = 100
	}
	if quality < 0 {
		quality = 0
	}
	return quality
}
