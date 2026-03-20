package middleware

import (
	"context"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

const (
	ContextKeyTenantID ContextKey = "tenantID"
)

// TenantMiddleware enforces multi-tenant usage limits on API requests.
type TenantMiddleware struct {
	tenantService *service.TenantService
	logger        *zap.Logger
	enabled       bool
}

// NewTenantMiddleware creates a new tenant middleware.
// When disabled (e.g., self-hosted mode), all requests pass through.
func NewTenantMiddleware(tenantService *service.TenantService, logger *zap.Logger, enabled bool) *TenantMiddleware {
	return &TenantMiddleware{
		tenantService: tenantService,
		logger:        logger,
		enabled:       enabled,
	}
}

// EnforceTraceLimit checks trace ingestion limits before allowing ingestion.
func (m *TenantMiddleware) EnforceTraceLimit() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if !m.enabled {
			return c.Next()
		}

		tenantID, ok := GetTenantID(c)
		if !ok {
			// No tenant context means self-hosted mode — allow through
			return c.Next()
		}

		allowed, remaining, err := m.tenantService.CheckUsageLimit(
			c.Context(), tenantID, domain.UsageEventTraceIngested,
		)
		if err != nil {
			m.logger.Error("failed to check tenant usage limit", zap.Error(err))
			// Fail open: allow the request on error to avoid blocking ingestion
			return c.Next()
		}

		if !allowed {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":   "UsageLimitExceeded",
				"message": "Monthly trace limit reached. Upgrade your plan for more capacity.",
				"plan":    "https://agenttrace.io/pricing",
			})
		}

		// Set remaining header for client visibility
		c.Set("X-AgentTrace-Traces-Remaining", formatInt64(remaining))

		return c.Next()
	}
}

// MeterTraceIngestion records a trace ingestion event after successful processing.
func (m *TenantMiddleware) MeterTraceIngestion() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Process the request first
		err := c.Next()
		if err != nil {
			return err
		}

		if !m.enabled {
			return nil
		}

		// Only meter successful ingestions
		if c.Response().StatusCode() >= 200 && c.Response().StatusCode() < 300 {
			tenantID, ok := GetTenantID(c)
			if !ok {
				return nil
			}

			// Record asynchronously with a dedicated context and timeout
			// to avoid goroutine leaks if the service is slow or unresponsive
			go func() {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				if meterErr := m.tenantService.RecordUsage(
					ctx, tenantID, domain.UsageEventTraceIngested, 1,
				); meterErr != nil {
					m.logger.Warn("failed to meter trace ingestion",
						zap.Error(meterErr),
						zap.String("tenantId", tenantID.String()),
					)
				}
			}()
		}

		return nil
	}
}

// GetTenantID gets the tenant ID from the request context.
func GetTenantID(c *fiber.Ctx) (uuid.UUID, bool) {
	tenantID, ok := c.Locals(string(ContextKeyTenantID)).(uuid.UUID)
	return tenantID, ok
}

func formatInt64(n int64) string {
	return fmt.Sprintf("%d", n)
}
