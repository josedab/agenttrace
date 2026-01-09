package domain

import (
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Prompt represents a managed prompt template
type Prompt struct {
	ID          uuid.UUID  `json:"id"`
	ProjectID   uuid.UUID  `json:"projectId"`
	Name        string     `json:"name"`
	Type        PromptType `json:"type"`
	Description string     `json:"description,omitempty"`
	Tags        []string   `json:"tags"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`

	// Latest version (populated by resolver)
	LatestVersion *PromptVersion `json:"latestVersion,omitempty"`
	// All versions (populated by resolver)
	Versions []PromptVersion `json:"versions,omitempty"`
}

// PromptVersion represents a version of a prompt
type PromptVersion struct {
	ID            uuid.UUID  `json:"id"`
	PromptID      uuid.UUID  `json:"promptId"`
	Version       int        `json:"version"`
	Content       string     `json:"content"`
	Config        string     `json:"config,omitempty"`
	Labels        []string   `json:"labels"`
	CreatedBy     *uuid.UUID `json:"createdBy,omitempty"`
	CommitMessage string     `json:"commitMessage,omitempty"`
	CreatedAt     time.Time  `json:"createdAt"`

	// Extracted variables
	Variables []string `json:"variables,omitempty"`
}

// PromptInput represents input for creating a prompt
type PromptInput struct {
	Name        string     `json:"name" validate:"required"`
	Type        PromptType `json:"type,omitempty"`
	Description *string    `json:"description,omitempty"`
	Tags        []string   `json:"tags,omitempty"`
	Content     string     `json:"content" validate:"required"`
	Config      any        `json:"config,omitempty"`
	Labels      []string   `json:"labels,omitempty"`
}

// PromptVersionInput represents input for creating a new prompt version
type PromptVersionInput struct {
	Content       string   `json:"content" validate:"required"`
	Config        any      `json:"config,omitempty"`
	Labels        []string `json:"labels,omitempty"`
	CommitMessage *string  `json:"commitMessage,omitempty"`
}

// PromptFilter represents filter options for querying prompts
type PromptFilter struct {
	ProjectID uuid.UUID
	Name      *string
	Tags      []string
	Label     *string
}

// PromptList represents a paginated list of prompts
type PromptList struct {
	Prompts    []Prompt `json:"prompts"`
	TotalCount int64    `json:"totalCount"`
	HasMore    bool     `json:"hasMore"`
}

// ChatMessage represents a message in a chat prompt
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// variableRegex matches {{variable}} patterns
var variableRegex = regexp.MustCompile(`\{\{([a-zA-Z_][a-zA-Z0-9_]*)\}\}`)

// ExtractVariables extracts variable names from prompt content
func ExtractVariables(content string) []string {
	matches := variableRegex.FindAllStringSubmatch(content, -1)
	seen := make(map[string]bool)
	var variables []string

	for _, match := range matches {
		if len(match) > 1 && !seen[match[1]] {
			seen[match[1]] = true
			variables = append(variables, match[1])
		}
	}

	return variables
}

// CompilePrompt compiles a prompt with variables
func CompilePrompt(content string, variables map[string]string) string {
	result := content
	for key, value := range variables {
		placeholder := "{{" + key + "}}"
		result = strings.ReplaceAll(result, placeholder, value)
	}
	return result
}

// CompileChatPrompt compiles a chat prompt with variables
func CompileChatPrompt(messages []ChatMessage, variables map[string]string) []ChatMessage {
	result := make([]ChatMessage, len(messages))
	for i, msg := range messages {
		result[i] = ChatMessage{
			Role:    msg.Role,
			Content: CompilePrompt(msg.Content, variables),
		}
	}
	return result
}

// ValidateVariables checks if all required variables are provided
func ValidateVariables(content string, variables map[string]string) []string {
	required := ExtractVariables(content)
	var missing []string

	for _, v := range required {
		if _, ok := variables[v]; !ok {
			missing = append(missing, v)
		}
	}

	return missing
}

// PromptWithVersion represents a prompt with a specific version
type PromptWithVersion struct {
	Prompt  *Prompt        `json:"prompt"`
	Version *PromptVersion `json:"version"`
}

// GetByLabel returns the prompt version with the specified label
func (p *Prompt) GetByLabel(label string) *PromptVersion {
	for i := range p.Versions {
		for _, l := range p.Versions[i].Labels {
			if l == label {
				return &p.Versions[i]
			}
		}
	}
	return nil
}

// GetByVersion returns the prompt version with the specified version number
func (p *Prompt) GetByVersion(version int) *PromptVersion {
	for i := range p.Versions {
		if p.Versions[i].Version == version {
			return &p.Versions[i]
		}
	}
	return nil
}
