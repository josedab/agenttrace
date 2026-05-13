package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/middleware"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// noEgressOutboundProbe counts requests that reach a remote endpoint so route
// tests can prove no-egress mode performs zero outbound calls.
func noEgressOutboundProbe(t *testing.T) (*httptest.Server, *int64) {
	t.Helper()
	var calls int64
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		atomic.AddInt64(&calls, 1)
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(server.Close)
	return server, &calls
}

func postWithProject(
	t *testing.T,
	app *fiber.App,
	path, body string,
) (int, map[string]any) {
	return requestWithProject(t, app, http.MethodPost, path, body)
}

func requestWithProject(
	t *testing.T,
	app *fiber.App,
	method, path, body string,
) (int, map[string]any) {
	t.Helper()
	request := httptest.NewRequestWithContext(
		t.Context(),
		method,
		path,
		strings.NewReader(body),
	)
	request.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	response, err := app.Test(request, -1)
	require.NoError(t, err)
	defer func() { require.NoError(t, response.Body.Close()) }()

	decoded := map[string]any{}
	if err := json.NewDecoder(response.Body).Decode(&decoded); err != nil {
		t.Fatalf("decode %s response: %v", path, err)
	}
	return response.StatusCode, decoded
}

func projectScopedApp(handlers func(router fiber.Router)) *fiber.App {
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals(string(middleware.ContextKeyProjectID), uuid.New())
		return c.Next()
	})
	handlers(app)
	return app
}

func assertNoEgressResponse(t *testing.T, status int, body map[string]any) {
	t.Helper()
	assert.Equal(t, http.StatusUnprocessableEntity, status)
	message := stringField(body, "message")
	if message == "" {
		message = stringField(body, "error")
	}
	assert.Contains(t, message, "no-egress")
}

// stringField reads a string field from a decoded JSON envelope.
func stringField(body map[string]any, key string) string {
	value, ok := body[key].(string)
	if !ok {
		return ""
	}
	return value
}

func TestFederationRoutesRejectNoEgress(t *testing.T) {
	endpoint, calls := noEgressOutboundProbe(t)
	policy := service.NewEgressPolicy(true, true)
	handler := NewFederationHandler(
		service.NewFederationService(zap.NewNop(), policy),
		zap.NewNop(),
	)
	app := projectScopedApp(func(router fiber.Router) {
		router.Post("/federation/peers", handler.AddPeer)
		router.Post("/federation/query", handler.FederatedQuery)
		router.Post("/federation/destinations", handler.CreateExportDestination)
	})

	status, body := postWithProject(t, app, "/federation/peers",
		`{"name":"remote","url":"`+endpoint.URL+`"}`)
	assertNoEgressResponse(t, status, body)

	status, body = postWithProject(t, app, "/federation/query", `{"query":"traces"}`)
	assertNoEgressResponse(t, status, body)

	status, body = postWithProject(t, app, "/federation/destinations",
		`{"name":"remote","type":"datadog","endpoint":"`+endpoint.URL+`"}`)
	assertNoEgressResponse(t, status, body)

	assert.Equal(t, int64(0), atomic.LoadInt64(calls))
}

func TestWarehouseRoutesRejectNoEgress(t *testing.T) {
	endpoint, calls := noEgressOutboundProbe(t)
	handler := NewWarehouseSyncHandler(
		service.NewWarehouseSyncService(zap.NewNop(), service.NewEgressPolicy(true, true)),
		zap.NewNop(),
	)
	app := projectScopedApp(func(router fiber.Router) {
		router.Post("/warehouse/connections", handler.CreateConnection)
		router.Post("/warehouse/connections/:connId/sync", handler.TriggerSync)
		router.Post("/warehouse/connections/:connId/test", handler.TestConnection)
	})

	status, body := postWithProject(t, app, "/warehouse/connections",
		`{"name":"analytics","type":"snowflake","config":{"account":"acct","database":"db"}}`)
	assertNoEgressResponse(t, status, body)

	connectionID := uuid.New().String()
	status, body = postWithProject(t, app, "/warehouse/connections/"+connectionID+"/sync", `{}`)
	assertNoEgressResponse(t, status, body)

	status, body = postWithProject(t, app, "/warehouse/connections/"+connectionID+"/test", `{}`)
	assertNoEgressResponse(t, status, body)

	_ = endpoint
	assert.Equal(t, int64(0), atomic.LoadInt64(calls))
}

func TestOTelDestinationRoutesRejectNoEgress(t *testing.T) {
	endpoint, calls := noEgressOutboundProbe(t)
	policy := service.NewEgressPolicy(true, true)
	compat := NewOTelCompatHandler(
		service.NewOTelCompatService(zap.NewNop(), policy),
		zap.NewNop(),
	)
	bridge := NewOTelBridgeHandler(
		service.NewOTelBridgeService(zap.NewNop(), nil, policy),
		zap.NewNop(),
	)
	otel := NewOTelHandler(
		zap.NewNop(),
		service.NewOTelExporterService(zap.NewNop(), policy),
	)
	app := projectScopedApp(func(router fiber.Router) {
		router.Post("/otel/destinations", compat.CreateExportDestination)
		router.Get("/otel-bridge/config", bridge.GetConfig)
		router.Put("/otel-bridge/config", bridge.UpdateConfig)
		router.Post("/otel-bridge/destinations", bridge.AddDestination)
		router.Post("/otel/exporters", otel.CreateExporter)
	})

	status, body := postWithProject(t, app, "/otel/destinations",
		`{"name":"remote","endpoint":"`+endpoint.URL+`","format":"otlp_http"}`)
	assertNoEgressResponse(t, status, body)

	configRequest := httptest.NewRequestWithContext(
		t.Context(),
		http.MethodGet,
		"/otel-bridge/config",
		nil,
	)
	configResponse, err := app.Test(configRequest, -1)
	require.NoError(t, err)
	require.NoError(t, configResponse.Body.Close())

	status, body = requestWithProject(
		t,
		app,
		http.MethodPut,
		"/otel-bridge/config",
		`{"exportEnabled":true}`,
	)
	assertNoEgressResponse(t, status, body)

	status, body = postWithProject(t, app, "/otel-bridge/destinations",
		`{"name":"remote","type":"otlp","endpoint":"`+endpoint.URL+`"}`)
	assertNoEgressResponse(t, status, body)

	status, body = postWithProject(t, app, "/otel/exporters",
		`{"name":"remote","type":"http","endpoint":"`+endpoint.URL+`"}`)
	assertNoEgressResponse(t, status, body)

	assert.Equal(t, int64(0), atomic.LoadInt64(calls))
}

func TestExportRouteRejectsRemoteDestinationInNoEgress(t *testing.T) {
	handler := NewExportHandler(nil, true, service.NewEgressPolicy(true, true), zap.NewNop())
	app := projectScopedApp(func(router fiber.Router) {
		router.Post("/export/data", handler.ExportData)
		router.Post("/export/dataset", handler.ExportDataset)
	})

	// A remote destination is refused before the queue availability check, so the
	// privacy decision is never masked by missing infrastructure.
	status, body := postWithProject(t, app, "/export/data",
		`{"type":"traces","format":"json","destination":{"type":"s3","bucket":"remote"}}`)
	assertNoEgressResponse(t, status, body)

	status, body = postWithProject(t, app, "/export/dataset",
		`{"datasetId":"`+uuid.New().String()+`","destination":{"type":"s3","bucket":"remote"}}`)
	assertNoEgressResponse(t, status, body)

	// Local exports stay available in no-egress mode; they stop only on
	// missing infrastructure.
	status, _ = postWithProject(t, app, "/export/data", `{"type":"traces","format":"json"}`)
	assert.Equal(t, http.StatusServiceUnavailable, status)
}
