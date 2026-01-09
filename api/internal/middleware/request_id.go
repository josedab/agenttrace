package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// RequestIDConfig configures the request ID middleware
type RequestIDConfig struct {
	// Header is the header key for the request ID
	Header string
	// Generator generates a new request ID
	Generator func() string
}

// DefaultRequestIDConfig returns default request ID config
func DefaultRequestIDConfig() RequestIDConfig {
	return RequestIDConfig{
		Header: "X-Request-ID",
		Generator: func() string {
			return uuid.New().String()
		},
	}
}

// RequestID creates a request ID middleware
func RequestID(config ...RequestIDConfig) fiber.Handler {
	cfg := DefaultRequestIDConfig()
	if len(config) > 0 {
		cfg = config[0]
	}

	return func(c *fiber.Ctx) error {
		// Check if request ID already exists
		requestID := c.Get(cfg.Header)
		if requestID == "" {
			requestID = cfg.Generator()
		}

		// Set request ID in response header
		c.Set(cfg.Header, requestID)

		// Store in locals for use in handlers
		c.Locals("requestID", requestID)

		return c.Next()
	}
}

// TraceRequestID creates a request ID using trace ID format
func TraceRequestID() fiber.Handler {
	return RequestID(RequestIDConfig{
		Header: "X-Request-ID",
		Generator: func() string {
			// Generate W3C trace ID format (32 hex chars)
			id := uuid.New()
			return id.String()[:8] + uuid.New().String()[:8] + uuid.New().String()[:8] + uuid.New().String()[:8]
		},
	})
}
