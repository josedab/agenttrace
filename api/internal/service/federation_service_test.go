package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

func TestNewFederationService(t *testing.T) {
	svc := NewFederationService(zap.NewNop())
	assert.NotNil(t, svc)
}

func TestFederationService_PeerManagement(t *testing.T) {
	svc := NewFederationService(zap.NewNop())
	ctx := context.Background()
	projectID := uuid.New()

	t.Run("adds peer", func(t *testing.T) {
		input := &domain.OTelFederationInput{
			Name:   "staging-cluster",
			URL:    "https://staging.example.com",
			APIKey: "test-key",
		}

		peer, err := svc.AddPeer(ctx, projectID, input)
		require.NoError(t, err)
		assert.Equal(t, "staging-cluster", peer.Name)
		assert.Equal(t, "https://staging.example.com", peer.URL)
		assert.Equal(t, "connected", peer.Status)
		assert.NotNil(t, peer.LastSeen)
	})

	t.Run("lists peers for project", func(t *testing.T) {
		peers := svc.ListPeers(ctx, projectID)
		assert.Len(t, peers, 1)
	})

	t.Run("does not list peers from other projects", func(t *testing.T) {
		otherProject := uuid.New()
		peers := svc.ListPeers(ctx, otherProject)
		assert.Len(t, peers, 0)
	})

	t.Run("removes peer", func(t *testing.T) {
		peers := svc.ListPeers(ctx, projectID)
		require.Len(t, peers, 1)

		err := svc.RemovePeer(ctx, peers[0].ID)
		require.NoError(t, err)

		peers = svc.ListPeers(ctx, projectID)
		assert.Len(t, peers, 0)
	})
}

func TestFederationService_ExportDestinations(t *testing.T) {
	svc := NewFederationService(zap.NewNop())
	ctx := context.Background()
	projectID := uuid.New()

	t.Run("creates destination with defaults", func(t *testing.T) {
		input := &domain.ExportDestinationInput{
			Name:     "Datadog Export",
			Type:     "datadog",
			Endpoint: "https://trace.agent.datadoghq.com",
		}

		dest, err := svc.CreateExportDestination(ctx, projectID, input)
		require.NoError(t, err)
		assert.Equal(t, "datadog", dest.Type)
		assert.Equal(t, "grpc", dest.Protocol)
		assert.Equal(t, 1.0, dest.Sampling)
		assert.Equal(t, 100, dest.BatchSize)
		assert.True(t, dest.Enabled)
		assert.Equal(t, "active", dest.Status)
	})

	t.Run("creates destination with custom sampling", func(t *testing.T) {
		sampling := 0.5
		batchSize := 250
		input := &domain.ExportDestinationInput{
			Name:      "Honeycomb",
			Type:      "honeycomb",
			Endpoint:  "https://api.honeycomb.io",
			Protocol:  "http",
			Sampling:  &sampling,
			BatchSize: &batchSize,
		}

		dest, err := svc.CreateExportDestination(ctx, projectID, input)
		require.NoError(t, err)
		assert.Equal(t, "http", dest.Protocol)
		assert.InDelta(t, 0.5, dest.Sampling, 0.001)
		assert.Equal(t, 250, dest.BatchSize)
	})

	t.Run("lists destinations for project", func(t *testing.T) {
		dests := svc.ListExportDestinations(ctx, projectID)
		assert.Len(t, dests, 2)
	})
}

func TestFederationService_FederatedQuery(t *testing.T) {
	svc := NewFederationService(zap.NewNop())
	ctx := context.Background()
	projectID := uuid.New()

	// Add some peers
	svc.AddPeer(ctx, projectID, &domain.OTelFederationInput{Name: "peer-1", URL: "http://peer1.example.com"})
	svc.AddPeer(ctx, projectID, &domain.OTelFederationInput{Name: "peer-2", URL: "http://peer2.example.com"})

	t.Run("queries all peers", func(t *testing.T) {
		query := &domain.FederationQuery{Query: "test query"}
		result, err := svc.FederatedQuery(ctx, projectID, query)
		require.NoError(t, err)
		assert.Len(t, result.Results, 2)
		assert.Greater(t, result.QueryTime, int64(-1))
	})

	t.Run("queries specific peers", func(t *testing.T) {
		peers := svc.ListPeers(ctx, projectID)
		query := &domain.FederationQuery{
			Query:   "test",
			PeerIDs: []uuid.UUID{peers[0].ID},
		}
		result, err := svc.FederatedQuery(ctx, projectID, query)
		require.NoError(t, err)
		assert.Len(t, result.Results, 1)
	})
}
