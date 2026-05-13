package handler

import (
	"bytes"
	"context"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/config"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

func TestOAuthCallbackRequiresTrustedServerSecret(t *testing.T) {
	authService := service.NewAuthService(
		&config.Config{OAuth: config.OAuthConfig{CallbackSecret: "trusted-secret"}},
		nil,
		nil,
		nil,
		nil,
	)
	handler := NewAuthHandler(authService, zap.NewNop())
	app := fiber.New()
	app.Post("/api/auth/callback/google", handler.OAuthCallback)

	request := httptest.NewRequestWithContext(
		context.Background(),
		fiber.MethodPost,
		"/api/auth/callback/google",
		bytes.NewBufferString(`{}`),
	)
	request.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	response, err := app.Test(request)

	require.NoError(t, err)
	defer func() {
		require.NoError(t, response.Body.Close())
	}()
	assert.Equal(t, fiber.StatusUnauthorized, response.StatusCode)
}
