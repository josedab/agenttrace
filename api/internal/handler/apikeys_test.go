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
	"github.com/agenttrace/agenttrace/api/internal/middleware"
	apperrors "github.com/agenttrace/agenttrace/api/internal/pkg/errors"
	"github.com/agenttrace/agenttrace/api/internal/testutil"
)

// MockAPIKeyService mocks the auth service methods for API key operations.
type MockAPIKeyService struct {
	mock.Mock
}

func (m *MockAPIKeyService) ListAPIKeys(ctx context.Context, projectID uuid.UUID) ([]domain.APIKey, error) {
	args := m.Called(ctx, projectID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.APIKey), args.Error(1)
}

func (m *MockAPIKeyService) CreateAPIKey(ctx context.Context, projectID uuid.UUID, input *domain.APIKeyInput, userID uuid.UUID) (*domain.APIKeyCreateResult, error) {
	args := m.Called(ctx, projectID, input, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.APIKeyCreateResult), args.Error(1)
}

func (m *MockAPIKeyService) DeleteAPIKey(ctx context.Context, keyID uuid.UUID) error {
	args := m.Called(ctx, keyID)
	return args.Error(0)
}

func setupAPIKeysTestApp(mockSvc *MockAPIKeyService, projectID, userID *uuid.UUID) *fiber.App {
	app := fiber.New()
	logger := zap.NewNop()

	// Apply middleware based on what's provided
	if projectID != nil {
		app.Use(testutil.TestProjectMiddleware(*projectID))
	}
	if userID != nil {
		app.Use(testutil.TestUserMiddleware(*userID))
	}

	// ListAPIKeys
	app.Get("/v1/api-keys", func(c *fiber.Ctx) error {
		pid, ok := middleware.GetProjectID(c)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "Unauthorized",
				"message": "Project ID not found",
			})
		}

		keys, err := mockSvc.ListAPIKeys(c.Context(), pid)
		if err != nil {
			logger.Error("failed to list API keys", zap.Error(err))
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "Internal Server Error",
				"message": "Failed to list API keys",
			})
		}

		response := make([]fiber.Map, len(keys))
		for i, key := range keys {
			response[i] = fiber.Map{
				"id":               key.ID,
				"name":             key.Name,
				"publicKey":        key.PublicKey,
				"secretKeyPreview": key.SecretKeyPreview,
				"scopes":           key.Scopes,
				"expiresAt":        key.ExpiresAt,
				"lastUsedAt":       key.LastUsedAt,
				"createdAt":        key.CreatedAt,
			}
		}

		return c.JSON(fiber.Map{
			"data": response,
		})
	})

	// CreateAPIKey
	app.Post("/v1/api-keys", func(c *fiber.Ctx) error {
		pid, ok := middleware.GetProjectID(c)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "Unauthorized",
				"message": "Project ID not found",
			})
		}

		uid, ok := middleware.GetUserID(c)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "Unauthorized",
				"message": "User ID not found",
			})
		}

		var input domain.APIKeyInput
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

		result, err := mockSvc.CreateAPIKey(c.Context(), pid, &input, uid)
		if err != nil {
			logger.Error("failed to create API key", zap.Error(err))
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "Internal Server Error",
				"message": "Failed to create API key",
			})
		}

		return c.Status(fiber.StatusCreated).JSON(fiber.Map{
			"id":               result.APIKey.ID,
			"name":             result.APIKey.Name,
			"publicKey":        result.APIKey.PublicKey,
			"secretKey":        result.SecretKey,
			"secretKeyPreview": result.APIKey.SecretKeyPreview,
			"scopes":           result.APIKey.Scopes,
			"expiresAt":        result.APIKey.ExpiresAt,
			"createdAt":        result.APIKey.CreatedAt,
			"note":             "This is the only time the full secret key will be shown. Please save it securely.",
		})
	})

	// DeleteAPIKey
	app.Delete("/v1/api-keys/:keyId", func(c *fiber.Ctx) error {
		keyID, err := uuid.Parse(c.Params("keyId"))
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Bad Request",
				"message": "Invalid key ID",
			})
		}

		if err := mockSvc.DeleteAPIKey(c.Context(), keyID); err != nil {
			if apperrors.IsNotFound(err) {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"error":   "Not Found",
					"message": "API key not found",
				})
			}
			logger.Error("failed to delete API key", zap.Error(err))
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "Internal Server Error",
				"message": "Failed to delete API key",
			})
		}

		return c.SendStatus(fiber.StatusNoContent)
	})

	return app
}

// --- ListAPIKeys Tests ---

func TestAPIKeysHandler_ListAPIKeys(t *testing.T) {
	t.Run("successfully lists API keys", func(t *testing.T) {
		mockSvc := new(MockAPIKeyService)
		projectID := uuid.New()
		userID := uuid.New()
		app := setupAPIKeysTestApp(mockSvc, &projectID, &userID)

		now := time.Now()
		expectedKeys := []domain.APIKey{
			{
				ID:               uuid.New(),
				Name:             "Production Key",
				PublicKey:        "pk-at-prod123",
				SecretKeyPreview: "sk-at-...xyz",
				Scopes:           []string{"traces:read", "traces:write"},
				CreatedAt:        now,
			},
			{
				ID:               uuid.New(),
				Name:             "Development Key",
				PublicKey:        "pk-at-dev456",
				SecretKeyPreview: "sk-at-...abc",
				Scopes:           []string{"traces:read"},
				CreatedAt:        now,
			},
		}

		mockSvc.On("ListAPIKeys", mock.Anything, projectID).Return(expectedKeys, nil)

		req := httptest.NewRequest(http.MethodGet, "/v1/api-keys", nil)

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)

		data := result["data"].([]interface{})
		assert.Len(t, data, 2)

		mockSvc.AssertExpectations(t)
	})

	t.Run("returns empty list when no keys exist", func(t *testing.T) {
		mockSvc := new(MockAPIKeyService)
		projectID := uuid.New()
		userID := uuid.New()
		app := setupAPIKeysTestApp(mockSvc, &projectID, &userID)

		mockSvc.On("ListAPIKeys", mock.Anything, projectID).Return([]domain.APIKey{}, nil)

		req := httptest.NewRequest(http.MethodGet, "/v1/api-keys", nil)

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)

		data := result["data"].([]interface{})
		assert.Len(t, data, 0)

		mockSvc.AssertExpectations(t)
	})

	t.Run("returns 401 when project ID not found", func(t *testing.T) {
		mockSvc := new(MockAPIKeyService)
		app := setupAPIKeysTestApp(mockSvc, nil, nil) // No project ID

		req := httptest.NewRequest(http.MethodGet, "/v1/api-keys", nil)

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("returns 500 on service error", func(t *testing.T) {
		mockSvc := new(MockAPIKeyService)
		projectID := uuid.New()
		userID := uuid.New()
		app := setupAPIKeysTestApp(mockSvc, &projectID, &userID)

		mockSvc.On("ListAPIKeys", mock.Anything, projectID).Return(nil, assert.AnError)

		req := httptest.NewRequest(http.MethodGet, "/v1/api-keys", nil)

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

		mockSvc.AssertExpectations(t)
	})
}

// --- CreateAPIKey Tests ---

func TestAPIKeysHandler_CreateAPIKey(t *testing.T) {
	t.Run("successfully creates API key", func(t *testing.T) {
		mockSvc := new(MockAPIKeyService)
		projectID := uuid.New()
		userID := uuid.New()
		app := setupAPIKeysTestApp(mockSvc, &projectID, &userID)

		keyID := uuid.New()
		expectedResult := &domain.APIKeyCreateResult{
			APIKey: &domain.APIKey{
				ID:               keyID,
				Name:             "New Production Key",
				PublicKey:        "pk-at-newkey123",
				SecretKeyPreview: "sk-at-...new",
				Scopes:           []string{"traces:read", "traces:write"},
				CreatedAt:        time.Now(),
			},
			SecretKey: "sk-at-full-secret-key-here",
		}

		mockSvc.On("CreateAPIKey", mock.Anything, projectID, mock.MatchedBy(func(input *domain.APIKeyInput) bool {
			return input.Name == "New Production Key"
		}), userID).Return(expectedResult, nil)

		body, _ := json.Marshal(map[string]interface{}{
			"name":   "New Production Key",
			"scopes": []string{"traces:read", "traces:write"},
		})
		req := httptest.NewRequest(http.MethodPost, "/v1/api-keys", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)

		assert.Equal(t, "New Production Key", result["name"])
		assert.Equal(t, "pk-at-newkey123", result["publicKey"])
		assert.Equal(t, "sk-at-full-secret-key-here", result["secretKey"])
		assert.Contains(t, result["note"].(string), "only time the full secret key")

		mockSvc.AssertExpectations(t)
	})

	t.Run("returns 400 for missing name", func(t *testing.T) {
		mockSvc := new(MockAPIKeyService)
		projectID := uuid.New()
		userID := uuid.New()
		app := setupAPIKeysTestApp(mockSvc, &projectID, &userID)

		body, _ := json.Marshal(map[string]interface{}{
			"scopes": []string{"traces:read"},
		})
		req := httptest.NewRequest(http.MethodPost, "/v1/api-keys", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		assert.Equal(t, "name is required", result["message"])
	})

	t.Run("returns 400 for invalid JSON", func(t *testing.T) {
		mockSvc := new(MockAPIKeyService)
		projectID := uuid.New()
		userID := uuid.New()
		app := setupAPIKeysTestApp(mockSvc, &projectID, &userID)

		req := httptest.NewRequest(http.MethodPost, "/v1/api-keys", bytes.NewReader([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("returns 401 when project ID not found", func(t *testing.T) {
		mockSvc := new(MockAPIKeyService)
		userID := uuid.New()
		app := setupAPIKeysTestApp(mockSvc, nil, &userID) // No project ID

		body, _ := json.Marshal(map[string]interface{}{
			"name": "Test Key",
		})
		req := httptest.NewRequest(http.MethodPost, "/v1/api-keys", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("returns 401 when user ID not found", func(t *testing.T) {
		mockSvc := new(MockAPIKeyService)
		projectID := uuid.New()
		app := setupAPIKeysTestApp(mockSvc, &projectID, nil) // No user ID

		body, _ := json.Marshal(map[string]interface{}{
			"name": "Test Key",
		})
		req := httptest.NewRequest(http.MethodPost, "/v1/api-keys", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("returns 500 on service error", func(t *testing.T) {
		mockSvc := new(MockAPIKeyService)
		projectID := uuid.New()
		userID := uuid.New()
		app := setupAPIKeysTestApp(mockSvc, &projectID, &userID)

		mockSvc.On("CreateAPIKey", mock.Anything, projectID, mock.Anything, userID).
			Return(nil, assert.AnError)

		body, _ := json.Marshal(map[string]interface{}{
			"name": "Test Key",
		})
		req := httptest.NewRequest(http.MethodPost, "/v1/api-keys", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

		mockSvc.AssertExpectations(t)
	})
}

// --- DeleteAPIKey Tests ---

func TestAPIKeysHandler_DeleteAPIKey(t *testing.T) {
	t.Run("successfully deletes API key", func(t *testing.T) {
		mockSvc := new(MockAPIKeyService)
		projectID := uuid.New()
		userID := uuid.New()
		app := setupAPIKeysTestApp(mockSvc, &projectID, &userID)

		keyID := uuid.New()
		mockSvc.On("DeleteAPIKey", mock.Anything, keyID).Return(nil)

		req := httptest.NewRequest(http.MethodDelete, "/v1/api-keys/"+keyID.String(), nil)

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNoContent, resp.StatusCode)

		mockSvc.AssertExpectations(t)
	})

	t.Run("returns 400 for invalid key ID", func(t *testing.T) {
		mockSvc := new(MockAPIKeyService)
		projectID := uuid.New()
		userID := uuid.New()
		app := setupAPIKeysTestApp(mockSvc, &projectID, &userID)

		req := httptest.NewRequest(http.MethodDelete, "/v1/api-keys/not-a-uuid", nil)

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		assert.Equal(t, "Invalid key ID", result["message"])
	})

	t.Run("returns 404 when key not found", func(t *testing.T) {
		mockSvc := new(MockAPIKeyService)
		projectID := uuid.New()
		userID := uuid.New()
		app := setupAPIKeysTestApp(mockSvc, &projectID, &userID)

		keyID := uuid.New()
		mockSvc.On("DeleteAPIKey", mock.Anything, keyID).
			Return(apperrors.NotFound("API key not found"))

		req := httptest.NewRequest(http.MethodDelete, "/v1/api-keys/"+keyID.String(), nil)

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)

		mockSvc.AssertExpectations(t)
	})

	t.Run("returns 500 on service error", func(t *testing.T) {
		mockSvc := new(MockAPIKeyService)
		projectID := uuid.New()
		userID := uuid.New()
		app := setupAPIKeysTestApp(mockSvc, &projectID, &userID)

		keyID := uuid.New()
		mockSvc.On("DeleteAPIKey", mock.Anything, keyID).Return(assert.AnError)

		req := httptest.NewRequest(http.MethodDelete, "/v1/api-keys/"+keyID.String(), nil)

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

		mockSvc.AssertExpectations(t)
	})
}
