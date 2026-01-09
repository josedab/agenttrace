package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// LoggerConfig configures the logger middleware
type LoggerConfig struct {
	// Logger instance
	Logger *zap.Logger
	// Skip function
	Skip func(*fiber.Ctx) bool
	// Fields to include
	IncludeHeaders bool
	IncludeBody    bool
	MaxBodySize    int
}

// DefaultLoggerConfig returns default logger config
func DefaultLoggerConfig(logger *zap.Logger) LoggerConfig {
	return LoggerConfig{
		Logger:         logger,
		Skip:           nil,
		IncludeHeaders: false,
		IncludeBody:    false,
		MaxBodySize:    1024,
	}
}

// LoggerMiddleware creates a request logging middleware
type LoggerMiddleware struct {
	config LoggerConfig
}

// NewLoggerMiddleware creates a new logger middleware
func NewLoggerMiddleware(config LoggerConfig) *LoggerMiddleware {
	return &LoggerMiddleware{
		config: config,
	}
}

// Handler returns the logger handler
func (m *LoggerMiddleware) Handler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Skip if skip function returns true
		if m.config.Skip != nil && m.config.Skip(c) {
			return c.Next()
		}

		start := time.Now()

		// Generate request ID if not present
		requestID := c.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
			c.Set("X-Request-ID", requestID)
		}

		// Store request ID in locals for use in handlers
		c.Locals("requestID", requestID)

		// Process request
		err := c.Next()

		// Calculate latency
		latency := time.Since(start)

		// Build log fields
		fields := []zap.Field{
			zap.String("request_id", requestID),
			zap.String("method", c.Method()),
			zap.String("path", c.Path()),
			zap.String("query", string(c.Request().URI().QueryString())),
			zap.Int("status", c.Response().StatusCode()),
			zap.Duration("latency", latency),
			zap.String("ip", c.IP()),
			zap.String("user_agent", c.Get("User-Agent")),
		}

		// Add user/project context if available
		if userID, ok := GetUserID(c); ok {
			fields = append(fields, zap.String("user_id", userID.String()))
		}
		if projectID, ok := GetProjectID(c); ok {
			fields = append(fields, zap.String("project_id", projectID.String()))
		}

		// Add headers if configured
		if m.config.IncludeHeaders {
			headers := make(map[string]string)
			c.Request().Header.VisitAll(func(key, value []byte) {
				k := string(key)
				// Skip sensitive headers
				if k != "Authorization" && k != "X-API-Key" && k != "Cookie" {
					headers[k] = string(value)
				}
			})
			fields = append(fields, zap.Any("headers", headers))
		}

		// Add error if present
		if err != nil {
			fields = append(fields, zap.Error(err))
		}

		// Log based on status code
		status := c.Response().StatusCode()
		switch {
		case status >= 500:
			m.config.Logger.Error("request completed", fields...)
		case status >= 400:
			m.config.Logger.Warn("request completed", fields...)
		default:
			m.config.Logger.Info("request completed", fields...)
		}

		return err
	}
}

// HealthSkipper skips logging for health check endpoints
func HealthSkipper(c *fiber.Ctx) bool {
	path := c.Path()
	return path == "/health" || path == "/healthz" || path == "/ready" || path == "/readyz" || path == "/live" || path == "/livez"
}

// StaticSkipper skips logging for static assets
func StaticSkipper(c *fiber.Ctx) bool {
	path := c.Path()
	return len(path) > 1 && (path[0:1] == "/" && (path[1:2] == "_" || // /_next/
		(len(path) > 7 && path[1:7] == "static"))) // /static/
}

// CombinedSkipper combines multiple skippers
func CombinedSkipper(skippers ...func(*fiber.Ctx) bool) func(*fiber.Ctx) bool {
	return func(c *fiber.Ctx) bool {
		for _, skip := range skippers {
			if skip(c) {
				return true
			}
		}
		return false
	}
}

// GetRequestID gets the request ID from context
func GetRequestID(c *fiber.Ctx) string {
	if requestID, ok := c.Locals("requestID").(string); ok {
		return requestID
	}
	return ""
}

// AccessLog creates a simple access log middleware
func AccessLog(logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		err := c.Next()

		logger.Info("access",
			zap.String("method", c.Method()),
			zap.String("path", c.Path()),
			zap.Int("status", c.Response().StatusCode()),
			zap.Duration("latency", time.Since(start)),
			zap.String("ip", c.IP()),
		)

		return err
	}
}
