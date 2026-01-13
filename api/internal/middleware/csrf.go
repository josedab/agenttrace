package middleware

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

const (
	// CSRFTokenLength is the length of the CSRF token in bytes (32 bytes = 256 bits)
	CSRFTokenLength = 32

	// CSRFCookieName is the name of the CSRF cookie
	CSRFCookieName = "_csrf"

	// CSRFHeaderName is the name of the CSRF header
	CSRFHeaderName = "X-CSRF-Token"

	// CSRFFormFieldName is the form field name for CSRF token
	CSRFFormFieldName = "csrf_token"

	// CSRFContextKey is the context key for the CSRF token
	CSRFContextKey = "csrfToken"
)

// CSRFConfig holds the configuration for CSRF middleware
type CSRFConfig struct {
	// CookieName is the name of the CSRF cookie
	CookieName string

	// CookiePath is the path for the CSRF cookie
	CookiePath string

	// CookieDomain is the domain for the CSRF cookie
	CookieDomain string

	// CookieSecure sets the Secure flag on the cookie
	CookieSecure bool

	// CookieSameSite sets the SameSite attribute
	CookieSameSite string

	// CookieHTTPOnly sets the HttpOnly flag on the cookie
	CookieHTTPOnly bool

	// TokenExpiry is how long the CSRF token is valid
	TokenExpiry time.Duration

	// SkipMethods are HTTP methods that don't need CSRF validation
	SkipMethods []string

	// TrustedOrigins are origins that can bypass CSRF checks
	TrustedOrigins []string

	// Enabled controls whether CSRF protection is active
	Enabled bool
}

// DefaultCSRFConfig returns the default CSRF configuration
func DefaultCSRFConfig() CSRFConfig {
	return CSRFConfig{
		CookieName:     CSRFCookieName,
		CookiePath:     "/",
		CookieSecure:   true,
		CookieSameSite: "Strict",
		CookieHTTPOnly: true,
		TokenExpiry:    24 * time.Hour,
		SkipMethods:    []string{"GET", "HEAD", "OPTIONS", "TRACE"},
		Enabled:        true,
	}
}

// CSRFMiddleware provides CSRF protection
type CSRFMiddleware struct {
	config CSRFConfig
}

// NewCSRFMiddleware creates a new CSRF middleware with default config
func NewCSRFMiddleware() *CSRFMiddleware {
	return &CSRFMiddleware{
		config: DefaultCSRFConfig(),
	}
}

// NewCSRFMiddlewareWithConfig creates a new CSRF middleware with custom config
func NewCSRFMiddlewareWithConfig(config CSRFConfig) *CSRFMiddleware {
	// Set defaults for empty fields
	if config.CookieName == "" {
		config.CookieName = CSRFCookieName
	}
	if config.CookiePath == "" {
		config.CookiePath = "/"
	}
	if len(config.SkipMethods) == 0 {
		config.SkipMethods = []string{"GET", "HEAD", "OPTIONS", "TRACE"}
	}

	return &CSRFMiddleware{
		config: config,
	}
}

// Handler returns the CSRF protection middleware handler
func (m *CSRFMiddleware) Handler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Skip if CSRF is disabled
		if !m.config.Enabled {
			return c.Next()
		}

		// Skip for API key authenticated requests (programmatic access)
		if authType, ok := c.Locals(string(ContextKeyAuthType)).(AuthType); ok && authType == AuthTypeAPIKey {
			return c.Next()
		}

		// Check if method should skip validation
		method := c.Method()
		if m.shouldSkipMethod(method) {
			// For safe methods, ensure token exists (generate if needed)
			token := m.getOrCreateToken(c)
			c.Locals(CSRFContextKey, token)
			return c.Next()
		}

		// For state-changing methods, validate the token
		if err := m.validateToken(c); err != nil {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error":   "Forbidden",
				"message": "Invalid or missing CSRF token",
			})
		}

		// Token is valid, generate a new one for the response
		token := m.generateToken()
		m.setTokenCookie(c, token)
		c.Locals(CSRFContextKey, token)

		return c.Next()
	}
}

// GetToken returns a handler that provides the CSRF token to the client
// This should be used on a GET endpoint for SPA applications
func (m *CSRFMiddleware) GetToken() fiber.Handler {
	return func(c *fiber.Ctx) error {
		token := m.getOrCreateToken(c)
		return c.JSON(fiber.Map{
			"csrfToken": token,
		})
	}
}

// shouldSkipMethod checks if the HTTP method should skip CSRF validation
func (m *CSRFMiddleware) shouldSkipMethod(method string) bool {
	for _, skip := range m.config.SkipMethods {
		if strings.EqualFold(method, skip) {
			return true
		}
	}
	return false
}

// validateToken validates the CSRF token from the request
func (m *CSRFMiddleware) validateToken(c *fiber.Ctx) error {
	// Get token from cookie
	cookieToken := c.Cookies(m.config.CookieName)
	if cookieToken == "" {
		return fiber.NewError(fiber.StatusForbidden, "CSRF cookie not found")
	}

	// Get token from header or form
	requestToken := c.Get(CSRFHeaderName)
	if requestToken == "" {
		// Try form field
		requestToken = c.FormValue(CSRFFormFieldName)
	}
	if requestToken == "" {
		return fiber.NewError(fiber.StatusForbidden, "CSRF token not provided")
	}

	// Compare tokens using constant-time comparison to prevent timing attacks
	if subtle.ConstantTimeCompare([]byte(cookieToken), []byte(requestToken)) != 1 {
		return fiber.NewError(fiber.StatusForbidden, "CSRF token mismatch")
	}

	return nil
}

// getOrCreateToken gets the existing CSRF token or creates a new one
func (m *CSRFMiddleware) getOrCreateToken(c *fiber.Ctx) string {
	// Check if token exists in cookie
	token := c.Cookies(m.config.CookieName)
	if token != "" {
		return token
	}

	// Generate new token
	token = m.generateToken()
	m.setTokenCookie(c, token)
	return token
}

// generateToken generates a new CSRF token
func (m *CSRFMiddleware) generateToken() string {
	bytes := make([]byte, CSRFTokenLength)
	if _, err := rand.Read(bytes); err != nil {
		// Fall back to less secure but still random method
		// This should never happen with crypto/rand
		return base64.RawURLEncoding.EncodeToString([]byte(time.Now().String()))
	}
	return base64.RawURLEncoding.EncodeToString(bytes)
}

// setTokenCookie sets the CSRF token cookie
func (m *CSRFMiddleware) setTokenCookie(c *fiber.Ctx, token string) {
	cookie := &fiber.Cookie{
		Name:     m.config.CookieName,
		Value:    token,
		Path:     m.config.CookiePath,
		Domain:   m.config.CookieDomain,
		Secure:   m.config.CookieSecure,
		HTTPOnly: m.config.CookieHTTPOnly,
	}

	// Set SameSite
	switch strings.ToLower(m.config.CookieSameSite) {
	case "strict":
		cookie.SameSite = "Strict"
	case "lax":
		cookie.SameSite = "Lax"
	case "none":
		cookie.SameSite = "None"
	default:
		cookie.SameSite = "Strict"
	}

	// Set expiry
	if m.config.TokenExpiry > 0 {
		cookie.Expires = time.Now().Add(m.config.TokenExpiry)
	}

	c.Cookie(cookie)
}

// GetCSRFToken gets the CSRF token from context
func GetCSRFToken(c *fiber.Ctx) string {
	token, ok := c.Locals(CSRFContextKey).(string)
	if !ok {
		return ""
	}
	return token
}
