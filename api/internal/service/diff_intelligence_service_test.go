package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

// MockDiffAnalysisRepository mocks the DiffAnalysisRepository interface
type MockDiffAnalysisRepository struct {
	mock.Mock
}

func (m *MockDiffAnalysisRepository) Save(ctx context.Context, analysis *domain.DiffAnalysis) error {
	args := m.Called(ctx, analysis)
	return args.Error(0)
}

func (m *MockDiffAnalysisRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.DiffAnalysis, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.DiffAnalysis), args.Error(1)
}

func (m *MockDiffAnalysisRepository) GetByTraceID(ctx context.Context, traceID uuid.UUID) ([]domain.DiffAnalysisSummary, error) {
	args := m.Called(ctx, traceID)
	return args.Get(0).([]domain.DiffAnalysisSummary), args.Error(1)
}

func (m *MockDiffAnalysisRepository) List(ctx context.Context, filter *domain.DiffAnalysisFilter, limit, offset int) ([]domain.DiffAnalysisSummary, int64, error) {
	args := m.Called(ctx, filter, limit, offset)
	return args.Get(0).([]domain.DiffAnalysisSummary), args.Get(1).(int64), args.Error(2)
}

func (m *MockDiffAnalysisRepository) GetQualityTrend(ctx context.Context, projectID uuid.UUID, since time.Time) (*domain.QualityTrend, error) {
	args := m.Called(ctx, projectID, since)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.QualityTrend), args.Error(1)
}

func TestNewDiffIntelligenceService(t *testing.T) {
	repo := new(MockDiffAnalysisRepository)
	svc := NewDiffIntelligenceService(zap.NewNop(), repo)
	assert.NotNil(t, svc)
}

func TestDiffIntelligenceService_AnalyzeDiff(t *testing.T) {
	repo := new(MockDiffAnalysisRepository)
	svc := NewDiffIntelligenceService(zap.NewNop(), repo)
	ctx := context.Background()
	projectID := uuid.New()

	t.Run("analyzes simple file change", func(t *testing.T) {
		repo.On("Save", mock.Anything, mock.AnythingOfType("*domain.DiffAnalysis")).Return(nil).Once()

		input := &domain.DiffAnalysisInput{
			TraceID: uuid.New(),
			FileChanges: []domain.FileChangeInput{
				{
					FilePath:      "main.go",
					Operation:     "modify",
					ContentBefore: "package main\nfunc main() {}\n",
					ContentAfter:  "package main\nfunc main() {\n\tfmt.Println(\"hello\")\n}\n",
					Diff:          "+\tfmt.Println(\"hello\")",
					Language:      "go",
				},
			},
		}

		analysis, err := svc.AnalyzeDiff(ctx, projectID, input)
		require.NoError(t, err)
		require.NotNil(t, analysis)
		assert.Equal(t, domain.DiffAnalysisCompleted, analysis.Status)
		assert.Equal(t, 0, analysis.FilesAdded)
		assert.Equal(t, 1, analysis.FilesModified)
		assert.Greater(t, analysis.OverallScore, 0.0)
		assert.LessOrEqual(t, analysis.OverallScore, 100.0)
		repo.AssertExpectations(t)
	})

	t.Run("analyzes file with potential security content", func(t *testing.T) {
		repo.On("Save", mock.Anything, mock.AnythingOfType("*domain.DiffAnalysis")).Return(nil).Once()

		input := &domain.DiffAnalysisInput{
			TraceID: uuid.New(),
			FileChanges: []domain.FileChangeInput{
				{
					FilePath:     "config.py",
					Operation:    "add",
					ContentAfter: "password = \"secret123\"\napi_key = \"sk-abc123\"\n",
					Language:     "python",
				},
			},
		}

		analysis, err := svc.AnalyzeDiff(ctx, projectID, input)
		require.NoError(t, err)
		assert.Equal(t, domain.DiffAnalysisCompleted, analysis.Status)
		assert.Equal(t, 1, analysis.FilesAdded)
		assert.Greater(t, analysis.OverallScore, 0.0)
		assert.LessOrEqual(t, analysis.OverallScore, 100.0)
		repo.AssertExpectations(t)
	})

	t.Run("detects quality findings", func(t *testing.T) {
		repo.On("Save", mock.Anything, mock.AnythingOfType("*domain.DiffAnalysis")).Return(nil).Once()

		// Create a large file with TODO
		longContent := "package main\n// TODO: fix this later\n"
		for i := 0; i < 600; i++ {
			longContent += "func placeholder() {}\n"
		}

		input := &domain.DiffAnalysisInput{
			TraceID: uuid.New(),
			FileChanges: []domain.FileChangeInput{
				{
					FilePath:     "big.go",
					Operation:    "add",
					ContentAfter: longContent,
					Language:     "go",
				},
			},
		}

		analysis, err := svc.AnalyzeDiff(ctx, projectID, input)
		require.NoError(t, err)

		qualityFindings := 0
		for _, f := range analysis.Findings {
			if f.Category == "quality" {
				qualityFindings++
			}
		}
		assert.Greater(t, qualityFindings, 0, "should detect quality issues (large file, TODO)")
		repo.AssertExpectations(t)
	})

	t.Run("handles multiple file changes", func(t *testing.T) {
		repo.On("Save", mock.Anything, mock.AnythingOfType("*domain.DiffAnalysis")).Return(nil).Once()

		input := &domain.DiffAnalysisInput{
			TraceID: uuid.New(),
			FileChanges: []domain.FileChangeInput{
				{FilePath: "a.ts", Operation: "add", ContentAfter: "const a = 1;\n"},
				{FilePath: "b.py", Operation: "modify", Diff: "+line1\n-line2\n"},
				{FilePath: "c.go", Operation: "delete"},
			},
		}

		analysis, err := svc.AnalyzeDiff(ctx, projectID, input)
		require.NoError(t, err)
		assert.Equal(t, 1, analysis.FilesAdded)
		assert.Equal(t, 1, analysis.FilesModified)
		assert.Equal(t, 1, analysis.FilesDeleted)
		assert.Len(t, analysis.FileAnalyses, 3)
		repo.AssertExpectations(t)
	})
}

func TestDiffIntelligenceService_LanguageDetection(t *testing.T) {
	svc := NewDiffIntelligenceService(zap.NewNop(), nil)

	tests := []struct {
		filePath string
		hint     string
		expected string
	}{
		{"main.go", "", "go"},
		{"app.tsx", "", "typescript"},
		{"script.py", "", "python"},
		{"lib.rs", "", "rust"},
		{"unknown.xyz", "", "unknown"},
		{"file.txt", "python", "python"},
	}

	for _, tt := range tests {
		t.Run(tt.filePath, func(t *testing.T) {
			result := svc.detectLanguage(tt.filePath, tt.hint)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDiffIntelligenceService_ScoreCalculation(t *testing.T) {
	svc := NewDiffIntelligenceService(zap.NewNop(), nil)

	t.Run("perfect score with no findings", func(t *testing.T) {
		analysis := &domain.DiffAnalysis{
			DimensionScores: make(map[domain.QualityDimension]float64),
			Findings:        []domain.DiffFinding{},
			FileAnalyses:    []domain.FileAnalysis{},
		}
		svc.calculateDimensionScores(analysis)
		svc.calculateOverallScore(analysis)
		assert.Equal(t, 100.0, analysis.DimensionScores[domain.QualitySecurity])
		assert.Greater(t, analysis.OverallScore, 90.0)
	})

	t.Run("lower score with critical findings", func(t *testing.T) {
		analysis := &domain.DiffAnalysis{
			DimensionScores: make(map[domain.QualityDimension]float64),
			Findings: []domain.DiffFinding{
				{Severity: domain.FindingSeverityCritical, Category: "security"},
				{Severity: domain.FindingSeverityCritical, Category: "security"},
			},
			FileAnalyses: []domain.FileAnalysis{},
		}
		svc.calculateDimensionScores(analysis)
		svc.calculateOverallScore(analysis)
		assert.Less(t, analysis.DimensionScores[domain.QualitySecurity], 60.0)
	})
}
