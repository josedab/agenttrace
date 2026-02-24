package domain

import (
	"time"

	"github.com/google/uuid"
)

// FrameworkType represents a supported AI agent framework
type FrameworkType string

const (
	FrameworkTypeLangChain  FrameworkType = "langchain"
	FrameworkTypeCrewAI     FrameworkType = "crewai"
	FrameworkTypeAutoGen    FrameworkType = "autogen"
	FrameworkTypeLlamaIndex FrameworkType = "llamaindex"
	FrameworkTypeOpenAI     FrameworkType = "openai"
	FrameworkTypeAnthropic  FrameworkType = "anthropic"
	FrameworkTypeCustom     FrameworkType = "custom"
)

// IsValid checks if the framework type is valid
func (f FrameworkType) IsValid() bool {
	switch f {
	case FrameworkTypeLangChain, FrameworkTypeCrewAI, FrameworkTypeAutoGen, FrameworkTypeLlamaIndex, FrameworkTypeOpenAI, FrameworkTypeAnthropic, FrameworkTypeCustom:
		return true
	}
	return false
}

// DiscoveryStatus represents the status of a discovered framework
type DiscoveryStatus string

const (
	DiscoveryStatusDetected     DiscoveryStatus = "detected"
	DiscoveryStatusConfirmed    DiscoveryStatus = "confirmed"
	DiscoveryStatusInstrumented DiscoveryStatus = "instrumented"
	DiscoveryStatusDisabled     DiscoveryStatus = "disabled"
)

// IsValid checks if the discovery status is valid
func (s DiscoveryStatus) IsValid() bool {
	switch s {
	case DiscoveryStatusDetected, DiscoveryStatusConfirmed, DiscoveryStatusInstrumented, DiscoveryStatusDisabled:
		return true
	}
	return false
}

// DiscoveredFramework represents an auto-discovered AI framework in a project
type DiscoveredFramework struct {
	ID               uuid.UUID             `json:"id"`
	ProjectID        uuid.UUID             `json:"projectId"`
	Framework        FrameworkType         `json:"framework"`
	Version          string                `json:"version"`
	Status           DiscoveryStatus       `json:"status"`
	DetectedAt       time.Time             `json:"detectedAt"`
	Components       []DiscoveredComponent `json:"components"`
	AutoInstrumented bool                  `json:"autoInstrumented"`
	Config           DiscoveryConfig       `json:"config"`
}

// DiscoveredComponent represents a single component discovered within a framework
type DiscoveredComponent struct {
	Name         string    `json:"name"`
	Type         string    `json:"type"`
	CallCount    int       `json:"callCount"`
	AvgLatencyMs float64   `json:"avgLatencyMs"`
	FirstSeen    time.Time `json:"firstSeen"`
	LastSeen     time.Time `json:"lastSeen"`
}

// DiscoveryConfig represents configuration for the auto-discovery engine
type DiscoveryConfig struct {
	Enabled         bool     `json:"enabled"`
	ExcludePatterns []string `json:"excludePatterns,omitempty"`
	SamplingRate    float64  `json:"samplingRate"`
	MaxDepth        int      `json:"maxDepth"`
}

// DiscoveryDashboard represents an overview of all discovered frameworks in a project
type DiscoveryDashboard struct {
	Frameworks              []DiscoveredFramework `json:"frameworks"`
	TotalComponents         int                   `json:"totalComponents"`
	InstrumentedComponents  int                   `json:"instrumentedComponents"`
	LastScanAt              *time.Time            `json:"lastScanAt,omitempty"`
}
