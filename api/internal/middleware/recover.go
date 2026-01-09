package middleware

import (
	"fmt"
	"runtime/debug"

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
}

// DefaultRecoverConfig returns default recover config
func DefaultRecoverConfig(logger *zap.Logger) RecoverConfig {
	return RecoverConfig{
		Logger:         logger,
		EnableStackAll: false,
		StackSize:      4 << 10, // 4 KB
		ErrorHandler:   nil,
	}
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
// (placeholder for Sentry integration)
func RecoverWithSentry(logger *zap.Logger, dsn string) fiber.Handler {
	// In production, initialize Sentry here
	// sentry.Init(sentry.ClientOptions{Dsn: dsn})

	return func(c *fiber.Ctx) (err error) {
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
					zap.String("stack", string(stack)),
				)

				// Report to Sentry (placeholder)
				// sentry.CaptureException(panicErr)
				// sentry.Flush(time.Second * 5)

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
