package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/middleware"
	apperrors "github.com/agenttrace/agenttrace/api/internal/pkg/errors"
	"github.com/agenttrace/agenttrace/api/internal/testutil"
)

// MockAuthService mocks the auth service for testing.
type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) Login(ctx context.Context, input *domain.LoginInput) (*domain.AuthResult, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.AuthResult), args.Error(1)
}

func (m *MockAuthService) Register(ctx context.Context, input *domain.RegisterInput) (*domain.AuthResult, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.AuthResult), args.Error(1)
}

func (m *MockAuthService) GetUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockAuthService) RefreshToken(ctx context.Context, refreshToken string) (*domain.AuthResult, error) {
	args := m.Called(ctx, refreshToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.AuthResult), args.Error(1)
}

func (m *MockAuthService) Logout(ctx context.Context, refreshToken string) error {
	args := m.Called(ctx, refreshToken)
	return args.Error(0)
}

func (m *MockAuthService) HandleOAuthCallback(ctx context.Context, input *domain.OAuthCallbackInput) (*domain.AuthResult, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.AuthResult), args.Error(1)
}

func setupAuthTestApp(mockSvc *MockAuthService, userID *uuid.UUID) *fiber.App {
	app := fiber.New()
	logger := zap.NewNop()

	if userID != nil {
		app.Use(testutil.TestUserMiddleware(*userID))
	}

	// Login
	app.Post("/auth/login", func(c *fiber.Ctx) error {
		var input struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}

		if err := c.BodyParser(&input); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Bad Request",
				"message": "Invalid request body: " + err.Error(),
			})
		}

		if input.Email == "" || input.Password == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Bad Request",
				"message": "Email and password are required",
			})
		}

		loginInput := &domain.LoginInput{
			Email:    input.Email,
			Password: input.Password,
		}

		result, err := mockSvc.Login(c.Context(), loginInput)
		if err != nil {
			if apperrors.IsUnauthorized(err) {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"error":   "Unauthorized",
					"message": "Invalid email or password",
				})
			}
			logger.Error("login failed", zap.Error(err))
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "Internal Server Error",
				"message": "Login failed",
			})
		}

		return c.JSON(fiber.Map{
			"accessToken":  result.AccessToken,
			"refreshToken": result.RefreshToken,
			"expiresAt":    result.ExpiresAt,
			"user": fiber.Map{
				"id":    result.User.ID,
				"email": result.User.Email,
				"name":  result.User.Name,
			},
		})
	})

	// Register
	app.Post("/auth/register", func(c *fiber.Ctx) error {
		var input struct {
			Email    string `json:"email"`
			Password string `json:"password"`
			Name     string `json:"name"`
		}

		if err := c.BodyParser(&input); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Bad Request",
				"message": "Invalid request body: " + err.Error(),
			})
		}

		if input.Email == "" || input.Password == "" || input.Name == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Bad Request",
				"message": "Email, password, and name are required",
			})
		}

		if len(input.Password) < 8 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Bad Request",
				"message": "Password must be at least 8 characters",
			})
		}

		registerInput := &domain.RegisterInput{
			Email:    input.Email,
			Password: input.Password,
			Name:     input.Name,
		}

		result, err := mockSvc.Register(c.Context(), registerInput)
		if err != nil {
			if apperrors.IsValidation(err) {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error":   "Bad Request",
					"message": err.Error(),
				})
			}
			logger.Error("registration failed", zap.Error(err))
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "Internal Server Error",
				"message": "Registration failed",
			})
		}

		return c.Status(fiber.StatusCreated).JSON(fiber.Map{
			"accessToken":  result.AccessToken,
			"refreshToken": result.RefreshToken,
			"expiresAt":    result.ExpiresAt,
			"user": fiber.Map{
				"id":    result.User.ID,
				"email": result.User.Email,
				"name":  result.User.Name,
			},
		})
	})

	// GetCurrentUser
	app.Get("/auth/me", func(c *fiber.Ctx) error {
		userID, ok := middleware.GetUserID(c)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "Unauthorized",
				"message": "User not authenticated",
			})
		}

		user, err := mockSvc.GetUserByID(c.Context(), userID)
		if err != nil {
			if apperrors.IsNotFound(err) {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"error":   "Not Found",
					"message": "User not found",
				})
			}
			logger.Error("failed to get user", zap.Error(err))
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "Internal Server Error",
				"message": "Failed to get user",
			})
		}

		return c.JSON(fiber.Map{
			"id":        user.ID,
			"email":     user.Email,
			"name":      user.Name,
			"image":     user.Image,
			"createdAt": user.CreatedAt,
		})
	})

	// RefreshToken
	app.Post("/auth/refresh", func(c *fiber.Ctx) error {
		var input struct {
			RefreshToken string `json:"refreshToken"`
		}

		if err := c.BodyParser(&input); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Bad Request",
				"message": "Invalid request body: " + err.Error(),
			})
		}

		if input.RefreshToken == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Bad Request",
				"message": "Refresh token is required",
			})
		}

		result, err := mockSvc.RefreshToken(c.Context(), input.RefreshToken)
		if err != nil {
			if apperrors.IsUnauthorized(err) {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"error":   "Unauthorized",
					"message": "Invalid refresh token",
				})
			}
			logger.Error("failed to refresh token", zap.Error(err))
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "Internal Server Error",
				"message": "Failed to refresh token",
			})
		}

		return c.JSON(fiber.Map{
			"accessToken":  result.AccessToken,
			"refreshToken": result.RefreshToken,
			"expiresAt":    result.ExpiresAt,
		})
	})

	// Logout
	app.Post("/auth/logout", func(c *fiber.Ctx) error {
		var input struct {
			RefreshToken string `json:"refreshToken"`
		}

		if err := c.BodyParser(&input); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Bad Request",
				"message": "Invalid request body: " + err.Error(),
			})
		}

		if input.RefreshToken == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Bad Request",
				"message": "Refresh token is required",
			})
		}

		if err := mockSvc.Logout(c.Context(), input.RefreshToken); err != nil {
			logger.Error("logout failed", zap.Error(err))
		}

		return c.JSON(fiber.Map{
			"message": "Logged out successfully",
		})
	})

	return app
}

// --- Login Tests ---

func TestAuthHandler_Login(t *testing.T) {
	t.Parallel()
	t.Run("successfully logs in with valid credentials", func(t *testing.T) {
		t.Parallel()
		mockSvc := new(MockAuthService)
		app := setupAuthTestApp(mockSvc, nil)

		userID := uuid.New()
		expectedResult := &domain.AuthResult{
			AccessToken:  "access-token-123",
			RefreshToken: "refresh-token-456",
			ExpiresAt:    time.Now().Add(time.Hour),
			User: &domain.User{
				ID:    userID,
				Email: "test@example.com",
				Name:  "Test User",
			},
		}

		mockSvc.On("Login", mock.Anything, &domain.LoginInput{
			Email:    "test@example.com",
			Password: "password123",
		}).Return(expectedResult, nil)

		body, _ := json.Marshal(map[string]string{
			"email":    "test@example.com",
			"password": "password123",
		})
		req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)

		assert.Equal(t, "access-token-123", result["accessToken"])
		assert.Equal(t, "refresh-token-456", result["refreshToken"])

		mockSvc.AssertExpectations(t)
	})

	t.Run("returns 400 for missing email", func(t *testing.T) {
		t.Parallel()
		mockSvc := new(MockAuthService)
		app := setupAuthTestApp(mockSvc, nil)

		body, _ := json.Marshal(map[string]string{
			"password": "password123",
		})
		req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		assert.Equal(t, "Email and password are required", result["message"])
	})

	t.Run("returns 400 for missing password", func(t *testing.T) {
		t.Parallel()
		mockSvc := new(MockAuthService)
		app := setupAuthTestApp(mockSvc, nil)

		body, _ := json.Marshal(map[string]string{
			"email": "test@example.com",
		})
		req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("returns 401 for invalid credentials", func(t *testing.T) {
		t.Parallel()
		mockSvc := new(MockAuthService)
		app := setupAuthTestApp(mockSvc, nil)

		mockSvc.On("Login", mock.Anything, mock.Anything).
			Return(nil, apperrors.Unauthorized("invalid credentials"))

		body, _ := json.Marshal(map[string]string{
			"email":    "test@example.com",
			"password": "wrongpassword",
		})
		req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

		mockSvc.AssertExpectations(t)
	})

	t.Run("returns 500 for service error", func(t *testing.T) {
		t.Parallel()
		mockSvc := new(MockAuthService)
		app := setupAuthTestApp(mockSvc, nil)

		mockSvc.On("Login", mock.Anything, mock.Anything).
			Return(nil, assert.AnError)

		body, _ := json.Marshal(map[string]string{
			"email":    "test@example.com",
			"password": "password123",
		})
		req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

		mockSvc.AssertExpectations(t)
	})
}

// --- Register Tests ---

func TestAuthHandler_Register(t *testing.T) {
	t.Parallel()
	t.Run("successfully registers new user", func(t *testing.T) {
		t.Parallel()
		mockSvc := new(MockAuthService)
		app := setupAuthTestApp(mockSvc, nil)

		userID := uuid.New()
		expectedResult := &domain.AuthResult{
			AccessToken:  "access-token-123",
			RefreshToken: "refresh-token-456",
			ExpiresAt:    time.Now().Add(time.Hour),
			User: &domain.User{
				ID:    userID,
				Email: "new@example.com",
				Name:  "New User",
			},
		}

		mockSvc.On("Register", mock.Anything, &domain.RegisterInput{
			Email:    "new@example.com",
			Password: "password123",
			Name:     "New User",
		}).Return(expectedResult, nil)

		body, _ := json.Marshal(map[string]string{
			"email":    "new@example.com",
			"password": "password123",
			"name":     "New User",
		})
		req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		assert.Equal(t, "access-token-123", result["accessToken"])

		mockSvc.AssertExpectations(t)
	})

	t.Run("returns 400 for missing email", func(t *testing.T) {
		mockSvc := new(MockAuthService)
		app := setupAuthTestApp(mockSvc, nil)

		body, _ := json.Marshal(map[string]string{
			"password": "password123",
			"name":     "New User",
		})
		req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("returns 400 for password too short", func(t *testing.T) {
		mockSvc := new(MockAuthService)
		app := setupAuthTestApp(mockSvc, nil)

		body, _ := json.Marshal(map[string]string{
			"email":    "new@example.com",
			"password": "short",
			"name":     "New User",
		})
		req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		assert.Equal(t, "Password must be at least 8 characters", result["message"])
	})

	t.Run("returns 400 for duplicate email", func(t *testing.T) {
		mockSvc := new(MockAuthService)
		app := setupAuthTestApp(mockSvc, nil)

		mockSvc.On("Register", mock.Anything, mock.Anything).
			Return(nil, apperrors.Validation("email already registered"))

		body, _ := json.Marshal(map[string]string{
			"email":    "existing@example.com",
			"password": "password123",
			"name":     "New User",
		})
		req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

		mockSvc.AssertExpectations(t)
	})
}

// --- GetCurrentUser Tests ---

func TestAuthHandler_GetCurrentUser(t *testing.T) {
	t.Parallel()
	t.Run("successfully gets current user", func(t *testing.T) {
		t.Parallel()
		mockSvc := new(MockAuthService)
		userID := uuid.New()
		app := setupAuthTestApp(mockSvc, &userID)

		expectedUser := &domain.User{
			ID:        userID,
			Email:     "test@example.com",
			Name:      "Test User",
			CreatedAt: time.Now(),
		}

		mockSvc.On("GetUserByID", mock.Anything, userID).Return(expectedUser, nil)

		req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		assert.Equal(t, "test@example.com", result["email"])
		assert.Equal(t, "Test User", result["name"])

		mockSvc.AssertExpectations(t)
	})

	t.Run("returns 401 for unauthenticated request", func(t *testing.T) {
		mockSvc := new(MockAuthService)
		app := setupAuthTestApp(mockSvc, nil) // No user ID

		req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("returns 404 for non-existent user", func(t *testing.T) {
		mockSvc := new(MockAuthService)
		userID := uuid.New()
		app := setupAuthTestApp(mockSvc, &userID)

		mockSvc.On("GetUserByID", mock.Anything, userID).
			Return(nil, apperrors.NotFound("user not found"))

		req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)

		mockSvc.AssertExpectations(t)
	})
}

// --- RefreshToken Tests ---

func TestAuthHandler_RefreshToken(t *testing.T) {
	t.Parallel()
	t.Run("successfully refreshes token", func(t *testing.T) {
		t.Parallel()
		mockSvc := new(MockAuthService)
		app := setupAuthTestApp(mockSvc, nil)

		expectedResult := &domain.AuthResult{
			AccessToken:  "new-access-token",
			RefreshToken: "new-refresh-token",
			ExpiresAt:    time.Now().Add(time.Hour),
		}

		mockSvc.On("RefreshToken", mock.Anything, "old-refresh-token").
			Return(expectedResult, nil)

		body, _ := json.Marshal(map[string]string{
			"refreshToken": "old-refresh-token",
		})
		req := httptest.NewRequest(http.MethodPost, "/auth/refresh", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		assert.Equal(t, "new-access-token", result["accessToken"])
		assert.Equal(t, "new-refresh-token", result["refreshToken"])

		mockSvc.AssertExpectations(t)
	})

	t.Run("returns 400 for missing refresh token", func(t *testing.T) {
		mockSvc := new(MockAuthService)
		app := setupAuthTestApp(mockSvc, nil)

		body, _ := json.Marshal(map[string]string{})
		req := httptest.NewRequest(http.MethodPost, "/auth/refresh", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("returns 401 for invalid refresh token", func(t *testing.T) {
		mockSvc := new(MockAuthService)
		app := setupAuthTestApp(mockSvc, nil)

		mockSvc.On("RefreshToken", mock.Anything, "invalid-token").
			Return(nil, apperrors.Unauthorized("invalid token"))

		body, _ := json.Marshal(map[string]string{
			"refreshToken": "invalid-token",
		})
		req := httptest.NewRequest(http.MethodPost, "/auth/refresh", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

		mockSvc.AssertExpectations(t)
	})
}

// --- Logout Tests ---

func TestAuthHandler_Logout(t *testing.T) {
	t.Parallel()
	t.Run("successfully logs out", func(t *testing.T) {
		t.Parallel()
		mockSvc := new(MockAuthService)
		app := setupAuthTestApp(mockSvc, nil)

		mockSvc.On("Logout", mock.Anything, "refresh-token").Return(nil)

		body, _ := json.Marshal(map[string]string{
			"refreshToken": "refresh-token",
		})
		req := httptest.NewRequest(http.MethodPost, "/auth/logout", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		assert.Equal(t, "Logged out successfully", result["message"])

		mockSvc.AssertExpectations(t)
	})

	t.Run("returns 400 for missing refresh token", func(t *testing.T) {
		mockSvc := new(MockAuthService)
		app := setupAuthTestApp(mockSvc, nil)

		body, _ := json.Marshal(map[string]string{})
		req := httptest.NewRequest(http.MethodPost, "/auth/logout", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("succeeds even if logout fails internally", func(t *testing.T) {
		mockSvc := new(MockAuthService)
		app := setupAuthTestApp(mockSvc, nil)

		mockSvc.On("Logout", mock.Anything, "refresh-token").Return(assert.AnError)

		body, _ := json.Marshal(map[string]string{
			"refreshToken": "refresh-token",
		})
		req := httptest.NewRequest(http.MethodPost, "/auth/logout", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode) // Still succeeds

		mockSvc.AssertExpectations(t)
	})
}
