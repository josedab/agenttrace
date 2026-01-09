package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	apperrors "github.com/agenttrace/agenttrace/api/internal/pkg/errors"
)

// DatasetRepository defines dataset repository operations
type DatasetRepository interface {
	Create(ctx context.Context, dataset *domain.Dataset) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Dataset, error)
	GetByName(ctx context.Context, projectID uuid.UUID, name string) (*domain.Dataset, error)
	Update(ctx context.Context, dataset *domain.Dataset) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, filter *domain.DatasetFilter, limit, offset int) (*domain.DatasetList, error)
	GetItemCount(ctx context.Context, datasetID uuid.UUID) (int64, error)
	GetRunCount(ctx context.Context, datasetID uuid.UUID) (int64, error)
	NameExists(ctx context.Context, projectID uuid.UUID, name string) (bool, error)

	// Item operations
	CreateItem(ctx context.Context, item *domain.DatasetItem) error
	GetItemByID(ctx context.Context, id uuid.UUID) (*domain.DatasetItem, error)
	UpdateItem(ctx context.Context, item *domain.DatasetItem) error
	DeleteItem(ctx context.Context, id uuid.UUID) error
	ListItems(ctx context.Context, filter *domain.DatasetItemFilter, limit, offset int) ([]domain.DatasetItem, int64, error)

	// Run operations
	CreateRun(ctx context.Context, run *domain.DatasetRun) error
	GetRunByID(ctx context.Context, id uuid.UUID) (*domain.DatasetRun, error)
	GetRunByName(ctx context.Context, datasetID uuid.UUID, name string) (*domain.DatasetRun, error)
	UpdateRun(ctx context.Context, run *domain.DatasetRun) error
	DeleteRun(ctx context.Context, id uuid.UUID) error
	ListRuns(ctx context.Context, datasetID uuid.UUID, limit, offset int) ([]domain.DatasetRun, int64, error)
	GetRunItemCount(ctx context.Context, runID uuid.UUID) (int64, error)

	// Run item operations
	CreateRunItem(ctx context.Context, item *domain.DatasetRunItem) error
	GetRunItemByID(ctx context.Context, id uuid.UUID) (*domain.DatasetRunItem, error)
	ListRunItems(ctx context.Context, runID uuid.UUID, limit, offset int) ([]domain.DatasetRunItem, int64, error)
	GetRunItemByDatasetItem(ctx context.Context, runID, itemID uuid.UUID) (*domain.DatasetRunItem, error)
}

// DatasetService handles dataset operations
type DatasetService struct {
	datasetRepo DatasetRepository
	traceRepo   TraceRepository
	scoreRepo   ScoreRepository
}

// NewDatasetService creates a new dataset service
func NewDatasetService(
	datasetRepo DatasetRepository,
	traceRepo TraceRepository,
	scoreRepo ScoreRepository,
) *DatasetService {
	return &DatasetService{
		datasetRepo: datasetRepo,
		traceRepo:   traceRepo,
		scoreRepo:   scoreRepo,
	}
}

// Create creates a new dataset
func (s *DatasetService) Create(ctx context.Context, projectID uuid.UUID, input *domain.DatasetInput) (*domain.Dataset, error) {
	// Check if name exists
	exists, err := s.datasetRepo.NameExists(ctx, projectID, input.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to check name: %w", err)
	}
	if exists {
		return nil, apperrors.Validation("dataset name already exists")
	}

	now := time.Now()

	var metadata string
	if input.Metadata != nil {
		metadataBytes, _ := json.Marshal(input.Metadata)
		metadata = string(metadataBytes)
	}

	dataset := &domain.Dataset{
		ID:          uuid.New(),
		ProjectID:   projectID,
		Name:        input.Name,
		Description: input.Description,
		Metadata:    metadata,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.datasetRepo.Create(ctx, dataset); err != nil {
		return nil, fmt.Errorf("failed to create dataset: %w", err)
	}

	return dataset, nil
}

// Get retrieves a dataset by ID
func (s *DatasetService) Get(ctx context.Context, id uuid.UUID) (*domain.Dataset, error) {
	dataset, err := s.datasetRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Load counts
	itemCount, _ := s.datasetRepo.GetItemCount(ctx, id)
	runCount, _ := s.datasetRepo.GetRunCount(ctx, id)
	dataset.ItemCount = itemCount
	dataset.RunCount = runCount

	return dataset, nil
}

// GetByName retrieves a dataset by name
func (s *DatasetService) GetByName(ctx context.Context, projectID uuid.UUID, name string) (*domain.Dataset, error) {
	dataset, err := s.datasetRepo.GetByName(ctx, projectID, name)
	if err != nil {
		return nil, err
	}

	// Load counts
	itemCount, _ := s.datasetRepo.GetItemCount(ctx, dataset.ID)
	runCount, _ := s.datasetRepo.GetRunCount(ctx, dataset.ID)
	dataset.ItemCount = itemCount
	dataset.RunCount = runCount

	return dataset, nil
}

// Update updates a dataset
func (s *DatasetService) Update(ctx context.Context, id uuid.UUID, input *domain.DatasetInput) (*domain.Dataset, error) {
	dataset, err := s.datasetRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if input.Name != "" && input.Name != dataset.Name {
		exists, err := s.datasetRepo.NameExists(ctx, dataset.ProjectID, input.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to check name: %w", err)
		}
		if exists {
			return nil, apperrors.Validation("dataset name already exists")
		}
		dataset.Name = input.Name
	}

	if input.Description != "" {
		dataset.Description = input.Description
	}
	if input.Metadata != nil {
		metadataBytes, _ := json.Marshal(input.Metadata)
		dataset.Metadata = string(metadataBytes)
	}

	dataset.UpdatedAt = time.Now()

	if err := s.datasetRepo.Update(ctx, dataset); err != nil {
		return nil, fmt.Errorf("failed to update dataset: %w", err)
	}

	return dataset, nil
}

// Delete deletes a dataset
func (s *DatasetService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.datasetRepo.Delete(ctx, id)
}

// List retrieves datasets with filtering
func (s *DatasetService) List(ctx context.Context, filter *domain.DatasetFilter, limit, offset int) (*domain.DatasetList, error) {
	list, err := s.datasetRepo.List(ctx, filter, limit, offset)
	if err != nil {
		return nil, err
	}

	// Load counts for each dataset
	for i := range list.Datasets {
		itemCount, _ := s.datasetRepo.GetItemCount(ctx, list.Datasets[i].ID)
		runCount, _ := s.datasetRepo.GetRunCount(ctx, list.Datasets[i].ID)
		list.Datasets[i].ItemCount = itemCount
		list.Datasets[i].RunCount = runCount
	}

	return list, nil
}

// AddItem adds an item to a dataset
func (s *DatasetService) AddItem(ctx context.Context, datasetID uuid.UUID, input *domain.DatasetItemInput) (*domain.DatasetItem, error) {
	// Verify dataset exists
	_, err := s.datasetRepo.GetByID(ctx, datasetID)
	if err != nil {
		return nil, err
	}

	now := time.Now()

	var inputStr string
	if input.Input != nil {
		inputBytes, _ := json.Marshal(input.Input)
		inputStr = string(inputBytes)
	}

	var expectedOutputStr *string
	if input.ExpectedOutput != nil {
		outputBytes, _ := json.Marshal(input.ExpectedOutput)
		s := string(outputBytes)
		expectedOutputStr = &s
	}

	var metadata string
	if input.Metadata != nil {
		metadataBytes, _ := json.Marshal(input.Metadata)
		metadata = string(metadataBytes)
	}

	item := &domain.DatasetItem{
		ID:                  uuid.New(),
		DatasetID:           datasetID,
		Input:               inputStr,
		ExpectedOutput:      expectedOutputStr,
		Metadata:            metadata,
		SourceTraceID:       input.SourceTraceID,
		SourceObservationID: input.SourceObservationID,
		Status:              domain.DatasetItemStatusActive,
		CreatedAt:           now,
		UpdatedAt:           now,
	}

	if err := s.datasetRepo.CreateItem(ctx, item); err != nil {
		return nil, fmt.Errorf("failed to create item: %w", err)
	}

	return item, nil
}

// AddItemFromTrace creates a dataset item from a trace/observation
func (s *DatasetService) AddItemFromTrace(ctx context.Context, datasetID uuid.UUID, projectID uuid.UUID, traceID string, observationID *string) (*domain.DatasetItem, error) {
	trace, err := s.traceRepo.GetByID(ctx, projectID, traceID)
	if err != nil {
		return nil, err
	}

	input := &domain.DatasetItemInput{
		SourceTraceID: &traceID,
	}

	// Use trace input/output
	if trace.Input != "" {
		var traceInput interface{}
		if err := json.Unmarshal([]byte(trace.Input), &traceInput); err == nil {
			input.Input = traceInput
		}
	}
	if trace.Output != "" {
		var traceOutput interface{}
		if err := json.Unmarshal([]byte(trace.Output), &traceOutput); err == nil {
			input.ExpectedOutput = traceOutput
		}
	}

	if observationID != nil {
		input.SourceObservationID = observationID
	}

	return s.AddItem(ctx, datasetID, input)
}

// UpdateItem updates a dataset item
func (s *DatasetService) UpdateItem(ctx context.Context, itemID uuid.UUID, input *domain.DatasetItemUpdateInput) (*domain.DatasetItem, error) {
	item, err := s.datasetRepo.GetItemByID(ctx, itemID)
	if err != nil {
		return nil, err
	}

	if input.Input != nil {
		inputBytes, _ := json.Marshal(input.Input)
		item.Input = string(inputBytes)
	}
	if input.ExpectedOutput != nil {
		outputBytes, _ := json.Marshal(input.ExpectedOutput)
		s := string(outputBytes)
		item.ExpectedOutput = &s
	}
	if input.Metadata != nil {
		metadataBytes, _ := json.Marshal(input.Metadata)
		item.Metadata = string(metadataBytes)
	}
	if input.Status != nil {
		item.Status = *input.Status
	}

	item.UpdatedAt = time.Now()

	if err := s.datasetRepo.UpdateItem(ctx, item); err != nil {
		return nil, fmt.Errorf("failed to update item: %w", err)
	}

	return item, nil
}

// DeleteItem deletes a dataset item
func (s *DatasetService) DeleteItem(ctx context.Context, itemID uuid.UUID) error {
	return s.datasetRepo.DeleteItem(ctx, itemID)
}

// ListItems retrieves dataset items
func (s *DatasetService) ListItems(ctx context.Context, filter *domain.DatasetItemFilter, limit, offset int) ([]domain.DatasetItem, int64, error) {
	return s.datasetRepo.ListItems(ctx, filter, limit, offset)
}

// CreateRun creates a new dataset run
func (s *DatasetService) CreateRun(ctx context.Context, datasetID uuid.UUID, input *domain.DatasetRunInput) (*domain.DatasetRun, error) {
	// Verify dataset exists
	_, err := s.datasetRepo.GetByID(ctx, datasetID)
	if err != nil {
		return nil, err
	}

	now := time.Now()

	var metadata string
	if input.Metadata != nil {
		metadataBytes, _ := json.Marshal(input.Metadata)
		metadata = string(metadataBytes)
	}

	run := &domain.DatasetRun{
		ID:          uuid.New(),
		DatasetID:   datasetID,
		Name:        input.Name,
		Description: input.Description,
		Metadata:    metadata,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.datasetRepo.CreateRun(ctx, run); err != nil {
		return nil, fmt.Errorf("failed to create run: %w", err)
	}

	return run, nil
}

// GetRun retrieves a dataset run by ID
func (s *DatasetService) GetRun(ctx context.Context, id uuid.UUID) (*domain.DatasetRun, error) {
	run, err := s.datasetRepo.GetRunByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Load item count
	itemCount, _ := s.datasetRepo.GetRunItemCount(ctx, id)
	run.ItemCount = itemCount

	return run, nil
}

// GetRunByName retrieves a dataset run by name
func (s *DatasetService) GetRunByName(ctx context.Context, datasetID uuid.UUID, name string) (*domain.DatasetRun, error) {
	run, err := s.datasetRepo.GetRunByName(ctx, datasetID, name)
	if err != nil {
		return nil, err
	}

	// Load item count
	itemCount, _ := s.datasetRepo.GetRunItemCount(ctx, run.ID)
	run.ItemCount = itemCount

	return run, nil
}

// ListRuns retrieves dataset runs
func (s *DatasetService) ListRuns(ctx context.Context, datasetID uuid.UUID, limit, offset int) ([]domain.DatasetRun, int64, error) {
	runs, totalCount, err := s.datasetRepo.ListRuns(ctx, datasetID, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	// Load counts for each run
	for i := range runs {
		itemCount, _ := s.datasetRepo.GetRunItemCount(ctx, runs[i].ID)
		runs[i].ItemCount = itemCount
	}

	return runs, totalCount, nil
}

// AddRunItem links a trace to a dataset run
func (s *DatasetService) AddRunItem(ctx context.Context, runID uuid.UUID, input *domain.DatasetRunItemInput) (*domain.DatasetRunItem, error) {
	// Verify run exists
	_, err := s.datasetRepo.GetRunByID(ctx, runID)
	if err != nil {
		return nil, err
	}

	// Parse dataset item ID
	itemID, err := uuid.Parse(input.DatasetItemID)
	if err != nil {
		return nil, apperrors.Validation("invalid dataset item ID")
	}

	// Verify dataset item exists
	_, err = s.datasetRepo.GetItemByID(ctx, itemID)
	if err != nil {
		return nil, err
	}

	now := time.Now()

	runItem := &domain.DatasetRunItem{
		ID:            uuid.New(),
		DatasetRunID:  runID,
		DatasetItemID: itemID,
		TraceID:       input.TraceID,
		ObservationID: input.ObservationID,
		CreatedAt:     now,
	}

	if err := s.datasetRepo.CreateRunItem(ctx, runItem); err != nil {
		return nil, fmt.Errorf("failed to create run item: %w", err)
	}

	return runItem, nil
}

// AddRunItemsBatch adds multiple items to a dataset run
func (s *DatasetService) AddRunItemsBatch(ctx context.Context, runID uuid.UUID, inputs []*domain.DatasetRunItemInput) ([]*domain.DatasetRunItem, error) {
	// Verify run exists
	_, err := s.datasetRepo.GetRunByID(ctx, runID)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	results := make([]*domain.DatasetRunItem, 0, len(inputs))

	for _, input := range inputs {
		// Parse dataset item ID
		itemID, err := uuid.Parse(input.DatasetItemID)
		if err != nil {
			return nil, apperrors.Validation(fmt.Sprintf("invalid dataset item ID: %s", input.DatasetItemID))
		}

		// Verify dataset item exists
		_, err = s.datasetRepo.GetItemByID(ctx, itemID)
		if err != nil {
			return nil, fmt.Errorf("dataset item not found: %s", input.DatasetItemID)
		}

		runItem := &domain.DatasetRunItem{
			ID:            uuid.New(),
			DatasetRunID:  runID,
			DatasetItemID: itemID,
			TraceID:       input.TraceID,
			ObservationID: input.ObservationID,
			CreatedAt:     now,
		}

		if err := s.datasetRepo.CreateRunItem(ctx, runItem); err != nil {
			return nil, fmt.Errorf("failed to create run item: %w", err)
		}

		results = append(results, runItem)
	}

	return results, nil
}

// GetRunResults retrieves run results with comparison
func (s *DatasetService) GetRunResults(ctx context.Context, projectID uuid.UUID, runID uuid.UUID) (*domain.DatasetRunResults, error) {
	run, err := s.datasetRepo.GetRunByID(ctx, runID)
	if err != nil {
		return nil, err
	}

	runItems, _, err := s.datasetRepo.ListRunItems(ctx, runID, 1000, 0)
	if err != nil {
		return nil, err
	}

	results := &domain.DatasetRunResults{
		Run:   run,
		Items: make([]domain.DatasetRunItemWithDiff, 0, len(runItems)),
	}

	for _, runItem := range runItems {
		itemWithDiff := domain.DatasetRunItemWithDiff{
			DatasetRunItem: runItem,
		}

		// Get expected output from dataset item
		if runItem.DatasetItem != nil && runItem.DatasetItem.ExpectedOutput != nil {
			itemWithDiff.ExpectedOutput = runItem.DatasetItem.ExpectedOutput
		}

		// Get actual output from trace
		trace, err := s.traceRepo.GetByID(ctx, projectID, runItem.TraceID)
		if err == nil && trace.Output != "" {
			itemWithDiff.ActualOutput = &trace.Output
		}

		// Load scores for the trace
		scores, err := s.scoreRepo.GetByTraceID(ctx, projectID, runItem.TraceID)
		if err == nil {
			runItem.Scores = scores
		}

		results.Items = append(results.Items, itemWithDiff)
	}

	return results, nil
}
