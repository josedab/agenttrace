package service

import (
	"context"
	"strings"
	"sync"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

// ModelPricing represents pricing for a model
type ModelPricing struct {
	Model            string  `json:"model"`
	Provider         string  `json:"provider"`
	InputPricePer1K  float64 `json:"inputPricePer1K"`
	OutputPricePer1K float64 `json:"outputPricePer1K"`
	CachePricePer1K  float64 `json:"cachePricePer1K,omitempty"`
}

// CostService handles cost calculation for LLM usage
type CostService struct {
	mu        sync.RWMutex
	pricing   map[string]*ModelPricing
	overrides map[uuid.UUID]map[string]*ModelPricing // project-specific overrides
	logger    *zap.Logger
}

// NewCostService creates a new cost service with default pricing
func NewCostService(logger *zap.Logger) *CostService {
	s := &CostService{
		pricing:   make(map[string]*ModelPricing),
		overrides: make(map[uuid.UUID]map[string]*ModelPricing),
		logger:    logger,
	}
	s.loadDefaultPricing()
	return s
}

// CalculateCost calculates the cost for a generation
func (s *CostService) CalculateCost(ctx context.Context, projectID uuid.UUID, model string, inputTokens, outputTokens int64) (*domain.CostDetails, error) {
	pricing := s.getPricing(projectID, model)
	if pricing == nil {
		return nil, nil // No pricing available
	}

	inputCost := float64(inputTokens) * pricing.InputPricePer1K / 1000
	outputCost := float64(outputTokens) * pricing.OutputPricePer1K / 1000
	totalCost := inputCost + outputCost

	return &domain.CostDetails{
		InputCost:  inputCost,
		OutputCost: outputCost,
		TotalCost:  totalCost,
		Currency:   "USD",
	}, nil
}

// CalculateCostWithCache calculates the cost including cache tokens
func (s *CostService) CalculateCostWithCache(ctx context.Context, projectID uuid.UUID, model string, inputTokens, outputTokens, cacheTokens int64) (*domain.CostDetails, error) {
	pricing := s.getPricing(projectID, model)
	if pricing == nil {
		return nil, nil
	}

	inputCost := float64(inputTokens) * pricing.InputPricePer1K / 1000
	outputCost := float64(outputTokens) * pricing.OutputPricePer1K / 1000
	cacheCost := float64(cacheTokens) * pricing.CachePricePer1K / 1000
	totalCost := inputCost + outputCost + cacheCost

	return &domain.CostDetails{
		InputCost:  inputCost,
		OutputCost: outputCost,
		TotalCost:  totalCost,
		Currency:   "USD",
	}, nil
}

// GetPricing returns pricing for a model
func (s *CostService) GetPricing(model string) *ModelPricing {
	return s.getPricing(uuid.Nil, model)
}

// SetProjectPricing sets custom pricing for a project
func (s *CostService) SetProjectPricing(projectID uuid.UUID, pricing *ModelPricing) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.overrides[projectID] == nil {
		s.overrides[projectID] = make(map[string]*ModelPricing)
	}
	s.overrides[projectID][normalizeModel(pricing.Model)] = pricing
}

// ListModels returns all available models with pricing
func (s *CostService) ListModels() []*ModelPricing {
	s.mu.RLock()
	defer s.mu.RUnlock()

	models := make([]*ModelPricing, 0, len(s.pricing))
	for _, p := range s.pricing {
		models = append(models, p)
	}
	return models
}

// getPricing retrieves pricing for a model, checking project overrides first
func (s *CostService) getPricing(projectID uuid.UUID, model string) *ModelPricing {
	s.mu.RLock()
	defer s.mu.RUnlock()

	normalizedModel := normalizeModel(model)

	// Check project overrides first
	if projectID != uuid.Nil {
		if overrides, ok := s.overrides[projectID]; ok {
			if pricing, ok := overrides[normalizedModel]; ok {
				return pricing
			}
		}
	}

	// Check default pricing
	if pricing, ok := s.pricing[normalizedModel]; ok {
		return pricing
	}

	// Try partial matches (for versioned models)
	for key, pricing := range s.pricing {
		if strings.HasPrefix(normalizedModel, key) || strings.HasPrefix(key, normalizedModel) {
			return pricing
		}
	}

	return nil
}

// normalizeModel normalizes a model name for lookup
func normalizeModel(model string) string {
	return strings.ToLower(strings.TrimSpace(model))
}

// loadDefaultPricing loads default pricing for all supported models
func (s *CostService) loadDefaultPricing() {
	s.mu.Lock()
	defer s.mu.Unlock()

	// OpenAI Models
	openaiModels := []ModelPricing{
		// GPT-4o
		{Model: "gpt-4o", Provider: "openai", InputPricePer1K: 0.0025, OutputPricePer1K: 0.010},
		{Model: "gpt-4o-2024-11-20", Provider: "openai", InputPricePer1K: 0.0025, OutputPricePer1K: 0.010},
		{Model: "gpt-4o-2024-08-06", Provider: "openai", InputPricePer1K: 0.0025, OutputPricePer1K: 0.010},
		{Model: "gpt-4o-2024-05-13", Provider: "openai", InputPricePer1K: 0.005, OutputPricePer1K: 0.015},
		{Model: "gpt-4o-audio-preview", Provider: "openai", InputPricePer1K: 0.0025, OutputPricePer1K: 0.010},
		{Model: "gpt-4o-audio-preview-2024-12-17", Provider: "openai", InputPricePer1K: 0.0025, OutputPricePer1K: 0.010},
		{Model: "gpt-4o-mini", Provider: "openai", InputPricePer1K: 0.00015, OutputPricePer1K: 0.0006},
		{Model: "gpt-4o-mini-2024-07-18", Provider: "openai", InputPricePer1K: 0.00015, OutputPricePer1K: 0.0006},
		{Model: "gpt-4o-mini-audio-preview", Provider: "openai", InputPricePer1K: 0.00015, OutputPricePer1K: 0.0006},
		{Model: "gpt-4o-mini-audio-preview-2024-12-17", Provider: "openai", InputPricePer1K: 0.00015, OutputPricePer1K: 0.0006},

		// o1
		{Model: "o1", Provider: "openai", InputPricePer1K: 0.015, OutputPricePer1K: 0.060},
		{Model: "o1-2024-12-17", Provider: "openai", InputPricePer1K: 0.015, OutputPricePer1K: 0.060},
		{Model: "o1-preview", Provider: "openai", InputPricePer1K: 0.015, OutputPricePer1K: 0.060},
		{Model: "o1-preview-2024-09-12", Provider: "openai", InputPricePer1K: 0.015, OutputPricePer1K: 0.060},
		{Model: "o1-mini", Provider: "openai", InputPricePer1K: 0.003, OutputPricePer1K: 0.012},
		{Model: "o1-mini-2024-09-12", Provider: "openai", InputPricePer1K: 0.003, OutputPricePer1K: 0.012},
		{Model: "o3-mini", Provider: "openai", InputPricePer1K: 0.0011, OutputPricePer1K: 0.0044},
		{Model: "o3-mini-2025-01-31", Provider: "openai", InputPricePer1K: 0.0011, OutputPricePer1K: 0.0044},

		// GPT-4 Turbo
		{Model: "gpt-4-turbo", Provider: "openai", InputPricePer1K: 0.010, OutputPricePer1K: 0.030},
		{Model: "gpt-4-turbo-2024-04-09", Provider: "openai", InputPricePer1K: 0.010, OutputPricePer1K: 0.030},
		{Model: "gpt-4-turbo-preview", Provider: "openai", InputPricePer1K: 0.010, OutputPricePer1K: 0.030},
		{Model: "gpt-4-0125-preview", Provider: "openai", InputPricePer1K: 0.010, OutputPricePer1K: 0.030},
		{Model: "gpt-4-1106-preview", Provider: "openai", InputPricePer1K: 0.010, OutputPricePer1K: 0.030},
		{Model: "gpt-4-vision-preview", Provider: "openai", InputPricePer1K: 0.010, OutputPricePer1K: 0.030},

		// GPT-4
		{Model: "gpt-4", Provider: "openai", InputPricePer1K: 0.030, OutputPricePer1K: 0.060},
		{Model: "gpt-4-0613", Provider: "openai", InputPricePer1K: 0.030, OutputPricePer1K: 0.060},
		{Model: "gpt-4-32k", Provider: "openai", InputPricePer1K: 0.060, OutputPricePer1K: 0.120},
		{Model: "gpt-4-32k-0613", Provider: "openai", InputPricePer1K: 0.060, OutputPricePer1K: 0.120},

		// GPT-3.5
		{Model: "gpt-3.5-turbo", Provider: "openai", InputPricePer1K: 0.0005, OutputPricePer1K: 0.0015},
		{Model: "gpt-3.5-turbo-0125", Provider: "openai", InputPricePer1K: 0.0005, OutputPricePer1K: 0.0015},
		{Model: "gpt-3.5-turbo-1106", Provider: "openai", InputPricePer1K: 0.001, OutputPricePer1K: 0.002},
		{Model: "gpt-3.5-turbo-instruct", Provider: "openai", InputPricePer1K: 0.0015, OutputPricePer1K: 0.002},
		{Model: "gpt-3.5-turbo-16k", Provider: "openai", InputPricePer1K: 0.003, OutputPricePer1K: 0.004},

		// Embedding models
		{Model: "text-embedding-3-small", Provider: "openai", InputPricePer1K: 0.00002, OutputPricePer1K: 0},
		{Model: "text-embedding-3-large", Provider: "openai", InputPricePer1K: 0.00013, OutputPricePer1K: 0},
		{Model: "text-embedding-ada-002", Provider: "openai", InputPricePer1K: 0.0001, OutputPricePer1K: 0},
	}

	// Anthropic Models
	anthropicModels := []ModelPricing{
		// Claude 4
		{Model: "claude-opus-4-20250514", Provider: "anthropic", InputPricePer1K: 0.015, OutputPricePer1K: 0.075},
		{Model: "claude-sonnet-4-20250514", Provider: "anthropic", InputPricePer1K: 0.003, OutputPricePer1K: 0.015},

		// Claude 3.5
		{Model: "claude-3-5-sonnet-20241022", Provider: "anthropic", InputPricePer1K: 0.003, OutputPricePer1K: 0.015},
		{Model: "claude-3-5-sonnet-20240620", Provider: "anthropic", InputPricePer1K: 0.003, OutputPricePer1K: 0.015},
		{Model: "claude-3-5-haiku-20241022", Provider: "anthropic", InputPricePer1K: 0.001, OutputPricePer1K: 0.005},

		// Claude 3
		{Model: "claude-3-opus-20240229", Provider: "anthropic", InputPricePer1K: 0.015, OutputPricePer1K: 0.075},
		{Model: "claude-3-sonnet-20240229", Provider: "anthropic", InputPricePer1K: 0.003, OutputPricePer1K: 0.015},
		{Model: "claude-3-haiku-20240307", Provider: "anthropic", InputPricePer1K: 0.00025, OutputPricePer1K: 0.00125},

		// Claude 2
		{Model: "claude-2.1", Provider: "anthropic", InputPricePer1K: 0.008, OutputPricePer1K: 0.024},
		{Model: "claude-2.0", Provider: "anthropic", InputPricePer1K: 0.008, OutputPricePer1K: 0.024},
		{Model: "claude-instant-1.2", Provider: "anthropic", InputPricePer1K: 0.0008, OutputPricePer1K: 0.0024},
	}

	// Google Models
	googleModels := []ModelPricing{
		// Gemini 2.0
		{Model: "gemini-2.0-flash-exp", Provider: "google", InputPricePer1K: 0.0, OutputPricePer1K: 0.0},
		{Model: "gemini-2.0-flash-thinking-exp", Provider: "google", InputPricePer1K: 0.0, OutputPricePer1K: 0.0},

		// Gemini 1.5
		{Model: "gemini-1.5-pro", Provider: "google", InputPricePer1K: 0.00125, OutputPricePer1K: 0.005},
		{Model: "gemini-1.5-pro-latest", Provider: "google", InputPricePer1K: 0.00125, OutputPricePer1K: 0.005},
		{Model: "gemini-1.5-flash", Provider: "google", InputPricePer1K: 0.000075, OutputPricePer1K: 0.0003},
		{Model: "gemini-1.5-flash-latest", Provider: "google", InputPricePer1K: 0.000075, OutputPricePer1K: 0.0003},
		{Model: "gemini-1.5-flash-8b", Provider: "google", InputPricePer1K: 0.0000375, OutputPricePer1K: 0.00015},

		// Gemini 1.0
		{Model: "gemini-1.0-pro", Provider: "google", InputPricePer1K: 0.0005, OutputPricePer1K: 0.0015},
		{Model: "gemini-pro", Provider: "google", InputPricePer1K: 0.0005, OutputPricePer1K: 0.0015},
		{Model: "gemini-pro-vision", Provider: "google", InputPricePer1K: 0.0005, OutputPricePer1K: 0.0015},
	}

	// Mistral Models
	mistralModels := []ModelPricing{
		{Model: "mistral-large-latest", Provider: "mistral", InputPricePer1K: 0.002, OutputPricePer1K: 0.006},
		{Model: "mistral-large-2411", Provider: "mistral", InputPricePer1K: 0.002, OutputPricePer1K: 0.006},
		{Model: "mistral-large-2407", Provider: "mistral", InputPricePer1K: 0.002, OutputPricePer1K: 0.006},
		{Model: "mistral-medium-latest", Provider: "mistral", InputPricePer1K: 0.00275, OutputPricePer1K: 0.0081},
		{Model: "mistral-small-latest", Provider: "mistral", InputPricePer1K: 0.0002, OutputPricePer1K: 0.0006},
		{Model: "mistral-small-2409", Provider: "mistral", InputPricePer1K: 0.0002, OutputPricePer1K: 0.0006},
		{Model: "open-mistral-7b", Provider: "mistral", InputPricePer1K: 0.00025, OutputPricePer1K: 0.00025},
		{Model: "open-mixtral-8x7b", Provider: "mistral", InputPricePer1K: 0.0007, OutputPricePer1K: 0.0007},
		{Model: "open-mixtral-8x22b", Provider: "mistral", InputPricePer1K: 0.002, OutputPricePer1K: 0.006},
		{Model: "codestral-latest", Provider: "mistral", InputPricePer1K: 0.0002, OutputPricePer1K: 0.0006},
		{Model: "codestral-2405", Provider: "mistral", InputPricePer1K: 0.001, OutputPricePer1K: 0.003},
		{Model: "pixtral-12b-2409", Provider: "mistral", InputPricePer1K: 0.00015, OutputPricePer1K: 0.00015},
		{Model: "pixtral-large-latest", Provider: "mistral", InputPricePer1K: 0.002, OutputPricePer1K: 0.006},
	}

	// Cohere Models
	cohereModels := []ModelPricing{
		{Model: "command-r-plus", Provider: "cohere", InputPricePer1K: 0.002, OutputPricePer1K: 0.01},
		{Model: "command-r-plus-08-2024", Provider: "cohere", InputPricePer1K: 0.002, OutputPricePer1K: 0.01},
		{Model: "command-r", Provider: "cohere", InputPricePer1K: 0.00015, OutputPricePer1K: 0.0006},
		{Model: "command-r-08-2024", Provider: "cohere", InputPricePer1K: 0.000375, OutputPricePer1K: 0.0015},
		{Model: "command", Provider: "cohere", InputPricePer1K: 0.001, OutputPricePer1K: 0.002},
		{Model: "command-light", Provider: "cohere", InputPricePer1K: 0.0003, OutputPricePer1K: 0.0006},
		{Model: "embed-english-v3.0", Provider: "cohere", InputPricePer1K: 0.0001, OutputPricePer1K: 0},
		{Model: "embed-multilingual-v3.0", Provider: "cohere", InputPricePer1K: 0.0001, OutputPricePer1K: 0},
	}

	// Meta/Llama Models (via cloud providers)
	llamaModels := []ModelPricing{
		{Model: "llama-3.3-70b-versatile", Provider: "groq", InputPricePer1K: 0.00059, OutputPricePer1K: 0.00079},
		{Model: "llama-3.2-90b-vision-preview", Provider: "groq", InputPricePer1K: 0.0009, OutputPricePer1K: 0.0009},
		{Model: "llama-3.2-11b-vision-preview", Provider: "groq", InputPricePer1K: 0.00018, OutputPricePer1K: 0.00018},
		{Model: "llama-3.2-3b-preview", Provider: "groq", InputPricePer1K: 0.00006, OutputPricePer1K: 0.00006},
		{Model: "llama-3.2-1b-preview", Provider: "groq", InputPricePer1K: 0.00004, OutputPricePer1K: 0.00004},
		{Model: "llama-3.1-405b", Provider: "groq", InputPricePer1K: 0.005, OutputPricePer1K: 0.015},
		{Model: "llama-3.1-70b-versatile", Provider: "groq", InputPricePer1K: 0.00059, OutputPricePer1K: 0.00079},
		{Model: "llama-3.1-8b-instant", Provider: "groq", InputPricePer1K: 0.00005, OutputPricePer1K: 0.00008},
		{Model: "llama3-70b-8192", Provider: "groq", InputPricePer1K: 0.00059, OutputPricePer1K: 0.00079},
		{Model: "llama3-8b-8192", Provider: "groq", InputPricePer1K: 0.00005, OutputPricePer1K: 0.00008},
		{Model: "mixtral-8x7b-32768", Provider: "groq", InputPricePer1K: 0.00024, OutputPricePer1K: 0.00024},
		{Model: "gemma2-9b-it", Provider: "groq", InputPricePer1K: 0.0002, OutputPricePer1K: 0.0002},
	}

	// AWS Bedrock Models
	bedrockModels := []ModelPricing{
		{Model: "amazon.titan-text-express-v1", Provider: "bedrock", InputPricePer1K: 0.0002, OutputPricePer1K: 0.0006},
		{Model: "amazon.titan-text-lite-v1", Provider: "bedrock", InputPricePer1K: 0.00015, OutputPricePer1K: 0.0002},
		{Model: "amazon.titan-text-premier-v1:0", Provider: "bedrock", InputPricePer1K: 0.0005, OutputPricePer1K: 0.0015},
		{Model: "amazon.titan-embed-text-v1", Provider: "bedrock", InputPricePer1K: 0.0001, OutputPricePer1K: 0},
		{Model: "amazon.titan-embed-text-v2:0", Provider: "bedrock", InputPricePer1K: 0.00002, OutputPricePer1K: 0},
		{Model: "amazon.nova-pro-v1:0", Provider: "bedrock", InputPricePer1K: 0.0008, OutputPricePer1K: 0.0032},
		{Model: "amazon.nova-lite-v1:0", Provider: "bedrock", InputPricePer1K: 0.00006, OutputPricePer1K: 0.00024},
		{Model: "amazon.nova-micro-v1:0", Provider: "bedrock", InputPricePer1K: 0.000035, OutputPricePer1K: 0.00014},
		{Model: "ai21.jamba-1-5-large-v1:0", Provider: "bedrock", InputPricePer1K: 0.002, OutputPricePer1K: 0.008},
		{Model: "ai21.jamba-1-5-mini-v1:0", Provider: "bedrock", InputPricePer1K: 0.0002, OutputPricePer1K: 0.0004},
		{Model: "meta.llama3-70b-instruct-v1:0", Provider: "bedrock", InputPricePer1K: 0.00265, OutputPricePer1K: 0.0035},
		{Model: "meta.llama3-8b-instruct-v1:0", Provider: "bedrock", InputPricePer1K: 0.0003, OutputPricePer1K: 0.0006},
		{Model: "cohere.command-r-plus-v1:0", Provider: "bedrock", InputPricePer1K: 0.003, OutputPricePer1K: 0.015},
		{Model: "cohere.command-r-v1:0", Provider: "bedrock", InputPricePer1K: 0.0005, OutputPricePer1K: 0.0015},
	}

	// Azure OpenAI Models
	azureModels := []ModelPricing{
		{Model: "gpt-4o", Provider: "azure", InputPricePer1K: 0.0025, OutputPricePer1K: 0.010},
		{Model: "gpt-4o-mini", Provider: "azure", InputPricePer1K: 0.00015, OutputPricePer1K: 0.0006},
		{Model: "gpt-4-turbo", Provider: "azure", InputPricePer1K: 0.010, OutputPricePer1K: 0.030},
		{Model: "gpt-4", Provider: "azure", InputPricePer1K: 0.030, OutputPricePer1K: 0.060},
		{Model: "gpt-35-turbo", Provider: "azure", InputPricePer1K: 0.0005, OutputPricePer1K: 0.0015},
	}

	// DeepSeek Models
	deepseekModels := []ModelPricing{
		{Model: "deepseek-chat", Provider: "deepseek", InputPricePer1K: 0.00014, OutputPricePer1K: 0.00028},
		{Model: "deepseek-coder", Provider: "deepseek", InputPricePer1K: 0.00014, OutputPricePer1K: 0.00028},
		{Model: "deepseek-reasoner", Provider: "deepseek", InputPricePer1K: 0.00055, OutputPricePer1K: 0.00219},
	}

	// xAI Models
	xaiModels := []ModelPricing{
		{Model: "grok-2", Provider: "xai", InputPricePer1K: 0.002, OutputPricePer1K: 0.010},
		{Model: "grok-2-mini", Provider: "xai", InputPricePer1K: 0.0002, OutputPricePer1K: 0.001},
		{Model: "grok-beta", Provider: "xai", InputPricePer1K: 0.005, OutputPricePer1K: 0.015},
		{Model: "grok-vision-beta", Provider: "xai", InputPricePer1K: 0.005, OutputPricePer1K: 0.015},
	}

	// Add all models to pricing map
	allModels := make([]ModelPricing, 0)
	allModels = append(allModels, openaiModels...)
	allModels = append(allModels, anthropicModels...)
	allModels = append(allModels, googleModels...)
	allModels = append(allModels, mistralModels...)
	allModels = append(allModels, cohereModels...)
	allModels = append(allModels, llamaModels...)
	allModels = append(allModels, bedrockModels...)
	allModels = append(allModels, azureModels...)
	allModels = append(allModels, deepseekModels...)
	allModels = append(allModels, xaiModels...)

	for i := range allModels {
		m := &allModels[i]
		s.pricing[normalizeModel(m.Model)] = m
	}

	if s.logger != nil {
		s.logger.Info("loaded pricing for models", zap.Int("model_count", len(s.pricing)))
	}
}
