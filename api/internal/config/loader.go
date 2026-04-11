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

	if err := bindEnvAliases(v); err != nil {
		return nil, err
	}

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
	cfg.Server.CSRFEnabled = v.GetBool("server_csrf_enabled")
	cfg.Server.SecureCookies = v.GetBool("server_secure_cookies")
	cfg.Server.StripeWebhookSecret = v.GetString("stripe_webhook_secret")
	cfg.Server.PublicURL = v.GetString("public_url")

	// Privacy and local/private mode
	cfg.Privacy.NoEgress = v.GetBool("privacy_no_egress")
	cfg.Privacy.RedactionEnabled = v.GetBool("privacy_redaction_enabled")

	// CORS
	cfg.CORS.AllowedOrigins = splitCSV(v.GetString("cors_allowed_origins"))
	cfg.CORS.AllowCredentials = v.GetBool("cors_allow_credentials")

	// PostgreSQL
	cfg.Postgres.Host = v.GetString("postgres_host")
	cfg.Postgres.Port = v.GetInt("postgres_port")
	cfg.Postgres.User = v.GetString("postgres_user")
	cfg.Postgres.Password = v.GetString("postgres_password")
	cfg.Postgres.Database = v.GetString("postgres_db")
	cfg.Postgres.SSLMode = v.GetString("postgres_ssl_mode")
	cfg.Postgres.AllowInsecure = v.GetBool("postgres_allow_insecure")
	cfg.Postgres.MaxConns = int32(v.GetInt("postgres_max_conns"))
	cfg.Postgres.MinConns = int32(v.GetInt("postgres_min_conns"))
	if databaseURL := v.GetString("database_url"); databaseURL != "" {
		if err := applyPostgresURL(&cfg.Postgres, databaseURL); err != nil {
			return nil, fmt.Errorf("invalid DATABASE_URL: %w", err)
		}
	}

	// ClickHouse
	cfg.ClickHouse.Host = v.GetString("clickhouse_host")
	cfg.ClickHouse.Port = v.GetInt("clickhouse_port")
	cfg.ClickHouse.HTTPPort = v.GetInt("clickhouse_http_port")
	cfg.ClickHouse.User = v.GetString("clickhouse_user")
	cfg.ClickHouse.Password = v.GetString("clickhouse_password")
	cfg.ClickHouse.Database = v.GetString("clickhouse_db")
	if clickHouseDSN := v.GetString("clickhouse_dsn"); clickHouseDSN != "" {
		if err := applyClickHouseURL(&cfg.ClickHouse, clickHouseDSN); err != nil {
			return nil, fmt.Errorf("invalid CLICKHOUSE_DSN: %w", err)
		}
	}

	// Redis
	cfg.Redis.Enabled = v.GetBool("redis_enabled")
	cfg.Redis.Host = v.GetString("redis_host")
	cfg.Redis.Port = v.GetInt("redis_port")
	cfg.Redis.Password = v.GetString("redis_password")
	cfg.Redis.DB = v.GetInt("redis_db")
	if redisURL := v.GetString("redis_url"); redisURL != "" {
		if err := applyRedisURL(&cfg.Redis, redisURL); err != nil {
			return nil, fmt.Errorf("invalid REDIS_URL: %w", err)
		}
	}

	// MinIO
	cfg.MinIO.Enabled = v.GetBool("minio_enabled")
	cfg.MinIO.Endpoint = v.GetString("minio_endpoint")
	cfg.MinIO.AccessKey = v.GetString("minio_access_key")
	cfg.MinIO.SecretKey = v.GetString("minio_secret_key")
	cfg.MinIO.UseSSL = v.GetBool("minio_use_ssl")
	cfg.MinIO.Bucket = v.GetString("minio_bucket")

	// JWT
	cfg.JWT.Secret = v.GetString("jwt_secret")
	cfg.JWT.ExpiryHours = v.GetInt("jwt_expiry_hours")
	cfg.JWT.RefreshExpiryDays = v.GetInt("jwt_refresh_expiry_days")
	cfg.JWT.AccessExpiry = v.GetInt("jwt_access_expiry_minutes")
	cfg.JWT.Issuer = v.GetString("jwt_issuer")
	cfg.JWT.Expiry = time.Duration(cfg.JWT.ExpiryHours) * time.Hour
	cfg.JWT.RefreshExpiry = time.Duration(cfg.JWT.RefreshExpiryDays) * 24 * time.Hour

	// OAuth
	cfg.OAuth.GoogleClientID = v.GetString("google_client_id")
	cfg.OAuth.GoogleClientSecret = v.GetString("google_client_secret")
	cfg.OAuth.GitHubClientID = v.GetString("github_client_id")
	cfg.OAuth.GitHubClientSecret = v.GetString("github_client_secret")
	cfg.OAuth.CallbackSecret = v.GetString("oauth_callback_secret")

	// Optional GitHub workflow reporting
	cfg.GitHub.ReportingEnabled = v.GetBool("github_reporting_enabled")
	cfg.GitHub.APIURL = v.GetString("github_api_url")
	cfg.GitHub.ReportToken = v.GetString("github_report_token")

	// Rate Limiting
	cfg.RateLimit.Enabled = v.GetBool("rate_limit_enabled")
	cfg.RateLimit.RequestsPerSecond = v.GetInt("rate_limit_requests_per_second")
	cfg.RateLimit.Burst = v.GetInt("rate_limit_burst")
	cfg.RateLimit.UserMaxPerMinute = v.GetInt("rate_limit_user_max_per_minute")

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
	cfg.OTel.ReceiverHTTPPort = v.GetInt("otel_receiver_http_port")
	cfg.OTel.ReceiverHTTPPath = v.GetString("otel_receiver_http_path")
	cfg.OTel.ExporterEnabled = v.GetBool("otel_exporter_enabled")
	cfg.OTel.DefaultBatchSize = v.GetInt("otel_default_batch_size")
	cfg.OTel.DefaultMaxQueueSize = v.GetInt("otel_default_max_queue_size")
	cfg.OTel.DefaultBatchTimeoutMs = v.GetInt("otel_default_batch_timeout_ms")
	cfg.OTel.DefaultExportTimeoutMs = v.GetInt("otel_default_export_timeout_ms")
	cfg.OTel.DefaultRetryEnabled = v.GetBool("otel_default_retry_enabled")
	cfg.OTel.DefaultRetryInitialInterval = v.GetInt("otel_default_retry_initial_interval_ms")
	cfg.OTel.DefaultRetryMaxInterval = v.GetInt("otel_default_retry_max_interval_ms")
	cfg.OTel.DefaultRetryMaxElapsedTime = v.GetInt("otel_default_retry_max_elapsed_time_ms")
	cfg.OTel.DefaultRetryMultiplier = v.GetFloat64("otel_default_retry_multiplier")
	cfg.OTel.WorkerQueue = v.GetString("otel_worker_queue")
	cfg.OTel.WorkerConcurrency = v.GetInt("otel_worker_concurrency")
	cfg.OTel.ServiceName = v.GetString("otel_service_name")
	cfg.OTel.ServiceVersion = v.GetString("otel_service_version")

	// Sentry
	cfg.Sentry.Enabled = v.GetBool("sentry_enabled")
	cfg.Sentry.DSN = v.GetString("sentry_dsn")
	cfg.Sentry.Environment = v.GetString("sentry_environment")
	cfg.Sentry.Release = v.GetString("sentry_release")
	cfg.Sentry.Debug = v.GetBool("sentry_debug")
	cfg.Sentry.SampleRate = v.GetFloat64("sentry_sample_rate")
	cfg.Sentry.TracesSampleRate = v.GetFloat64("sentry_traces_sample_rate")

	// Validate required fields
	if err := validate(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func bindEnvAliases(v *viper.Viper) error {
	aliases := map[string][]string{
		"server_host":          {"SERVER_HOST", "API_HOST"},
		"server_port":          {"SERVER_PORT", "API_PORT"},
		"server_env":           {"SERVER_ENV", "ENVIRONMENT"},
		"clickhouse_dsn":       {"CLICKHOUSE_DSN", "CLICKHOUSE_URL"},
		"minio_access_key":     {"MINIO_ACCESS_KEY", "MINIO_ROOT_USER"},
		"minio_secret_key":     {"MINIO_SECRET_KEY", "MINIO_ROOT_PASSWORD"},
		"eval_api_key":         {"EVAL_API_KEY", "OPENAI_API_KEY"},
		"github_client_id":     {"GITHUB_CLIENT_ID", "GITHUB_ID"},
		"github_client_secret": {"GITHUB_CLIENT_SECRET", "GITHUB_SECRET"},
		"github_report_token":  {"GITHUB_REPORT_TOKEN"},
	}
	for key, names := range aliases {
		envNames := append([]string{key}, names...)
		if err := v.BindEnv(envNames...); err != nil {
			return fmt.Errorf("bind environment aliases for %s: %w", key, err)
		}
	}
	return nil
}

func splitCSV(value string) []string {
	var values []string
	for _, item := range strings.Split(value, ",") {
		if item = strings.TrimSpace(item); item != "" {
			values = append(values, item)
		}
	}
	return values
}

func setDefaults(v *viper.Viper) {
	// Server defaults
	v.SetDefault("server_host", "0.0.0.0")
	v.SetDefault("server_port", 8080)
	v.SetDefault("server_env", "development")
	v.SetDefault("server_csrf_enabled", true)
	v.SetDefault("server_secure_cookies", false)
	v.SetDefault("stripe_webhook_secret", "")
	v.SetDefault("public_url", "")
	v.SetDefault("cors_allowed_origins", "*")
	v.SetDefault("cors_allow_credentials", false)
	v.SetDefault("privacy_no_egress", false)
	v.SetDefault("privacy_redaction_enabled", true)
	v.SetDefault("github_reporting_enabled", false)
	v.SetDefault("github_api_url", "https://api.github.com")
	v.SetDefault("github_report_token", "")

	// PostgreSQL defaults
	v.SetDefault("postgres_host", "localhost")
	v.SetDefault("postgres_port", 5432)
	v.SetDefault("postgres_user", "agenttrace")
	v.SetDefault("postgres_password", "agenttrace")
	v.SetDefault("postgres_db", "agenttrace")
	v.SetDefault("postgres_ssl_mode", "disable")
	v.SetDefault("postgres_allow_insecure", false)
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
	v.SetDefault("redis_enabled", true)
	v.SetDefault("redis_host", "localhost")
	v.SetDefault("redis_port", 6379)
	v.SetDefault("redis_password", "")
	v.SetDefault("redis_db", 0)

	// MinIO defaults
	v.SetDefault("minio_enabled", true)
	v.SetDefault("minio_endpoint", "localhost:9002")
	v.SetDefault("minio_access_key", "agenttrace")
	v.SetDefault("minio_secret_key", "agenttrace123")
	v.SetDefault("minio_use_ssl", false)
	v.SetDefault("minio_bucket", "agenttrace-exports")

	// JWT defaults
	v.SetDefault("jwt_secret", "change-me-in-production")
	v.SetDefault("jwt_expiry_hours", 24)
	v.SetDefault("jwt_refresh_expiry_days", 7)
	v.SetDefault("jwt_access_expiry_minutes", 1440)
	v.SetDefault("jwt_issuer", "agenttrace")
	v.SetDefault("oauth_callback_secret", "")

	// Rate limiting defaults
	v.SetDefault("rate_limit_enabled", true)
	v.SetDefault("rate_limit_requests_per_second", 100)
	v.SetDefault("rate_limit_burst", 200)
	v.SetDefault("rate_limit_user_max_per_minute", 100)

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
	v.SetDefault("otel_receiver_http_port", 4318)
	v.SetDefault("otel_receiver_http_path", "/v1/traces")
	v.SetDefault("otel_exporter_enabled", false)
	v.SetDefault("otel_default_batch_size", 100)
	v.SetDefault("otel_default_max_queue_size", 1000)
	v.SetDefault("otel_default_batch_timeout_ms", 5000)
	v.SetDefault("otel_default_export_timeout_ms", 30000)
	v.SetDefault("otel_default_retry_enabled", true)
	v.SetDefault("otel_default_retry_initial_interval_ms", 1000)
	v.SetDefault("otel_default_retry_max_interval_ms", 30000)
	v.SetDefault("otel_default_retry_max_elapsed_time_ms", 300000)
	v.SetDefault("otel_default_retry_multiplier", 2.0)
	v.SetDefault("otel_worker_queue", "default")
	v.SetDefault("otel_worker_concurrency", 5)
	v.SetDefault("otel_service_name", "agenttrace")
	v.SetDefault("otel_service_version", "0.1.0")

	// Sentry defaults
	v.SetDefault("sentry_enabled", false)
	v.SetDefault("sentry_dsn", "")
	v.SetDefault("sentry_environment", "")
	v.SetDefault("sentry_release", "")
	v.SetDefault("sentry_debug", false)
	v.SetDefault("sentry_sample_rate", 1.0)
	v.SetDefault("sentry_traces_sample_rate", 0.1)
}

func validate(cfg *Config) error {
	var errs []string

	// JWT secret must not use default in production
	if (cfg.JWT.Secret == "change-me-in-production" || isPlaceholder(cfg.JWT.Secret)) && cfg.IsProduction() {
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
	if cfg.ClickHouse.Database != "" && cfg.ClickHouse.Database != "agenttrace" {
		errs = append(errs, "clickhouse_db must be agenttrace because migrations use the canonical database")
	}

	// Redis validation
	if cfg.Redis.Enabled {
		if cfg.Redis.Host == "" {
			errs = append(errs, "redis_host is required")
		}
		if cfg.Redis.Port <= 0 || cfg.Redis.Port > 65535 {
			errs = append(errs, "redis_port must be between 1 and 65535")
		}
	}

	// Server validation
	if cfg.Server.Port <= 0 || cfg.Server.Port > 65535 {
		errs = append(errs, "server_port must be between 1 and 65535")
	}
	if cfg.JWT.AccessExpiry <= 0 {
		errs = append(errs, "jwt_access_expiry_minutes must be greater than zero")
	}
	if cfg.JWT.Issuer == "" {
		errs = append(errs, "jwt_issuer is required")
	}
	googleConfigured := cfg.OAuth.GoogleClientID != "" || cfg.OAuth.GoogleClientSecret != ""
	if googleConfigured && (cfg.OAuth.GoogleClientID == "" || cfg.OAuth.GoogleClientSecret == "") {
		errs = append(errs, "google_client_id and google_client_secret must be configured together")
	}
	githubConfigured := cfg.OAuth.GitHubClientID != "" || cfg.OAuth.GitHubClientSecret != ""
	if githubConfigured && (cfg.OAuth.GitHubClientID == "" || cfg.OAuth.GitHubClientSecret == "") {
		errs = append(errs, "github_client_id and github_client_secret must be configured together")
	}
	if (googleConfigured || githubConfigured) &&
		(cfg.OAuth.CallbackSecret == "" || isPlaceholder(cfg.OAuth.CallbackSecret)) {
		errs = append(errs, "oauth_callback_secret is required when OAuth providers are configured")
	}

	if cfg.Privacy.NoEgress {
		if !cfg.Privacy.RedactionEnabled {
			errs = append(errs, "privacy_redaction_enabled must be true when privacy_no_egress is enabled")
		}
		if cfg.GitHub.ReportingEnabled {
			errs = append(errs, "github_reporting_enabled conflicts with privacy_no_egress")
		}
		if cfg.OTel.ExporterEnabled {
			errs = append(errs, "otel_exporter_enabled conflicts with privacy_no_egress")
		}
		if cfg.Sentry.Enabled {
			errs = append(errs, "sentry_enabled conflicts with privacy_no_egress")
		}
		if cfg.Eval.APIKey != "" {
			errs = append(errs, "external evaluation providers conflict with privacy_no_egress")
		}
		if googleConfigured || githubConfigured {
			errs = append(errs, "OAuth providers conflict with privacy_no_egress")
		}
	}

	// Production-specific checks
	if cfg.IsProduction() {
		if cfg.Postgres.Password == "" {
			errs = append(errs, "postgres_password is required in production")
		} else if cfg.Postgres.Password == "agenttrace" || isPlaceholder(cfg.Postgres.Password) {
			errs = append(errs, "postgres_password must be changed from default in production")
		}
		if cfg.Postgres.SSLMode == "disable" && !cfg.Postgres.AllowInsecure {
			errs = append(errs, "postgres_ssl_mode must use TLS in production unless postgres_allow_insecure is true")
		}
		if cfg.MinIO.Enabled && (cfg.MinIO.SecretKey == "" || cfg.MinIO.SecretKey == "agenttrace123") {
			errs = append(errs, "minio_secret_key must be changed from default in production")
		}
		if cfg.ClickHouse.Password == "" || isPlaceholder(cfg.ClickHouse.Password) {
			errs = append(errs, "clickhouse_password is required and must not be a placeholder in production")
		}
		if cfg.Redis.Enabled && (cfg.Redis.Password == "" || isPlaceholder(cfg.Redis.Password)) {
			errs = append(errs, "redis_password is required and must not be a placeholder in production")
		}
		if cfg.MinIO.Enabled && isPlaceholder(cfg.MinIO.SecretKey) {
			errs = append(errs, "minio_secret_key must not be a placeholder in production")
		}
		if len(cfg.CORS.AllowedOrigins) == 0 {
			errs = append(errs, "cors_allowed_origins is required in production")
		}

		for _, origin := range cfg.CORS.AllowedOrigins {
			if origin == "*" {
				errs = append(errs, "cors_allowed_origins must not contain '*' in production")
				break
			}
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("configuration validation failed:\n  - %s", strings.Join(errs, "\n  - "))
	}
	return nil
}

func isPlaceholder(value string) bool {
	return strings.HasPrefix(strings.ToLower(strings.TrimSpace(value)), "change-me")
}
