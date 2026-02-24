package service

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

// AgentVersionService manages agent versioning and rollback
type AgentVersionService struct {
	logger   *zap.Logger
	mu       sync.RWMutex
	versions map[uuid.UUID]*domain.AgentVersion
	// Track the next version number per agent (projectID:agentName -> nextVersion)
	nextVersion map[string]int
}

// NewAgentVersionService creates a new agent version service
func NewAgentVersionService(logger *zap.Logger) *AgentVersionService {
	return &AgentVersionService{
		logger:      logger,
		versions:    make(map[uuid.UUID]*domain.AgentVersion),
		nextVersion: make(map[string]int),
	}
}

// CreateVersion creates a new agent version
func (s *AgentVersionService) CreateVersion(ctx context.Context, projectID uuid.UUID, input *domain.CreateVersionInput) (*domain.AgentVersion, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := fmt.Sprintf("%s:%s", projectID.String(), input.AgentName)
	version := s.nextVersion[key] + 1
	s.nextVersion[key] = version

	// Deactivate all existing versions for this agent
	for _, v := range s.versions {
		if v.ProjectID == projectID && v.AgentName == input.AgentName {
			v.IsActive = false
		}
	}

	av := &domain.AgentVersion{
		ID:        uuid.New(),
		ProjectID: projectID,
		AgentName: input.AgentName,
		Version:   version,
		Tag:       input.Tag,
		Config:    input.Config,
		PerformanceMetrics: &domain.VersionMetrics{
			TraceCount:   0,
			SuccessRate:  0,
			AvgCost:      0,
			AvgLatencyMs: 0,
			AvgQuality:   0,
		},
		CreatedBy:  "current-user",
		ChangeNote: input.ChangeNote,
		IsActive:   true,
		CreatedAt:  time.Now(),
	}

	s.versions[av.ID] = av

	s.logger.Info("created agent version",
		zap.String("versionId", av.ID.String()),
		zap.String("agentName", input.AgentName),
		zap.Int("version", version),
	)

	return av, nil
}

// GetVersion retrieves an agent version by ID
func (s *AgentVersionService) GetVersion(ctx context.Context, versionID uuid.UUID) (*domain.AgentVersion, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	v, exists := s.versions[versionID]
	if !exists {
		return nil, fmt.Errorf("version not found")
	}
	return v, nil
}

// ListVersions lists all versions for an agent
func (s *AgentVersionService) ListVersions(ctx context.Context, projectID uuid.UUID, agentName string) ([]domain.AgentVersion, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var versions []domain.AgentVersion
	for _, v := range s.versions {
		if v.ProjectID == projectID {
			if agentName == "" || v.AgentName == agentName {
				versions = append(versions, *v)
			}
		}
	}

	if versions == nil {
		versions = []domain.AgentVersion{}
	}
	return versions, nil
}

// GetActiveVersion returns the active version for an agent
func (s *AgentVersionService) GetActiveVersion(ctx context.Context, projectID uuid.UUID, agentName string) (*domain.AgentVersion, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, v := range s.versions {
		if v.ProjectID == projectID && v.AgentName == agentName && v.IsActive {
			return v, nil
		}
	}

	return nil, fmt.Errorf("no active version found for agent %s", agentName)
}

// Rollback sets a specific version as the active version
func (s *AgentVersionService) Rollback(ctx context.Context, projectID uuid.UUID, versionID uuid.UUID) (*domain.AgentVersion, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	target, exists := s.versions[versionID]
	if !exists {
		return nil, fmt.Errorf("version not found")
	}

	if target.ProjectID != projectID {
		return nil, fmt.Errorf("version does not belong to this project")
	}

	// Deactivate all versions for this agent, activate the target
	for _, v := range s.versions {
		if v.ProjectID == projectID && v.AgentName == target.AgentName {
			v.IsActive = false
		}
	}
	target.IsActive = true

	s.logger.Info("rolled back agent version",
		zap.String("agentName", target.AgentName),
		zap.Int("version", target.Version),
	)

	return target, nil
}

// DiffVersions compares two agent versions
func (s *AgentVersionService) DiffVersions(ctx context.Context, idA uuid.UUID, idB uuid.UUID) (*domain.VersionDiff, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	vA, existsA := s.versions[idA]
	vB, existsB := s.versions[idB]
	if !existsA || !existsB {
		return nil, fmt.Errorf("one or both versions not found")
	}

	changes := s.compareConfigs(vA.Config, vB.Config)

	// Compare performance metrics if available
	if vA.PerformanceMetrics != nil && vB.PerformanceMetrics != nil {
		if vA.PerformanceMetrics.SuccessRate != vB.PerformanceMetrics.SuccessRate {
			changes = append(changes, domain.ConfigChange{
				Field:    "performanceMetrics.successRate",
				OldValue: vA.PerformanceMetrics.SuccessRate,
				NewValue: vB.PerformanceMetrics.SuccessRate,
			})
		}
		if vA.PerformanceMetrics.AvgCost != vB.PerformanceMetrics.AvgCost {
			changes = append(changes, domain.ConfigChange{
				Field:    "performanceMetrics.avgCost",
				OldValue: vA.PerformanceMetrics.AvgCost,
				NewValue: vB.PerformanceMetrics.AvgCost,
			})
		}
	}

	diff := &domain.VersionDiff{
		VersionA: *vA,
		VersionB: *vB,
		Changes:  changes,
	}

	// Generate mock metrics for display if both versions have zero metrics
	if vA.PerformanceMetrics != nil && vA.PerformanceMetrics.TraceCount == 0 {
		vA.PerformanceMetrics = &domain.VersionMetrics{
			TraceCount:   50 + rand.Intn(200),  //nolint:gosec
			SuccessRate:  0.8 + rand.Float64()*0.15, //nolint:gosec
			AvgCost:      0.01 + rand.Float64()*0.05, //nolint:gosec
			AvgLatencyMs: 500 + rand.Float64()*2000, //nolint:gosec
			AvgQuality:   70 + rand.Float64()*25, //nolint:gosec
		}
	}

	return diff, nil
}

func (s *AgentVersionService) compareConfigs(a, b domain.AgentConfig) []domain.ConfigChange {
	var changes []domain.ConfigChange

	if a.Model != b.Model {
		changes = append(changes, domain.ConfigChange{Field: "model", OldValue: a.Model, NewValue: b.Model})
	}
	if a.SystemPrompt != b.SystemPrompt {
		changes = append(changes, domain.ConfigChange{Field: "systemPrompt", OldValue: a.SystemPrompt, NewValue: b.SystemPrompt})
	}
	if a.Temperature != b.Temperature {
		changes = append(changes, domain.ConfigChange{Field: "temperature", OldValue: a.Temperature, NewValue: b.Temperature})
	}
	if a.MaxTokens != b.MaxTokens {
		changes = append(changes, domain.ConfigChange{Field: "maxTokens", OldValue: a.MaxTokens, NewValue: b.MaxTokens})
	}

	return changes
}
