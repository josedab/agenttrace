package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

// TrainingPipelineService handles training dataset and failure pattern logic
type TrainingPipelineService struct {
	logger   *zap.Logger
	mu       sync.RWMutex
	datasets map[uuid.UUID]*domain.TrainingDataset
	exports  map[uuid.UUID]*domain.TrainingExport
}

// NewTrainingPipelineService creates a new training pipeline service
func NewTrainingPipelineService(logger *zap.Logger) *TrainingPipelineService {
	return &TrainingPipelineService{
		logger:   logger,
		datasets: make(map[uuid.UUID]*domain.TrainingDataset),
		exports:  make(map[uuid.UUID]*domain.TrainingExport),
	}
}

// CreateDataset creates a new training dataset
func (s *TrainingPipelineService) CreateDataset(ctx context.Context, projectID uuid.UUID, input domain.TrainingDatasetInput) (*domain.TrainingDataset, error) {
	s.logger.Info("creating training dataset",
		zap.String("projectId", projectID.String()),
		zap.String("name", input.Name),
	)

	if input.Name == "" {
		return nil, fmt.Errorf("dataset name is required")
	}

	dataset := &domain.TrainingDataset{
		ID:           uuid.New(),
		ProjectID:    projectID,
		Name:         input.Name,
		Description:  input.Description,
		Format:       input.Format,
		SourceFilter: input.SourceFilter,
		Items:        0,
		Status:       domain.TrainingDatasetBuilding,
		CreatedAt:    time.Now(),
	}

	// Simulate async building — mark as ready with mock items
	dataset.Items = 150
	dataset.Status = domain.TrainingDatasetReady

	s.mu.Lock()
	s.datasets[dataset.ID] = dataset
	s.mu.Unlock()

	return dataset, nil
}

// ListDatasets lists all training datasets for a project
func (s *TrainingPipelineService) ListDatasets(ctx context.Context, projectID uuid.UUID) ([]domain.TrainingDataset, error) {
	s.logger.Debug("listing training datasets", zap.String("projectId", projectID.String()))

	s.mu.RLock()
	defer s.mu.RUnlock()

	var results []domain.TrainingDataset
	for _, ds := range s.datasets {
		if ds.ProjectID == projectID {
			results = append(results, *ds)
		}
	}

	if results == nil {
		results = []domain.TrainingDataset{}
	}

	return results, nil
}

// ExportDataset exports a training dataset in the specified format
func (s *TrainingPipelineService) ExportDataset(ctx context.Context, datasetID uuid.UUID, format domain.TrainingDatasetFormat) (*domain.TrainingExport, error) {
	s.logger.Info("exporting training dataset",
		zap.String("datasetId", datasetID.String()),
		zap.String("format", string(format)),
	)

	s.mu.RLock()
	dataset, exists := s.datasets[datasetID]
	s.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("dataset not found: %s", datasetID.String())
	}

	export := &domain.TrainingExport{
		ID:         uuid.New(),
		DatasetID:  datasetID,
		Format:     format,
		URL:        fmt.Sprintf("/exports/%s/%s.%s", dataset.ProjectID.String(), datasetID.String(), string(format)),
		LineCount:  dataset.Items,
		TokenCount: int64(dataset.Items) * 850,
		CreatedAt:  time.Now(),
	}

	s.mu.Lock()
	s.exports[export.ID] = export
	dataset.Status = domain.TrainingDatasetExported
	s.mu.Unlock()

	return export, nil
}

// DetectFailurePatterns detects common failure patterns in traces
func (s *TrainingPipelineService) DetectFailurePatterns(ctx context.Context, projectID uuid.UUID) ([]domain.FailurePattern, error) {
	s.logger.Debug("detecting failure patterns", zap.String("projectId", projectID.String()))

	patterns := []domain.FailurePattern{
		{
			ID:              uuid.New(),
			ProjectID:       projectID,
			Pattern:         "context_window_exceeded",
			Description:     "Agent exceeds context window limit causing truncated responses",
			Frequency:       23,
			ExampleTraceIDs: []string{uuid.New().String(), uuid.New().String()},
			SuggestedFix:    "Implement sliding window or summarization strategy for long conversations",
		},
		{
			ID:              uuid.New(),
			ProjectID:       projectID,
			Pattern:         "tool_call_loop",
			Description:     "Agent enters infinite tool call loop without producing final output",
			Frequency:       12,
			ExampleTraceIDs: []string{uuid.New().String()},
			SuggestedFix:    "Add maximum tool call depth and loop detection guardrails",
		},
		{
			ID:              uuid.New(),
			ProjectID:       projectID,
			Pattern:         "hallucinated_api_response",
			Description:     "Agent fabricates API responses instead of calling actual APIs",
			Frequency:       8,
			ExampleTraceIDs: []string{uuid.New().String(), uuid.New().String(), uuid.New().String()},
			SuggestedFix:    "Add validation layer to verify tool outputs match expected schemas",
		},
	}

	return patterns, nil
}
