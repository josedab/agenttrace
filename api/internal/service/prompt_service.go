package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	apperrors "github.com/agenttrace/agenttrace/api/internal/pkg/errors"
)

// PromptRepository defines prompt repository operations
type PromptRepository interface {
	Create(ctx context.Context, prompt *domain.Prompt) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Prompt, error)
	GetByName(ctx context.Context, projectID uuid.UUID, name string) (*domain.Prompt, error)
	Update(ctx context.Context, prompt *domain.Prompt) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, filter *domain.PromptFilter, limit, offset int) (*domain.PromptList, error)
	NameExists(ctx context.Context, projectID uuid.UUID, name string) (bool, error)

	// Version operations
	CreateVersion(ctx context.Context, version *domain.PromptVersion) error
	GetVersion(ctx context.Context, promptID uuid.UUID, version int) (*domain.PromptVersion, error)
	GetLatestVersion(ctx context.Context, promptID uuid.UUID) (*domain.PromptVersion, error)
	GetVersionByLabel(ctx context.Context, promptID uuid.UUID, label string) (*domain.PromptVersion, error)
	ListVersions(ctx context.Context, promptID uuid.UUID) ([]domain.PromptVersion, error)
	UpdateVersionLabels(ctx context.Context, versionID uuid.UUID, labels []string) error
}

// PromptService handles prompt management operations
type PromptService struct {
	promptRepo PromptRepository
}

// NewPromptService creates a new prompt service
func NewPromptService(promptRepo PromptRepository) *PromptService {
	return &PromptService{
		promptRepo: promptRepo,
	}
}

// Create creates a new prompt with an initial version
func (s *PromptService) Create(ctx context.Context, projectID uuid.UUID, input *domain.PromptInput, userID uuid.UUID) (*domain.Prompt, error) {
	// Check if name exists
	exists, err := s.promptRepo.NameExists(ctx, projectID, input.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to check name: %w", err)
	}
	if exists {
		return nil, apperrors.Validation("prompt name already exists")
	}

	now := time.Now()

	// Determine type
	promptType := domain.PromptTypeText
	if input.Type != "" {
		promptType = input.Type
	}

	prompt := &domain.Prompt{
		ID:        uuid.New(),
		ProjectID: projectID,
		Name:      input.Name,
		Type:      promptType,
		Tags:      input.Tags,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if input.Description != nil {
		prompt.Description = *input.Description
	}

	if err := s.promptRepo.Create(ctx, prompt); err != nil {
		return nil, fmt.Errorf("failed to create prompt: %w", err)
	}

	// Create initial version if content provided
	if input.Content != "" {
		initialMessage := "Initial version"
		versionInput := &domain.PromptVersionInput{
			Content:       input.Content,
			CommitMessage: &initialMessage,
		}

		if input.Config != nil {
			configJSON, _ := json.Marshal(input.Config)
			versionInput.Config = string(configJSON)
		}

		version, err := s.CreateVersion(ctx, prompt.ID, versionInput, userID)
		if err != nil {
			return nil, fmt.Errorf("failed to create initial version: %w", err)
		}

		prompt.LatestVersion = version
	}

	return prompt, nil
}

// Get retrieves a prompt by ID
func (s *PromptService) Get(ctx context.Context, id uuid.UUID) (*domain.Prompt, error) {
	prompt, err := s.promptRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Load latest version
	latestVersion, err := s.promptRepo.GetLatestVersion(ctx, id)
	if err != nil && !apperrors.IsNotFound(err) {
		return nil, fmt.Errorf("failed to get latest version: %w", err)
	}
	prompt.LatestVersion = latestVersion

	return prompt, nil
}

// GetByName retrieves a prompt by name
func (s *PromptService) GetByName(ctx context.Context, projectID uuid.UUID, name string) (*domain.Prompt, error) {
	prompt, err := s.promptRepo.GetByName(ctx, projectID, name)
	if err != nil {
		return nil, err
	}

	// Load latest version
	latestVersion, err := s.promptRepo.GetLatestVersion(ctx, prompt.ID)
	if err != nil && !apperrors.IsNotFound(err) {
		return nil, fmt.Errorf("failed to get latest version: %w", err)
	}
	prompt.LatestVersion = latestVersion

	return prompt, nil
}

// GetByNameAndVersion retrieves a specific version of a prompt
func (s *PromptService) GetByNameAndVersion(ctx context.Context, projectID uuid.UUID, name string, version int) (*domain.Prompt, error) {
	prompt, err := s.promptRepo.GetByName(ctx, projectID, name)
	if err != nil {
		return nil, err
	}

	promptVersion, err := s.promptRepo.GetVersion(ctx, prompt.ID, version)
	if err != nil {
		return nil, err
	}
	prompt.LatestVersion = promptVersion

	return prompt, nil
}

// GetByNameAndLabel retrieves a prompt version by label (e.g., "production")
func (s *PromptService) GetByNameAndLabel(ctx context.Context, projectID uuid.UUID, name, label string) (*domain.Prompt, error) {
	prompt, err := s.promptRepo.GetByName(ctx, projectID, name)
	if err != nil {
		return nil, err
	}

	promptVersion, err := s.promptRepo.GetVersionByLabel(ctx, prompt.ID, label)
	if err != nil {
		return nil, err
	}
	prompt.LatestVersion = promptVersion

	return prompt, nil
}

// Update updates a prompt
func (s *PromptService) Update(ctx context.Context, id uuid.UUID, input *domain.PromptInput) (*domain.Prompt, error) {
	prompt, err := s.promptRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Check if new name conflicts
	if input.Name != "" && input.Name != prompt.Name {
		exists, err := s.promptRepo.NameExists(ctx, prompt.ProjectID, input.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to check name: %w", err)
		}
		if exists {
			return nil, apperrors.Validation("prompt name already exists")
		}
		prompt.Name = input.Name
	}

	if input.Description != nil {
		prompt.Description = *input.Description
	}
	if input.Type != "" {
		prompt.Type = input.Type
	}
	if len(input.Tags) > 0 {
		prompt.Tags = input.Tags
	}

	prompt.UpdatedAt = time.Now()

	if err := s.promptRepo.Update(ctx, prompt); err != nil {
		return nil, fmt.Errorf("failed to update prompt: %w", err)
	}

	return prompt, nil
}

// Delete deletes a prompt
func (s *PromptService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.promptRepo.Delete(ctx, id)
}

// List retrieves prompts with filtering
func (s *PromptService) List(ctx context.Context, filter *domain.PromptFilter, limit, offset int) (*domain.PromptList, error) {
	return s.promptRepo.List(ctx, filter, limit, offset)
}

// CreateVersion creates a new version of a prompt
func (s *PromptService) CreateVersion(ctx context.Context, promptID uuid.UUID, input *domain.PromptVersionInput, userID uuid.UUID) (*domain.PromptVersion, error) {
	// Verify prompt exists
	_, err := s.promptRepo.GetByID(ctx, promptID)
	if err != nil {
		return nil, err
	}

	now := time.Now()

	// Convert config to string
	var configStr string
	if input.Config != nil {
		configBytes, _ := json.Marshal(input.Config)
		configStr = string(configBytes)
	}

	// Handle optional commit message
	var commitMessage string
	if input.CommitMessage != nil {
		commitMessage = *input.CommitMessage
	}

	version := &domain.PromptVersion{
		ID:            uuid.New(),
		PromptID:      promptID,
		Content:       input.Content,
		Config:        configStr,
		Labels:        input.Labels,
		CreatedBy:     &userID,
		CommitMessage: commitMessage,
		CreatedAt:     now,
	}

	// Extract variables from content
	version.Variables = domain.ExtractVariables(input.Content)

	if err := s.promptRepo.CreateVersion(ctx, version); err != nil {
		return nil, fmt.Errorf("failed to create version: %w", err)
	}

	return version, nil
}

// GetVersion retrieves a specific version
func (s *PromptService) GetVersion(ctx context.Context, promptID uuid.UUID, version int) (*domain.PromptVersion, error) {
	return s.promptRepo.GetVersion(ctx, promptID, version)
}

// ListVersions retrieves all versions of a prompt
func (s *PromptService) ListVersions(ctx context.Context, promptID uuid.UUID) ([]domain.PromptVersion, error) {
	return s.promptRepo.ListVersions(ctx, promptID)
}

// SetVersionLabel adds or removes a label from a version
func (s *PromptService) SetVersionLabel(ctx context.Context, promptID uuid.UUID, version int, label string, add bool) error {
	promptVersion, err := s.promptRepo.GetVersion(ctx, promptID, version)
	if err != nil {
		return err
	}

	labels := promptVersion.Labels

	if add {
		// Remove label from other versions first (labels are unique per prompt)
		versions, err := s.promptRepo.ListVersions(ctx, promptID)
		if err != nil {
			return err
		}

		for _, v := range versions {
			if v.ID == promptVersion.ID {
				continue
			}
			newLabels := make([]string, 0)
			for _, l := range v.Labels {
				if l != label {
					newLabels = append(newLabels, l)
				}
			}
			if len(newLabels) != len(v.Labels) {
				if err := s.promptRepo.UpdateVersionLabels(ctx, v.ID, newLabels); err != nil {
					return err
				}
			}
		}

		// Add label to current version
		found := false
		for _, l := range labels {
			if l == label {
				found = true
				break
			}
		}
		if !found {
			labels = append(labels, label)
		}
	} else {
		// Remove label from current version
		newLabels := make([]string, 0)
		for _, l := range labels {
			if l != label {
				newLabels = append(newLabels, l)
			}
		}
		labels = newLabels
	}

	return s.promptRepo.UpdateVersionLabels(ctx, promptVersion.ID, labels)
}

// Compile compiles a prompt with variables
func (s *PromptService) Compile(ctx context.Context, projectID uuid.UUID, name string, variables map[string]string, options *CompileOptions) (*CompiledPrompt, error) {
	var prompt *domain.Prompt
	var err error

	if options != nil && options.Version != nil {
		prompt, err = s.GetByNameAndVersion(ctx, projectID, name, *options.Version)
	} else if options != nil && options.Label != "" {
		prompt, err = s.GetByNameAndLabel(ctx, projectID, name, options.Label)
	} else {
		prompt, err = s.GetByName(ctx, projectID, name)
	}

	if err != nil {
		return nil, err
	}

	if prompt.LatestVersion == nil {
		return nil, apperrors.NotFound("prompt version")
	}

	// Validate variables
	if missingVars := domain.ValidateVariables(prompt.LatestVersion.Content, variables); len(missingVars) > 0 {
		return nil, apperrors.Validation(fmt.Sprintf("missing variables: %v", missingVars))
	}

	// Compile content
	compiled := domain.CompilePrompt(prompt.LatestVersion.Content, variables)

	return &CompiledPrompt{
		Prompt:    prompt,
		Version:   prompt.LatestVersion.Version,
		Compiled:  compiled,
		Variables: variables,
	}, nil
}

// CompileOptions represents options for prompt compilation
type CompileOptions struct {
	Version *int
	Label   string
}

// CompiledPrompt represents a compiled prompt ready for use
type CompiledPrompt struct {
	Prompt    *domain.Prompt         `json:"prompt"`
	Version   int                    `json:"version"`
	Compiled  string                 `json:"compiled"`
	Variables map[string]string      `json:"variables"`
}
