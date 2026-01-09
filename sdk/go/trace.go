package agenttrace

import (
	"time"

	"github.com/google/uuid"
)

// Trace represents a trace in AgentTrace.
type Trace struct {
	client    *Client
	id        string
	name      string
	userID    string
	sessionID string
	metadata  map[string]any
	tags      []string
	input     any
	output    any
	public    bool
	startTime time.Time
	endTime   *time.Time
	ended     bool
}

// ID returns the trace ID.
func (t *Trace) ID() string {
	return t.id
}

// Name returns the trace name.
func (t *Trace) Name() string {
	return t.name
}

func (t *Trace) sendCreate() {
	if !t.client.Enabled() {
		return
	}

	t.client.addEvent(map[string]any{
		"type": "trace-create",
		"body": map[string]any{
			"id":        t.id,
			"name":      t.name,
			"userId":    t.userID,
			"sessionId": t.sessionID,
			"metadata":  t.metadata,
			"tags":      t.tags,
			"input":     t.input,
			"public":    t.public,
			"timestamp": t.startTime.Format(time.RFC3339Nano),
		},
	})
}

// Span creates a span within this trace.
func (t *Trace) Span(opts SpanOptions) *Span {
	if opts.ID == "" {
		opts.ID = uuid.New().String()
	}

	span := &Span{
		client:              t.client,
		traceID:             t.id,
		id:                  opts.ID,
		name:                opts.Name,
		parentObservationID: opts.ParentObservationID,
		metadata:            opts.Metadata,
		input:               opts.Input,
		level:               opts.Level,
		startTime:           time.Now().UTC(),
	}

	if span.metadata == nil {
		span.metadata = make(map[string]any)
	}
	if span.level == "" {
		span.level = "DEFAULT"
	}

	span.sendCreate()
	return span
}

// Generation creates a generation (LLM call) within this trace.
func (t *Trace) Generation(opts GenerationOptions) *Generation {
	if opts.ID == "" {
		opts.ID = uuid.New().String()
	}

	gen := &Generation{
		client:              t.client,
		traceID:             t.id,
		id:                  opts.ID,
		name:                opts.Name,
		parentObservationID: opts.ParentObservationID,
		model:               opts.Model,
		modelParameters:     opts.ModelParameters,
		input:               opts.Input,
		metadata:            opts.Metadata,
		level:               opts.Level,
		startTime:           time.Now().UTC(),
	}

	if gen.metadata == nil {
		gen.metadata = make(map[string]any)
	}
	if gen.modelParameters == nil {
		gen.modelParameters = make(map[string]any)
	}
	if gen.level == "" {
		gen.level = "DEFAULT"
	}

	gen.sendCreate()
	return gen
}

// TraceUpdateOptions holds options for updating a trace.
type TraceUpdateOptions struct {
	Name      *string
	UserID    *string
	SessionID *string
	Metadata  map[string]any
	Tags      []string
	Input     any
	Output    any
	Public    *bool
}

// Update updates the trace properties.
func (t *Trace) Update(opts TraceUpdateOptions) {
	if opts.Name != nil {
		t.name = *opts.Name
	}
	if opts.UserID != nil {
		t.userID = *opts.UserID
	}
	if opts.SessionID != nil {
		t.sessionID = *opts.SessionID
	}
	if opts.Metadata != nil {
		for k, v := range opts.Metadata {
			t.metadata[k] = v
		}
	}
	if opts.Tags != nil {
		t.tags = opts.Tags
	}
	if opts.Input != nil {
		t.input = opts.Input
	}
	if opts.Output != nil {
		t.output = opts.Output
	}
	if opts.Public != nil {
		t.public = *opts.Public
	}

	if t.client.Enabled() {
		t.client.addEvent(map[string]any{
			"type": "trace-update",
			"body": map[string]any{
				"id":        t.id,
				"name":      t.name,
				"userId":    t.userID,
				"sessionId": t.sessionID,
				"metadata":  t.metadata,
				"tags":      t.tags,
				"input":     t.input,
				"output":    t.output,
				"public":    t.public,
			},
		})
	}
}

// TraceEndOptions holds options for ending a trace.
type TraceEndOptions struct {
	Output any
}

// End ends the trace.
func (t *Trace) End(opts *TraceEndOptions) {
	if t.ended {
		return
	}

	t.ended = true
	now := time.Now().UTC()
	t.endTime = &now

	if opts != nil && opts.Output != nil {
		t.output = opts.Output
	}

	t.Update(TraceUpdateOptions{Output: t.output})
}

// Score adds a score to this trace.
func (t *Trace) Score(name string, value any, opts *ScoreAddOptions) {
	scoreOpts := ScoreOptions{
		TraceID: t.id,
		Name:    name,
		Value:   value,
	}

	if opts != nil {
		scoreOpts.DataType = opts.DataType
		scoreOpts.Comment = opts.Comment
	}

	t.client.Score(scoreOpts)
}

// ScoreAddOptions holds additional options for adding a score.
type ScoreAddOptions struct {
	DataType string
	Comment  string
}
