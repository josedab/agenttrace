package circuitbreaker

import (
	"context"
	"errors"
	"sync"
	"time"
)

var (
	// ErrCircuitOpen is returned when the circuit breaker is open
	ErrCircuitOpen = errors.New("circuit breaker is open")
	// ErrTooManyRequests is returned when the circuit breaker is half-open and already processing a request
	ErrTooManyRequests = errors.New("too many requests, circuit breaker is half-open")
)

// State represents the circuit breaker state
type State int

const (
	// StateClosed allows requests to pass through
	StateClosed State = iota
	// StateOpen blocks all requests
	StateOpen
	// StateHalfOpen allows one request to test the service
	StateHalfOpen
)

func (s State) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateOpen:
		return "open"
	case StateHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

// Config holds circuit breaker configuration
type Config struct {
	// Name of the circuit breaker (for logging/metrics)
	Name string
	// MaxFailures is the number of failures before opening the circuit
	MaxFailures int
	// Timeout is how long to wait before transitioning from open to half-open
	Timeout time.Duration
	// MaxHalfOpenRequests is the number of requests allowed in half-open state
	MaxHalfOpenRequests int
	// OnStateChange is called when the circuit state changes
	OnStateChange func(name string, from, to State)
}

// DefaultConfig returns a default circuit breaker configuration
func DefaultConfig(name string) Config {
	return Config{
		Name:                name,
		MaxFailures:         5,
		Timeout:             30 * time.Second,
		MaxHalfOpenRequests: 1,
	}
}

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
	config Config

	mu               sync.Mutex
	state            State
	failures         int
	successes        int
	lastFailureTime  time.Time
	halfOpenRequests int
}

// New creates a new circuit breaker with the given configuration
func New(config Config) *CircuitBreaker {
	if config.MaxFailures <= 0 {
		config.MaxFailures = 5
	}
	if config.Timeout <= 0 {
		config.Timeout = 30 * time.Second
	}
	if config.MaxHalfOpenRequests <= 0 {
		config.MaxHalfOpenRequests = 1
	}

	return &CircuitBreaker{
		config: config,
		state:  StateClosed,
	}
}

// Execute runs the given function with circuit breaker protection
func (cb *CircuitBreaker) Execute(ctx context.Context, fn func() error) error {
	if err := cb.beforeRequest(); err != nil {
		return err
	}

	// Check context before executing
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	err := fn()
	cb.afterRequest(err)
	return err
}

// ExecuteWithResult runs the given function and returns its result with circuit breaker protection
func ExecuteWithResult[T any](cb *CircuitBreaker, ctx context.Context, fn func() (T, error)) (T, error) {
	var zero T

	if err := cb.beforeRequest(); err != nil {
		return zero, err
	}

	// Check context before executing
	select {
	case <-ctx.Done():
		return zero, ctx.Err()
	default:
	}

	result, err := fn()
	cb.afterRequest(err)
	return result, err
}

// beforeRequest checks if the request should be allowed
func (cb *CircuitBreaker) beforeRequest() error {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case StateClosed:
		return nil

	case StateOpen:
		// Check if timeout has passed
		if time.Since(cb.lastFailureTime) >= cb.config.Timeout {
			cb.transitionTo(StateHalfOpen)
			cb.halfOpenRequests++
			return nil
		}
		return ErrCircuitOpen

	case StateHalfOpen:
		// Allow limited requests in half-open state
		if cb.halfOpenRequests >= cb.config.MaxHalfOpenRequests {
			return ErrTooManyRequests
		}
		cb.halfOpenRequests++
		return nil
	}

	return nil
}

// afterRequest records the result of the request
func (cb *CircuitBreaker) afterRequest(err error) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		cb.recordFailure()
	} else {
		cb.recordSuccess()
	}
}

// recordFailure records a failed request
func (cb *CircuitBreaker) recordFailure() {
	cb.failures++
	cb.lastFailureTime = time.Now()

	switch cb.state {
	case StateClosed:
		if cb.failures >= cb.config.MaxFailures {
			cb.transitionTo(StateOpen)
		}
	case StateHalfOpen:
		// Failed test request, go back to open
		cb.transitionTo(StateOpen)
	}
}

// recordSuccess records a successful request
func (cb *CircuitBreaker) recordSuccess() {
	switch cb.state {
	case StateClosed:
		// Reset failures on success
		cb.failures = 0

	case StateHalfOpen:
		cb.successes++
		// If we've had enough successes, close the circuit
		if cb.successes >= cb.config.MaxHalfOpenRequests {
			cb.transitionTo(StateClosed)
		}
	}
}

// transitionTo changes the circuit breaker state
func (cb *CircuitBreaker) transitionTo(newState State) {
	if cb.state == newState {
		return
	}

	oldState := cb.state
	cb.state = newState

	// Reset counters on state change
	switch newState {
	case StateClosed:
		cb.failures = 0
		cb.successes = 0
		cb.halfOpenRequests = 0
	case StateOpen:
		cb.successes = 0
		cb.halfOpenRequests = 0
	case StateHalfOpen:
		cb.halfOpenRequests = 0
		cb.successes = 0
	}

	// Notify listener
	if cb.config.OnStateChange != nil {
		go cb.config.OnStateChange(cb.config.Name, oldState, newState)
	}
}

// State returns the current state of the circuit breaker
func (cb *CircuitBreaker) State() State {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.state
}

// Failures returns the current failure count
func (cb *CircuitBreaker) Failures() int {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.failures
}

// Reset resets the circuit breaker to closed state
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.transitionTo(StateClosed)
}

// Registry is a global registry of circuit breakers
type Registry struct {
	mu       sync.RWMutex
	breakers map[string]*CircuitBreaker
}

// NewRegistry creates a new circuit breaker registry
func NewRegistry() *Registry {
	return &Registry{
		breakers: make(map[string]*CircuitBreaker),
	}
}

// Get returns a circuit breaker by name, creating it if it doesn't exist
func (r *Registry) Get(name string, config ...Config) *CircuitBreaker {
	r.mu.RLock()
	if cb, ok := r.breakers[name]; ok {
		r.mu.RUnlock()
		return cb
	}
	r.mu.RUnlock()

	// Create new circuit breaker
	r.mu.Lock()
	defer r.mu.Unlock()

	// Double-check after acquiring write lock
	if cb, ok := r.breakers[name]; ok {
		return cb
	}

	cfg := DefaultConfig(name)
	if len(config) > 0 {
		cfg = config[0]
	}

	cb := New(cfg)
	r.breakers[name] = cb
	return cb
}

// Stats returns statistics for all circuit breakers
func (r *Registry) Stats() map[string]map[string]interface{} {
	r.mu.RLock()
	defer r.mu.RUnlock()

	stats := make(map[string]map[string]interface{})
	for name, cb := range r.breakers {
		stats[name] = map[string]interface{}{
			"state":    cb.State().String(),
			"failures": cb.Failures(),
		}
	}
	return stats
}

// Global registry instance
var globalRegistry = NewRegistry()

// GetCircuitBreaker returns a circuit breaker from the global registry
func GetCircuitBreaker(name string, config ...Config) *CircuitBreaker {
	return globalRegistry.Get(name, config...)
}

// GlobalStats returns statistics for all global circuit breakers
func GlobalStats() map[string]map[string]interface{} {
	return globalRegistry.Stats()
}
