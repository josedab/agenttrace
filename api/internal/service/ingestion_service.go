package service

import (
	"context"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

// TraceRepository defines the interface for trace persistence operations.
// Implementations may use ClickHouse, PostgreSQL, or other storage backends.
// All methods must be safe for concurrent use.
type TraceRepository interface {
	// Create persists a new trace to storage.
	Create(ctx context.Context, trace *domain.Trace) error
	// CreateBatch persists multiple traces in a single operation for efficiency.
	CreateBatch(ctx context.Context, traces []*domain.Trace) error
	// GetByID retrieves a trace by its project-scoped ID.
	GetByID(ctx context.Context, projectID uuid.UUID, traceID string) (*domain.Trace, error)
	// Update modifies an existing trace's mutable fields.
	Update(ctx context.Context, trace *domain.Trace) error
	// UpdateCosts updates the aggregated cost fields for a trace.
	UpdateCosts(ctx context.Context, projectID uuid.UUID, traceID string, inputCost, outputCost, totalCost float64) error
	// List returns traces matching the filter with pagination.
	List(ctx context.Context, filter *domain.TraceFilter, limit, offset int) (*domain.TraceList, error)
	// SetBookmark marks or unmarks a trace as bookmarked.
	SetBookmark(ctx context.Context, projectID uuid.UUID, traceID string, bookmarked bool) error
	// GetBySessionID retrieves all traces belonging to a session.
	GetBySessionID(ctx context.Context, projectID uuid.UUID, sessionID string) ([]domain.Trace, error)
	// Delete removes a trace by ID.
	// Note: This is a heavy operation in ClickHouse (ALTER TABLE DELETE).
	Delete(ctx context.Context, projectID uuid.UUID, traceID string) error
}

// ObservationRepository defines the interface for observation persistence operations.
// Observations include spans (generic operations) and generations (LLM calls).
// All methods must be safe for concurrent use.
type ObservationRepository interface {
	// Create persists a new observation to storage.
	Create(ctx context.Context, obs *domain.Observation) error
	// CreateBatch persists multiple observations in a single operation for efficiency.
	CreateBatch(ctx context.Context, observations []*domain.Observation) error
	// GetByID retrieves an observation by its project-scoped ID.
	GetByID(ctx context.Context, projectID uuid.UUID, observationID string) (*domain.Observation, error)
	// GetByTraceID retrieves all observations belonging to a trace.
	GetByTraceID(ctx context.Context, projectID uuid.UUID, traceID string) ([]domain.Observation, error)
	// Update modifies an existing observation's mutable fields.
	Update(ctx context.Context, obs *domain.Observation) error
	// UpdateCosts updates the cost fields for an observation.
	UpdateCosts(ctx context.Context, projectID uuid.UUID, observationID string, inputCost, outputCost, totalCost float64) error
	// List returns observations matching the filter with pagination.
	List(ctx context.Context, filter *domain.ObservationFilter, limit, offset int) ([]domain.Observation, int64, error)
	// GetGenerationsWithoutCost retrieves generations that need cost calculation.
	GetGenerationsWithoutCost(ctx context.Context, projectID uuid.UUID, limit int) ([]domain.Observation, error)
	// GetTree retrieves observations as a hierarchical tree for visualization.
	GetTree(ctx context.Context, projectID uuid.UUID, traceID string) (*domain.ObservationTree, error)
}

// SessionRepository defines the interface for session persistence operations.
// Sessions group related traces together (e.g., a user conversation).
type SessionRepository interface {
	// Upsert creates or updates a session, typically called when traces reference it.
	Upsert(ctx context.Context, session *domain.Session) error
	// GetByID retrieves a session by its project-scoped ID.
	GetByID(ctx context.Context, projectID uuid.UUID, sessionID string) (*domain.Session, error)
	// List returns sessions matching the filter with pagination.
	List(ctx context.Context, filter *domain.SessionFilter, limit, offset int) (*domain.SessionList, error)
}

// IngestionService handles trace and observation ingestion from SDKs and APIs.
//
// This is the core service for receiving telemetry data from instrumented applications.
// It processes incoming traces, spans, and LLM generations, persisting them to storage
// while handling:
//   - ID generation for entities without explicit IDs
//   - Metadata and payload JSON marshaling
//   - Timestamp normalization and duration calculation
//   - Cost calculation for LLM generations (via CostService)
//   - Asynchronous evaluation triggering (via EvalService)
//
// The service is safe for concurrent use and designed for high-throughput ingestion.
//
// Implementation is split across files by domain:
//   - ingestion_service.go: struct, interfaces, constructor, types
//   - ingestion_trace.go: trace ingestion and updates
//   - ingestion_span.go: observation/span ingestion and updates
//   - ingestion_generation.go: LLM generation ingestion
//   - ingestion_batch.go: batch ingestion
type IngestionService struct {
	traceRepo        TraceRepository
	observationRepo  ObservationRepository
	costService      *CostService
	evalService      *EvalService
	guardrailService *GuardrailService
	logger           *zap.Logger
}

// NewIngestionService creates a new IngestionService with the provided dependencies.
//
// Parameters:
//   - logger: Structured logger for observability (required)
//   - traceRepo: Repository for trace persistence (required)
//   - observationRepo: Repository for observation persistence (required)
//   - costService: Service for LLM cost calculation (optional, costs won't be calculated if nil)
//   - evalService: Service for triggering evaluations (optional, evals won't trigger if nil)
//
// Returns a configured IngestionService ready for use.
func NewIngestionService(
	logger *zap.Logger,
	traceRepo TraceRepository,
	observationRepo ObservationRepository,
	costService *CostService,
	evalService *EvalService,
) *IngestionService {
	return &IngestionService{
		logger:          logger.Named("ingestion"),
		traceRepo:       traceRepo,
		observationRepo: observationRepo,
		costService:     costService,
		evalService:     evalService,
	}
}

// SetGuardrailService sets the guardrail service for ingestion-time evaluation.
// Called after service initialization to avoid circular dependencies.
func (s *IngestionService) SetGuardrailService(gs *GuardrailService) {
	s.guardrailService = gs
}

// IngestionBatchInput represents a batch of telemetry items for bulk ingestion.
//
// SDKs typically buffer telemetry locally and send batches periodically to
// reduce network overhead. This struct mirrors the domain.IngestionBatch but
// uses input types for deserialization from API requests.
//
// All arrays are optional - a batch can contain any combination of traces,
// observations, and generations. Items within each array are processed together
// in efficient batch database operations.
//
// Example JSON:
//
//	{
//	  "traces": [{"id": "trace-1", "name": "api-request"}],
//	  "observations": [{"traceId": "trace-1", "name": "db-query"}],
//	  "generations": [{"traceId": "trace-1", "model": "gpt-4", "usage": {...}}]
//	}
type IngestionBatchInput struct {
	// Traces to create (parent containers for observations)
	Traces []*domain.TraceInput `json:"traces"`
	// Observations to create (spans, events, generic operations)
	Observations []*domain.ObservationInput `json:"observations"`
	// Generations to create (LLM calls with model/usage/cost tracking)
	Generations []*domain.GenerationInput `json:"generations"`
}

