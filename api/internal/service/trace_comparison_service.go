package service

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

// TraceComparisonRepository defines repository operations for trace comparisons
type TraceComparisonRepository interface {
	GetTraceByID(ctx context.Context, traceID string) (*domain.TraceComparisonEntry, error)
	GetTraceSpans(ctx context.Context, traceID string) ([]domain.ToolCallSummary, error)
	GetTraceSpanNames(ctx context.Context, traceID string) ([]string, error)
	SaveComparison(ctx context.Context, comparison *domain.TraceComparisonMatrix) error
	GetComparison(ctx context.Context, id uuid.UUID) (*domain.TraceComparisonMatrix, error)
	GetComparisonByShareToken(ctx context.Context, token string) (*domain.TraceComparisonMatrix, error)
}

// TraceComparisonService handles trace comparison operations
type TraceComparisonService struct {
	logger   *zap.Logger
	traceRepo TraceComparisonRepository
}

// NewTraceComparisonService creates a new trace comparison service
func NewTraceComparisonService(logger *zap.Logger, traceRepo TraceComparisonRepository) *TraceComparisonService {
	return &TraceComparisonService{
		logger:    logger,
		traceRepo: traceRepo,
	}
}

// CompareTraces creates a comparison matrix for the given trace IDs
func (s *TraceComparisonService) CompareTraces(ctx context.Context, projectID, userID uuid.UUID, input *domain.TraceComparisonInput) (*domain.TraceComparisonMatrix, error) {
	if len(input.TraceIDs) < 2 {
		return nil, fmt.Errorf("at least 2 traces are required for comparison")
	}
	if len(input.TraceIDs) > 10 {
		return nil, fmt.Errorf("maximum 10 traces can be compared at once")
	}

	// Fetch trace data
	var entries []domain.TraceComparisonEntry
	for _, traceID := range input.TraceIDs {
		entry, err := s.traceRepo.GetTraceByID(ctx, traceID)
		if err != nil {
			return nil, fmt.Errorf("failed to get trace %s: %w", traceID, err)
		}
		entries = append(entries, *entry)
	}

	// Build metric comparison grid
	metrics := s.buildMetricGrid(entries)

	// Build tool usage comparison
	toolUsage := s.buildToolUsageComparison(ctx, entries)

	// Build topology comparison
	topology := s.buildTopologyComparison(ctx, entries)

	// Generate share token
	shareToken := uuid.New().String()[:12]

	matrix := &domain.TraceComparisonMatrix{
		ID:         uuid.New(),
		ProjectID:  projectID,
		TraceIDs:   input.TraceIDs,
		Traces:     entries,
		Metrics:    *metrics,
		ToolUsage:  *toolUsage,
		Topology:   *topology,
		ShareToken: shareToken,
		CreatedAt:  time.Now(),
		CreatedBy:  userID,
	}

	if err := s.traceRepo.SaveComparison(ctx, matrix); err != nil {
		s.logger.Warn("failed to persist comparison", zap.Error(err))
	}

	s.logger.Info("trace comparison created",
		zap.String("comparisonId", matrix.ID.String()),
		zap.Int("traceCount", len(input.TraceIDs)),
	)

	return matrix, nil
}

// GetComparison retrieves a saved comparison
func (s *TraceComparisonService) GetComparison(ctx context.Context, id uuid.UUID) (*domain.TraceComparisonMatrix, error) {
	return s.traceRepo.GetComparison(ctx, id)
}

// GetComparisonByShareToken retrieves a comparison by its share token
func (s *TraceComparisonService) GetComparisonByShareToken(ctx context.Context, token string) (*domain.TraceComparisonMatrix, error) {
	return s.traceRepo.GetComparisonByShareToken(ctx, token)
}

// ExportComparison exports a comparison in the specified format
func (s *TraceComparisonService) ExportComparison(ctx context.Context, comparisonID uuid.UUID, format string) (*domain.TraceComparisonExport, error) {
	comparison, err := s.traceRepo.GetComparison(ctx, comparisonID)
	if err != nil {
		return nil, fmt.Errorf("failed to get comparison: %w", err)
	}

	switch format {
	case "csv":
		return s.exportCSV(comparison)
	case "json":
		return s.exportJSON(comparison)
	default:
		return nil, fmt.Errorf("unsupported export format: %s", format)
	}
}

func (s *TraceComparisonService) buildMetricGrid(entries []domain.TraceComparisonEntry) *domain.MetricComparisonGrid {
	grid := &domain.MetricComparisonGrid{
		Latency:      s.buildMetricRow("Latency", "ms", entries, func(e domain.TraceComparisonEntry) float64 { return float64(e.Duration) }, false),
		TotalCost:    s.buildMetricRow("Total Cost", "USD", entries, func(e domain.TraceComparisonEntry) float64 { return e.TotalCost }, false),
		TotalTokens:  s.buildMetricRow("Total Tokens", "tokens", entries, func(e domain.TraceComparisonEntry) float64 { return float64(e.TotalTokens) }, false),
		SpanCount:    s.buildMetricRow("Span Count", "spans", entries, func(e domain.TraceComparisonEntry) float64 { return float64(e.SpanCount) }, false),
		ErrorRate:    s.buildMetricRow("Error Rate", "%", entries, func(e domain.TraceComparisonEntry) float64 {
			if e.SpanCount == 0 {
				return 0
			}
			return float64(e.ErrorCount) / float64(e.SpanCount) * 100
		}, false),
	}
	return grid
}

func (s *TraceComparisonService) buildMetricRow(name, unit string, entries []domain.TraceComparisonEntry, extractFn func(domain.TraceComparisonEntry) float64, higherIsBetter bool) domain.MetricRow {
	row := domain.MetricRow{
		Name: name,
		Unit: unit,
	}

	var values []float64
	for _, entry := range entries {
		val := extractFn(entry)
		values = append(values, val)
		row.Values = append(row.Values, domain.MetricValue{
			TraceID: entry.TraceID,
			Value:   val,
		})
	}

	// Compute stats
	if len(values) > 0 {
		minVal, maxVal := values[0], values[0]
		sum := 0.0
		for _, v := range values {
			sum += v
			if v < minVal {
				minVal = v
			}
			if v > maxVal {
				maxVal = v
			}
		}

		row.Stats = domain.MetricComparisonStats{
			Min:   minVal,
			Max:   maxVal,
			Avg:   sum / float64(len(values)),
			Range: maxVal - minVal,
		}

		// Assign ranks and find best
		sorted := make([]int, len(values))
		for i := range sorted {
			sorted[i] = i
		}
		sort.Slice(sorted, func(i, j int) bool {
			if higherIsBetter {
				return values[sorted[i]] > values[sorted[j]]
			}
			return values[sorted[i]] < values[sorted[j]]
		})
		for rank, idx := range sorted {
			row.Values[idx].Rank = rank + 1
		}
		row.Stats.BestTrace = row.Values[sorted[0]].TraceID
	}

	return row
}

func (s *TraceComparisonService) buildToolUsageComparison(ctx context.Context, entries []domain.TraceComparisonEntry) *domain.ToolUsageComparison {
	comparison := &domain.ToolUsageComparison{
		ByTrace: make(map[string][]domain.ToolCallSummary),
	}

	allToolsSet := make(map[string]bool)
	traceTools := make(map[string]map[string]bool)

	for _, entry := range entries {
		spans, err := s.traceRepo.GetTraceSpans(ctx, entry.TraceID)
		if err != nil {
			s.logger.Warn("failed to get spans for trace", zap.String("traceId", entry.TraceID), zap.Error(err))
			continue
		}
		comparison.ByTrace[entry.TraceID] = spans
		traceTools[entry.TraceID] = make(map[string]bool)
		for _, span := range spans {
			allToolsSet[span.ToolName] = true
			traceTools[entry.TraceID][span.ToolName] = true
		}
	}

	for tool := range allToolsSet {
		comparison.AllTools = append(comparison.AllTools, tool)
	}
	sort.Strings(comparison.AllTools)

	// Detect divergences
	for _, tool := range comparison.AllTools {
		var presentIn, missingIn []string
		for _, entry := range entries {
			if traceTools[entry.TraceID][tool] {
				presentIn = append(presentIn, entry.TraceID)
			} else {
				missingIn = append(missingIn, entry.TraceID)
			}
		}
		if len(missingIn) > 0 && len(presentIn) > 0 {
			comparison.Divergences = append(comparison.Divergences, domain.ToolDivergence{
				ToolName:    tool,
				Type:        "missing_in",
				Description: fmt.Sprintf("Tool '%s' used in %d traces but missing in %d", tool, len(presentIn), len(missingIn)),
				TraceIDs:    missingIn,
			})
		}
	}

	return comparison
}

func (s *TraceComparisonService) buildTopologyComparison(ctx context.Context, entries []domain.TraceComparisonEntry) *domain.TopologyComparison {
	topology := &domain.TopologyComparison{
		UniqueSpans: make(map[string][]string),
	}

	spanSets := make(map[string]map[string]bool)
	allSpans := make(map[string]int) // span name -> count of traces containing it

	for _, entry := range entries {
		spanNames, err := s.traceRepo.GetTraceSpanNames(ctx, entry.TraceID)
		if err != nil {
			continue
		}
		spanSets[entry.TraceID] = make(map[string]bool)
		for _, name := range spanNames {
			spanSets[entry.TraceID][name] = true
			allSpans[name]++
		}
	}

	// Find common and unique spans
	for span, count := range allSpans {
		if count == len(entries) {
			topology.CommonSpans = append(topology.CommonSpans, span)
		}
	}
	sort.Strings(topology.CommonSpans)

	for _, entry := range entries {
		for span := range spanSets[entry.TraceID] {
			if allSpans[span] < len(entries) {
				topology.UniqueSpans[entry.TraceID] = append(topology.UniqueSpans[entry.TraceID], span)
			}
		}
	}

	// Detect structural diffs
	depths := make(map[string]int)
	for _, entry := range entries {
		depths[entry.TraceID] = entry.SpanCount
	}
	var minDepth, maxDepth int
	first := true
	for _, d := range depths {
		if first {
			minDepth, maxDepth = d, d
			first = false
		}
		if d < minDepth {
			minDepth = d
		}
		if d > maxDepth {
			maxDepth = d
		}
	}
	if float64(maxDepth-minDepth)/math.Max(float64(minDepth), 1) > 0.5 {
		topology.StructureDiffs = append(topology.StructureDiffs, domain.StructureDiff{
			Type:        "depth_diff",
			Description: fmt.Sprintf("Significant span count difference: %d to %d", minDepth, maxDepth),
		})
	}

	return topology
}

func (s *TraceComparisonService) exportCSV(comparison *domain.TraceComparisonMatrix) (*domain.TraceComparisonExport, error) {
	var buf strings.Builder
	writer := csv.NewWriter(&buf)

	// Header
	header := []string{"Metric"}
	for _, trace := range comparison.Traces {
		header = append(header, trace.TraceID)
	}
	writer.Write(header)

	// Rows
	writeMetricRow := func(row domain.MetricRow) {
		record := []string{row.Name}
		for _, v := range row.Values {
			record = append(record, fmt.Sprintf("%.4f", v.Value))
		}
		writer.Write(record)
	}

	writeMetricRow(comparison.Metrics.Latency)
	writeMetricRow(comparison.Metrics.TotalCost)
	writeMetricRow(comparison.Metrics.TotalTokens)
	writeMetricRow(comparison.Metrics.SpanCount)
	writeMetricRow(comparison.Metrics.ErrorRate)

	writer.Flush()

	return &domain.TraceComparisonExport{
		Format:   "csv",
		Data:     []byte(buf.String()),
		Filename: fmt.Sprintf("trace-comparison-%s.csv", comparison.ID.String()[:8]),
	}, nil
}

func (s *TraceComparisonService) exportJSON(comparison *domain.TraceComparisonMatrix) (*domain.TraceComparisonExport, error) {
	data, err := json.MarshalIndent(comparison, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal comparison: %w", err)
	}

	return &domain.TraceComparisonExport{
		Format:   "json",
		Data:     data,
		Filename: fmt.Sprintf("trace-comparison-%s.json", comparison.ID.String()[:8]),
	}, nil
}
