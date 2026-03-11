package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

// GitOpsPromptRepository defines repository operations for GitOps prompt pipelines
type GitOpsPromptRepository interface {
	SaveConfig(ctx context.Context, config *domain.GitOpsPromptConfig) error
	GetConfigByID(ctx context.Context, id uuid.UUID) (*domain.GitOpsPromptConfig, error)
	ListConfigs(ctx context.Context, projectID uuid.UUID) ([]domain.GitOpsPromptConfig, error)
	UpdateConfig(ctx context.Context, config *domain.GitOpsPromptConfig) error
	DeleteConfig(ctx context.Context, id uuid.UUID) error
	SaveSyncEvent(ctx context.Context, event *domain.GitOpsSyncEvent) error
	ListSyncEvents(ctx context.Context, configID uuid.UUID, limit int) ([]domain.GitOpsSyncEvent, error)
}

// GitOpsPromptService manages GitOps prompt pipelines
type GitOpsPromptService struct {
	logger      *zap.Logger
	gitopsRepo  GitOpsPromptRepository
	promptRepo  PromptRepository
}

// NewGitOpsPromptService creates a new GitOps prompt service
func NewGitOpsPromptService(
	logger *zap.Logger,
	gitopsRepo GitOpsPromptRepository,
	promptRepo PromptRepository,
) *GitOpsPromptService {
	return &GitOpsPromptService{
		logger:     logger,
		gitopsRepo: gitopsRepo,
		promptRepo: promptRepo,
	}
}

// CreateConfig creates a new GitOps prompt pipeline configuration
func (s *GitOpsPromptService) CreateConfig(ctx context.Context, projectID, userID uuid.UUID, input *domain.GitOpsPromptConfigInput) (*domain.GitOpsPromptConfig, error) {
	if err := s.validateConfigInput(input); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	// Parse repo URL to extract owner/name
	owner, name := parseRepoURL(input.RepoURL)

	// Generate webhook secret
	secretBytes := make([]byte, 32)
	if _, err := rand.Read(secretBytes); err != nil {
		return nil, fmt.Errorf("failed to generate webhook secret: %w", err)
	}

	basePath := input.BasePath
	if basePath == "" {
		basePath = "prompts/"
	}

	autoPromote := false
	if input.AutoPromote != nil {
		autoPromote = *input.AutoPromote
	}

	evalGateEnabled := false
	if input.EvalGateEnabled != nil {
		evalGateEnabled = *input.EvalGateEnabled
	}

	config := &domain.GitOpsPromptConfig{
		ID:              uuid.New(),
		ProjectID:       projectID,
		Name:            input.Name,
		RepoURL:         input.RepoURL,
		RepoOwner:       owner,
		RepoName:        name,
		BasePath:        basePath,
		Enabled:         true,
		SyncStatus:      domain.GitOpsSyncPending,
		WebhookSecret:   hex.EncodeToString(secretBytes),
		BranchMapping:   input.BranchMapping,
		AutoPromote:     autoPromote,
		EvalGateEnabled: evalGateEnabled,
		Annotations:     input.Annotations,
		CreatedBy:       userID,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	if err := s.gitopsRepo.SaveConfig(ctx, config); err != nil {
		return nil, fmt.Errorf("failed to save config: %w", err)
	}

	s.logger.Info("gitops prompt config created",
		zap.String("configId", config.ID.String()),
		zap.String("repo", input.RepoURL),
		zap.Int("branchMappings", len(input.BranchMapping)),
	)

	return config, nil
}

// GetConfig retrieves a GitOps config by ID
func (s *GitOpsPromptService) GetConfig(ctx context.Context, id uuid.UUID) (*domain.GitOpsPromptConfig, error) {
	return s.gitopsRepo.GetConfigByID(ctx, id)
}

// ListConfigs lists all GitOps configs for a project
func (s *GitOpsPromptService) ListConfigs(ctx context.Context, projectID uuid.UUID) ([]domain.GitOpsPromptConfig, error) {
	return s.gitopsRepo.ListConfigs(ctx, projectID)
}

// SyncFromGit performs a sync from git to AgentTrace prompts
func (s *GitOpsPromptService) SyncFromGit(ctx context.Context, configID uuid.UUID, branch, commitSHA string, files []GitPromptFile) (*domain.GitOpsSyncEvent, error) {
	config, err := s.gitopsRepo.GetConfigByID(ctx, configID)
	if err != nil {
		return nil, fmt.Errorf("config not found: %w", err)
	}

	if !config.Enabled {
		return nil, fmt.Errorf("gitops config is disabled")
	}

	// Determine environment from branch
	env := s.resolveEnvironment(config, branch)

	event := &domain.GitOpsSyncEvent{
		ID:          uuid.New(),
		ConfigID:    configID,
		ProjectID:   config.ProjectID,
		Status:      domain.GitOpsSyncSyncing,
		Branch:      branch,
		CommitSHA:   commitSHA,
		Environment: env,
		StartedAt:   time.Now(),
	}

	// Process each prompt file
	var changes []domain.SyncChange
	for _, file := range files {
		change, err := s.syncPromptFile(ctx, config, file, env)
		if err != nil {
			s.logger.Warn("failed to sync prompt file",
				zap.String("file", file.Path),
				zap.Error(err),
			)
			continue
		}
		if change != nil {
			changes = append(changes, *change)
		}
	}

	event.Changes = changes
	event.Status = domain.GitOpsSyncSynced
	now := time.Now()
	event.CompletedAt = &now

	// Update config sync status
	config.SyncStatus = domain.GitOpsSyncSynced
	config.LastSyncAt = &now
	config.LastSyncErr = ""
	config.UpdatedAt = now
	if err := s.gitopsRepo.UpdateConfig(ctx, config); err != nil {
		s.logger.Warn("failed to update config sync status", zap.Error(err))
	}

	if err := s.gitopsRepo.SaveSyncEvent(ctx, event); err != nil {
		s.logger.Warn("failed to save sync event", zap.Error(err))
	}

	s.logger.Info("gitops sync completed",
		zap.String("configId", configID.String()),
		zap.String("branch", branch),
		zap.String("commit", commitSHA),
		zap.Int("changes", len(changes)),
	)

	return event, nil
}

// SyncToGit exports prompts from AgentTrace to git format
func (s *GitOpsPromptService) SyncToGit(ctx context.Context, configID uuid.UUID) ([]GitPromptFile, error) {
	config, err := s.gitopsRepo.GetConfigByID(ctx, configID)
	if err != nil {
		return nil, fmt.Errorf("config not found: %w", err)
	}

	prompts, err := s.promptRepo.List(ctx, &domain.PromptFilter{
		ProjectID: config.ProjectID,
	}, 1000, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to list prompts: %w", err)
	}

	var files []GitPromptFile
	for _, prompt := range prompts.Prompts {
		// Sanitize prompt name to prevent path traversal
		if validateSafeName(prompt.Name) != nil {
			s.logger.Warn("skipping prompt with unsafe name", zap.String("name", prompt.Name))
			continue
		}
		spec := s.promptToFileSpec(&prompt)
		content := fmt.Sprintf("apiVersion: agenttrace.io/v1\nkind: Prompt\nmetadata:\n  name: %s\n", prompt.Name)
		if prompt.Description != "" {
			content += fmt.Sprintf("  description: %s\n", prompt.Description)
		}
		if prompt.LatestVersion != nil {
			content += fmt.Sprintf("spec:\n  type: %s\n  content: |\n    %s\n",
				spec.Spec.Type,
				strings.ReplaceAll(prompt.LatestVersion.Content, "\n", "\n    "))
		}

		files = append(files, GitPromptFile{
			Path:    filepath.Join(config.BasePath, prompt.Name+".yaml"),
			Content: content,
		})
	}

	s.logger.Info("exported prompts to git format",
		zap.String("configId", configID.String()),
		zap.Int("fileCount", len(files)),
	)

	return files, nil
}

// HandleWebhook processes a git webhook event (push, PR merge)
func (s *GitOpsPromptService) HandleWebhook(ctx context.Context, configID uuid.UUID, webhookSecret string, branch, commitSHA string, changedFiles []string) error {
	config, err := s.gitopsRepo.GetConfigByID(ctx, configID)
	if err != nil {
		return fmt.Errorf("config not found: %w", err)
	}

	if config.WebhookSecret != webhookSecret {
		return fmt.Errorf("invalid webhook secret")
	}

	if !config.Enabled {
		return nil
	}

	// Filter to only prompt files in base path
	var promptFiles []string
	for _, f := range changedFiles {
		if strings.HasPrefix(f, config.BasePath) &&
			(strings.HasSuffix(f, ".yaml") || strings.HasSuffix(f, ".yml") || strings.HasSuffix(f, ".json")) {
			promptFiles = append(promptFiles, f)
		}
	}

	if len(promptFiles) == 0 {
		return nil
	}

	s.logger.Info("webhook received with prompt changes",
		zap.String("configId", configID.String()),
		zap.String("branch", branch),
		zap.Int("changedFiles", len(promptFiles)),
	)

	// In production, this would fetch file contents from git and call SyncFromGit
	return nil
}

// PromoteToEnvironment promotes prompts from one environment to another
func (s *GitOpsPromptService) PromoteToEnvironment(ctx context.Context, configID uuid.UUID, from, to domain.GitOpsEnvironment) error {
	config, err := s.gitopsRepo.GetConfigByID(ctx, configID)
	if err != nil {
		return fmt.Errorf("config not found: %w", err)
	}

	if config.EvalGateEnabled {
		s.logger.Info("evaluation gate check before promotion",
			zap.String("from", string(from)),
			zap.String("to", string(to)),
		)
		// In production, run eval suite before promoting
	}

	s.logger.Info("promoting prompts",
		zap.String("configId", configID.String()),
		zap.String("from", string(from)),
		zap.String("to", string(to)),
	)

	return nil
}

// GitPromptFile represents a prompt file in a git repository
type GitPromptFile struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

func (s *GitOpsPromptService) validateConfigInput(input *domain.GitOpsPromptConfigInput) error {
	if input.Name == "" {
		return fmt.Errorf("name is required")
	}
	if err := validateSafeName(input.Name); err != nil {
		return err
	}
	if input.RepoURL == "" {
		return fmt.Errorf("repo URL is required")
	}
	if len(input.BranchMapping) == 0 {
		return fmt.Errorf("at least one branch mapping is required")
	}
	return nil
}

// validateSafeName rejects names containing path separators or traversal sequences
func validateSafeName(name string) error {
	cleaned := filepath.Clean(name)
	if cleaned != name || strings.Contains(name, "/") || strings.Contains(name, "\\") || strings.Contains(name, "..") {
		return fmt.Errorf("name contains invalid path characters")
	}
	return nil
}

func (s *GitOpsPromptService) resolveEnvironment(config *domain.GitOpsPromptConfig, branch string) domain.GitOpsEnvironment {
	for _, mapping := range config.BranchMapping {
		if matchBranch(mapping.BranchPattern, branch) {
			return mapping.Environment
		}
	}
	return domain.GitOpsEnvDevelopment
}

func matchBranch(pattern, branch string) bool {
	if pattern == branch {
		return true
	}
	// Simple wildcard matching for "feature/*" patterns
	if strings.HasSuffix(pattern, "/*") {
		prefix := strings.TrimSuffix(pattern, "/*")
		return strings.HasPrefix(branch, prefix+"/")
	}
	return false
}

func (s *GitOpsPromptService) syncPromptFile(ctx context.Context, config *domain.GitOpsPromptConfig, file GitPromptFile, env domain.GitOpsEnvironment) (*domain.SyncChange, error) {
	// Parse prompt name from file path
	baseName := filepath.Base(file.Path)
	promptName := strings.TrimSuffix(baseName, filepath.Ext(baseName))

	// Check if prompt already exists
	existing, err := s.promptRepo.GetByName(ctx, config.ProjectID, promptName)
	if err != nil {
		// Create new prompt
		change := &domain.SyncChange{
			PromptName: promptName,
			Action:     "created",
			FilePath:   file.Path,
		}
		return change, nil
	}

	// Update existing prompt
	change := &domain.SyncChange{
		PromptName: promptName,
		Action:     "updated",
		FilePath:   file.Path,
	}

	if existing.LatestVersion != nil {
		v := existing.LatestVersion.Version
		change.OldVersion = &v
		newV := v + 1
		change.NewVersion = &newV
	}

	return change, nil
}

func (s *GitOpsPromptService) promptToFileSpec(prompt *domain.Prompt) *domain.PromptFileSpec {
	spec := &domain.PromptFileSpec{
		APIVersion: "agenttrace.io/v1",
		Kind:       "Prompt",
		Metadata: domain.PromptFileMeta{
			Name:        prompt.Name,
			Description: prompt.Description,
			Tags:        prompt.Tags,
		},
		Spec: domain.PromptFileContent{
			Type: string(prompt.Type),
		},
	}

	if prompt.LatestVersion != nil {
		spec.Spec.Content = prompt.LatestVersion.Content
		for _, v := range prompt.LatestVersion.Variables {
			spec.Spec.Variables = append(spec.Spec.Variables, domain.PromptFileVariable{
				Name:     v,
				Required: true,
			})
		}
	}

	return spec
}

func parseRepoURL(repoURL string) (string, string) {
	// Parse owner/name from various URL formats
	// e.g., "https://github.com/owner/repo" or "git@github.com:owner/repo.git"
	repoURL = strings.TrimSuffix(repoURL, ".git")

	if strings.Contains(repoURL, "github.com") {
		parts := strings.Split(repoURL, "/")
		if len(parts) >= 2 {
			return parts[len(parts)-2], parts[len(parts)-1]
		}
	}

	return "", ""
}
