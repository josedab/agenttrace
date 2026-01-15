package handler

import (
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/dto"
	"github.com/agenttrace/agenttrace/api/internal/middleware"
	apperrors "github.com/agenttrace/agenttrace/api/internal/pkg/errors"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// AuthHandler handles authentication endpoints
type AuthHandler struct {
	authService *service.AuthService
	logger      *zap.Logger
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authService *service.AuthService, logger *zap.Logger) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		logger:      logger,
	}
}

// Login handles POST /auth/login
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req dto.LoginRequest
	if err := dto.ParseAndValidate(c, &req); err != nil {
		return err
	}

	loginInput := &domain.LoginInput{
		Email:    req.Email,
		Password: req.Password,
	}

	result, err := h.authService.Login(c.Context(), loginInput)
	if err != nil {
		if apperrors.IsUnauthorized(err) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "Unauthorized",
				"message": "Invalid email or password",
			})
		}
		h.logger.Error("login failed", zap.Error(err))
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
}

// Register handles POST /auth/register
func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req dto.RegisterRequest
	if err := dto.ParseAndValidate(c, &req); err != nil {
		return err
	}

	registerInput := &domain.RegisterInput{
		Email:    req.Email,
		Password: req.Password,
		Name:     req.Name,
	}

	result, err := h.authService.Register(c.Context(), registerInput)
	if err != nil {
		if apperrors.IsValidation(err) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Bad Request",
				"message": err.Error(),
			})
		}
		h.logger.Error("registration failed", zap.Error(err))
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
}

// GetCurrentUser handles GET /auth/me
func (h *AuthHandler) GetCurrentUser(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "User not authenticated",
		})
	}

	user, err := h.authService.GetUserByID(c.Context(), userID)
	if err != nil {
		if apperrors.IsNotFound(err) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   "Not Found",
				"message": "User not found",
			})
		}
		h.logger.Error("failed to get user", zap.Error(err))
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
}

// RefreshToken handles POST /auth/refresh
func (h *AuthHandler) RefreshToken(c *fiber.Ctx) error {
	var req dto.RefreshTokenRequest
	if err := dto.ParseAndValidate(c, &req); err != nil {
		return err
	}

	result, err := h.authService.RefreshToken(c.Context(), req.RefreshToken)
	if err != nil {
		if apperrors.IsUnauthorized(err) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "Unauthorized",
				"message": "Invalid refresh token",
			})
		}
		h.logger.Error("failed to refresh token", zap.Error(err))
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
}

// Logout handles POST /auth/logout
func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	var req dto.LogoutRequest
	if err := dto.ParseAndValidate(c, &req); err != nil {
		return err
	}

	if err := h.authService.Logout(c.Context(), req.RefreshToken); err != nil {
		h.logger.Error("logout failed", zap.Error(err))
		// Don't expose error details for logout
	}

	return c.JSON(fiber.Map{
		"message": "Logged out successfully",
	})
}

// OAuthCallback handles OAuth provider callbacks
func (h *AuthHandler) OAuthCallback(c *fiber.Ctx) error {
	provider := c.Params("provider")
	if provider == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Provider required",
		})
	}

	// Parse OAuth callback data from query params or body
	var callbackInput domain.OAuthCallbackInput
	callbackInput.Provider = provider

	// Get code from query params
	code := c.Query("code")
	if code == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Authorization code required",
		})
	}

	// In a real implementation, we would exchange the code for tokens
	// and get user info from the provider. For now, parse from body.
	if err := c.BodyParser(&callbackInput); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid callback data",
		})
	}

	result, err := h.authService.HandleOAuthCallback(c.Context(), &callbackInput)
	if err != nil {
		h.logger.Error("OAuth callback failed",
			zap.String("provider", provider),
			zap.Error(err),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "OAuth authentication failed",
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
			"image": result.User.Image,
		},
	})
}

// RegisterRoutes registers auth routes
func (h *AuthHandler) RegisterRoutes(app *fiber.App, authMiddleware *middleware.AuthMiddleware) {
	auth := app.Group("/auth")

	// Public routes
	auth.Post("/login", h.Login)
	auth.Post("/register", h.Register)
	auth.Post("/refresh", h.RefreshToken)
	auth.Post("/logout", h.Logout)
	auth.Get("/callback/:provider", h.OAuthCallback)

	// Protected routes
	protected := auth.Group("", authMiddleware.RequireJWT())
	protected.Get("/me", h.GetCurrentUser)
}
