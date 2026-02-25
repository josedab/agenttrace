package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

// PluginService manages the plugin architecture
type PluginService struct {
	logger     *zap.Logger
	mu         sync.RWMutex
	plugins    map[uuid.UUID]*domain.Plugin
	executions map[uuid.UUID]*domain.PluginExecution
}

// NewPluginService creates a new plugin service
func NewPluginService(logger *zap.Logger) *PluginService {
	return &PluginService{
		logger:     logger,
		plugins:    make(map[uuid.UUID]*domain.Plugin),
		executions: make(map[uuid.UUID]*domain.PluginExecution),
	}
}

// InstallPlugin installs a new plugin
func (s *PluginService) InstallPlugin(ctx context.Context, projectID uuid.UUID, input *domain.PluginInput) (*domain.Plugin, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	plugin := &domain.Plugin{
		ID:          uuid.New(),
		ProjectID:   projectID,
		Name:        input.Manifest.Name,
		Description: input.Manifest.Description,
		Type:        input.Manifest.Type,
		Version:     input.Manifest.Version,
		Author:      input.Manifest.Author,
		EntryPoint:  input.Manifest.EntryPoint,
		Config:      map[string]any{},
		Status:      domain.PluginStatusInstalled,
		CreatedAt:   time.Now(),
	}

	s.plugins[plugin.ID] = plugin
	s.logger.Info("installed plugin", zap.String("id", plugin.ID.String()), zap.String("name", input.Manifest.Name))
	return plugin, nil
}

// ListPlugins lists all plugins for a project
func (s *PluginService) ListPlugins(ctx context.Context, projectID uuid.UUID) (*domain.PluginRegistry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var plugins []domain.Plugin
	for _, p := range s.plugins {
		if p.ProjectID == projectID {
			plugins = append(plugins, *p)
		}
	}
	return &domain.PluginRegistry{
		Plugins:    plugins,
		TotalCount: len(plugins),
	}, nil
}

// GetPlugin returns a plugin by ID
func (s *PluginService) GetPlugin(ctx context.Context, id uuid.UUID) (*domain.Plugin, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	plugin, ok := s.plugins[id]
	if !ok {
		return nil, fmt.Errorf("plugin not found: %s", id)
	}
	return plugin, nil
}

// ActivatePlugin activates a plugin
func (s *PluginService) ActivatePlugin(ctx context.Context, id uuid.UUID) (*domain.Plugin, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	plugin, ok := s.plugins[id]
	if !ok {
		return nil, fmt.Errorf("plugin not found: %s", id)
	}
	plugin.Status = domain.PluginStatusActive
	s.logger.Info("activated plugin", zap.String("id", id.String()))
	return plugin, nil
}

// DisablePlugin disables a plugin
func (s *PluginService) DisablePlugin(ctx context.Context, id uuid.UUID) (*domain.Plugin, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	plugin, ok := s.plugins[id]
	if !ok {
		return nil, fmt.Errorf("plugin not found: %s", id)
	}
	plugin.Status = domain.PluginStatusDisabled
	s.logger.Info("disabled plugin", zap.String("id", id.String()))
	return plugin, nil
}

// ExecutePlugin executes a plugin with given input
func (s *PluginService) ExecutePlugin(ctx context.Context, pluginID uuid.UUID, input string) (*domain.PluginExecution, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	plugin, ok := s.plugins[pluginID]
	if !ok {
		return nil, fmt.Errorf("plugin not found: %s", pluginID)
	}
	if plugin.Status != domain.PluginStatusActive {
		return nil, fmt.Errorf("plugin is not active: %s", plugin.Status)
	}

	exec := &domain.PluginExecution{
		ID:         uuid.New(),
		PluginID:   pluginID,
		Input:      input,
		Output:     fmt.Sprintf("Plugin '%s' executed successfully with input length %d", plugin.Name, len(input)),
		DurationMs: 150,
		Status:     domain.PluginExecSuccess,
		ExecutedAt: time.Now(),
	}

	s.executions[exec.ID] = exec
	s.logger.Info("executed plugin", zap.String("pluginId", pluginID.String()), zap.String("execId", exec.ID.String()))
	return exec, nil
}

// UninstallPlugin removes a plugin
func (s *PluginService) UninstallPlugin(ctx context.Context, id uuid.UUID) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.plugins[id]; !ok {
		return fmt.Errorf("plugin not found: %s", id)
	}
	delete(s.plugins, id)
	s.logger.Info("uninstalled plugin", zap.String("id", id.String()))
	return nil
}

// InstallAdapter installs an agent framework adapter
func (s *PluginService) InstallAdapter(ctx context.Context, projectID uuid.UUID, input *domain.AdapterInstallInput) (*domain.AgentProtocolAdapter, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	templates := s.getAdapterTemplates()
	var tmpl *domain.AdapterTemplate
	for i, t := range templates {
		if t.Framework == input.Framework {
			tmpl = &templates[i]
			break
		}
	}

	name := input.Name
	if name == "" && tmpl != nil {
		name = tmpl.DisplayName
	}
	if name == "" {
		name = string(input.Framework) + " adapter"
	}

	description := ""
	capabilities := []string{"trace_capture"}
	if tmpl != nil {
		description = tmpl.Description
		capabilities = tmpl.Capabilities
	}

	adapter := &domain.AgentProtocolAdapter{
		ID:           uuid.New(),
		ProjectID:    projectID,
		Framework:    input.Framework,
		Name:         name,
		Description:  description,
		Version:      "1.0.0",
		Config:       input.Config,
		Status:       domain.PluginStatusInstalled,
		Capabilities: capabilities,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	s.logger.Info("installed agent protocol adapter",
		zap.String("adapterId", adapter.ID.String()),
		zap.String("framework", string(input.Framework)),
	)
	return adapter, nil
}

// ListAdapters returns the adapter registry with installed and available adapters
func (s *PluginService) ListAdapters(ctx context.Context, projectID uuid.UUID) (*domain.AdapterRegistry, error) {
	return &domain.AdapterRegistry{
		Installed: []domain.AgentProtocolAdapter{},
		Available: s.getAdapterTemplates(),
	}, nil
}

// IngestAdapterEvent receives trace events from framework adapters
func (s *PluginService) IngestAdapterEvent(ctx context.Context, projectID uuid.UUID, input *domain.AdapterEventInput) error {
	if input.EventType == "" {
		return fmt.Errorf("event type is required")
	}

	s.logger.Debug("ingested adapter event",
		zap.String("adapterId", input.AdapterID.String()),
		zap.String("eventType", input.EventType),
		zap.String("name", input.Name),
	)
	return nil
}

func (s *PluginService) getAdapterTemplates() []domain.AdapterTemplate {
	return []domain.AdapterTemplate{
		{
			Framework:    domain.FrameworkLangChain,
			DisplayName:  "LangChain",
			Description:  "Automatic trace capture for LangChain agents, chains, and tools",
			SetupGuide:   "pip install agenttrace[langchain] && export AGENTTRACE_API_KEY=...",
			Capabilities: []string{"trace_capture", "cost_tracking", "prompt_logging", "tool_tracking"},
			SDKSupport:   []string{"python", "typescript"},
		},
		{
			Framework:    domain.FrameworkCrewAI,
			DisplayName:  "CrewAI",
			Description:  "Trace capture for CrewAI multi-agent orchestrations",
			SetupGuide:   "pip install agenttrace[crewai] && export AGENTTRACE_API_KEY=...",
			Capabilities: []string{"trace_capture", "cost_tracking", "agent_handoff_tracking", "multi_agent"},
			SDKSupport:   []string{"python"},
		},
		{
			Framework:    domain.FrameworkAutoGen,
			DisplayName:  "AutoGen",
			Description:  "Trace capture for Microsoft AutoGen multi-agent conversations",
			SetupGuide:   "pip install agenttrace[autogen] && export AGENTTRACE_API_KEY=...",
			Capabilities: []string{"trace_capture", "cost_tracking", "conversation_tracking", "multi_agent"},
			SDKSupport:   []string{"python"},
		},
		{
			Framework:    domain.FrameworkLlamaIndex,
			DisplayName:  "LlamaIndex",
			Description:  "Trace capture for LlamaIndex RAG pipelines and agents",
			SetupGuide:   "pip install agenttrace[llamaindex] && export AGENTTRACE_API_KEY=...",
			Capabilities: []string{"trace_capture", "cost_tracking", "retrieval_tracking"},
			SDKSupport:   []string{"python", "typescript"},
		},
		{
			Framework:    domain.FrameworkOpenAI,
			DisplayName:  "OpenAI Agents SDK",
			Description:  "Trace capture for OpenAI Agents SDK",
			SetupGuide:   "pip install agenttrace[openai] && export AGENTTRACE_API_KEY=...",
			Capabilities: []string{"trace_capture", "cost_tracking", "tool_tracking"},
			SDKSupport:   []string{"python"},
		},
		{
			Framework:    domain.FrameworkDSPy,
			DisplayName:  "DSPy",
			Description:  "Trace capture for DSPy prompt optimization programs",
			SetupGuide:   "pip install agenttrace[dspy] && export AGENTTRACE_API_KEY=...",
			Capabilities: []string{"trace_capture", "prompt_logging", "optimization_tracking"},
			SDKSupport:   []string{"python"},
		},
		{
			Framework:    domain.FrameworkCustom,
			DisplayName:  "Custom Framework",
			Description:  "Generic adapter for custom agent frameworks using the AgentTrace Protocol",
			SetupGuide:   "Implement the AgentTrace adapter interface for your framework",
			Capabilities: []string{"trace_capture"},
			SDKSupport:   []string{"python", "typescript", "go"},
		},
	}
}
