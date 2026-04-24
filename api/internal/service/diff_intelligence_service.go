package service

import (
	"context"
	"fmt"
	"math"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

// DiffAnalysisRepository interface for the service
type DiffAnalysisRepository interface {
	Save(ctx context.Context, analysis *domain.DiffAnalysis) error
	GetByID(ctx context.Context, projectID, id uuid.UUID) (*domain.DiffAnalysis, error)
	GetByTraceID(ctx context.Context, projectID, traceID uuid.UUID) ([]domain.DiffAnalysisSummary, error)
	List(ctx context.Context, filter *domain.DiffAnalysisFilter, limit, offset int) ([]domain.DiffAnalysisSummary, int64, error)
	GetQualityTrend(ctx context.Context, projectID uuid.UUID, since time.Time) (*domain.QualityTrend, error)
}

// DiffIntelligenceService provides code diff analysis capabilities
type DiffIntelligenceService struct {
	logger *zap.Logger
	repo   DiffAnalysisRepository
}

// NewDiffIntelligenceService creates a new diff intelligence service
func NewDiffIntelligenceService(logger *zap.Logger, repo DiffAnalysisRepository) *DiffIntelligenceService {
	return &DiffIntelligenceService{
		logger: logger,
		repo:   repo,
	}
}

// AnalyzeDiff creates and runs a diff analysis
func (s *DiffIntelligenceService) AnalyzeDiff(ctx context.Context, projectID uuid.UUID, input *domain.DiffAnalysisInput) (*domain.DiffAnalysis, error) {
	analysis := &domain.DiffAnalysis{
		ID:              uuid.New(),
		ProjectID:       projectID,
		TraceID:         input.TraceID,
		Status:          domain.DiffAnalysisRunning,
		DimensionScores: make(map[domain.QualityDimension]float64),
		GitCommitSha:    input.GitCommitSha,
		GitBranch:       input.GitBranch,
		CreatedAt:       time.Now(),
	}

	// Analyze each file change
	for _, fc := range input.FileChanges {
		fileAnalysis := s.analyzeFileChange(fc)
		analysis.FileAnalyses = append(analysis.FileAnalyses, fileAnalysis)
		analysis.Findings = append(analysis.Findings, fileAnalysis.Findings...)

		switch fc.Operation {
		case "add":
			analysis.FilesAdded++
		case "modify":
			analysis.FilesModified++
		case "delete":
			analysis.FilesDeleted++
		}
		analysis.LinesAdded += fileAnalysis.LinesAdded
		analysis.LinesRemoved += fileAnalysis.LinesRemoved
	}

	// Calculate aggregate scores
	s.calculateDimensionScores(analysis)
	s.calculateOverallScore(analysis)

	now := time.Now()
	analysis.CompletedAt = &now
	analysis.Status = domain.DiffAnalysisCompleted

	// Persist
	if err := s.repo.Save(ctx, analysis); err != nil {
		return nil, fmt.Errorf("failed to save diff analysis: %w", err)
	}

	s.logger.Info("diff analysis completed",
		zap.String("analysisId", analysis.ID.String()),
		zap.String("traceId", analysis.TraceID.String()),
		zap.Float64("overallScore", analysis.OverallScore),
		zap.Int("findingCount", len(analysis.Findings)),
	)

	return analysis, nil
}

// GetAnalysis retrieves a specific analysis
func (s *DiffIntelligenceService) GetAnalysis(ctx context.Context, projectID, id uuid.UUID) (*domain.DiffAnalysis, error) {
	return s.repo.GetByID(ctx, projectID, id)
}

// GetTraceAnalyses retrieves analyses for a trace
func (s *DiffIntelligenceService) GetTraceAnalyses(ctx context.Context, projectID, traceID uuid.UUID) ([]domain.DiffAnalysisSummary, error) {
	return s.repo.GetByTraceID(ctx, projectID, traceID)
}

// ListAnalyses retrieves analyses with filtering
func (s *DiffIntelligenceService) ListAnalyses(ctx context.Context, filter *domain.DiffAnalysisFilter, limit, offset int) ([]domain.DiffAnalysisSummary, int64, error) {
	return s.repo.List(ctx, filter, limit, offset)
}

// GetQualityTrend retrieves quality trend data
func (s *DiffIntelligenceService) GetQualityTrend(ctx context.Context, projectID uuid.UUID, days int) (*domain.QualityTrend, error) {
	since := time.Now().AddDate(0, 0, -days)
	return s.repo.GetQualityTrend(ctx, projectID, since)
}

func (s *DiffIntelligenceService) analyzeFileChange(fc domain.FileChangeInput) domain.FileAnalysis {
	fa := domain.FileAnalysis{
		FilePath: fc.FilePath,
		Language: s.detectLanguage(fc.FilePath, fc.Language),
		Diff:     fc.Diff,
	}

	// Count line changes
	if fc.Diff != "" {
		for _, line := range strings.Split(fc.Diff, "\n") {
			if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++") {
				fa.LinesAdded++
			} else if strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "---") {
				fa.LinesRemoved++
			}
		}
	}

	// Run static analysis checks
	fa.Findings = s.runStaticChecks(fc, fa.Language)

	// Calculate complexity delta
	fa.ComplexityDelta = s.estimateComplexityDelta(fc)

	// Calculate file quality score
	fa.QualityScore = s.calculateFileQuality(fa)

	return fa
}

func (s *DiffIntelligenceService) runStaticChecks(fc domain.FileChangeInput, language string) []domain.DiffFinding {
	var findings []domain.DiffFinding
	content := fc.ContentAfter
	if content == "" {
		content = fc.Diff
	}

	// Security checks
	securityPatterns := map[string]string{
		`(?i)password\s*=\s*["']`:          "Potential hardcoded password",
		`(?i)api[_-]?key\s*=\s*["']`:       "Potential hardcoded API key",
		`(?i)secret\s*=\s*["']`:            "Potential hardcoded secret",
		`(?i)token\s*=\s*["'][a-zA-Z0-9]+`: "Potential hardcoded token",
		`(?i)eval\s*\(`:                    "Use of eval() - potential code injection",
		`(?i)exec\s*\(`:                    "Use of exec() - potential command injection",
		`(?i)innerHTML\s*=`:                "Direct innerHTML assignment - XSS risk",
		`(?i)SELECT\s+.*\s+FROM.*\+`:       "Potential SQL injection via string concatenation",
	}

	for pattern, title := range securityPatterns {
		if strings.Contains(strings.ToLower(content), strings.ToLower(strings.Split(pattern, `\s`)[0])) {
			findings = append(findings, domain.DiffFinding{
				ID:          uuid.New().String(),
				Severity:    domain.FindingSeverityWarning,
				Category:    "security",
				Title:       title,
				Description: fmt.Sprintf("Detected pattern that may indicate a security issue in %s", fc.FilePath),
				FilePath:    fc.FilePath,
				Confidence:  0.6,
			})
		}
	}

	// Quality checks
	if fc.ContentAfter != "" {
		lines := strings.Split(fc.ContentAfter, "\n")

		// Long file check
		if len(lines) > 500 {
			findings = append(findings, domain.DiffFinding{
				ID:          uuid.New().String(),
				Severity:    domain.FindingSeverityInfo,
				Category:    "quality",
				Title:       "Large file",
				Description: fmt.Sprintf("File has %d lines, consider splitting into smaller modules", len(lines)),
				FilePath:    fc.FilePath,
				Confidence:  0.8,
			})
		}

		// Long line check
		for i, line := range lines {
			if len(line) > 200 {
				findings = append(findings, domain.DiffFinding{
					ID:         uuid.New().String(),
					Severity:   domain.FindingSeverityInfo,
					Category:   "quality",
					Title:      "Line exceeds 200 characters",
					FilePath:   fc.FilePath,
					StartLine:  i + 1,
					EndLine:    i + 1,
					Confidence: 0.9,
				})
				break // Only report first occurrence
			}
		}

		// TODO/FIXME check
		for i, line := range lines {
			lower := strings.ToLower(line)
			if strings.Contains(lower, "todo") || strings.Contains(lower, "fixme") || strings.Contains(lower, "hack") {
				findings = append(findings, domain.DiffFinding{
					ID:          uuid.New().String(),
					Severity:    domain.FindingSeverityInfo,
					Category:    "quality",
					Title:       "TODO/FIXME comment found",
					Description: strings.TrimSpace(line),
					FilePath:    fc.FilePath,
					StartLine:   i + 1,
					Confidence:  0.95,
				})
			}
		}
	}

	// Missing error handling (language-specific)
	if language == "go" && fc.ContentAfter != "" {
		if strings.Contains(fc.ContentAfter, "_ = ") || strings.Contains(fc.ContentAfter, ", _ =") {
			findings = append(findings, domain.DiffFinding{
				ID:          uuid.New().String(),
				Severity:    domain.FindingSeverityWarning,
				Category:    "quality",
				Title:       "Ignored error value",
				Description: "Error return value is being ignored. Consider handling the error.",
				FilePath:    fc.FilePath,
				Confidence:  0.7,
			})
		}
	}

	return findings
}

func (s *DiffIntelligenceService) detectLanguage(filePath, hint string) string {
	if hint != "" {
		return hint
	}
	ext := strings.ToLower(filepath.Ext(filePath))
	langMap := map[string]string{
		".go":    "go",
		".py":    "python",
		".ts":    "typescript",
		".tsx":   "typescript",
		".js":    "javascript",
		".jsx":   "javascript",
		".rs":    "rust",
		".java":  "java",
		".rb":    "ruby",
		".cpp":   "cpp",
		".c":     "c",
		".cs":    "csharp",
		".swift": "swift",
		".kt":    "kotlin",
		".sh":    "shell",
		".yml":   "yaml",
		".yaml":  "yaml",
		".json":  "json",
		".md":    "markdown",
	}
	if lang, ok := langMap[ext]; ok {
		return lang
	}
	return "unknown"
}

func (s *DiffIntelligenceService) estimateComplexityDelta(fc domain.FileChangeInput) int {
	content := fc.ContentAfter
	if content == "" {
		return 0
	}

	// Simple cyclomatic complexity estimation based on control flow keywords
	complexityKeywords := []string{"if ", "else ", "for ", "while ", "switch ", "case ", "catch ", "&&", "||", "?"}
	delta := 0
	for _, kw := range complexityKeywords {
		afterCount := strings.Count(content, kw)
		beforeCount := strings.Count(fc.ContentBefore, kw)
		delta += afterCount - beforeCount
	}
	return delta
}

func (s *DiffIntelligenceService) calculateFileQuality(fa domain.FileAnalysis) float64 {
	score := 100.0

	// Deduct for findings based on severity
	for _, f := range fa.Findings {
		switch f.Severity {
		case domain.FindingSeverityCritical:
			score -= 20
		case domain.FindingSeverityError:
			score -= 10
		case domain.FindingSeverityWarning:
			score -= 5
		case domain.FindingSeverityInfo:
			score -= 1
		}
	}

	// Deduct for high complexity
	if fa.ComplexityDelta > 10 {
		score -= float64(fa.ComplexityDelta - 10)
	}

	return math.Max(0, math.Min(100, score))
}

func (s *DiffIntelligenceService) calculateDimensionScores(analysis *domain.DiffAnalysis) {
	// Security score
	securityDeductions := 0.0
	for _, f := range analysis.Findings {
		if f.Category == "security" {
			switch f.Severity {
			case domain.FindingSeverityCritical:
				securityDeductions += 25
			case domain.FindingSeverityError:
				securityDeductions += 15
			case domain.FindingSeverityWarning:
				securityDeductions += 8
			}
		}
	}
	analysis.DimensionScores[domain.QualitySecurity] = math.Max(0, 100-securityDeductions)

	// Complexity score
	totalComplexityDelta := 0
	for _, fa := range analysis.FileAnalyses {
		if fa.ComplexityDelta > 0 {
			totalComplexityDelta += fa.ComplexityDelta
		}
	}
	complexityScore := 100.0 - math.Min(50, float64(totalComplexityDelta)*2)
	analysis.DimensionScores[domain.QualityComplexity] = math.Max(0, complexityScore)

	// Readability score (based on file sizes and line lengths)
	readabilityDeductions := 0.0
	for _, f := range analysis.Findings {
		if f.Category == "quality" {
			readabilityDeductions += 3
		}
	}
	analysis.DimensionScores[domain.QualityReadability] = math.Max(0, 100-readabilityDeductions)

	// Maintainability (composite)
	analysis.DimensionScores[domain.QualityMaintainability] = (analysis.DimensionScores[domain.QualityComplexity] +
		analysis.DimensionScores[domain.QualityReadability]) / 2
}

func (s *DiffIntelligenceService) calculateOverallScore(analysis *domain.DiffAnalysis) {
	weights := map[domain.QualityDimension]float64{
		domain.QualitySecurity:        0.30,
		domain.QualityComplexity:      0.25,
		domain.QualityReadability:     0.20,
		domain.QualityMaintainability: 0.25,
	}

	total := 0.0
	weightSum := 0.0
	for dim, weight := range weights {
		if score, ok := analysis.DimensionScores[dim]; ok {
			total += score * weight
			weightSum += weight
		}
	}

	if weightSum > 0 {
		analysis.OverallScore = total / weightSum
	}
}
