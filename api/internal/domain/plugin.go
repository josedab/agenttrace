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
