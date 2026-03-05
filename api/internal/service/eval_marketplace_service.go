package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

// EvalMarketplaceService handles evaluation dataset marketplace operations
type EvalMarketplaceService struct {
	logger   *zap.Logger
	dataset  *DatasetService
	listings map[uuid.UUID]*domain.EvalDatasetListing
}

// NewEvalMarketplaceService creates a new eval marketplace service
func NewEvalMarketplaceService(logger *zap.Logger, dataset *DatasetService) *EvalMarketplaceService {
	svc := &EvalMarketplaceService{
		logger:   logger,
		dataset:  dataset,
		listings: make(map[uuid.UUID]*domain.EvalDatasetListing),
	}
	svc.seedDefaults()
	return svc
}

// ListDatasets searches and filters marketplace datasets
func (s *EvalMarketplaceService) ListDatasets(ctx context.Context, search *domain.EvalMarketplaceSearch) ([]domain.EvalDatasetListing, error) {
	s.logger.Info("listing marketplace datasets",
		zap.String("query", search.Query),
		zap.Int("limit", search.Limit),
		zap.Int("offset", search.Offset),
	)

	results := make([]domain.EvalDatasetListing, 0)
	query := strings.ToLower(search.Query)

	for _, listing := range s.listings {
		if !s.matchesSearch(listing, query, search) {
			continue
		}
		results = append(results, *listing)
	}

	// Apply pagination
	if search.Offset >= len(results) {
		return []domain.EvalDatasetListing{}, nil
	}
	end := search.Offset + search.Limit
	if end > len(results) {
		end = len(results)
	}

	return results[search.Offset:end], nil
}

// GetDataset returns a single marketplace listing
func (s *EvalMarketplaceService) GetDataset(ctx context.Context, datasetID uuid.UUID) (*domain.EvalDatasetListing, error) {
	listing, ok := s.listings[datasetID]
	if !ok {
		return nil, fmt.Errorf("marketplace dataset not found: %s", datasetID)
	}
	return listing, nil
}

// PublishDataset publishes a dataset to the marketplace
func (s *EvalMarketplaceService) PublishDataset(ctx context.Context, projectID, authorID uuid.UUID, input *domain.EvalDatasetPublishInput) (*domain.EvalDatasetListing, error) {
	s.logger.Info("publishing dataset to marketplace",
		zap.String("projectId", projectID.String()),
		zap.String("name", input.Name),
	)

	now := time.Now()
	listing := &domain.EvalDatasetListing{
		ID:          uuid.New(),
		Name:        input.Name,
		Description: input.Description,
		Author:      "user-" + authorID.String()[:8],
		AuthorID:    authorID,
		Category:    input.Category,
		TaskType:    input.TaskType,
		SampleCount: 0,
		ScoringRubric: input.ScoringRubric,
		BaselineScores: []domain.BaselineScore{},
		Tags:        input.Tags,
		Downloads:   0,
		Rating:      0,
		RatingCount: 0,
		IsVerified:  false,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	s.listings[listing.ID] = listing

	s.logger.Info("dataset published to marketplace",
		zap.String("listingId", listing.ID.String()),
		zap.String("name", listing.Name),
	)

	return listing, nil
}

// ImportDataset imports a marketplace dataset into a project
func (s *EvalMarketplaceService) ImportDataset(ctx context.Context, input *domain.EvalDatasetImportInput) error {
	s.logger.Info("importing marketplace dataset",
		zap.String("datasetId", input.DatasetID.String()),
		zap.String("projectId", input.ProjectID.String()),
	)

	listing, ok := s.listings[input.DatasetID]
	if !ok {
		return fmt.Errorf("marketplace dataset not found: %s", input.DatasetID)
	}

	listing.Downloads++
	listing.UpdatedAt = time.Now()

	s.logger.Info("dataset imported",
		zap.String("datasetId", input.DatasetID.String()),
		zap.Int("downloads", listing.Downloads),
	)

	return nil
}

// RateDataset rates a marketplace dataset (1-5)
func (s *EvalMarketplaceService) RateDataset(ctx context.Context, datasetID, userID uuid.UUID, score int, review string) error {
	s.logger.Info("rating marketplace dataset",
		zap.String("datasetId", datasetID.String()),
		zap.String("userId", userID.String()),
		zap.Int("score", score),
	)

	if score < 1 || score > 5 {
		return fmt.Errorf("rating score must be between 1 and 5, got %d", score)
	}

	listing, ok := s.listings[datasetID]
	if !ok {
		return fmt.Errorf("marketplace dataset not found: %s", datasetID)
	}

	// Update running average
	totalScore := listing.Rating * float64(listing.RatingCount)
	listing.RatingCount++
	listing.Rating = (totalScore + float64(score)) / float64(listing.RatingCount)
	listing.UpdatedAt = time.Now()

	return nil
}

// ListCategories returns the fixed list of marketplace categories
func (s *EvalMarketplaceService) ListCategories(ctx context.Context) ([]domain.MarketplaceCategory, error) {
	return []domain.MarketplaceCategory{
		{Name: "Code Generation", Description: "Datasets for evaluating code generation quality", Count: 15, Icon: "code"},
		{Name: "Question Answering", Description: "Datasets for evaluating QA capabilities", Count: 12, Icon: "help-circle"},
		{Name: "Summarization", Description: "Datasets for evaluating text summarization", Count: 8, Icon: "file-text"},
		{Name: "Classification", Description: "Datasets for evaluating classification tasks", Count: 10, Icon: "tag"},
		{Name: "Reasoning", Description: "Datasets for evaluating logical reasoning", Count: 6, Icon: "brain"},
		{Name: "Tool Use", Description: "Datasets for evaluating tool and API usage", Count: 9, Icon: "wrench"},
		{Name: "Safety", Description: "Datasets for evaluating model safety and alignment", Count: 5, Icon: "shield"},
		{Name: "Multi-turn", Description: "Datasets for evaluating multi-turn conversations", Count: 7, Icon: "message-circle"},
	}, nil
}

func (s *EvalMarketplaceService) matchesSearch(listing *domain.EvalDatasetListing, query string, search *domain.EvalMarketplaceSearch) bool {
	if query != "" {
		name := strings.ToLower(listing.Name)
		desc := strings.ToLower(listing.Description)
		if !strings.Contains(name, query) && !strings.Contains(desc, query) {
			match := false
			for _, tag := range listing.Tags {
				if strings.Contains(strings.ToLower(tag), query) {
					match = true
					break
				}
			}
			if !match {
				return false
			}
		}
	}
	if search.Category != nil && *search.Category != listing.Category {
		return false
	}
	if search.TaskType != nil && *search.TaskType != listing.TaskType {
		return false
	}
	if search.MinRating != nil && listing.Rating < *search.MinRating {
		return false
	}
	return true
}

func (s *EvalMarketplaceService) seedDefaults() {
	baseTime := time.Date(2024, 6, 1, 10, 0, 0, 0, time.UTC)

	defaults := []domain.EvalDatasetListing{
		{
			ID:          uuid.New(),
			Name:        "HumanEval-Plus",
			Description: "Extended version of HumanEval with additional test cases for Python code generation",
			Author:      "agenttrace-team",
			AuthorID:    uuid.New(),
			Category:    "Code Generation",
			TaskType:    "coding",
			SampleCount: 164,
			ScoringRubric: domain.ScoringRubric{
				Criteria: []domain.RubricCriterion{
					{Name: "Correctness", Description: "Code passes all test cases", Weight: 0.7, MaxPoints: 100},
					{Name: "Efficiency", Description: "Code runs within time limits", Weight: 0.3, MaxPoints: 100},
				},
				MaxScore:     100,
				PassingScore: 70,
			},
			BaselineScores: []domain.BaselineScore{
				{Model: "gpt-4", Score: 87.1, EvaluatedAt: baseTime, SampleSize: 164},
				{Model: "claude-3-opus", Score: 84.9, EvaluatedAt: baseTime, SampleSize: 164},
			},
			Tags:        []string{"python", "code-generation", "function-completion"},
			Downloads:   1250,
			Rating:      4.7,
			RatingCount: 89,
			IsVerified:  true,
			CreatedAt:   baseTime,
			UpdatedAt:   baseTime.Add(30 * 24 * time.Hour),
		},
		{
			ID:          uuid.New(),
			Name:        "AgentBench-QA",
			Description: "Multi-domain question answering benchmark for AI agents",
			Author:      "benchmark-lab",
			AuthorID:    uuid.New(),
			Category:    "Question Answering",
			TaskType:    "qa",
			SampleCount: 500,
			ScoringRubric: domain.ScoringRubric{
				Criteria: []domain.RubricCriterion{
					{Name: "Accuracy", Description: "Answer is factually correct", Weight: 0.5, MaxPoints: 100},
					{Name: "Completeness", Description: "Answer covers all aspects", Weight: 0.3, MaxPoints: 100},
					{Name: "Relevance", Description: "Answer is relevant to the question", Weight: 0.2, MaxPoints: 100},
				},
				MaxScore:     100,
				PassingScore: 65,
			},
			BaselineScores: []domain.BaselineScore{
				{Model: "gpt-4", Score: 78.3, EvaluatedAt: baseTime, SampleSize: 500},
			},
			Tags:        []string{"qa", "multi-domain", "knowledge"},
			Downloads:   832,
			Rating:      4.4,
			RatingCount: 56,
			IsVerified:  true,
			CreatedAt:   baseTime.Add(7 * 24 * time.Hour),
			UpdatedAt:   baseTime.Add(45 * 24 * time.Hour),
		},
		{
			ID:          uuid.New(),
			Name:        "ToolUse-Eval",
			Description: "Evaluation dataset for measuring AI tool usage accuracy and efficiency",
			Author:      "tools-research",
			AuthorID:    uuid.New(),
			Category:    "Tool Use",
			TaskType:    "custom",
			SampleCount: 320,
			ScoringRubric: domain.ScoringRubric{
				Criteria: []domain.RubricCriterion{
					{Name: "Tool Selection", Description: "Correct tool chosen for the task", Weight: 0.4, MaxPoints: 100},
					{Name: "Parameter Accuracy", Description: "Correct parameters passed", Weight: 0.4, MaxPoints: 100},
					{Name: "Efficiency", Description: "Minimal unnecessary tool calls", Weight: 0.2, MaxPoints: 100},
				},
				MaxScore:     100,
				PassingScore: 60,
			},
			BaselineScores: []domain.BaselineScore{
				{Model: "gpt-4", Score: 72.5, EvaluatedAt: baseTime, SampleSize: 320},
				{Model: "claude-3-opus", Score: 71.8, EvaluatedAt: baseTime, SampleSize: 320},
			},
			Tags:        []string{"tool-use", "function-calling", "api"},
			Downloads:   445,
			Rating:      4.2,
			RatingCount: 34,
			IsVerified:  false,
			CreatedAt:   baseTime.Add(14 * 24 * time.Hour),
			UpdatedAt:   baseTime.Add(60 * 24 * time.Hour),
		},
	}

	for i := range defaults {
		s.listings[defaults[i].ID] = &defaults[i]
	}
}
