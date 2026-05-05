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
	"github.com/stretchr/testify/require"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

type traceProjectReaderStub struct {
	projectID uuid.UUID
	traceID   string
	err       error
}

func (s *traceProjectReaderStub) GetByID(
	_ context.Context,
	projectID uuid.UUID,
	traceID string,
) (*domain.Trace, error) {
	s.projectID = projectID
	s.traceID = traceID
	if s.err != nil {
		return nil, s.err
	}
	return &domain.Trace{ID: traceID, ProjectID: projectID}, nil
}

func TestExtractAPIKey(t *testing.T) {
	t.Parallel()
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
			name: "canonical secret API key from bearer header",
			setupRequest: func(req *http.Request) {
				req.Header.Set(
					"Authorization",
					"Bearer sk-at-0123456789abcdef0123456789abcdef.secret",
				)
			},
			expectedKey: "sk-at-0123456789abcdef0123456789abcdef.secret",
		},
		{
			name: "legacy key pair from basic auth",
			setupRequest: func(req *http.Request) {
				req.Header.Set(
					"Authorization",
					"Basic cGstbGVnYWN5OnNrLWxlZ2FjeQ==",
				)
			},
			expectedKey: "pk-legacy:sk-legacy",
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
	t.Parallel()
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
			name: "JWT token from WebSocket subprotocol",
			setupRequest: func(req *http.Request) {
				req.Header.Set("Sec-WebSocket-Protocol", "agenttrace, jwt.protocol.token")
			},
			expectedToken: "jwt.protocol.token",
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
	t.Parallel()
	t.Run("returns user ID from context", func(t *testing.T) {
		t.Parallel()
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
		t.Parallel()
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
	t.Parallel()
	t.Run("returns project ID from context", func(t *testing.T) {
		t.Parallel()
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
		t.Parallel()
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
	t.Parallel()
	t.Run("returns API key auth type", func(t *testing.T) {
		t.Parallel()
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
		t.Parallel()
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
		t.Parallel()
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
	t.Parallel()
	t.Run("context key values", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, ContextKey("userID"), ContextKeyUserID)
		assert.Equal(t, ContextKey("projectID"), ContextKeyProjectID)
		assert.Equal(t, ContextKey("orgID"), ContextKeyOrgID)
		assert.Equal(t, ContextKey("apiKeyID"), ContextKeyAPIKeyID)
		assert.Equal(t, ContextKey("authType"), ContextKeyAuthType)
	})

	t.Run("auth type values", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, AuthType("api_key"), AuthTypeAPIKey)
		assert.Equal(t, AuthType("jwt"), AuthTypeJWT)
	})
}

func TestRequiredAPIKeyScope(t *testing.T) {
	t.Parallel()

	tests := []struct {
		method   string
		path     string
		expected string
	}{
		{method: fiber.MethodPost, path: "/api/public/ingestion", expected: ""},
		{method: fiber.MethodGet, path: "/api/public/traces/id/scores", expected: "scores:read"},
		{method: fiber.MethodDelete, path: "/api/public/prompts/name", expected: "prompts:delete"},
		{method: fiber.MethodPost, path: "/v1/traces", expected: "traces:write"},
		{method: fiber.MethodPost, path: "/api/public/spans", expected: "observations:write"},
		{method: fiber.MethodPost, path: "/api/public/generations", expected: "observations:write"},
		{method: fiber.MethodPost, path: "/api/public/events", expected: "observations:write"},
		{method: fiber.MethodPost, path: "/v1/metrics", expected: "metrics:write"},
		{method: fiber.MethodPost, path: "/v1/logs", expected: "logs:write"},
		{method: fiber.MethodGet, path: "/api/public/outcomes", expected: "metrics:read"},
		{method: fiber.MethodPost, path: "/api/public/replay-plans/id/execute", expected: "traces:write"},
		{method: fiber.MethodGet, path: "/api/public/eval-hub/packages", expected: "evaluators:read"},
		{method: fiber.MethodGet, path: "/api/public/privacy/capabilities", expected: "metrics:read"},
		{method: fiber.MethodGet, path: "/api/public/marketplace", expected: "admin:write"},
	}

	for _, tt := range tests {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, requiredAPIKeyScope(tt.method, tt.path))
		})
	}
}

func TestRequiredProjectRole(t *testing.T) {
	t.Parallel()
	assert.Equal(t, domain.OrgRoleViewer, requiredProjectRole(fiber.MethodGet, "/api/public/traces"))
	assert.Equal(t, domain.OrgRoleAdmin, requiredProjectRole(fiber.MethodGet, "/api/public/webhooks"))
	assert.Equal(t, domain.OrgRoleMember, requiredProjectRole(fiber.MethodPost, "/api/public/traces"))
	assert.Equal(t, domain.OrgRoleAdmin, requiredProjectRole(fiber.MethodDelete, "/api/public/traces/id"))
}

func TestExtractWebSocketProjectID(t *testing.T) {
	t.Parallel()
	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString(extractWebSocketProjectID(c))
	})
	request := httptest.NewRequestWithContext(context.Background(), "GET", "/test", nil)
	request.Header.Set(
		"Sec-WebSocket-Protocol",
		"agenttrace, jwt.protocol.token, project.00000000-0000-0000-0000-000000000001",
	)

	response, err := app.Test(request)

	require.NoError(t, err)
	defer func() {
		require.NoError(t, response.Body.Close())
	}()
	body, err := io.ReadAll(response.Body)
	require.NoError(t, err)
	assert.Equal(t, "00000000-0000-0000-0000-000000000001", string(body))
}

func TestRequireTraceAccess(t *testing.T) {
	t.Parallel()
	projectID := uuid.New()
	repo := &traceProjectReaderStub{}
	authMiddleware := NewAuthMiddleware(nil, nil)
	app := fiber.New()
	app.Get(
		"/traces/:traceId",
		func(c *fiber.Ctx) error {
			c.Locals(string(ContextKeyProjectID), projectID)
			return c.Next()
		},
		authMiddleware.RequireTraceAccess(repo),
		func(c *fiber.Ctx) error {
			return c.SendStatus(fiber.StatusNoContent)
		},
	)

	request := httptest.NewRequestWithContext(
		context.Background(),
		"GET",
		"/traces/trace-1",
		nil,
	)
	response, err := app.Test(request)

	require.NoError(t, err)
	defer func() {
		require.NoError(t, response.Body.Close())
	}()
	assert.Equal(t, fiber.StatusNoContent, response.StatusCode)
	assert.Equal(t, projectID, repo.projectID)
	assert.Equal(t, "trace-1", repo.traceID)
}

func TestHasAPIKeyScope(t *testing.T) {
	t.Parallel()
	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		c.Locals(string(ContextKeyAuthType), AuthTypeAPIKey)
		c.Locals(string(ContextKeyAPIKeyScopes), []string{"traces:read"})
		return c.JSON(fiber.Map{
			"read":  HasAPIKeyScope(c, "traces:read"),
			"write": HasAPIKeyScope(c, "traces:write"),
		})
	})

	request := httptest.NewRequestWithContext(context.Background(), "GET", "/test", nil)
	response, err := app.Test(request)

	require.NoError(t, err)
	defer func() {
		require.NoError(t, response.Body.Close())
	}()
	body, err := io.ReadAll(response.Body)
	require.NoError(t, err)
	assert.JSONEq(t, `{"read":true,"write":false}`, string(body))
}

func TestNewAuthMiddleware(t *testing.T) {
	t.Parallel()
	t.Run("creates auth middleware", func(t *testing.T) {
		t.Parallel()
		// Note: In a real test we'd mock the AuthService
		// For this unit test we just verify the constructor works
		middleware := NewAuthMiddleware(nil, nil)
		assert.NotNil(t, middleware)
	})
}

func TestRequireAPIKeyHandler(t *testing.T) {
	t.Parallel()
	// Note: Full integration tests would require setting up the full
	// AuthService with database. These tests verify the middleware
	// returns appropriate error responses when no API key is provided.

	t.Run("returns 401 when no API key provided", func(t *testing.T) {
		t.Parallel()
		app := fiber.New()

		middleware := NewAuthMiddleware(nil, nil)
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
	t.Parallel()
	t.Run("returns 401 when no JWT provided", func(t *testing.T) {
		t.Parallel()
		app := fiber.New()

		middleware := NewAuthMiddleware(nil, nil)
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
	t.Parallel()
	t.Run("returns 401 when no auth provided", func(t *testing.T) {
		t.Parallel()
		app := fiber.New()

		middleware := NewAuthMiddleware(nil, nil)
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
	t.Parallel()
	t.Run("continues without auth", func(t *testing.T) {
		t.Parallel()
		app := fiber.New()

		middleware := NewAuthMiddleware(nil, nil)
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

func TestSetAPIKeyContext(t *testing.T) {
	t.Parallel()

	t.Run("owned API key exposes creator as user context", func(t *testing.T) {
		t.Parallel()
		app := fiber.New()
		ownerID := uuid.New()
		apiKeyID := uuid.New()
		projectID := uuid.New()

		app.Get("/test", func(c *fiber.Ctx) error {
			setAPIKeyContext(c, &domain.APIKeyContext{
				APIKeyID:  apiKeyID,
				ProjectID: projectID,
				Scopes:    []string{"traces:write"},
				UserID:    &ownerID,
			})

			userID, ok := GetUserID(c)
			assert.True(t, ok, "owned key must expose a user context")
			assert.Equal(t, ownerID, userID)

			keyID, ok := GetAPIKeyID(c)
			assert.True(t, ok)
			assert.Equal(t, apiKeyID, keyID)

			pid, ok := GetProjectID(c)
			assert.True(t, ok)
			assert.Equal(t, projectID, pid)
			return c.SendStatus(200)
		})

		resp, err := app.Test(httptest.NewRequestWithContext(context.Background(), "GET", "/test", nil))
		require.NoError(t, err)
		require.NoError(t, resp.Body.Close())
	})

	t.Run("unowned legacy API key leaves user context unset", func(t *testing.T) {
		t.Parallel()
		app := fiber.New()
		apiKeyID := uuid.New()

		app.Get("/test", func(c *fiber.Ctx) error {
			setAPIKeyContext(c, &domain.APIKeyContext{
				APIKeyID:  apiKeyID,
				ProjectID: uuid.New(),
				Scopes:    []string{"traces:write"},
				UserID:    nil,
			})

			_, ok := GetUserID(c)
			assert.False(t, ok, "unowned key must not fabricate a user context")

			keyID, ok := GetAPIKeyID(c)
			assert.True(t, ok)
			assert.Equal(t, apiKeyID, keyID)
			return c.SendStatus(200)
		})

		resp, err := app.Test(httptest.NewRequestWithContext(context.Background(), "GET", "/test", nil))
		require.NoError(t, err)
		require.NoError(t, resp.Body.Close())
	})
}
