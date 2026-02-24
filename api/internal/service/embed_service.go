package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

// EmbedService manages white-label embed configurations and tokens
type EmbedService struct {
	logger  *zap.Logger
	mu      sync.RWMutex
	configs map[uuid.UUID]*domain.EmbedConfig
	tokens  map[string]*domain.EmbedToken
}

// NewEmbedService creates a new embed service
func NewEmbedService(logger *zap.Logger) *EmbedService {
	return &EmbedService{
		logger:  logger,
		configs: make(map[uuid.UUID]*domain.EmbedConfig),
		tokens:  make(map[string]*domain.EmbedToken),
	}
}

// CreateConfig creates a new embed configuration
func (s *EmbedService) CreateConfig(ctx context.Context, projectID uuid.UUID, input *domain.EmbedConfigInput) (*domain.EmbedConfig, error) {
	now := time.Now()
	config := &domain.EmbedConfig{
		ID:             uuid.New(),
		ProjectID:      projectID,
		Theme:          input.Theme,
		AllowedOrigins: input.AllowedOrigins,
		Features:       input.Features,
		APIKeyID:       uuid.New(),
		Enabled:        true,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if config.AllowedOrigins == nil {
		config.AllowedOrigins = []string{}
	}

	s.mu.Lock()
	s.configs[config.ID] = config
	s.mu.Unlock()

	s.logger.Info("created embed config",
		zap.String("configId", config.ID.String()),
		zap.String("projectId", projectID.String()),
	)

	return config, nil
}

// GetConfig retrieves an embed configuration for a project
func (s *EmbedService) GetConfig(ctx context.Context, projectID uuid.UUID) (*domain.EmbedConfig, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, config := range s.configs {
		if config.ProjectID == projectID {
			return config, nil
		}
	}

	return nil, fmt.Errorf("embed config not found for project")
}

// UpdateConfig updates an existing embed configuration
func (s *EmbedService) UpdateConfig(ctx context.Context, projectID uuid.UUID, input *domain.EmbedConfigInput) (*domain.EmbedConfig, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, config := range s.configs {
		if config.ProjectID == projectID {
			config.Theme = input.Theme
			config.AllowedOrigins = input.AllowedOrigins
			if config.AllowedOrigins == nil {
				config.AllowedOrigins = []string{}
			}
			config.Features = input.Features
			config.UpdatedAt = time.Now()

			s.logger.Info("updated embed config",
				zap.String("configId", config.ID.String()),
			)

			return config, nil
		}
	}

	return nil, fmt.Errorf("embed config not found for project")
}

// GenerateToken generates a new embed access token
func (s *EmbedService) GenerateToken(ctx context.Context, projectID uuid.UUID) (*domain.EmbedToken, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var config *domain.EmbedConfig
	for _, c := range s.configs {
		if c.ProjectID == projectID {
			config = c
			break
		}
	}
	if config == nil {
		return nil, fmt.Errorf("embed config not found for project")
	}

	// Generate a simple token using SHA-256 hash
	tokenData := fmt.Sprintf("%s:%s:%d", config.ID.String(), projectID.String(), time.Now().UnixNano())
	hash := sha256.Sum256([]byte(tokenData))
	tokenStr := "at_embed_" + hex.EncodeToString(hash[:])

	token := &domain.EmbedToken{
		Token:     tokenStr,
		ConfigID:  config.ID,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	s.tokens[tokenStr] = token

	s.logger.Info("generated embed token",
		zap.String("configId", config.ID.String()),
	)

	return token, nil
}

// ValidateToken validates an embed token
func (s *EmbedService) ValidateToken(ctx context.Context, tokenStr string) (*domain.EmbedToken, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	token, exists := s.tokens[tokenStr]
	if !exists {
		return nil, fmt.Errorf("invalid token")
	}

	if time.Now().After(token.ExpiresAt) {
		return nil, fmt.Errorf("token expired")
	}

	return token, nil
}
