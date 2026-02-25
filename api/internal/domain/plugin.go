package domain

import (
	"time"

	"github.com/google/uuid"
)

// PluginType represents the type of plugin
type PluginType string

const (
	PluginTypeEvaluator        PluginType = "evaluator"
	PluginTypeTraceProcessor   PluginType = "trace_processor"
	PluginTypeDashboardWidget  PluginType = "dashboard_widget"
	PluginTypeMarketplace      PluginType = "marketplace_package"
)

// PluginStatus represents the status of a plugin
type PluginStatus string

const (
	PluginStatusInstalled PluginStatus = "installed"
	PluginStatusActive    PluginStatus = "active"
	PluginStatusDisabled  PluginStatus = "disabled"
	PluginStatusError     PluginStatus = "error"
)

// PluginExecutionStatus represents the status of a plugin execution
type PluginExecutionStatus string

const (
	PluginExecSuccess PluginExecutionStatus = "success"
	PluginExecError   PluginExecutionStatus = "error"
)

// Plugin represents an installed plugin
type Plugin struct {
	ID          uuid.UUID      `json:"id"`
	ProjectID   uuid.UUID      `json:"projectId"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Type        PluginType     `json:"type"`
	Version     string         `json:"version"`
	Author      string         `json:"author"`
	EntryPoint  string         `json:"entryPoint"`
	Config      map[string]any `json:"config,omitempty"`
	Status      PluginStatus   `json:"status"`
	CreatedAt   time.Time      `json:"createdAt"`
}

// PluginManifest represents a plugin manifest for installation
type PluginManifest struct {
	Name         string         `json:"name" validate:"required"`
	Version      string         `json:"version" validate:"required"`
	Type         PluginType     `json:"type" validate:"required"`
	Description  string         `json:"description"`
	Author       string         `json:"author"`
	EntryPoint   string         `json:"entryPoint"`
	Permissions  []string       `json:"permissions,omitempty"`
	ConfigSchema map[string]any `json:"configSchema,omitempty"`
}

// PluginInput represents input for installing a plugin
type PluginInput struct {
	Manifest PluginManifest `json:"manifest" validate:"required"`
}

// PluginExecution represents a plugin execution record
type PluginExecution struct {
	ID         uuid.UUID             `json:"id"`
	PluginID   uuid.UUID             `json:"pluginId"`
	Input      string                `json:"input"`
	Output     string                `json:"output"`
	DurationMs int64                 `json:"durationMs"`
	Status     PluginExecutionStatus `json:"status"`
	ExecutedAt time.Time             `json:"executedAt"`
}

// PluginRegistry represents the plugin registry
type PluginRegistry struct {
	Plugins    []Plugin `json:"plugins"`
	TotalCount int      `json:"totalCount"`
}

// AgentProtocolAdapter represents a framework-specific adapter plugin
type AgentProtocolAdapter struct {
	ID          uuid.UUID              `json:"id"`
	ProjectID   uuid.UUID              `json:"projectId"`
	Framework   AgentFramework         `json:"framework"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Version     string                 `json:"version"`
	Config      map[string]any         `json:"config,omitempty"`
	Status      PluginStatus           `json:"status"`
	Capabilities []string              `json:"capabilities"` // trace_capture, cost_tracking, prompt_logging
	TracesCaptured int64               `json:"tracesCaptured"`
	LastActiveAt   *time.Time          `json:"lastActiveAt,omitempty"`
	CreatedAt   time.Time              `json:"createdAt"`
	UpdatedAt   time.Time              `json:"updatedAt"`
}

// AgentFramework represents a supported agent framework
type AgentFramework string

const (
	FrameworkLangChain   AgentFramework = "langchain"
	FrameworkCrewAI      AgentFramework = "crewai"
	FrameworkAutoGen     AgentFramework = "autogen"
	FrameworkLlamaIndex  AgentFramework = "llamaindex"
	FrameworkCustom      AgentFramework = "custom"
	FrameworkOpenAI      AgentFramework = "openai_agents"
	FrameworkDSPy        AgentFramework = "dspy"
	FrameworkHaystack    AgentFramework = "haystack"
	FrameworkSemantic    AgentFramework = "semantic_kernel"
)

// AdapterInstallInput represents input for installing an adapter
type AdapterInstallInput struct {
	Framework    AgentFramework `json:"framework" validate:"required"`
	Name         string         `json:"name,omitempty"`
	Config       map[string]any `json:"config,omitempty"`
}

// AdapterEventInput represents a trace event from an adapter
type AdapterEventInput struct {
	AdapterID   uuid.UUID      `json:"adapterId" validate:"required"`
	EventType   string         `json:"eventType" validate:"required"` // llm_call, tool_use, agent_step, chain_start, chain_end
	TraceID     string         `json:"traceId"`
	SpanID      string         `json:"spanId"`
	ParentID    string         `json:"parentId,omitempty"`
	Name        string         `json:"name"`
	Input       any            `json:"input,omitempty"`
	Output      any            `json:"output,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
	StartTime   time.Time      `json:"startTime"`
	EndTime     *time.Time     `json:"endTime,omitempty"`
	Error       string         `json:"error,omitempty"`
}

// AdapterRegistry represents the list of available and installed adapters
type AdapterRegistry struct {
	Installed []AgentProtocolAdapter `json:"installed"`
	Available []AdapterTemplate      `json:"available"`
}

// AdapterTemplate represents a framework adapter template available for installation
type AdapterTemplate struct {
	Framework    AgentFramework `json:"framework"`
	DisplayName  string         `json:"displayName"`
	Description  string         `json:"description"`
	SetupGuide   string         `json:"setupGuide"`
	Capabilities []string       `json:"capabilities"`
	SDKSupport   []string       `json:"sdkSupport"` // python, typescript, go
}
