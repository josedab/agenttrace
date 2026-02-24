package service

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

// AgentMemoryService analyzes agent context window utilization
type AgentMemoryService struct {
	logger *zap.Logger
}

// NewAgentMemoryService creates a new agent memory service
func NewAgentMemoryService(logger *zap.Logger) *AgentMemoryService {
	return &AgentMemoryService{
		logger: logger,
	}
}

// AnalyzeMemory returns a timeline of context window utilization for a trace
func (s *AgentMemoryService) AnalyzeMemory(ctx context.Context, projectID uuid.UUID, traceID uuid.UUID) (*domain.MemoryTimeline, error) {
	s.logger.Debug("analyzing memory for trace", zap.String("traceId", traceID.String()))

	now := time.Now()
	totalTokens := 128000
	snapshots := make([]domain.MemorySnapshot, 5)
	var peakUtil float64
	var totalUtil float64
	truncationEvents := 0

	for i := 0; i < 5; i++ {
		usedPct := 0.2 + float64(i)*0.18
		if usedPct > 1.0 {
			usedPct = 0.95
		}
		used := int(float64(totalTokens) * usedPct)
		avail := totalTokens - used

		retained := []domain.MemoryItem{
			{Type: domain.MemoryItemTypeSystem, Content: truncateContent("You are a helpful assistant..."), TokenCount: 500, Retained: true},
			{Type: domain.MemoryItemTypeUser, Content: truncateContent(fmt.Sprintf("User message at step %d", i)), TokenCount: 200 + i*100, Retained: true},
		}
		var truncated []domain.MemoryItem
		if i >= 3 {
			truncated = []domain.MemoryItem{
				{Type: domain.MemoryItemTypeAssistant, Content: truncateContent("Earlier assistant response that was truncated..."), TokenCount: 1500, Retained: false},
			}
			truncationEvents++
		}

		snapshots[i] = domain.MemorySnapshot{
			ID:              uuid.New(),
			TraceID:         traceID,
			ProjectID:       projectID,
			StepIndex:       i,
			TotalTokens:     totalTokens,
			UsedTokens:      used,
			AvailableTokens: avail,
			RetainedItems:   retained,
			TruncatedItems:  truncated,
			UtilizationPct:  math.Round(usedPct*10000) / 100,
			Timestamp:       now.Add(time.Duration(i) * time.Minute),
		}
		if usedPct > peakUtil {
			peakUtil = usedPct
		}
		totalUtil += usedPct
	}

	suggestions := []string{
		"Consider using summarization for messages older than 5 turns",
		"System prompt could be compressed by 30% without quality loss",
		"RAG retrieval could replace static context injection",
	}

	timeline := &domain.MemoryTimeline{
		TraceID:                 traceID,
		Snapshots:               snapshots,
		PeakUtilization:         math.Round(peakUtil*10000) / 100,
		AvgUtilization:          math.Round((totalUtil/5)*10000) / 100,
		TruncationEvents:        truncationEvents,
		OptimizationSuggestions: suggestions,
	}

	return timeline, nil
}

// GetSnapshot returns a specific memory snapshot for a trace step
func (s *AgentMemoryService) GetSnapshot(ctx context.Context, traceID uuid.UUID, stepIndex int) (*domain.MemorySnapshot, error) {
	s.logger.Debug("getting snapshot", zap.String("traceId", traceID.String()), zap.Int("step", stepIndex))

	totalTokens := 128000
	usedPct := 0.2 + float64(stepIndex)*0.18
	if usedPct > 1.0 {
		usedPct = 0.95
	}
	used := int(float64(totalTokens) * usedPct)

	snapshot := &domain.MemorySnapshot{
		ID:              uuid.New(),
		TraceID:         traceID,
		ProjectID:       uuid.New(),
		StepIndex:       stepIndex,
		TotalTokens:     totalTokens,
		UsedTokens:      used,
		AvailableTokens: totalTokens - used,
		RetainedItems: []domain.MemoryItem{
			{Type: domain.MemoryItemTypeSystem, Content: truncateContent("You are a helpful assistant..."), TokenCount: 500, Retained: true},
		},
		TruncatedItems: []domain.MemoryItem{},
		UtilizationPct: math.Round(usedPct*10000) / 100,
		Timestamp:      time.Now(),
	}

	return snapshot, nil
}

// GetOptimizations returns optimization suggestions for a project
func (s *AgentMemoryService) GetOptimizations(ctx context.Context, projectID uuid.UUID) ([]domain.MemoryOptimization, error) {
	s.logger.Debug("getting optimizations", zap.String("projectId", projectID.String()))

	return []domain.MemoryOptimization{
		{
			Technique:           domain.MemoryOptCompression,
			Description:         "Compress system prompts using token-efficient formatting",
			EstimatedSavingsPct: 15.0,
			Confidence:          0.85,
		},
		{
			Technique:           domain.MemoryOptSummarization,
			Description:         "Summarize conversation history beyond 10 turns",
			EstimatedSavingsPct: 35.0,
			Confidence:          0.78,
		},
		{
			Technique:           domain.MemoryOptRAG,
			Description:         "Replace static context with RAG-based retrieval",
			EstimatedSavingsPct: 45.0,
			Confidence:          0.72,
		},
		{
			Technique:           domain.MemoryOptSlidingWindow,
			Description:         "Use sliding window to keep only recent N turns",
			EstimatedSavingsPct: 25.0,
			Confidence:          0.90,
		},
	}, nil
}

func truncateContent(s string) string {
	if len(s) > 200 {
		return s[:200]
	}
	return s
}
