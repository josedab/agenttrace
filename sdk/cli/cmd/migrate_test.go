package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMigrateCommand(t *testing.T) {
	t.Run("has expected flags", func(t *testing.T) {
		f := migrateCmd.PersistentFlags()
		assert.NotNil(t, f.Lookup("source"))
		assert.NotNil(t, f.Lookup("source-dsn"))
		assert.NotNil(t, f.Lookup("target-host"))
		assert.NotNil(t, f.Lookup("dry-run"))
		assert.NotNil(t, f.Lookup("incremental"))
	})

	t.Run("has validate subcommand", func(t *testing.T) {
		found := false
		for _, c := range migrateCmd.Commands() {
			if c.Name() == "validate" {
				found = true
				break
			}
		}
		assert.True(t, found, "migrate should have validate subcommand")
	})

	t.Run("source defaults to langfuse", func(t *testing.T) {
		f := migrateCmd.PersistentFlags().Lookup("source")
		assert.Equal(t, "langfuse", f.DefValue)
	})
}

func TestMaskDSN(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "postgres://user:password@host:5432/db",
			expected: "postgres://user:****@host:5432/db",
		},
		{
			input:    "postgres://admin:s3cr3t@localhost:5432/langfuse",
			expected: "postgres://admin:****@localhost:5432/langfuse",
		},
		{
			input:    "no-scheme-dsn",
			expected: "no-scheme-dsn",
		},
		{
			input:    "postgres://user@host/db",
			expected: "postgres://user@host/db",
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.expected, maskDSN(tt.input))
		})
	}
}

func TestGetMigrateHost(t *testing.T) {
	t.Run("uses flag when set", func(t *testing.T) {
		original := migrateHost
		defer func() { migrateHost = original }()

		migrateHost = "http://custom:8080"
		assert.Equal(t, "http://custom:8080", getMigrateHost())
	})

	t.Run("falls back to env var", func(t *testing.T) {
		original := migrateHost
		defer func() { migrateHost = original }()

		migrateHost = ""
		t.Setenv("AGENTTRACE_API_URL", "http://env-host:8080")
		assert.Equal(t, "http://env-host:8080", getMigrateHost())
	})
}

func TestRunMigrate_RequiresAPIKey(t *testing.T) {
	original := apiKey
	defer func() { apiKey = original }()

	apiKey = ""
	t.Setenv("AGENTTRACE_API_KEY", "")

	err := runMigrate(migrateCmd, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API key required")
}

func TestRunMigrate_RequiresSourceDSN(t *testing.T) {
	original := apiKey
	origDSN := migrateSourceDSN
	defer func() { apiKey = original; migrateSourceDSN = origDSN }()

	apiKey = "test-key"
	migrateSourceDSN = ""

	err := runMigrate(migrateCmd, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "--source-dsn is required")
}
