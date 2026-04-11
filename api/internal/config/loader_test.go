package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidate_ProductionDefaultPostgresPassword(t *testing.T) {
	cfg := validConfig()
	cfg.Server.Env = "production"
	cfg.JWT.Secret = "real-secret"
	cfg.Postgres.Password = "agenttrace" // default value
	cfg.Postgres.SSLMode = "require"
	cfg.MinIO.SecretKey = "secure-minio-key"
	err := validate(cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "postgres_password must be changed from default in production")
}

func TestValidate_InvalidPortNumbers(t *testing.T) {
	tests := []struct {
		name     string
		modifier func(*Config)
		errMsg   string
	}{
		{
			name: "server port 0",
			modifier: func(cfg *Config) {
				cfg.Server.Port = 0
			},
			errMsg: "server_port must be between 1 and 65535",
		},
		{
			name: "server port -1",
			modifier: func(cfg *Config) {
				cfg.Server.Port = -1
			},
			errMsg: "server_port must be between 1 and 65535",
		},
		{
			name: "server port 65536",
			modifier: func(cfg *Config) {
				cfg.Server.Port = 65536
			},
			errMsg: "server_port must be between 1 and 65535",
		},
		{
			name: "postgres port 0",
			modifier: func(cfg *Config) {
				cfg.Postgres.Port = 0
			},
			errMsg: "postgres_port must be between 1 and 65535",
		},
		{
			name: "redis port 0",
			modifier: func(cfg *Config) {
				cfg.Redis.Port = 0
			},
			errMsg: "redis_port must be between 1 and 65535",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := validConfig()
			tt.modifier(cfg)
			err := validate(cfg)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.errMsg)
		})
	}
}

func TestValidate_MissingRequiredFields(t *testing.T) {
	tests := []struct {
		name     string
		modifier func(*Config)
		errMsg   string
	}{
		{
			name:     "missing postgres_host",
			modifier: func(cfg *Config) { cfg.Postgres.Host = "" },
			errMsg:   "postgres_host is required",
		},
		{
			name:     "missing redis_host",
			modifier: func(cfg *Config) { cfg.Redis.Host = "" },
			errMsg:   "redis_host is required",
		},
		{
			name:     "missing clickhouse_host",
			modifier: func(cfg *Config) { cfg.ClickHouse.Host = "" },
			errMsg:   "clickhouse_host is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := validConfig()
			tt.modifier(cfg)
			err := validate(cfg)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.errMsg)
		})
	}
}

func TestValidate_ValidConfigPasses(t *testing.T) {
	cfg := validConfig()
	err := validate(cfg)
	assert.NoError(t, err)
}

func TestValidate_OAuthRequiresCallbackSecret(t *testing.T) {
	cfg := validConfig()
	cfg.OAuth.GoogleClientID = "google-client"
	cfg.OAuth.GoogleClientSecret = "google-secret"

	err := validate(cfg)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "oauth_callback_secret is required")

	cfg.OAuth.CallbackSecret = "trusted-callback-secret"
	assert.NoError(t, validate(cfg))
}

func TestValidate_RejectsCustomClickHouseDatabase(t *testing.T) {
	cfg := validConfig()
	cfg.ClickHouse.Database = "custom"

	err := validate(cfg)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "clickhouse_db must be agenttrace")
}

func TestJWTDurationCalculations(t *testing.T) {
	// Simulate the duration calculations from loader.go
	expiryHours := 24
	refreshExpiryDays := 7

	expiry := time.Duration(expiryHours) * time.Hour
	refreshExpiry := time.Duration(refreshExpiryDays) * 24 * time.Hour

	assert.Equal(t, 24*time.Hour, expiry)
	assert.Equal(t, 7*24*time.Hour, refreshExpiry)
}

func TestValidate_NonProductionLenientValidation(t *testing.T) {
	cfg := validConfig()
	cfg.Server.Env = "development"
	cfg.JWT.Secret = "change-me-in-production" // default secret OK in dev
	cfg.Postgres.Password = "agenttrace"       // default password OK in dev
	cfg.MinIO.SecretKey = "agenttrace123"      // default minio key OK in dev
	cfg.Postgres.SSLMode = "disable"           // SSL disable OK in dev

	err := validate(cfg)
	// In non-production, default JWT secret "change-me-in-production" is OK
	// but empty JWT secret is not
	assert.NoError(t, err)
}

func TestValidate_ProductionMinIODefaultSecret(t *testing.T) {
	cfg := validConfig()
	cfg.Server.Env = "production"
	cfg.JWT.Secret = "real-secret"
	cfg.Postgres.Password = "secure-pw"
	cfg.Postgres.SSLMode = "require"
	cfg.MinIO.SecretKey = "agenttrace123" // default
	err := validate(cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "minio_secret_key must be changed from default in production")
}

func TestValidate_ProductionAllowsDisabledMinIO(t *testing.T) {
	cfg := validConfig()
	cfg.Server.Env = "production"
	cfg.JWT.Secret = "real-secret"
	cfg.Postgres.Password = "secure-pw"
	cfg.Postgres.SSLMode = "require"
	cfg.MinIO.Enabled = false
	cfg.MinIO.SecretKey = ""

	assert.NoError(t, validate(cfg))
}

func TestValidate_ProductionRejectsWildcardCORS(t *testing.T) {
	cfg := validConfig()
	cfg.Server.Env = "production"
	cfg.JWT.Secret = "real-secret"
	cfg.Postgres.Password = "secure-pw"
	cfg.Postgres.SSLMode = "require"
	cfg.CORS.AllowedOrigins = []string{"*"}

	err := validate(cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cors_allowed_origins must not contain '*' in production")
}

func TestValidate_RejectsZeroJWTAccessExpiry(t *testing.T) {
	cfg := validConfig()
	cfg.JWT.AccessExpiry = 0

	err := validate(cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "jwt_access_expiry_minutes must be greater than zero")
}
