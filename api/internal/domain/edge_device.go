package domain

import (
	"time"

	"github.com/google/uuid"
)

// EdgePlatform represents the platform type for edge devices
type EdgePlatform string

const (
	EdgePlatformIOS     EdgePlatform = "ios"
	EdgePlatformAndroid EdgePlatform = "android"
	EdgePlatformWASM    EdgePlatform = "wasm"
	EdgePlatformIoT     EdgePlatform = "iot"
	EdgePlatformDesktop EdgePlatform = "desktop"
)

// EdgeDeviceStatus represents the status of an edge device
type EdgeDeviceStatus string

const (
	EdgeDeviceOnline  EdgeDeviceStatus = "online"
	EdgeDeviceOffline EdgeDeviceStatus = "offline"
	EdgeDeviceSyncing EdgeDeviceStatus = "syncing"
)

// EdgeDevice represents a registered edge/mobile device
type EdgeDevice struct {
	ID           uuid.UUID        `json:"id"`
	ProjectID    uuid.UUID        `json:"projectId"`
	DeviceID     string           `json:"deviceId"`
	Name         string           `json:"name"`
	Platform     EdgePlatform     `json:"platform"`
	SDKVersion   string           `json:"sdkVersion"`
	Status       EdgeDeviceStatus `json:"status"`
	LastSeenAt   time.Time        `json:"lastSeenAt"`
	BufferedEvents int            `json:"bufferedEvents"`
	TotalSynced  int64            `json:"totalSynced"`
	Metadata     map[string]string `json:"metadata,omitempty"`
	CreatedAt    time.Time        `json:"createdAt"`
}

// EdgeTraceBatch represents a batch of traces from an edge device
type EdgeTraceBatch struct {
	DeviceID    string      `json:"deviceId"`
	BatchID     string      `json:"batchId"`
	Events      []EdgeEvent `json:"events"`
	Compressed  bool        `json:"compressed"`
	OfflineMode bool        `json:"offlineMode"`
	Timestamp   time.Time   `json:"timestamp"`
}

// EdgeEvent represents a single trace event from an edge device
type EdgeEvent struct {
	ID        string            `json:"id"`
	Type      string            `json:"type"` // "trace", "span", "metric", "log"
	Name      string            `json:"name"`
	Input     string            `json:"input,omitempty"`
	Output    string            `json:"output,omitempty"`
	Model     string            `json:"model,omitempty"`
	LatencyMs int64             `json:"latencyMs,omitempty"`
	TokensIn  int               `json:"tokensIn,omitempty"`
	TokensOut int               `json:"tokensOut,omitempty"`
	Metadata  map[string]string `json:"metadata,omitempty"`
	Timestamp time.Time         `json:"timestamp"`
}

// EdgeIngestResult represents the result of ingesting an edge batch
type EdgeIngestResult struct {
	BatchID       string `json:"batchId"`
	Accepted      int    `json:"accepted"`
	Rejected      int    `json:"rejected"`
	Deduplicated  int    `json:"deduplicated"`
	ServerTraceID string `json:"serverTraceId,omitempty"`
}

// EdgeSyncRequest represents a request to sync offline data
type EdgeSyncRequest struct {
	DeviceID     string           `json:"deviceId"`
	Batches      []EdgeTraceBatch `json:"batches"`
	LastSyncedAt *time.Time       `json:"lastSyncedAt,omitempty"`
}

// EdgeDeviceInput represents input for registering an edge device
type EdgeDeviceInput struct {
	DeviceID   string           `json:"deviceId" validate:"required"`
	Name       string           `json:"name" validate:"required"`
	Platform   EdgePlatform     `json:"platform" validate:"required"`
	SDKVersion string           `json:"sdkVersion,omitempty"`
	Metadata   map[string]string `json:"metadata,omitempty"`
}

// EdgeStats represents edge ingestion statistics
type EdgeStats struct {
	TotalDevices    int   `json:"totalDevices"`
	OnlineDevices   int   `json:"onlineDevices"`
	TotalBatches    int64 `json:"totalBatches"`
	TotalEvents     int64 `json:"totalEvents"`
	OfflineSyncs    int64 `json:"offlineSyncs"`
	AvgBatchSize    int   `json:"avgBatchSize"`
	BandwidthSaved  int64 `json:"bandwidthSavedBytes"`
}
