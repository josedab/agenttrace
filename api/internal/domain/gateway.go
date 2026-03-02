package domain

import (
	"time"

	"github.com/google/uuid"
)

// GatewayProvider represents a supported LLM provider
type GatewayProvider string

const (
	ProviderOpenAI    GatewayProvider = "openai"
	ProviderAnthropic GatewayProvider = "anthropic"
	ProviderGoogle    GatewayProvider = "google"
	ProviderLocal     GatewayProvider = "local"
)

// RoutingStrategy defines how requests are routed to providers
type RoutingStrategy string

const (
	RoutingCheapest   RoutingStrategy = "cheapest"
	RoutingFastest    RoutingStrategy = "fastest"
	RoutingRoundRobin RoutingStrategy = "round_robin"
	RoutingPriority   RoutingStrategy = "priority"
	RoutingFallback   RoutingStrategy = "fallback"
)

// GatewayConfig represents the configuration for the LLM gateway
type GatewayConfig struct {
	ID              uuid.UUID       `json:"id"`
	ProjectID       uuid.UUID       `json:"projectId"`
	Name            string          `json:"name"`
	Strategy        RoutingStrategy `json:"strategy"`
	Providers       []ProviderConfig `json:"providers"`
	FallbackChain   []string        `json:"fallbackChain,omitempty"`
	RateLimitRPM    int             `json:"rateLimitRpm"`
	RateLimitTPM    int             `json:"rateLimitTpm"`
	MaxRetries      int             `json:"maxRetries"`
	TimeoutSeconds  int             `json:"timeoutSeconds"`
	TracingEnabled  bool            `json:"tracingEnabled"`
	Enabled         bool            `json:"enabled"`
	CreatedAt       time.Time       `json:"createdAt"`
	UpdatedAt       time.Time       `json:"updatedAt"`
}

// ProviderConfig represents configuration for a single LLM provider
type ProviderConfig struct {
	Provider    GatewayProvider `json:"provider"`
	BaseURL     string          `json:"baseUrl,omitempty"`
	Models      []string        `json:"models"`
	Priority    int             `json:"priority"`
	Weight      float64         `json:"weight"`
	MaxRPM      int             `json:"maxRpm"`
	CostPer1K   float64         `json:"costPer1k"`
	Enabled     bool            `json:"enabled"`
}

// GatewayRequest represents a proxied LLM API request
type GatewayRequest struct {
	ID         uuid.UUID       `json:"id"`
	ProjectID  uuid.UUID       `json:"projectId"`
	Provider   GatewayProvider `json:"provider"`
	Model      string          `json:"model"`
	Messages   []GatewayMessage `json:"messages"`
	MaxTokens  int             `json:"maxTokens,omitempty"`
	Temperature *float64       `json:"temperature,omitempty"`
	Stream     bool            `json:"stream,omitempty"`
	Metadata   map[string]string `json:"metadata,omitempty"`
}

// GatewayMessage represents a chat message in the unified API
type GatewayMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// GatewayResponse represents the response from an LLM provider
type GatewayResponse struct {
	ID              string          `json:"id"`
	Provider        GatewayProvider `json:"provider"`
	Model           string          `json:"model"`
	Choices         []GatewayChoice `json:"choices"`
	Usage           GatewayUsage    `json:"usage"`
	LatencyMs       int64           `json:"latencyMs"`
	EstimatedCost   float64         `json:"estimatedCost"`
	TraceID         string          `json:"traceId,omitempty"`
	FallbackUsed    bool            `json:"fallbackUsed"`
	OriginalProvider GatewayProvider `json:"originalProvider,omitempty"`
}

// GatewayChoice represents a response choice
type GatewayChoice struct {
	Index        int            `json:"index"`
	Message      GatewayMessage `json:"message"`
	FinishReason string         `json:"finishReason"`
}

// GatewayUsage represents token usage information
type GatewayUsage struct {
	PromptTokens     int `json:"promptTokens"`
	CompletionTokens int `json:"completionTokens"`
	TotalTokens      int `json:"totalTokens"`
}

// GatewayConfigInput represents input for creating/updating a gateway config
type GatewayConfigInput struct {
	Name           string           `json:"name" validate:"required"`
	Strategy       RoutingStrategy  `json:"strategy" validate:"required"`
	Providers      []ProviderConfig `json:"providers" validate:"required"`
	FallbackChain  []string         `json:"fallbackChain,omitempty"`
	RateLimitRPM   int              `json:"rateLimitRpm,omitempty"`
	RateLimitTPM   int              `json:"rateLimitTpm,omitempty"`
	MaxRetries     int              `json:"maxRetries,omitempty"`
	TimeoutSeconds int              `json:"timeoutSeconds,omitempty"`
	TracingEnabled *bool            `json:"tracingEnabled,omitempty"`
}

// RoutingRule defines a smart routing rule
type RoutingRule struct {
	ID          uuid.UUID `json:"id"`
	ConfigID    uuid.UUID `json:"configId"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Condition   GatewayRuleCondition `json:"condition"`
	Action      RuleAction    `json:"action"`
	Priority    int       `json:"priority"`
	Enabled     bool      `json:"enabled"`
	CreatedAt   time.Time `json:"createdAt"`
}

// RuleCondition defines when a routing rule applies
type GatewayRuleCondition struct {
	MaxTokens      *int     `json:"maxTokens,omitempty"`
	MinTokens      *int     `json:"minTokens,omitempty"`
	ModelPattern   string   `json:"modelPattern,omitempty"`
	CostThreshold  *float64 `json:"costThreshold,omitempty"`
	TaskComplexity string   `json:"taskComplexity,omitempty"` // "simple", "moderate", "complex"
}

// RuleAction defines what to do when a rule matches
type RuleAction struct {
	RouteToProvider GatewayProvider `json:"routeToProvider,omitempty"`
	RouteToModel    string          `json:"routeToModel,omitempty"`
	ApplyRateLimit  *int            `json:"applyRateLimit,omitempty"`
}

// GatewayStats represents gateway usage statistics
type GatewayStats struct {
	TotalRequests    int64              `json:"totalRequests"`
	TotalTokens      int64              `json:"totalTokens"`
	TotalCost        float64            `json:"totalCost"`
	AvgLatencyMs     float64            `json:"avgLatencyMs"`
	ErrorRate        float64            `json:"errorRate"`
	ProviderBreakdown map[string]int64  `json:"providerBreakdown"`
	FallbackRate     float64            `json:"fallbackRate"`
}
