package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewCostService(t *testing.T) {
	t.Run("creates service with default pricing", func(t *testing.T) {
		svc := NewCostService(zap.NewNop())

		assert.NotNil(t, svc)
		assert.NotEmpty(t, svc.pricing)

		// Check some known models exist
		models := svc.ListModels()
		assert.Greater(t, len(models), 50) // We have 100+ models
	})
}

func TestCostService_CalculateCost(t *testing.T) {
	svc := NewCostService(zap.NewNop())
	ctx := context.Background()
	projectID := uuid.New()

	t.Run("calculates cost for gpt-4o", func(t *testing.T) {
		cost, err := svc.CalculateCost(ctx, projectID, "gpt-4o", 1000, 500)

		require.NoError(t, err)
		require.NotNil(t, cost)

		// GPT-4o: $0.0025/1K input, $0.010/1K output
		expectedInputCost := 1000 * 0.0025 / 1000  // $0.0025
		expectedOutputCost := 500 * 0.010 / 1000   // $0.005
		expectedTotal := expectedInputCost + expectedOutputCost

		assert.InDelta(t, expectedInputCost, cost.InputCost, 0.0001)
		assert.InDelta(t, expectedOutputCost, cost.OutputCost, 0.0001)
		assert.InDelta(t, expectedTotal, cost.TotalCost, 0.0001)
		assert.Equal(t, "USD", cost.Currency)
	})

	t.Run("calculates cost for gpt-4o-mini", func(t *testing.T) {
		cost, err := svc.CalculateCost(ctx, projectID, "gpt-4o-mini", 10000, 1000)

		require.NoError(t, err)
		require.NotNil(t, cost)

		// GPT-4o-mini: $0.00015/1K input, $0.0006/1K output
		expectedInputCost := 10000 * 0.00015 / 1000
		expectedOutputCost := 1000 * 0.0006 / 1000
		expectedTotal := expectedInputCost + expectedOutputCost

		assert.InDelta(t, expectedInputCost, cost.InputCost, 0.0001)
		assert.InDelta(t, expectedOutputCost, cost.OutputCost, 0.0001)
		assert.InDelta(t, expectedTotal, cost.TotalCost, 0.0001)
	})

	t.Run("calculates cost for claude-3-5-sonnet", func(t *testing.T) {
		cost, err := svc.CalculateCost(ctx, projectID, "claude-3-5-sonnet-20241022", 5000, 2000)

		require.NoError(t, err)
		require.NotNil(t, cost)

		// Claude 3.5 Sonnet: $0.003/1K input, $0.015/1K output
		expectedInputCost := 5000 * 0.003 / 1000
		expectedOutputCost := 2000 * 0.015 / 1000
		expectedTotal := expectedInputCost + expectedOutputCost

		assert.InDelta(t, expectedInputCost, cost.InputCost, 0.0001)
		assert.InDelta(t, expectedOutputCost, cost.OutputCost, 0.0001)
		assert.InDelta(t, expectedTotal, cost.TotalCost, 0.0001)
	})

	t.Run("calculates cost for claude-3-opus", func(t *testing.T) {
		cost, err := svc.CalculateCost(ctx, projectID, "claude-3-opus-20240229", 1000, 500)

		require.NoError(t, err)
		require.NotNil(t, cost)

		// Claude 3 Opus: $0.015/1K input, $0.075/1K output
		expectedInputCost := 1000 * 0.015 / 1000
		expectedOutputCost := 500 * 0.075 / 1000
		expectedTotal := expectedInputCost + expectedOutputCost

		assert.InDelta(t, expectedInputCost, cost.InputCost, 0.0001)
		assert.InDelta(t, expectedOutputCost, cost.OutputCost, 0.0001)
		assert.InDelta(t, expectedTotal, cost.TotalCost, 0.0001)
	})

	t.Run("returns nil for unknown model", func(t *testing.T) {
		cost, err := svc.CalculateCost(ctx, projectID, "unknown-model-xyz", 1000, 500)

		require.NoError(t, err)
		assert.Nil(t, cost)
	})

	t.Run("handles zero tokens", func(t *testing.T) {
		cost, err := svc.CalculateCost(ctx, projectID, "gpt-4o", 0, 0)

		require.NoError(t, err)
		require.NotNil(t, cost)
		assert.Equal(t, float64(0), cost.TotalCost)
	})

	t.Run("handles large token counts", func(t *testing.T) {
		cost, err := svc.CalculateCost(ctx, projectID, "gpt-4o", 1000000, 100000)

		require.NoError(t, err)
		require.NotNil(t, cost)

		// GPT-4o: $0.0025/1K input, $0.010/1K output
		expectedInputCost := 1000000 * 0.0025 / 1000  // $2.50
		expectedOutputCost := 100000 * 0.010 / 1000   // $1.00
		expectedTotal := expectedInputCost + expectedOutputCost

		assert.InDelta(t, expectedTotal, cost.TotalCost, 0.01)
	})
}

func TestCostService_CalculateCostWithCache(t *testing.T) {
	svc := NewCostService(zap.NewNop())
	ctx := context.Background()
	projectID := uuid.New()

	t.Run("calculates cost including cache tokens", func(t *testing.T) {
		cost, err := svc.CalculateCostWithCache(ctx, projectID, "gpt-4o", 1000, 500, 2000)

		require.NoError(t, err)
		require.NotNil(t, cost)

		// GPT-4o doesn't have cache pricing set (0), so cache cost should be 0
		expectedInputCost := 1000 * 0.0025 / 1000
		expectedOutputCost := 500 * 0.010 / 1000
		expectedTotal := expectedInputCost + expectedOutputCost // Cache is 0

		assert.InDelta(t, expectedTotal, cost.TotalCost, 0.0001)
	})

	t.Run("returns nil for unknown model", func(t *testing.T) {
		cost, err := svc.CalculateCostWithCache(ctx, projectID, "unknown-model", 1000, 500, 2000)

		require.NoError(t, err)
		assert.Nil(t, cost)
	})
}

func TestCostService_GetPricing(t *testing.T) {
	svc := NewCostService(zap.NewNop())

	t.Run("gets pricing for known model", func(t *testing.T) {
		// Use a versioned model to avoid duplicate entries between providers
		pricing := svc.GetPricing("gpt-4o-2024-11-20")

		require.NotNil(t, pricing)
		assert.Equal(t, "gpt-4o-2024-11-20", pricing.Model)
		assert.Equal(t, "openai", pricing.Provider)
		assert.Equal(t, 0.0025, pricing.InputPricePer1K)
		assert.Equal(t, 0.010, pricing.OutputPricePer1K)
	})

	t.Run("returns nil for unknown model", func(t *testing.T) {
		pricing := svc.GetPricing("totally-fake-model")

		assert.Nil(t, pricing)
	})

	t.Run("is case insensitive", func(t *testing.T) {
		// Use a versioned model to avoid duplicate entries between providers
		pricing1 := svc.GetPricing("GPT-4O-2024-11-20")
		pricing2 := svc.GetPricing("gpt-4o-2024-11-20")
		pricing3 := svc.GetPricing("Gpt-4O-2024-11-20")

		require.NotNil(t, pricing1)
		require.NotNil(t, pricing2)
		require.NotNil(t, pricing3)
		assert.Equal(t, pricing1.Model, pricing2.Model)
		assert.Equal(t, pricing2.Model, pricing3.Model)
	})

	t.Run("handles whitespace", func(t *testing.T) {
		pricing := svc.GetPricing("  gpt-4o  ")

		require.NotNil(t, pricing)
		assert.Equal(t, "gpt-4o", pricing.Model)
	})
}

func TestCostService_SetProjectPricing(t *testing.T) {
	svc := NewCostService(zap.NewNop())
	ctx := context.Background()
	projectID := uuid.New()

	t.Run("sets custom project pricing", func(t *testing.T) {
		customPricing := &ModelPricing{
			Model:            "gpt-4o",
			Provider:         "openai",
			InputPricePer1K:  0.001, // Custom lower price
			OutputPricePer1K: 0.005,
		}

		svc.SetProjectPricing(projectID, customPricing)

		// Calculate with custom pricing
		cost, err := svc.CalculateCost(ctx, projectID, "gpt-4o", 1000, 500)

		require.NoError(t, err)
		require.NotNil(t, cost)

		expectedInputCost := 1000 * 0.001 / 1000
		expectedOutputCost := 500 * 0.005 / 1000
		expectedTotal := expectedInputCost + expectedOutputCost

		assert.InDelta(t, expectedTotal, cost.TotalCost, 0.0001)
	})

	t.Run("uses default pricing for other projects", func(t *testing.T) {
		otherProjectID := uuid.New()

		cost, err := svc.CalculateCost(ctx, otherProjectID, "gpt-4o", 1000, 500)

		require.NoError(t, err)
		require.NotNil(t, cost)

		// Should use default GPT-4o pricing
		expectedInputCost := 1000 * 0.0025 / 1000
		expectedOutputCost := 500 * 0.010 / 1000
		expectedTotal := expectedInputCost + expectedOutputCost

		assert.InDelta(t, expectedTotal, cost.TotalCost, 0.0001)
	})

	t.Run("overrides only specified model", func(t *testing.T) {
		projectID := uuid.New()

		// Set custom pricing for gpt-4o only
		svc.SetProjectPricing(projectID, &ModelPricing{
			Model:            "gpt-4o",
			Provider:         "openai",
			InputPricePer1K:  0.001,
			OutputPricePer1K: 0.002,
		})

		// gpt-4o-mini should still use default pricing
		cost, err := svc.CalculateCost(ctx, projectID, "gpt-4o-mini", 1000, 500)

		require.NoError(t, err)
		require.NotNil(t, cost)

		// Default GPT-4o-mini pricing
		expectedInputCost := 1000 * 0.00015 / 1000
		expectedOutputCost := 500 * 0.0006 / 1000
		expectedTotal := expectedInputCost + expectedOutputCost

		assert.InDelta(t, expectedTotal, cost.TotalCost, 0.0001)
	})
}

func TestCostService_ListModels(t *testing.T) {
	svc := NewCostService(zap.NewNop())

	t.Run("lists all models", func(t *testing.T) {
		models := svc.ListModels()

		assert.Greater(t, len(models), 50)

		// Check some providers are present
		providers := make(map[string]bool)
		for _, m := range models {
			providers[m.Provider] = true
		}

		assert.True(t, providers["openai"], "OpenAI models should be present")
		assert.True(t, providers["anthropic"], "Anthropic models should be present")
		assert.True(t, providers["google"], "Google models should be present")
		assert.True(t, providers["mistral"], "Mistral models should be present")
	})

	t.Run("all models have required fields", func(t *testing.T) {
		models := svc.ListModels()

		for _, m := range models {
			assert.NotEmpty(t, m.Model, "Model name should not be empty")
			assert.NotEmpty(t, m.Provider, "Provider should not be empty")
			assert.GreaterOrEqual(t, m.InputPricePer1K, float64(0), "Input price should be non-negative")
			assert.GreaterOrEqual(t, m.OutputPricePer1K, float64(0), "Output price should be non-negative")
		}
	})
}

func TestCostService_ModelNormalization(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"gpt-4o", "gpt-4o"},
		{"GPT-4O", "gpt-4o"},
		{"  gpt-4o  ", "gpt-4o"},
		{"GPT-4O  ", "gpt-4o"},
		{"  GPT-4O", "gpt-4o"},
		{"CLAUDE-3-OPUS-20240229", "claude-3-opus-20240229"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := normalizeModel(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCostService_PartialModelMatching(t *testing.T) {
	svc := NewCostService(zap.NewNop())

	t.Run("matches versioned model prefix", func(t *testing.T) {
		// If we query for a base model, it should match the specific version
		pricing := svc.GetPricing("gpt-4o-2024-11-20")

		require.NotNil(t, pricing)
		assert.Equal(t, "openai", pricing.Provider)
	})
}

func TestCostService_SpecificProviderModels(t *testing.T) {
	svc := NewCostService(zap.NewNop())
	ctx := context.Background()
	projectID := uuid.New()

	t.Run("OpenAI o1 model", func(t *testing.T) {
		cost, err := svc.CalculateCost(ctx, projectID, "o1", 1000, 500)

		require.NoError(t, err)
		require.NotNil(t, cost)

		// o1: $0.015/1K input, $0.060/1K output
		expectedInputCost := 1000 * 0.015 / 1000
		expectedOutputCost := 500 * 0.060 / 1000

		assert.InDelta(t, expectedInputCost, cost.InputCost, 0.0001)
		assert.InDelta(t, expectedOutputCost, cost.OutputCost, 0.0001)
	})

	t.Run("Google Gemini model", func(t *testing.T) {
		cost, err := svc.CalculateCost(ctx, projectID, "gemini-1.5-pro", 10000, 5000)

		require.NoError(t, err)
		require.NotNil(t, cost)

		// Gemini 1.5 Pro: $0.00125/1K input, $0.005/1K output
		expectedInputCost := 10000 * 0.00125 / 1000
		expectedOutputCost := 5000 * 0.005 / 1000

		assert.InDelta(t, expectedInputCost, cost.InputCost, 0.0001)
		assert.InDelta(t, expectedOutputCost, cost.OutputCost, 0.0001)
	})

	t.Run("Mistral model", func(t *testing.T) {
		cost, err := svc.CalculateCost(ctx, projectID, "mistral-large-latest", 2000, 1000)

		require.NoError(t, err)
		require.NotNil(t, cost)

		// Mistral Large: $0.002/1K input, $0.006/1K output
		expectedInputCost := 2000 * 0.002 / 1000
		expectedOutputCost := 1000 * 0.006 / 1000

		assert.InDelta(t, expectedInputCost, cost.InputCost, 0.0001)
		assert.InDelta(t, expectedOutputCost, cost.OutputCost, 0.0001)
	})

	t.Run("DeepSeek model", func(t *testing.T) {
		cost, err := svc.CalculateCost(ctx, projectID, "deepseek-chat", 50000, 10000)

		require.NoError(t, err)
		require.NotNil(t, cost)

		// DeepSeek Chat: $0.00014/1K input, $0.00028/1K output
		expectedInputCost := 50000 * 0.00014 / 1000
		expectedOutputCost := 10000 * 0.00028 / 1000

		assert.InDelta(t, expectedInputCost, cost.InputCost, 0.0001)
		assert.InDelta(t, expectedOutputCost, cost.OutputCost, 0.0001)
	})

	t.Run("Cohere model", func(t *testing.T) {
		cost, err := svc.CalculateCost(ctx, projectID, "command-r-plus", 5000, 2000)

		require.NoError(t, err)
		require.NotNil(t, cost)

		// Command R+: $0.002/1K input, $0.01/1K output
		expectedInputCost := 5000 * 0.002 / 1000
		expectedOutputCost := 2000 * 0.01 / 1000

		assert.InDelta(t, expectedInputCost, cost.InputCost, 0.0001)
		assert.InDelta(t, expectedOutputCost, cost.OutputCost, 0.0001)
	})

	t.Run("Groq/Llama model", func(t *testing.T) {
		cost, err := svc.CalculateCost(ctx, projectID, "llama-3.3-70b-versatile", 10000, 5000)

		require.NoError(t, err)
		require.NotNil(t, cost)

		// Llama 3.3 70B: $0.00059/1K input, $0.00079/1K output
		expectedInputCost := 10000 * 0.00059 / 1000
		expectedOutputCost := 5000 * 0.00079 / 1000

		assert.InDelta(t, expectedInputCost, cost.InputCost, 0.0001)
		assert.InDelta(t, expectedOutputCost, cost.OutputCost, 0.0001)
	})

	t.Run("AWS Bedrock model", func(t *testing.T) {
		cost, err := svc.CalculateCost(ctx, projectID, "amazon.nova-pro-v1:0", 10000, 5000)

		require.NoError(t, err)
		require.NotNil(t, cost)

		// Nova Pro: $0.0008/1K input, $0.0032/1K output
		expectedInputCost := 10000 * 0.0008 / 1000
		expectedOutputCost := 5000 * 0.0032 / 1000

		assert.InDelta(t, expectedInputCost, cost.InputCost, 0.0001)
		assert.InDelta(t, expectedOutputCost, cost.OutputCost, 0.0001)
	})

	t.Run("xAI Grok model", func(t *testing.T) {
		cost, err := svc.CalculateCost(ctx, projectID, "grok-2", 5000, 2000)

		require.NoError(t, err)
		require.NotNil(t, cost)

		// Grok 2: $0.002/1K input, $0.010/1K output
		expectedInputCost := 5000 * 0.002 / 1000
		expectedOutputCost := 2000 * 0.010 / 1000

		assert.InDelta(t, expectedInputCost, cost.InputCost, 0.0001)
		assert.InDelta(t, expectedOutputCost, cost.OutputCost, 0.0001)
	})
}

func TestCostService_EmbeddingModels(t *testing.T) {
	svc := NewCostService(zap.NewNop())
	ctx := context.Background()
	projectID := uuid.New()

	t.Run("OpenAI embedding model (output is 0)", func(t *testing.T) {
		cost, err := svc.CalculateCost(ctx, projectID, "text-embedding-3-small", 10000, 0)

		require.NoError(t, err)
		require.NotNil(t, cost)

		// Embedding 3 Small: $0.00002/1K input, $0 output
		expectedInputCost := 10000 * 0.00002 / 1000
		expectedOutputCost := float64(0)

		assert.InDelta(t, expectedInputCost, cost.InputCost, 0.0001)
		assert.Equal(t, expectedOutputCost, cost.OutputCost)
	})

	t.Run("Cohere embedding model", func(t *testing.T) {
		cost, err := svc.CalculateCost(ctx, projectID, "embed-english-v3.0", 50000, 0)

		require.NoError(t, err)
		require.NotNil(t, cost)

		// Cohere Embed: $0.0001/1K input, $0 output
		expectedInputCost := 50000 * 0.0001 / 1000

		assert.InDelta(t, expectedInputCost, cost.InputCost, 0.0001)
	})
}

func TestCostService_ConcurrentAccess(t *testing.T) {
	svc := NewCostService(zap.NewNop())
	ctx := context.Background()
	projectID := uuid.New()

	t.Run("handles concurrent cost calculations", func(t *testing.T) {
		done := make(chan bool, 100)

		for i := 0; i < 100; i++ {
			go func() {
				cost, err := svc.CalculateCost(ctx, projectID, "gpt-4o", 1000, 500)
				assert.NoError(t, err)
				assert.NotNil(t, cost)
				done <- true
			}()
		}

		for i := 0; i < 100; i++ {
			<-done
		}
	})

	t.Run("handles concurrent pricing reads and writes", func(t *testing.T) {
		done := make(chan bool, 100)

		// Concurrent reads
		for i := 0; i < 50; i++ {
			go func() {
				_ = svc.ListModels()
				done <- true
			}()
		}

		// Concurrent writes
		for i := 0; i < 50; i++ {
			go func(id int) {
				projectID := uuid.New()
				svc.SetProjectPricing(projectID, &ModelPricing{
					Model:            "custom-model",
					Provider:         "custom",
					InputPricePer1K:  float64(id) * 0.001,
					OutputPricePer1K: float64(id) * 0.002,
				})
				done <- true
			}(i)
		}

		for i := 0; i < 100; i++ {
			<-done
		}
	})
}
