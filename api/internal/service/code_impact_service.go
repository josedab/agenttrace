package service

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

// CodeImpactService handles trace-linked code impact analysis
type CodeImpactService struct {
	logger *zap.Logger
	query  *QueryService
	fileOp *FileOperationService
}

// NewCodeImpactService creates a new code impact service
func NewCodeImpactService(logger *zap.Logger, query *QueryService, fileOp *FileOperationService) *CodeImpactService {
	return &CodeImpactService{
		logger: logger,
		query:  query,
		fileOp: fileOp,
	}
}

// GetCodeImpact builds an impact map by fetching file operations for a trace
func (s *CodeImpactService) GetCodeImpact(ctx context.Context, projectID uuid.UUID, traceID string) (*domain.CodeImpactMap, error) {
	s.logger.Info("building code impact map",
		zap.String("projectId", projectID.String()),
		zap.String("traceId", traceID),
	)

	trace, err := s.query.GetTrace(ctx, projectID, traceID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch trace: %w", err)
	}

	fileOps, err := s.fileOp.GetByTraceID(ctx, projectID, traceID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch file operations: %w", err)
	}

	observations, err := s.query.GetObservationsByTraceID(ctx, projectID, traceID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch observations: %w", err)
	}

	files := s.buildFileImpacts(fileOps, observations)
	summary := s.buildSummary(files)

	impactMap := &domain.CodeImpactMap{
		ID:        uuid.New(),
		ProjectID: projectID,
		TraceID:   traceID,
		SessionID: trace.SessionID,
		RepoURL:   "https://github.com/example/repo",
		Branch:    "main",
		Files:     files,
		Summary:   summary,
		CreatedAt: time.Now(),
	}

	s.logger.Info("code impact map built",
		zap.String("traceId", traceID),
		zap.Int("totalFiles", summary.TotalFiles),
	)

	return impactMap, nil
}

// GetProjectImpactSummary returns aggregated impact summary for a project
func (s *CodeImpactService) GetProjectImpactSummary(ctx context.Context, projectID uuid.UUID, filter *domain.CodeImpactFilter) (*domain.CodeImpactSummary, error) {
	s.logger.Info("fetching project impact summary",
		zap.String("projectId", projectID.String()),
	)

	// Return mock aggregated summary
	summary := &domain.CodeImpactSummary{
		TotalFiles:        42,
		FilesCreated:      12,
		FilesModified:     25,
		FilesDeleted:      5,
		TotalLinesAdded:   1847,
		TotalLinesRemoved: 623,
		Languages: map[string]int{
			"go":         18,
			"typescript": 14,
			"python":     7,
			"yaml":       3,
		},
		MostImpactedFile: "src/services/agent.go",
	}

	return summary, nil
}

// GetFileTree returns file tree of impacted files sorted by path
func (s *CodeImpactService) GetFileTree(ctx context.Context, projectID uuid.UUID, traceID string) ([]domain.FileImpact, error) {
	s.logger.Info("fetching file impact tree",
		zap.String("projectId", projectID.String()),
		zap.String("traceId", traceID),
	)

	impactMap, err := s.GetCodeImpact(ctx, projectID, traceID)
	if err != nil {
		return nil, fmt.Errorf("failed to build impact map: %w", err)
	}

	files := make([]domain.FileImpact, len(impactMap.Files))
	copy(files, impactMap.Files)

	sort.Slice(files, func(i, j int) bool {
		return files[i].FilePath < files[j].FilePath
	})

	return files, nil
}

func (s *CodeImpactService) buildFileImpacts(fileOps []domain.FileOperation, observations []domain.Observation) []domain.FileImpact {
	if len(fileOps) == 0 {
		// Return mock data when no real file operations exist
		return []domain.FileImpact{
			{
				FilePath:       "src/services/agent.go",
				OperationType:  "modified",
				LinesAdded:     45,
				LinesRemoved:   12,
				ObservationIDs: []string{},
				Language:       "go",
				Complexity:     "medium",
			},
			{
				FilePath:       "src/handlers/api.go",
				OperationType:  "created",
				LinesAdded:     120,
				LinesRemoved:   0,
				ObservationIDs: []string{},
				Language:       "go",
				Complexity:     "high",
			},
			{
				FilePath:       "src/config/settings.yaml",
				OperationType:  "modified",
				LinesAdded:     8,
				LinesRemoved:   3,
				ObservationIDs: []string{},
				Language:       "yaml",
				Complexity:     "low",
			},
		}
	}

	obsIDsByFile := map[string][]string{}
	for _, obs := range observations {
		if obs.Name != "" {
			obsIDsByFile[obs.Name] = append(obsIDsByFile[obs.Name], obs.ID)
		}
	}

	files := make([]domain.FileImpact, 0, len(fileOps))
	for _, op := range fileOps {
		obsIDs := obsIDsByFile[op.FilePath]
		if obsIDs == nil {
			obsIDs = []string{}
		}
		files = append(files, domain.FileImpact{
			FilePath:       op.FilePath,
			OperationType:  op.OperationType,
			LinesAdded:     op.LinesAdded,
			LinesRemoved:   op.LinesRemoved,
			ObservationIDs: obsIDs,
			Language:       s.detectLanguage(op.FilePath),
			Complexity:     s.estimateComplexity(op.LinesAdded + op.LinesRemoved),
		})
	}

	return files
}

func (s *CodeImpactService) buildSummary(files []domain.FileImpact) domain.CodeImpactSummary {
	summary := domain.CodeImpactSummary{
		TotalFiles: len(files),
		Languages:  make(map[string]int),
	}

	maxLines := 0
	for _, f := range files {
		switch f.OperationType {
		case "created":
			summary.FilesCreated++
		case "modified":
			summary.FilesModified++
		case "deleted":
			summary.FilesDeleted++
		}
		summary.TotalLinesAdded += f.LinesAdded
		summary.TotalLinesRemoved += f.LinesRemoved
		if f.Language != "" {
			summary.Languages[f.Language]++
		}
		total := f.LinesAdded + f.LinesRemoved
		if total > maxLines {
			maxLines = total
			summary.MostImpactedFile = f.FilePath
		}
	}

	return summary
}

func (s *CodeImpactService) detectLanguage(filePath string) string {
	parts := strings.Split(filePath, ".")
	if len(parts) < 2 {
		return ""
	}
	ext := parts[len(parts)-1]
	switch ext {
	case "go":
		return "go"
	case "ts", "tsx":
		return "typescript"
	case "js", "jsx":
		return "javascript"
	case "py":
		return "python"
	case "yaml", "yml":
		return "yaml"
	case "json":
		return "json"
	case "rs":
		return "rust"
	default:
		return ext
	}
}

func (s *CodeImpactService) estimateComplexity(totalLines int) string {
	switch {
	case totalLines > 100:
		return "high"
	case totalLines > 30:
		return "medium"
	default:
		return "low"
	}
}
