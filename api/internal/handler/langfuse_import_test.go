package handler

import (
	"context"
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/middleware"
)

type langfuseImportUseCaseStub struct {
	projectID uuid.UUID
	actorID   uuid.UUID
}

func (s *langfuseImportUseCaseStub) ImportBatch(
	_ context.Context,
	projectID, actorID uuid.UUID,
	batch domain.LangfuseImportBatch,
) (*domain.MigrationJob, error) {
	s.projectID, s.actorID = projectID, actorID
	return &domain.MigrationJob{
		ID:        batch.JobID,
		ProjectID: projectID,
		Source:    "langfuse",
		Status:    domain.MigrationStatusCompleted,
		Config: domain.MigrationConfig{
			SourceDSN: "must-not-leak",
		},
	}, nil
}

func TestLangfuseImportHandlerRequiresProject(t *testing.T) {
	handler := NewLangfuseImportHandler(&langfuseImportUseCaseStub{}, zap.NewNop())
	app := fiber.New()
	app.Post("/migrations/langfuse/import", handler.ImportBatch)

	response, err := app.Test(
		httptest.NewRequestWithContext(
			context.Background(),
			"POST",
			"/migrations/langfuse/import",
			nil,
		),
	)

	require.NoError(t, err)
	defer func() { require.NoError(t, response.Body.Close()) }()
	assert.Equal(t, fiber.StatusUnauthorized, response.StatusCode)
}

func TestLangfuseImportHandlerUsesScopedActorAndRedactsConfig(t *testing.T) {
	projectID := uuid.New()
	userID := uuid.New()
	useCase := &langfuseImportUseCaseStub{}
	handler := NewLangfuseImportHandler(useCase, zap.NewNop())
	app := fiber.New()
	app.Post(
		"/migrations/langfuse/import",
		func(c *fiber.Ctx) error {
			c.Locals(string(middleware.ContextKeyProjectID), projectID)
			c.Locals(string(middleware.ContextKeyUserID), userID)
			return c.Next()
		},
		handler.ImportBatch,
	)
	request := httptest.NewRequestWithContext(
		context.Background(),
		"POST",
		"/migrations/langfuse/import",
		strings.NewReader(`{
			"jobId":"00000000-0000-0000-0000-000000000001",
			"fingerprint":"0123456789abcdef",
			"records":{"traces":[{"id":"trace","startTime":"2026-07-25T10:00:00Z"}]}
		}`),
	)
	request.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

	response, err := app.Test(request)

	require.NoError(t, err)
	defer func() { require.NoError(t, response.Body.Close()) }()
	assert.Equal(t, fiber.StatusOK, response.StatusCode)
	assert.Equal(t, projectID, useCase.projectID)
	assert.Equal(t, userID, useCase.actorID)
	body, err := io.ReadAll(response.Body)
	require.NoError(t, err)
	assert.NotContains(t, string(body), "must-not-leak")
}
