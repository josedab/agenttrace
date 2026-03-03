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

// EdgeIngestService manages edge/mobile device registration and bandwidth-optimized ingestion
type EdgeIngestService struct {
	logger       *zap.Logger
	mu           sync.RWMutex
	devices      map[string]*domain.EdgeDevice
	batchCount   int64
	eventCount   int64
	offlineSyncs int64
}

// NewEdgeIngestService creates a new edge ingest service
func NewEdgeIngestService(logger *zap.Logger) *EdgeIngestService {
	return &EdgeIngestService{
		logger:  logger,
		devices: make(map[string]*domain.EdgeDevice),
	}
}

// RegisterDevice registers a new edge/mobile device
func (s *EdgeIngestService) RegisterDevice(ctx context.Context, projectID uuid.UUID, input *domain.EdgeDeviceInput) (*domain.EdgeDevice, error) {
	if input.DeviceID == "" {
		return nil, fmt.Errorf("device ID is required")
	}
	if input.Name == "" {
		return nil, fmt.Errorf("device name is required")
	}
	if input.Platform == "" {
		return nil, fmt.Errorf("platform is required")
	}

	sdkVersion := input.SDKVersion
	if sdkVersion == "" {
		sdkVersion = "1.0.0"
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Update existing or create new
	if existing, exists := s.devices[input.DeviceID]; exists {
		existing.Name = input.Name
		existing.SDKVersion = sdkVersion
		existing.Status = domain.EdgeDeviceOnline
		existing.LastSeenAt = time.Now()
		if input.Metadata != nil {
			existing.Metadata = input.Metadata
		}
		return existing, nil
	}

	device := &domain.EdgeDevice{
		ID:         uuid.New(),
		ProjectID:  projectID,
		DeviceID:   input.DeviceID,
		Name:       input.Name,
		Platform:   input.Platform,
		SDKVersion: sdkVersion,
		Status:     domain.EdgeDeviceOnline,
		LastSeenAt: time.Now(),
		Metadata:   input.Metadata,
		CreatedAt:  time.Now(),
	}

	s.devices[input.DeviceID] = device

	s.logger.Info("edge device registered",
		zap.String("deviceId", input.DeviceID),
		zap.String("platform", string(input.Platform)),
	)

	return device, nil
}

// IngestBatch ingests a batch of trace events from an edge device
func (s *EdgeIngestService) IngestBatch(ctx context.Context, projectID uuid.UUID, batch *domain.EdgeTraceBatch) (*domain.EdgeIngestResult, error) {
	if batch.DeviceID == "" {
		return nil, fmt.Errorf("device ID is required")
	}
	if len(batch.Events) == 0 {
		return nil, fmt.Errorf("at least one event is required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Update device status
	if device, exists := s.devices[batch.DeviceID]; exists {
		device.Status = domain.EdgeDeviceOnline
		device.LastSeenAt = time.Now()
		device.TotalSynced += int64(len(batch.Events))
	}

	accepted := len(batch.Events)
	s.batchCount++
	s.eventCount += int64(accepted)

	if batch.OfflineMode {
		s.offlineSyncs++
	}

	result := &domain.EdgeIngestResult{
		BatchID:       batch.BatchID,
		Accepted:      accepted,
		Rejected:      0,
		Deduplicated:  0,
		ServerTraceID: uuid.New().String(),
	}

	s.logger.Info("edge batch ingested",
		zap.String("deviceId", batch.DeviceID),
		zap.String("batchId", batch.BatchID),
		zap.Int("events", accepted),
		zap.Bool("offline", batch.OfflineMode),
	)

	return result, nil
}

// ListDevices lists registered edge devices for a project
func (s *EdgeIngestService) ListDevices(ctx context.Context, projectID uuid.UUID) ([]domain.EdgeDevice, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var devices []domain.EdgeDevice
	for _, device := range s.devices {
		if device.ProjectID == projectID {
			devices = append(devices, *device)
		}
	}

	sort.Slice(devices, func(i, j int) bool {
		return devices[i].LastSeenAt.After(devices[j].LastSeenAt)
	})

	return devices, nil
}

// GetDeviceStatus returns the status of a specific device
func (s *EdgeIngestService) GetDeviceStatus(ctx context.Context, deviceID string) (*domain.EdgeDevice, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	device, exists := s.devices[deviceID]
	if !exists {
		return nil, fmt.Errorf("device not found")
	}

	// Mark as offline if not seen recently
	if time.Since(device.LastSeenAt) > 5*time.Minute {
		device.Status = domain.EdgeDeviceOffline
	}

	return device, nil
}

// SyncOfflineData processes offline data sync from an edge device
func (s *EdgeIngestService) SyncOfflineData(ctx context.Context, projectID uuid.UUID, syncReq *domain.EdgeSyncRequest) ([]domain.EdgeIngestResult, error) {
	if syncReq.DeviceID == "" {
		return nil, fmt.Errorf("device ID is required")
	}

	var results []domain.EdgeIngestResult
	for _, batch := range syncReq.Batches {
		batch.OfflineMode = true
		result, err := s.IngestBatch(ctx, projectID, &batch)
		if err != nil {
			s.logger.Warn("failed to sync batch", zap.Error(err), zap.String("batchId", batch.BatchID))
			continue
		}
		results = append(results, *result)
	}

	return results, nil
}

// GetStats returns edge ingestion statistics
func (s *EdgeIngestService) GetStats(ctx context.Context, projectID uuid.UUID) *domain.EdgeStats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	totalDevices := 0
	onlineDevices := 0
	for _, device := range s.devices {
		if device.ProjectID == projectID {
			totalDevices++
			if device.Status == domain.EdgeDeviceOnline {
				onlineDevices++
			}
		}
	}

	avgBatch := 0
	if s.batchCount > 0 {
		avgBatch = int(s.eventCount / s.batchCount)
	}

	return &domain.EdgeStats{
		TotalDevices:   totalDevices,
		OnlineDevices:  onlineDevices,
		TotalBatches:   s.batchCount,
		TotalEvents:    s.eventCount,
		OfflineSyncs:   s.offlineSyncs,
		AvgBatchSize:   avgBatch,
		BandwidthSaved: s.eventCount * 200, // estimate 200 bytes saved per event via batching
	}
}
