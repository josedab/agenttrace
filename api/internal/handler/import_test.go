package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/middleware"
)

// MockDatasetServiceForImport mocks the dataset service for import tests
type MockDatasetServiceForImport struct {
	mock.Mock
}

func (m *MockDatasetServiceForImport) Create(ctx context.Context, projectID uuid.UUID, input *domain.DatasetInput) (*domain.Dataset, error) {
	args := m.Called(ctx, projectID, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Dataset), args.Error(1)
}

func (m *MockDatasetServiceForImport) AddItem(ctx context.Context, datasetID uuid.UUID, input *domain.DatasetItemInput) (*domain.DatasetItem, error) {
	args := m.Called(ctx, datasetID, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.DatasetItem), args.Error(1)
}

// MockPromptServiceForImport mocks the prompt service for import tests
type MockPromptServiceForImport struct {
	mock.Mock
}

func (m *MockPromptServiceForImport) Create(ctx context.Context, projectID uuid.UUID, input *domain.PromptInput, userID uuid.UUID) (*domain.Prompt, error) {
	args := m.Called(ctx, projectID, input, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Prompt), args.Error(1)
}

// importTestProjectMiddleware injects a test project ID for import tests
func importTestProjectMiddleware(projectID uuid.UUID) fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Locals(string(middleware.ContextKeyProjectID), projectID)
		return c.Next()
	}
}

func setupImportTestApp(
	mockDatasetSvc *MockDatasetServiceForImport,
	mockPromptSvc *MockPromptServiceForImport,
	projectID uuid.UUID,
) *fiber.App {
	app := fiber.New()
	logger := zap.NewNop()

	app.Use(importTestProjectMiddleware(projectID))

	// Import dataset endpoint
	app.Post("/v1/import/dataset", func(c *fiber.Ctx) error {
		pid, ok := middleware.GetProjectID(c)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "Unauthorized",
				"message": "Project ID not found",
			})
		}

		var request ImportDatasetRequest
		if err := c.BodyParser(&request); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Bad Request",
				"message": "Invalid request body: " + err.Error(),
			})
		}

		if request.Name == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Bad Request",
				"message": "name is required",
			})
		}

		if len(request.Items) == 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Bad Request",
				"message": "at least one item is required",
			})
		}

		datasetInput := &domain.DatasetInput{
			Name:        request.Name,
			Description: request.Description,
			Metadata:    request.Metadata,
		}

		dataset, err := mockDatasetSvc.Create(c.Context(), pid, datasetInput)
		if err != nil {
			logger.Error("failed to create dataset for import", zap.Error(err))
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "Internal Server Error",
				"message": "Failed to create dataset",
			})
		}

		importedCount := 0
		for _, item := range request.Items {
			itemInput := &domain.DatasetItemInput{
				Input:          item.Input,
				ExpectedOutput: item.ExpectedOutput,
				Metadata:       item.Metadata,
			}

			_, err := mockDatasetSvc.AddItem(c.Context(), dataset.ID, itemInput)
			if err != nil {
				continue
			}
			importedCount++
		}

		return c.Status(fiber.StatusCreated).JSON(ImportResponse{
			ID:            dataset.ID.String(),
			ImportedCount: importedCount,
			Message:       "Successfully imported items",
		})
	})

	// Import prompt endpoint
	app.Post("/v1/import/prompt", func(c *fiber.Ctx) error {
		pid, ok := middleware.GetProjectID(c)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "Unauthorized",
				"message": "Project ID not found",
			})
		}

		var request ImportPromptRequest
		if err := c.BodyParser(&request); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Bad Request",
				"message": "Invalid request body: " + err.Error(),
			})
		}

		if request.Name == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Bad Request",
				"message": "name is required",
			})
		}

		if request.Content == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Bad Request",
				"message": "content is required",
			})
		}

		promptInput := &domain.PromptInput{
			Name:        request.Name,
			Content:     request.Content,
			Description: request.Description,
			Config:      request.Config,
			Labels:      request.Labels,
			Tags:        request.Tags,
		}

		prompt, err := mockPromptSvc.Create(c.Context(), pid, promptInput, uuid.Nil)
		if err != nil {
			logger.Error("failed to create prompt for import", zap.Error(err))
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "Internal Server Error",
				"message": "Failed to create prompt",
			})
		}

		return c.Status(fiber.StatusCreated).JSON(fiber.Map{
			"id":      prompt.ID.String(),
			"name":    prompt.Name,
			"message": "Prompt imported successfully",
		})
	})

	return app
}

func TestImportHandler_ImportDataset(t *testing.T) {
	projectID := uuid.New()

	t.Run("successfully imports dataset", func(t *testing.T) {
		mockDatasetSvc := new(MockDatasetServiceForImport)
		mockPromptSvc := new(MockPromptServiceForImport)
		app := setupImportTestApp(mockDatasetSvc, mockPromptSvc, projectID)

		datasetID := uuid.New()
		dataset := &domain.Dataset{
			ID:        datasetID,
			ProjectID: projectID,
			Name:      "test-dataset",
		}
		mockDatasetSvc.On("Create", mock.Anything, projectID, mock.AnythingOfType("*domain.DatasetInput")).Return(dataset, nil)
		mockDatasetSvc.On("AddItem", mock.Anything, datasetID, mock.AnythingOfType("*domain.DatasetItemInput")).Return(&domain.DatasetItem{}, nil)

		reqBody := ImportDatasetRequest{
			Name:        "test-dataset",
			Description: "Test dataset description",
			Items: []DatasetItemImport{
				{
					Input:          "What is 2+2?",
					ExpectedOutput: "4",
				},
				{
					Input:          "What is 3+3?",
					ExpectedOutput: "6",
				},
			},
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/v1/import/dataset", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var result ImportResponse
		respBody, _ := io.ReadAll(resp.Body)
		err = json.Unmarshal(respBody, &result)
		require.NoError(t, err)

		assert.Equal(t, datasetID.String(), result.ID)
		assert.Equal(t, 2, result.ImportedCount)

		mockDatasetSvc.AssertExpectations(t)
	})

	t.Run("returns 400 for missing name", func(t *testing.T) {
		mockDatasetSvc := new(MockDatasetServiceForImport)
		mockPromptSvc := new(MockPromptServiceForImport)
		app := setupImportTestApp(mockDatasetSvc, mockPromptSvc, projectID)

		reqBody := ImportDatasetRequest{
			Items: []DatasetItemImport{
				{Input: "test"},
			},
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/v1/import/dataset", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

		var result map[string]string
		respBody, _ := io.ReadAll(resp.Body)
		err = json.Unmarshal(respBody, &result)
		require.NoError(t, err)

		assert.Contains(t, result["message"], "name is required")
	})

	t.Run("returns 400 for empty items", func(t *testing.T) {
		mockDatasetSvc := new(MockDatasetServiceForImport)
		mockPromptSvc := new(MockPromptServiceForImport)
		app := setupImportTestApp(mockDatasetSvc, mockPromptSvc, projectID)

		reqBody := ImportDatasetRequest{
			Name:  "test-dataset",
			Items: []DatasetItemImport{},
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/v1/import/dataset", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

		var result map[string]string
		respBody, _ := io.ReadAll(resp.Body)
		err = json.Unmarshal(respBody, &result)
		require.NoError(t, err)

		assert.Contains(t, result["message"], "at least one item is required")
	})

	t.Run("returns 400 for invalid JSON", func(t *testing.T) {
		mockDatasetSvc := new(MockDatasetServiceForImport)
		mockPromptSvc := new(MockPromptServiceForImport)
		app := setupImportTestApp(mockDatasetSvc, mockPromptSvc, projectID)

		req := httptest.NewRequest(http.MethodPost, "/v1/import/dataset", strings.NewReader("invalid json"))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("returns 500 when dataset creation fails", func(t *testing.T) {
		mockDatasetSvc := new(MockDatasetServiceForImport)
		mockPromptSvc := new(MockPromptServiceForImport)
		app := setupImportTestApp(mockDatasetSvc, mockPromptSvc, projectID)

		mockDatasetSvc.On("Create", mock.Anything, projectID, mock.AnythingOfType("*domain.DatasetInput")).Return(nil, errors.New("db error"))

		reqBody := ImportDatasetRequest{
			Name: "test-dataset",
			Items: []DatasetItemImport{
				{Input: "test"},
			},
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/v1/import/dataset", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

		mockDatasetSvc.AssertExpectations(t)
	})

	t.Run("handles partial item import failures gracefully", func(t *testing.T) {
		mockDatasetSvc := new(MockDatasetServiceForImport)
		mockPromptSvc := new(MockPromptServiceForImport)
		app := setupImportTestApp(mockDatasetSvc, mockPromptSvc, projectID)

		datasetID := uuid.New()
		dataset := &domain.Dataset{
			ID:        datasetID,
			ProjectID: projectID,
			Name:      "test-dataset",
		}
		mockDatasetSvc.On("Create", mock.Anything, projectID, mock.AnythingOfType("*domain.DatasetInput")).Return(dataset, nil)
		// First item succeeds, second fails
		mockDatasetSvc.On("AddItem", mock.Anything, datasetID, mock.AnythingOfType("*domain.DatasetItemInput")).Return(&domain.DatasetItem{}, nil).Once()
		mockDatasetSvc.On("AddItem", mock.Anything, datasetID, mock.AnythingOfType("*domain.DatasetItemInput")).Return(nil, errors.New("item error")).Once()
		mockDatasetSvc.On("AddItem", mock.Anything, datasetID, mock.AnythingOfType("*domain.DatasetItemInput")).Return(&domain.DatasetItem{}, nil).Once()

		reqBody := ImportDatasetRequest{
			Name: "test-dataset",
			Items: []DatasetItemImport{
				{Input: "item1"},
				{Input: "item2"},
				{Input: "item3"},
			},
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/v1/import/dataset", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var result ImportResponse
		respBody, _ := io.ReadAll(resp.Body)
		err = json.Unmarshal(respBody, &result)
		require.NoError(t, err)

		// Only 2 items succeeded (first and third)
		assert.Equal(t, 2, result.ImportedCount)

		mockDatasetSvc.AssertExpectations(t)
	})
}

func TestImportHandler_ImportPrompt(t *testing.T) {
	projectID := uuid.New()

	t.Run("successfully imports prompt", func(t *testing.T) {
		mockDatasetSvc := new(MockDatasetServiceForImport)
		mockPromptSvc := new(MockPromptServiceForImport)
		app := setupImportTestApp(mockDatasetSvc, mockPromptSvc, projectID)

		promptID := uuid.New()
		prompt := &domain.Prompt{
			ID:        promptID,
			ProjectID: projectID,
			Name:      "test-prompt",
		}
		mockPromptSvc.On("Create", mock.Anything, projectID, mock.AnythingOfType("*domain.PromptInput"), uuid.Nil).Return(prompt, nil)

		description := "A test prompt"
		reqBody := ImportPromptRequest{
			Name:        "test-prompt",
			Content:     "You are a helpful assistant. {{instruction}}",
			Description: &description,
			Labels:      []string{"production"},
			Tags:        []string{"assistant", "general"},
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/v1/import/prompt", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var result map[string]interface{}
		respBody, _ := io.ReadAll(resp.Body)
		err = json.Unmarshal(respBody, &result)
		require.NoError(t, err)

		assert.Equal(t, promptID.String(), result["id"])
		assert.Equal(t, "test-prompt", result["name"])
		assert.Equal(t, "Prompt imported successfully", result["message"])

		mockPromptSvc.AssertExpectations(t)
	})

	t.Run("returns 400 for missing name", func(t *testing.T) {
		mockDatasetSvc := new(MockDatasetServiceForImport)
		mockPromptSvc := new(MockPromptServiceForImport)
		app := setupImportTestApp(mockDatasetSvc, mockPromptSvc, projectID)

		reqBody := ImportPromptRequest{
			Content: "Some content",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/v1/import/prompt", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

		var result map[string]string
		respBody, _ := io.ReadAll(resp.Body)
		err = json.Unmarshal(respBody, &result)
		require.NoError(t, err)

		assert.Contains(t, result["message"], "name is required")
	})

	t.Run("returns 400 for missing content", func(t *testing.T) {
		mockDatasetSvc := new(MockDatasetServiceForImport)
		mockPromptSvc := new(MockPromptServiceForImport)
		app := setupImportTestApp(mockDatasetSvc, mockPromptSvc, projectID)

		reqBody := ImportPromptRequest{
			Name: "test-prompt",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/v1/import/prompt", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

		var result map[string]string
		respBody, _ := io.ReadAll(resp.Body)
		err = json.Unmarshal(respBody, &result)
		require.NoError(t, err)

		assert.Contains(t, result["message"], "content is required")
	})

	t.Run("returns 400 for invalid JSON", func(t *testing.T) {
		mockDatasetSvc := new(MockDatasetServiceForImport)
		mockPromptSvc := new(MockPromptServiceForImport)
		app := setupImportTestApp(mockDatasetSvc, mockPromptSvc, projectID)

		req := httptest.NewRequest(http.MethodPost, "/v1/import/prompt", strings.NewReader("not valid json"))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("returns 500 when prompt creation fails", func(t *testing.T) {
		mockDatasetSvc := new(MockDatasetServiceForImport)
		mockPromptSvc := new(MockPromptServiceForImport)
		app := setupImportTestApp(mockDatasetSvc, mockPromptSvc, projectID)

		mockPromptSvc.On("Create", mock.Anything, projectID, mock.AnythingOfType("*domain.PromptInput"), uuid.Nil).Return(nil, errors.New("db error"))

		reqBody := ImportPromptRequest{
			Name:    "test-prompt",
			Content: "Some content",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/v1/import/prompt", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

		mockPromptSvc.AssertExpectations(t)
	})

	t.Run("imports prompt with config", func(t *testing.T) {
		mockDatasetSvc := new(MockDatasetServiceForImport)
		mockPromptSvc := new(MockPromptServiceForImport)
		app := setupImportTestApp(mockDatasetSvc, mockPromptSvc, projectID)

		promptID := uuid.New()
		prompt := &domain.Prompt{
			ID:        promptID,
			ProjectID: projectID,
			Name:      "test-prompt",
		}
		mockPromptSvc.On("Create", mock.Anything, projectID, mock.MatchedBy(func(input *domain.PromptInput) bool {
			if input.Config == nil {
				return false
			}
			config, ok := input.Config.(map[string]any)
			if !ok {
				return false
			}
			return config["model"] == "gpt-4" && config["temperature"] == 0.7
		}), uuid.Nil).Return(prompt, nil)

		reqBody := ImportPromptRequest{
			Name:    "test-prompt",
			Content: "You are a helpful assistant.",
			Config: map[string]any{
				"model":       "gpt-4",
				"temperature": 0.7,
			},
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/v1/import/prompt", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		mockPromptSvc.AssertExpectations(t)
	})
}

// Helper function to create multipart form data for CSV import tests
func createMultipartForm(fieldName, fileName, content string, fields map[string]string) (*bytes.Buffer, string, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add form fields
	for key, value := range fields {
		if err := writer.WriteField(key, value); err != nil {
			return nil, "", err
		}
	}

	// Add file
	part, err := writer.CreateFormFile(fieldName, fileName)
	if err != nil {
		return nil, "", err
	}
	if _, err := io.Copy(part, strings.NewReader(content)); err != nil {
		return nil, "", err
	}

	if err := writer.Close(); err != nil {
		return nil, "", err
	}

	return body, writer.FormDataContentType(), nil
}
