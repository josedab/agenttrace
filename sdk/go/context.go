package agenttrace

import (
	"context"
	"sync"
)

type contextKey int

const (
	traceKey contextKey = iota
	observationKey
)

var (
	globalClient *Client
	globalMu     sync.RWMutex
)

// SetGlobalClient sets the global AgentTrace client.
func SetGlobalClient(c *Client) {
	globalMu.Lock()
	defer globalMu.Unlock()
	globalClient = c
}

// GetGlobalClient returns the global AgentTrace client.
func GetGlobalClient() *Client {
	globalMu.RLock()
	defer globalMu.RUnlock()
	return globalClient
}

// SetCurrentTrace sets the current trace in the context.
func SetCurrentTrace(ctx context.Context, trace *Trace) context.Context {
	return context.WithValue(ctx, traceKey, trace)
}

// GetCurrentTrace returns the current trace from the context.
func GetCurrentTrace(ctx context.Context) *Trace {
	if trace, ok := ctx.Value(traceKey).(*Trace); ok {
		return trace
	}
	return nil
}

// SetCurrentObservation sets the current observation in the context.
func SetCurrentObservation(ctx context.Context, obs any) context.Context {
	return context.WithValue(ctx, observationKey, obs)
}

// GetCurrentObservation returns the current observation from the context.
func GetCurrentObservation(ctx context.Context) any {
	return ctx.Value(observationKey)
}

// GetCurrentSpan returns the current span from the context.
func GetCurrentSpan(ctx context.Context) *Span {
	if span, ok := ctx.Value(observationKey).(*Span); ok {
		return span
	}
	return nil
}

// GetCurrentGeneration returns the current generation from the context.
func GetCurrentGeneration(ctx context.Context) *Generation {
	if gen, ok := ctx.Value(observationKey).(*Generation); ok {
		return gen
	}
	return nil
}

// WithTrace returns a new context with the trace.
func WithTrace(ctx context.Context, trace *Trace) context.Context {
	return SetCurrentTrace(ctx, trace)
}

// WithSpan returns a new context with the span.
func WithSpan(ctx context.Context, span *Span) context.Context {
	return SetCurrentObservation(ctx, span)
}

// WithGeneration returns a new context with the generation.
func WithGeneration(ctx context.Context, gen *Generation) context.Context {
	return SetCurrentObservation(ctx, gen)
}

// StartTrace creates a new trace using the global client.
// If no client is configured, returns nil.
func StartTrace(ctx context.Context, opts TraceOptions) (*Trace, context.Context) {
	client := GetGlobalClient()
	if client == nil {
		return nil, ctx
	}

	trace := client.Trace(ctx, opts)
	return trace, WithTrace(ctx, trace)
}

// StartSpan creates a new span within the current trace.
// If no trace is in context, returns nil.
func StartSpan(ctx context.Context, opts SpanOptions) (*Span, context.Context) {
	trace := GetCurrentTrace(ctx)
	if trace == nil {
		return nil, ctx
	}

	// Set parent observation ID from context
	if opts.ParentObservationID == "" {
		if span := GetCurrentSpan(ctx); span != nil {
			opts.ParentObservationID = span.ID()
		} else if gen := GetCurrentGeneration(ctx); gen != nil {
			opts.ParentObservationID = gen.ID()
		}
	}

	span := trace.Span(opts)
	return span, WithSpan(ctx, span)
}

// StartGeneration creates a new generation within the current trace.
// If no trace is in context, returns nil.
func StartGeneration(ctx context.Context, opts GenerationOptions) (*Generation, context.Context) {
	trace := GetCurrentTrace(ctx)
	if trace == nil {
		return nil, ctx
	}

	// Set parent observation ID from context
	if opts.ParentObservationID == "" {
		if span := GetCurrentSpan(ctx); span != nil {
			opts.ParentObservationID = span.ID()
		} else if gen := GetCurrentGeneration(ctx); gen != nil {
			opts.ParentObservationID = gen.ID()
		}
	}

	generation := trace.Generation(opts)
	return generation, WithGeneration(ctx, generation)
}
