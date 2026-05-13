package handler

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/middleware"
)

type outcomeUseCaseStub struct {
	projectID uuid.UUID
	overview  *domain.OutcomeOverview
	digest    *domain.OutcomeDigest
}

func (s *outcomeUseCaseStub) GetOverview(
	_ context.Context,
	projectID uuid.UUID,
	_, _ time.Time,
) (*domain.OutcomeOverview, error) {
	s.projectID = projectID
	return s.overview, nil
}

func (s *outcomeUseCaseStub) GetDigest(
	_ context.Context,
	projectID uuid.UUID,
	_, _ time.Time,
) (*domain.OutcomeDigest, error) {
	s.projectID = projectID
	return s.digest, nil
}

func TestOutcomeHandlerRequiresProjectContext(t *testing.T) {
	handler := NewOutcomeHandler(&outcomeUseCaseStub{}, zap.NewNop())
	app := fiber.New()
	app.Get("/outcomes", handler.GetOverview)

	response, err := app.Test(httptest.NewRequestWithContext(
		context.Background(),
		"GET",
		"/outcomes",
		nil,
	))

	require.NoError(t, err)
	defer func() { require.NoError(t, response.Body.Close()) }()
	assert.Equal(t, fiber.StatusUnauthorized, response.StatusCode)
}

func TestOutcomeHandlerUsesAuthorizedProject(t *testing.T) {
	projectID := uuid.New()
	useCase := &outcomeUseCaseStub{
		overview: &domain.OutcomeOverview{ProjectID: projectID},
	}
	handler := NewOutcomeHandler(useCase, zap.NewNop())
	app := fiber.New()
	app.Get(
		"/outcomes",
		func(c *fiber.Ctx) error {
			c.Locals(string(middleware.ContextKeyProjectID), projectID)
			return c.Next()
		},
		handler.GetOverview,
	)

	response, err := app.Test(httptest.NewRequestWithContext(
		context.Background(),
		"GET",
		"/outcomes?window=7d",
		nil,
	))

	require.NoError(t, err)
	defer func() { require.NoError(t, response.Body.Close()) }()
	assert.Equal(t, fiber.StatusOK, response.StatusCode)
	assert.Equal(t, projectID, useCase.projectID)

	var body domain.OutcomeOverview
	require.NoError(t, json.NewDecoder(response.Body).Decode(&body))
	assert.Equal(t, projectID, body.ProjectID)
}

func TestOutcomeHandlerRejectsInvalidPeriod(t *testing.T) {
	projectID := uuid.New()
	handler := NewOutcomeHandler(&outcomeUseCaseStub{}, zap.NewNop())
	app := fiber.New()
	app.Get(
		"/outcomes",
		func(c *fiber.Ctx) error {
			c.Locals(string(middleware.ContextKeyProjectID), projectID)
			return c.Next()
		},
		handler.GetOverview,
	)

	response, err := app.Test(httptest.NewRequestWithContext(
		context.Background(),
		"GET",
		"/outcomes?window=forever",
		nil,
	))

	require.NoError(t, err)
	defer func() { require.NoError(t, response.Body.Close()) }()
	assert.Equal(t, fiber.StatusBadRequest, response.StatusCode)
}

func TestOutcomeHandlerRendersMarkdownDigest(t *testing.T) {
	projectID := uuid.New()
	useCase := &outcomeUseCaseStub{
		digest: &domain.OutcomeDigest{
			ProjectID: projectID,
			Period: domain.OutcomePeriod{
				From: time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
				To:   time.Date(2026, 7, 8, 0, 0, 0, 0, time.UTC),
			},
			Title:   "Agent outcome digest",
			Summary: "Eight successful outcomes.",
		},
	}
	handler := NewOutcomeHandler(useCase, zap.NewNop())
	app := fiber.New()
	app.Get(
		"/digest",
		func(c *fiber.Ctx) error {
			c.Locals(string(middleware.ContextKeyProjectID), projectID)
			return c.Next()
		},
		handler.GetDigest,
	)

	response, err := app.Test(httptest.NewRequestWithContext(
		context.Background(),
		"GET",
		"/digest?format=markdown",
		nil,
	))

	require.NoError(t, err)
	defer func() { require.NoError(t, response.Body.Close()) }()
	assert.Equal(t, fiber.StatusOK, response.StatusCode)
	assert.Contains(t, response.Header.Get("Content-Type"), "text/markdown")
}
