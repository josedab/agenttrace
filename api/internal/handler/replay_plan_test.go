package handler

import (
	"bytes"
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

type replayPlanUseCaseStub struct {
	projectID uuid.UUID
	traceID   string
	input     domain.ReplayPlanInput
	plan      *domain.ReplayPlan
}

func (s *replayPlanUseCaseStub) AssessCapabilities(
	_ context.Context,
	projectID uuid.UUID,
	traceID string,
	input domain.ReplayPlanInput,
) (*domain.ReplayCapabilityReport, error) {
	s.projectID, s.traceID, s.input = projectID, traceID, input
	return &domain.ReplayCapabilityReport{CanInspectTimeline: true}, nil
}

func (s *replayPlanUseCaseStub) CreatePlan(
	_ context.Context,
	projectID uuid.UUID,
	traceID string,
	_ *uuid.UUID,
	input domain.ReplayPlanInput,
) (*domain.ReplayPlan, error) {
	s.projectID, s.traceID, s.input = projectID, traceID, input
	return s.plan, nil
}

func (s *replayPlanUseCaseStub) GetPlan(
	_ context.Context,
	projectID, _ uuid.UUID,
) (*domain.ReplayPlan, error) {
	s.projectID = projectID
	return s.plan, nil
}

func (s *replayPlanUseCaseStub) ExecutePlan(
	_ context.Context,
	projectID, _ uuid.UUID,
) (*domain.ReplayPlan, error) {
	s.projectID = projectID
	return s.plan, nil
}

func (s *replayPlanUseCaseStub) RetryPlan(
	_ context.Context,
	projectID, _ uuid.UUID,
) (*domain.ReplayPlan, error) {
	s.projectID = projectID
	return s.plan, nil
}

func (s *replayPlanUseCaseStub) GetComparison(
	_ context.Context,
	projectID, _ uuid.UUID,
) (*domain.ReplayPlanComparison, error) {
	s.projectID = projectID
	return &domain.ReplayPlanComparison{Verdict: "recorded_equivalent"}, nil
}

func TestReplayPlanHandlerRequiresProjectContext(t *testing.T) {
	handler := NewReplayPlanHandler(&replayPlanUseCaseStub{}, zap.NewNop())
	app := fiber.New()
	app.Post("/traces/:traceId/replay-plans", handler.CreatePlan)

	response, err := app.Test(
		httptest.NewRequestWithContext(
			context.Background(),
			"POST",
			"/traces/trace-1/replay-plans",
			nil,
		),
	)

	require.NoError(t, err)
	defer func() { require.NoError(t, response.Body.Close()) }()
	assert.Equal(t, fiber.StatusUnauthorized, response.StatusCode)
}

func TestReplayPlanHandlerUsesAuthorizedProjectAndTrace(t *testing.T) {
	projectID := uuid.New()
	traceID := uuid.New().String()
	planID := uuid.New()
	useCase := &replayPlanUseCaseStub{
		plan: &domain.ReplayPlan{
			ID:        planID,
			ProjectID: projectID,
			TraceID:   traceID,
			Status:    domain.ReplayPlanReady,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}
	handler := NewReplayPlanHandler(useCase, zap.NewNop())
	app := fiber.New()
	app.Post(
		"/traces/:traceId/replay-plans",
		func(c *fiber.Ctx) error {
			c.Locals(string(middleware.ContextKeyProjectID), projectID)
			return c.Next()
		},
		handler.CreatePlan,
	)

	body, err := json.Marshal(domain.ReplayPlanInput{
		Mode: domain.ReplayModeRecordedGeneration,
	})
	require.NoError(t, err)
	request := httptest.NewRequestWithContext(
		context.Background(),
		"POST",
		"/traces/"+traceID+"/replay-plans",
		bytes.NewReader(body),
	)
	request.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

	response, err := app.Test(request)

	require.NoError(t, err)
	defer func() { require.NoError(t, response.Body.Close()) }()
	assert.Equal(t, fiber.StatusCreated, response.StatusCode)
	assert.Equal(t, projectID, useCase.projectID)
	assert.Equal(t, traceID, useCase.traceID)
	assert.Equal(t, domain.ReplayModeRecordedGeneration, useCase.input.Mode)
}

func TestReplayPlanHandlerRejectsInvalidPlanID(t *testing.T) {
	projectID := uuid.New()
	handler := NewReplayPlanHandler(&replayPlanUseCaseStub{}, zap.NewNop())
	app := fiber.New()
	app.Get(
		"/replay-plans/:planId",
		func(c *fiber.Ctx) error {
			c.Locals(string(middleware.ContextKeyProjectID), projectID)
			return c.Next()
		},
		handler.GetPlan,
	)

	response, err := app.Test(httptest.NewRequestWithContext(
		context.Background(),
		"GET",
		"/replay-plans/not-a-uuid",
		nil,
	))

	require.NoError(t, err)
	defer func() { require.NoError(t, response.Body.Close()) }()
	assert.Equal(t, fiber.StatusBadRequest, response.StatusCode)
}

func TestReplayPlanHandlerRetriesWithinAuthorizedProject(t *testing.T) {
	projectID := uuid.New()
	planID := uuid.New()
	useCase := &replayPlanUseCaseStub{
		plan: &domain.ReplayPlan{
			ID:        planID,
			ProjectID: projectID,
			Status:    domain.ReplayPlanReady,
		},
	}
	handler := NewReplayPlanHandler(useCase, zap.NewNop())
	app := fiber.New()
	app.Post(
		"/replay-plans/:planId/retry",
		func(c *fiber.Ctx) error {
			c.Locals(string(middleware.ContextKeyProjectID), projectID)
			return c.Next()
		},
		handler.RetryPlan,
	)

	response, err := app.Test(httptest.NewRequestWithContext(
		context.Background(),
		"POST",
		"/replay-plans/"+planID.String()+"/retry",
		nil,
	))

	require.NoError(t, err)
	defer func() { require.NoError(t, response.Body.Close()) }()
	assert.Equal(t, fiber.StatusOK, response.StatusCode)
	assert.Equal(t, projectID, useCase.projectID)
}
