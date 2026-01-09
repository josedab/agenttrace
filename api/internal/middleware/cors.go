package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
)

// CORSConfig configures the CORS middleware
type CORSConfig struct {
	// AllowOrigins is a list of allowed origins
	AllowOrigins []string
	// AllowMethods is a list of allowed methods
	AllowMethods []string
	// AllowHeaders is a list of allowed headers
	AllowHeaders []string
	// ExposeHeaders is a list of headers to expose
	ExposeHeaders []string
	// AllowCredentials indicates whether credentials are allowed
	AllowCredentials bool
	// MaxAge indicates how long the results of a preflight request can be cached
	MaxAge int
}

// DefaultCORSConfig returns default CORS config
func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{
			fiber.MethodGet,
			fiber.MethodPost,
			fiber.MethodPut,
			fiber.MethodPatch,
			fiber.MethodDelete,
			fiber.MethodOptions,
			fiber.MethodHead,
		},
		AllowHeaders: []string{
			"Origin",
			"Content-Type",
			"Accept",
			"Authorization",
			"X-API-Key",
			"X-Request-ID",
			"X-Requested-With",
		},
		ExposeHeaders: []string{
			"X-Request-ID",
			"X-RateLimit-Limit",
			"X-RateLimit-Remaining",
			"X-RateLimit-Reset",
		},
		AllowCredentials: true,
		MaxAge:           86400, // 24 hours
	}
}

// ProductionCORSConfig returns CORS config for production
func ProductionCORSConfig(allowedOrigins []string) CORSConfig {
	config := DefaultCORSConfig()
	config.AllowOrigins = allowedOrigins
	return config
}

// CORSMiddleware creates a CORS middleware
type CORSMiddleware struct {
	config CORSConfig
}

// NewCORSMiddleware creates a new CORS middleware
func NewCORSMiddleware(config CORSConfig) *CORSMiddleware {
	return &CORSMiddleware{
		config: config,
	}
}

// Handler returns the CORS handler
func (m *CORSMiddleware) Handler() fiber.Handler {
	allowMethods := strings.Join(m.config.AllowMethods, ", ")
	allowHeaders := strings.Join(m.config.AllowHeaders, ", ")
	exposeHeaders := strings.Join(m.config.ExposeHeaders, ", ")

	return func(c *fiber.Ctx) error {
		origin := c.Get("Origin")

		// Check if origin is allowed
		allowOrigin := ""
		if len(m.config.AllowOrigins) == 1 && m.config.AllowOrigins[0] == "*" {
			if m.config.AllowCredentials {
				// Can't use * with credentials, so reflect the origin
				allowOrigin = origin
			} else {
				allowOrigin = "*"
			}
		} else {
			for _, o := range m.config.AllowOrigins {
				if o == origin || o == "*" {
					allowOrigin = origin
					break
				}
				// Support wildcard subdomains (e.g., *.example.com)
				if strings.HasPrefix(o, "*.") {
					domain := o[1:] // Remove *
					if strings.HasSuffix(origin, domain) {
						allowOrigin = origin
						break
					}
				}
			}
		}

		// Not allowed origin
		if allowOrigin == "" && origin != "" {
			return c.Next()
		}

		// Set CORS headers
		c.Set("Access-Control-Allow-Origin", allowOrigin)

		if m.config.AllowCredentials {
			c.Set("Access-Control-Allow-Credentials", "true")
		}

		if exposeHeaders != "" {
			c.Set("Access-Control-Expose-Headers", exposeHeaders)
		}

		// Handle preflight request
		if c.Method() == fiber.MethodOptions {
			c.Set("Access-Control-Allow-Methods", allowMethods)
			c.Set("Access-Control-Allow-Headers", allowHeaders)

			if m.config.MaxAge > 0 {
				c.Set("Access-Control-Max-Age", string(rune(m.config.MaxAge)))
			}

			c.Set("Content-Length", "0")
			return c.SendStatus(fiber.StatusNoContent)
		}

		return c.Next()
	}
}

// SimpleCORS creates a simple CORS middleware that allows all origins
func SimpleCORS() fiber.Handler {
	return func(c *fiber.Ctx) error {
		origin := c.Get("Origin")
		if origin == "" {
			origin = "*"
		}

		c.Set("Access-Control-Allow-Origin", origin)
		c.Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS, HEAD")
		c.Set("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-API-Key, X-Request-ID")
		c.Set("Access-Control-Allow-Credentials", "true")
		c.Set("Access-Control-Expose-Headers", "X-Request-ID, X-RateLimit-Limit, X-RateLimit-Remaining")

		if c.Method() == fiber.MethodOptions {
			c.Set("Access-Control-Max-Age", "86400")
			return c.SendStatus(fiber.StatusNoContent)
		}

		return c.Next()
	}
}

// SecureCORS creates a CORS middleware with specific allowed origins
func SecureCORS(allowedOrigins ...string) fiber.Handler {
	originMap := make(map[string]bool)
	for _, o := range allowedOrigins {
		originMap[o] = true
	}

	return func(c *fiber.Ctx) error {
		origin := c.Get("Origin")

		if origin != "" && !originMap[origin] {
			// Check for wildcard patterns
			allowed := false
			for o := range originMap {
				if strings.HasPrefix(o, "*.") {
					domain := o[1:]
					if strings.HasSuffix(origin, domain) {
						allowed = true
						break
					}
				}
			}
			if !allowed {
				return c.Next()
			}
		}

		c.Set("Access-Control-Allow-Origin", origin)
		c.Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Set("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-API-Key")
		c.Set("Access-Control-Allow-Credentials", "true")

		if c.Method() == fiber.MethodOptions {
			return c.SendStatus(fiber.StatusNoContent)
		}

		return c.Next()
	}
}
