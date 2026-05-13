package handler

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/middleware"
	apperrors "github.com/agenttrace/agenttrace/api/internal/pkg/errors"
)

// LangfuseImportUseCase defines resumable batch import behavior.
type LangfuseImportUseCase interface {
	ImportBatch(
		ctx context.Context,
		projectID, actorID uuid.UUID,
		batch domain.LangfuseImportBatch,
	) (*domain.MigrationJob, error)
}

// LangfuseImportHandler transports JSON export batches.
type LangfuseImportHandler struct {
	service LangfuseImportUseCase
	logger  *zap.Logger
}

// NewLangfuseImportHandler creates a Langfuse import handler.
func NewLangfuseImportHandler(
	service LangfuseImportUseCase,
	logger *zap.Logger,
) *LangfuseImportHandler {
	return &LangfuseImportHandler{service: service, logger: logger}
}

// ImportBatch handles POST /migrations/langfuse/import.
func (h *LangfuseImportHandler) ImportBatch(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return langfuseImportError(c, fiber.StatusUnauthorized, "Project ID not found")
	}
	actorID, ok := migrationActorID(c)
	if !ok {
		return langfuseImportError(c, fiber.StatusUnauthorized, "Actor ID not found")
	}
	var batch domain.LangfuseImportBatch
	if err := c.BodyParser(&batch); err != nil {
		return langfuseImportError(c, fiber.StatusBadRequest, "Invalid request body")
	}
	job, err := h.service.ImportBatch(c.Context(), projectID, actorID, batch)
	if err != nil {
		if appErr := apperrors.GetAppError(err); appErr != nil {
			return langfuseImportError(c, appErr.StatusCode, appErr.Message)
		}
		h.logger.Error("Langfuse import batch failed", zap.Error(err))
		return langfuseImportError(
			c,
			fiber.StatusInternalServerError,
			"Langfuse import batch failed",
		)
	}
	return c.JSON(redactMigrationJob(job))
}

func migrationActorID(c *fiber.Ctx) (uuid.UUID, bool) {
	return roadmapActorID(c)
}

func langfuseImportError(c *fiber.Ctx, status int, message string) error {
	return roadmapError(c, status, message)
}
