package service

import (
	"context"
	"fmt"
	"sort"
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
	return &EvalMarketplaceService{
		logger:   logger,
		dataset:  dataset,
		listings: make(map[uuid.UUID]*domain.EvalDatasetListing),
	}
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
		ID:             uuid.New(),
		Name:           input.Name,
		Description:    input.Description,
		Author:         "user-" + authorID.String()[:8],
		AuthorID:       authorID,
		Category:       input.Category,
		TaskType:       input.TaskType,
		SampleCount:    0,
		ScoringRubric:  input.ScoringRubric,
		BaselineScores: []domain.BaselineScore{},
		Tags:           input.Tags,
		Downloads:      0,
		Rating:         0,
		RatingCount:    0,
		IsVerified:     false,
		CreatedAt:      now,
		UpdatedAt:      now,
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

// ListCategories derives categories from published compatibility listings.
func (s *EvalMarketplaceService) ListCategories(ctx context.Context) ([]domain.MarketplaceCategory, error) {
	counts := make(map[string]int)
	for _, listing := range s.listings {
		if listing.Category != "" {
			counts[listing.Category]++
		}
	}
	names := make([]string, 0, len(counts))
	for name := range counts {
		names = append(names, name)
	}
	sort.Strings(names)

	categories := make([]domain.MarketplaceCategory, 0, len(names))
	for _, name := range names {
		categories = append(categories, domain.MarketplaceCategory{
			Name:        name,
			Description: "Published evaluation datasets",
			Count:       counts[name],
			Icon:        "package",
		})
	}
	return categories, nil
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
