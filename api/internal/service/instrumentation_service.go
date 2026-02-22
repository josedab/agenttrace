package service

import (
	"context"
	"fmt"
	"strings"

	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

// InstrumentationService provides setup instructions and validates
// instrumentation configurations for supported agent frameworks.
type InstrumentationService struct {
	logger *zap.Logger
}

// NewInstrumentationService creates a new instrumentation service
func NewInstrumentationService(logger *zap.Logger) *InstrumentationService {
	return &InstrumentationService{
		logger: logger,
	}
}

// GetSetupInstructions returns framework-specific setup instructions for the
// given language.
func (s *InstrumentationService) GetSetupInstructions(framework, language string) (*domain.InstrumentationSetup, error) {
	fw := domain.InstrumentationFramework(strings.ToUpper(framework))
	if !fw.IsValid() {
		return nil, fmt.Errorf("unsupported instrumentation framework: %s", framework)
	}

	lang := strings.ToLower(language)

	setup := &domain.InstrumentationSetup{
		Framework: fw,
		Language:  lang,
	}

	switch fw {
	case domain.InstrumentationFrameworkCrewAI:
		setup.InstallCommand = "pip install agenttrace[crewai]"
		setup.ConfigSnippet = `from agenttrace import configure\nconfigure(framework="crewai")`
		setup.ExampleCode = `from crewai import Agent\n# Traces are captured automatically`
	case domain.InstrumentationFrameworkAutogen:
		setup.InstallCommand = "pip install agenttrace[autogen]"
		setup.ConfigSnippet = `from agenttrace import configure\nconfigure(framework="autogen")`
		setup.ExampleCode = `import autogen\n# Traces are captured automatically`
	case domain.InstrumentationFrameworkLangGraph:
		setup.InstallCommand = "pip install agenttrace[langgraph]"
		setup.ConfigSnippet = `from agenttrace import configure\nconfigure(framework="langgraph")`
		setup.ExampleCode = `from langgraph.graph import StateGraph\n# Traces are captured automatically`
	case domain.InstrumentationFrameworkLlamaIndex:
		setup.InstallCommand = "pip install agenttrace[llamaindex]"
		setup.ConfigSnippet = `from agenttrace import configure\nconfigure(framework="llamaindex")`
		setup.ExampleCode = `from llama_index import VectorStoreIndex\n# Traces are captured automatically`
	case domain.InstrumentationFrameworkSemanticKernel:
		if lang == "csharp" || lang == "dotnet" {
			setup.InstallCommand = "dotnet add package AgentTrace.SemanticKernel"
			setup.ConfigSnippet = `builder.Services.AddAgentTrace(o => o.Framework = "semantic_kernel");`
			setup.ExampleCode = `var kernel = Kernel.CreateBuilder().Build();\n// Traces are captured automatically`
		} else {
			setup.InstallCommand = "pip install agenttrace[semantic-kernel]"
			setup.ConfigSnippet = `from agenttrace import configure\nconfigure(framework="semantic_kernel")`
			setup.ExampleCode = `import semantic_kernel as sk\n# Traces are captured automatically`
		}
	}

	s.logger.Debug("returned setup instructions",
		zap.String("framework", string(fw)),
		zap.String("language", lang),
	)

	return setup, nil
}

// ListFrameworks returns all supported instrumentation frameworks.
func (s *InstrumentationService) ListFrameworks() []domain.InstrumentationFramework {
	return []domain.InstrumentationFramework{
		domain.InstrumentationFrameworkCrewAI,
		domain.InstrumentationFrameworkAutogen,
		domain.InstrumentationFrameworkLangGraph,
		domain.InstrumentationFrameworkLlamaIndex,
		domain.InstrumentationFrameworkSemanticKernel,
	}
}

// ValidateConfig validates an instrumentation configuration and returns any
// issues found.
func (s *InstrumentationService) ValidateConfig(ctx context.Context, config domain.InstrumentationConfig) (bool, []string) {
	var issues []string

	if !config.Framework.IsValid() {
		issues = append(issues, fmt.Sprintf("unsupported framework: %s", config.Framework))
	}

	if config.ProjectID.String() == "00000000-0000-0000-0000-000000000000" {
		issues = append(issues, "projectId is required")
	}

	if !config.Enabled {
		issues = append(issues, "instrumentation is disabled")
	}

	valid := len(issues) == 0

	s.logger.Debug("validated instrumentation config",
		zap.String("framework", string(config.Framework)),
		zap.Bool("valid", valid),
		zap.Int("issueCount", len(issues)),
	)

	return valid, issues
}
