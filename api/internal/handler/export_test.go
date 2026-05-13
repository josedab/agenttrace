package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/middleware"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

func TestExportHandler_QueueUnavailable(t *testing.T) {
	tests := []struct {
		name    string
		handler func(*ExportHandler, *fiber.Ctx) error
	}{
		{
			name:    "trace export",
			handler: (*ExportHandler).ExportData,
		},
		{
			name:    "dataset export",
			handler: (*ExportHandler).ExportDataset,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()
			handler := NewExportHandler(nil, true, service.AllowAllOutbound(), zap.NewNop())

			app.Post("/export", func(c *fiber.Ctx) error {
				c.Locals(string(middleware.ContextKeyProjectID), uuid.New())
				return tt.handler(handler, c)
			})

			req := httptest.NewRequestWithContext(
				t.Context(),
				http.MethodPost,
				"/export",
				strings.NewReader("{}"),
			)
			req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
			resp, err := app.Test(req, -1)
			require.NoError(t, err)
			assert.Equal(t, http.StatusServiceUnavailable, resp.StatusCode)

			var body map[string]string
			require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
			assert.Contains(t, body["message"], "full local stack")
			require.NoError(t, resp.Body.Close())
		})
	}
}

func TestExportHandler_StorageUnavailable(t *testing.T) {
	client := asynq.NewClient(asynq.RedisClientOpt{Addr: "localhost:1"})
	t.Cleanup(func() {
		require.NoError(t, client.Close())
	})

	app := fiber.New()
	handler := NewExportHandler(client, false, service.AllowAllOutbound(), zap.NewNop())
	app.Post("/export", func(c *fiber.Ctx) error {
		c.Locals(string(middleware.ContextKeyProjectID), uuid.New())
		return handler.ExportData(c)
	})

	req := httptest.NewRequestWithContext(
		t.Context(),
		http.MethodPost,
		"/export",
		strings.NewReader("{}"),
	)
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	assert.Equal(t, http.StatusServiceUnavailable, resp.StatusCode)

	var body map[string]string
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	assert.Contains(t, body["message"], "MinIO")
	require.NoError(t, resp.Body.Close())
}
