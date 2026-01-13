package handler

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// PromptLibraryHandler handles prompt library HTTP requests
type PromptLibraryHandler struct {
	logger               *zap.Logger
	promptLibraryService *service.PromptLibraryService
}

// NewPromptLibraryHandler creates a new prompt library handler
func NewPromptLibraryHandler(
	logger *zap.Logger,
	promptLibraryService *service.PromptLibraryService,
) *PromptLibraryHandler {
	return &PromptLibraryHandler{
		logger:               logger,
		promptLibraryService: promptLibraryService,
	}
}

// ListPrompts returns prompts from the library
// @Summary List library prompts
// @Description Get prompts from the community library
// @Tags prompt-library
// @Accept json
// @Produce json
// @Param category query string false "Filter by category"
// @Param visibility query string false "Filter by visibility"
// @Param tags query []string false "Filter by tags"
// @Param search query string false "Search by name or description"
// @Param sortBy query string false "Sort by: popular, recent, stars, usage"
// @Param limit query int false "Limit results" default(20)
// @Param offset query int false "Offset for pagination" default(0)
// @Success 200 {object} domain.LibraryPromptList
// @Failure 400 {object} ErrorResponse
// @Router /api/public/library/prompts [get]
func (h *PromptLibraryHandler) ListPrompts(c *fiber.Ctx) error {
	filter := domain.LibraryPromptFilter{
		Search:    c.Query("search"),
		SortBy:    c.Query("sortBy", "popular"),
		SortOrder: c.Query("sortOrder", "desc"),
	}

	if categoryStr := c.Query("category"); categoryStr != "" {
		cat := domain.PromptCategory(categoryStr)
		filter.Category = &cat
	}

	if visibilityStr := c.Query("visibility"); visibilityStr != "" {
		vis := domain.PromptVisibility(visibilityStr)
		filter.Visibility = &vis
	}

	tags := c.Query("tags")
	if tags != "" {
		filter.Tags = strings.Split(tags, ",")
	}

	h.logger.Debug("List library prompts",
		zap.String("search", filter.Search),
		zap.String("sortBy", filter.SortBy),
	)

	// Return empty list for now
	result := domain.LibraryPromptList{
		Prompts:    []domain.LibraryPrompt{},
		TotalCount: 0,
		HasMore:    false,
	}

	return c.JSON(result)
}

// GetPrompt returns a specific library prompt
// @Summary Get library prompt
// @Description Get a specific prompt from the library
// @Tags prompt-library
// @Accept json
// @Produce json
// @Param id path string true "Prompt ID"
// @Success 200 {object} domain.LibraryPrompt
// @Failure 404 {object} ErrorResponse
// @Router /api/public/library/prompts/{id} [get]
func (h *PromptLibraryHandler) GetPrompt(c *fiber.Ctx) error {
	promptID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid prompt ID")
	}

	h.logger.Debug("Get library prompt", zap.String("promptId", promptID.String()))

	// In real implementation, increment view count
	return errorResponse(c, fiber.StatusNotFound, "Prompt not found")
}

// GetPromptBySlug returns a library prompt by its slug
// @Summary Get prompt by slug
// @Description Get a prompt from the library by its URL-friendly slug
// @Tags prompt-library
// @Accept json
// @Produce json
// @Param slug path string true "Prompt slug"
// @Success 200 {object} domain.LibraryPrompt
// @Failure 404 {object} ErrorResponse
// @Router /api/public/library/prompts/slug/{slug} [get]
func (h *PromptLibraryHandler) GetPromptBySlug(c *fiber.Ctx) error {
	slug := c.Params("slug")
	if slug == "" {
		return errorResponse(c, fiber.StatusBadRequest, "Slug is required")
	}

	h.logger.Debug("Get library prompt by slug", zap.String("slug", slug))

	return errorResponse(c, fiber.StatusNotFound, "Prompt not found")
}

// CreatePrompt creates a new library prompt
// @Summary Create library prompt
// @Description Create a new prompt in the library
// @Tags prompt-library
// @Accept json
// @Produce json
// @Param prompt body domain.LibraryPromptInput true "Prompt configuration"
// @Success 201 {object} domain.LibraryPrompt
// @Failure 400 {object} ErrorResponse
// @Router /api/public/library/prompts [post]
func (h *PromptLibraryHandler) CreatePrompt(c *fiber.Ctx) error {
	userID := uuid.New()   // In real implementation, get from auth context
	userName := "user"     // In real implementation, get from auth context
	var projectID *uuid.UUID // Optional project association

	var input domain.LibraryPromptInput
	if err := c.BodyParser(&input); err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid request body")
	}

	prompt, err := h.promptLibraryService.CreatePrompt(c.Context(), userID, userName, projectID, &input)
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	return c.Status(fiber.StatusCreated).JSON(prompt)
}

// UpdatePrompt updates a library prompt
// @Summary Update library prompt
// @Description Update a prompt in the library
// @Tags prompt-library
// @Accept json
// @Produce json
// @Param id path string true "Prompt ID"
// @Param prompt body domain.LibraryPromptUpdateInput true "Updated configuration"
// @Success 200 {object} domain.LibraryPrompt
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/public/library/prompts/{id} [patch]
func (h *PromptLibraryHandler) UpdatePrompt(c *fiber.Ctx) error {
	promptID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid prompt ID")
	}

	var input domain.LibraryPromptUpdateInput
	if err := c.BodyParser(&input); err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid request body")
	}

	h.logger.Debug("Update library prompt", zap.String("promptId", promptID.String()))

	return errorResponse(c, fiber.StatusNotFound, "Prompt not found")
}

// DeletePrompt deletes a library prompt
// @Summary Delete library prompt
// @Description Delete a prompt from the library
// @Tags prompt-library
// @Accept json
// @Produce json
// @Param id path string true "Prompt ID"
// @Success 204 "No Content"
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/public/library/prompts/{id} [delete]
func (h *PromptLibraryHandler) DeletePrompt(c *fiber.Ctx) error {
	promptID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid prompt ID")
	}

	h.logger.Info("Delete library prompt", zap.String("promptId", promptID.String()))

	return errorResponse(c, fiber.StatusNotFound, "Prompt not found")
}

// ForkPrompt creates a fork of a library prompt
// @Summary Fork prompt
// @Description Create a fork of an existing library prompt
// @Tags prompt-library
// @Accept json
// @Produce json
// @Param id path string true "Prompt ID to fork"
// @Param body body domain.ForkInput true "Fork configuration"
// @Success 201 {object} domain.LibraryPrompt
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/public/library/prompts/{id}/fork [post]
func (h *PromptLibraryHandler) ForkPrompt(c *fiber.Ctx) error {
	promptID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid prompt ID")
	}

	var input domain.ForkInput
	if err := c.BodyParser(&input); err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid request body")
	}

	h.logger.Info("Fork library prompt", zap.String("promptId", promptID.String()))

	return errorResponse(c, fiber.StatusNotFound, "Prompt not found")
}

// PublishPrompt makes a prompt publicly visible
// @Summary Publish prompt
// @Description Make a prompt publicly visible in the library
// @Tags prompt-library
// @Accept json
// @Produce json
// @Param id path string true "Prompt ID"
// @Success 200 {object} domain.LibraryPrompt
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/public/library/prompts/{id}/publish [post]
func (h *PromptLibraryHandler) PublishPrompt(c *fiber.Ctx) error {
	promptID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid prompt ID")
	}

	h.logger.Info("Publish library prompt", zap.String("promptId", promptID.String()))

	return errorResponse(c, fiber.StatusNotFound, "Prompt not found")
}

// StarPrompt adds a star to a prompt
// @Summary Star prompt
// @Description Add a star to a library prompt
// @Tags prompt-library
// @Accept json
// @Produce json
// @Param id path string true "Prompt ID"
// @Success 200 {object} domain.PromptStar
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/public/library/prompts/{id}/star [post]
func (h *PromptLibraryHandler) StarPrompt(c *fiber.Ctx) error {
	promptID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid prompt ID")
	}

	userID := uuid.New() // In real implementation, get from auth context

	star, err := h.promptLibraryService.StarPrompt(c.Context(), promptID, userID)
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	return c.JSON(star)
}

// UnstarPrompt removes a star from a prompt
// @Summary Unstar prompt
// @Description Remove a star from a library prompt
// @Tags prompt-library
// @Accept json
// @Produce json
// @Param id path string true "Prompt ID"
// @Success 204 "No Content"
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/public/library/prompts/{id}/star [delete]
func (h *PromptLibraryHandler) UnstarPrompt(c *fiber.Ctx) error {
	promptID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid prompt ID")
	}

	userID := uuid.New() // In real implementation, get from auth context

	if err := h.promptLibraryService.UnstarPrompt(c.Context(), promptID, userID); err != nil {
		return errorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// GetVersions returns version history for a prompt
// @Summary Get prompt versions
// @Description Get version history for a library prompt
// @Tags prompt-library
// @Accept json
// @Produce json
// @Param id path string true "Prompt ID"
// @Param limit query int false "Limit results" default(20)
// @Param offset query int false "Offset for pagination" default(0)
// @Success 200 {object} domain.PromptVersionList
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/public/library/prompts/{id}/versions [get]
func (h *PromptLibraryHandler) GetVersions(c *fiber.Ctx) error {
	promptID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid prompt ID")
	}

	h.logger.Debug("Get prompt versions", zap.String("promptId", promptID.String()))

	result := domain.PromptVersionList{
		Versions:   []domain.PromptVersion{},
		TotalCount: 0,
		HasMore:    false,
	}

	return c.JSON(result)
}

// GetVersion returns a specific version of a prompt
// @Summary Get prompt version
// @Description Get a specific version of a library prompt
// @Tags prompt-library
// @Accept json
// @Produce json
// @Param id path string true "Prompt ID"
// @Param version path int true "Version number"
// @Success 200 {object} domain.PromptVersion
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/public/library/prompts/{id}/versions/{version} [get]
func (h *PromptLibraryHandler) GetVersion(c *fiber.Ctx) error {
	promptID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid prompt ID")
	}

	version, err := c.ParamsInt("version")
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid version number")
	}

	h.logger.Debug("Get prompt version",
		zap.String("promptId", promptID.String()),
		zap.Int("version", version),
	)

	return errorResponse(c, fiber.StatusNotFound, "Version not found")
}

// RunBenchmark runs a benchmark for a prompt
// @Summary Run benchmark
// @Description Run a benchmark for a library prompt
// @Tags prompt-library
// @Accept json
// @Produce json
// @Param id path string true "Prompt ID"
// @Param body body domain.BenchmarkInput true "Benchmark configuration"
// @Success 201 {object} domain.PromptBenchmark
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/public/library/prompts/{id}/benchmark [post]
func (h *PromptLibraryHandler) RunBenchmark(c *fiber.Ctx) error {
	promptID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid prompt ID")
	}

	var input domain.BenchmarkInput
	if err := c.BodyParser(&input); err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid request body")
	}

	if input.Model == "" {
		return errorResponse(c, fiber.StatusBadRequest, "Model is required")
	}

	h.logger.Info("Run prompt benchmark",
		zap.String("promptId", promptID.String()),
		zap.String("model", input.Model),
	)

	return errorResponse(c, fiber.StatusNotFound, "Prompt not found")
}

// GetBenchmarks returns benchmarks for a prompt
// @Summary Get benchmarks
// @Description Get benchmark results for a library prompt
// @Tags prompt-library
// @Accept json
// @Produce json
// @Param id path string true "Prompt ID"
// @Success 200 {array} domain.PromptBenchmark
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/public/library/prompts/{id}/benchmarks [get]
func (h *PromptLibraryHandler) GetBenchmarks(c *fiber.Ctx) error {
	promptID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid prompt ID")
	}

	h.logger.Debug("Get prompt benchmarks", zap.String("promptId", promptID.String()))

	return c.JSON([]domain.PromptBenchmark{})
}

// RenderPrompt renders a prompt template with variables
// @Summary Render prompt
// @Description Render a prompt template with provided variables
// @Tags prompt-library
// @Accept json
// @Produce json
// @Param id path string true "Prompt ID"
// @Param body body RenderPromptRequest true "Variables for rendering"
// @Success 200 {object} RenderPromptResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/public/library/prompts/{id}/render [post]
func (h *PromptLibraryHandler) RenderPrompt(c *fiber.Ctx) error {
	promptID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid prompt ID")
	}

	var req RenderPromptRequest
	if err := c.BodyParser(&req); err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid request body")
	}

	h.logger.Debug("Render prompt", zap.String("promptId", promptID.String()))

	return errorResponse(c, fiber.StatusNotFound, "Prompt not found")
}

// RenderPromptRequest represents the request to render a prompt
type RenderPromptRequest struct {
	Variables map[string]any `json:"variables"`
	Version   *int           `json:"version,omitempty"`
}

// RenderPromptResponse represents the response from rendering a prompt
type RenderPromptResponse struct {
	RenderedPrompt string `json:"renderedPrompt"`
	PromptID       string `json:"promptId"`
	Version        int    `json:"version"`
}

// GetCategories returns available prompt categories
// @Summary Get categories
// @Description Get available prompt categories
// @Tags prompt-library
// @Accept json
// @Produce json
// @Success 200 {array} CategoryInfo
// @Router /api/public/library/categories [get]
func (h *PromptLibraryHandler) GetCategories(c *fiber.Ctx) error {
	categories := []CategoryInfo{
		{ID: string(domain.PromptCategoryAgent), Name: "Agent", Description: "Prompts for autonomous agents"},
		{ID: string(domain.PromptCategoryChat), Name: "Chat", Description: "Conversational prompts"},
		{ID: string(domain.PromptCategoryCompletion), Name: "Completion", Description: "Text completion prompts"},
		{ID: string(domain.PromptCategorySummarization), Name: "Summarization", Description: "Document summarization"},
		{ID: string(domain.PromptCategoryExtraction), Name: "Extraction", Description: "Information extraction"},
		{ID: string(domain.PromptCategoryClassification), Name: "Classification", Description: "Text classification"},
		{ID: string(domain.PromptCategoryCodeGen), Name: "Code Generation", Description: "Code generation prompts"},
		{ID: string(domain.PromptCategoryTranslation), Name: "Translation", Description: "Language translation"},
		{ID: string(domain.PromptCategoryCustom), Name: "Custom", Description: "Other prompt types"},
	}

	return c.JSON(categories)
}

// CategoryInfo represents information about a prompt category
type CategoryInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// GetPopularTags returns popular tags
// @Summary Get popular tags
// @Description Get the most popular tags in the library
// @Tags prompt-library
// @Accept json
// @Produce json
// @Param limit query int false "Limit results" default(20)
// @Success 200 {array} service.TagCount
// @Router /api/public/library/tags [get]
func (h *PromptLibraryHandler) GetPopularTags(c *fiber.Ctx) error {
	limit := c.QueryInt("limit", 20)
	tags := h.promptLibraryService.GetPopularTags(c.Context(), limit)
	return c.JSON(tags)
}

// RecordUsage records that a prompt was used
// @Summary Record usage
// @Description Record that a library prompt was used
// @Tags prompt-library
// @Accept json
// @Produce json
// @Param id path string true "Prompt ID"
// @Param body body RecordUsageRequest true "Usage details"
// @Success 201 {object} domain.PromptUsageRecord
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/public/library/prompts/{id}/usage [post]
func (h *PromptLibraryHandler) RecordUsage(c *fiber.Ctx) error {
	promptID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid prompt ID")
	}

	var req RecordUsageRequest
	if err := c.BodyParser(&req); err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid request body")
	}

	userID := uuid.New() // In real implementation, get from auth context

	record, err := h.promptLibraryService.RecordUsage(
		c.Context(),
		promptID,
		req.Version,
		userID,
		req.ProjectID,
		req.TraceID,
	)
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	return c.Status(fiber.StatusCreated).JSON(record)
}

// RecordUsageRequest represents the request to record prompt usage
type RecordUsageRequest struct {
	Version   int        `json:"version"`
	ProjectID uuid.UUID  `json:"projectId"`
	TraceID   *uuid.UUID `json:"traceId,omitempty"`
}

// GetStarredPrompts returns prompts starred by the current user
// @Summary Get starred prompts
// @Description Get prompts starred by the current user
// @Tags prompt-library
// @Accept json
// @Produce json
// @Param limit query int false "Limit results" default(20)
// @Param offset query int false "Offset for pagination" default(0)
// @Success 200 {object} domain.LibraryPromptList
// @Failure 400 {object} ErrorResponse
// @Router /api/public/library/prompts/starred [get]
func (h *PromptLibraryHandler) GetStarredPrompts(c *fiber.Ctx) error {
	userID := uuid.New() // In real implementation, get from auth context

	h.logger.Debug("Get starred prompts", zap.String("userId", userID.String()))

	result := domain.LibraryPromptList{
		Prompts:    []domain.LibraryPrompt{},
		TotalCount: 0,
		HasMore:    false,
	}

	return c.JSON(result)
}

// GetMyPrompts returns prompts created by the current user
// @Summary Get my prompts
// @Description Get prompts created by the current user
// @Tags prompt-library
// @Accept json
// @Produce json
// @Param limit query int false "Limit results" default(20)
// @Param offset query int false "Offset for pagination" default(0)
// @Success 200 {object} domain.LibraryPromptList
// @Failure 400 {object} ErrorResponse
// @Router /api/public/library/prompts/mine [get]
func (h *PromptLibraryHandler) GetMyPrompts(c *fiber.Ctx) error {
	userID := uuid.New() // In real implementation, get from auth context

	h.logger.Debug("Get my prompts", zap.String("userId", userID.String()))

	result := domain.LibraryPromptList{
		Prompts:    []domain.LibraryPrompt{},
		TotalCount: 0,
		HasMore:    false,
	}

	return c.JSON(result)
}
