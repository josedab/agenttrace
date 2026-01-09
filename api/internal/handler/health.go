package handler

import (
	"context"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

// HealthHandler handles health check endpoints
type HealthHandler struct {
	postgres   *pgxpool.Pool
	clickhouse clickhouse.Conn
	redis      *redis.Client
	version    string
	startTime  time.Time
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(
	postgres *pgxpool.Pool,
	clickhouse clickhouse.Conn,
	redis *redis.Client,
	version string,
) *HealthHandler {
	return &HealthHandler{
		postgres:   postgres,
		clickhouse: clickhouse,
		redis:      redis,
		version:    version,
		startTime:  time.Now(),
	}
}

// HealthStatus represents health check status
type HealthStatus struct {
	Status    string            `json:"status"`
	Version   string            `json:"version"`
	Uptime    string            `json:"uptime"`
	Timestamp string            `json:"timestamp"`
	Checks    map[string]string `json:"checks"`
}

// Health handles GET /health
func (h *HealthHandler) Health(c *fiber.Ctx) error {
	status := HealthStatus{
		Status:    "healthy",
		Version:   h.version,
		Uptime:    time.Since(h.startTime).String(),
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Checks:    make(map[string]string),
	}

	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	// Check PostgreSQL
	if err := h.postgres.Ping(ctx); err != nil {
		status.Status = "unhealthy"
		status.Checks["postgres"] = "unhealthy: " + err.Error()
	} else {
		status.Checks["postgres"] = "healthy"
	}

	// Check ClickHouse
	if err := h.clickhouse.Ping(ctx); err != nil {
		status.Status = "unhealthy"
		status.Checks["clickhouse"] = "unhealthy: " + err.Error()
	} else {
		status.Checks["clickhouse"] = "healthy"
	}

	// Check Redis
	if _, err := h.redis.Ping(ctx).Result(); err != nil {
		status.Status = "unhealthy"
		status.Checks["redis"] = "unhealthy: " + err.Error()
	} else {
		status.Checks["redis"] = "healthy"
	}

	statusCode := fiber.StatusOK
	if status.Status != "healthy" {
		statusCode = fiber.StatusServiceUnavailable
	}

	return c.Status(statusCode).JSON(status)
}

// Liveness handles GET /livez - basic liveness probe
func (h *HealthHandler) Liveness(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status": "alive",
	})
}

// Readiness handles GET /readyz - readiness probe
func (h *HealthHandler) Readiness(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.Context(), 3*time.Second)
	defer cancel()

	// Check all dependencies
	if err := h.postgres.Ping(ctx); err != nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"status": "not ready",
			"reason": "postgres unavailable",
		})
	}

	if err := h.clickhouse.Ping(ctx); err != nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"status": "not ready",
			"reason": "clickhouse unavailable",
		})
	}

	if _, err := h.redis.Ping(ctx).Result(); err != nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"status": "not ready",
			"reason": "redis unavailable",
		})
	}

	return c.JSON(fiber.Map{
		"status": "ready",
	})
}

// Version handles GET /version
func (h *HealthHandler) Version(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"version": h.version,
		"uptime":  time.Since(h.startTime).String(),
	})
}

// RegisterRoutes registers health check routes
func (h *HealthHandler) RegisterRoutes(app *fiber.App) {
	app.Get("/health", h.Health)
	app.Get("/healthz", h.Health)
	app.Get("/livez", h.Liveness)
	app.Get("/live", h.Liveness)
	app.Get("/readyz", h.Readiness)
	app.Get("/ready", h.Readiness)
	app.Get("/version", h.Version)
}
