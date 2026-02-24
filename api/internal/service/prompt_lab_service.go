package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

type PromptLabService struct {
	logger      *zap.Logger
	experiments map[uuid.UUID]*domain.PromptExperiment
}

func NewPromptLabService(logger *zap.Logger) *PromptLabService {
	return &PromptLabService{
		logger:      logger,
		experiments: make(map[uuid.UUID]*domain.PromptExperiment),
	}
}

func (s *PromptLabService) CreateExperiment(ctx context.Context, projectID uuid.UUID, input *domain.PromptExperimentInput) (*domain.PromptExperiment, error) {
	exp := &domain.PromptExperiment{
		ID:          uuid.New(),
		ProjectID:   projectID,
		Name:        input.Name,
		Description: input.Description,
		PromptName:  input.PromptName,
		Status:      domain.PromptExpStatusDraft,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	totalWeight := 0.0
	for _, v := range input.Variants {
		variant := domain.PromptVariant{
			ID:            uuid.New(),
			ExperimentID:  exp.ID,
			Name:          v.Name,
			PromptContent: v.PromptContent,
			TrafficWeight: v.TrafficWeight,
			IsControl:     v.IsControl,
		}
		if variant.TrafficWeight == 0 {
			variant.TrafficWeight = 1.0 / float64(len(input.Variants))
		}
		totalWeight += variant.TrafficWeight
		exp.Variants = append(exp.Variants, variant)
	}

	// Normalize weights
	if totalWeight > 0 {
		for i := range exp.Variants {
			exp.Variants[i].TrafficWeight /= totalWeight
		}
	}

	s.experiments[exp.ID] = exp
	s.logger.Info("created prompt experiment",
		zap.String("experimentId", exp.ID.String()),
		zap.String("promptName", exp.PromptName),
		zap.Int("variants", len(exp.Variants)),
	)
	return exp, nil
}

func (s *PromptLabService) GetExperiment(ctx context.Context, id uuid.UUID) (*domain.PromptExperiment, error) {
	exp, ok := s.experiments[id]
	if !ok {
		return nil, fmt.Errorf("experiment not found")
	}
	return exp, nil
}

func (s *PromptLabService) ListExperiments(ctx context.Context, projectID uuid.UUID) ([]domain.PromptExperiment, error) {
	var result []domain.PromptExperiment
	for _, exp := range s.experiments {
		if exp.ProjectID == projectID {
			result = append(result, *exp)
		}
	}
	return result, nil
}

func (s *PromptLabService) StartExperiment(ctx context.Context, id uuid.UUID) (*domain.PromptExperiment, error) {
	exp, ok := s.experiments[id]
	if !ok {
		return nil, fmt.Errorf("experiment not found")
	}
	exp.Status = domain.PromptExpStatusRunning
	exp.UpdatedAt = time.Now()
	return exp, nil
}

func (s *PromptLabService) CompleteExperiment(ctx context.Context, id uuid.UUID) (*domain.PromptExperiment, error) {
	exp, ok := s.experiments[id]
	if !ok {
		return nil, fmt.Errorf("experiment not found")
	}

	exp.Status = domain.PromptExpStatusCompleted
	now := time.Now()
	exp.CompletedAt = &now
	exp.UpdatedAt = now

	// Determine winner by highest quality score
	var bestID uuid.UUID
	bestQuality := -1.0
	for _, v := range exp.Variants {
		if v.Metrics.AvgQuality > bestQuality {
			bestQuality = v.Metrics.AvgQuality
			bestID = v.ID
		}
	}
	if bestQuality >= 0 {
		exp.WinnerID = &bestID
	}

	return exp, nil
}

func (s *PromptLabService) GetOptimizationSuggestions(ctx context.Context, projectID uuid.UUID, promptName string) ([]domain.OptimizationSuggestion, error) {
	return []domain.OptimizationSuggestion{
		{
			ID: uuid.New(), ProjectID: projectID, PromptName: promptName,
			Technique: "compression", OriginalTokens: 500, OptimizedTokens: 320,
			SavingsPercent: 36.0, Description: "Remove redundant instructions and merge similar directives",
			Suggestion: "Consolidate the 3 formatting instructions into a single concise directive",
			Confidence: 0.85, CreatedAt: time.Now(),
		},
		{
			ID: uuid.New(), ProjectID: projectID, PromptName: promptName,
			Technique: "caching", OriginalTokens: 500, OptimizedTokens: 150,
			SavingsPercent: 70.0, Description: "Cache static context that doesn't change between calls",
			Suggestion: "Move system instructions to a cached prefix; only send dynamic context per request",
			Confidence: 0.72, CreatedAt: time.Now(),
		},
	}, nil
}
