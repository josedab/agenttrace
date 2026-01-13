package middleware

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

// MockAuthService mocks the AuthService for testing
type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) ValidateAPIKeyPublicOnly(ctx context.Context, key string) (*uuid.UUID, error) {
	args := m.Called(ctx, key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	id := args.Get(0).(uuid.UUID)
	return &id, args.Error(1)
}

func (m *MockAuthService) ValidateJWT(ctx context.Context, token string) (*domain.JWTClaims, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.JWTClaims), args.Error(1)
}

func (m *MockAuthService) CheckProjectAccess(ctx context.Context, projectID, userID uuid.UUID, role domain.OrgRole) error {
	args := m.Called(ctx, projectID, userID, role)
	return args.Error(0)
}

func TestExtractAPIKey(t *testing.T) {
	tests := []struct {
		name          string
		setupRequest  func(*http.Request)
		expectedKey   string
		expectedEmpty bool
	}{
		{
			name: "API key from Bearer header with at_ prefix",
			setupRequest: func(req *http.Request) {
				req.Header.Set("Authorization", "Bearer at_test_key_123")
			},
			expectedKey: "at_test_key_123",
		},
		{
			name: "API key from Bearer header with pk- prefix",
			setupRequest: func(req *http.Request) {
				req.Header.Set("Authorization", "Bearer pk-test_key_456")
			},
			expectedKey: "pk-test_key_456",
		},
		{
			name: "API key from Bearer header with pk_ prefix",
			setupRequest: func(req *http.Request) {
				req.Header.Set("Authorization", "Bearer pk_test_key_789")
			},
			expectedKey: "pk_test_key_789",
		},
		{
			name: "API key from X-API-Key header",
			setupRequest: func(req *http.Request) {
				req.Header.Set("X-API-Key", "sk_secret_key")
			},
			expectedKey: "sk_secret_key",
		},
		{
			name: "No API key - Bearer token is JWT (no prefix)",
			setupRequest: func(req *http.Request) {
				req.Header.Set("Authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9")
			},
			expectedEmpty: true,
		},
		{
			name:          "No Authorization header",
			setupRequest:  func(req *http.Request) {},
			expectedEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()

			var extractedKey string
			app.Get("/test", func(c *fiber.Ctx) error {
				extractedKey = extractAPIKey(c)
				return c.SendStatus(200)
			})

			req := httptest.NewRequest("GET", "/test", nil)
			tt.setupRequest(req)

			_, err := app.Test(req)
			require.NoError(t, err)

			if tt.expectedEmpty {
				assert.Empty(t, extractedKey)
			} else {
				assert.Equal(t, tt.expectedKey, extractedKey)
			}
		})
	}
}

func TestExtractBearerToken(t *testing.T) {
	tests := []struct {
		name          string
		setupRequest  func(*http.Request)
		expectedToken string
		expectedEmpty bool
	}{
		{
			name: "JWT token from Bearer header",
			setupRequest: func(req *http.Request) {
				req.Header.Set("Authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.dozjgNryP4J3jVmNHl0w5N_XgL0n3I9PlFUP0THsR8U")
			},
			expectedToken: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.dozjgNryP4J3jVmNHl0w5N_XgL0n3I9PlFUP0THsR8U",
		},
		{
			name: "API key token not returned (at_ prefix)",
			setupRequest: func(req *http.Request) {
				req.Header.Set("Authorization", "Bearer at_api_key")
			},
			expectedEmpty: true,
		},
		{
			name: "API key token not returned (pk- prefix)",
			setupRequest: func(req *http.Request) {
				req.Header.Set("Authorization", "Bearer pk-api_key")
			},
			expectedEmpty: true,
		},
		{
			name: "API key token not returned (sk- prefix)",
			setupRequest: func(req *http.Request) {
				req.Header.Set("Authorization", "Bearer sk-api_key")
			},
			expectedEmpty: true,
		},
		{
			name:          "No Authorization header",
			setupRequest:  func(req *http.Request) {},
			expectedEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()

			var extractedToken string
			app.Get("/test", func(c *fiber.Ctx) error {
				extractedToken = extractBearerToken(c)
				return c.SendStatus(200)
			})

			req := httptest.NewRequest("GET", "/test", nil)
			tt.setupRequest(req)

			_, err := app.Test(req)
			require.NoError(t, err)

			if tt.expectedEmpty {
				assert.Empty(t, extractedToken)
			} else {
				assert.Equal(t, tt.expectedToken, extractedToken)
			}
		})
	}
}

func TestGetUserID(t *testing.T) {
	t.Run("returns user ID from context", func(t *testing.T) {
		app := fiber.New()
		userID := uuid.New()

		app.Get("/test", func(c *fiber.Ctx) error {
			c.Locals(string(ContextKeyUserID), userID)
			id, ok := GetUserID(c)
			assert.True(t, ok)
			assert.Equal(t, userID, id)
			return c.SendStatus(200)
		})

		req := httptest.NewRequest("GET", "/test", nil)
		_, err := app.Test(req)
		require.NoError(t, err)
	})

	t.Run("returns false when user ID not in context", func(t *testing.T) {
		app := fiber.New()

		app.Get("/test", func(c *fiber.Ctx) error {
			id, ok := GetUserID(c)
			assert.False(t, ok)
			assert.Equal(t, uuid.UUID{}, id)
			return c.SendStatus(200)
		})

		req := httptest.NewRequest("GET", "/test", nil)
		_, err := app.Test(req)
		require.NoError(t, err)
	})
}

func TestGetProjectID(t *testing.T) {
	t.Run("returns project ID from context", func(t *testing.T) {
		app := fiber.New()
		projectID := uuid.New()

		app.Get("/test", func(c *fiber.Ctx) error {
			c.Locals(string(ContextKeyProjectID), projectID)
			id, ok := GetProjectID(c)
			assert.True(t, ok)
			assert.Equal(t, projectID, id)
			return c.SendStatus(200)
		})

		req := httptest.NewRequest("GET", "/test", nil)
		_, err := app.Test(req)
		require.NoError(t, err)
	})

	t.Run("returns false when project ID not in context", func(t *testing.T) {
		app := fiber.New()

		app.Get("/test", func(c *fiber.Ctx) error {
			id, ok := GetProjectID(c)
			assert.False(t, ok)
			assert.Equal(t, uuid.UUID{}, id)
			return c.SendStatus(200)
		})

		req := httptest.NewRequest("GET", "/test", nil)
		_, err := app.Test(req)
		require.NoError(t, err)
	})
}

func TestGetAuthType(t *testing.T) {
	t.Run("returns API key auth type", func(t *testing.T) {
		app := fiber.New()

		app.Get("/test", func(c *fiber.Ctx) error {
			c.Locals(string(ContextKeyAuthType), AuthTypeAPIKey)
			authType, ok := GetAuthType(c)
			assert.True(t, ok)
			assert.Equal(t, AuthTypeAPIKey, authType)
			return c.SendStatus(200)
		})

		req := httptest.NewRequest("GET", "/test", nil)
		_, err := app.Test(req)
		require.NoError(t, err)
	})

	t.Run("returns JWT auth type", func(t *testing.T) {
		app := fiber.New()

		app.Get("/test", func(c *fiber.Ctx) error {
			c.Locals(string(ContextKeyAuthType), AuthTypeJWT)
			authType, ok := GetAuthType(c)
			assert.True(t, ok)
			assert.Equal(t, AuthTypeJWT, authType)
			return c.SendStatus(200)
		})

		req := httptest.NewRequest("GET", "/test", nil)
		_, err := app.Test(req)
		require.NoError(t, err)
	})

	t.Run("returns false when auth type not in context", func(t *testing.T) {
		app := fiber.New()

		app.Get("/test", func(c *fiber.Ctx) error {
			authType, ok := GetAuthType(c)
			assert.False(t, ok)
			assert.Empty(t, authType)
			return c.SendStatus(200)
		})

		req := httptest.NewRequest("GET", "/test", nil)
		_, err := app.Test(req)
		require.NoError(t, err)
	})
}

func TestAuthConstants(t *testing.T) {
	t.Run("context key values", func(t *testing.T) {
		assert.Equal(t, ContextKey("userID"), ContextKeyUserID)
		assert.Equal(t, ContextKey("projectID"), ContextKeyProjectID)
		assert.Equal(t, ContextKey("orgID"), ContextKeyOrgID)
		assert.Equal(t, ContextKey("apiKeyID"), ContextKeyAPIKeyID)
		assert.Equal(t, ContextKey("authType"), ContextKeyAuthType)
	})

	t.Run("auth type values", func(t *testing.T) {
		assert.Equal(t, AuthType("api_key"), AuthTypeAPIKey)
		assert.Equal(t, AuthType("jwt"), AuthTypeJWT)
	})
}

func TestNewAuthMiddleware(t *testing.T) {
	t.Run("creates auth middleware", func(t *testing.T) {
		// Note: In a real test we'd mock the AuthService
		// For this unit test we just verify the constructor works
		middleware := NewAuthMiddleware(nil)
		assert.NotNil(t, middleware)
	})
}

func TestRequireAPIKeyHandler(t *testing.T) {
	// Note: Full integration tests would require setting up the full
	// AuthService with database. These tests verify the middleware
	// returns appropriate error responses when no API key is provided.

	t.Run("returns 401 when no API key provided", func(t *testing.T) {
		app := fiber.New()

		middleware := NewAuthMiddleware(nil)
		app.Use(middleware.RequireAPIKey())
		app.Get("/test", func(c *fiber.Ctx) error {
			return c.SendStatus(200)
		})

		req := httptest.NewRequest("GET", "/test", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(body), "API key required")
	})
}

func TestRequireJWTHandler(t *testing.T) {
	t.Run("returns 401 when no JWT provided", func(t *testing.T) {
		app := fiber.New()

		middleware := NewAuthMiddleware(nil)
		app.Use(middleware.RequireJWT())
		app.Get("/test", func(c *fiber.Ctx) error {
			return c.SendStatus(200)
		})

		req := httptest.NewRequest("GET", "/test", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(body), "Authorization header required")
	})
}

func TestRequireAuthHandler(t *testing.T) {
	t.Run("returns 401 when no auth provided", func(t *testing.T) {
		app := fiber.New()

		middleware := NewAuthMiddleware(nil)
		app.Use(middleware.RequireAuth())
		app.Get("/test", func(c *fiber.Ctx) error {
			return c.SendStatus(200)
		})

		req := httptest.NewRequest("GET", "/test", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(body), "Valid authentication required")
	})
}

func TestOptionalAuthHandler(t *testing.T) {
	t.Run("continues without auth", func(t *testing.T) {
		app := fiber.New()

		middleware := NewAuthMiddleware(nil)
		app.Use(middleware.OptionalAuth())
		app.Get("/test", func(c *fiber.Ctx) error {
			return c.SendStatus(200)
		})

		req := httptest.NewRequest("GET", "/test", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)

		// Optional auth should allow the request through
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	})
}
