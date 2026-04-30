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

func TestEgressPolicyReportsAndEnforcesPrivateMode(t *testing.T) {
	policy := NewEgressPolicy(true, true)

	err := policy.Require(EgressWebhooks)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no-egress")

	capabilities := policy.Capabilities()
	assert.Equal(t, "local_private", capabilities.Mode)
	assert.True(t, capabilities.NoEgress)
	assert.True(t, capabilities.RedactionEnabled)
	assert.False(t, capabilities.Capabilities[string(EgressGitHub)].Available)
	assert.True(t, capabilities.Capabilities["localTraceStorage"].Available)
	assert.True(t, capabilities.Capabilities["redactedShareLinks"].Available)
}

func TestNotificationServiceBlocksNoEgressBeforeNetwork(t *testing.T) {
	service := NewNotificationService(
		zap.NewNop(),
		"",
		NewEgressPolicy(true, true),
	)

	_, err := service.SendNotification(
		context.Background(),
		&domain.Webhook{
			ID:        uuid.New(),
			Type:      domain.WebhookTypeGeneric,
			URL:       "https://8.8.8.8/webhook",
			IsEnabled: true,
		},
		domain.EventTypeTeamDigest,
		map[string]any{"summary": "private"},
	)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "no-egress")
}
