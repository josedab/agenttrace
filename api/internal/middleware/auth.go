package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// ContextKey type for context keys
type ContextKey string

const (
	// Context keys
	ContextKeyUserID    ContextKey = "userID"
	ContextKeyProjectID ContextKey = "projectID"
	ContextKeyOrgID     ContextKey = "orgID"
	ContextKeyAPIKeyID  ContextKey = "apiKeyID"
	ContextKeyAuthType  ContextKey = "authType"
)

// AuthType represents the type of authentication used
type AuthType string

const (
	AuthTypeAPIKey AuthType = "api_key"
	AuthTypeJWT    AuthType = "jwt"
)

// AuthMiddleware handles authentication
type AuthMiddleware struct {
	authService *service.AuthService
}

// NewAuthMiddleware creates a new auth middleware
func NewAuthMiddleware(authService *service.AuthService) *AuthMiddleware {
	return &AuthMiddleware{
		authService: authService,
	}
}

// RequireAPIKey validates API key authentication
func (m *AuthMiddleware) RequireAPIKey() fiber.Handler {
	return func(c *fiber.Ctx) error {
		apiKey := extractAPIKey(c)
		if apiKey == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "Unauthorized",
				"message": "API key required",
			})
		}

		// For public key only validation (no secret key)
		projectID, err := m.authService.ValidateAPIKeyPublicOnly(c.Context(), apiKey)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "Unauthorized",
				"message": "Invalid API key",
			})
		}

		// Set context values
		c.Locals(string(ContextKeyProjectID), *projectID)
		c.Locals(string(ContextKeyAuthType), AuthTypeAPIKey)

		return c.Next()
	}
}

// RequireJWT validates JWT authentication
func (m *AuthMiddleware) RequireJWT() fiber.Handler {
	return func(c *fiber.Ctx) error {
		token := extractBearerToken(c)
		if token == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "Unauthorized",
				"message": "Authorization header required",
			})
		}

		claims, err := m.authService.ValidateJWT(c.Context(), token)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "Unauthorized",
				"message": "Invalid or expired token",
			})
		}

		// Parse user ID from claims
		userID, err := uuid.Parse(claims.UserID)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "Unauthorized",
				"message": "Invalid user ID in token",
			})
		}

		// Set context values
		c.Locals(string(ContextKeyUserID), userID)
		c.Locals(string(ContextKeyAuthType), AuthTypeJWT)

		return c.Next()
	}
}

// RequireAuth validates either API key or JWT authentication
func (m *AuthMiddleware) RequireAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Try API key first
		apiKey := extractAPIKey(c)
		if apiKey != "" {
			projectID, err := m.authService.ValidateAPIKeyPublicOnly(c.Context(), apiKey)
			if err == nil {
				c.Locals(string(ContextKeyProjectID), *projectID)
				c.Locals(string(ContextKeyAuthType), AuthTypeAPIKey)
				return c.Next()
			}
		}

		// Try JWT
		token := extractBearerToken(c)
		if token != "" {
			claims, err := m.authService.ValidateJWT(c.Context(), token)
			if err == nil {
				if userID, err := uuid.Parse(claims.UserID); err == nil {
					c.Locals(string(ContextKeyUserID), userID)
					c.Locals(string(ContextKeyAuthType), AuthTypeJWT)
					return c.Next()
				}
			}
		}

		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Valid authentication required",
		})
	}
}

// RequireProjectAccess ensures the user has access to the specified project
func (m *AuthMiddleware) RequireProjectAccess() fiber.Handler {
	return func(c *fiber.Ctx) error {
		projectIDParam := c.Params("projectId")
		if projectIDParam == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Bad Request",
				"message": "Project ID required",
			})
		}

		projectID, err := uuid.Parse(projectIDParam)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Bad Request",
				"message": "Invalid project ID",
			})
		}

		// Check if using API key (already validated for project)
		if authType, ok := c.Locals(string(ContextKeyAuthType)).(AuthType); ok && authType == AuthTypeAPIKey {
			keyProjectID, ok := c.Locals(string(ContextKeyProjectID)).(uuid.UUID)
			if ok && keyProjectID == projectID {
				return c.Next()
			}
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error":   "Forbidden",
				"message": "API key not valid for this project",
			})
		}

		// Check JWT user has access
		userID, ok := c.Locals(string(ContextKeyUserID)).(uuid.UUID)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "Unauthorized",
				"message": "User not authenticated",
			})
		}

		if err := m.authService.CheckProjectAccess(c.Context(), projectID, userID, domain.OrgRoleViewer); err != nil {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error":   "Forbidden",
				"message": "No access to this project",
			})
		}

		c.Locals(string(ContextKeyProjectID), projectID)
		return c.Next()
	}
}

// OptionalAuth tries to authenticate but continues even if it fails
func (m *AuthMiddleware) OptionalAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Try API key first
		apiKey := extractAPIKey(c)
		if apiKey != "" {
			projectID, err := m.authService.ValidateAPIKeyPublicOnly(c.Context(), apiKey)
			if err == nil {
				c.Locals(string(ContextKeyProjectID), *projectID)
				c.Locals(string(ContextKeyAuthType), AuthTypeAPIKey)
				return c.Next()
			}
		}

		// Try JWT
		token := extractBearerToken(c)
		if token != "" {
			claims, err := m.authService.ValidateJWT(c.Context(), token)
			if err == nil {
				if userID, err := uuid.Parse(claims.UserID); err == nil {
					c.Locals(string(ContextKeyUserID), userID)
					c.Locals(string(ContextKeyAuthType), AuthTypeJWT)
				}
			}
		}

		return c.Next()
	}
}

// extractAPIKey extracts API key from request
func extractAPIKey(c *fiber.Ctx) string {
	// Check Authorization header with Bearer prefix
	auth := c.Get("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		token := strings.TrimPrefix(auth, "Bearer ")
		// API keys start with "at_" prefix
		if strings.HasPrefix(token, "at_") {
			return token
		}
	}

	// Check X-API-Key header
	if apiKey := c.Get("X-API-Key"); apiKey != "" {
		return apiKey
	}

	// Check query parameter
	if apiKey := c.Query("api_key"); apiKey != "" {
		return apiKey
	}

	return ""
}

// extractBearerToken extracts JWT from Authorization header
func extractBearerToken(c *fiber.Ctx) string {
	auth := c.Get("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		token := strings.TrimPrefix(auth, "Bearer ")
		// JWT tokens don't start with "at_"
		if !strings.HasPrefix(token, "at_") {
			return token
		}
	}
	return ""
}

// GetUserID gets the user ID from context
func GetUserID(c *fiber.Ctx) (uuid.UUID, bool) {
	userID, ok := c.Locals(string(ContextKeyUserID)).(uuid.UUID)
	return userID, ok
}

// GetProjectID gets the project ID from context
func GetProjectID(c *fiber.Ctx) (uuid.UUID, bool) {
	projectID, ok := c.Locals(string(ContextKeyProjectID)).(uuid.UUID)
	return projectID, ok
}

// GetAPIKeyID gets the API key ID from context
func GetAPIKeyID(c *fiber.Ctx) (uuid.UUID, bool) {
	apiKeyID, ok := c.Locals(string(ContextKeyAPIKeyID)).(uuid.UUID)
	return apiKeyID, ok
}

// GetAuthType gets the authentication type from context
func GetAuthType(c *fiber.Ctx) (AuthType, bool) {
	authType, ok := c.Locals(string(ContextKeyAuthType)).(AuthType)
	return authType, ok
}
