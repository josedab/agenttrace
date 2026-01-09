package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
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
	"github.com/agenttrace/agenttrace/api/internal/middleware"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// MockPromptService mocks the prompt service
type MockPromptService struct {
	mock.Mock
}

func (m *MockPromptService) Create(ctx context.Context, projectID uuid.UUID, input *domain.PromptInput, userID uuid.UUID) (*domain.Prompt, error) {
	args := m.Called(ctx, projectID, input, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Prompt), args.Error(1)
}

func (m *MockPromptService) Get(ctx context.Context, id uuid.UUID) (*domain.Prompt, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Prompt), args.Error(1)
}

func (m *MockPromptService) GetByName(ctx context.Context, projectID uuid.UUID, name string) (*domain.Prompt, error) {
	args := m.Called(ctx, projectID, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Prompt), args.Error(1)
}

func (m *MockPromptService) GetByNameAndVersion(ctx context.Context, projectID uuid.UUID, name string, version int) (*domain.Prompt, error) {
	args := m.Called(ctx, projectID, name, version)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Prompt), args.Error(1)
}

func (m *MockPromptService) GetByNameAndLabel(ctx context.Context, projectID uuid.UUID, name, label string) (*domain.Prompt, error) {
	args := m.Called(ctx, projectID, name, label)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Prompt), args.Error(1)
}

func (m *MockPromptService) Update(ctx context.Context, id uuid.UUID, input *domain.PromptInput) (*domain.Prompt, error) {
	args := m.Called(ctx, id, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Prompt), args.Error(1)
}

func (m *MockPromptService) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockPromptService) List(ctx context.Context, filter *domain.PromptFilter, limit, offset int) (*domain.PromptList, error) {
	args := m.Called(ctx, filter, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.PromptList), args.Error(1)
}

func (m *MockPromptService) CreateVersion(ctx context.Context, promptID uuid.UUID, input *domain.PromptVersionInput, userID uuid.UUID) (*domain.PromptVersion, error) {
	args := m.Called(ctx, promptID, input, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.PromptVersion), args.Error(1)
}

func (m *MockPromptService) ListVersions(ctx context.Context, promptID uuid.UUID) ([]domain.PromptVersion, error) {
	args := m.Called(ctx, promptID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.PromptVersion), args.Error(1)
}

func (m *MockPromptService) SetVersionLabel(ctx context.Context, promptID uuid.UUID, version int, label string, add bool) error {
	args := m.Called(ctx, promptID, version, label, add)
	return args.Error(0)
}

func (m *MockPromptService) Compile(ctx context.Context, projectID uuid.UUID, name string, variables map[string]string, options *service.CompileOptions) (*service.CompiledPrompt, error) {
	args := m.Called(ctx, projectID, name, variables, options)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.CompiledPrompt), args.Error(1)
}

func setupPromptsTestApp(mockSvc *MockPromptService, projectID uuid.UUID) *fiber.App {
	app := fiber.New()
	logger := zap.NewNop()

	app.Use(testProjectMiddleware(projectID))

	// ListPrompts
	app.Get("/v1/prompts", func(c *fiber.Ctx) error {
		pid, ok := middleware.GetProjectID(c)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized",
			})
		}

		filter := &domain.PromptFilter{
			ProjectID: pid,
		}

		limit := parseIntParam(c, "limit", 50)
		offset := parseIntParam(c, "offset", 0)

		list, err := mockSvc.List(c.Context(), filter, limit, offset)
		if err != nil {
			logger.Error("failed to list prompts")
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Internal Server Error",
			})
		}

		return c.JSON(list)
	})

	// GetPrompt
	app.Get("/v1/prompts/:promptName", func(c *fiber.Ctx) error {
		pid, ok := middleware.GetProjectID(c)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized",
			})
		}

		promptName := c.Params("promptName")
		if promptName == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Bad Request",
				"message": "Prompt name required",
			})
		}

		version := parseIntParam(c, "version", 0)
		label := c.Query("label")

		var prompt *domain.Prompt
		var err error

		if version > 0 {
			prompt, err = mockSvc.GetByNameAndVersion(c.Context(), pid, promptName, version)
		} else if label != "" {
			prompt, err = mockSvc.GetByNameAndLabel(c.Context(), pid, promptName, label)
		} else {
			prompt, err = mockSvc.GetByName(c.Context(), pid, promptName)
		}

		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   "Not Found",
				"message": "Prompt not found",
			})
		}

		return c.JSON(prompt)
	})

	// CreatePrompt
	app.Post("/v1/prompts", func(c *fiber.Ctx) error {
		pid, ok := middleware.GetProjectID(c)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized",
			})
		}

		var input domain.PromptInput
		if err := c.BodyParser(&input); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Bad Request",
				"message": "Invalid request body",
			})
		}

		if input.Name == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Bad Request",
				"message": "name is required",
			})
		}

		if input.Content == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Bad Request",
				"message": "content is required",
			})
		}

		prompt, err := mockSvc.Create(c.Context(), pid, &input, uuid.Nil)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Internal Server Error",
			})
		}

		return c.Status(fiber.StatusCreated).JSON(prompt)
	})

	// UpdatePrompt
	app.Patch("/v1/prompts/:promptId", func(c *fiber.Ctx) error {
		promptIDStr := c.Params("promptId")
		if promptIDStr == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Bad Request",
				"message": "Prompt ID required",
			})
		}

		promptID, err := uuid.Parse(promptIDStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Bad Request",
				"message": "Invalid prompt ID",
			})
		}

		var input domain.PromptInput
		if err := c.BodyParser(&input); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Bad Request",
				"message": "Invalid request body",
			})
		}

		prompt, err := mockSvc.Update(c.Context(), promptID, &input)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   "Not Found",
				"message": "Prompt not found",
			})
		}

		return c.JSON(prompt)
	})

	// DeletePrompt
	app.Delete("/v1/prompts/:promptId", func(c *fiber.Ctx) error {
		promptIDStr := c.Params("promptId")
		if promptIDStr == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Bad Request",
				"message": "Prompt ID required",
			})
		}

		promptID, err := uuid.Parse(promptIDStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Bad Request",
				"message": "Invalid prompt ID",
			})
		}

		if err := mockSvc.Delete(c.Context(), promptID); err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   "Not Found",
				"message": "Prompt not found",
			})
		}

		return c.SendStatus(fiber.StatusNoContent)
	})

	// ListVersions
	app.Get("/v1/prompts/:promptName/versions", func(c *fiber.Ctx) error {
		pid, ok := middleware.GetProjectID(c)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized",
			})
		}

		promptName := c.Params("promptName")
		if promptName == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Bad Request",
				"message": "Prompt name required",
			})
		}

		// Get prompt first
		prompt, err := mockSvc.GetByName(c.Context(), pid, promptName)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   "Not Found",
				"message": "Prompt not found",
			})
		}

		versions, err := mockSvc.ListVersions(c.Context(), prompt.ID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Internal Server Error",
			})
		}

		return c.JSON(fiber.Map{
			"data":       versions,
			"totalCount": len(versions),
		})
	})

	// CompilePrompt
	app.Post("/v1/prompts/:promptName/compile", func(c *fiber.Ctx) error {
		pid, ok := middleware.GetProjectID(c)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized",
			})
		}

		promptName := c.Params("promptName")
		if promptName == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Bad Request",
				"message": "Prompt name required",
			})
		}

		var input struct {
			Variables map[string]string `json:"variables"`
			Version   *int              `json:"version,omitempty"`
			Label     *string           `json:"label,omitempty"`
		}

		if err := c.BodyParser(&input); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Bad Request",
				"message": "Invalid request body",
			})
		}

		options := &service.CompileOptions{
			Version: input.Version,
		}
		if input.Label != nil {
			options.Label = *input.Label
		}

		compiled, err := mockSvc.Compile(c.Context(), pid, promptName, input.Variables, options)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   "Not Found",
				"message": "Prompt not found",
			})
		}

		return c.JSON(compiled)
	})

	return app
}

func TestPromptsHandler_ListPrompts(t *testing.T) {
	t.Run("successfully lists prompts", func(t *testing.T) {
		mockSvc := new(MockPromptService)
		projectID := uuid.New()
		app := setupPromptsTestApp(mockSvc, projectID)

		expectedList := &domain.PromptList{
			Prompts: []domain.Prompt{
				{
					ID:        uuid.New(),
					ProjectID: projectID,
					Name:      "greeting-prompt",
					Type:      domain.PromptTypeText,
					CreatedAt: time.Now(),
				},
				{
					ID:        uuid.New(),
					ProjectID: projectID,
					Name:      "summary-prompt",
					Type:      domain.PromptTypeText,
					CreatedAt: time.Now(),
				},
			},
			TotalCount: 2,
			HasMore:    false,
		}

		mockSvc.On("List", mock.Anything, mock.AnythingOfType("*domain.PromptFilter"), 50, 0).
			Return(expectedList, nil)

		req := httptest.NewRequest(http.MethodGet, "/v1/prompts", nil)

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result domain.PromptList
		respBody, _ := io.ReadAll(resp.Body)
		err = json.Unmarshal(respBody, &result)
		require.NoError(t, err)
		assert.Len(t, result.Prompts, 2)
		assert.Equal(t, int64(2), result.TotalCount)

		mockSvc.AssertExpectations(t)
	})
}

func TestPromptsHandler_GetPrompt(t *testing.T) {
	t.Run("successfully gets prompt by name", func(t *testing.T) {
		mockSvc := new(MockPromptService)
		projectID := uuid.New()
		app := setupPromptsTestApp(mockSvc, projectID)

		promptID := uuid.New()
		expectedPrompt := &domain.Prompt{
			ID:          promptID,
			ProjectID:   projectID,
			Name:        "greeting-prompt",
			Type:        domain.PromptTypeText,
			Description: "A greeting prompt",
			CreatedAt:   time.Now(),
		}

		mockSvc.On("GetByName", mock.Anything, projectID, "greeting-prompt").
			Return(expectedPrompt, nil)

		req := httptest.NewRequest(http.MethodGet, "/v1/prompts/greeting-prompt", nil)

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result domain.Prompt
		respBody, _ := io.ReadAll(resp.Body)
		err = json.Unmarshal(respBody, &result)
		require.NoError(t, err)
		assert.Equal(t, "greeting-prompt", result.Name)

		mockSvc.AssertExpectations(t)
	})

	t.Run("gets prompt by specific version", func(t *testing.T) {
		mockSvc := new(MockPromptService)
		projectID := uuid.New()
		app := setupPromptsTestApp(mockSvc, projectID)

		promptID := uuid.New()
		expectedPrompt := &domain.Prompt{
			ID:        promptID,
			ProjectID: projectID,
			Name:      "greeting-prompt",
			Type:      domain.PromptTypeText,
			LatestVersion: &domain.PromptVersion{
				ID:      uuid.New(),
				Version: 2,
				Content: "Hello {{name}} - version 2",
			},
		}

		mockSvc.On("GetByNameAndVersion", mock.Anything, projectID, "greeting-prompt", 2).
			Return(expectedPrompt, nil)

		req := httptest.NewRequest(http.MethodGet, "/v1/prompts/greeting-prompt?version=2", nil)

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		mockSvc.AssertExpectations(t)
	})

	t.Run("gets prompt by label", func(t *testing.T) {
		mockSvc := new(MockPromptService)
		projectID := uuid.New()
		app := setupPromptsTestApp(mockSvc, projectID)

		promptID := uuid.New()
		expectedPrompt := &domain.Prompt{
			ID:        promptID,
			ProjectID: projectID,
			Name:      "greeting-prompt",
			Type:      domain.PromptTypeText,
			LatestVersion: &domain.PromptVersion{
				ID:      uuid.New(),
				Version: 1,
				Content: "Hello {{name}}",
				Labels:  []string{"production"},
			},
		}

		mockSvc.On("GetByNameAndLabel", mock.Anything, projectID, "greeting-prompt", "production").
			Return(expectedPrompt, nil)

		req := httptest.NewRequest(http.MethodGet, "/v1/prompts/greeting-prompt?label=production", nil)

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		mockSvc.AssertExpectations(t)
	})

	t.Run("returns 404 for non-existent prompt", func(t *testing.T) {
		mockSvc := new(MockPromptService)
		projectID := uuid.New()
		app := setupPromptsTestApp(mockSvc, projectID)

		mockSvc.On("GetByName", mock.Anything, projectID, "non-existent").
			Return(nil, assert.AnError)

		req := httptest.NewRequest(http.MethodGet, "/v1/prompts/non-existent", nil)

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)

		mockSvc.AssertExpectations(t)
	})
}

func TestPromptsHandler_CreatePrompt(t *testing.T) {
	t.Run("successfully creates prompt", func(t *testing.T) {
		mockSvc := new(MockPromptService)
		projectID := uuid.New()
		app := setupPromptsTestApp(mockSvc, projectID)

		promptID := uuid.New()
		expectedPrompt := &domain.Prompt{
			ID:          promptID,
			ProjectID:   projectID,
			Name:        "new-prompt",
			Type:        domain.PromptTypeText,
			Description: "A new prompt",
			CreatedAt:   time.Now(),
		}

		mockSvc.On("Create", mock.Anything, projectID, mock.AnythingOfType("*domain.PromptInput"), uuid.Nil).
			Return(expectedPrompt, nil)

		body := map[string]interface{}{
			"name":        "new-prompt",
			"content":     "Hello {{name}}",
			"description": "A new prompt",
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/v1/prompts", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var result domain.Prompt
		respBody, _ := io.ReadAll(resp.Body)
		err = json.Unmarshal(respBody, &result)
		require.NoError(t, err)
		assert.Equal(t, "new-prompt", result.Name)

		mockSvc.AssertExpectations(t)
	})

	t.Run("returns 400 for missing name", func(t *testing.T) {
		mockSvc := new(MockPromptService)
		projectID := uuid.New()
		app := setupPromptsTestApp(mockSvc, projectID)

		body := map[string]interface{}{
			"content": "Hello {{name}}",
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/v1/prompts", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("returns 400 for missing content", func(t *testing.T) {
		mockSvc := new(MockPromptService)
		projectID := uuid.New()
		app := setupPromptsTestApp(mockSvc, projectID)

		body := map[string]interface{}{
			"name": "new-prompt",
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/v1/prompts", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

func TestPromptsHandler_UpdatePrompt(t *testing.T) {
	t.Run("successfully updates prompt", func(t *testing.T) {
		mockSvc := new(MockPromptService)
		projectID := uuid.New()
		app := setupPromptsTestApp(mockSvc, projectID)

		promptID := uuid.New()
		expectedPrompt := &domain.Prompt{
			ID:          promptID,
			ProjectID:   projectID,
			Name:        "updated-prompt",
			Type:        domain.PromptTypeText,
			Description: "Updated description",
			UpdatedAt:   time.Now(),
		}

		mockSvc.On("Update", mock.Anything, promptID, mock.AnythingOfType("*domain.PromptInput")).
			Return(expectedPrompt, nil)

		body := map[string]interface{}{
			"description": "Updated description",
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPatch, "/v1/prompts/"+promptID.String(), bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result domain.Prompt
		respBody, _ := io.ReadAll(resp.Body)
		err = json.Unmarshal(respBody, &result)
		require.NoError(t, err)
		assert.Equal(t, "Updated description", result.Description)

		mockSvc.AssertExpectations(t)
	})

	t.Run("returns 400 for invalid prompt ID", func(t *testing.T) {
		mockSvc := new(MockPromptService)
		projectID := uuid.New()
		app := setupPromptsTestApp(mockSvc, projectID)

		body := map[string]interface{}{
			"description": "Updated description",
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPatch, "/v1/prompts/invalid-id", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

func TestPromptsHandler_DeletePrompt(t *testing.T) {
	t.Run("successfully deletes prompt", func(t *testing.T) {
		mockSvc := new(MockPromptService)
		projectID := uuid.New()
		app := setupPromptsTestApp(mockSvc, projectID)

		promptID := uuid.New()
		mockSvc.On("Delete", mock.Anything, promptID).
			Return(nil)

		req := httptest.NewRequest(http.MethodDelete, "/v1/prompts/"+promptID.String(), nil)

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNoContent, resp.StatusCode)

		mockSvc.AssertExpectations(t)
	})

	t.Run("returns 404 for non-existent prompt", func(t *testing.T) {
		mockSvc := new(MockPromptService)
		projectID := uuid.New()
		app := setupPromptsTestApp(mockSvc, projectID)

		promptID := uuid.New()
		mockSvc.On("Delete", mock.Anything, promptID).
			Return(assert.AnError)

		req := httptest.NewRequest(http.MethodDelete, "/v1/prompts/"+promptID.String(), nil)

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)

		mockSvc.AssertExpectations(t)
	})
}

func TestPromptsHandler_ListVersions(t *testing.T) {
	t.Run("successfully lists versions", func(t *testing.T) {
		mockSvc := new(MockPromptService)
		projectID := uuid.New()
		app := setupPromptsTestApp(mockSvc, projectID)

		promptID := uuid.New()
		prompt := &domain.Prompt{
			ID:        promptID,
			ProjectID: projectID,
			Name:      "greeting-prompt",
		}

		expectedVersions := []domain.PromptVersion{
			{
				ID:       uuid.New(),
				PromptID: promptID,
				Version:  1,
				Content:  "Hello {{name}}",
				Labels:   []string{"staging"},
			},
			{
				ID:       uuid.New(),
				PromptID: promptID,
				Version:  2,
				Content:  "Hi {{name}}",
				Labels:   []string{"production"},
			},
		}

		mockSvc.On("GetByName", mock.Anything, projectID, "greeting-prompt").
			Return(prompt, nil)
		mockSvc.On("ListVersions", mock.Anything, promptID).
			Return(expectedVersions, nil)

		req := httptest.NewRequest(http.MethodGet, "/v1/prompts/greeting-prompt/versions", nil)

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		respBody, _ := io.ReadAll(resp.Body)
		err = json.Unmarshal(respBody, &result)
		require.NoError(t, err)
		assert.Equal(t, float64(2), result["totalCount"])

		mockSvc.AssertExpectations(t)
	})
}

func TestPromptsHandler_CompilePrompt(t *testing.T) {
	t.Run("successfully compiles prompt", func(t *testing.T) {
		mockSvc := new(MockPromptService)
		projectID := uuid.New()
		app := setupPromptsTestApp(mockSvc, projectID)

		promptID := uuid.New()
		prompt := &domain.Prompt{
			ID:        promptID,
			ProjectID: projectID,
			Name:      "greeting-prompt",
			LatestVersion: &domain.PromptVersion{
				ID:       uuid.New(),
				PromptID: promptID,
				Version:  1,
				Content:  "Hello {{name}}",
			},
		}

		expectedCompiled := &service.CompiledPrompt{
			Prompt:    prompt,
			Version:   1,
			Compiled:  "Hello World",
			Variables: map[string]string{"name": "World"},
		}

		mockSvc.On("Compile", mock.Anything, projectID, "greeting-prompt", map[string]string{"name": "World"}, mock.AnythingOfType("*service.CompileOptions")).
			Return(expectedCompiled, nil)

		body := map[string]interface{}{
			"variables": map[string]string{"name": "World"},
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/v1/prompts/greeting-prompt/compile", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result service.CompiledPrompt
		respBody, _ := io.ReadAll(resp.Body)
		err = json.Unmarshal(respBody, &result)
		require.NoError(t, err)
		assert.Equal(t, "Hello World", result.Compiled)

		mockSvc.AssertExpectations(t)
	})

	t.Run("returns 404 for non-existent prompt", func(t *testing.T) {
		mockSvc := new(MockPromptService)
		projectID := uuid.New()
		app := setupPromptsTestApp(mockSvc, projectID)

		mockSvc.On("Compile", mock.Anything, projectID, "non-existent", map[string]string{}, mock.AnythingOfType("*service.CompileOptions")).
			Return(nil, assert.AnError)

		body := map[string]interface{}{
			"variables": map[string]string{},
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/v1/prompts/non-existent/compile", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)

		mockSvc.AssertExpectations(t)
	})
}
