package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/middleware"
	apperrors "github.com/agenttrace/agenttrace/api/internal/pkg/errors"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// APIKeysHandler handles API key endpoints
type APIKeysHandler struct {
	authService *service.AuthService
	logger      *zap.Logger
}

// NewAPIKeysHandler creates a new API keys handler
func NewAPIKeysHandler(authService *service.AuthService, logger *zap.Logger) *APIKeysHandler {
	return &APIKeysHandler{
		authService: authService,
		logger:      logger,
	}
}

// ListAPIKeys handles GET /v1/api-keys
func (h *APIKeysHandler) ListAPIKeys(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	keys, err := h.authService.ListAPIKeys(c.Context(), projectID)
	if err != nil {
		h.logger.Error("failed to list API keys", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to list API keys",
		})
	}

	// Transform to response format
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
}

// CreateAPIKey handles POST /v1/api-keys
func (h *APIKeysHandler) CreateAPIKey(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	userID, ok := middleware.GetUserID(c)
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

	result, err := h.authService.CreateAPIKey(c.Context(), projectID, &input, userID)
	if err != nil {
		h.logger.Error("failed to create API key", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to create API key",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"id":               result.APIKey.ID,
		"name":             result.APIKey.Name,
		"publicKey":        result.APIKey.PublicKey,
		"secretKey":        result.SecretKey, // Only returned on creation
		"secretKeyPreview": result.APIKey.SecretKeyPreview,
		"scopes":           result.APIKey.Scopes,
		"expiresAt":        result.APIKey.ExpiresAt,
		"createdAt":        result.APIKey.CreatedAt,
		"note":             "This is the only time the full secret key will be shown. Please save it securely.",
	})
}

// DeleteAPIKey handles DELETE /v1/api-keys/:keyId
func (h *APIKeysHandler) DeleteAPIKey(c *fiber.Ctx) error {
	keyID, err := uuid.Parse(c.Params("keyId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid key ID",
		})
	}

	if err := h.authService.DeleteAPIKey(c.Context(), keyID); err != nil {
		if apperrors.IsNotFound(err) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   "Not Found",
				"message": "API key not found",
			})
		}
		h.logger.Error("failed to delete API key", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to delete API key",
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// RegisterRoutes registers API key routes
func (h *APIKeysHandler) RegisterRoutes(app *fiber.App, authMiddleware *middleware.AuthMiddleware) {
	// API key management requires JWT auth
	v1 := app.Group("/v1")

	// These routes need JWT authentication (user logged in via web)
	jwtAuth := v1.Group("", authMiddleware.RequireJWT(), authMiddleware.RequireProjectAccess())
	jwtAuth.Get("/projects/:projectId/api-keys", h.ListAPIKeys)
	jwtAuth.Post("/projects/:projectId/api-keys", h.CreateAPIKey)
	jwtAuth.Delete("/projects/:projectId/api-keys/:keyId", h.DeleteAPIKey)
}
