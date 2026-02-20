package config

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func validConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Host: "0.0.0.0",
			Port: 8080,
			Env:  "development",
		},
		Postgres: PostgresConfig{
			Host:     "localhost",
			Port:     5432,
			User:     "agenttrace",
			Password: "secret",
			Database: "agenttrace",
			SSLMode:  "disable",
		},
		ClickHouse: ClickHouseConfig{
			Host: "localhost",
			Port: 9000,
		},
		Redis: RedisConfig{
			Host: "localhost",
			Port: 6379,
		},
		JWT: JWTConfig{
			Secret: "test-secret-key",
		},
	}
}

func TestValidate_ValidConfig(t *testing.T) {
	cfg := validConfig()
	err := validate(cfg)
	assert.NoError(t, err)
}

func TestValidate_EmptyJWTSecret(t *testing.T) {
	cfg := validConfig()
	cfg.JWT.Secret = ""
	err := validate(cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "jwt_secret is required")
}

func TestValidate_DefaultJWTSecretInProduction(t *testing.T) {
	cfg := validConfig()
	cfg.JWT.Secret = "change-me-in-production"
	cfg.Server.Env = "production"
	cfg.Postgres.Password = "secure-pw"
	cfg.Postgres.SSLMode = "require"
	err := validate(cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "jwt_secret must be changed from default in production")
}

func TestValidate_EmptyPostgresHost(t *testing.T) {
	cfg := validConfig()
	cfg.Postgres.Host = ""
	err := validate(cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "postgres_host is required")
}

func TestValidate_InvalidPostgresPort(t *testing.T) {
	cfg := validConfig()
	cfg.Postgres.Port = 0
	err := validate(cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "postgres_port must be between 1 and 65535")
}

func TestValidate_EmptyPostgresUser(t *testing.T) {
	cfg := validConfig()
	cfg.Postgres.User = ""
	err := validate(cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "postgres_user is required")
}

func TestValidate_EmptyPostgresDB(t *testing.T) {
	cfg := validConfig()
	cfg.Postgres.Database = ""
	err := validate(cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "postgres_db is required")
}

func TestValidate_EmptyClickHouseHost(t *testing.T) {
	cfg := validConfig()
	cfg.ClickHouse.Host = ""
	err := validate(cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "clickhouse_host is required")
}

func TestValidate_InvalidClickHousePort(t *testing.T) {
	cfg := validConfig()
	cfg.ClickHouse.Port = 70000
	err := validate(cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "clickhouse_port must be between 1 and 65535")
}

func TestValidate_EmptyRedisHost(t *testing.T) {
	cfg := validConfig()
	cfg.Redis.Host = ""
	err := validate(cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "redis_host is required")
}

func TestValidate_InvalidServerPort(t *testing.T) {
	cfg := validConfig()
	cfg.Server.Port = -1
	err := validate(cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "server_port must be between 1 and 65535")
}

func TestValidate_ProductionRequiresPostgresPassword(t *testing.T) {
	cfg := validConfig()
	cfg.Server.Env = "production"
	cfg.JWT.Secret = "real-secret"
	cfg.Postgres.Password = ""
	cfg.Postgres.SSLMode = "require"
	err := validate(cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "postgres_password is required in production")
}

func TestValidate_ProductionWarnsDisabledSSL(t *testing.T) {
	cfg := validConfig()
	cfg.Server.Env = "production"
	cfg.JWT.Secret = "real-secret"
	cfg.Postgres.Password = "secure-pw"
	cfg.Postgres.SSLMode = "disable"
	err := validate(cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "postgres_ssl_mode should not be 'disable' in production")
}

func TestValidate_MultipleErrors(t *testing.T) {
	cfg := validConfig()
	cfg.JWT.Secret = ""
	cfg.Postgres.Host = ""
	cfg.Redis.Host = ""
	err := validate(cfg)
	require.Error(t, err)

	errMsg := err.Error()
	assert.True(t, strings.Contains(errMsg, "jwt_secret is required"))
	assert.True(t, strings.Contains(errMsg, "postgres_host is required"))
	assert.True(t, strings.Contains(errMsg, "redis_host is required"))
}

func TestPostgresDSN(t *testing.T) {
	cfg := PostgresConfig{
		Host:     "db.example.com",
		Port:     5432,
		User:     "admin",
		Password: "secret",
		Database: "mydb",
		SSLMode:  "require",
	}
	expected := "postgres://admin:secret@db.example.com:5432/mydb?sslmode=require"
	assert.Equal(t, expected, cfg.DSN())
}

func TestRedisAddr(t *testing.T) {
	cfg := RedisConfig{
		Host: "redis.example.com",
		Port: 6380,
	}
	assert.Equal(t, "redis.example.com:6380", cfg.Addr())
}

func TestIsDevelopment(t *testing.T) {
	cfg := Config{Server: ServerConfig{Env: "development"}}
	assert.True(t, cfg.IsDevelopment())
	assert.False(t, cfg.IsProduction())
}

func TestIsProduction(t *testing.T) {
	cfg := Config{Server: ServerConfig{Env: "production"}}
	assert.True(t, cfg.IsProduction())
	assert.False(t, cfg.IsDevelopment())
}
