package domain

import (
	"time"

	"github.com/google/uuid"
)

// WarehouseType represents the type of data warehouse
type WarehouseType string

const (
	WarehouseSnowflake  WarehouseType = "snowflake"
	WarehouseBigQuery   WarehouseType = "bigquery"
	WarehouseDatabricks WarehouseType = "databricks"
	WarehouseS3Parquet  WarehouseType = "s3_parquet"
)

// SyncDirection represents the direction of data sync
type SyncDirection string

const (
	SyncDirectionExport SyncDirection = "export"
	SyncDirectionImport SyncDirection = "import"
	SyncDirectionBidir  SyncDirection = "bidirectional"
)

// SyncStatus represents the status of a sync operation
type SyncStatus string

const (
	SyncStatusIdle      SyncStatus = "idle"
	SyncStatusRunning   SyncStatus = "running"
	SyncStatusCompleted SyncStatus = "completed"
	SyncStatusFailed    SyncStatus = "failed"
)

// WarehouseConnection represents a connection to a data warehouse
type WarehouseConnection struct {
	ID            uuid.UUID     `json:"id"`
	ProjectID     uuid.UUID     `json:"projectId"`
	Name          string        `json:"name"`
	Type          WarehouseType `json:"type"`
	Direction     SyncDirection `json:"direction"`
	Config        WarehouseConfig `json:"config"`
	SchemaMapping []SchemaMap   `json:"schemaMapping,omitempty"`
	SyncSchedule  string        `json:"syncSchedule,omitempty"` // cron expression
	LastSyncAt    *time.Time    `json:"lastSyncAt,omitempty"`
	LastSyncStatus SyncStatus   `json:"lastSyncStatus"`
	Enabled       bool          `json:"enabled"`
	CreatedAt     time.Time     `json:"createdAt"`
	UpdatedAt     time.Time     `json:"updatedAt"`
}

// WarehouseConfig holds provider-specific configuration
type WarehouseConfig struct {
	// Snowflake
	Account   string `json:"account,omitempty"`
	Database  string `json:"database,omitempty"`
	Schema    string `json:"schema,omitempty"`
	Warehouse string `json:"warehouse,omitempty"`

	// BigQuery
	ProjectGCP string `json:"projectGcp,omitempty"`
	Dataset    string `json:"dataset,omitempty"`

	// Databricks
	Host      string `json:"host,omitempty"`
	Catalog   string `json:"catalog,omitempty"`

	// S3/Parquet
	Bucket    string `json:"bucket,omitempty"`
	Prefix    string `json:"prefix,omitempty"`
	Region    string `json:"region,omitempty"`
	Format    string `json:"format,omitempty"` // "parquet", "csv", "json"
}

// SchemaMap represents a mapping between AgentTrace fields and warehouse columns
type SchemaMap struct {
	SourceField string `json:"sourceField"`
	TargetField string `json:"targetField"`
	Transform   string `json:"transform,omitempty"` // "none", "lowercase", "timestamp", "json_extract"
}

// SyncOperation represents a single sync operation
type SyncOperation struct {
	ID            uuid.UUID  `json:"id"`
	ConnectionID  uuid.UUID  `json:"connectionId"`
	Status        SyncStatus `json:"status"`
	Direction     SyncDirection `json:"direction"`
	RecordsTotal  int64      `json:"recordsTotal"`
	RecordsSynced int64      `json:"recordsSynced"`
	RecordsFailed int64      `json:"recordsFailed"`
	BytesSynced   int64      `json:"bytesSynced"`
	Error         string     `json:"error,omitempty"`
	StartedAt     time.Time  `json:"startedAt"`
	CompletedAt   *time.Time `json:"completedAt,omitempty"`
}

// WarehouseConnectionInput represents input for creating a warehouse connection
type WarehouseConnectionInput struct {
	Name          string        `json:"name" validate:"required"`
	Type          WarehouseType `json:"type" validate:"required"`
	Direction     SyncDirection `json:"direction,omitempty"`
	Config        WarehouseConfig `json:"config" validate:"required"`
	SchemaMapping []SchemaMap   `json:"schemaMapping,omitempty"`
	SyncSchedule  string        `json:"syncSchedule,omitempty"`
}

// WarehouseConnectionTest reports whether a connection is usable before syncing.
type WarehouseConnectionTest struct {
	ConnectionID uuid.UUID `json:"connectionId"`
	Reachable    bool      `json:"reachable"`
	Message      string    `json:"message,omitempty"`
	CheckedAt    time.Time `json:"checkedAt"`
}
