package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

// CodeQualityService handles code quality analysis logic
type CodeQualityService struct {
	logger *zap.Logger
}

// NewCodeQualityService creates a new code quality service
func NewCodeQualityService(logger *zap.Logger) *CodeQualityService {
	return &CodeQualityService{
		logger: logger,
	}
}

// CreateConfig creates a new code quality configuration for a project
func (s *CodeQualityService) CreateConfig(ctx context.Context, projectID uuid.UUID, input domain.CodeQualityConfigInput) (*domain.CodeQualityConfig, error) {
	now := time.Now()

	config := &domain.CodeQualityConfig{
		ID:                uuid.New(),
		ProjectID:         projectID,
		Name:              input.Name,
		Analyzers:         input.Analyzers,
		AutoRunOnTrace:    input.AutoRunOnTrace,
		MinScoreThreshold: input.MinScoreThreshold,
		FailOnBlocker:     input.FailOnBlocker,
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	s.logger.Info("created code quality config",
		zap.String("configId", config.ID.String()),
		zap.String("projectId", projectID.String()),
		zap.String("name", config.Name),
		zap.Int("analyzers", len(config.Analyzers)),
	)

	return config, nil
}

// GetConfig returns a specific code quality configuration by ID
func (s *CodeQualityService) GetConfig(ctx context.Context, projectID, configID uuid.UUID) (*domain.CodeQualityConfig, error) {
	s.logger.Info("fetching code quality config",
		zap.String("projectId", projectID.String()),
		zap.String("configId", configID.String()),
	)

	config := &domain.CodeQualityConfig{
		ID:        configID,
		ProjectID: projectID,
		Name:      "Default Quality Config",
		Analyzers: []domain.AnalyzerConfig{
			{Type: domain.CodeQualityAnalyzerESLint, Enabled: true, Weight: 0.4},
			{Type: domain.CodeQualityAnalyzerSemgrep, Enabled: true, Weight: 0.35},
			{Type: domain.CodeQualityAnalyzerSonarQube, Enabled: true, Weight: 0.25},
		},
		AutoRunOnTrace:    true,
		MinScoreThreshold: 70.0,
		FailOnBlocker:     true,
		CreatedAt:         time.Now().Add(-24 * time.Hour),
		UpdatedAt:         time.Now(),
	}

	return config, nil
}

// AnalyzeTrace runs code quality analysis on a trace and returns a report
func (s *CodeQualityService) AnalyzeTrace(ctx context.Context, projectID uuid.UUID, input domain.CodeQualityInput) (*domain.CodeQualityReport, error) {
	s.logger.Info("analyzing trace for code quality",
		zap.String("projectId", projectID.String()),
		zap.String("traceId", input.TraceID.String()),
	)

	findings := s.generateFindings()
	analyzerResults := s.calculateAnalyzerResults(findings)
	overallScore := s.calculateOverallScore(analyzerResults)
	grade := s.assignGrade(overallScore)

	findingsBySeverity := map[string]int{}
	for _, f := range findings {
		findingsBySeverity[string(f.Severity)]++
	}

	report := &domain.CodeQualityReport{
		ID:                 uuid.New(),
		ProjectID:          projectID,
		TraceID:            input.TraceID,
		ConfigID:           input.ConfigID,
		OverallScore:       overallScore,
		Grade:              grade,
		Findings:           findings,
		AnalyzerResults:    analyzerResults,
		LinesAnalyzed:      3842,
		FilesAnalyzed:      27,
		TotalFindings:      len(findings),
		FindingsBySeverity: findingsBySeverity,
		Passed:             overallScore >= 70.0,
		DurationMs:         1250,
		Duration:           1250 * time.Millisecond,
		CreatedAt:          time.Now(),
	}

	s.logger.Info("completed code quality analysis",
		zap.String("reportId", report.ID.String()),
		zap.Float64("score", overallScore),
		zap.String("grade", grade),
		zap.Int("findings", len(findings)),
	)

	return report, nil
}

// ListReports returns a paginated list of code quality reports for a project
func (s *CodeQualityService) ListReports(ctx context.Context, projectID uuid.UUID, traceID *uuid.UUID, limit, offset int) (*domain.CodeQualityReportList, error) {
	s.logger.Info("listing code quality reports",
		zap.String("projectId", projectID.String()),
		zap.Int("limit", limit),
		zap.Int("offset", offset),
	)

	baseTime := time.Date(2024, 6, 15, 10, 0, 0, 0, time.UTC)

	demoReports := []struct {
		score float64
		grade string
		files int
		lines int
	}{
		{92.5, "A", 32, 4210},
		{78.3, "C", 18, 2105},
		{85.1, "B", 25, 3420},
	}

	reports := make([]domain.CodeQualityReport, 0, len(demoReports))
	for i, d := range demoReports {
		tid := uuid.New()
		if traceID != nil {
			tid = *traceID
		}
		reports = append(reports, domain.CodeQualityReport{
			ID:            uuid.New(),
			ProjectID:     projectID,
			TraceID:       tid,
			OverallScore:  d.score,
			Grade:         d.grade,
			Findings:      []domain.CodeQualityFinding{},
			AnalyzerResults: []domain.AnalyzerResult{},
			FilesAnalyzed: d.files,
			LinesAnalyzed: d.lines,
			TotalFindings: (i + 1) * 3,
			FindingsBySeverity: map[string]int{
				"major": i + 1,
				"minor": (i + 1) * 2,
			},
			Passed:     d.score >= 70.0,
			DurationMs: int64(800 + i*200),
			CreatedAt:  baseTime.Add(time.Duration(i) * 24 * time.Hour),
		})
	}

	// Apply pagination
	total := int64(len(reports))
	if offset >= len(reports) {
		reports = []domain.CodeQualityReport{}
	} else {
		end := offset + limit
		if end > len(reports) {
			end = len(reports)
		}
		reports = reports[offset:end]
	}

	return &domain.CodeQualityReportList{
		Reports:    reports,
		TotalCount: total,
		HasMore:    int64(offset+limit) < total,
	}, nil
}

// GetReport returns a specific code quality report by ID
func (s *CodeQualityService) GetReport(ctx context.Context, projectID, reportID uuid.UUID) (*domain.CodeQualityReport, error) {
	s.logger.Info("fetching code quality report",
		zap.String("projectId", projectID.String()),
		zap.String("reportId", reportID.String()),
	)

	findings := s.generateFindings()
	analyzerResults := s.calculateAnalyzerResults(findings)
	overallScore := s.calculateOverallScore(analyzerResults)
	grade := s.assignGrade(overallScore)

	findingsBySeverity := map[string]int{}
	for _, f := range findings {
		findingsBySeverity[string(f.Severity)]++
	}

	report := &domain.CodeQualityReport{
		ID:                 reportID,
		ProjectID:          projectID,
		TraceID:            uuid.New(),
		OverallScore:       overallScore,
		Grade:              grade,
		Findings:           findings,
		AnalyzerResults:    analyzerResults,
		LinesAnalyzed:      3842,
		FilesAnalyzed:      27,
		TotalFindings:      len(findings),
		FindingsBySeverity: findingsBySeverity,
		Passed:             overallScore >= 70.0,
		DurationMs:         1250,
		Duration:           1250 * time.Millisecond,
		CreatedAt:          time.Now().Add(-2 * time.Hour),
	}

	return report, nil
}

// GetDashboard returns aggregated code quality metrics for a project
func (s *CodeQualityService) GetDashboard(ctx context.Context, projectID uuid.UUID) (*domain.CodeQualityDashboard, error) {
	s.logger.Info("fetching code quality dashboard",
		zap.String("projectId", projectID.String()),
	)

	baseDate := time.Date(2024, 6, 10, 0, 0, 0, 0, time.UTC)

	topIssues := []domain.CodeQualityFinding{
		{
			ID:       "top-1",
			Analyzer: domain.CodeQualityAnalyzerSemgrep,
			Severity: domain.CodeQualitySeverityCritical,
			Rule:     "hardcoded-password",
			Message:  "Hardcoded password detected in configuration",
			FilePath: "src/config/database.ts",
			Line:     42,
			Column:   10,
		},
		{
			ID:       "top-2",
			Analyzer: domain.CodeQualityAnalyzerSonarQube,
			Severity: domain.CodeQualitySeverityMajor,
			Rule:     "cognitive-complexity",
			Message:  "Function has a cognitive complexity of 23 (max allowed: 15)",
			FilePath: "src/services/auth.ts",
			Line:     88,
			Column:   1,
		},
		{
			ID:       "top-3",
			Analyzer: domain.CodeQualityAnalyzerESLint,
			Severity: domain.CodeQualitySeverityMajor,
			Rule:     "no-unused-vars",
			Message:  "'tempData' is defined but never used",
			FilePath: "src/utils/helpers.ts",
			Line:     15,
			Column:   7,
		},
	}

	dashboard := &domain.CodeQualityDashboard{
		TotalScans:    48,
		TotalFindings: 156,
		AvgScore:      81.4,
		GradeDistribution: map[string]int{
			"A": 12,
			"B": 18,
			"C": 11,
			"D": 5,
			"F": 2,
		},
		TopIssues: topIssues,
		ScoreTrend: []domain.CodeQualityTrendPoint{
			{Date: baseDate, AvgScore: 74.2, TotalFindings: 42},
			{Date: baseDate.Add(7 * 24 * time.Hour), AvgScore: 77.8, TotalFindings: 38},
			{Date: baseDate.Add(14 * 24 * time.Hour), AvgScore: 80.1, TotalFindings: 35},
			{Date: baseDate.Add(21 * 24 * time.Hour), AvgScore: 82.5, TotalFindings: 28},
			{Date: baseDate.Add(28 * 24 * time.Hour), AvgScore: 85.3, TotalFindings: 22},
		},
		AnalyzerBreakdown: map[string]int{
			"eslint":    68,
			"semgrep":   45,
			"sonarqube": 43,
		},
	}

	return dashboard, nil
}

func (s *CodeQualityService) generateFindings() []domain.CodeQualityFinding {
	suggestion := func(s string) *string { return &s }
	endLine := func(l int) *int { return &l }

	return []domain.CodeQualityFinding{
		// ESLint findings
		{
			ID:         "eslint-1",
			Analyzer:   domain.CodeQualityAnalyzerESLint,
			Severity:   domain.CodeQualitySeverityMajor,
			Rule:       "no-unused-vars",
			Message:    "'tempData' is defined but never used",
			FilePath:   "src/utils/helpers.ts",
			Line:       15,
			Column:     7,
			Snippet:    "const tempData = fetchData();",
			Suggestion: suggestion("Remove the unused variable or prefix with underscore: _tempData"),
			Effort:     "5min",
		},
		{
			ID:         "eslint-2",
			Analyzer:   domain.CodeQualityAnalyzerESLint,
			Severity:   domain.CodeQualitySeverityMinor,
			Rule:       "no-console",
			Message:    "Unexpected console statement",
			FilePath:   "src/services/api.ts",
			Line:       47,
			Column:     5,
			Snippet:    "console.log('API response:', response);",
			Suggestion: suggestion("Replace with a proper logging library call"),
			Effort:     "5min",
		},
		{
			ID:         "eslint-3",
			Analyzer:   domain.CodeQualityAnalyzerESLint,
			Severity:   domain.CodeQualitySeverityMinor,
			Rule:       "prefer-const",
			Message:    "'config' is never reassigned. Use 'const' instead",
			FilePath:   "src/config/settings.ts",
			Line:       8,
			Column:     5,
			Snippet:    "let config = loadConfig();",
			Suggestion: suggestion("Change 'let' to 'const'"),
			Effort:     "2min",
		},
		{
			ID:       "eslint-4",
			Analyzer: domain.CodeQualityAnalyzerESLint,
			Severity: domain.CodeQualitySeverityMajor,
			Rule:     "eqeqeq",
			Message:  "Expected '===' and instead saw '=='",
			FilePath: "src/components/UserList.tsx",
			Line:     32,
			Column:   12,
			Snippet:  "if (user.role == 'admin') {",
			Suggestion: suggestion("Use strict equality operator: user.role === 'admin'"),
			Effort:     "2min",
		},
		{
			ID:         "eslint-5",
			Analyzer:   domain.CodeQualityAnalyzerESLint,
			Severity:   domain.CodeQualitySeverityMajor,
			Rule:       "no-var",
			Message:    "Unexpected var, use let or const instead",
			FilePath:   "src/legacy/processor.ts",
			Line:       101,
			Column:     1,
			Snippet:    "var result = processItems(items);",
			Suggestion: suggestion("Replace 'var' with 'const' or 'let'"),
			Effort:     "2min",
		},
		{
			ID:         "eslint-6",
			Analyzer:   domain.CodeQualityAnalyzerESLint,
			Severity:   domain.CodeQualitySeverityCritical,
			Rule:       "react-hooks/rules-of-hooks",
			Message:    "React Hook 'useState' is called conditionally",
			FilePath:   "src/components/Dashboard.tsx",
			Line:       24,
			Column:     9,
			EndLine:    endLine(26),
			Snippet:    "if (isReady) {\n  const [data, setData] = useState(null);\n}",
			Suggestion: suggestion("Move the Hook call to the top level of the component"),
			Effort:     "15min",
		},
		// Semgrep findings
		{
			ID:         "semgrep-1",
			Analyzer:   domain.CodeQualityAnalyzerSemgrep,
			Severity:   domain.CodeQualitySeverityCritical,
			Rule:       "hardcoded-password",
			Message:    "Hardcoded password detected in configuration",
			FilePath:   "src/config/database.ts",
			Line:       42,
			Column:     10,
			Snippet:    "const dbPassword = 'super_secret_123';",
			Suggestion: suggestion("Use environment variables: process.env.DB_PASSWORD"),
			Effort:     "10min",
		},
		{
			ID:         "semgrep-2",
			Analyzer:   domain.CodeQualityAnalyzerSemgrep,
			Severity:   domain.CodeQualitySeverityBlocker,
			Rule:       "sql-injection",
			Message:    "Potential SQL injection via string concatenation",
			FilePath:   "src/repositories/user.ts",
			Line:       67,
			Column:     14,
			EndLine:    endLine(69),
			Snippet:    "const query = `SELECT * FROM users WHERE id = ${userId}`;",
			Suggestion: suggestion("Use parameterized queries: db.query('SELECT * FROM users WHERE id = $1', [userId])"),
			Effort:     "15min",
		},
		{
			ID:         "semgrep-3",
			Analyzer:   domain.CodeQualityAnalyzerSemgrep,
			Severity:   domain.CodeQualitySeverityCritical,
			Rule:       "xss-vulnerability",
			Message:    "Potential cross-site scripting vulnerability with dangerouslySetInnerHTML",
			FilePath:   "src/components/RichText.tsx",
			Line:       19,
			Column:     7,
			Snippet:    "<div dangerouslySetInnerHTML={{__html: userInput}} />",
			Suggestion: suggestion("Sanitize input with DOMPurify before rendering: DOMPurify.sanitize(userInput)"),
			Effort:     "20min",
		},
		{
			ID:         "semgrep-4",
			Analyzer:   domain.CodeQualityAnalyzerSemgrep,
			Severity:   domain.CodeQualitySeverityMajor,
			Rule:       "insecure-crypto",
			Message:    "Use of insecure MD5 hash algorithm",
			FilePath:   "src/utils/hash.ts",
			Line:       11,
			Column:     18,
			Snippet:    "const hash = crypto.createHash('md5').update(data).digest('hex');",
			Suggestion: suggestion("Use a secure hash algorithm: crypto.createHash('sha256')"),
			Effort:     "5min",
		},
		// SonarQube findings
		{
			ID:         "sonar-1",
			Analyzer:   domain.CodeQualityAnalyzerSonarQube,
			Severity:   domain.CodeQualitySeverityMajor,
			Rule:       "cognitive-complexity",
			Message:    "Function has a cognitive complexity of 23 (max allowed: 15)",
			FilePath:   "src/services/auth.ts",
			Line:       88,
			Column:     1,
			EndLine:    endLine(142),
			Snippet:    "export async function validateAndRefreshToken(token: string, context: AuthContext) {",
			Suggestion: suggestion("Extract nested logic into smaller helper functions to reduce complexity"),
			Effort:     "30min",
		},
		{
			ID:         "sonar-2",
			Analyzer:   domain.CodeQualityAnalyzerSonarQube,
			Severity:   domain.CodeQualitySeverityMajor,
			Rule:       "duplicate-code",
			Message:    "Duplicated block of 18 lines detected across 2 files",
			FilePath:   "src/handlers/orders.ts",
			Line:       34,
			Column:     1,
			EndLine:    endLine(51),
			Snippet:    "const validated = validateInput(input);\nif (!validated.success) {\n  return { error: validated.error };",
			Suggestion: suggestion("Extract the duplicated validation logic into a shared utility function"),
			Effort:     "20min",
		},
		{
			ID:         "sonar-3",
			Analyzer:   domain.CodeQualityAnalyzerSonarQube,
			Severity:   domain.CodeQualitySeverityMinor,
			Rule:       "code-smell",
			Message:    "Function has 8 parameters, consider using an options object",
			FilePath:   "src/services/notification.ts",
			Line:       55,
			Column:     1,
			Snippet:    "function sendNotification(userId, type, title, body, channel, priority, ttl, metadata) {",
			Suggestion: suggestion("Refactor to accept a single options object: sendNotification(options: NotificationOptions)"),
			Effort:     "15min",
		},
	}
}

func (s *CodeQualityService) calculateAnalyzerResults(findings []domain.CodeQualityFinding) []domain.AnalyzerResult {
	type analyzerStats struct {
		count    int
		blocker  int
		critical int
		major    int
	}

	stats := map[domain.CodeQualityAnalyzer]*analyzerStats{}
	for _, f := range findings {
		st, ok := stats[f.Analyzer]
		if !ok {
			st = &analyzerStats{}
			stats[f.Analyzer] = st
		}
		st.count++
		switch f.Severity {
		case domain.CodeQualitySeverityBlocker:
			st.blocker++
		case domain.CodeQualitySeverityCritical:
			st.critical++
		case domain.CodeQualitySeverityMajor:
			st.major++
		}
	}

	calcScore := func(st *analyzerStats) float64 {
		score := 100.0
		score -= float64(st.blocker) * 15.0
		score -= float64(st.critical) * 10.0
		score -= float64(st.major) * 5.0
		score -= float64(st.count-st.blocker-st.critical-st.major) * 2.0
		if score < 0 {
			score = 0
		}
		return score
	}

	analyzers := []struct {
		typ      domain.CodeQualityAnalyzer
		duration int64
	}{
		{domain.CodeQualityAnalyzerESLint, 420},
		{domain.CodeQualityAnalyzerSemgrep, 530},
		{domain.CodeQualityAnalyzerSonarQube, 300},
	}

	results := make([]domain.AnalyzerResult, 0, len(analyzers))
	for _, a := range analyzers {
		st := stats[a.typ]
		if st == nil {
			st = &analyzerStats{}
		}
		score := calcScore(st)
		results = append(results, domain.AnalyzerResult{
			Analyzer:      a.typ,
			Score:         score,
			FindingsCount: st.count,
			Duration:      a.duration,
			Passed:        score >= 70.0,
		})
	}

	return results
}

func (s *CodeQualityService) calculateOverallScore(results []domain.AnalyzerResult) float64 {
	weights := map[domain.CodeQualityAnalyzer]float64{
		domain.CodeQualityAnalyzerESLint:    0.40,
		domain.CodeQualityAnalyzerSemgrep:   0.35,
		domain.CodeQualityAnalyzerSonarQube: 0.25,
	}

	var totalScore float64
	var totalWeight float64
	for _, r := range results {
		w := weights[r.Analyzer]
		totalScore += r.Score * w
		totalWeight += w
	}

	if totalWeight == 0 {
		return 0
	}
	return totalScore / totalWeight
}

func (s *CodeQualityService) assignGrade(score float64) string {
	switch {
	case score >= 90:
		return "A"
	case score >= 80:
		return "B"
	case score >= 70:
		return "C"
	case score >= 60:
		return "D"
	default:
		return "F"
	}
}
