package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	apperrors "github.com/agenttrace/agenttrace/api/internal/pkg/errors"
	"github.com/agenttrace/agenttrace/api/internal/testutil"
)

// MockDatasetService mocks the dataset service for testing
type MockDatasetService struct {
	mock.Mock
}

func (m *MockDatasetService) Get(ctx context.Context, id uuid.UUID) (*domain.Dataset, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Dataset), args.Error(1)
}

func (m *MockDatasetService) GetByName(ctx context.Context, projectID uuid.UUID, name string) (*domain.Dataset, error) {
	args := m.Called(ctx, projectID, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Dataset), args.Error(1)
}

func (m *MockDatasetService) List(ctx context.Context, filter *domain.DatasetFilter, limit, offset int) (*domain.DatasetList, error) {
	args := m.Called(ctx, filter, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.DatasetList), args.Error(1)
}

func (m *MockDatasetService) Create(ctx context.Context, projectID uuid.UUID, input *domain.DatasetInput) (*domain.Dataset, error) {
	args := m.Called(ctx, projectID, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Dataset), args.Error(1)
}

func (m *MockDatasetService) Update(ctx context.Context, id uuid.UUID, input *domain.DatasetInput) (*domain.Dataset, error) {
	args := m.Called(ctx, id, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Dataset), args.Error(1)
}

func (m *MockDatasetService) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockDatasetService) AddItem(ctx context.Context, datasetID uuid.UUID, input *domain.DatasetItemInput) (*domain.DatasetItem, error) {
	args := m.Called(ctx, datasetID, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.DatasetItem), args.Error(1)
}

func (m *MockDatasetService) ListItems(ctx context.Context, filter *domain.DatasetItemFilter, limit, offset int) ([]domain.DatasetItem, int64, error) {
	args := m.Called(ctx, filter, limit, offset)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]domain.DatasetItem), args.Get(1).(int64), args.Error(2)
}

func (m *MockDatasetService) UpdateItem(ctx context.Context, id uuid.UUID, input *domain.DatasetItemUpdateInput) (*domain.DatasetItem, error) {
	args := m.Called(ctx, id, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.DatasetItem), args.Error(1)
}

func (m *MockDatasetService) DeleteItem(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockDatasetService) CreateRun(ctx context.Context, datasetID uuid.UUID, input *domain.DatasetRunInput) (*domain.DatasetRun, error) {
	args := m.Called(ctx, datasetID, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.DatasetRun), args.Error(1)
}

func (m *MockDatasetService) GetRun(ctx context.Context, id uuid.UUID) (*domain.DatasetRun, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.DatasetRun), args.Error(1)
}

func (m *MockDatasetService) ListRuns(ctx context.Context, datasetID uuid.UUID, limit, offset int) ([]domain.DatasetRun, int64, error) {
	args := m.Called(ctx, datasetID, limit, offset)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]domain.DatasetRun), args.Get(1).(int64), args.Error(2)
}

func (m *MockDatasetService) AddRunItem(ctx context.Context, runID uuid.UUID, input *domain.DatasetRunItemInput) (*domain.DatasetRunItem, error) {
	args := m.Called(ctx, runID, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.DatasetRunItem), args.Error(1)
}

func setupDatasetTestApp(mockSvc *MockDatasetService, projectID *uuid.UUID) *fiber.App {
	app := fiber.New()
	logger := zap.NewNop()

	if projectID != nil {
		app.Use(testutil.TestProjectMiddleware(*projectID))
	}

	// ListDatasets
	app.Get("/v1/datasets", func(c *fiber.Ctx) error {
		projectID, ok := c.Locals("projectID").(uuid.UUID)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "Unauthorized",
				"message": "Project ID not found",
			})
		}

		filter := &domain.DatasetFilter{
			ProjectID: projectID,
		}

		limit := 50
		offset := 0

		list, err := mockSvc.List(c.Context(), filter, limit, offset)
		if err != nil {
			logger.Error("failed to list datasets", zap.Error(err))
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "Internal Server Error",
				"message": "Failed to list datasets",
			})
		}

		return c.JSON(list)
	})

	// GetDataset
	app.Get("/v1/datasets/:datasetId", func(c *fiber.Ctx) error {
		datasetID, err := uuid.Parse(c.Params("datasetId"))
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Bad Request",
				"message": "Invalid dataset ID",
			})
		}

		dataset, err := mockSvc.Get(c.Context(), datasetID)
		if err != nil {
			if apperrors.IsNotFound(err) {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"error":   "Not Found",
					"message": "Dataset not found",
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "Internal Server Error",
				"message": "Failed to get dataset",
			})
		}

		return c.JSON(dataset)
	})

	// GetDatasetByName
	app.Get("/v1/datasets/name/:name", func(c *fiber.Ctx) error {
		projectID, ok := c.Locals("projectID").(uuid.UUID)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "Unauthorized",
				"message": "Project ID not found",
			})
		}

		name := c.Params("name")
		if name == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Bad Request",
				"message": "Dataset name required",
			})
		}

		dataset, err := mockSvc.GetByName(c.Context(), projectID, name)
		if err != nil {
			if apperrors.IsNotFound(err) {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"error":   "Not Found",
					"message": "Dataset not found",
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "Internal Server Error",
				"message": "Failed to get dataset",
			})
		}

		return c.JSON(dataset)
	})

	// CreateDataset
	app.Post("/v1/datasets", func(c *fiber.Ctx) error {
		projectID, ok := c.Locals("projectID").(uuid.UUID)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "Unauthorized",
				"message": "Project ID not found",
			})
		}

		var input domain.DatasetInput
		if err := c.BodyParser(&input); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Bad Request",
				"message": "Invalid request body: " + err.Error(),
			})
		}

		if input.Name == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Bad Request",
				"message": "name is required",
			})
		}

		dataset, err := mockSvc.Create(c.Context(), projectID, &input)
		if err != nil {
			if apperrors.IsValidation(err) {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error":   "Bad Request",
					"message": err.Error(),
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "Internal Server Error",
				"message": "Failed to create dataset",
			})
		}

		return c.Status(fiber.StatusCreated).JSON(dataset)
	})

	// UpdateDataset
	app.Patch("/v1/datasets/:datasetId", func(c *fiber.Ctx) error {
		datasetID, err := uuid.Parse(c.Params("datasetId"))
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Bad Request",
				"message": "Invalid dataset ID",
			})
		}

		var input domain.DatasetInput
		if err := c.BodyParser(&input); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Bad Request",
				"message": "Invalid request body: " + err.Error(),
			})
		}

		dataset, err := mockSvc.Update(c.Context(), datasetID, &input)
		if err != nil {
			if apperrors.IsNotFound(err) {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"error":   "Not Found",
					"message": "Dataset not found",
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "Internal Server Error",
				"message": "Failed to update dataset",
			})
		}

		return c.JSON(dataset)
	})

	// DeleteDataset
	app.Delete("/v1/datasets/:datasetId", func(c *fiber.Ctx) error {
		datasetID, err := uuid.Parse(c.Params("datasetId"))
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Bad Request",
				"message": "Invalid dataset ID",
			})
		}

		if err := mockSvc.Delete(c.Context(), datasetID); err != nil {
			if apperrors.IsNotFound(err) {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"error":   "Not Found",
					"message": "Dataset not found",
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "Internal Server Error",
				"message": "Failed to delete dataset",
			})
		}

		return c.SendStatus(fiber.StatusNoContent)
	})

	// ListItems
	app.Get("/v1/datasets/:datasetId/items", func(c *fiber.Ctx) error {
		datasetID, err := uuid.Parse(c.Params("datasetId"))
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Bad Request",
				"message": "Invalid dataset ID",
			})
		}

		filter := &domain.DatasetItemFilter{
			DatasetID: datasetID,
		}

		items, totalCount, err := mockSvc.ListItems(c.Context(), filter, 50, 0)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "Internal Server Error",
				"message": "Failed to list items",
			})
		}

		return c.JSON(fiber.Map{
			"data":       items,
			"totalCount": totalCount,
		})
	})

	// CreateItem
	app.Post("/v1/datasets/:datasetId/items", func(c *fiber.Ctx) error {
		datasetID, err := uuid.Parse(c.Params("datasetId"))
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Bad Request",
				"message": "Invalid dataset ID",
			})
		}

		var input domain.DatasetItemInput
		if err := c.BodyParser(&input); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Bad Request",
				"message": "Invalid request body: " + err.Error(),
			})
		}

		item, err := mockSvc.AddItem(c.Context(), datasetID, &input)
		if err != nil {
			if apperrors.IsNotFound(err) {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"error":   "Not Found",
					"message": "Dataset not found",
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "Internal Server Error",
				"message": "Failed to create item",
			})
		}

		return c.Status(fiber.StatusCreated).JSON(item)
	})

	// DeleteItem
	app.Delete("/v1/datasets/:datasetId/items/:itemId", func(c *fiber.Ctx) error {
		itemID, err := uuid.Parse(c.Params("itemId"))
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Bad Request",
				"message": "Invalid item ID",
			})
		}

		if err := mockSvc.DeleteItem(c.Context(), itemID); err != nil {
			if apperrors.IsNotFound(err) {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"error":   "Not Found",
					"message": "Item not found",
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "Internal Server Error",
				"message": "Failed to delete item",
			})
		}

		return c.SendStatus(fiber.StatusNoContent)
	})

	// ListRuns
	app.Get("/v1/datasets/:datasetId/runs", func(c *fiber.Ctx) error {
		datasetID, err := uuid.Parse(c.Params("datasetId"))
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Bad Request",
				"message": "Invalid dataset ID",
			})
		}

		runs, totalCount, err := mockSvc.ListRuns(c.Context(), datasetID, 50, 0)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "Internal Server Error",
				"message": "Failed to list runs",
			})
		}

		return c.JSON(fiber.Map{
			"data":       runs,
			"totalCount": totalCount,
		})
	})

	// GetRun
	app.Get("/v1/datasets/:datasetId/runs/:runId", func(c *fiber.Ctx) error {
		runID, err := uuid.Parse(c.Params("runId"))
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Bad Request",
				"message": "Invalid run ID",
			})
		}

		run, err := mockSvc.GetRun(c.Context(), runID)
		if err != nil {
			if apperrors.IsNotFound(err) {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"error":   "Not Found",
					"message": "Run not found",
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "Internal Server Error",
				"message": "Failed to get run",
			})
		}

		return c.JSON(run)
	})

	// CreateRun
	app.Post("/v1/datasets/:datasetId/runs", func(c *fiber.Ctx) error {
		datasetID, err := uuid.Parse(c.Params("datasetId"))
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Bad Request",
				"message": "Invalid dataset ID",
			})
		}

		var input domain.DatasetRunInput
		if err := c.BodyParser(&input); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Bad Request",
				"message": "Invalid request body: " + err.Error(),
			})
		}

		if input.Name == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Bad Request",
				"message": "name is required",
			})
		}

		run, err := mockSvc.CreateRun(c.Context(), datasetID, &input)
		if err != nil {
			if apperrors.IsNotFound(err) {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"error":   "Not Found",
					"message": "Dataset not found",
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "Internal Server Error",
				"message": "Failed to create run",
			})
		}

		return c.Status(fiber.StatusCreated).JSON(run)
	})

	// AddRunItem
	app.Post("/v1/datasets/:datasetId/runs/:runId/items", func(c *fiber.Ctx) error {
		runID, err := uuid.Parse(c.Params("runId"))
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Bad Request",
				"message": "Invalid run ID",
			})
		}

		var input domain.DatasetRunItemInput
		if err := c.BodyParser(&input); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Bad Request",
				"message": "Invalid request body: " + err.Error(),
			})
		}

		if input.DatasetItemID == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Bad Request",
				"message": "datasetItemId is required",
			})
		}

		if input.TraceID == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Bad Request",
				"message": "traceId is required",
			})
		}

		item, err := mockSvc.AddRunItem(c.Context(), runID, &input)
		if err != nil {
			if apperrors.IsNotFound(err) {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"error":   "Not Found",
					"message": "Run or dataset item not found",
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "Internal Server Error",
				"message": "Failed to add run item",
			})
		}

		return c.Status(fiber.StatusCreated).JSON(item)
	})

	return app
}

// --- ListDatasets Tests ---

func TestDatasetsHandler_ListDatasets(t *testing.T) {
	t.Run("successfully lists datasets", func(t *testing.T) {
		mockSvc := new(MockDatasetService)
		projectID := uuid.New()
		app := setupDatasetTestApp(mockSvc, &projectID)

		list := &domain.DatasetList{
			Datasets: []domain.Dataset{
				{ID: uuid.New(), Name: "Dataset 1"},
				{ID: uuid.New(), Name: "Dataset 2"},
			},
			TotalCount: 2,
		}

		mockSvc.On("List", mock.Anything, mock.AnythingOfType("*domain.DatasetFilter"), 50, 0).Return(list, nil)

		req := httptest.NewRequest(http.MethodGet, "/v1/datasets", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result domain.DatasetList
		json.NewDecoder(resp.Body).Decode(&result)
		assert.Len(t, result.Datasets, 2)

		mockSvc.AssertExpectations(t)
	})

	t.Run("returns 401 without project ID", func(t *testing.T) {
		mockSvc := new(MockDatasetService)
		app := setupDatasetTestApp(mockSvc, nil)

		req := httptest.NewRequest(http.MethodGet, "/v1/datasets", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}

// --- GetDataset Tests ---

func TestDatasetsHandler_GetDataset(t *testing.T) {
	t.Run("successfully gets dataset", func(t *testing.T) {
		mockSvc := new(MockDatasetService)
		projectID := uuid.New()
		app := setupDatasetTestApp(mockSvc, &projectID)

		datasetID := uuid.New()
		dataset := &domain.Dataset{
			ID:        datasetID,
			Name:      "Test Dataset",
			CreatedAt: time.Now(),
		}

		mockSvc.On("Get", mock.Anything, datasetID).Return(dataset, nil)

		req := httptest.NewRequest(http.MethodGet, "/v1/datasets/"+datasetID.String(), nil)
		resp, err := app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result domain.Dataset
		json.NewDecoder(resp.Body).Decode(&result)
		assert.Equal(t, "Test Dataset", result.Name)

		mockSvc.AssertExpectations(t)
	})

	t.Run("returns 400 for invalid UUID", func(t *testing.T) {
		mockSvc := new(MockDatasetService)
		projectID := uuid.New()
		app := setupDatasetTestApp(mockSvc, &projectID)

		req := httptest.NewRequest(http.MethodGet, "/v1/datasets/invalid", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("returns 404 for non-existent dataset", func(t *testing.T) {
		mockSvc := new(MockDatasetService)
		projectID := uuid.New()
		app := setupDatasetTestApp(mockSvc, &projectID)

		datasetID := uuid.New()
		mockSvc.On("Get", mock.Anything, datasetID).Return(nil, apperrors.NotFound("not found"))

		req := httptest.NewRequest(http.MethodGet, "/v1/datasets/"+datasetID.String(), nil)
		resp, err := app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
		mockSvc.AssertExpectations(t)
	})
}

// --- GetDatasetByName Tests ---

func TestDatasetsHandler_GetDatasetByName(t *testing.T) {
	t.Run("successfully gets dataset by name", func(t *testing.T) {
		mockSvc := new(MockDatasetService)
		projectID := uuid.New()
		app := setupDatasetTestApp(mockSvc, &projectID)

		dataset := &domain.Dataset{
			ID:   uuid.New(),
			Name: "test-dataset",
		}

		mockSvc.On("GetByName", mock.Anything, projectID, "test-dataset").Return(dataset, nil)

		req := httptest.NewRequest(http.MethodGet, "/v1/datasets/name/test-dataset", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		mockSvc.AssertExpectations(t)
	})

	t.Run("returns 404 for non-existent name", func(t *testing.T) {
		mockSvc := new(MockDatasetService)
		projectID := uuid.New()
		app := setupDatasetTestApp(mockSvc, &projectID)

		mockSvc.On("GetByName", mock.Anything, projectID, "nonexistent").Return(nil, apperrors.NotFound("not found"))

		req := httptest.NewRequest(http.MethodGet, "/v1/datasets/name/nonexistent", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
		mockSvc.AssertExpectations(t)
	})
}

// --- CreateDataset Tests ---

func TestDatasetsHandler_CreateDataset(t *testing.T) {
	t.Run("successfully creates dataset", func(t *testing.T) {
		mockSvc := new(MockDatasetService)
		projectID := uuid.New()
		app := setupDatasetTestApp(mockSvc, &projectID)

		expectedDataset := &domain.Dataset{
			ID:   uuid.New(),
			Name: "New Dataset",
		}

		mockSvc.On("Create", mock.Anything, projectID, mock.AnythingOfType("*domain.DatasetInput")).Return(expectedDataset, nil)

		body, _ := json.Marshal(map[string]string{"name": "New Dataset"})
		req := httptest.NewRequest(http.MethodPost, "/v1/datasets", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusCreated, resp.StatusCode)
		mockSvc.AssertExpectations(t)
	})

	t.Run("returns 400 for missing name", func(t *testing.T) {
		mockSvc := new(MockDatasetService)
		projectID := uuid.New()
		app := setupDatasetTestApp(mockSvc, &projectID)

		body, _ := json.Marshal(map[string]string{})
		req := httptest.NewRequest(http.MethodPost, "/v1/datasets", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("returns 400 for duplicate name", func(t *testing.T) {
		mockSvc := new(MockDatasetService)
		projectID := uuid.New()
		app := setupDatasetTestApp(mockSvc, &projectID)

		mockSvc.On("Create", mock.Anything, projectID, mock.Anything).Return(nil, apperrors.Validation("dataset already exists"))

		body, _ := json.Marshal(map[string]string{"name": "Existing"})
		req := httptest.NewRequest(http.MethodPost, "/v1/datasets", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		mockSvc.AssertExpectations(t)
	})
}

// --- UpdateDataset Tests ---

func TestDatasetsHandler_UpdateDataset(t *testing.T) {
	t.Run("successfully updates dataset", func(t *testing.T) {
		mockSvc := new(MockDatasetService)
		projectID := uuid.New()
		app := setupDatasetTestApp(mockSvc, &projectID)

		datasetID := uuid.New()
		updatedDataset := &domain.Dataset{
			ID:   datasetID,
			Name: "Updated Name",
		}

		mockSvc.On("Update", mock.Anything, datasetID, mock.AnythingOfType("*domain.DatasetInput")).Return(updatedDataset, nil)

		body, _ := json.Marshal(map[string]string{"name": "Updated Name"})
		req := httptest.NewRequest(http.MethodPatch, "/v1/datasets/"+datasetID.String(), bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		mockSvc.AssertExpectations(t)
	})

	t.Run("returns 404 for non-existent dataset", func(t *testing.T) {
		mockSvc := new(MockDatasetService)
		projectID := uuid.New()
		app := setupDatasetTestApp(mockSvc, &projectID)

		datasetID := uuid.New()
		mockSvc.On("Update", mock.Anything, datasetID, mock.Anything).Return(nil, apperrors.NotFound("not found"))

		body, _ := json.Marshal(map[string]string{"name": "New Name"})
		req := httptest.NewRequest(http.MethodPatch, "/v1/datasets/"+datasetID.String(), bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
		mockSvc.AssertExpectations(t)
	})
}

// --- DeleteDataset Tests ---

func TestDatasetsHandler_DeleteDataset(t *testing.T) {
	t.Run("successfully deletes dataset", func(t *testing.T) {
		mockSvc := new(MockDatasetService)
		projectID := uuid.New()
		app := setupDatasetTestApp(mockSvc, &projectID)

		datasetID := uuid.New()
		mockSvc.On("Delete", mock.Anything, datasetID).Return(nil)

		req := httptest.NewRequest(http.MethodDelete, "/v1/datasets/"+datasetID.String(), nil)
		resp, err := app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusNoContent, resp.StatusCode)
		mockSvc.AssertExpectations(t)
	})

	t.Run("returns 404 for non-existent dataset", func(t *testing.T) {
		mockSvc := new(MockDatasetService)
		projectID := uuid.New()
		app := setupDatasetTestApp(mockSvc, &projectID)

		datasetID := uuid.New()
		mockSvc.On("Delete", mock.Anything, datasetID).Return(apperrors.NotFound("not found"))

		req := httptest.NewRequest(http.MethodDelete, "/v1/datasets/"+datasetID.String(), nil)
		resp, err := app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
		mockSvc.AssertExpectations(t)
	})
}

// --- Items Tests ---

func TestDatasetsHandler_ListItems(t *testing.T) {
	t.Run("successfully lists items", func(t *testing.T) {
		mockSvc := new(MockDatasetService)
		projectID := uuid.New()
		app := setupDatasetTestApp(mockSvc, &projectID)

		datasetID := uuid.New()
		items := []domain.DatasetItem{
			{ID: uuid.New(), DatasetID: datasetID},
		}

		mockSvc.On("ListItems", mock.Anything, mock.AnythingOfType("*domain.DatasetItemFilter"), 50, 0).Return(items, int64(1), nil)

		req := httptest.NewRequest(http.MethodGet, "/v1/datasets/"+datasetID.String()+"/items", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		mockSvc.AssertExpectations(t)
	})
}

func TestDatasetsHandler_CreateItem(t *testing.T) {
	t.Run("successfully creates item", func(t *testing.T) {
		mockSvc := new(MockDatasetService)
		projectID := uuid.New()
		app := setupDatasetTestApp(mockSvc, &projectID)

		datasetID := uuid.New()
		item := &domain.DatasetItem{
			ID:        uuid.New(),
			DatasetID: datasetID,
		}

		mockSvc.On("AddItem", mock.Anything, datasetID, mock.AnythingOfType("*domain.DatasetItemInput")).Return(item, nil)

		body, _ := json.Marshal(map[string]interface{}{"input": map[string]string{"key": "value"}})
		req := httptest.NewRequest(http.MethodPost, "/v1/datasets/"+datasetID.String()+"/items", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusCreated, resp.StatusCode)
		mockSvc.AssertExpectations(t)
	})
}

func TestDatasetsHandler_DeleteItem(t *testing.T) {
	t.Run("successfully deletes item", func(t *testing.T) {
		mockSvc := new(MockDatasetService)
		projectID := uuid.New()
		app := setupDatasetTestApp(mockSvc, &projectID)

		datasetID := uuid.New()
		itemID := uuid.New()
		mockSvc.On("DeleteItem", mock.Anything, itemID).Return(nil)

		req := httptest.NewRequest(http.MethodDelete, "/v1/datasets/"+datasetID.String()+"/items/"+itemID.String(), nil)
		resp, err := app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusNoContent, resp.StatusCode)
		mockSvc.AssertExpectations(t)
	})
}

// --- Runs Tests ---

func TestDatasetsHandler_ListRuns(t *testing.T) {
	t.Run("successfully lists runs", func(t *testing.T) {
		mockSvc := new(MockDatasetService)
		projectID := uuid.New()
		app := setupDatasetTestApp(mockSvc, &projectID)

		datasetID := uuid.New()
		runs := []domain.DatasetRun{
			{ID: uuid.New(), DatasetID: datasetID, Name: "Run 1"},
		}

		mockSvc.On("ListRuns", mock.Anything, datasetID, 50, 0).Return(runs, int64(1), nil)

		req := httptest.NewRequest(http.MethodGet, "/v1/datasets/"+datasetID.String()+"/runs", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		mockSvc.AssertExpectations(t)
	})
}

func TestDatasetsHandler_GetRun(t *testing.T) {
	t.Run("successfully gets run", func(t *testing.T) {
		mockSvc := new(MockDatasetService)
		projectID := uuid.New()
		app := setupDatasetTestApp(mockSvc, &projectID)

		datasetID := uuid.New()
		runID := uuid.New()
		run := &domain.DatasetRun{
			ID:        runID,
			DatasetID: datasetID,
			Name:      "Test Run",
		}

		mockSvc.On("GetRun", mock.Anything, runID).Return(run, nil)

		req := httptest.NewRequest(http.MethodGet, "/v1/datasets/"+datasetID.String()+"/runs/"+runID.String(), nil)
		resp, err := app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		mockSvc.AssertExpectations(t)
	})

	t.Run("returns 404 for non-existent run", func(t *testing.T) {
		mockSvc := new(MockDatasetService)
		projectID := uuid.New()
		app := setupDatasetTestApp(mockSvc, &projectID)

		datasetID := uuid.New()
		runID := uuid.New()
		mockSvc.On("GetRun", mock.Anything, runID).Return(nil, apperrors.NotFound("not found"))

		req := httptest.NewRequest(http.MethodGet, "/v1/datasets/"+datasetID.String()+"/runs/"+runID.String(), nil)
		resp, err := app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
		mockSvc.AssertExpectations(t)
	})
}

func TestDatasetsHandler_CreateRun(t *testing.T) {
	t.Run("successfully creates run", func(t *testing.T) {
		mockSvc := new(MockDatasetService)
		projectID := uuid.New()
		app := setupDatasetTestApp(mockSvc, &projectID)

		datasetID := uuid.New()
		run := &domain.DatasetRun{
			ID:        uuid.New(),
			DatasetID: datasetID,
			Name:      "New Run",
		}

		mockSvc.On("CreateRun", mock.Anything, datasetID, mock.AnythingOfType("*domain.DatasetRunInput")).Return(run, nil)

		body, _ := json.Marshal(map[string]string{"name": "New Run"})
		req := httptest.NewRequest(http.MethodPost, "/v1/datasets/"+datasetID.String()+"/runs", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusCreated, resp.StatusCode)
		mockSvc.AssertExpectations(t)
	})

	t.Run("returns 400 for missing name", func(t *testing.T) {
		mockSvc := new(MockDatasetService)
		projectID := uuid.New()
		app := setupDatasetTestApp(mockSvc, &projectID)

		datasetID := uuid.New()
		body, _ := json.Marshal(map[string]string{})
		req := httptest.NewRequest(http.MethodPost, "/v1/datasets/"+datasetID.String()+"/runs", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

func TestDatasetsHandler_AddRunItem(t *testing.T) {
	t.Run("successfully adds run item", func(t *testing.T) {
		mockSvc := new(MockDatasetService)
		projectID := uuid.New()
		app := setupDatasetTestApp(mockSvc, &projectID)

		datasetID := uuid.New()
		runID := uuid.New()
		datasetItemID := uuid.New()

		runItem := &domain.DatasetRunItem{
			ID:            uuid.New(),
			DatasetRunID:  runID,
			DatasetItemID: datasetItemID,
		}

		mockSvc.On("AddRunItem", mock.Anything, runID, mock.AnythingOfType("*domain.DatasetRunItemInput")).Return(runItem, nil)

		body, _ := json.Marshal(map[string]string{
			"datasetItemId": datasetItemID.String(),
			"traceId":       "trace-123",
		})
		req := httptest.NewRequest(http.MethodPost, "/v1/datasets/"+datasetID.String()+"/runs/"+runID.String()+"/items", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusCreated, resp.StatusCode)
		mockSvc.AssertExpectations(t)
	})

	t.Run("returns 400 for missing datasetItemId", func(t *testing.T) {
		mockSvc := new(MockDatasetService)
		projectID := uuid.New()
		app := setupDatasetTestApp(mockSvc, &projectID)

		datasetID := uuid.New()
		runID := uuid.New()

		body, _ := json.Marshal(map[string]string{
			"traceId": "trace-123",
		})
		req := httptest.NewRequest(http.MethodPost, "/v1/datasets/"+datasetID.String()+"/runs/"+runID.String()+"/items", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("returns 400 for missing traceId", func(t *testing.T) {
		mockSvc := new(MockDatasetService)
		projectID := uuid.New()
		app := setupDatasetTestApp(mockSvc, &projectID)

		datasetID := uuid.New()
		runID := uuid.New()
		datasetItemID := uuid.New()

		body, _ := json.Marshal(map[string]string{
			"datasetItemId": datasetItemID.String(),
		})
		req := httptest.NewRequest(http.MethodPost, "/v1/datasets/"+datasetID.String()+"/runs/"+runID.String()+"/items", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

func TestNewDatasetsHandler(t *testing.T) {
	t.Run("creates handler with dependencies", func(t *testing.T) {
		logger := zap.NewNop()
		handler := NewDatasetsHandler(nil, logger)

		require.NotNil(t, handler)
		assert.Equal(t, logger, handler.logger)
	})
}
