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

// GenerateWidgetScript generates the lightweight embeddable JavaScript widget
func (s *EmbedService) GenerateWidgetScript(ctx context.Context, token string) (string, error) {
	embedToken, err := s.ValidateToken(ctx, token)
	if err != nil {
		return "", fmt.Errorf("invalid embed token: %w", err)
	}

	config, exists := s.configs[embedToken.ConfigID]
	if !exists {
		return "", fmt.Errorf("embed config not found")
	}

	primaryColor := config.Theme.PrimaryColor
	if primaryColor == "" {
		primaryColor = "#3b82f6"
	}
	bgColor := config.Theme.BackgroundColor
	if bgColor == "" {
		bgColor = "#ffffff"
	}
	textColor := config.Theme.TextColor
	if textColor == "" {
		textColor = "#1f2937"
	}
	borderRadius := config.Theme.BorderRadius
	if borderRadius == "" {
		borderRadius = "8px"
	}

	script := fmt.Sprintf(`(function(){
"use strict";
var AT_CONFIG={token:"%s",configId:"%s",features:%s,theme:{primary:"%s",bg:"%s",text:"%s",radius:"%s"}};
var root=document.getElementById("agenttrace-widget");
if(!root){root=document.createElement("div");root.id="agenttrace-widget";document.body.appendChild(root);}
root.style.cssText="font-family:system-ui,-apple-system,sans-serif;background:"+AT_CONFIG.theme.bg+";color:"+AT_CONFIG.theme.text+";border:1px solid #e5e7eb;border-radius:"+AT_CONFIG.theme.radius+";padding:16px;max-width:400px;font-size:14px;";
var h=document.createElement("div");h.style.cssText="font-weight:600;font-size:16px;margin-bottom:12px;display:flex;align-items:center;gap:8px;";
h.innerHTML='<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="'+AT_CONFIG.theme.primary+'" stroke-width="2"><circle cx="12" cy="12" r="10"/><path d="M12 6v6l4 2"/></svg>AI Activity';
root.appendChild(h);
var c=document.createElement("div");c.id="at-content";c.style.cssText="display:flex;flex-direction:column;gap:8px;";
c.innerHTML='<div style="padding:8px;background:#f9fafb;border-radius:4px;font-size:13px;">Loading trace data...</div>';
root.appendChild(c);
%s
window.AgentTraceWidget={refresh:function(){},destroy:function(){root.remove();}};
})();`,
		token,
		embedToken.ConfigID.String(),
		s.featuresToJSON(config.Features),
		primaryColor, bgColor, textColor, borderRadius,
		s.generateFeatureCode(config.Features),
	)

	return script, nil
}

func (s *EmbedService) featuresToJSON(f domain.EmbedFeatures) string {
	return fmt.Sprintf(`{traceViewer:%v,costDashboard:%v,qualityScore:%v,activityFeed:%v}`,
		f.TraceViewer, f.CostDashboard, f.QualityScore, f.ActivityFeed)
}

func (s *EmbedService) generateFeatureCode(f domain.EmbedFeatures) string {
	code := ""
	if f.ActivityFeed {
		code += `
var feed=document.createElement("div");feed.style.cssText="font-size:13px;";
feed.innerHTML='<div style="padding:6px 0;border-bottom:1px solid #f3f4f6;">✅ Agent completed task</div><div style="padding:6px 0;border-bottom:1px solid #f3f4f6;">🔍 Searched knowledge base</div><div style="padding:6px 0;">💬 Generated response</div>';
c.innerHTML="";c.appendChild(feed);`
	}
	if f.CostDashboard {
		code += `
var cost=document.createElement("div");cost.style.cssText="display:flex;gap:12px;padding:8px;background:#f0fdf4;border-radius:4px;font-size:13px;";
cost.innerHTML='<span>💰 Cost: $0.0042</span><span>⚡ Tokens: 1,247</span>';
c.appendChild(cost);`
	}
	if f.QualityScore {
		code += `
var quality=document.createElement("div");quality.style.cssText="padding:8px;background:#eff6ff;border-radius:4px;font-size:13px;";
quality.innerHTML='<span>📊 Quality Score: <strong>94/100</strong></span>';
c.appendChild(quality);`
	}
	return code
}
