package domain

import "github.com/google/uuid"

// InstrumentationFramework represents a supported agent framework for instrumentation
type InstrumentationFramework string

const (
	InstrumentationFrameworkCrewAI         InstrumentationFramework = "CREWAI"
	InstrumentationFrameworkAutogen        InstrumentationFramework = "AUTOGEN"
	InstrumentationFrameworkLangGraph      InstrumentationFramework = "LANGGRAPH"
	InstrumentationFrameworkLlamaIndex     InstrumentationFramework = "LLAMAINDEX"
	InstrumentationFrameworkSemanticKernel InstrumentationFramework = "SEMANTIC_KERNEL"
)

// IsValid checks if the instrumentation framework is valid
func (f InstrumentationFramework) IsValid() bool {
	switch f {
	case InstrumentationFrameworkCrewAI,
		InstrumentationFrameworkAutogen,
		InstrumentationFrameworkLangGraph,
		InstrumentationFrameworkLlamaIndex,
		InstrumentationFrameworkSemanticKernel:
		return true
	}
	return false
}

// InstrumentationConfig holds configuration for instrumenting a specific framework
type InstrumentationConfig struct {
	Framework       InstrumentationFramework `json:"framework"`
	Enabled         bool                     `json:"enabled"`
	AutoTraceAgents bool                     `json:"autoTraceAgents"`
	AutoTraceTools  bool                     `json:"autoTraceTools"`
	AutoTraceMessages bool                   `json:"autoTraceMessages"`
	CaptureIO       bool                     `json:"captureIO"`
	ProjectID       uuid.UUID                `json:"projectId"`
}

// InstrumentationSetup provides setup instructions for a framework and language
type InstrumentationSetup struct {
	Framework      InstrumentationFramework `json:"framework"`
	Language       string                   `json:"language"`
	InstallCommand string                   `json:"installCommand"`
	ConfigSnippet  string                   `json:"configSnippet"`
	ExampleCode    string                   `json:"exampleCode"`
}
