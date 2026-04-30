package service

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	apperrors "github.com/agenttrace/agenttrace/api/internal/pkg/errors"
)

// requireNoEgressError asserts an outbound refusal is reported as Unprocessable
// so transports answer 422 instead of a generic failure.
func requireNoEgressError(t *testing.T, err error) {
	t.Helper()
	require.Error(t, err)
	appErr := apperrors.GetAppError(err)
	require.NotNil(t, appErr, "expected an application error, got %v", err)
	assert.Equal(t, apperrors.CodeUnprocessable, appErr.Code)
	assert.Equal(t, http.StatusUnprocessableEntity, appErr.StatusCode)
	assert.Contains(t, appErr.Message, "no-egress")
}

// countingEndpoint records every inbound request so tests can prove that
// no-egress mode performs zero outbound calls.
func countingEndpoint(t *testing.T) (*httptest.Server, *int64) {
	t.Helper()
	var calls int64
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		atomic.AddInt64(&calls, 1)
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(server.Close)
	return server, &calls
}

func TestOTelExporterServiceBlocksEgressWithoutNetworkCalls(t *testing.T) {
	endpoint, calls := countingEndpoint(t)
	service := NewOTelExporterService(zap.NewNop(), NewEgressPolicy(true, true))
	t.Cleanup(service.Stop)

	_, err := service.CreateExporter(
		context.Background(),
		uuid.New(),
		uuid.New(),
		&domain.OTelExporterInput{
			Name:     "collector",
			Endpoint: endpoint.URL,
			Type:     domain.OTelExporterTypeHTTP,
		},
	)
	requireNoEgressError(t, err)

	exporter := &domain.OTelExporter{
		ID:           uuid.New(),
		ProjectID:    uuid.New(),
		Enabled:      true,
		Type:         domain.OTelExporterTypeHTTP,
		Endpoint:     endpoint.URL,
		SamplingRate: 1,
		Timeout:      5,
		BatchConfig:  domain.OTelBatchConfig{MaxBatchSize: 1, MaxQueueSize: 1},
	}
	requireNoEgressError(t, service.TestExporter(context.Background(), exporter))
	requireNoEgressError(t, service.QueueSpansForExport(exporter, []*domain.OTelSpan{{
		TraceID: "00000000000000000000000000000001",
		SpanID:  "0000000000000001",
		Name:    "blocked",
	}}))

	assert.Equal(t, int64(0), atomic.LoadInt64(calls))
}

func TestOTelExporterServiceAllowsExportWhenEgressPermitted(t *testing.T) {
	service := NewOTelExporterService(zap.NewNop(), AllowAllOutbound())
	t.Cleanup(service.Stop)

	exporter, err := service.CreateExporter(
		context.Background(),
		uuid.New(),
		uuid.New(),
		&domain.OTelExporterInput{
			Name:     "collector",
			Endpoint: "localhost:4317",
			Type:     domain.OTelExporterTypeGRPC,
		},
	)
	require.NoError(t, err)
	assert.Equal(t, "collector", exporter.Name)
}

func TestOTelCompatServiceBlocksDestinationCreation(t *testing.T) {
	service := NewOTelCompatService(zap.NewNop(), NewEgressPolicy(true, true))

	_, err := service.CreateDestination(
		context.Background(),
		uuid.New(),
		domain.OTelExportDestination{
			Name:     "remote",
			Endpoint: "https://collector.example.com",
			Format:   domain.OTelExportFormatOTLPHTTP,
		},
	)
	requireNoEgressError(t, err)
}

func TestOTelBridgeServiceBlocksDestinationCreation(t *testing.T) {
	service := NewOTelBridgeService(zap.NewNop(), nil, NewEgressPolicy(true, true))
	projectID := uuid.New()
	_, err := service.GetConfig(context.Background(), projectID)
	require.NoError(t, err)

	exportEnabled := true
	_, err = service.UpdateConfig(context.Background(), projectID, &domain.OTelBridgeConfigInput{
		ExportEnabled: &exportEnabled,
	})
	requireNoEgressError(t, err)

	_, err = service.AddDestination(context.Background(), projectID, &domain.OTelDestinationInput{
		Name:     "remote",
		Type:     "otlp",
		Endpoint: "https://collector.example.com",
	})
	requireNoEgressError(t, err)

	destinations, err := service.ListDestinations(context.Background(), projectID)
	require.NoError(t, err)
	assert.Empty(t, destinations)
}

func TestOTelBridgeDestinationRemovalIsProjectScoped(t *testing.T) {
	service := NewOTelBridgeService(zap.NewNop(), nil, AllowAllOutbound())
	ownerProjectID := uuid.New()
	otherProjectID := uuid.New()
	ctx := context.Background()
	_, err := service.GetConfig(ctx, ownerProjectID)
	require.NoError(t, err)
	_, err = service.GetConfig(ctx, otherProjectID)
	require.NoError(t, err)

	destination, err := service.AddDestination(ctx, ownerProjectID, &domain.OTelDestinationInput{
		Name:     "collector",
		Type:     "otlp",
		Endpoint: "https://collector.example.com",
	})
	require.NoError(t, err)

	err = service.RemoveDestination(ctx, otherProjectID, destination.ID)
	require.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err))

	destinations, err := service.ListDestinations(ctx, ownerProjectID)
	require.NoError(t, err)
	assert.Len(t, destinations, 1)

	require.NoError(t, service.RemoveDestination(ctx, ownerProjectID, destination.ID))
	destinations, err = service.ListDestinations(ctx, ownerProjectID)
	require.NoError(t, err)
	assert.Empty(t, destinations)
}

func TestFederationServiceBlocksRemoteInstanceUse(t *testing.T) {
	endpoint, calls := countingEndpoint(t)
	service := NewFederationService(zap.NewNop(), NewEgressPolicy(true, true))
	projectID := uuid.New()
	ctx := context.Background()

	_, err := service.AddPeer(ctx, projectID, &domain.OTelFederationInput{
		Name: "remote",
		URL:  endpoint.URL,
	})
	requireNoEgressError(t, err)
	assert.Empty(t, service.ListPeers(ctx, projectID))

	_, err = service.FederatedQuery(ctx, projectID, &domain.FederationQuery{Query: "traces"})
	requireNoEgressError(t, err)

	_, err = service.CreateExportDestination(ctx, projectID, &domain.ExportDestinationInput{
		Name:     "remote",
		Type:     "datadog",
		Endpoint: endpoint.URL,
	})
	requireNoEgressError(t, err)
	assert.Empty(t, service.ListExportDestinations(ctx, projectID))
	assert.Equal(t, int64(0), atomic.LoadInt64(calls))
}

func TestFederationServiceScopesPeerRemovalToProject(t *testing.T) {
	service := NewFederationService(zap.NewNop(), AllowAllOutbound())
	ctx := context.Background()
	ownerProject := uuid.New()

	peer, err := service.AddPeer(ctx, ownerProject, &domain.OTelFederationInput{
		Name: "peer",
		URL:  "https://peer.example.com",
	})
	require.NoError(t, err)

	err = service.RemovePeer(ctx, uuid.New(), peer.ID)
	require.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err))
	assert.Len(t, service.ListPeers(ctx, ownerProject), 1)

	require.NoError(t, service.RemovePeer(ctx, ownerProject, peer.ID))
	assert.Empty(t, service.ListPeers(ctx, ownerProject))
}

func TestWarehouseSyncServiceBlocksConnectionsAndSyncs(t *testing.T) {
	blocked := NewWarehouseSyncService(zap.NewNop(), NewEgressPolicy(true, true))
	ctx := context.Background()
	projectID := uuid.New()

	_, err := blocked.CreateConnection(ctx, projectID, &domain.WarehouseConnectionInput{
		Name:   "warehouse",
		Type:   domain.WarehouseSnowflake,
		Config: domain.WarehouseConfig{Account: "acct", Database: "db"},
	})
	requireNoEgressError(t, err)

	connections, err := blocked.ListConnections(ctx, projectID)
	require.NoError(t, err)
	assert.Empty(t, connections)

	// A connection created before no-egress mode was enabled must not sync either.
	allowed := NewWarehouseSyncService(zap.NewNop(), AllowAllOutbound())
	connection, err := allowed.CreateConnection(ctx, projectID, &domain.WarehouseConnectionInput{
		Name:   "warehouse",
		Type:   domain.WarehouseSnowflake,
		Config: domain.WarehouseConfig{Account: "acct", Database: "db"},
	})
	require.NoError(t, err)

	blocked.mu.Lock()
	blocked.connections[connection.ID] = connection
	blocked.mu.Unlock()

	_, err = blocked.TriggerSync(ctx, projectID, connection.ID)
	requireNoEgressError(t, err)
	_, err = blocked.TestConnection(ctx, projectID, connection.ID)
	requireNoEgressError(t, err)

	operations, err := blocked.GetSyncStatus(ctx, projectID, connection.ID)
	require.NoError(t, err)
	assert.Empty(t, operations)
	assert.Equal(t, domain.SyncStatusIdle, connection.LastSyncStatus)
}

func TestWarehouseSyncServiceScopesConnectionTestToProject(t *testing.T) {
	service := NewWarehouseSyncService(zap.NewNop(), AllowAllOutbound())
	ctx := context.Background()
	projectID := uuid.New()

	connection, err := service.CreateConnection(ctx, projectID, &domain.WarehouseConnectionInput{
		Name:   "warehouse",
		Type:   domain.WarehouseSnowflake,
		Config: domain.WarehouseConfig{Account: "acct", Database: "db"},
	})
	require.NoError(t, err)

	_, err = service.TestConnection(ctx, uuid.New(), connection.ID)
	require.Error(t, err)

	result, err := service.TestConnection(ctx, projectID, connection.ID)
	require.NoError(t, err)
	assert.True(t, result.Reachable)
	assert.Equal(t, connection.ID, result.ConnectionID)
}

func TestEgressPolicyCapabilitiesCoverEveryGuardedSurface(t *testing.T) {
	capabilities := NewEgressPolicy(true, true).Capabilities()

	for _, capability := range outboundCapabilities {
		item, ok := capabilities.Capabilities[string(capability)]
		require.True(t, ok, "capability %s must be reported", capability)
		assert.False(t, item.Available)
		assert.NotEmpty(t, item.Reason)
	}

	allowed := AllowAllOutbound().Capabilities()
	assert.Equal(t, "standard", allowed.Mode)
	for _, capability := range outboundCapabilities {
		assert.True(t, allowed.Capabilities[string(capability)].Available)
	}
}

func TestMigrationServiceBlocksRemoteSourceImports(t *testing.T) {
	service := NewMigrationService(
		zap.NewNop(),
		nil,
		nil,
		nil,
		nil,
		NewEgressPolicy(true, true),
	)
	ctx := context.Background()

	_, err := service.StartMigration(ctx, uuid.New(), &domain.MigrationInput{
		Source: "langfuse",
		Config: domain.MigrationConfig{SourceDSN: "https://cloud.langfuse.com"},
	})
	requireNoEgressError(t, err)

	_, _, err = service.ValidateSource(ctx, "langfuse", "https://cloud.langfuse.com")
	requireNoEgressError(t, err)

	// The documented local JSON export path stays available in no-egress mode.
	valid, message, err := service.ValidateSource(ctx, "langfuse", LocalMigrationDSN)
	require.NoError(t, err)
	assert.True(t, valid)
	assert.Contains(t, message, "JSON export")
}

func TestLLMClientReportsUnconfiguredInNoEgressMode(t *testing.T) {
	blocked := NewLLMClient(zap.NewNop(), NewEgressPolicy(true, true))
	blocked.apiKey = "sk-test-key"

	// Callers branch on IsConfigured to choose the local fallback, so no-egress
	// mode must report the provider as unusable rather than dialing it.
	assert.False(t, blocked.IsConfigured())

	response, err := blocked.ChatCompletion(context.Background(), "system", "user")
	require.NoError(t, err)
	assert.NotEmpty(t, response)

	allowed := NewLLMClient(zap.NewNop(), AllowAllOutbound())
	allowed.apiKey = "sk-test-key"
	assert.True(t, allowed.IsConfigured())
}
