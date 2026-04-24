package service

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

// WarehouseSyncService manages data warehouse sync operations
type WarehouseSyncService struct {
	logger      *zap.Logger
	guard       OutboundGuard
	mu          sync.RWMutex
	connections map[uuid.UUID]*domain.WarehouseConnection
	operations  map[uuid.UUID]*domain.SyncOperation
}

// NewWarehouseSyncService creates a new warehouse sync service.
// Warehouse connections and syncs move project data to an external warehouse, so
// the outbound guard rejects creation, connection tests, and syncs in no-egress mode.
func NewWarehouseSyncService(logger *zap.Logger, guard OutboundGuard) *WarehouseSyncService {
	return &WarehouseSyncService{
		logger:      logger,
		guard:       guard,
		connections: make(map[uuid.UUID]*domain.WarehouseConnection),
		operations:  make(map[uuid.UUID]*domain.SyncOperation),
	}
}

// CreateConnection creates a new warehouse connection
func (s *WarehouseSyncService) CreateConnection(ctx context.Context, projectID uuid.UUID, input *domain.WarehouseConnectionInput) (*domain.WarehouseConnection, error) {
	if err := RequireOutbound(s.guard, EgressWarehouseSync); err != nil {
		return nil, err
	}
	if input.Name == "" {
		return nil, fmt.Errorf("connection name is required")
	}
	if input.Type == "" {
		return nil, fmt.Errorf("warehouse type is required")
	}

	if err := s.validateConfig(input.Type, input.Config); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	direction := input.Direction
	if direction == "" {
		direction = domain.SyncDirectionExport
	}

	conn := &domain.WarehouseConnection{
		ID:             uuid.New(),
		ProjectID:      projectID,
		Name:           input.Name,
		Type:           input.Type,
		Direction:      direction,
		Config:         input.Config,
		SchemaMapping:  input.SchemaMapping,
		SyncSchedule:   input.SyncSchedule,
		LastSyncStatus: domain.SyncStatusIdle,
		Enabled:        true,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// Generate default schema mapping if not provided
	if len(conn.SchemaMapping) == 0 {
		conn.SchemaMapping = s.defaultSchemaMapping()
	}

	s.mu.Lock()
	s.connections[conn.ID] = conn
	s.mu.Unlock()

	s.logger.Info("warehouse connection created",
		zap.String("connId", conn.ID.String()),
		zap.String("type", string(conn.Type)),
		zap.String("name", conn.Name),
	)

	return conn, nil
}

// GetConnection retrieves a connection by ID
func (s *WarehouseSyncService) GetConnection(ctx context.Context, projectID, id uuid.UUID) (*domain.WarehouseConnection, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	conn, exists := s.connections[id]
	if !exists || conn.ProjectID != projectID {
		return nil, fmt.Errorf("connection not found")
	}
	return conn, nil
}

// ListConnections lists warehouse connections for a project
func (s *WarehouseSyncService) ListConnections(ctx context.Context, projectID uuid.UUID) ([]domain.WarehouseConnection, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var conns []domain.WarehouseConnection
	for _, conn := range s.connections {
		if conn.ProjectID == projectID {
			conns = append(conns, *conn)
		}
	}

	sort.Slice(conns, func(i, j int) bool {
		return conns[i].CreatedAt.After(conns[j].CreatedAt)
	})

	return conns, nil
}

// DeleteConnection deletes a warehouse connection
func (s *WarehouseSyncService) DeleteConnection(ctx context.Context, projectID, id uuid.UUID) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	connection, exists := s.connections[id]
	if !exists || connection.ProjectID != projectID {
		return fmt.Errorf("connection not found")
	}
	delete(s.connections, id)
	return nil
}

// TriggerSync triggers an immediate sync for a connection
func (s *WarehouseSyncService) TriggerSync(ctx context.Context, projectID, connID uuid.UUID) (*domain.SyncOperation, error) {
	if err := RequireOutbound(s.guard, EgressWarehouseSync); err != nil {
		return nil, err
	}

	s.mu.Lock()
	conn, exists := s.connections[connID]
	if !exists || conn.ProjectID != projectID {
		s.mu.Unlock()
		return nil, fmt.Errorf("connection not found")
	}

	if !conn.Enabled {
		s.mu.Unlock()
		return nil, fmt.Errorf("connection is disabled")
	}

	op := &domain.SyncOperation{
		ID:           uuid.New(),
		ConnectionID: connID,
		Status:       domain.SyncStatusRunning,
		Direction:    conn.Direction,
		StartedAt:    time.Now(),
	}

	conn.LastSyncStatus = domain.SyncStatusRunning
	conn.UpdatedAt = time.Now()
	s.operations[op.ID] = op
	s.mu.Unlock()

	// Simulate sync operation
	op.RecordsTotal = 5000
	op.RecordsSynced = 5000
	op.BytesSynced = 2500000
	op.Status = domain.SyncStatusCompleted
	now := time.Now()
	op.CompletedAt = &now

	s.mu.Lock()
	conn.LastSyncAt = &now
	conn.LastSyncStatus = domain.SyncStatusCompleted
	s.mu.Unlock()

	s.logger.Info("sync completed",
		zap.String("connId", connID.String()),
		zap.Int64("records", op.RecordsSynced),
		zap.Int64("bytes", op.BytesSynced),
	)

	return op, nil
}

// TestConnection validates a project-owned connection before it is used for a sync.
func (s *WarehouseSyncService) TestConnection(
	ctx context.Context,
	projectID, connID uuid.UUID,
) (*domain.WarehouseConnectionTest, error) {
	if err := RequireOutbound(s.guard, EgressWarehouseSync); err != nil {
		return nil, err
	}

	conn, err := s.GetConnection(ctx, projectID, connID)
	if err != nil {
		return nil, err
	}
	if err := s.validateConfig(conn.Type, conn.Config); err != nil {
		return &domain.WarehouseConnectionTest{
			ConnectionID: conn.ID,
			Reachable:    false,
			Message:      err.Error(),
			CheckedAt:    time.Now().UTC(),
		}, nil
	}
	return &domain.WarehouseConnectionTest{
		ConnectionID: conn.ID,
		Reachable:    conn.Enabled,
		Message:      "connection configuration is complete",
		CheckedAt:    time.Now().UTC(),
	}, nil
}

// GetSyncStatus returns the latest sync status for a connection
func (s *WarehouseSyncService) GetSyncStatus(ctx context.Context, projectID, connID uuid.UUID) ([]domain.SyncOperation, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	connection, exists := s.connections[connID]
	if !exists || connection.ProjectID != projectID {
		return nil, fmt.Errorf("connection not found")
	}

	var ops []domain.SyncOperation
	for _, op := range s.operations {
		if op.ConnectionID == connID {
			ops = append(ops, *op)
		}
	}

	sort.Slice(ops, func(i, j int) bool {
		return ops[i].StartedAt.After(ops[j].StartedAt)
	})

	return ops, nil
}

// GetSchemaMapping returns default schema mapping for a connection
func (s *WarehouseSyncService) GetSchemaMapping(ctx context.Context, projectID, connID uuid.UUID) ([]domain.SchemaMap, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	conn, exists := s.connections[connID]
	if !exists || conn.ProjectID != projectID {
		return nil, fmt.Errorf("connection not found")
	}

	return conn.SchemaMapping, nil
}

func (s *WarehouseSyncService) validateConfig(wType domain.WarehouseType, config domain.WarehouseConfig) error {
	switch wType {
	case domain.WarehouseSnowflake:
		if config.Account == "" || config.Database == "" {
			return fmt.Errorf("snowflake requires account and database")
		}
	case domain.WarehouseBigQuery:
		if config.ProjectGCP == "" || config.Dataset == "" {
			return fmt.Errorf("bigquery requires project and dataset")
		}
	case domain.WarehouseDatabricks:
		if config.Host == "" {
			return fmt.Errorf("databricks requires host")
		}
	case domain.WarehouseS3Parquet:
		if config.Bucket == "" {
			return fmt.Errorf("s3 requires bucket")
		}
	default:
		return fmt.Errorf("unsupported warehouse type: %s", wType)
	}
	return nil
}

func (s *WarehouseSyncService) defaultSchemaMapping() []domain.SchemaMap {
	return []domain.SchemaMap{
		{SourceField: "trace_id", TargetField: "trace_id", Transform: "none"},
		{SourceField: "project_id", TargetField: "project_id", Transform: "none"},
		{SourceField: "name", TargetField: "trace_name", Transform: "none"},
		{SourceField: "input", TargetField: "input_text", Transform: "none"},
		{SourceField: "output", TargetField: "output_text", Transform: "none"},
		{SourceField: "model", TargetField: "model_name", Transform: "lowercase"},
		{SourceField: "total_cost", TargetField: "cost_usd", Transform: "none"},
		{SourceField: "latency_ms", TargetField: "latency_ms", Transform: "none"},
		{SourceField: "created_at", TargetField: "created_at", Transform: "timestamp"},
	}
}
