package middleware

import (
	"fmt"
	"runtime/debug"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// RecoverConfig configures the recover middleware
type RecoverConfig struct {
	// Logger instance
	Logger *zap.Logger
	// EnableStackAll enables stack trace for all goroutines
	EnableStackAll bool
	// StackSize limits the stack trace size
	StackSize int
	// Custom error handler
	ErrorHandler func(*fiber.Ctx, error) error
	// Sentry configuration
	SentryEnabled bool
}

// SentryConfig holds Sentry-specific configuration
type SentryConfig struct {
	DSN              string
	Environment      string
	Release          string
	Debug            bool
	SampleRate       float64
	TracesSampleRate float64
	FlushTimeout     time.Duration
}

// DefaultSentryConfig returns default Sentry configuration
func DefaultSentryConfig() SentryConfig {
	return SentryConfig{
		DSN:              "",
		Environment:      "development",
		Release:          "",
		Debug:            false,
		SampleRate:       1.0,
		TracesSampleRate: 0.1,
		FlushTimeout:     5 * time.Second,
	}
}

// DefaultRecoverConfig returns default recover config
func DefaultRecoverConfig(logger *zap.Logger) RecoverConfig {
	return RecoverConfig{
		Logger:         logger,
		EnableStackAll: false,
		StackSize:      4 << 10, // 4 KB
		ErrorHandler:   nil,
		SentryEnabled:  false,
	}
}

// InitSentry initializes the Sentry SDK
func InitSentry(config SentryConfig) error {
	if config.DSN == "" {
		return nil // Sentry disabled if no DSN
	}

	err := sentry.Init(sentry.ClientOptions{
		Dsn:              config.DSN,
		Environment:      config.Environment,
		Release:          config.Release,
		Debug:            config.Debug,
		SampleRate:       config.SampleRate,
		TracesSampleRate: config.TracesSampleRate,
		AttachStacktrace: true,
	})
	if err != nil {
		return fmt.Errorf("failed to initialize Sentry: %w", err)
	}

	return nil
}

// FlushSentry flushes any buffered events to Sentry
func FlushSentry(timeout time.Duration) {
	sentry.Flush(timeout)
}

// RecoverMiddleware creates a panic recovery middleware
type RecoverMiddleware struct {
	config RecoverConfig
}

// NewRecoverMiddleware creates a new recover middleware
func NewRecoverMiddleware(config RecoverConfig) *RecoverMiddleware {
	return &RecoverMiddleware{
		config: config,
	}
}

// Handler returns the recover handler
func (m *RecoverMiddleware) Handler() fiber.Handler {
	return func(c *fiber.Ctx) (err error) {
		defer func() {
			if r := recover(); r != nil {
				// Get stack trace
				stack := debug.Stack()
				if len(stack) > m.config.StackSize {
					stack = stack[:m.config.StackSize]
				}

				// Convert panic to error
				var panicErr error
				switch v := r.(type) {
				case error:
					panicErr = v
				case string:
					panicErr = fmt.Errorf("%s", v)
				default:
					panicErr = fmt.Errorf("%v", v)
				}

				// Log the panic
				m.config.Logger.Error("panic recovered",
					zap.Error(panicErr),
					zap.String("path", c.Path()),
					zap.String("method", c.Method()),
					zap.String("ip", c.IP()),
					zap.String("stack", string(stack)),
					zap.String("request_id", GetRequestID(c)),
				)

				// Custom error handler
				if m.config.ErrorHandler != nil {
					err = m.config.ErrorHandler(c, panicErr)
					return
				}

				// Default error response
				err = c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error":      "Internal Server Error",
					"message":    "An unexpected error occurred",
					"request_id": GetRequestID(c),
				})
			}
		}()

		return c.Next()
	}
}

// SimpleRecover creates a simple recovery middleware
func SimpleRecover(logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) (err error) {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("panic recovered",
					zap.Any("error", r),
					zap.String("path", c.Path()),
					zap.String("stack", string(debug.Stack())),
				)

				err = c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error":   "Internal Server Error",
					"message": "An unexpected error occurred",
				})
			}
		}()

		return c.Next()
	}
}

// RecoverWithSentry creates a recovery middleware that reports to Sentry
func RecoverWithSentry(logger *zap.Logger, sentryEnabled bool) fiber.Handler {
	return func(c *fiber.Ctx) (err error) {
		// Create a Sentry hub for this request if Sentry is enabled
		var hub *sentry.Hub
		if sentryEnabled {
			hub = sentry.CurrentHub().Clone()
			setSentryRequestContext(hub, c)
			hub.Scope().SetTag("request_id", GetRequestID(c))
		}

		defer func() {
			if r := recover(); r != nil {
				stack := debug.Stack()

				// Convert to error
				var panicErr error
				switch v := r.(type) {
				case error:
					panicErr = v
				default:
					panicErr = fmt.Errorf("%v", v)
				}

				// Log locally
				logger.Error("panic recovered",
					zap.Error(panicErr),
					zap.String("path", c.Path()),
					zap.String("method", c.Method()),
					zap.String("ip", c.IP()),
					zap.String("stack", string(stack)),
					zap.String("request_id", GetRequestID(c)),
				)

				// Report to Sentry if enabled
				if sentryEnabled && hub != nil {
					hub.Scope().SetExtra("stack_trace", string(stack))
					hub.Scope().SetExtra("path", c.Path())
					hub.Scope().SetExtra("method", c.Method())
					hub.Scope().SetExtra("ip", c.IP())
					hub.Scope().SetLevel(sentry.LevelFatal)

					eventID := hub.RecoverWithContext(c.Context(), r)
					if eventID != nil {
						logger.Info("panic reported to Sentry",
							zap.String("event_id", string(*eventID)),
						)
					}

					// Flush to ensure the event is sent
					hub.Flush(2 * time.Second)
				}

				err = c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error":      "Internal Server Error",
					"message":    "An unexpected error occurred",
					"request_id": GetRequestID(c),
				})
			}
		}()

		return c.Next()
	}
}

// SentryMiddleware creates a middleware that adds Sentry context to requests
func SentryMiddleware(enabled bool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if !enabled {
			return c.Next()
		}

		hub := sentry.CurrentHub().Clone()
		setSentryRequestContext(hub, c)
		hub.Scope().SetTag("request_id", GetRequestID(c))

		// Store hub in context for later use
		c.Locals("sentry_hub", hub)

		return c.Next()
	}
}

// CaptureError reports an error to Sentry from a Fiber context
func CaptureError(c *fiber.Ctx, err error) {
	hub, ok := c.Locals("sentry_hub").(*sentry.Hub)
	if !ok || hub == nil {
		// Try using the current hub if not in context
		hub = sentry.CurrentHub()
	}

	hub.Scope().SetExtra("path", c.Path())
	hub.Scope().SetExtra("method", c.Method())
	hub.Scope().SetTag("request_id", GetRequestID(c))

	hub.CaptureException(err)
}

// CaptureMessage reports a message to Sentry
func CaptureMessage(c *fiber.Ctx, message string, level sentry.Level) {
	hub, ok := c.Locals("sentry_hub").(*sentry.Hub)
	if !ok || hub == nil {
		hub = sentry.CurrentHub()
	}

	hub.Scope().SetLevel(level)
	hub.Scope().SetExtra("path", c.Path())
	hub.Scope().SetExtra("method", c.Method())
	hub.Scope().SetTag("request_id", GetRequestID(c))

	hub.CaptureMessage(message)
}

// setSentryRequestContext sets request context on a Sentry hub from Fiber context
func setSentryRequestContext(hub *sentry.Hub, c *fiber.Ctx) {
	headers := make(map[string]string)
	c.Request().Header.VisitAll(func(key, value []byte) {
		k := string(key)
		// Don't include sensitive headers
		if k != "Authorization" && k != "Cookie" && k != "X-Api-Key" {
			headers[k] = string(value)
		}
	})

	hub.Scope().SetContext("Request", map[string]interface{}{
		"url":          c.OriginalURL(),
		"method":       c.Method(),
		"headers":      headers,
		"query_string": string(c.Request().URI().QueryString()),
		"remote_addr":  c.IP(),
	})
}
