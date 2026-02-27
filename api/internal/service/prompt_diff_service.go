package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

// PromptDiffService handles prompt version diffing and impact analysis
type PromptDiffService struct {
	logger     *zap.Logger
	promptRepo PromptRepository
}

// NewPromptDiffService creates a new prompt diff service
func NewPromptDiffService(logger *zap.Logger, promptRepo PromptRepository) *PromptDiffService {
	return &PromptDiffService{
		logger:     logger,
		promptRepo: promptRepo,
	}
}

// DiffVersions computes a semantic diff between two prompt versions
func (s *PromptDiffService) DiffVersions(ctx context.Context, promptID uuid.UUID, versionA, versionB int) (*domain.PromptVersionDiff, error) {
	verA, err := s.promptRepo.GetVersion(ctx, promptID, versionA)
	if err != nil {
		return nil, fmt.Errorf("failed to get version %d: %w", versionA, err)
	}

	verB, err := s.promptRepo.GetVersion(ctx, promptID, versionB)
	if err != nil {
		return nil, fmt.Errorf("failed to get version %d: %w", versionB, err)
	}

	// Compute content diff
	contentDiff := s.computeSemanticDiff(verA.Content, verB.Content)

	// Compute variable diff
	varsA := domain.ExtractVariables(verA.Content)
	varsB := domain.ExtractVariables(verB.Content)
	variableDiff := s.computeVariableDiff(varsA, varsB)

	// Compute config diff
	var configDiff *domain.ConfigDiff
	if verA.Config != "" || verB.Config != "" {
		configDiff = s.computeConfigDiff(verA.Config, verB.Config)
	}

	// Categorize changes
	changes := s.categorizeChanges(contentDiff, variableDiff, configDiff)

	// Compute summary
	summary := s.computeDiffSummary(changes)

	diff := &domain.PromptVersionDiff{
		PromptID:     promptID,
		VersionA:     versionA,
		VersionB:     versionB,
		ContentDiff:  *contentDiff,
		ConfigDiff:   configDiff,
		VariableDiff: *variableDiff,
		Changes:      changes,
		Summary:      *summary,
	}

	s.logger.Info("computed prompt version diff",
		zap.String("promptId", promptID.String()),
		zap.Int("versionA", versionA),
		zap.Int("versionB", versionB),
		zap.Int("totalChanges", summary.TotalChanges),
	)

	return diff, nil
}

// AnalyzeImpact analyzes the impact of a prompt version change using trace metrics
func (s *PromptDiffService) AnalyzeImpact(ctx context.Context, promptID uuid.UUID, versionBefore, versionAfter int) (*domain.PromptImpactAnalysis, error) {
	// Placeholder: in production, this would query ClickHouse trace data
	// correlated with prompt version usage
	analysis := &domain.PromptImpactAnalysis{
		PromptID:      promptID,
		VersionBefore: versionBefore,
		VersionAfter:  versionAfter,
		Metrics: domain.PromptImpactMetrics{
			Before: domain.PromptMetricsSnapshot{},
			After:  domain.PromptMetricsSnapshot{},
			Deltas: domain.PromptMetricsDeltas{},
		},
		TraceComparison: domain.PromptTraceComparison{
			TopChanges: []string{},
		},
		Recommendation: "Insufficient trace data for impact analysis. Deploy the new version to staging first.",
		AnalyzedAt:     time.Now(),
	}

	s.logger.Info("analyzed prompt version impact",
		zap.String("promptId", promptID.String()),
		zap.Int("before", versionBefore),
		zap.Int("after", versionAfter),
	)

	return analysis, nil
}

// ReviewVersion creates or updates a review for a prompt version
func (s *PromptDiffService) ReviewVersion(ctx context.Context, promptID, versionID, reviewerID uuid.UUID, version int, input *domain.PromptVersionReviewInput) (*domain.PromptVersionReview, error) {
	review := &domain.PromptVersionReview{
		ID:         uuid.New(),
		PromptID:   promptID,
		VersionID:  versionID,
		Version:    version,
		Status:     input.Status,
		ReviewerID: reviewerID,
		Comment:    input.Comment,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	s.logger.Info("prompt version reviewed",
		zap.String("promptId", promptID.String()),
		zap.Int("version", version),
		zap.String("status", string(input.Status)),
	)

	return review, nil
}

// RollbackVersion creates a new version that reverts to a previous version's content
func (s *PromptDiffService) RollbackVersion(ctx context.Context, promptID uuid.UUID, input *domain.PromptRollbackInput, userID uuid.UUID) (*domain.PromptVersion, error) {
	targetVer, err := s.promptRepo.GetVersion(ctx, promptID, input.TargetVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to get target version %d: %w", input.TargetVersion, err)
	}

	commitMsg := fmt.Sprintf("Rollback to version %d", input.TargetVersion)
	if input.Reason != "" {
		commitMsg += ": " + input.Reason
	}

	versionInput := &domain.PromptVersionInput{
		Content:       targetVer.Content,
		CommitMessage: &commitMsg,
		Labels:        targetVer.Labels,
	}

	if targetVer.Config != "" {
		var configObj interface{}
		if err := json.Unmarshal([]byte(targetVer.Config), &configObj); err == nil {
			versionInput.Config = configObj
		}
	}

	ps := NewPromptService(s.promptRepo)
	newVersion, err := ps.CreateVersion(ctx, promptID, versionInput, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to create rollback version: %w", err)
	}

	s.logger.Info("prompt version rolled back",
		zap.String("promptId", promptID.String()),
		zap.Int("targetVersion", input.TargetVersion),
		zap.Int("newVersion", newVersion.Version),
	)

	return newVersion, nil
}

func (s *PromptDiffService) computeSemanticDiff(contentA, contentB string) *domain.SemanticDiff {
	linesA := strings.Split(contentA, "\n")
	linesB := strings.Split(contentB, "\n")

	diff := &domain.SemanticDiff{
		TotalLines: max(len(linesA), len(linesB)),
	}

	// Simple line-by-line diff using LCS approach
	hunks := s.computeLineDiff(linesA, linesB)
	diff.Hunks = hunks

	for _, hunk := range hunks {
		for _, line := range hunk.Lines {
			switch line.Type {
			case domain.PromptDiffTypeAdded:
				diff.AddedLines++
			case domain.PromptDiffTypeRemoved:
				diff.RemovedLines++
			}
		}
	}

	return diff
}

func (s *PromptDiffService) computeLineDiff(linesA, linesB []string) []domain.DiffHunk {
	var hunks []domain.DiffHunk
	var currentHunk *domain.DiffHunk

	maxLen := max(len(linesA), len(linesB))
	contextLines := 3

	for i := 0; i < maxLen; i++ {
		var lineA, lineB string
		if i < len(linesA) {
			lineA = linesA[i]
		}
		if i < len(linesB) {
			lineB = linesB[i]
		}

		if i < len(linesA) && i < len(linesB) && lineA == lineB {
			if currentHunk != nil {
				lineNum := i + 1
				currentHunk.Lines = append(currentHunk.Lines, domain.DiffLine{
					Type:    domain.PromptDiffTypeUnchanged,
					Content: lineA,
					LineA:   &lineNum,
					LineB:   &lineNum,
				})
				// End hunk after context lines of unchanged content
				unchangedCount := 0
				for j := len(currentHunk.Lines) - 1; j >= 0; j-- {
					if currentHunk.Lines[j].Type == domain.PromptDiffTypeUnchanged {
						unchangedCount++
					} else {
						break
					}
				}
				if unchangedCount >= contextLines*2 {
					hunks = append(hunks, *currentHunk)
					currentHunk = nil
				}
			}
			continue
		}

		if currentHunk == nil {
			startA := i + 1
			startB := i + 1
			currentHunk = &domain.DiffHunk{
				StartLineA: startA,
				StartLineB: startB,
			}
			// Add context lines before the change
			for j := max(0, i-contextLines); j < i; j++ {
				if j < len(linesA) {
					lineNum := j + 1
					currentHunk.Lines = append(currentHunk.Lines, domain.DiffLine{
						Type:    domain.PromptDiffTypeUnchanged,
						Content: linesA[j],
						LineA:   &lineNum,
						LineB:   &lineNum,
					})
				}
			}
		}

		if i < len(linesA) && (i >= len(linesB) || lineA != lineB) {
			lineNum := i + 1
			currentHunk.Lines = append(currentHunk.Lines, domain.DiffLine{
				Type:    domain.PromptDiffTypeRemoved,
				Content: lineA,
				LineA:   &lineNum,
			})
		}

		if i < len(linesB) && (i >= len(linesA) || lineA != lineB) {
			lineNum := i + 1
			currentHunk.Lines = append(currentHunk.Lines, domain.DiffLine{
				Type:    domain.PromptDiffTypeAdded,
				Content: lineB,
				LineB:   &lineNum,
			})
		}
	}

	if currentHunk != nil {
		hunks = append(hunks, *currentHunk)
	}

	return hunks
}

func (s *PromptDiffService) computeVariableDiff(varsA, varsB []string) *domain.VariableDiff {
	setA := make(map[string]bool)
	setB := make(map[string]bool)
	for _, v := range varsA {
		setA[v] = true
	}
	for _, v := range varsB {
		setB[v] = true
	}

	diff := &domain.VariableDiff{}

	for v := range setA {
		if setB[v] {
			diff.Common = append(diff.Common, v)
		} else {
			diff.Removed = append(diff.Removed, v)
		}
	}

	for v := range setB {
		if !setA[v] {
			diff.Added = append(diff.Added, v)
		}
	}

	return diff
}

func (s *PromptDiffService) computeConfigDiff(configA, configB string) *domain.ConfigDiff {
	var mapA, mapB map[string]interface{}

	if configA != "" {
		json.Unmarshal([]byte(configA), &mapA)
	}
	if configB != "" {
		json.Unmarshal([]byte(configB), &mapB)
	}

	if mapA == nil {
		mapA = make(map[string]interface{})
	}
	if mapB == nil {
		mapB = make(map[string]interface{})
	}

	diff := &domain.ConfigDiff{
		Added:   make(map[string]interface{}),
		Removed: make(map[string]interface{}),
		Changed: make(map[string]domain.ConfigFieldChange),
	}

	for k, v := range mapA {
		if vB, ok := mapB[k]; ok {
			aJSON, _ := json.Marshal(v)
			bJSON, _ := json.Marshal(vB)
			if string(aJSON) != string(bJSON) {
				diff.Changed[k] = domain.ConfigFieldChange{OldValue: v, NewValue: vB}
			}
		} else {
			diff.Removed[k] = v
		}
	}

	for k, v := range mapB {
		if _, ok := mapA[k]; !ok {
			diff.Added[k] = v
		}
	}

	return diff
}

func (s *PromptDiffService) categorizeChanges(contentDiff *domain.SemanticDiff, varDiff *domain.VariableDiff, configDiff *domain.ConfigDiff) []domain.PromptChange {
	var changes []domain.PromptChange

	if len(varDiff.Added) > 0 {
		changes = append(changes, domain.PromptChange{
			Category:    domain.PromptChangeCategoryVariable,
			Description: fmt.Sprintf("Added variables: %s", strings.Join(varDiff.Added, ", ")),
			Severity:    "medium",
		})
	}

	if len(varDiff.Removed) > 0 {
		changes = append(changes, domain.PromptChange{
			Category:    domain.PromptChangeCategoryVariable,
			Description: fmt.Sprintf("Removed variables: %s", strings.Join(varDiff.Removed, ", ")),
			Severity:    "high",
		})
	}

	if contentDiff.AddedLines > 0 || contentDiff.RemovedLines > 0 {
		severity := "low"
		if contentDiff.AddedLines+contentDiff.RemovedLines > 10 {
			severity = "medium"
		}
		if contentDiff.AddedLines+contentDiff.RemovedLines > 30 {
			severity = "high"
		}
		changes = append(changes, domain.PromptChange{
			Category:    domain.PromptChangeCategoryContent,
			Description: fmt.Sprintf("Content changed: +%d/-%d lines", contentDiff.AddedLines, contentDiff.RemovedLines),
			Severity:    severity,
		})
	}

	if configDiff != nil {
		if len(configDiff.Changed) > 0 || len(configDiff.Added) > 0 || len(configDiff.Removed) > 0 {
			changes = append(changes, domain.PromptChange{
				Category:    domain.PromptChangeCategoryConfig,
				Description: "Configuration parameters changed",
				Severity:    "medium",
			})
		}
	}

	return changes
}

func (s *PromptDiffService) computeDiffSummary(changes []domain.PromptChange) *domain.DiffSummary {
	summary := &domain.DiffSummary{
		TotalChanges: len(changes),
		ByCategory:   make(map[domain.PromptChangeCategory]int),
		RiskLevel:    "low",
	}

	hasHigh := false
	hasMedium := false

	for _, change := range changes {
		summary.ByCategory[change.Category]++
		if change.Severity == "high" {
			hasHigh = true
		}
		if change.Severity == "medium" {
			hasMedium = true
		}
	}

	if hasHigh {
		summary.RiskLevel = "high"
		summary.RiskExplanation = "High-risk changes detected (variable removal or major content changes). Review carefully before deploying."
	} else if hasMedium {
		summary.RiskLevel = "medium"
		summary.RiskExplanation = "Moderate changes detected. Consider testing with a subset of traffic first."
	} else {
		summary.RiskLevel = "low"
		summary.RiskExplanation = "Minor changes. Safe to deploy."
	}

	return summary
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
