package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/pkg/database"
	apperrors "github.com/agenttrace/agenttrace/api/internal/pkg/errors"
)

// createTestDataset creates a dataset with test data
func createTestDataset(name string, projectID uuid.UUID) *domain.Dataset {
	now := time.Now()
	return &domain.Dataset{
		ID:          uuid.New(),
		ProjectID:   projectID,
		Name:        name,
		Description: "Test dataset description",
		Metadata:    `{"source": "test"}`,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// createTestDatasetItem creates a dataset item with test data
func createTestDatasetItem(datasetID uuid.UUID) *domain.DatasetItem {
	now := time.Now()
	expectedOutput := `{"response": "expected output"}`
	return &domain.DatasetItem{
		ID:             uuid.New(),
		DatasetID:      datasetID,
		Input:          `{"prompt": "test input"}`,
		ExpectedOutput: &expectedOutput,
		Metadata:       `{"source": "test"}`,
		Status:         "active",
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

// createTestDatasetRun creates a dataset run with test data
func createTestDatasetRun(datasetID uuid.UUID, name string) *domain.DatasetRun {
	now := time.Now()
	return &domain.DatasetRun{
		ID:          uuid.New(),
		DatasetID:   datasetID,
		Name:        name,
		Description: "Test run description",
		Metadata:    `{"model": "gpt-4"}`,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

func cleanupDatasets(t *testing.T, db *database.PostgresDB, names ...string) {
	ctx := context.Background()
	for _, name := range names {
		// Clean up in dependency order
		_, _ = db.Pool.Exec(ctx, "DELETE FROM dataset_run_items WHERE dataset_run_id IN (SELECT id FROM dataset_runs WHERE dataset_id IN (SELECT id FROM datasets WHERE name = $1))", name)
		_, _ = db.Pool.Exec(ctx, "DELETE FROM dataset_runs WHERE dataset_id IN (SELECT id FROM datasets WHERE name = $1)", name)
		_, _ = db.Pool.Exec(ctx, "DELETE FROM dataset_items WHERE dataset_id IN (SELECT id FROM datasets WHERE name = $1)", name)
		_, _ = db.Pool.Exec(ctx, "DELETE FROM datasets WHERE name = $1", name)
	}
}

func TestDatasetRepository_Create(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	orgRepo := NewOrgRepository(db)
	projectRepo := NewProjectRepository(db)
	datasetRepo := NewDatasetRepository(db)
	ctx := context.Background()

	orgName := "Test Org for Dataset Create"
	projectName := "Test Project for Dataset Create"
	datasetName := "Test Dataset Create"

	cleanupDatasets(t, db, datasetName)
	cleanupProjects(t, db, projectName)
	cleanupOrgs(t, db, orgName)
	defer cleanupDatasets(t, db, datasetName)
	defer cleanupProjects(t, db, projectName)
	defer cleanupOrgs(t, db, orgName)

	org := createTestOrg(orgName)
	err := orgRepo.Create(ctx, org)
	require.NoError(t, err)

	project := createTestProject(projectName, org.ID)
	err = projectRepo.Create(ctx, project)
	require.NoError(t, err)

	dataset := createTestDataset(datasetName, project.ID)
	err = datasetRepo.Create(ctx, dataset)
	require.NoError(t, err)

	// Verify by fetching
	fetched, err := datasetRepo.GetByID(ctx, dataset.ID)
	require.NoError(t, err)
	assert.Equal(t, dataset.ID, fetched.ID)
	assert.Equal(t, dataset.Name, fetched.Name)
}

func TestDatasetRepository_GetByID(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	orgRepo := NewOrgRepository(db)
	projectRepo := NewProjectRepository(db)
	datasetRepo := NewDatasetRepository(db)
	ctx := context.Background()

	orgName := "Test Org for Dataset GetByID"
	projectName := "Test Project for Dataset GetByID"
	datasetName := "Test Dataset GetByID"

	cleanupDatasets(t, db, datasetName)
	cleanupProjects(t, db, projectName)
	cleanupOrgs(t, db, orgName)
	defer cleanupDatasets(t, db, datasetName)
	defer cleanupProjects(t, db, projectName)
	defer cleanupOrgs(t, db, orgName)

	org := createTestOrg(orgName)
	err := orgRepo.Create(ctx, org)
	require.NoError(t, err)

	project := createTestProject(projectName, org.ID)
	err = projectRepo.Create(ctx, project)
	require.NoError(t, err)

	dataset := createTestDataset(datasetName, project.ID)
	err = datasetRepo.Create(ctx, dataset)
	require.NoError(t, err)

	t.Run("existing dataset", func(t *testing.T) {
		fetched, err := datasetRepo.GetByID(ctx, dataset.ID)
		require.NoError(t, err)
		assert.Equal(t, dataset.ID, fetched.ID)
	})

	t.Run("non-existent dataset", func(t *testing.T) {
		_, err := datasetRepo.GetByID(ctx, uuid.New())
		assert.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

func TestDatasetRepository_GetByName(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	orgRepo := NewOrgRepository(db)
	projectRepo := NewProjectRepository(db)
	datasetRepo := NewDatasetRepository(db)
	ctx := context.Background()

	orgName := "Test Org for Dataset GetByName"
	projectName := "Test Project for Dataset GetByName"
	datasetName := "Test Dataset GetByName"

	cleanupDatasets(t, db, datasetName)
	cleanupProjects(t, db, projectName)
	cleanupOrgs(t, db, orgName)
	defer cleanupDatasets(t, db, datasetName)
	defer cleanupProjects(t, db, projectName)
	defer cleanupOrgs(t, db, orgName)

	org := createTestOrg(orgName)
	err := orgRepo.Create(ctx, org)
	require.NoError(t, err)

	project := createTestProject(projectName, org.ID)
	err = projectRepo.Create(ctx, project)
	require.NoError(t, err)

	dataset := createTestDataset(datasetName, project.ID)
	err = datasetRepo.Create(ctx, dataset)
	require.NoError(t, err)

	t.Run("existing name", func(t *testing.T) {
		fetched, err := datasetRepo.GetByName(ctx, project.ID, datasetName)
		require.NoError(t, err)
		assert.Equal(t, dataset.ID, fetched.ID)
	})

	t.Run("non-existent name", func(t *testing.T) {
		_, err := datasetRepo.GetByName(ctx, project.ID, "nonexistent")
		assert.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

func TestDatasetRepository_Update(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	orgRepo := NewOrgRepository(db)
	projectRepo := NewProjectRepository(db)
	datasetRepo := NewDatasetRepository(db)
	ctx := context.Background()

	orgName := "Test Org for Dataset Update"
	projectName := "Test Project for Dataset Update"
	datasetName := "Test Dataset Update"

	cleanupDatasets(t, db, datasetName)
	cleanupProjects(t, db, projectName)
	cleanupOrgs(t, db, orgName)
	defer cleanupDatasets(t, db, datasetName)
	defer cleanupProjects(t, db, projectName)
	defer cleanupOrgs(t, db, orgName)

	org := createTestOrg(orgName)
	err := orgRepo.Create(ctx, org)
	require.NoError(t, err)

	project := createTestProject(projectName, org.ID)
	err = projectRepo.Create(ctx, project)
	require.NoError(t, err)

	dataset := createTestDataset(datasetName, project.ID)
	err = datasetRepo.Create(ctx, dataset)
	require.NoError(t, err)

	// Update
	dataset.Description = "Updated description"
	err = datasetRepo.Update(ctx, dataset)
	require.NoError(t, err)

	// Verify
	fetched, err := datasetRepo.GetByID(ctx, dataset.ID)
	require.NoError(t, err)
	assert.Equal(t, "Updated description", fetched.Description)
}

func TestDatasetRepository_Delete(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	orgRepo := NewOrgRepository(db)
	projectRepo := NewProjectRepository(db)
	datasetRepo := NewDatasetRepository(db)
	ctx := context.Background()

	orgName := "Test Org for Dataset Delete"
	projectName := "Test Project for Dataset Delete"
	datasetName := "Test Dataset Delete"

	cleanupDatasets(t, db, datasetName)
	cleanupProjects(t, db, projectName)
	cleanupOrgs(t, db, orgName)
	defer cleanupProjects(t, db, projectName)
	defer cleanupOrgs(t, db, orgName)

	org := createTestOrg(orgName)
	err := orgRepo.Create(ctx, org)
	require.NoError(t, err)

	project := createTestProject(projectName, org.ID)
	err = projectRepo.Create(ctx, project)
	require.NoError(t, err)

	dataset := createTestDataset(datasetName, project.ID)
	err = datasetRepo.Create(ctx, dataset)
	require.NoError(t, err)

	// Delete
	err = datasetRepo.Delete(ctx, dataset.ID)
	require.NoError(t, err)

	// Verify deleted
	_, err = datasetRepo.GetByID(ctx, dataset.ID)
	assert.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err))
}

func TestDatasetRepository_List(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	orgRepo := NewOrgRepository(db)
	projectRepo := NewProjectRepository(db)
	datasetRepo := NewDatasetRepository(db)
	ctx := context.Background()

	orgName := "Test Org for Dataset List"
	projectName := "Test Project for Dataset List"
	datasetName1 := "Test Dataset List A"
	datasetName2 := "Test Dataset List B"

	cleanupDatasets(t, db, datasetName1, datasetName2)
	cleanupProjects(t, db, projectName)
	cleanupOrgs(t, db, orgName)
	defer cleanupDatasets(t, db, datasetName1, datasetName2)
	defer cleanupProjects(t, db, projectName)
	defer cleanupOrgs(t, db, orgName)

	org := createTestOrg(orgName)
	err := orgRepo.Create(ctx, org)
	require.NoError(t, err)

	project := createTestProject(projectName, org.ID)
	err = projectRepo.Create(ctx, project)
	require.NoError(t, err)

	for _, name := range []string{datasetName1, datasetName2} {
		dataset := createTestDataset(name, project.ID)
		err = datasetRepo.Create(ctx, dataset)
		require.NoError(t, err)
	}

	t.Run("basic list", func(t *testing.T) {
		filter := &domain.DatasetFilter{ProjectID: project.ID}
		list, err := datasetRepo.List(ctx, filter, 10, 0)
		require.NoError(t, err)
		assert.Equal(t, int64(2), list.TotalCount)
	})

	t.Run("filter by name", func(t *testing.T) {
		nameFilter := "List A"
		filter := &domain.DatasetFilter{
			ProjectID: project.ID,
			Name:      &nameFilter,
		}
		list, err := datasetRepo.List(ctx, filter, 10, 0)
		require.NoError(t, err)
		assert.Equal(t, int64(1), list.TotalCount)
	})
}

func TestDatasetRepository_Items(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	orgRepo := NewOrgRepository(db)
	projectRepo := NewProjectRepository(db)
	datasetRepo := NewDatasetRepository(db)
	ctx := context.Background()

	orgName := "Test Org for Dataset Items"
	projectName := "Test Project for Dataset Items"
	datasetName := "Test Dataset Items"

	cleanupDatasets(t, db, datasetName)
	cleanupProjects(t, db, projectName)
	cleanupOrgs(t, db, orgName)
	defer cleanupDatasets(t, db, datasetName)
	defer cleanupProjects(t, db, projectName)
	defer cleanupOrgs(t, db, orgName)

	org := createTestOrg(orgName)
	err := orgRepo.Create(ctx, org)
	require.NoError(t, err)

	project := createTestProject(projectName, org.ID)
	err = projectRepo.Create(ctx, project)
	require.NoError(t, err)

	dataset := createTestDataset(datasetName, project.ID)
	err = datasetRepo.Create(ctx, dataset)
	require.NoError(t, err)

	var itemID uuid.UUID

	t.Run("create item", func(t *testing.T) {
		item := createTestDatasetItem(dataset.ID)
		itemID = item.ID
		err := datasetRepo.CreateItem(ctx, item)
		require.NoError(t, err)
	})

	t.Run("get item by ID", func(t *testing.T) {
		fetched, err := datasetRepo.GetItemByID(ctx, itemID)
		require.NoError(t, err)
		assert.Equal(t, itemID, fetched.ID)
		assert.Equal(t, dataset.ID, fetched.DatasetID)
	})

	t.Run("get item count", func(t *testing.T) {
		count, err := datasetRepo.GetItemCount(ctx, dataset.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(1), count)
	})

	t.Run("update item", func(t *testing.T) {
		item, err := datasetRepo.GetItemByID(ctx, itemID)
		require.NoError(t, err)

		item.Status = "archived"
		err = datasetRepo.UpdateItem(ctx, item)
		require.NoError(t, err)

		fetched, err := datasetRepo.GetItemByID(ctx, itemID)
		require.NoError(t, err)
		assert.Equal(t, "archived", fetched.Status)
	})

	t.Run("list items", func(t *testing.T) {
		// Create another item
		item2 := createTestDatasetItem(dataset.ID)
		err := datasetRepo.CreateItem(ctx, item2)
		require.NoError(t, err)

		filter := &domain.DatasetItemFilter{DatasetID: dataset.ID}
		items, count, err := datasetRepo.ListItems(ctx, filter, 10, 0)
		require.NoError(t, err)
		assert.Len(t, items, 2)
		assert.Equal(t, int64(2), count)
	})

	t.Run("delete item", func(t *testing.T) {
		err := datasetRepo.DeleteItem(ctx, itemID)
		require.NoError(t, err)

		_, err = datasetRepo.GetItemByID(ctx, itemID)
		assert.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

func TestDatasetRepository_Runs(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	orgRepo := NewOrgRepository(db)
	projectRepo := NewProjectRepository(db)
	datasetRepo := NewDatasetRepository(db)
	ctx := context.Background()

	orgName := "Test Org for Dataset Runs"
	projectName := "Test Project for Dataset Runs"
	datasetName := "Test Dataset Runs"

	cleanupDatasets(t, db, datasetName)
	cleanupProjects(t, db, projectName)
	cleanupOrgs(t, db, orgName)
	defer cleanupDatasets(t, db, datasetName)
	defer cleanupProjects(t, db, projectName)
	defer cleanupOrgs(t, db, orgName)

	org := createTestOrg(orgName)
	err := orgRepo.Create(ctx, org)
	require.NoError(t, err)

	project := createTestProject(projectName, org.ID)
	err = projectRepo.Create(ctx, project)
	require.NoError(t, err)

	dataset := createTestDataset(datasetName, project.ID)
	err = datasetRepo.Create(ctx, dataset)
	require.NoError(t, err)

	var runID uuid.UUID

	t.Run("create run", func(t *testing.T) {
		run := createTestDatasetRun(dataset.ID, "test-run-1")
		runID = run.ID
		err := datasetRepo.CreateRun(ctx, run)
		require.NoError(t, err)
	})

	t.Run("get run by ID", func(t *testing.T) {
		fetched, err := datasetRepo.GetRunByID(ctx, runID)
		require.NoError(t, err)
		assert.Equal(t, runID, fetched.ID)
	})

	t.Run("get run by name", func(t *testing.T) {
		fetched, err := datasetRepo.GetRunByName(ctx, dataset.ID, "test-run-1")
		require.NoError(t, err)
		assert.Equal(t, runID, fetched.ID)
	})

	t.Run("get run count", func(t *testing.T) {
		count, err := datasetRepo.GetRunCount(ctx, dataset.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(1), count)
	})

	t.Run("update run", func(t *testing.T) {
		run, err := datasetRepo.GetRunByID(ctx, runID)
		require.NoError(t, err)

		run.Description = "Updated description"
		err = datasetRepo.UpdateRun(ctx, run)
		require.NoError(t, err)

		fetched, err := datasetRepo.GetRunByID(ctx, runID)
		require.NoError(t, err)
		assert.Equal(t, "Updated description", fetched.Description)
	})

	t.Run("list runs", func(t *testing.T) {
		// Create another run
		run2 := createTestDatasetRun(dataset.ID, "test-run-2")
		err := datasetRepo.CreateRun(ctx, run2)
		require.NoError(t, err)

		runs, count, err := datasetRepo.ListRuns(ctx, dataset.ID, 10, 0)
		require.NoError(t, err)
		assert.Len(t, runs, 2)
		assert.Equal(t, int64(2), count)
	})

	t.Run("delete run", func(t *testing.T) {
		err := datasetRepo.DeleteRun(ctx, runID)
		require.NoError(t, err)

		_, err = datasetRepo.GetRunByID(ctx, runID)
		assert.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

func TestDatasetRepository_NameExists(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	orgRepo := NewOrgRepository(db)
	projectRepo := NewProjectRepository(db)
	datasetRepo := NewDatasetRepository(db)
	ctx := context.Background()

	orgName := "Test Org for Dataset NameExists"
	projectName := "Test Project for Dataset NameExists"
	datasetName := "Test Dataset NameExists"

	cleanupDatasets(t, db, datasetName)
	cleanupProjects(t, db, projectName)
	cleanupOrgs(t, db, orgName)
	defer cleanupDatasets(t, db, datasetName)
	defer cleanupProjects(t, db, projectName)
	defer cleanupOrgs(t, db, orgName)

	org := createTestOrg(orgName)
	err := orgRepo.Create(ctx, org)
	require.NoError(t, err)

	project := createTestProject(projectName, org.ID)
	err = projectRepo.Create(ctx, project)
	require.NoError(t, err)

	t.Run("name does not exist", func(t *testing.T) {
		exists, err := datasetRepo.NameExists(ctx, project.ID, datasetName)
		require.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("name exists", func(t *testing.T) {
		dataset := createTestDataset(datasetName, project.ID)
		err := datasetRepo.Create(ctx, dataset)
		require.NoError(t, err)

		exists, err := datasetRepo.NameExists(ctx, project.ID, datasetName)
		require.NoError(t, err)
		assert.True(t, exists)
	})
}
