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

// MigrationHandler handles migration HTTP requests
type MigrationHandler struct {
	migrationService *service.MigrationService
	logger           *zap.Logger
}

// NewMigrationHandler creates a new migration handler
func NewMigrationHandler(migrationService *service.MigrationService, logger *zap.Logger) *MigrationHandler {
	return &MigrationHandler{
		migrationService: migrationService,
		logger:           logger,
	}
}

// StartMigration handles POST /migrations
func (h *MigrationHandler) StartMigration(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	var input domain.MigrationInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid request body",
		})
	}

	if input.Source == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "source is required",
		})
	}

	job, err := h.migrationService.StartMigration(c.Context(), projectID, &input)
	if err != nil {
		if appErr := apperrors.GetAppError(err); appErr != nil {
			return errorResponse(c, appErr.StatusCode, appErr.Message)
		}
		h.logger.Error("failed to start migration", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to start migration",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(redactMigrationJob(job))
}

// GetMigration handles GET /migrations/:jobId
func (h *MigrationHandler) GetMigration(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	jobID, err := uuid.Parse(c.Params("jobId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid job ID",
		})
	}

	job, err := h.migrationService.GetMigration(c.Context(), projectID, jobID)
	if err != nil {
		h.logger.Error("failed to get migration",
			zap.String("jobId", jobID.String()),
			zap.Error(err),
		)
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":   "Not Found",
			"message": "Migration job not found",
		})
	}

	return c.JSON(redactMigrationJob(job))
}

// ListMigrations handles GET /migrations
func (h *MigrationHandler) ListMigrations(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	jobs, err := h.migrationService.ListMigrations(c.Context(), projectID)
	if err != nil {
		h.logger.Error("failed to list migrations", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to list migrations",
		})
	}

	for i := range jobs {
		jobs[i] = *redactMigrationJob(&jobs[i])
	}
	return c.JSON(jobs)
}

func redactMigrationJob(job *domain.MigrationJob) *domain.MigrationJob {
	redacted := *job
	redacted.Config.SourceDSN = ""
	return &redacted
}

// ValidateSource handles POST /migrations/validate
func (h *MigrationHandler) ValidateSource(c *fiber.Ctx) error {
	_, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	var body struct {
		Source string `json:"source"`
		DSN    string `json:"dsn"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid request body",
		})
	}

	if body.Source == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "source is required",
		})
	}

	valid, message, err := h.migrationService.ValidateSource(c.Context(), body.Source, body.DSN)
	if err != nil {
		if appErr := apperrors.GetAppError(err); appErr != nil {
			return errorResponse(c, appErr.StatusCode, appErr.Message)
		}
		h.logger.Error("failed to validate migration source", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to validate source",
		})
	}

	return c.JSON(fiber.Map{
		"valid":   valid,
		"message": message,
	})
}
