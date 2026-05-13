package handler

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/middleware"
	apperrors "github.com/agenttrace/agenttrace/api/internal/pkg/errors"
)

type shareLinkUseCaseStub struct {
	projectID uuid.UUID
	actorID   uuid.UUID
	input     domain.ShareLinkInput
}

func (s *shareLinkUseCaseStub) Create(
	_ context.Context,
	projectID, actorID uuid.UUID,
	input domain.ShareLinkInput,
) (*domain.ShareLinkCreated, error) {
	s.projectID, s.actorID, s.input = projectID, actorID, input
	return &domain.ShareLinkCreated{
		ShareLink: domain.ShareLink{
			ID:           uuid.New(),
			ResourceType: input.ResourceType,
			ResourceID:   input.ResourceID,
		},
		Token: "token",
		URL:   "/share/token",
	}, nil
}

func (s *shareLinkUseCaseStub) Resolve(
	_ context.Context,
	_ string,
) (*domain.SharedResourceView, error) {
	return nil, apperrors.NotFound("share link")
}

func (s *shareLinkUseCaseStub) Revoke(
	_ context.Context,
	_ uuid.UUID,
	_ uuid.UUID,
) error {
	return nil
}

func TestShareLinkHandlerRequiresProjectContext(t *testing.T) {
	handler := NewShareLinkHandler(&shareLinkUseCaseStub{}, zap.NewNop())
	app := fiber.New()
	app.Post("/traces/:traceId/share-links", handler.CreateTrace)

	response, err := app.Test(httptest.NewRequestWithContext(
		context.Background(),
		"POST",
		"/traces/trace-1/share-links",
		nil,
	))

	require.NoError(t, err)
	defer func() { require.NoError(t, response.Body.Close()) }()
	assert.Equal(t, fiber.StatusUnauthorized, response.StatusCode)
}

func TestShareLinkHandlerScopesTraceCreation(t *testing.T) {
	projectID := uuid.New()
	userID := uuid.New()
	traceID := uuid.New().String()
	useCase := &shareLinkUseCaseStub{}
	handler := NewShareLinkHandler(useCase, zap.NewNop())
	app := fiber.New()
	app.Post(
		"/traces/:traceId/share-links",
		func(c *fiber.Ctx) error {
			c.Locals(string(middleware.ContextKeyProjectID), projectID)
			c.Locals(string(middleware.ContextKeyUserID), userID)
			return c.Next()
		},
		handler.CreateTrace,
	)

	response, err := app.Test(
		httptest.NewRequestWithContext(
			context.Background(),
			"POST",
			"/traces/"+traceID+"/share-links",
			nil,
		),
	)

	require.NoError(t, err)
	defer func() { require.NoError(t, response.Body.Close()) }()
	assert.Equal(t, fiber.StatusCreated, response.StatusCode)
	assert.Equal(t, projectID, useCase.projectID)
	assert.Equal(t, userID, useCase.actorID)
	assert.Equal(t, domain.ShareResourceTrace, useCase.input.ResourceType)
	assert.Equal(t, traceID, useCase.input.ResourceID)
}

func TestShareLinkHandlerHidesPublicResolutionErrors(t *testing.T) {
	handler := NewShareLinkHandler(&shareLinkUseCaseStub{}, zap.NewNop())
	app := fiber.New()
	app.Get("/api/share/:token", handler.Resolve)

	response, err := app.Test(httptest.NewRequestWithContext(
		context.Background(),
		"GET",
		"/api/share/invalid",
		nil,
	))

	require.NoError(t, err)
	defer func() { require.NoError(t, response.Body.Close()) }()
	assert.Equal(t, fiber.StatusNotFound, response.StatusCode)
	assert.Equal(t, "private, no-store", response.Header.Get(fiber.HeaderCacheControl))
}
