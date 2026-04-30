package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateWebhookURL(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	require.NoError(t, ValidateWebhookURL(ctx, "https://8.8.8.8/webhook"))
	assert.Error(t, ValidateWebhookURL(ctx, "http://8.8.8.8/webhook"))
	assert.Error(t, ValidateWebhookURL(ctx, "https://127.0.0.1/webhook"))
	assert.Error(t, ValidateWebhookURL(ctx, "https://169.254.169.254/latest/meta-data"))
	assert.Error(t, ValidateWebhookURL(ctx, "https://user:password@8.8.8.8/webhook"))
}
