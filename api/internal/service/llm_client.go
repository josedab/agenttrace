package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"go.uber.org/zap"
)

// LLMClient provides LLM API calling capabilities
type LLMClient struct {
	logger     *zap.Logger
	httpClient *http.Client
	apiKey     string
	baseURL    string
	model      string
	guard      OutboundGuard
}

// LLMRequest represents a chat completion request
type LLMRequest struct {
	Model       string       `json:"model"`
	Messages    []LLMMessage `json:"messages"`
	Temperature float64      `json:"temperature,omitempty"`
	MaxTokens   int          `json:"max_tokens,omitempty"`
}

// LLMMessage represents a chat message
type LLMMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// LLMResponse represents a chat completion response
type LLMResponse struct {
	ID      string `json:"id"`
	Choices []struct {
		Message LLMMessage `json:"message"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// NewLLMClient creates a new LLM client.
// The optional outbound guard rejects provider calls in no-egress mode; the
// client then falls back to its local response instead of leaving the network.
func NewLLMClient(logger *zap.Logger, guards ...OutboundGuard) *LLMClient {
	var guard OutboundGuard
	if len(guards) > 0 {
		guard = guards[0]
	}
	apiKey := os.Getenv("OPENAI_API_KEY")
	baseURL := os.Getenv("LLM_BASE_URL")
	model := os.Getenv("LLM_MODEL")

	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}
	if model == "" {
		model = "gpt-4o-mini"
	}

	return &LLMClient{
		logger:     logger,
		httpClient: &http.Client{Timeout: 60 * time.Second},
		apiKey:     apiKey,
		baseURL:    baseURL,
		model:      model,
		guard:      guard,
	}
}

// IsConfigured returns whether an external provider call is possible.
// No-egress mode reports the provider as unconfigured so callers keep using the
// deterministic local fallback.
func (c *LLMClient) IsConfigured() bool {
	if RequireOutbound(c.guard, EgressExternalModel) != nil {
		return false
	}
	return c.apiKey != ""
}

// ChatCompletion sends a chat completion request
func (c *LLMClient) ChatCompletion(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	if !c.IsConfigured() {
		return c.fallbackResponse(systemPrompt, userPrompt), nil
	}

	req := LLMRequest{
		Model: c.model,
		Messages: []LLMMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
		Temperature: 0.3,
		MaxTokens:   2000,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		c.logger.Warn("LLM API call failed, using fallback", zap.Error(err))
		return c.fallbackResponse(systemPrompt, userPrompt), nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		c.logger.Warn("LLM API returned error",
			zap.Int("status", resp.StatusCode),
			zap.String("body", string(respBody)),
		)
		return c.fallbackResponse(systemPrompt, userPrompt), nil
	}

	var llmResp LLMResponse
	if err := json.NewDecoder(resp.Body).Decode(&llmResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if len(llmResp.Choices) == 0 {
		return c.fallbackResponse(systemPrompt, userPrompt), nil
	}

	c.logger.Debug("LLM API call succeeded",
		zap.Int("promptTokens", llmResp.Usage.PromptTokens),
		zap.Int("completionTokens", llmResp.Usage.CompletionTokens),
	)

	return llmResp.Choices[0].Message.Content, nil
}

// fallbackResponse generates a response when LLM API is not available
func (c *LLMClient) fallbackResponse(systemPrompt, userPrompt string) string {
	return fmt.Sprintf("Analysis generated using heuristic fallback (no LLM API key configured). "+
		"Configure OPENAI_API_KEY environment variable for AI-powered analysis. "+
		"System context: %d chars, User input: %d chars.",
		len(systemPrompt), len(userPrompt))
}
