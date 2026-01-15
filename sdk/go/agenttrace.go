// Package agenttrace provides the Go SDK for AgentTrace observability platform.
//
// Example usage:
//
//	client := agenttrace.New(agenttrace.Config{
//		APIKey: "your-api-key",
//		Host:   "https://api.agenttrace.io",
//	})
//	defer client.Shutdown()
//
//	ctx := context.Background()
//	trace := client.Trace(ctx, agenttrace.TraceOptions{
//		Name: "my-trace",
//	})
//
//	gen := trace.Generation(agenttrace.GenerationOptions{
//		Name:  "llm-call",
//		Model: "gpt-4",
//		Input: map[string]any{"query": "Hello"},
//	})
//	gen.End(agenttrace.GenerationEndOptions{
//		Output: "Hi there!",
//	})
//
//	trace.End(nil)
//	client.Flush()
package agenttrace

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
)

const (
	// DefaultMaxQueueSize is the maximum number of events in the queue before dropping oldest
	DefaultMaxQueueSize = 10000
)

// Config holds the configuration for the AgentTrace client.
type Config struct {
	// APIKey is the API key for authentication.
	APIKey string

	// Host is the AgentTrace API host URL.
	Host string

	// PublicKey is an optional public key for client-side usage.
	PublicKey string

	// ProjectID is an optional project ID override.
	ProjectID string

	// Enabled controls whether tracing is enabled. Defaults to true.
	Enabled *bool

	// FlushAt is the number of events before auto-flush. Defaults to 20.
	FlushAt int

	// FlushInterval is the duration between auto-flushes. Defaults to 5 seconds.
	FlushInterval time.Duration

	// MaxRetries is the number of retries for failed requests. Defaults to 3.
	MaxRetries int

	// Timeout is the request timeout. Defaults to 10 seconds.
	Timeout time.Duration

	// MaxQueueSize is the maximum number of events in the queue. Defaults to 10000.
	// When exceeded, oldest events are dropped.
	MaxQueueSize int

	// OnError is an optional callback for handling errors that occur during
	// background operations like flushing. If nil, errors are logged to stderr.
	OnError func(err error)
}

// Client is the main AgentTrace client.
type Client struct {
	config        Config
	httpClient    *http.Client
	queue         []map[string]any
	queueMu       sync.Mutex
	flushCh       chan struct{}
	doneCh        chan struct{}
	wg            sync.WaitGroup
	droppedEvents int64 // Counter for dropped events due to queue overflow
}

// New creates a new AgentTrace client.
func New(config Config) *Client {
	// Set defaults
	if config.Host == "" {
		config.Host = "https://api.agenttrace.io"
	}
	if config.Enabled == nil {
		enabled := true
		config.Enabled = &enabled
	}
	if config.FlushAt == 0 {
		config.FlushAt = 20
	}
	if config.FlushInterval == 0 {
		config.FlushInterval = 5 * time.Second
	}
	if config.MaxRetries == 0 {
		config.MaxRetries = 3
	}
	if config.Timeout == 0 {
		config.Timeout = 10 * time.Second
	}
	if config.MaxQueueSize == 0 {
		config.MaxQueueSize = DefaultMaxQueueSize
	}

	c := &Client{
		config: config,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
		queue:   make([]map[string]any, 0),
		flushCh: make(chan struct{}, 1),
		doneCh:  make(chan struct{}),
	}

	// Start background flush goroutine
	c.wg.Add(1)
	go c.flushLoop()

	// Set as global client
	SetGlobalClient(c)

	return c
}

// Enabled returns whether tracing is enabled.
func (c *Client) Enabled() bool {
	return c.config.Enabled != nil && *c.config.Enabled
}

// Trace creates a new trace.
func (c *Client) Trace(ctx context.Context, opts TraceOptions) *Trace {
	if opts.ID == "" {
		opts.ID = uuid.New().String()
	}

	trace := &Trace{
		client:    c,
		id:        opts.ID,
		name:      opts.Name,
		userID:    opts.UserID,
		sessionID: opts.SessionID,
		metadata:  opts.Metadata,
		tags:      opts.Tags,
		input:     opts.Input,
		public:    opts.Public,
		startTime: time.Now().UTC(),
	}

	if trace.metadata == nil {
		trace.metadata = make(map[string]any)
	}
	if trace.tags == nil {
		trace.tags = []string{}
	}

	trace.sendCreate()

	// Set in context
	SetCurrentTrace(ctx, trace)

	return trace
}

// Score submits a score for a trace or observation.
func (c *Client) Score(opts ScoreOptions) {
	if !c.Enabled() {
		return
	}

	c.addEvent(map[string]any{
		"type": "score-create",
		"body": map[string]any{
			"id":            uuid.New().String(),
			"traceId":       opts.TraceID,
			"observationId": opts.ObservationID,
			"name":          opts.Name,
			"value":         opts.Value,
			"dataType":      opts.DataType,
			"comment":       opts.Comment,
			"source":        "API",
		},
	})
}

// Flush sends all pending events to the server.
func (c *Client) Flush() {
	c.queueMu.Lock()
	if len(c.queue) == 0 {
		c.queueMu.Unlock()
		return
	}

	events := c.queue
	c.queue = make([]map[string]any, 0)
	c.queueMu.Unlock()

	c.sendBatch(events)
}

// Shutdown shuts down the client and flushes remaining events.
func (c *Client) Shutdown() {
	close(c.doneCh)
	c.wg.Wait()
	c.Flush()
}

func (c *Client) addEvent(event map[string]any) {
	c.queueMu.Lock()

	// Enforce queue size limit - drop oldest events if necessary
	if len(c.queue) >= c.config.MaxQueueSize {
		dropped := len(c.queue) - c.config.MaxQueueSize + 1
		c.queue = c.queue[dropped:]
		c.droppedEvents += int64(dropped)
		c.reportError(fmt.Errorf("agenttrace: queue overflow, dropped %d events (total dropped: %d)", dropped, c.droppedEvents))
	}

	c.queue = append(c.queue, event)
	shouldFlush := len(c.queue) >= c.config.FlushAt
	c.queueMu.Unlock()

	if shouldFlush {
		select {
		case c.flushCh <- struct{}{}:
		default:
		}
	}
}

func (c *Client) flushLoop() {
	defer c.wg.Done()

	ticker := time.NewTicker(c.config.FlushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-c.doneCh:
			return
		case <-c.flushCh:
			c.Flush()
		case <-ticker.C:
			c.Flush()
		}
	}
}

func (c *Client) sendBatch(events []map[string]any) {
	c.sendBatchWithContext(context.Background(), events)
}

func (c *Client) sendBatchWithContext(ctx context.Context, events []map[string]any) {
	if len(events) == 0 {
		return
	}

	payload := map[string]any{
		"batch": events,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		c.reportError(fmt.Errorf("agenttrace: failed to marshal batch: %w", err))
		return
	}

	url := c.config.Host + "/api/public/ingestion"

	var lastErr error
	for attempt := 0; attempt < c.config.MaxRetries; attempt++ {
		// Check if context is cancelled
		if ctx.Err() != nil {
			c.reportError(fmt.Errorf("agenttrace: context cancelled: %w", ctx.Err()))
			return
		}

		// Create a timeout context for this specific request
		reqCtx, cancel := context.WithTimeout(ctx, c.config.Timeout)

		req, err := http.NewRequestWithContext(reqCtx, "POST", url, bytes.NewReader(data))
		if err != nil {
			cancel()
			lastErr = fmt.Errorf("failed to create request: %w", err)
			continue
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+c.config.APIKey)
		req.Header.Set("User-Agent", "agenttrace-go/0.1.0")

		resp, err := c.httpClient.Do(req)
		cancel() // Always cancel after request completes

		if err != nil {
			lastErr = fmt.Errorf("request failed: %w", err)
			// Check if it was a context cancellation
			if ctx.Err() != nil {
				c.reportError(fmt.Errorf("agenttrace: context cancelled during request: %w", ctx.Err()))
				return
			}
			time.Sleep(time.Duration(1<<attempt) * 500 * time.Millisecond)
			continue
		}
		resp.Body.Close()

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return
		}

		if resp.StatusCode == 429 {
			retryAfter := 5
			if h := resp.Header.Get("Retry-After"); h != "" {
				fmt.Sscanf(h, "%d", &retryAfter)
			}
			lastErr = fmt.Errorf("rate limited (429), retry after %ds", retryAfter)
			time.Sleep(time.Duration(retryAfter) * time.Second)
			continue
		}

		if resp.StatusCode >= 500 {
			lastErr = fmt.Errorf("server error: %d", resp.StatusCode)
			time.Sleep(time.Duration(1<<attempt) * 500 * time.Millisecond)
			continue
		}

		// Client error - don't retry
		c.reportError(fmt.Errorf("agenttrace: client error %d, not retrying", resp.StatusCode))
		return
	}

	// All retries exhausted
	if lastErr != nil {
		c.reportError(fmt.Errorf("agenttrace: failed after %d attempts: %w", c.config.MaxRetries, lastErr))
	}
}

// reportError reports an error using the configured callback or logs to stderr
func (c *Client) reportError(err error) {
	if c.config.OnError != nil {
		c.config.OnError(err)
	} else {
		log.Printf("[agenttrace] %v", err)
	}
}

// DroppedEvents returns the total number of events dropped due to queue overflow
func (c *Client) DroppedEvents() int64 {
	c.queueMu.Lock()
	defer c.queueMu.Unlock()
	return c.droppedEvents
}

// TraceOptions holds options for creating a trace.
type TraceOptions struct {
	Name      string
	ID        string
	UserID    string
	SessionID string
	Metadata  map[string]any
	Tags      []string
	Input     any
	Public    bool
}

// ScoreOptions holds options for creating a score.
type ScoreOptions struct {
	TraceID       string
	Name          string
	Value         any
	ObservationID string
	DataType      string
	Comment       string
}
