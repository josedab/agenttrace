package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
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

type evalHubUseCaseStub struct {
	projectID uuid.UUID
	userID    uuid.UUID
	input     domain.EvalHubPublishInput
	pkg       *domain.EvalHubPackage
}

func (s *evalHubUseCaseStub) ListPackages(
	_ context.Context,
	projectID uuid.UUID,
	_ domain.EvalHubPackageFilter,
) (*domain.EvalHubPackageList, error) {
	s.projectID = projectID
	return &domain.EvalHubPackageList{}, nil
}

func (s *evalHubUseCaseStub) GetPackage(
	_ context.Context,
	projectID, _ uuid.UUID,
) (*domain.EvalHubPackage, error) {
	s.projectID = projectID
	return s.pkg, nil
}

func (s *evalHubUseCaseStub) Publish(
	_ context.Context,
	projectID, userID uuid.UUID,
	input domain.EvalHubPublishInput,
) (*domain.EvalHubPackage, error) {
	s.projectID, s.userID, s.input = projectID, userID, input
	return s.pkg, nil
}

func (s *evalHubUseCaseStub) Fork(
	_ context.Context,
	projectID, userID, _ uuid.UUID,
	_ domain.EvalHubForkInput,
) (*domain.EvalHubPackage, error) {
	s.projectID, s.userID = projectID, userID
	return s.pkg, nil
}

func (s *evalHubUseCaseStub) Run(
	_ context.Context,
	projectID, userID, packageID uuid.UUID,
	_ domain.EvalHubRunInput,
) (*domain.EvalHubRun, error) {
	s.projectID, s.userID = projectID, userID
	return &domain.EvalHubRun{PackageID: packageID, ProjectID: projectID}, nil
}

func (s *evalHubUseCaseStub) GetRun(
	_ context.Context,
	projectID, runID uuid.UUID,
) (*domain.EvalHubRun, error) {
	s.projectID = projectID
	return &domain.EvalHubRun{ID: runID, ProjectID: projectID}, nil
}

func (s *evalHubUseCaseStub) ListRuns(
	_ context.Context,
	projectID uuid.UUID,
	_, _ int,
) (*domain.EvalHubRunList, error) {
	s.projectID = projectID
	return &domain.EvalHubRunList{}, nil
}

func TestEvalHubHandlerRequiresProjectContext(t *testing.T) {
	handler := NewEvalHubHandler(&evalHubUseCaseStub{}, zap.NewNop())
	app := fiber.New()
	app.Get("/eval-hub/packages", handler.ListPackages)

	response, err := app.Test(httptest.NewRequestWithContext(
		context.Background(),
		"GET",
		"/eval-hub/packages",
		nil,
	))

	require.NoError(t, err)
	defer func() { require.NoError(t, response.Body.Close()) }()
	assert.Equal(t, fiber.StatusUnauthorized, response.StatusCode)
}

func TestEvalHubHandlerRejectsInvalidKind(t *testing.T) {
	projectID := uuid.New()
	handler := NewEvalHubHandler(&evalHubUseCaseStub{}, zap.NewNop())
	app := fiber.New()
	app.Get(
		"/eval-hub/packages",
		func(c *fiber.Ctx) error {
			c.Locals(string(middleware.ContextKeyProjectID), projectID)
			return c.Next()
		},
		handler.ListPackages,
	)

	response, err := app.Test(
		httptest.NewRequestWithContext(
			context.Background(),
			"GET",
			"/eval-hub/packages?kind=unknown",
			nil,
		),
	)

	require.NoError(t, err)
	defer func() { require.NoError(t, response.Body.Close()) }()
	assert.Equal(t, fiber.StatusBadRequest, response.StatusCode)
}

func TestEvalHubHandlerPublishesForOwnedAPIKey(t *testing.T) {
	projectID := uuid.New()
	userID := uuid.New()
	apiKeyID := uuid.New()
	packageID := uuid.New()
	useCase := &evalHubUseCaseStub{
		pkg: &domain.EvalHubPackage{ID: packageID, OwnerProjectID: projectID},
	}
	handler := NewEvalHubHandler(useCase, zap.NewNop())
	app := fiber.New()
	app.Post(
		"/eval-hub/packages",
		func(c *fiber.Ctx) error {
			c.Locals(string(middleware.ContextKeyProjectID), projectID)
			c.Locals(string(middleware.ContextKeyAPIKeyID), apiKeyID)
			c.Locals(string(middleware.ContextKeyAuthType), middleware.AuthTypeAPIKey)
			c.Locals(string(middleware.ContextKeyUserID), userID)
			return c.Next()
		},
		handler.Publish,
	)

	input := domain.EvalHubPublishInput{
		Kind:             domain.EvalHubDataset,
		SourceResourceID: uuid.New(),
		Visibility:       domain.EvalHubVisibilityPrivate,
	}
	body, err := json.Marshal(input)
	require.NoError(t, err)
	request := httptest.NewRequestWithContext(
		context.Background(),
		"POST",
		"/eval-hub/packages",
		bytes.NewReader(body),
	)
	request.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

	response, err := app.Test(request)

	require.NoError(t, err)
	defer func() { require.NoError(t, response.Body.Close()) }()
	assert.Equal(t, fiber.StatusCreated, response.StatusCode)
	assert.Equal(t, projectID, useCase.projectID)
	assert.Equal(t, userID, useCase.userID)
	assert.Equal(t, input.SourceResourceID, useCase.input.SourceResourceID)
}

func TestEvalHubHandlerRejectsUnownedLegacyAPIKey(t *testing.T) {
	projectID := uuid.New()
	useCase := &evalHubUseCaseStub{
		pkg: &domain.EvalHubPackage{ID: uuid.New(), OwnerProjectID: projectID},
	}
	handler := NewEvalHubHandler(useCase, zap.NewNop())
	app := fiber.New()
	app.Post(
		"/eval-hub/packages",
		func(c *fiber.Ctx) error {
			c.Locals(string(middleware.ContextKeyProjectID), projectID)
			c.Locals(string(middleware.ContextKeyAPIKeyID), uuid.New())
			c.Locals(string(middleware.ContextKeyAuthType), middleware.AuthTypeAPIKey)
			return c.Next()
		},
		handler.Publish,
	)

	body := `{"kind":"dataset","sourceResourceId":"` + uuid.New().String() +
		`","visibility":"private"}`
	request := httptest.NewRequestWithContext(
		context.Background(),
		http.MethodPost,
		"/eval-hub/packages",
		strings.NewReader(body),
	)
	request.Header.Set("Content-Type", "application/json")
	response, err := app.Test(request)
	require.NoError(t, err)
	defer func() { require.NoError(t, response.Body.Close()) }()

	assert.Equal(t, fiber.StatusUnauthorized, response.StatusCode)
	assert.Equal(t, uuid.Nil, useCase.userID, "persistence must not be called with the API-key UUID")
}
