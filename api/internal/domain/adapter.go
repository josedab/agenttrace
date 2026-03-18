package domain

import (
	"time"

	"github.com/google/uuid"
)

// AdapterStatus represents the registration status of an adapter
type AdapterStatus string

const (
	AdapterStatusRegistered AdapterStatus = "registered"
	AdapterStatusActive     AdapterStatus = "active"
	AdapterStatusInactive   AdapterStatus = "inactive"
	AdapterStatusDeprecated AdapterStatus = "deprecated"
)

// AdapterFramework represents a supported agent framework
type AdapterFramework string

const (
	AdapterFrameworkLangChain      AdapterFramework = "langchain"
	AdapterFrameworkCrewAI         AdapterFramework = "crewai"
	AdapterFrameworkAutoGen        AdapterFramework = "autogen"
	AdapterFrameworkLangGraph      AdapterFramework = "langgraph"
	AdapterFrameworkOpenHands      AdapterFramework = "openhands"
	AdapterFrameworkSemanticKernel AdapterFramework = "semantic_kernel"
	AdapterFrameworkCustom         AdapterFramework = "custom"
)

// AgentAdapter represents a registered protocol adapter
type AgentAdapter struct {
	ID             uuid.UUID        `json:"id"`
	ProjectID      uuid.UUID        `json:"projectId"`
	Name           string           `json:"name"`
	Framework      AdapterFramework `json:"framework"`
	Version        string           `json:"version"`
	Status         AdapterStatus    `json:"status"`
	Config         AdapterConfig    `json:"config"`
	Capabilities   []string         `json:"capabilities"`
	LifecycleHooks []LifecycleHook  `json:"lifecycleHooks"`
	Stats          AdapterStats     `json:"stats"`
	CreatedAt      time.Time        `json:"createdAt"`
	UpdatedAt      time.Time        `json:"updatedAt"`
}

// AdapterConfig represents adapter-specific configuration
type AdapterConfig struct {
	AutoInstrument  bool              `json:"autoInstrument"`
	CaptureIO       bool              `json:"captureIO"`
	CaptureMetadata bool              `json:"captureMetadata"`
	MaxSpanDepth    int               `json:"maxSpanDepth"`
	SamplingRate    float64           `json:"samplingRate"`
	CustomHeaders   map[string]string `json:"customHeaders,omitempty"`
	TransformRules  []TransformRule   `json:"transformRules,omitempty"`
}

// TransformRule defines how to map framework-specific data to AgentTrace schema
type TransformRule struct {
	SourceField string `json:"sourceField"`
	TargetField string `json:"targetField"`
	Transform   string `json:"transform"` // direct, regex, jmespath, template
	Expression  string `json:"expression,omitempty"`
}

// LifecycleHook represents a hook into the adapter lifecycle
type LifecycleHook struct {
	Name       string `json:"name"`       // on_start, on_complete, on_error, on_tool_call, on_llm_call, on_chain_start, on_agent_action
	Enabled    bool   `json:"enabled"`
	WebhookURL string `json:"webhookUrl,omitempty"`
	FilterExpr string `json:"filterExpr,omitempty"`
}

// AdapterStats tracks adapter usage metrics
type AdapterStats struct {
	TotalTraces   int64      `json:"totalTraces"`
	TotalSpans    int64      `json:"totalSpans"`
	AvgLatencyMs  float64    `json:"avgLatencyMs"`
	ErrorRate     float64    `json:"errorRate"`
	LastActiveAt  *time.Time `json:"lastActiveAt,omitempty"`
	TracesPerHour float64    `json:"tracesPerHour"`
}

// AdapterInput represents input for registering an adapter
type AdapterInput struct {
	Name           string           `json:"name" validate:"required,min=1,max=200"`
	Framework      AdapterFramework `json:"framework" validate:"required"`
	Version        string           `json:"version,omitempty"`
	Config         *AdapterConfig   `json:"config,omitempty"`
	Capabilities   []string         `json:"capabilities,omitempty"`
	LifecycleHooks []LifecycleHook  `json:"lifecycleHooks,omitempty"`
}

// AdapterUpdateInput represents input for updating an adapter
type AdapterUpdateInput struct {
	Name           *string          `json:"name,omitempty"`
	Status         *AdapterStatus   `json:"status,omitempty"`
	Config         *AdapterConfig   `json:"config,omitempty"`
	LifecycleHooks []LifecycleHook  `json:"lifecycleHooks,omitempty"`
}

// AdapterTestResult represents the result of testing an adapter
type AdapterTestResult struct {
	AdapterID   uuid.UUID         `json:"adapterId"`
	Framework   AdapterFramework  `json:"framework"`
	Passed      bool              `json:"passed"`
	TestResults []AdapterTestCase `json:"testResults"`
	Summary     string            `json:"summary"`
	TestedAt    time.Time         `json:"testedAt"`
}

// AdapterTestCase represents a single test case result
type AdapterTestCase struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Passed      bool   `json:"passed"`
	Error       string `json:"error,omitempty"`
	DurationMs  int64  `json:"durationMs"`
}

// AdapterEvent represents an incoming event from an adapter
type AdapterEvent struct {
	AdapterID  uuid.UUID              `json:"adapterId"`
	Framework  AdapterFramework       `json:"framework"`
	EventType  string                 `json:"eventType"` // trace_start, span_start, span_end, trace_end, error
	TraceID    string                 `json:"traceId,omitempty"`
	SpanID     string                 `json:"spanId,omitempty"`
	ParentID   string                 `json:"parentId,omitempty"`
	Name       string                 `json:"name"`
	Input      map[string]interface{} `json:"input,omitempty"`
	Output     map[string]interface{} `json:"output,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	StartTime  time.Time              `json:"startTime"`
	EndTime    *time.Time             `json:"endTime,omitempty"`
	StatusCode string                 `json:"statusCode,omitempty"` // ok, error
	Error      string                 `json:"error,omitempty"`
}

// AdapterTemplateV2 provides quickstart templates for frameworks
type AdapterTemplateV2 struct {
	Framework    AdapterFramework `json:"framework"`
	Name         string           `json:"name"`
	Description  string           `json:"description"`
	SetupCode    string           `json:"setupCode"`
	Language     string           `json:"language"` // python, typescript, csharp
	Dependencies []string         `json:"dependencies"`
}
