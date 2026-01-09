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

// PromptsHandler handles prompt endpoints
type PromptsHandler struct {
	promptService *service.PromptService
	logger        *zap.Logger
}

// NewPromptsHandler creates a new prompts handler
func NewPromptsHandler(promptService *service.PromptService, logger *zap.Logger) *PromptsHandler {
	return &PromptsHandler{
		promptService: promptService,
		logger:        logger,
	}
}

// ListPrompts handles GET /v1/prompts
func (h *PromptsHandler) ListPrompts(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	filter := h.parsePromptFilter(c)
	filter.ProjectID = projectID
	limit := parseIntParam(c, "limit", 50)
	offset := parseIntParam(c, "offset", 0)

	if limit > 100 {
		limit = 100
	}

	list, err := h.promptService.List(c.Context(), filter, limit, offset)
	if err != nil {
		h.logger.Error("failed to list prompts", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to list prompts",
		})
	}

	return c.JSON(list)
}

// GetPrompt handles GET /v1/prompts/:promptName
func (h *PromptsHandler) GetPrompt(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	promptName := c.Params("promptName")
	if promptName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Prompt name required",
		})
	}

	// Check for version or label query params
	version := parseIntParam(c, "version", 0)
	label := c.Query("label")

	var prompt *domain.Prompt
	var err error

	if version > 0 {
		prompt, err = h.promptService.GetByNameAndVersion(c.Context(), projectID, promptName, version)
	} else if label != "" {
		prompt, err = h.promptService.GetByNameAndLabel(c.Context(), projectID, promptName, label)
	} else {
		// Default to production label or latest
		prompt, err = h.promptService.GetByNameAndLabel(c.Context(), projectID, promptName, "production")
		if apperrors.IsNotFound(err) {
			prompt, err = h.promptService.GetByName(c.Context(), projectID, promptName)
		}
	}

	if err != nil {
		if apperrors.IsNotFound(err) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   "Not Found",
				"message": "Prompt not found",
			})
		}
		h.logger.Error("failed to get prompt", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to get prompt",
		})
	}

	return c.JSON(prompt)
}

// CreatePrompt handles POST /v1/prompts
func (h *PromptsHandler) CreatePrompt(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	var input domain.PromptInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid request body: " + err.Error(),
		})
	}

	// Validation
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

	// Get user ID (required for creating prompts)
	userID, ok := middleware.GetUserID(c)
	if !ok {
		// For API key auth, use a system user ID
		userID = uuid.Nil
	}

	prompt, err := h.promptService.Create(c.Context(), projectID, &input, userID)
	if err != nil {
		if apperrors.IsValidation(err) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Bad Request",
				"message": err.Error(),
			})
		}
		h.logger.Error("failed to create prompt", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to create prompt",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(prompt)
}

// UpdatePrompt handles PATCH /v1/prompts/:promptId
func (h *PromptsHandler) UpdatePrompt(c *fiber.Ctx) error {
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
			"message": "Invalid request body: " + err.Error(),
		})
	}

	prompt, err := h.promptService.Update(c.Context(), promptID, &input)
	if err != nil {
		if apperrors.IsNotFound(err) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   "Not Found",
				"message": "Prompt not found",
			})
		}
		h.logger.Error("failed to update prompt", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to update prompt",
		})
	}

	return c.JSON(prompt)
}

// DeletePrompt handles DELETE /v1/prompts/:promptId
func (h *PromptsHandler) DeletePrompt(c *fiber.Ctx) error {
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

	if err := h.promptService.Delete(c.Context(), promptID); err != nil {
		if apperrors.IsNotFound(err) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   "Not Found",
				"message": "Prompt not found",
			})
		}
		h.logger.Error("failed to delete prompt", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to delete prompt",
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// ListVersions handles GET /v1/prompts/:promptName/versions
func (h *PromptsHandler) ListVersions(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	promptName := c.Params("promptName")
	if promptName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Prompt name required",
		})
	}

	// First get the prompt by name to get the prompt ID
	prompt, err := h.promptService.GetByName(c.Context(), projectID, promptName)
	if err != nil {
		if apperrors.IsNotFound(err) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   "Not Found",
				"message": "Prompt not found",
			})
		}
		h.logger.Error("failed to get prompt", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to get prompt",
		})
	}

	versions, err := h.promptService.ListVersions(c.Context(), prompt.ID)
	if err != nil {
		h.logger.Error("failed to list versions", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to list versions",
		})
	}

	return c.JSON(fiber.Map{
		"data":       versions,
		"totalCount": len(versions),
	})
}

// CreateVersion handles POST /v1/prompts/:promptName/versions
func (h *PromptsHandler) CreateVersion(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	promptName := c.Params("promptName")
	if promptName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Prompt name required",
		})
	}

	// First get the prompt by name to get the prompt ID
	prompt, err := h.promptService.GetByName(c.Context(), projectID, promptName)
	if err != nil {
		if apperrors.IsNotFound(err) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   "Not Found",
				"message": "Prompt not found",
			})
		}
		h.logger.Error("failed to get prompt", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to get prompt",
		})
	}

	var input domain.PromptVersionInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid request body: " + err.Error(),
		})
	}

	// Get user ID
	userID, ok := middleware.GetUserID(c)
	if !ok {
		userID = uuid.Nil
	}

	version, err := h.promptService.CreateVersion(c.Context(), prompt.ID, &input, userID)
	if err != nil {
		h.logger.Error("failed to create version", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to create version",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(version)
}

// SetLabel handles POST /v1/prompts/:promptName/labels
func (h *PromptsHandler) SetLabel(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	promptName := c.Params("promptName")
	if promptName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Prompt name required",
		})
	}

	// First get the prompt by name to get the prompt ID
	prompt, err := h.promptService.GetByName(c.Context(), projectID, promptName)
	if err != nil {
		if apperrors.IsNotFound(err) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   "Not Found",
				"message": "Prompt not found",
			})
		}
		h.logger.Error("failed to get prompt", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to get prompt",
		})
	}

	var input struct {
		Label   string `json:"label"`
		Version int    `json:"version"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid request body: " + err.Error(),
		})
	}

	if input.Label == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "label is required",
		})
	}

	if input.Version <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "version must be positive",
		})
	}

	if err := h.promptService.SetVersionLabel(c.Context(), prompt.ID, input.Version, input.Label, true); err != nil {
		if apperrors.IsNotFound(err) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   "Not Found",
				"message": "Prompt or version not found",
			})
		}
		h.logger.Error("failed to set label", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to set label",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Label set successfully",
	})
}

// RemoveLabel handles DELETE /v1/prompts/:promptName/versions/:version/labels/:label
func (h *PromptsHandler) RemoveLabel(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	promptName := c.Params("promptName")
	label := c.Params("label")
	version := parseIntParam(c, "version", 0)

	if promptName == "" || label == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Prompt name and label required",
		})
	}

	// First get the prompt by name to get the prompt ID
	prompt, err := h.promptService.GetByName(c.Context(), projectID, promptName)
	if err != nil {
		if apperrors.IsNotFound(err) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   "Not Found",
				"message": "Prompt not found",
			})
		}
		h.logger.Error("failed to get prompt", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to get prompt",
		})
	}

	if err := h.promptService.SetVersionLabel(c.Context(), prompt.ID, version, label, false); err != nil {
		if apperrors.IsNotFound(err) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   "Not Found",
				"message": "Label not found",
			})
		}
		h.logger.Error("failed to remove label", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to remove label",
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// CompilePrompt handles POST /v1/prompts/:promptName/compile
func (h *PromptsHandler) CompilePrompt(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
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
			"message": "Invalid request body: " + err.Error(),
		})
	}

	options := &service.CompileOptions{
		Version: input.Version,
	}
	if input.Label != nil {
		options.Label = *input.Label
	}

	compiled, err := h.promptService.Compile(c.Context(), projectID, promptName, input.Variables, options)
	if err != nil {
		if apperrors.IsNotFound(err) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   "Not Found",
				"message": "Prompt not found",
			})
		}
		h.logger.Error("failed to compile prompt", zap.Error(err))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Failed to compile prompt: " + err.Error(),
		})
	}

	return c.JSON(compiled)
}

// parsePromptFilter parses prompt filter from query params
func (h *PromptsHandler) parsePromptFilter(c *fiber.Ctx) *domain.PromptFilter {
	filter := &domain.PromptFilter{}

	if name := c.Query("name"); name != "" {
		filter.Name = &name
	}

	if label := c.Query("label"); label != "" {
		filter.Label = &label
	}

	if tags := c.Query("tags"); tags != "" {
		filter.Tags = []string{tags}
	}

	return filter
}

// RegisterRoutes registers prompt routes
func (h *PromptsHandler) RegisterRoutes(app *fiber.App, authMiddleware *middleware.AuthMiddleware) {
	v1 := app.Group("/v1", authMiddleware.RequireAPIKey())

	v1.Get("/prompts", h.ListPrompts)
	v1.Get("/prompts/:promptName", h.GetPrompt)
	v1.Post("/prompts", h.CreatePrompt)
	v1.Patch("/prompts/:promptId", h.UpdatePrompt)
	v1.Delete("/prompts/:promptId", h.DeletePrompt)

	v1.Get("/prompts/:promptName/versions", h.ListVersions)
	v1.Post("/prompts/:promptName/versions", h.CreateVersion)
	v1.Post("/prompts/:promptName/labels", h.SetLabel)
	v1.Delete("/prompts/:promptName/versions/:version/labels/:label", h.RemoveLabel)
	v1.Post("/prompts/:promptName/compile", h.CompilePrompt)
}
