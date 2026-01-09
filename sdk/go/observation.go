package agenttrace

import (
	"time"
)

// SpanOptions holds options for creating a span.
type SpanOptions struct {
	Name                string
	ID                  string
	ParentObservationID string
	Metadata            map[string]any
	Input               any
	Level               string
}

// Span represents a span within a trace.
type Span struct {
	client              *Client
	traceID             string
	id                  string
	name                string
	parentObservationID string
	metadata            map[string]any
	input               any
	output              any
	level               string
	startTime           time.Time
	endTime             *time.Time
	ended               bool
}

// ID returns the span ID.
func (s *Span) ID() string {
	return s.id
}

// TraceID returns the trace ID.
func (s *Span) TraceID() string {
	return s.traceID
}

func (s *Span) sendCreate() {
	if !s.client.Enabled() {
		return
	}

	s.client.addEvent(map[string]any{
		"type": "span-create",
		"body": map[string]any{
			"id":                  s.id,
			"traceId":             s.traceID,
			"parentObservationId": s.parentObservationID,
			"name":                s.name,
			"metadata":            s.metadata,
			"input":               s.input,
			"level":               s.level,
			"startTime":           s.startTime.Format(time.RFC3339Nano),
		},
	})
}

// SpanEndOptions holds options for ending a span.
type SpanEndOptions struct {
	Output any
}

// End ends the span.
func (s *Span) End(opts *SpanEndOptions) {
	if s.ended {
		return
	}

	s.ended = true
	now := time.Now().UTC()
	s.endTime = &now

	if opts != nil && opts.Output != nil {
		s.output = opts.Output
	}

	if s.client.Enabled() {
		s.client.addEvent(map[string]any{
			"type": "span-update",
			"body": map[string]any{
				"id":      s.id,
				"output":  s.output,
				"endTime": s.endTime.Format(time.RFC3339Nano),
			},
		})
	}
}

// GenerationOptions holds options for creating a generation.
type GenerationOptions struct {
	Name                string
	ID                  string
	ParentObservationID string
	Model               string
	ModelParameters     map[string]any
	Input               any
	Metadata            map[string]any
	Level               string
}

// Generation represents an LLM generation within a trace.
type Generation struct {
	client              *Client
	traceID             string
	id                  string
	name                string
	parentObservationID string
	model               string
	modelParameters     map[string]any
	input               any
	output              any
	metadata            map[string]any
	level               string
	startTime           time.Time
	endTime             *time.Time
	usage               *UsageDetails
	ended               bool
}

// UsageDetails holds token usage information.
type UsageDetails struct {
	InputTokens  int `json:"inputTokens,omitempty"`
	OutputTokens int `json:"outputTokens,omitempty"`
	TotalTokens  int `json:"totalTokens,omitempty"`
}

// ID returns the generation ID.
func (g *Generation) ID() string {
	return g.id
}

// TraceID returns the trace ID.
func (g *Generation) TraceID() string {
	return g.traceID
}

func (g *Generation) sendCreate() {
	if !g.client.Enabled() {
		return
	}

	g.client.addEvent(map[string]any{
		"type": "generation-create",
		"body": map[string]any{
			"id":                  g.id,
			"traceId":             g.traceID,
			"parentObservationId": g.parentObservationID,
			"name":                g.name,
			"model":               g.model,
			"modelParameters":     g.modelParameters,
			"input":               g.input,
			"metadata":            g.metadata,
			"level":               g.level,
			"startTime":           g.startTime.Format(time.RFC3339Nano),
		},
	})
}

// GenerationUpdateOptions holds options for updating a generation.
type GenerationUpdateOptions struct {
	Output   any
	Usage    *UsageDetails
	Model    string
	Metadata map[string]any
}

// Update updates the generation.
func (g *Generation) Update(opts GenerationUpdateOptions) {
	if opts.Output != nil {
		g.output = opts.Output
	}
	if opts.Usage != nil {
		g.usage = opts.Usage
	}
	if opts.Model != "" {
		g.model = opts.Model
	}
	if opts.Metadata != nil {
		for k, v := range opts.Metadata {
			g.metadata[k] = v
		}
	}
}

// GenerationEndOptions holds options for ending a generation.
type GenerationEndOptions struct {
	Output any
	Usage  *UsageDetails
	Model  string
}

// End ends the generation.
func (g *Generation) End(opts *GenerationEndOptions) {
	if g.ended {
		return
	}

	g.ended = true
	now := time.Now().UTC()
	g.endTime = &now

	if opts != nil {
		if opts.Output != nil {
			g.output = opts.Output
		}
		if opts.Usage != nil {
			g.usage = opts.Usage
		}
		if opts.Model != "" {
			g.model = opts.Model
		}
	}

	if g.client.Enabled() {
		body := map[string]any{
			"id":      g.id,
			"output":  g.output,
			"model":   g.model,
			"endTime": g.endTime.Format(time.RFC3339Nano),
		}

		if g.usage != nil {
			body["usage"] = map[string]any{
				"inputTokens":  g.usage.InputTokens,
				"outputTokens": g.usage.OutputTokens,
				"totalTokens":  g.usage.TotalTokens,
			}
		}

		g.client.addEvent(map[string]any{
			"type": "generation-update",
			"body": body,
		})
	}
}
