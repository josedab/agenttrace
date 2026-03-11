package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Load loads configuration from environment variables and config files
func Load() (*Config, error) {
	v := viper.New()

	// Set defaults
	setDefaults(v)

	// Read from environment variables
	v.SetEnvPrefix("")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Optionally read from config file
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("./config")
	v.AddConfigPath("/etc/agenttrace")

	// Ignore error if config file not found
	_ = v.ReadInConfig()

	var cfg Config

	// Server
	cfg.Server.Host = v.GetString("server_host")
	cfg.Server.Port = v.GetInt("server_port")
	cfg.Server.Env = v.GetString("server_env")

	// PostgreSQL
	cfg.Postgres.Host = v.GetString("postgres_host")
	cfg.Postgres.Port = v.GetInt("postgres_port")
	cfg.Postgres.User = v.GetString("postgres_user")
	cfg.Postgres.Password = v.GetString("postgres_password")
	cfg.Postgres.Database = v.GetString("postgres_db")
	cfg.Postgres.SSLMode = v.GetString("postgres_ssl_mode")
	cfg.Postgres.MaxConns = int32(v.GetInt("postgres_max_conns"))
	cfg.Postgres.MinConns = int32(v.GetInt("postgres_min_conns"))

	// ClickHouse
	cfg.ClickHouse.Host = v.GetString("clickhouse_host")
	cfg.ClickHouse.Port = v.GetInt("clickhouse_port")
	cfg.ClickHouse.HTTPPort = v.GetInt("clickhouse_http_port")
	cfg.ClickHouse.User = v.GetString("clickhouse_user")
	cfg.ClickHouse.Password = v.GetString("clickhouse_password")
	cfg.ClickHouse.Database = v.GetString("clickhouse_db")

	// Redis
	cfg.Redis.Host = v.GetString("redis_host")
	cfg.Redis.Port = v.GetInt("redis_port")
	cfg.Redis.Password = v.GetString("redis_password")
	cfg.Redis.DB = v.GetInt("redis_db")

	// MinIO
	cfg.MinIO.Endpoint = v.GetString("minio_endpoint")
	cfg.MinIO.AccessKey = v.GetString("minio_access_key")
	cfg.MinIO.SecretKey = v.GetString("minio_secret_key")
	cfg.MinIO.UseSSL = v.GetBool("minio_use_ssl")
	cfg.MinIO.Bucket = v.GetString("minio_bucket")

	// JWT
	cfg.JWT.Secret = v.GetString("jwt_secret")
	cfg.JWT.ExpiryHours = v.GetInt("jwt_expiry_hours")
	cfg.JWT.RefreshExpiryDays = v.GetInt("jwt_refresh_expiry_days")
	cfg.JWT.Expiry = time.Duration(cfg.JWT.ExpiryHours) * time.Hour
	cfg.JWT.RefreshExpiry = time.Duration(cfg.JWT.RefreshExpiryDays) * 24 * time.Hour

	// OAuth
	cfg.OAuth.GoogleClientID = v.GetString("google_client_id")
	cfg.OAuth.GoogleClientSecret = v.GetString("google_client_secret")
	cfg.OAuth.GitHubClientID = v.GetString("github_client_id")
	cfg.OAuth.GitHubClientSecret = v.GetString("github_client_secret")

	// Rate Limiting
	cfg.RateLimit.Enabled = v.GetBool("rate_limit_enabled")
	cfg.RateLimit.RequestsPerSecond = v.GetInt("rate_limit_requests_per_second")
	cfg.RateLimit.Burst = v.GetInt("rate_limit_burst")

	// Worker
	cfg.Worker.Concurrency = v.GetInt("worker_concurrency")
	cfg.Worker.QueueCritical = v.GetString("worker_queue_critical")
	cfg.Worker.QueueDefault = v.GetString("worker_queue_default")
	cfg.Worker.QueueLow = v.GetString("worker_queue_low")
	cfg.Worker.CostEnabled = v.GetBool("cost_worker_enabled")
	cfg.Worker.CostBatchSize = v.GetInt("cost_worker_batch_size")

	// Logging
	cfg.Log.Level = v.GetString("log_level")
	cfg.Log.Format = v.GetString("log_format")

	// Evaluation
	cfg.Eval.Enabled = v.GetBool("eval_worker_enabled")
	cfg.Eval.DefaultModel = v.GetString("eval_default_model")
	cfg.Eval.APIKey = v.GetString("eval_api_key")

	// Retention
	cfg.Retention.Days = v.GetInt("retention_days")
	cfg.Retention.Enabled = v.GetBool("retention_worker_enabled")

	// OpenTelemetry
	cfg.OTel.ReceiverEnabled = v.GetBool("otel_receiver_enabled")
	cfg.OTel.ReceiverGRPCPort = v.GetInt("otel_receiver_grpc_port")

	// Validate required fields
	if err := validate(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func setDefaults(v *viper.Viper) {
	// Server defaults
	v.SetDefault("server_host", "0.0.0.0")
	v.SetDefault("server_port", 8080)
	v.SetDefault("server_env", "development")

	// PostgreSQL defaults
	v.SetDefault("postgres_host", "localhost")
	v.SetDefault("postgres_port", 5432)
	v.SetDefault("postgres_user", "agenttrace")
	v.SetDefault("postgres_password", "agenttrace")
	v.SetDefault("postgres_db", "agenttrace")
	v.SetDefault("postgres_ssl_mode", "disable")
	v.SetDefault("postgres_max_conns", 25)
	v.SetDefault("postgres_min_conns", 5)

	// ClickHouse defaults
	v.SetDefault("clickhouse_host", "localhost")
	v.SetDefault("clickhouse_port", 9000)
	v.SetDefault("clickhouse_http_port", 8123)
	v.SetDefault("clickhouse_user", "agenttrace")
	v.SetDefault("clickhouse_password", "agenttrace")
	v.SetDefault("clickhouse_db", "agenttrace")

	// Redis defaults
	v.SetDefault("redis_host", "localhost")
	v.SetDefault("redis_port", 6379)
	v.SetDefault("redis_password", "")
	v.SetDefault("redis_db", 0)

	// MinIO defaults
	v.SetDefault("minio_endpoint", "localhost:9002")
	v.SetDefault("minio_access_key", "agenttrace")
	v.SetDefault("minio_secret_key", "agenttrace123")
	v.SetDefault("minio_use_ssl", false)
	v.SetDefault("minio_bucket", "agenttrace-exports")

	// JWT defaults
	v.SetDefault("jwt_secret", "change-me-in-production")
	v.SetDefault("jwt_expiry_hours", 24)
	v.SetDefault("jwt_refresh_expiry_days", 7)

	// Rate limiting defaults
	v.SetDefault("rate_limit_enabled", true)
	v.SetDefault("rate_limit_requests_per_second", 100)
	v.SetDefault("rate_limit_burst", 200)

	// Worker defaults
	v.SetDefault("worker_concurrency", 10)
	v.SetDefault("worker_queue_critical", "critical")
	v.SetDefault("worker_queue_default", "default")
	v.SetDefault("worker_queue_low", "low")
	v.SetDefault("cost_worker_enabled", true)
	v.SetDefault("cost_worker_batch_size", 100)

	// Logging defaults
	v.SetDefault("log_level", "info")
	v.SetDefault("log_format", "json")

	// Evaluation defaults
	v.SetDefault("eval_worker_enabled", true)
	v.SetDefault("eval_default_model", "gpt-4o-mini")

	// Retention defaults
	v.SetDefault("retention_days", 90)
	v.SetDefault("retention_worker_enabled", true)

	// OpenTelemetry defaults
	v.SetDefault("otel_receiver_enabled", false)
	v.SetDefault("otel_receiver_grpc_port", 4317)
}

func validate(cfg *Config) error {
	var errs []string

	// JWT secret must not use default in production
	if cfg.JWT.Secret == "change-me-in-production" && cfg.IsProduction() {
		errs = append(errs, "jwt_secret must be changed from default in production")
	}

	// JWT secret must not be empty
	if cfg.JWT.Secret == "" {
		errs = append(errs, "jwt_secret is required")
	}

	// PostgreSQL validation
	if cfg.Postgres.Host == "" {
		errs = append(errs, "postgres_host is required")
	}
	if cfg.Postgres.Port <= 0 || cfg.Postgres.Port > 65535 {
		errs = append(errs, "postgres_port must be between 1 and 65535")
	}
	if cfg.Postgres.User == "" {
		errs = append(errs, "postgres_user is required")
	}
	if cfg.Postgres.Database == "" {
		errs = append(errs, "postgres_db is required")
	}

	// ClickHouse validation
	if cfg.ClickHouse.Host == "" {
		errs = append(errs, "clickhouse_host is required")
	}
	if cfg.ClickHouse.Port <= 0 || cfg.ClickHouse.Port > 65535 {
		errs = append(errs, "clickhouse_port must be between 1 and 65535")
	}

	// Redis validation
	if cfg.Redis.Host == "" {
		errs = append(errs, "redis_host is required")
	}
	if cfg.Redis.Port <= 0 || cfg.Redis.Port > 65535 {
		errs = append(errs, "redis_port must be between 1 and 65535")
	}

	// Server validation
	if cfg.Server.Port <= 0 || cfg.Server.Port > 65535 {
		errs = append(errs, "server_port must be between 1 and 65535")
	}

	// Production-specific checks
	if cfg.IsProduction() {
		if cfg.Postgres.Password == "" || cfg.Postgres.Password == "agenttrace" {
			errs = append(errs, "postgres_password must be changed from default in production")
		}
		if cfg.Postgres.SSLMode == "disable" {
			cfg.Postgres.SSLMode = "require"
		}
		if cfg.MinIO.SecretKey == "" || cfg.MinIO.SecretKey == "agenttrace123" {
			errs = append(errs, "minio_secret_key must be changed from default in production")
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("configuration validation failed:\n  - %s", strings.Join(errs, "\n  - "))
	}
	return nil
}
