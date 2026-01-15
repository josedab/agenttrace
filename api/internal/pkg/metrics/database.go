// Package metrics provides Prometheus metrics recording for internal packages.
// This package exists to avoid import cycles between database and middleware packages.
package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// dbQueryDuration tracks database query duration in seconds
	dbQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "agenttrace_db_query_duration_seconds",
			Help:    "Database query duration in seconds",
			Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5},
		},
		[]string{"database", "operation"},
	)

	// dbQueryTotal tracks total database queries
	dbQueryTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "agenttrace_db_queries_total",
			Help: "Total number of database queries",
		},
		[]string{"database", "operation"},
	)

	// dbQueryErrors tracks database query errors
	dbQueryErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "agenttrace_db_query_errors_total",
			Help: "Total number of database query errors",
		},
		[]string{"database", "operation"},
	)

	// dbSlowQueries tracks slow database queries
	dbSlowQueries = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "agenttrace_db_slow_queries_total",
			Help: "Total number of slow database queries (>100ms)",
		},
		[]string{"database", "operation"},
	)
)

// RecordDBQuery records database query metrics
func RecordDBQuery(database, operation string, duration time.Duration) {
	dbQueryTotal.WithLabelValues(database, operation).Inc()
	dbQueryDuration.WithLabelValues(database, operation).Observe(duration.Seconds())

	// Track slow queries
	if duration > 100*time.Millisecond {
		dbSlowQueries.WithLabelValues(database, operation).Inc()
	}
}

// RecordDBError records a database query error
func RecordDBError(database, operation string) {
	dbQueryErrors.WithLabelValues(database, operation).Inc()
}
