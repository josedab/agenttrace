package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApplyPostgresURL(t *testing.T) {
	cfg := PostgresConfig{}
	err := applyPostgresURL(
		&cfg,
		"postgres://agenttrace:p%40ss@postgres:5433/observability?sslmode=require",
	)
	require.NoError(t, err)

	assert.Equal(t, "postgres", cfg.Host)
	assert.Equal(t, 5433, cfg.Port)
	assert.Equal(t, "agenttrace", cfg.User)
	assert.Equal(t, "p@ss", cfg.Password)
	assert.Equal(t, "observability", cfg.Database)
	assert.Equal(t, "require", cfg.SSLMode)
}

func TestApplyClickHouseURL(t *testing.T) {
	cfg := ClickHouseConfig{}
	err := applyClickHouseURL(
		&cfg,
		"clickhouse://clickhouse:9001?username=agenttrace&password=secret&database=traces",
	)
	require.NoError(t, err)

	assert.Equal(t, "clickhouse", cfg.Host)
	assert.Equal(t, 9001, cfg.Port)
	assert.Equal(t, "agenttrace", cfg.User)
	assert.Equal(t, "secret", cfg.Password)
	assert.Equal(t, "traces", cfg.Database)
}

func TestApplyRedisURL(t *testing.T) {
	cfg := RedisConfig{}
	err := applyRedisURL(&cfg, "redis://:secret@redis:6380/2")
	require.NoError(t, err)

	assert.True(t, cfg.Enabled)
	assert.Equal(t, "redis", cfg.Host)
	assert.Equal(t, 6380, cfg.Port)
	assert.Equal(t, "secret", cfg.Password)
	assert.Equal(t, 2, cfg.DB)
}

func TestLoadDeploymentAliases(t *testing.T) {
	t.Setenv("SERVER_HOST", "")
	t.Setenv("SERVER_PORT", "")
	t.Setenv("SERVER_ENV", "")
	t.Setenv("API_HOST", "127.0.0.1")
	t.Setenv("API_PORT", "9090")
	t.Setenv("ENVIRONMENT", "production")
	t.Setenv("DATABASE_URL", "postgres://agenttrace:secure@postgres:5432/agenttrace?sslmode=require")
	t.Setenv("CLICKHOUSE_DSN", "clickhouse://agenttrace:secure@clickhouse:9000/agenttrace")
	t.Setenv("REDIS_URL", "redis://:secure@redis:6379/1")
	t.Setenv("MINIO_ENABLED", "false")
	t.Setenv("JWT_SECRET", "production-secret")
	t.Setenv("CORS_ALLOWED_ORIGINS", "https://app.example.com, https://admin.example.com")
	t.Setenv("CORS_ALLOW_CREDENTIALS", "true")
	t.Setenv("SERVER_SECURE_COOKIES", "true")

	cfg, err := Load()
	require.NoError(t, err)

	assert.Equal(t, "127.0.0.1", cfg.Server.Host)
	assert.Equal(t, 9090, cfg.Server.Port)
	assert.Equal(t, "production", cfg.Server.Env)
	assert.True(t, cfg.Server.SecureCookies)
	assert.Equal(t, "postgres", cfg.Postgres.Host)
	assert.Equal(t, "clickhouse", cfg.ClickHouse.Host)
	assert.Equal(t, "redis", cfg.Redis.Host)
	assert.Equal(t, 1, cfg.Redis.DB)
	assert.Equal(t, 1440, cfg.JWT.AccessExpiry)
	assert.Equal(t, "agenttrace", cfg.JWT.Issuer)
	assert.Equal(t, []string{"https://app.example.com", "https://admin.example.com"}, cfg.CORS.AllowedOrigins)
	assert.True(t, cfg.CORS.AllowCredentials)
}
