package middleware

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HTTP metrics
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "agenttrace_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "agenttrace_http_request_duration_seconds",
			Help:    "HTTP request latency in seconds",
			Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		},
		[]string{"method", "path"},
	)

	httpRequestSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "agenttrace_http_request_size_bytes",
			Help:    "HTTP request size in bytes",
			Buckets: []float64{100, 1000, 10000, 100000, 1000000, 10000000},
		},
		[]string{"method", "path"},
	)

	httpResponseSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "agenttrace_http_response_size_bytes",
			Help:    "HTTP response size in bytes",
			Buckets: []float64{100, 1000, 10000, 100000, 1000000, 10000000},
		},
		[]string{"method", "path"},
	)

	httpActiveRequests = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "agenttrace_http_active_requests",
			Help: "Number of active HTTP requests",
		},
		[]string{"method"},
	)

	// Ingestion metrics
	tracesIngested = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "agenttrace_traces_ingested_total",
			Help: "Total number of traces ingested",
		},
		[]string{"project_id"},
	)

	observationsIngested = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "agenttrace_observations_ingested_total",
			Help: "Total number of observations ingested",
		},
		[]string{"project_id", "type"},
	)

	ingestionLatency = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "agenttrace_ingestion_latency_seconds",
			Help:    "Trace ingestion latency in seconds",
			Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1},
		},
		[]string{"project_id"},
	)

	// Cost metrics
	totalCostUSD = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "agenttrace_total_cost_usd",
			Help: "Total cost in USD",
		},
		[]string{"project_id", "model"},
	)

	totalTokens = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "agenttrace_total_tokens",
			Help: "Total tokens processed",
		},
		[]string{"project_id", "model", "type"},
	)

	// Database metrics
	dbQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "agenttrace_db_query_duration_seconds",
			Help:    "Database query duration in seconds",
			Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5},
		},
		[]string{"database", "operation"},
	)

	// Evaluation metrics
	evaluationsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "agenttrace_evaluations_total",
			Help: "Total number of evaluations executed",
		},
		[]string{"project_id", "evaluator_id", "status"},
	)

	evaluationDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "agenttrace_evaluation_duration_seconds",
			Help:    "Evaluation execution duration in seconds",
			Buckets: []float64{.1, .25, .5, 1, 2.5, 5, 10, 30},
		},
		[]string{"project_id", "evaluator_id"},
	)
)

// MetricsConfig configures the metrics middleware
type MetricsConfig struct {
	// Skip function
	Skip func(*fiber.Ctx) bool
	// PathNormalizer normalizes paths for metrics labels
	PathNormalizer func(string) string
}

// DefaultMetricsConfig returns default metrics config
func DefaultMetricsConfig() MetricsConfig {
	return MetricsConfig{
		Skip:           HealthSkipper,
		PathNormalizer: DefaultPathNormalizer,
	}
}

// DefaultPathNormalizer normalizes paths by replacing IDs with placeholders
func DefaultPathNormalizer(path string) string {
	// This is a simple normalizer - in production you might want something more sophisticated
	return path
}

// MetricsMiddleware creates a Prometheus metrics middleware
type MetricsMiddleware struct {
	config MetricsConfig
}

// NewMetricsMiddleware creates a new metrics middleware
func NewMetricsMiddleware(config MetricsConfig) *MetricsMiddleware {
	return &MetricsMiddleware{
		config: config,
	}
}

// Handler returns the metrics handler
func (m *MetricsMiddleware) Handler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Skip if configured
		if m.config.Skip != nil && m.config.Skip(c) {
			return c.Next()
		}

		start := time.Now()
		method := c.Method()
		path := m.config.PathNormalizer(c.Path())

		// Track active requests
		httpActiveRequests.WithLabelValues(method).Inc()
		defer httpActiveRequests.WithLabelValues(method).Dec()

		// Process request
		err := c.Next()

		// Record metrics
		duration := time.Since(start).Seconds()
		status := strconv.Itoa(c.Response().StatusCode())
		requestSize := float64(len(c.Request().Body()))
		responseSize := float64(len(c.Response().Body()))

		httpRequestsTotal.WithLabelValues(method, path, status).Inc()
		httpRequestDuration.WithLabelValues(method, path).Observe(duration)
		httpRequestSize.WithLabelValues(method, path).Observe(requestSize)
		httpResponseSize.WithLabelValues(method, path).Observe(responseSize)

		return err
	}
}

// RecordTraceIngested records a trace ingestion
func RecordTraceIngested(projectID string) {
	tracesIngested.WithLabelValues(projectID).Inc()
}

// RecordObservationIngested records an observation ingestion
func RecordObservationIngested(projectID, obsType string) {
	observationsIngested.WithLabelValues(projectID, obsType).Inc()
}

// RecordIngestionLatency records ingestion latency
func RecordIngestionLatency(projectID string, duration time.Duration) {
	ingestionLatency.WithLabelValues(projectID).Observe(duration.Seconds())
}

// RecordCost records cost metrics
func RecordCost(projectID, model string, costUSD float64) {
	totalCostUSD.WithLabelValues(projectID, model).Add(costUSD)
}

// RecordTokens records token usage
func RecordTokens(projectID, model, tokenType string, count int) {
	totalTokens.WithLabelValues(projectID, model, tokenType).Add(float64(count))
}

// RecordDBQuery records database query metrics
func RecordDBQuery(database, operation string, duration time.Duration) {
	dbQueryDuration.WithLabelValues(database, operation).Observe(duration.Seconds())
}

// RecordEvaluation records evaluation metrics
func RecordEvaluation(projectID, evaluatorID, status string) {
	evaluationsTotal.WithLabelValues(projectID, evaluatorID, status).Inc()
}

// RecordEvaluationDuration records evaluation duration
func RecordEvaluationDuration(projectID, evaluatorID string, duration time.Duration) {
	evaluationDuration.WithLabelValues(projectID, evaluatorID).Observe(duration.Seconds())
}

// SimpleMetrics creates a simple metrics middleware
func SimpleMetrics() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if c.Path() == "/health" || c.Path() == "/metrics" {
			return c.Next()
		}

		start := time.Now()

		err := c.Next()

		httpRequestsTotal.WithLabelValues(
			c.Method(),
			c.Path(),
			strconv.Itoa(c.Response().StatusCode()),
		).Inc()

		httpRequestDuration.WithLabelValues(
			c.Method(),
			c.Path(),
		).Observe(time.Since(start).Seconds())

		return err
	}
}
