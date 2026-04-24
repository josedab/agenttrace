package service

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	apperrors "github.com/agenttrace/agenttrace/api/internal/pkg/errors"
)

// forkDatasetRepositoryStub implements only the operations a dataset fork uses.
// The embedded interface makes any unexpected call fail loudly instead of
// silently succeeding.
type forkDatasetRepositoryStub struct {
	DatasetRepository
	datasets      map[uuid.UUID]*domain.Dataset
	items         map[uuid.UUID][]domain.DatasetItem
	failItemAfter int
	failDelete    bool
	createdItems  int
	deleted       []uuid.UUID
}

func newForkDatasetRepositoryStub() *forkDatasetRepositoryStub {
	return &forkDatasetRepositoryStub{
		datasets:      map[uuid.UUID]*domain.Dataset{},
		items:         map[uuid.UUID][]domain.DatasetItem{},
		failItemAfter: -1,
	}
}

func (r *forkDatasetRepositoryStub) Create(_ context.Context, dataset *domain.Dataset) error {
	stored := *dataset
	r.datasets[dataset.ID] = &stored
	return nil
}

func (r *forkDatasetRepositoryStub) GetByID(
	_ context.Context,
	id uuid.UUID,
) (*domain.Dataset, error) {
	dataset, ok := r.datasets[id]
	if !ok {
		return nil, apperrors.NotFound("dataset")
	}
	copied := *dataset
	return &copied, nil
}

func (r *forkDatasetRepositoryStub) GetItemCount(
	_ context.Context,
	datasetID uuid.UUID,
) (int64, error) {
	return int64(len(r.items[datasetID])), nil
}

func (r *forkDatasetRepositoryStub) GetRunCount(
	_ context.Context,
	_ uuid.UUID,
) (int64, error) {
	return 0, nil
}

func (r *forkDatasetRepositoryStub) NameExists(
	_ context.Context,
	_ uuid.UUID,
	_ string,
) (bool, error) {
	return false, nil
}

func (r *forkDatasetRepositoryStub) CreateItem(
	_ context.Context,
	item *domain.DatasetItem,
) error {
	if r.failItemAfter >= 0 && r.createdItems >= r.failItemAfter {
		return errors.New("storage rejected the item")
	}
	r.createdItems++
	r.items[item.DatasetID] = append(r.items[item.DatasetID], *item)
	return nil
}

func (r *forkDatasetRepositoryStub) Delete(_ context.Context, id uuid.UUID) error {
	if r.failDelete {
		return errors.New("delete failed")
	}
	r.deleted = append(r.deleted, id)
	delete(r.datasets, id)
	delete(r.items, id)
	return nil
}

func datasetForkManifest(t *testing.T, itemCount int) json.RawMessage {
	t.Helper()
	items := make([]domain.DatasetItem, 0, itemCount)
	for index := range itemCount {
		expected := `{"answer":"b"}`
		items = append(items, domain.DatasetItem{
			ID:             uuid.New(),
			Input:          `{"question":"q` + string(rune('0'+index)) + `"}`,
			ExpectedOutput: &expected,
		})
	}
	manifest, err := json.Marshal(datasetAssetManifest{
		Dataset: domain.Dataset{Name: "source", Description: "source dataset"},
		Items:   items,
	})
	require.NoError(t, err)
	return manifest
}

func TestForkDatasetMaterializesEveryItemIntoTargetProject(t *testing.T) {
	repository := newForkDatasetRepositoryStub()
	manager := NewDefaultEvalHubAssetManager(
		NewDatasetService(repository, nil, nil),
		nil,
		nil,
		nil,
	)
	projectID := uuid.New()

	resourceID, manifest, err := manager.Fork(
		context.Background(),
		projectID,
		uuid.New(),
		domain.EvalHubDataset,
		"forked dataset",
		datasetForkManifest(t, 3),
	)
	require.NoError(t, err)

	created, ok := repository.datasets[resourceID]
	require.True(t, ok)
	assert.Equal(t, projectID, created.ProjectID)
	assert.Len(t, repository.items[resourceID], 3)

	var forked datasetAssetManifest
	require.NoError(t, json.Unmarshal(manifest, &forked))
	assert.Len(t, forked.Items, 3)
	assert.Equal(t, resourceID, forked.Dataset.ID)
}

func TestForkDatasetRollsBackPartialMaterialization(t *testing.T) {
	repository := newForkDatasetRepositoryStub()
	repository.failItemAfter = 2
	manager := NewDefaultEvalHubAssetManager(
		NewDatasetService(repository, nil, nil),
		nil,
		nil,
		nil,
	)
	projectID := uuid.New()

	_, _, err := manager.Fork(
		context.Background(),
		projectID,
		uuid.New(),
		domain.EvalHubDataset,
		"forked dataset",
		datasetForkManifest(t, 4),
	)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "materialize dataset item")
	// No half-populated dataset is left behind for the caller to discover later.
	assert.Empty(t, repository.datasets)
	assert.Len(t, repository.deleted, 1)
}

func TestForkDatasetReportsFailedRollback(t *testing.T) {
	repository := newForkDatasetRepositoryStub()
	repository.failItemAfter = 1
	repository.failDelete = true
	manager := NewDefaultEvalHubAssetManager(
		NewDatasetService(repository, nil, nil),
		nil,
		nil,
		nil,
	)

	_, _, err := manager.Fork(
		context.Background(),
		uuid.New(),
		uuid.New(),
		domain.EvalHubDataset,
		"forked dataset",
		datasetForkManifest(t, 3),
	)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "could not be removed")
}

func TestForkDatasetValidatesEveryItemBeforeCreatingAnything(t *testing.T) {
	repository := newForkDatasetRepositoryStub()
	manager := NewDefaultEvalHubAssetManager(
		NewDatasetService(repository, nil, nil),
		nil,
		nil,
		nil,
	)

	manifest, err := json.Marshal(datasetAssetManifest{
		Dataset: domain.Dataset{Name: "source"},
		Items: []domain.DatasetItem{
			{ID: uuid.New(), Input: `{"question":"ok"}`},
			{ID: uuid.New(), Input: ""},
		},
	})
	require.NoError(t, err)

	_, _, forkErr := manager.Fork(
		context.Background(),
		uuid.New(),
		uuid.New(),
		domain.EvalHubDataset,
		"forked dataset",
		manifest,
	)

	require.Error(t, forkErr)
	assert.True(t, apperrors.IsValidation(forkErr))
	assert.Contains(t, forkErr.Error(), "item 1")
	assert.Empty(t, repository.datasets)
	assert.Zero(t, repository.createdItems)
}
