// Package middleware provides HTTP middleware for automatic tracing.
package middleware

import (
	"net/http"
	"time"

	agenttrace "github.com/agenttrace/agenttrace-go"
)

// HTTPMiddlewareConfig holds configuration for the HTTP middleware.
type HTTPMiddlewareConfig struct {
	// TraceName is the name template for traces. Defaults to "{method} {path}".
	TraceName func(r *http.Request) string

	// CaptureRequestBody enables capturing request bodies.
	CaptureRequestBody bool

	// CaptureResponseBody enables capturing response bodies.
	CaptureResponseBody bool

	// SkipPaths is a list of paths to skip tracing.
	SkipPaths []string

	// ExtractUserID extracts a user ID from the request.
	ExtractUserID func(r *http.Request) string

	// ExtractSessionID extracts a session ID from the request.
	ExtractSessionID func(r *http.Request) string
}

// HTTP returns an HTTP middleware that traces requests.
func HTTP(config *HTTPMiddlewareConfig) func(http.Handler) http.Handler {
	if config == nil {
		config = &HTTPMiddlewareConfig{}
	}

	if config.TraceName == nil {
		config.TraceName = func(r *http.Request) string {
			return r.Method + " " + r.URL.Path
		}
	}

	skipPaths := make(map[string]struct{})
	for _, path := range config.SkipPaths {
		skipPaths[path] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check if path should be skipped
			if _, skip := skipPaths[r.URL.Path]; skip {
				next.ServeHTTP(w, r)
				return
			}

			client := agenttrace.GetGlobalClient()
			if client == nil || !client.Enabled() {
				next.ServeHTTP(w, r)
				return
			}

			// Build trace options
			opts := agenttrace.TraceOptions{
				Name: config.TraceName(r),
				Metadata: map[string]any{
					"http.method":     r.Method,
					"http.url":        r.URL.String(),
					"http.path":       r.URL.Path,
					"http.host":       r.Host,
					"http.user_agent": r.UserAgent(),
				},
			}

			if config.ExtractUserID != nil {
				opts.UserID = config.ExtractUserID(r)
			}

			if config.ExtractSessionID != nil {
				opts.SessionID = config.ExtractSessionID(r)
			}

			// Create trace
			trace, ctx := agenttrace.StartTrace(r.Context(), opts)
			if trace == nil {
				next.ServeHTTP(w, r)
				return
			}

			// Wrap response writer
			rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			// Call next handler
			start := time.Now()
			next.ServeHTTP(rw, r.WithContext(ctx))
			duration := time.Since(start)

			// End trace
			trace.Update(agenttrace.TraceUpdateOptions{
				Metadata: map[string]any{
					"http.status_code": rw.statusCode,
					"http.duration_ms": duration.Milliseconds(),
				},
			})

			output := map[string]any{
				"status_code": rw.statusCode,
			}

			trace.End(&agenttrace.TraceEndOptions{Output: output})
		})
	}
}

// responseWriter wraps http.ResponseWriter to capture status code.
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
