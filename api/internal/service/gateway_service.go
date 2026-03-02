package service

import (
	"context"
	"fmt"
	"math"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

// GatewayService manages the LLM gateway and smart routing
type GatewayService struct {
	logger   *zap.Logger
	mu       sync.RWMutex
	configs  map[uuid.UUID]*domain.GatewayConfig
	rules    map[uuid.UUID][]domain.RoutingRule
	stats    *gatewayStats
	reqCount atomic.Int64
}

type gatewayStats struct {
	mu             sync.Mutex
	totalRequests  int64
	totalTokens    int64
	totalCost      float64
	totalLatency   int64
	errorCount     int64
	fallbackCount  int64
	providerCounts map[string]int64
}

// NewGatewayService creates a new gateway service
func NewGatewayService(logger *zap.Logger) *GatewayService {
	return &GatewayService{
		logger:  logger,
		configs: make(map[uuid.UUID]*domain.GatewayConfig),
		rules:   make(map[uuid.UUID][]domain.RoutingRule),
		stats: &gatewayStats{
			providerCounts: make(map[string]int64),
		},
	}
}

// CreateConfig creates a new gateway configuration
func (s *GatewayService) CreateConfig(ctx context.Context, projectID uuid.UUID, input *domain.GatewayConfigInput) (*domain.GatewayConfig, error) {
	if input.Name == "" {
		return nil, fmt.Errorf("gateway config name is required")
	}
	if len(input.Providers) == 0 {
		return nil, fmt.Errorf("at least one provider is required")
	}

	tracingEnabled := true
	if input.TracingEnabled != nil {
		tracingEnabled = *input.TracingEnabled
	}

	rateLimitRPM := input.RateLimitRPM
	if rateLimitRPM <= 0 {
		rateLimitRPM = 1000
	}

	timeoutSec := input.TimeoutSeconds
	if timeoutSec <= 0 {
		timeoutSec = 30
	}

	maxRetries := input.MaxRetries
	if maxRetries <= 0 {
		maxRetries = 2
	}

	config := &domain.GatewayConfig{
		ID:             uuid.New(),
		ProjectID:      projectID,
		Name:           input.Name,
		Strategy:       input.Strategy,
		Providers:      input.Providers,
		FallbackChain:  input.FallbackChain,
		RateLimitRPM:   rateLimitRPM,
		RateLimitTPM:   input.RateLimitTPM,
		MaxRetries:     maxRetries,
		TimeoutSeconds: timeoutSec,
		TracingEnabled: tracingEnabled,
		Enabled:        true,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	s.mu.Lock()
	s.configs[config.ID] = config
	s.mu.Unlock()

	s.logger.Info("gateway config created",
		zap.String("configId", config.ID.String()),
		zap.String("name", config.Name),
		zap.String("strategy", string(config.Strategy)),
		zap.Int("providers", len(config.Providers)),
	)

	return config, nil
}

// GetConfig retrieves a gateway configuration by ID
func (s *GatewayService) GetConfig(ctx context.Context, id uuid.UUID) (*domain.GatewayConfig, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	config, exists := s.configs[id]
	if !exists {
		return nil, fmt.Errorf("gateway config not found")
	}
	return config, nil
}

// ListConfigs lists gateway configurations for a project
func (s *GatewayService) ListConfigs(ctx context.Context, projectID uuid.UUID) ([]domain.GatewayConfig, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var configs []domain.GatewayConfig
	for _, cfg := range s.configs {
		if cfg.ProjectID == projectID {
			configs = append(configs, *cfg)
		}
	}

	sort.Slice(configs, func(i, j int) bool {
		return configs[i].CreatedAt.After(configs[j].CreatedAt)
	})

	return configs, nil
}

// UpdateConfig updates an existing gateway configuration
func (s *GatewayService) UpdateConfig(ctx context.Context, id uuid.UUID, input *domain.GatewayConfigInput) (*domain.GatewayConfig, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	config, exists := s.configs[id]
	if !exists {
		return nil, fmt.Errorf("gateway config not found")
	}

	if input.Name != "" {
		config.Name = input.Name
	}
	if input.Strategy != "" {
		config.Strategy = input.Strategy
	}
	if len(input.Providers) > 0 {
		config.Providers = input.Providers
	}
	if len(input.FallbackChain) > 0 {
		config.FallbackChain = input.FallbackChain
	}
	if input.RateLimitRPM > 0 {
		config.RateLimitRPM = input.RateLimitRPM
	}
	if input.MaxRetries > 0 {
		config.MaxRetries = input.MaxRetries
	}
	if input.TimeoutSeconds > 0 {
		config.TimeoutSeconds = input.TimeoutSeconds
	}
	if input.TracingEnabled != nil {
		config.TracingEnabled = *input.TracingEnabled
	}
	config.UpdatedAt = time.Now()

	return config, nil
}

// DeleteConfig deletes a gateway configuration
func (s *GatewayService) DeleteConfig(ctx context.Context, id uuid.UUID) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.configs[id]; !exists {
		return fmt.Errorf("gateway config not found")
	}
	delete(s.configs, id)
	delete(s.rules, id)
	return nil
}

// AddRoutingRule adds a smart routing rule to a configuration
func (s *GatewayService) AddRoutingRule(ctx context.Context, configID uuid.UUID, rule *domain.RoutingRule) (*domain.RoutingRule, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.configs[configID]; !exists {
		return nil, fmt.Errorf("gateway config not found")
	}

	rule.ID = uuid.New()
	rule.ConfigID = configID
	rule.Enabled = true
	rule.CreatedAt = time.Now()

	s.rules[configID] = append(s.rules[configID], *rule)

	// Sort rules by priority
	sort.Slice(s.rules[configID], func(i, j int) bool {
		return s.rules[configID][i].Priority < s.rules[configID][j].Priority
	})

	s.logger.Info("routing rule added",
		zap.String("ruleId", rule.ID.String()),
		zap.String("configId", configID.String()),
		zap.String("name", rule.Name),
	)

	return rule, nil
}

// ListRoutingRules lists routing rules for a configuration
func (s *GatewayService) ListRoutingRules(ctx context.Context, configID uuid.UUID) ([]domain.RoutingRule, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rules, exists := s.rules[configID]
	if !exists {
		return []domain.RoutingRule{}, nil
	}
	return rules, nil
}

// DeleteRoutingRule deletes a routing rule
func (s *GatewayService) DeleteRoutingRule(ctx context.Context, configID, ruleID uuid.UUID) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	rules, exists := s.rules[configID]
	if !exists {
		return fmt.Errorf("config not found")
	}

	for i, rule := range rules {
		if rule.ID == ruleID {
			s.rules[configID] = append(rules[:i], rules[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("rule not found")
}

// ProxyRequest routes and proxies an LLM API request through the gateway
func (s *GatewayService) ProxyRequest(ctx context.Context, configID uuid.UUID, req *domain.GatewayRequest) (*domain.GatewayResponse, error) {
	s.mu.RLock()
	config, exists := s.configs[configID]
	if !exists {
		s.mu.RUnlock()
		return nil, fmt.Errorf("gateway config not found")
	}
	rules := s.rules[configID]
	s.mu.RUnlock()

	if !config.Enabled {
		return nil, fmt.Errorf("gateway config is disabled")
	}

	start := time.Now()
	s.reqCount.Add(1)

	// Resolve provider using routing strategy and rules
	provider, model := s.resolveProvider(config, rules, req)

	// Simulate proxied LLM call (real implementation would make HTTP calls)
	response := &domain.GatewayResponse{
		ID:       uuid.New().String(),
		Provider: provider,
		Model:    model,
		Choices: []domain.GatewayChoice{
			{
				Index:        0,
				Message:      domain.GatewayMessage{Role: "assistant", Content: "Response from " + string(provider) + "/" + model},
				FinishReason: "stop",
			},
		},
		Usage: domain.GatewayUsage{
			PromptTokens:     estimateTokens(req.Messages),
			CompletionTokens: 150,
			TotalTokens:      estimateTokens(req.Messages) + 150,
		},
		LatencyMs:     time.Since(start).Milliseconds(),
		FallbackUsed:  provider != req.Provider,
		EstimatedCost: s.estimateCost(provider, estimateTokens(req.Messages)+150),
	}

	if response.FallbackUsed {
		response.OriginalProvider = req.Provider
	}

	if config.TracingEnabled {
		response.TraceID = uuid.New().String()
	}

	// Update stats
	s.stats.mu.Lock()
	s.stats.totalRequests++
	s.stats.totalTokens += int64(response.Usage.TotalTokens)
	s.stats.totalCost += response.EstimatedCost
	s.stats.totalLatency += response.LatencyMs
	s.stats.providerCounts[string(provider)]++
	if response.FallbackUsed {
		s.stats.fallbackCount++
	}
	s.stats.mu.Unlock()

	s.logger.Info("gateway request proxied",
		zap.String("provider", string(provider)),
		zap.String("model", model),
		zap.Int64("latencyMs", response.LatencyMs),
		zap.Float64("cost", response.EstimatedCost),
	)

	return response, nil
}

// GetStats returns gateway usage statistics
func (s *GatewayService) GetStats(ctx context.Context) *domain.GatewayStats {
	s.stats.mu.Lock()
	defer s.stats.mu.Unlock()

	var avgLatency float64
	if s.stats.totalRequests > 0 {
		avgLatency = float64(s.stats.totalLatency) / float64(s.stats.totalRequests)
	}

	var errorRate float64
	if s.stats.totalRequests > 0 {
		errorRate = float64(s.stats.errorCount) / float64(s.stats.totalRequests)
	}

	var fallbackRate float64
	if s.stats.totalRequests > 0 {
		fallbackRate = float64(s.stats.fallbackCount) / float64(s.stats.totalRequests)
	}

	providerBreakdown := make(map[string]int64)
	for k, v := range s.stats.providerCounts {
		providerBreakdown[k] = v
	}

	return &domain.GatewayStats{
		TotalRequests:     s.stats.totalRequests,
		TotalTokens:       s.stats.totalTokens,
		TotalCost:         math.Round(s.stats.totalCost*10000) / 10000,
		AvgLatencyMs:      math.Round(avgLatency*100) / 100,
		ErrorRate:         errorRate,
		ProviderBreakdown: providerBreakdown,
		FallbackRate:      fallbackRate,
	}
}

func (s *GatewayService) resolveProvider(config *domain.GatewayConfig, rules []domain.RoutingRule, req *domain.GatewayRequest) (domain.GatewayProvider, string) {
	// Check routing rules first
	for _, rule := range rules {
		if !rule.Enabled {
			continue
		}
		if s.matchesRule(rule, req) {
			if rule.Action.RouteToProvider != "" {
				model := rule.Action.RouteToModel
				if model == "" {
					model = req.Model
				}
				return rule.Action.RouteToProvider, model
			}
		}
	}

	// Apply routing strategy
	switch config.Strategy {
	case domain.RoutingCheapest:
		return s.resolveCheapest(config, req)
	case domain.RoutingFastest:
		return s.resolveFastest(config, req)
	case domain.RoutingFallback:
		return s.resolveFallback(config, req)
	case domain.RoutingRoundRobin:
		return s.resolveRoundRobin(config, req)
	default:
		return s.resolvePriority(config, req)
	}
}

func (s *GatewayService) matchesRule(rule domain.RoutingRule, req *domain.GatewayRequest) bool {
	cond := rule.Condition
	tokenCount := estimateTokens(req.Messages)

	if cond.MaxTokens != nil && tokenCount > *cond.MaxTokens {
		return false
	}
	if cond.MinTokens != nil && tokenCount < *cond.MinTokens {
		return false
	}
	return true
}

func (s *GatewayService) resolveCheapest(config *domain.GatewayConfig, req *domain.GatewayRequest) (domain.GatewayProvider, string) {
	var cheapest *domain.ProviderConfig
	for i, p := range config.Providers {
		if !p.Enabled {
			continue
		}
		if cheapest == nil || p.CostPer1K < cheapest.CostPer1K {
			cheapest = &config.Providers[i]
		}
	}
	if cheapest != nil {
		model := req.Model
		if len(cheapest.Models) > 0 {
			model = cheapest.Models[0]
		}
		return cheapest.Provider, model
	}
	return req.Provider, req.Model
}

func (s *GatewayService) resolveFastest(config *domain.GatewayConfig, req *domain.GatewayRequest) (domain.GatewayProvider, string) {
	// Priority-based selection favoring local/lower-latency providers
	for _, p := range config.Providers {
		if !p.Enabled {
			continue
		}
		if p.Provider == domain.ProviderLocal {
			model := req.Model
			if len(p.Models) > 0 {
				model = p.Models[0]
			}
			return p.Provider, model
		}
	}
	return s.resolvePriority(config, req)
}

func (s *GatewayService) resolveFallback(config *domain.GatewayConfig, req *domain.GatewayRequest) (domain.GatewayProvider, string) {
	for _, providerName := range config.FallbackChain {
		for _, p := range config.Providers {
			if string(p.Provider) == providerName && p.Enabled {
				model := req.Model
				if len(p.Models) > 0 {
					model = p.Models[0]
				}
				return p.Provider, model
			}
		}
	}
	return s.resolvePriority(config, req)
}

func (s *GatewayService) resolveRoundRobin(config *domain.GatewayConfig, req *domain.GatewayRequest) (domain.GatewayProvider, string) {
	var enabledProviders []domain.ProviderConfig
	for _, p := range config.Providers {
		if p.Enabled {
			enabledProviders = append(enabledProviders, p)
		}
	}
	if len(enabledProviders) == 0 {
		return req.Provider, req.Model
	}

	idx := int(s.reqCount.Load()) % len(enabledProviders)
	p := enabledProviders[idx]
	model := req.Model
	if len(p.Models) > 0 {
		model = p.Models[0]
	}
	return p.Provider, model
}

func (s *GatewayService) resolvePriority(config *domain.GatewayConfig, req *domain.GatewayRequest) (domain.GatewayProvider, string) {
	sorted := make([]domain.ProviderConfig, len(config.Providers))
	copy(sorted, config.Providers)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Priority < sorted[j].Priority })

	for _, p := range sorted {
		if p.Enabled {
			model := req.Model
			if len(p.Models) > 0 {
				model = p.Models[0]
			}
			return p.Provider, model
		}
	}
	return req.Provider, req.Model
}

func (s *GatewayService) estimateCost(provider domain.GatewayProvider, totalTokens int) float64 {
	costPer1K := map[domain.GatewayProvider]float64{
		domain.ProviderOpenAI:    0.03,
		domain.ProviderAnthropic: 0.025,
		domain.ProviderGoogle:    0.02,
		domain.ProviderLocal:     0.0,
	}
	rate, ok := costPer1K[provider]
	if !ok {
		rate = 0.03
	}
	return float64(totalTokens) / 1000.0 * rate
}

func estimateTokens(messages []domain.GatewayMessage) int {
	total := 0
	for _, msg := range messages {
		total += len(msg.Content) / 4 // rough estimate: ~4 chars per token
	}
	if total == 0 {
		total = 10
	}
	return total
}
