package service

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

// PromptLibraryService handles prompt library logic
type PromptLibraryService struct {
	logger *zap.Logger
}

// NewPromptLibraryService creates a new prompt library service
func NewPromptLibraryService(logger *zap.Logger) *PromptLibraryService {
	return &PromptLibraryService{
		logger: logger,
	}
}

// CreatePrompt creates a new library prompt
func (s *PromptLibraryService) CreatePrompt(
	ctx context.Context,
	authorID uuid.UUID,
	authorName string,
	projectID *uuid.UUID,
	input *domain.LibraryPromptInput,
) (*domain.LibraryPrompt, error) {
	// Validate input
	if input.Name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if input.Template == "" {
		return nil, fmt.Errorf("template is required")
	}
	if input.Category == "" {
		return nil, fmt.Errorf("category is required")
	}

	// Generate slug
	slug := s.GenerateSlug(input.Name)

	// Set default visibility
	visibility := input.Visibility
	if visibility == "" {
		visibility = domain.PromptVisibilityPrivate
	}

	// Validate variables in template
	if err := s.ValidateTemplate(input.Template, input.Variables); err != nil {
		return nil, fmt.Errorf("template validation failed: %w", err)
	}

	now := time.Now()
	prompt := &domain.LibraryPrompt{
		ID:                uuid.New(),
		AuthorID:          authorID,
		AuthorName:        authorName,
		ProjectID:         projectID,
		Name:              input.Name,
		Slug:              slug,
		Description:       input.Description,
		Visibility:        visibility,
		Category:          input.Category,
		Tags:              s.normalizeTags(input.Tags),
		Template:          input.Template,
		Variables:         input.Variables,
		Examples:          input.Examples,
		RecommendedModels: input.RecommendedModels,
		ModelParams:       input.ModelParams,
		Version:           1,
		LatestVersion:     1,
		VersionNotes:      input.VersionNotes,
		ForkOf:            nil,
		ForkCount:         0,
		UsageCount:        0,
		StarCount:         0,
		ViewCount:         0,
		Benchmarks:        []domain.PromptBenchmark{},
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	s.logger.Info("Created library prompt",
		zap.String("promptId", prompt.ID.String()),
		zap.String("name", prompt.Name),
		zap.String("slug", prompt.Slug),
	)

	return prompt, nil
}

// UpdatePrompt updates a library prompt
func (s *PromptLibraryService) UpdatePrompt(
	ctx context.Context,
	prompt *domain.LibraryPrompt,
	input *domain.LibraryPromptUpdateInput,
) (*domain.LibraryPrompt, error) {
	if input.Name != nil {
		prompt.Name = *input.Name
		prompt.Slug = s.GenerateSlug(*input.Name)
	}
	if input.Description != nil {
		prompt.Description = *input.Description
	}
	if input.Visibility != nil {
		prompt.Visibility = *input.Visibility
	}
	if input.Category != nil {
		prompt.Category = *input.Category
	}
	if input.Tags != nil {
		prompt.Tags = s.normalizeTags(input.Tags)
	}
	if input.Template != nil {
		// Validate new template
		vars := prompt.Variables
		if input.Variables != nil {
			vars = input.Variables
		}
		if err := s.ValidateTemplate(*input.Template, vars); err != nil {
			return nil, fmt.Errorf("template validation failed: %w", err)
		}
		prompt.Template = *input.Template
	}
	if input.Variables != nil {
		prompt.Variables = input.Variables
	}
	if input.Examples != nil {
		prompt.Examples = input.Examples
	}
	if input.RecommendedModels != nil {
		prompt.RecommendedModels = input.RecommendedModels
	}
	if input.ModelParams != nil {
		prompt.ModelParams = input.ModelParams
	}

	// Bump version if requested and template changed
	if input.BumpVersion && input.Template != nil {
		prompt.Version++
		prompt.LatestVersion = prompt.Version
		prompt.VersionNotes = input.VersionNotes
	}

	prompt.UpdatedAt = time.Now()

	s.logger.Info("Updated library prompt",
		zap.String("promptId", prompt.ID.String()),
		zap.Int("version", prompt.Version),
	)

	return prompt, nil
}

// ForkPrompt creates a fork of an existing prompt
func (s *PromptLibraryService) ForkPrompt(
	ctx context.Context,
	sourcePrompt *domain.LibraryPrompt,
	forkerID uuid.UUID,
	forkerName string,
	projectID *uuid.UUID,
	input *domain.ForkInput,
) (*domain.LibraryPrompt, error) {
	name := sourcePrompt.Name
	if input != nil && input.Name != "" {
		name = input.Name
	}

	visibility := domain.PromptVisibilityPrivate
	if input != nil && input.Visibility != "" {
		visibility = input.Visibility
	}

	now := time.Now()
	fork := &domain.LibraryPrompt{
		ID:                uuid.New(),
		AuthorID:          forkerID,
		AuthorName:        forkerName,
		ProjectID:         projectID,
		Name:              name,
		Slug:              s.GenerateSlug(name),
		Description:       sourcePrompt.Description,
		Visibility:        visibility,
		Category:          sourcePrompt.Category,
		Tags:              append([]string{}, sourcePrompt.Tags...), // Copy tags
		Template:          sourcePrompt.Template,
		Variables:         append([]domain.PromptVariable{}, sourcePrompt.Variables...), // Copy variables
		Examples:          append([]domain.PromptExample{}, sourcePrompt.Examples...),   // Copy examples
		RecommendedModels: append([]string{}, sourcePrompt.RecommendedModels...),
		ModelParams:       copyMap(sourcePrompt.ModelParams),
		Version:           1,
		LatestVersion:     1,
		VersionNotes:      fmt.Sprintf("Forked from %s v%d", sourcePrompt.Name, sourcePrompt.Version),
		ForkOf:            &sourcePrompt.ID,
		ForkCount:         0,
		UsageCount:        0,
		StarCount:         0,
		ViewCount:         0,
		Benchmarks:        []domain.PromptBenchmark{},
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	// Record the fork
	_ = domain.PromptFork{
		ID:             uuid.New(),
		SourcePromptID: sourcePrompt.ID,
		SourceVersion:  sourcePrompt.Version,
		ForkedPromptID: fork.ID,
		ForkedBy:       forkerID,
		ForkedAt:       now,
	}

	s.logger.Info("Forked library prompt",
		zap.String("sourcePromptId", sourcePrompt.ID.String()),
		zap.String("forkedPromptId", fork.ID.String()),
		zap.String("forkedBy", forkerID.String()),
	)

	return fork, nil
}

// PublishPrompt makes a prompt publicly visible
func (s *PromptLibraryService) PublishPrompt(
	ctx context.Context,
	prompt *domain.LibraryPrompt,
) error {
	if prompt.Visibility == domain.PromptVisibilityPublic {
		return fmt.Errorf("prompt is already public")
	}

	prompt.Visibility = domain.PromptVisibilityPublic
	now := time.Now()
	prompt.PublishedAt = &now
	prompt.UpdatedAt = now

	s.logger.Info("Published library prompt",
		zap.String("promptId", prompt.ID.String()),
	)

	return nil
}

// StarPrompt adds a star to a prompt
func (s *PromptLibraryService) StarPrompt(
	ctx context.Context,
	promptID uuid.UUID,
	userID uuid.UUID,
) (*domain.PromptStar, error) {
	star := &domain.PromptStar{
		UserID:    userID,
		PromptID:  promptID,
		StarredAt: time.Now(),
	}

	s.logger.Debug("User starred prompt",
		zap.String("promptId", promptID.String()),
		zap.String("userId", userID.String()),
	)

	return star, nil
}

// UnstarPrompt removes a star from a prompt
func (s *PromptLibraryService) UnstarPrompt(
	ctx context.Context,
	promptID uuid.UUID,
	userID uuid.UUID,
) error {
	s.logger.Debug("User unstarred prompt",
		zap.String("promptId", promptID.String()),
		zap.String("userId", userID.String()),
	)

	return nil
}

// RecordUsage records that a prompt was used
func (s *PromptLibraryService) RecordUsage(
	ctx context.Context,
	promptID uuid.UUID,
	version int,
	userID uuid.UUID,
	projectID uuid.UUID,
	traceID *uuid.UUID,
) (*domain.PromptUsageRecord, error) {
	record := &domain.PromptUsageRecord{
		ID:        uuid.New(),
		PromptID:  promptID,
		Version:   version,
		UserID:    userID,
		ProjectID: projectID,
		TraceID:   traceID,
		UsedAt:    time.Now(),
	}

	s.logger.Debug("Recorded prompt usage",
		zap.String("promptId", promptID.String()),
		zap.Int("version", version),
	)

	return record, nil
}

// CreateVersion creates a new version of a prompt
func (s *PromptLibraryService) CreateVersion(
	ctx context.Context,
	prompt *domain.LibraryPrompt,
	createdBy uuid.UUID,
) (*domain.PromptVersion, error) {
	version := &domain.PromptVersion{
		PromptID:     prompt.ID,
		Version:      prompt.Version,
		Template:     prompt.Template,
		Variables:    prompt.Variables,
		VersionNotes: prompt.VersionNotes,
		CreatedAt:    time.Now(),
		CreatedBy:    createdBy,
	}

	s.logger.Debug("Created prompt version",
		zap.String("promptId", prompt.ID.String()),
		zap.Int("version", version.Version),
	)

	return version, nil
}

// RunBenchmark runs a benchmark for a prompt
func (s *PromptLibraryService) RunBenchmark(
	ctx context.Context,
	prompt *domain.LibraryPrompt,
	input *domain.BenchmarkInput,
	runBy uuid.UUID,
) (*domain.PromptBenchmark, error) {
	// In real implementation:
	// 1. Load dataset if provided
	// 2. Generate samples if no dataset
	// 3. Run prompt against each sample
	// 4. Run evaluators on results
	// 5. Calculate metrics

	startTime := time.Now()

	// Mock benchmark results
	benchmark := &domain.PromptBenchmark{
		ID:            uuid.New(),
		PromptID:      prompt.ID,
		PromptVersion: prompt.Version,
		Model:         input.Model,
		DatasetID:     input.DatasetID,
		SampleCount:   input.SampleCount,
		Metrics: domain.BenchmarkMetrics{
			AvgLatency:     500,
			P95Latency:     800,
			AvgTokens:      150,
			TotalCost:      0.05,
			AvgCostPerCall: 0.001,
			SuccessRate:    0.98,
			ErrorCount:     2,
		},
		RunBy:    runBy,
		RunAt:    startTime,
		Duration: int(time.Since(startTime).Seconds()),
	}

	s.logger.Info("Ran prompt benchmark",
		zap.String("promptId", prompt.ID.String()),
		zap.String("model", input.Model),
		zap.Int("sampleCount", input.SampleCount),
	)

	return benchmark, nil
}

// GenerateSlug generates a URL-friendly slug from a name
func (s *PromptLibraryService) GenerateSlug(name string) string {
	// Convert to lowercase
	slug := strings.ToLower(name)

	// Replace spaces with hyphens
	slug = strings.ReplaceAll(slug, " ", "-")

	// Remove special characters
	reg := regexp.MustCompile(`[^a-z0-9\-]`)
	slug = reg.ReplaceAllString(slug, "")

	// Remove multiple consecutive hyphens
	reg = regexp.MustCompile(`-+`)
	slug = reg.ReplaceAllString(slug, "-")

	// Trim hyphens from ends
	slug = strings.Trim(slug, "-")

	// Limit length
	if len(slug) > 100 {
		slug = slug[:100]
	}

	return slug
}

// ValidateTemplate validates that a template's variables are defined
func (s *PromptLibraryService) ValidateTemplate(template string, variables []domain.PromptVariable) error {
	// Extract variable names from template (e.g., {{variable_name}})
	reg := regexp.MustCompile(`\{\{([a-zA-Z_][a-zA-Z0-9_]*)\}\}`)
	matches := reg.FindAllStringSubmatch(template, -1)

	// Build set of defined variable names
	defined := make(map[string]bool)
	for _, v := range variables {
		defined[v.Name] = true
	}

	// Check all template variables are defined
	var undefined []string
	for _, match := range matches {
		if len(match) > 1 {
			varName := match[1]
			if !defined[varName] {
				undefined = append(undefined, varName)
			}
		}
	}

	if len(undefined) > 0 {
		return fmt.Errorf("undefined variables in template: %s", strings.Join(undefined, ", "))
	}

	return nil
}

// RenderTemplate renders a template with provided variables
func (s *PromptLibraryService) RenderTemplate(template string, variables map[string]any) (string, error) {
	result := template

	for name, value := range variables {
		placeholder := fmt.Sprintf("{{%s}}", name)
		valueStr := fmt.Sprintf("%v", value)
		result = strings.ReplaceAll(result, placeholder, valueStr)
	}

	// Check for unsubstituted variables
	reg := regexp.MustCompile(`\{\{([a-zA-Z_][a-zA-Z0-9_]*)\}\}`)
	if matches := reg.FindAllString(result, -1); len(matches) > 0 {
		return "", fmt.Errorf("unsubstituted variables: %s", strings.Join(matches, ", "))
	}

	return result, nil
}

// normalizeTags normalizes tag names
func (s *PromptLibraryService) normalizeTags(tags []string) []string {
	if tags == nil {
		return []string{}
	}

	normalized := make([]string, 0, len(tags))
	seen := make(map[string]bool)

	for _, tag := range tags {
		t := strings.ToLower(strings.TrimSpace(tag))
		if t != "" && !seen[t] {
			normalized = append(normalized, t)
			seen[t] = true
		}
	}

	return normalized
}

// GetPopularTags returns the most popular tags
func (s *PromptLibraryService) GetPopularTags(
	ctx context.Context,
	limit int,
) []TagCount {
	// In real implementation, query database for tag counts
	return []TagCount{
		{Tag: "agent", Count: 150},
		{Tag: "chat", Count: 120},
		{Tag: "code", Count: 100},
		{Tag: "summarization", Count: 80},
		{Tag: "extraction", Count: 60},
	}
}

// TagCount represents a tag and its usage count
type TagCount struct {
	Tag   string `json:"tag"`
	Count int    `json:"count"`
}

// copyMap creates a deep copy of a map
func copyMap(src map[string]any) map[string]any {
	if src == nil {
		return nil
	}
	dst := make(map[string]any, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}
