package config

import "time"

// Config holds all configuration for the application
type Config struct {
	Server     ServerConfig
	Postgres   PostgresConfig
	ClickHouse ClickHouseConfig
	Redis      RedisConfig
	MinIO      MinIOConfig
	JWT        JWTConfig
	OAuth      OAuthConfig
	RateLimit  RateLimitConfig
	Worker     WorkerConfig
	Log        LogConfig
	Eval       EvalConfig
	Retention  RetentionConfig
	OTel       OTelConfig
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
	Env  string `mapstructure:"env"`
}

// PostgresConfig holds PostgreSQL configuration
type PostgresConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Database string `mapstructure:"database"`
	SSLMode  string `mapstructure:"ssl_mode"`
	MaxConns int32  `mapstructure:"max_conns"`
	MinConns int32  `mapstructure:"min_conns"`
}

// DSN returns the PostgreSQL connection string
func (c PostgresConfig) DSN() string {
	return "postgres://" + c.User + ":" + c.Password + "@" + c.Host + ":" +
		string(rune(c.Port)) + "/" + c.Database + "?sslmode=" + c.SSLMode
}

// ClickHouseConfig holds ClickHouse configuration
type ClickHouseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	HTTPPort int    `mapstructure:"http_port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Database string `mapstructure:"database"`
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

// Addr returns the Redis address
func (c RedisConfig) Addr() string {
	return c.Host + ":" + string(rune(c.Port))
}

// MinIOConfig holds MinIO configuration
type MinIOConfig struct {
	Endpoint  string `mapstructure:"endpoint"`
	AccessKey string `mapstructure:"access_key"`
	SecretKey string `mapstructure:"secret_key"`
	UseSSL    bool   `mapstructure:"use_ssl"`
	Bucket    string `mapstructure:"bucket"`
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	Secret            string        `mapstructure:"secret"`
	ExpiryHours       int           `mapstructure:"expiry_hours"`
	RefreshExpiryDays int           `mapstructure:"refresh_expiry_days"`
	Expiry            time.Duration `mapstructure:"-"`
	RefreshExpiry     time.Duration `mapstructure:"-"`
	AccessExpiry      int           `mapstructure:"access_expiry"`  // Access token expiry in minutes
	Issuer            string        `mapstructure:"issuer"`         // JWT issuer
}

// OAuthConfig holds OAuth provider configuration
type OAuthConfig struct {
	GoogleClientID     string `mapstructure:"google_client_id"`
	GoogleClientSecret string `mapstructure:"google_client_secret"`
	GitHubClientID     string `mapstructure:"github_client_id"`
	GitHubClientSecret string `mapstructure:"github_client_secret"`
}

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	Enabled           bool `mapstructure:"enabled"`
	RequestsPerSecond int  `mapstructure:"requests_per_second"`
	Burst             int  `mapstructure:"burst"`
}

// WorkerConfig holds background worker configuration
type WorkerConfig struct {
	Concurrency    int    `mapstructure:"concurrency"`
	QueueCritical  string `mapstructure:"queue_critical"`
	QueueDefault   string `mapstructure:"queue_default"`
	QueueLow       string `mapstructure:"queue_low"`
	CostEnabled    bool   `mapstructure:"cost_enabled"`
	CostBatchSize  int    `mapstructure:"cost_batch_size"`
}

// LogConfig holds logging configuration
type LogConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

// EvalConfig holds evaluation configuration
type EvalConfig struct {
	Enabled      bool   `mapstructure:"enabled"`
	DefaultModel string `mapstructure:"default_model"`
	APIKey       string `mapstructure:"api_key"`
}

// RetentionConfig holds data retention configuration
type RetentionConfig struct {
	Days    int  `mapstructure:"days"`
	Enabled bool `mapstructure:"enabled"`
}

// OTelConfig holds OpenTelemetry configuration
type OTelConfig struct {
	// Receiver configuration - for ingesting OTLP data
	ReceiverEnabled  bool   `mapstructure:"receiver_enabled"`
	ReceiverGRPCPort int    `mapstructure:"receiver_grpc_port"`
	ReceiverHTTPPort int    `mapstructure:"receiver_http_port"`
	ReceiverHTTPPath string `mapstructure:"receiver_http_path"`

	// Exporter configuration - for sending traces to external backends
	ExporterEnabled bool `mapstructure:"exporter_enabled"`

	// Default batch settings for new exporters
	DefaultBatchSize      int `mapstructure:"default_batch_size"`
	DefaultMaxQueueSize   int `mapstructure:"default_max_queue_size"`
	DefaultBatchTimeoutMs int `mapstructure:"default_batch_timeout_ms"`
	DefaultExportTimeoutMs int `mapstructure:"default_export_timeout_ms"`

	// Default retry settings
	DefaultRetryEnabled         bool    `mapstructure:"default_retry_enabled"`
	DefaultRetryInitialInterval int     `mapstructure:"default_retry_initial_interval_ms"`
	DefaultRetryMaxInterval     int     `mapstructure:"default_retry_max_interval_ms"`
	DefaultRetryMaxElapsedTime  int     `mapstructure:"default_retry_max_elapsed_time_ms"`
	DefaultRetryMultiplier      float64 `mapstructure:"default_retry_multiplier"`

	// Worker queue for async exports
	WorkerQueue       string `mapstructure:"worker_queue"`
	WorkerConcurrency int    `mapstructure:"worker_concurrency"`

	// Service name for outbound traces
	ServiceName    string `mapstructure:"service_name"`
	ServiceVersion string `mapstructure:"service_version"`
}

// IsDevelopment returns true if running in development mode
func (c Config) IsDevelopment() bool {
	return c.Server.Env == "development"
}

// IsProduction returns true if running in production mode
func (c Config) IsProduction() bool {
	return c.Server.Env == "production"
}
