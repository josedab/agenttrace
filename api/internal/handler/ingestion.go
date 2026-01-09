package handler

import (
	"encoding/json"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/middleware"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// IngestionHandler handles trace ingestion endpoints
type IngestionHandler struct {
	ingestionService *service.IngestionService
	scoreService     *service.ScoreService
	logger           *zap.Logger
}

// NewIngestionHandler creates a new ingestion handler
func NewIngestionHandler(ingestionService *service.IngestionService, scoreService *service.ScoreService, logger *zap.Logger) *IngestionHandler {
	return &IngestionHandler{
		ingestionService: ingestionService,
		scoreService:     scoreService,
		logger:           logger,
	}
}

// BatchIngestion handles POST /api/public/ingestion
// This is the main Langfuse-compatible batch ingestion endpoint
func (h *IngestionHandler) BatchIngestion(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	var request struct {
		Batch    []json.RawMessage `json:"batch"`
		Metadata map[string]any    `json:"metadata,omitempty"`
	}

	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid request body: " + err.Error(),
		})
	}

	if len(request.Batch) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Batch is empty",
		})
	}

	successes := make([]string, 0)
	errors := make([]map[string]any, 0)

	for _, item := range request.Batch {
		// Parse common fields to determine type
		var common struct {
			ID        string `json:"id"`
			Type      string `json:"type"`
			Timestamp string `json:"timestamp,omitempty"`
			Body      any    `json:"body,omitempty"`
		}

		if err := json.Unmarshal(item, &common); err != nil {
			errors = append(errors, map[string]any{
				"id":      "unknown",
				"status":  400,
				"message": "Invalid JSON: " + err.Error(),
			})
			continue
		}

		eventID := common.ID
		if eventID == "" {
			eventID = uuid.New().String()
		}

		switch common.Type {
		case "trace-create":
			var traceInput domain.TraceInput
			if err := json.Unmarshal(item, &struct {
				Body *domain.TraceInput `json:"body"`
			}{Body: &traceInput}); err != nil {
				errors = append(errors, map[string]any{
					"id":      eventID,
					"status":  400,
					"message": "Invalid trace input: " + err.Error(),
				})
				continue
			}
			if _, err := h.ingestionService.IngestTrace(c.Context(), projectID, &traceInput); err != nil {
				errors = append(errors, map[string]any{
					"id":      eventID,
					"status":  500,
					"message": err.Error(),
				})
				continue
			}
			successes = append(successes, eventID)

		case "span-create", "generation-create", "event-create":
			var obsInput domain.ObservationInput
			if err := json.Unmarshal(item, &struct {
				Body *domain.ObservationInput `json:"body"`
			}{Body: &obsInput}); err != nil {
				errors = append(errors, map[string]any{
					"id":      eventID,
					"status":  400,
					"message": "Invalid observation input: " + err.Error(),
				})
				continue
			}

			// Set type based on event type
			var obsType domain.ObservationType
			switch common.Type {
			case "span-create":
				obsType = domain.ObservationTypeSpan
			case "generation-create":
				obsType = domain.ObservationTypeGeneration
			case "event-create":
				obsType = domain.ObservationTypeEvent
			}
			obsInput.Type = &obsType

			if _, err := h.ingestionService.IngestObservation(c.Context(), projectID, &obsInput); err != nil {
				errors = append(errors, map[string]any{
					"id":      eventID,
					"status":  500,
					"message": err.Error(),
				})
				continue
			}
			successes = append(successes, eventID)

		case "score-create":
			var scoreInput domain.ScoreInput
			if err := json.Unmarshal(item, &struct {
				Body *domain.ScoreInput `json:"body"`
			}{Body: &scoreInput}); err != nil {
				errors = append(errors, map[string]any{
					"id":      eventID,
					"status":  400,
					"message": "Invalid score input: " + err.Error(),
				})
				continue
			}
			if _, err := h.scoreService.Create(c.Context(), projectID, &scoreInput); err != nil {
				errors = append(errors, map[string]any{
					"id":      eventID,
					"status":  500,
					"message": err.Error(),
				})
				continue
			}
			successes = append(successes, eventID)

		case "sdk-log":
			// SDK logs are silently acknowledged
			successes = append(successes, eventID)

		default:
			errors = append(errors, map[string]any{
				"id":      eventID,
				"status":  400,
				"message": "Unknown event type: " + common.Type,
			})
		}
	}

	return c.JSON(fiber.Map{
		"successes": successes,
		"errors":    errors,
	})
}

// CreateTrace handles POST /v1/traces
func (h *IngestionHandler) CreateTrace(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	var input domain.TraceInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid request body: " + err.Error(),
		})
	}

	trace, err := h.ingestionService.IngestTrace(c.Context(), projectID, &input)
	if err != nil {
		h.logger.Error("failed to create trace",
			zap.Error(err),
			zap.String("project_id", projectID.String()),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to create trace",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(trace)
}

// UpdateTrace handles PATCH /v1/traces/:traceId
func (h *IngestionHandler) UpdateTrace(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	traceID := c.Params("traceId")
	if traceID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Trace ID required",
		})
	}

	var input domain.TraceInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid request body: " + err.Error(),
		})
	}

	// Set the trace ID from path
	input.ID = traceID

	trace, err := h.ingestionService.IngestTrace(c.Context(), projectID, &input)
	if err != nil {
		h.logger.Error("failed to update trace",
			zap.Error(err),
			zap.String("trace_id", traceID),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to update trace",
		})
	}

	return c.JSON(trace)
}

// CreateSpan handles POST /v1/spans
func (h *IngestionHandler) CreateSpan(c *fiber.Ctx) error {
	return h.createObservation(c, domain.ObservationTypeSpan)
}

// CreateGeneration handles POST /v1/generations
func (h *IngestionHandler) CreateGeneration(c *fiber.Ctx) error {
	return h.createObservation(c, domain.ObservationTypeGeneration)
}

// UpdateGeneration handles PATCH /v1/generations/:generationId
func (h *IngestionHandler) UpdateGeneration(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	generationID := c.Params("generationId")
	if generationID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Generation ID required",
		})
	}

	var input domain.ObservationInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid request body: " + err.Error(),
		})
	}

	input.ID = &generationID
	obsType := domain.ObservationTypeGeneration
	input.Type = &obsType

	obs, err := h.ingestionService.IngestObservation(c.Context(), projectID, &input)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to update generation",
		})
	}

	return c.JSON(obs)
}

// CreateEvent handles POST /v1/events
func (h *IngestionHandler) CreateEvent(c *fiber.Ctx) error {
	return h.createObservation(c, domain.ObservationTypeEvent)
}

func (h *IngestionHandler) createObservation(c *fiber.Ctx, obsType domain.ObservationType) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	var input domain.ObservationInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid request body: " + err.Error(),
		})
	}

	input.Type = &obsType

	obs, err := h.ingestionService.IngestObservation(c.Context(), projectID, &input)
	if err != nil {
		h.logger.Error("failed to create observation",
			zap.Error(err),
			zap.String("type", string(obsType)),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to create " + string(obsType),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(obs)
}

// CreateScore handles POST /v1/scores
func (h *IngestionHandler) CreateScore(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	var input domain.ScoreInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid request body: " + err.Error(),
		})
	}

	score, err := h.scoreService.Create(c.Context(), projectID, &input)
	if err != nil {
		h.logger.Error("failed to create score",
			zap.Error(err),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to create score",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(score)
}

// SDKLog handles POST /api/public/sdk-log
func (h *IngestionHandler) SDKLog(c *fiber.Ctx) error {
	// SDK logs are useful for debugging but we just acknowledge them
	var log struct {
		Level   string `json:"level"`
		Message string `json:"message"`
		Args    any    `json:"args,omitempty"`
	}

	if err := c.BodyParser(&log); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid request body",
		})
	}

	h.logger.Debug("SDK log received",
		zap.String("level", log.Level),
		zap.String("message", log.Message),
		zap.Any("args", log.Args),
	)

	return c.JSON(fiber.Map{
		"status": "ok",
	})
}

// IngestionMetrics handles GET /api/public/metrics
func (h *IngestionHandler) IngestionMetrics(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	// Return basic ingestion metrics
	return c.JSON(fiber.Map{
		"project_id": projectID.String(),
		"timestamp":  time.Now().UTC().Format(time.RFC3339),
		// In a real implementation, these would be actual metrics
		"traces_24h":       0,
		"observations_24h": 0,
		"scores_24h":       0,
	})
}

// RegisterRoutes registers ingestion routes
func (h *IngestionHandler) RegisterRoutes(app *fiber.App, authMiddleware *middleware.AuthMiddleware) {
	// Public API (Langfuse-compatible)
	public := app.Group("/api/public", authMiddleware.RequireAPIKey())
	public.Post("/ingestion", h.BatchIngestion)
	public.Post("/sdk-log", h.SDKLog)
	public.Get("/metrics", h.IngestionMetrics)

	// REST API v1
	v1 := app.Group("/v1", authMiddleware.RequireAPIKey())
	v1.Post("/traces", h.CreateTrace)
	v1.Patch("/traces/:traceId", h.UpdateTrace)
	v1.Post("/spans", h.CreateSpan)
	v1.Post("/generations", h.CreateGeneration)
	v1.Patch("/generations/:generationId", h.UpdateGeneration)
	v1.Post("/events", h.CreateEvent)
	v1.Post("/scores", h.CreateScore)
}
