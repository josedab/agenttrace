package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestMigrationValidateSourceDoesNotClaimUnverifiedLangfuseDSN(t *testing.T) {
	service := NewMigrationService(zap.NewNop(), nil, nil, nil, nil)

	valid, message, err := service.ValidateSource(
		context.Background(),
		"langfuse",
		"postgres://user:secret@example.com/langfuse",
	)

	require.NoError(t, err)
	assert.False(t, valid)
	assert.Contains(t, message, "--source-file")
	assert.NotContains(t, message, "secret")
}

func TestMigrationValidateSourceAcceptsDocumentedJSONMode(t *testing.T) {
	service := NewMigrationService(zap.NewNop(), nil, nil, nil, nil)

	valid, _, err := service.ValidateSource(
		context.Background(),
		"langfuse",
		"json-export",
	)

	require.NoError(t, err)
	assert.True(t, valid)
}
