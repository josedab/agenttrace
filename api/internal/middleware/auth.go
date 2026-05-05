package middleware

import (
	"context"
	"encoding/base64"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// ContextKey type for context keys
type ContextKey string

const (
	// ContextKeyUserID stores the authenticated user ID.
	ContextKeyUserID ContextKey = "userID"
	// ContextKeyProjectID stores the authorized project ID.
	ContextKeyProjectID ContextKey = "projectID"
	// ContextKeyOrgID stores the authorized organization ID.
	ContextKeyOrgID ContextKey = "orgID"
	// ContextKeyAPIKeyID stores the authenticated API-key ID.
	ContextKeyAPIKeyID ContextKey = "apiKeyID"
	// ContextKeyAPIKeyScopes stores the authenticated API-key scopes.
	ContextKeyAPIKeyScopes ContextKey = "apiKeyScopes"
	// ContextKeyAuthType stores the authentication mechanism.
	ContextKeyAuthType ContextKey = "authType"
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
	orgService  *service.OrgService
}

type traceProjectReader interface {
	GetByID(ctx context.Context, projectID uuid.UUID, traceID string) (*domain.Trace, error)
}

// NewAuthMiddleware creates a new auth middleware
func NewAuthMiddleware(authService *service.AuthService, orgService *service.OrgService) *AuthMiddleware {
	return &AuthMiddleware{
		authService: authService,
		orgService:  orgService,
	}
}

// RequireAPIKey validates API key authentication
func (m *AuthMiddleware) RequireAPIKey() fiber.Handler {
	return func(c *fiber.Ctx) error {
		credential := extractAPIKey(c)
		if credential == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "Unauthorized",
				"message": "API key required",
			})
		}

		apiKeyContext, err := m.authService.AuthenticateAPIKey(c.Context(), credential)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "Unauthorized",
				"message": "Invalid API key",
			})
		}
		if requiredScope := requiredAPIKeyScope(c.Method(), c.Path()); requiredScope != "" &&
			!apiKeyContext.HasScope(requiredScope) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error":   "Forbidden",
				"message": "API key scope does not permit this operation",
			})
		}

		setAPIKeyContext(c, apiKeyContext)

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
		credential := extractAPIKey(c)
		if credential != "" {
			apiKeyContext, err := m.authService.AuthenticateAPIKey(c.Context(), credential)
			if err == nil {
				if requiredScope := requiredAPIKeyScope(c.Method(), c.Path()); requiredScope != "" &&
					!apiKeyContext.HasScope(requiredScope) {
					return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
						"error":   "Forbidden",
						"message": "API key scope does not permit this operation",
					})
				}
				setAPIKeyContext(c, apiKeyContext)
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

					projectIDValue := c.Get("X-Project-ID")
					if projectIDValue == "" {
						projectIDValue = extractWebSocketProjectID(c)
					}
					if projectIDValue == "" {
						return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
							"error":   "Bad Request",
							"message": "Project context required",
						})
					}
					projectID, err := uuid.Parse(projectIDValue)
					if err != nil {
						return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
							"error":   "Bad Request",
							"message": "Invalid project ID",
						})
					}
					if err := m.authService.CheckProjectAccess(
						c.Context(),
						projectID,
						userID,
						requiredProjectRole(c.Method(), c.Path()),
					); err != nil {
						return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
							"error":   "Forbidden",
							"message": "Project access denied",
						})
					}
					c.Locals(string(ContextKeyProjectID), projectID)

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

// RequireTraceAccess verifies that the authenticated project can read the trace.
func (m *AuthMiddleware) RequireTraceAccess(traceRepo traceProjectReader) fiber.Handler {
	return func(c *fiber.Ctx) error {
		projectID, ok := GetProjectID(c)
		if !ok {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Bad Request",
				"message": "Project context required",
			})
		}
		traceID := c.Params("traceId")
		if traceID == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Bad Request",
				"message": "Trace ID required",
			})
		}
		if _, err := traceRepo.GetByID(c.Context(), projectID, traceID); err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   "Not Found",
				"message": "Trace not found",
			})
		}
		return c.Next()
	}
}

// RequireProjectAccess ensures the user has access to the specified project
func (m *AuthMiddleware) RequireProjectAccess() fiber.Handler {
	return m.RequireProjectRole(domain.OrgRoleViewer)
}

// RequireProjectRole ensures the user has the requested project role.
func (m *AuthMiddleware) RequireProjectRole(requiredRole domain.OrgRole) fiber.Handler {
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

		if err := m.authService.CheckProjectAccess(c.Context(), projectID, userID, requiredRole); err != nil {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error":   "Forbidden",
				"message": "No access to this project",
			})
		}

		c.Locals(string(ContextKeyProjectID), projectID)
		return c.Next()
	}
}

func requiredProjectRole(method, path string) domain.OrgRole {
	switch method {
	case fiber.MethodGet, fiber.MethodHead, fiber.MethodOptions:
		if requiredAPIKeyScope(method, path) == "admin:write" {
			return domain.OrgRoleAdmin
		}
		return domain.OrgRoleViewer
	case fiber.MethodDelete:
		return domain.OrgRoleAdmin
	default:
		return domain.OrgRoleMember
	}
}

// RequireOrgAccess ensures the authenticated user has at least the specified role in the organization.
// The organization ID is extracted from the :orgId route parameter.
func (m *AuthMiddleware) RequireOrgAccess(requiredRole domain.OrgRole) fiber.Handler {
	return func(c *fiber.Ctx) error {
		orgIDParam := c.Params("orgId")
		if orgIDParam == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Bad Request",
				"message": "Organization ID required",
			})
		}

		orgID, err := uuid.Parse(orgIDParam)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Bad Request",
				"message": "Invalid organization ID",
			})
		}

		userID, ok := c.Locals(string(ContextKeyUserID)).(uuid.UUID)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "Unauthorized",
				"message": "User not authenticated",
			})
		}

		if err := m.orgService.CheckAccess(c.Context(), orgID, userID, requiredRole); err != nil {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error":   "Forbidden",
				"message": "No access to this organization",
			})
		}

		c.Locals(string(ContextKeyOrgID), orgID)
		return c.Next()
	}
}

// OptionalAuth tries to authenticate but continues even if it fails
func (m *AuthMiddleware) OptionalAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Try API key first
		credential := extractAPIKey(c)
		if credential != "" {
			apiKeyContext, err := m.authService.AuthenticateAPIKey(c.Context(), credential)
			if err == nil {
				setAPIKeyContext(c, apiKeyContext)
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
		token := strings.TrimSpace(strings.TrimPrefix(auth, "Bearer "))
		if isAPIKeyCredential(token) {
			return token
		}
	}
	if strings.HasPrefix(auth, "Basic ") {
		encoded := strings.TrimSpace(strings.TrimPrefix(auth, "Basic "))
		decoded, err := base64.StdEncoding.DecodeString(encoded)
		if err == nil {
			publicKey, secretKey, found := strings.Cut(string(decoded), ":")
			if found && publicKey != "" && secretKey != "" {
				return publicKey + ":" + secretKey
			}
		}
	}

	// Check X-API-Key header
	if apiKey := c.Get("X-API-Key"); apiKey != "" {
		return strings.TrimSpace(apiKey)
	}
	return ""
}

func isAPIKeyCredential(token string) bool {
	return strings.HasPrefix(token, "sk-at-") ||
		strings.HasPrefix(token, "at_") ||
		strings.HasPrefix(token, "pk-") ||
		strings.HasPrefix(token, "pk_")
}

func setAPIKeyContext(c *fiber.Ctx, apiKeyContext *domain.APIKeyContext) {
	c.Locals(string(ContextKeyProjectID), apiKeyContext.ProjectID)
	c.Locals(string(ContextKeyAPIKeyID), apiKeyContext.APIKeyID)
	c.Locals(string(ContextKeyAPIKeyScopes), apiKeyContext.Scopes)
	c.Locals(string(ContextKeyAuthType), AuthTypeAPIKey)
	// When the API key is owned by a real user, expose that identity as the
	// normal user context so that user-foreign-key constrained operations
	// attribute to the owner instead of the API key UUID. Legacy/unowned keys
	// leave the user context unset so callers can reject them explicitly rather
	// than triggering a foreign-key violation.
	if apiKeyContext.UserID != nil {
		c.Locals(string(ContextKeyUserID), *apiKeyContext.UserID)
	}
}

func requiredAPIKeyScope(method, path string) string {
	if strings.Trim(path, "/") == "api/public/ingestion" {
		return ""
	}
	resource := apiKeyScopeResource(path)
	if resource == "" {
		return "admin:write"
	}

	action := "write"
	switch method {
	case fiber.MethodGet, fiber.MethodHead, fiber.MethodOptions:
		action = "read"
	case fiber.MethodDelete:
		action = "delete"
	}
	return resource + ":" + action
}

func apiKeyScopeResource(path string) string {
	path = strings.Trim(path, "/")
	path = strings.TrimPrefix(path, "api/public/")
	path = strings.TrimPrefix(path, "v1/")
	segments := strings.Split(path, "/")

	for _, candidate := range []string{
		"scores",
		"observations",
		"spans",
		"generations",
		"prompts",
		"datasets",
		"evaluators",
		"exports",
		"metrics",
		"logs",
	} {
		for _, segment := range segments {
			if segment == candidate || (candidate == "exports" && segment == "export") {
				if candidate == "spans" || candidate == "generations" {
					return "observations"
				}
				return candidate
			}
		}
	}

	if len(segments) == 0 {
		return ""
	}
	switch segments[0] {
	case "traces", "sessions", "checkpoints", "git-links", "file-operations", "terminal-commands", "ci-runs", "replay", "replay-plans", "ws":
		return "traces"
	case "outcomes":
		return "metrics"
	case "eval-hub", "eval-marketplace":
		return "evaluators"
	case "share-links":
		return "traces"
	case "privacy":
		if len(segments) > 1 && segments[1] == "capabilities" {
			return "metrics"
		}
		return ""
	case "events":
		return "observations"
	case "feedback":
		return "scores"
	default:
		return ""
	}
}

// extractBearerToken extracts JWT from Authorization header
func extractBearerToken(c *fiber.Ctx) string {
	auth := c.Get("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		token := strings.TrimPrefix(auth, "Bearer ")
		// JWT tokens don't start with API key prefixes
		if !strings.HasPrefix(token, "at_") &&
			!strings.HasPrefix(token, "pk-") &&
			!strings.HasPrefix(token, "pk_") &&
			!strings.HasPrefix(token, "sk-") &&
			!strings.HasPrefix(token, "sk_") {
			return token
		}
	}
	protocols := strings.Split(c.Get("Sec-WebSocket-Protocol"), ",")
	for index, protocol := range protocols {
		if strings.TrimSpace(protocol) == "agenttrace" && index+1 < len(protocols) {
			if token := strings.TrimSpace(protocols[index+1]); token != "" {
				return token
			}
		}
	}
	return ""
}

func extractWebSocketProjectID(c *fiber.Ctx) string {
	for _, protocol := range strings.Split(c.Get("Sec-WebSocket-Protocol"), ",") {
		protocol = strings.TrimSpace(protocol)
		if strings.HasPrefix(protocol, "project.") {
			return strings.TrimPrefix(protocol, "project.")
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

// HasAPIKeyScope reports whether the authenticated request may use a scope.
// JWT requests are authorized by project role instead.
func HasAPIKeyScope(c *fiber.Ctx, scope string) bool {
	authType, _ := GetAuthType(c)
	if authType != AuthTypeAPIKey {
		return true
	}
	scopes, ok := c.Locals(string(ContextKeyAPIKeyScopes)).([]string)
	if !ok {
		return false
	}
	return (&domain.APIKeyContext{Scopes: scopes}).HasScope(scope)
}

// GetOrganizationID gets the organization ID from context
func GetOrganizationID(c *fiber.Ctx) (uuid.UUID, bool) {
	orgID, ok := c.Locals(string(ContextKeyOrgID)).(uuid.UUID)
	return orgID, ok
}

// GetAuthType gets the authentication type from context
func GetAuthType(c *fiber.Ctx) (AuthType, bool) {
	authType, ok := c.Locals(string(ContextKeyAuthType)).(AuthType)
	return authType, ok
}
