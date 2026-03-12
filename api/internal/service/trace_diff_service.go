package service

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

// TraceDiffService handles trace diffing and regression bisect operations
type TraceDiffService struct {
	logger   *zap.Logger
	query    *QueryService
	sessions map[uuid.UUID]*domain.BisectSession
}

// NewTraceDiffService creates a new trace diff service
func NewTraceDiffService(logger *zap.Logger, query *QueryService) *TraceDiffService {
	return &TraceDiffService{
		logger:   logger,
		query:    query,
		sessions: make(map[uuid.UUID]*domain.BisectSession),
	}
}

// DiffTraces computes a structural diff between two traces
func (s *TraceDiffService) DiffTraces(ctx context.Context, projectID uuid.UUID, input *domain.TraceDiffInput) (*domain.TraceDiffResult, error) {
	if input.LeftTraceID == input.RightTraceID {
		return nil, fmt.Errorf("cannot diff a trace with itself")
	}

	leftObs, err := s.query.GetObservationsByTraceID(ctx, projectID, input.LeftTraceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get left trace observations: %w", err)
	}

	rightObs, err := s.query.GetObservationsByTraceID(ctx, projectID, input.RightTraceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get right trace observations: %w", err)
	}

	leftTree := domain.BuildObservationTree(leftObs)
	rightTree := domain.BuildObservationTree(rightObs)

	rootDiffs := s.diffTrees(leftTree, rightTree)
	summary := s.computeSummary(rootDiffs, leftObs, rightObs)

	result := &domain.TraceDiffResult{
		ID:           uuid.New(),
		ProjectID:    projectID,
		LeftTraceID:  input.LeftTraceID,
		RightTraceID: input.RightTraceID,
		RootDiffs:    rootDiffs,
		Summary:      summary,
		CreatedAt:    time.Now(),
	}

	s.logger.Info("trace diff computed",
		zap.String("leftTraceId", input.LeftTraceID),
		zap.String("rightTraceId", input.RightTraceID),
		zap.Int("totalDiffs", summary.AddedCount+summary.RemovedCount+summary.ModifiedCount),
	)

	return result, nil
}

// diffTrees performs a tree-level diff matching by span name
func (s *TraceDiffService) diffTrees(left, right []*domain.ObservationTree) []*domain.TraceDiffNode {
	var diffs []*domain.TraceDiffNode

	leftMap := make(map[string]*domain.ObservationTree)
	rightMap := make(map[string]*domain.ObservationTree)
	var leftOrder, rightOrder []string

	for _, node := range left {
		key := node.Observation.Name
		leftMap[key] = node
		leftOrder = append(leftOrder, key)
	}
	for _, node := range right {
		key := node.Observation.Name
		rightMap[key] = node
		rightOrder = append(rightOrder, key)
	}

	matched := make(map[string]bool)

	// Match by name
	for _, name := range leftOrder {
		leftNode := leftMap[name]
		if rightNode, ok := rightMap[name]; ok {
			matched[name] = true
			diff := s.diffNodes(leftNode, rightNode)
			diffs = append(diffs, diff)
		} else {
			diffs = append(diffs, s.removedNode(leftNode))
		}
	}

	for _, name := range rightOrder {
		if !matched[name] {
			diffs = append(diffs, s.addedNode(rightMap[name]))
		}
	}

	return diffs
}

func (s *TraceDiffService) diffNodes(left, right *domain.ObservationTree) *domain.TraceDiffNode {
	leftSnap := s.snapshotSpan(left.Observation)
	rightSnap := s.snapshotSpan(right.Observation)

	propDiffs := s.diffProperties(leftSnap, rightSnap)
	diffType := domain.DiffTypeUnchanged
	if len(propDiffs) > 0 {
		diffType = domain.DiffTypeModified
	}

	childDiffs := s.diffTrees(left.Children, right.Children)

	return &domain.TraceDiffNode{
		DiffType:      diffType,
		SpanName:      left.Observation.Name,
		LeftSpanID:    left.Observation.ID,
		RightSpanID:   right.Observation.ID,
		LeftValue:     leftSnap,
		RightValue:    rightSnap,
		PropertyDiffs: propDiffs,
		Children:      childDiffs,
	}
}

func (s *TraceDiffService) addedNode(node *domain.ObservationTree) *domain.TraceDiffNode {
	snap := s.snapshotSpan(node.Observation)
	children := make([]*domain.TraceDiffNode, 0, len(node.Children))
	for _, child := range node.Children {
		children = append(children, s.addedNode(child))
	}
	return &domain.TraceDiffNode{
		DiffType:    domain.DiffTypeAdded,
		SpanName:    node.Observation.Name,
		RightSpanID: node.Observation.ID,
		RightValue:  snap,
		Children:    children,
	}
}

func (s *TraceDiffService) removedNode(node *domain.ObservationTree) *domain.TraceDiffNode {
	snap := s.snapshotSpan(node.Observation)
	children := make([]*domain.TraceDiffNode, 0, len(node.Children))
	for _, child := range node.Children {
		children = append(children, s.removedNode(child))
	}
	return &domain.TraceDiffNode{
		DiffType:   domain.DiffTypeRemoved,
		SpanName:   node.Observation.Name,
		LeftSpanID: node.Observation.ID,
		LeftValue:  snap,
		Children:   children,
	}
}

func (s *TraceDiffService) snapshotSpan(obs *domain.Observation) *domain.SpanSnapshot {
	return &domain.SpanSnapshot{
		ID:          obs.ID,
		Name:        obs.Name,
		Type:        string(obs.Type),
		Model:       obs.Model,
		DurationMs:  obs.DurationMs,
		TotalTokens: obs.UsageDetails.TotalTokens,
		TotalCost:   obs.CostDetails.TotalCost,
		Input:       obs.Input,
		Output:      obs.Output,
		Level:       string(obs.Level),
	}
}

func (s *TraceDiffService) diffProperties(left, right *domain.SpanSnapshot) []domain.PropertyDiff {
	var diffs []domain.PropertyDiff

	if left.Model != right.Model {
		diffs = append(diffs, domain.PropertyDiff{Property: "model", LeftValue: left.Model, RightValue: right.Model, ChangeType: "changed"})
	}
	if left.DurationMs != right.DurationMs {
		ct := "increased"
		if right.DurationMs < left.DurationMs {
			ct = "decreased"
		}
		diffs = append(diffs, domain.PropertyDiff{Property: "durationMs", LeftValue: left.DurationMs, RightValue: right.DurationMs, ChangeType: ct})
	}
	if left.TotalTokens != right.TotalTokens {
		ct := "increased"
		if right.TotalTokens < left.TotalTokens {
			ct = "decreased"
		}
		diffs = append(diffs, domain.PropertyDiff{Property: "totalTokens", LeftValue: left.TotalTokens, RightValue: right.TotalTokens, ChangeType: ct})
	}
	if left.TotalCost != right.TotalCost {
		ct := "increased"
		if right.TotalCost < left.TotalCost {
			ct = "decreased"
		}
		diffs = append(diffs, domain.PropertyDiff{Property: "totalCost", LeftValue: left.TotalCost, RightValue: right.TotalCost, ChangeType: ct})
	}
	if left.Level != right.Level {
		diffs = append(diffs, domain.PropertyDiff{Property: "level", LeftValue: left.Level, RightValue: right.Level, ChangeType: "changed"})
	}

	return diffs
}

func (s *TraceDiffService) computeSummary(diffs []*domain.TraceDiffNode, leftObs, rightObs []domain.Observation) domain.TraceDiffSummary {
	summary := domain.TraceDiffSummary{}
	s.countDiffs(diffs, &summary)

	var leftCost, rightCost, leftDur, rightDur float64
	var leftTokens, rightTokens int64
	for _, o := range leftObs {
		leftCost += o.CostDetails.TotalCost
		leftDur += o.DurationMs
		leftTokens += int64(o.UsageDetails.TotalTokens)
	}
	for _, o := range rightObs {
		rightCost += o.CostDetails.TotalCost
		rightDur += o.DurationMs
		rightTokens += int64(o.UsageDetails.TotalTokens)
	}

	summary.CostDelta = rightCost - leftCost
	summary.LatencyDelta = rightDur - leftDur
	summary.TokenDelta = rightTokens - leftTokens
	return summary
}

func (s *TraceDiffService) countDiffs(diffs []*domain.TraceDiffNode, summary *domain.TraceDiffSummary) {
	for _, d := range diffs {
		summary.TotalNodes++
		switch d.DiffType {
		case domain.DiffTypeAdded:
			summary.AddedCount++
		case domain.DiffTypeRemoved:
			summary.RemovedCount++
		case domain.DiffTypeModified:
			summary.ModifiedCount++
		case domain.DiffTypeUnchanged:
			summary.UnchangedCount++
		}
		if d.Children != nil {
			s.countDiffs(d.Children, summary)
		}
	}
}

// StartBisect begins a new regression bisect session
func (s *TraceDiffService) StartBisect(ctx context.Context, projectID, userID uuid.UUID, input *domain.BisectStartInput) (*domain.BisectSession, error) {
	if len(input.TraceHistory) < 2 {
		return nil, fmt.Errorf("trace history must contain at least 2 entries")
	}

	// Validate good/bad traces exist in history
	goodIdx, badIdx := -1, -1
	for i, tid := range input.TraceHistory {
		if tid == input.GoodTraceID {
			goodIdx = i
		}
		if tid == input.BadTraceID {
			badIdx = i
		}
	}
	if goodIdx == -1 || badIdx == -1 {
		return nil, fmt.Errorf("good and bad traces must be in trace history")
	}
	if goodIdx >= badIdx {
		return nil, fmt.Errorf("good trace must precede bad trace in history")
	}

	midIdx := (goodIdx + badIdx) / 2

	session := &domain.BisectSession{
		ID:           uuid.New(),
		ProjectID:    projectID,
		Status:       domain.BisectStatusActive,
		GoodTraceID:  input.GoodTraceID,
		BadTraceID:   input.BadTraceID,
		TraceHistory: input.TraceHistory,
		CurrentIndex: midIdx,
		LowIndex:     goodIdx,
		HighIndex:    badIdx,
		Steps:        []domain.BisectStep{},
		MetricName:   input.MetricName,
		Threshold:    input.Threshold,
		CreatedAt:    time.Now(),
		CreatedBy:    userID,
	}

	s.sessions[session.ID] = session

	s.logger.Info("bisect session started",
		zap.String("sessionId", session.ID.String()),
		zap.Int("historySize", len(input.TraceHistory)),
	)

	return session, nil
}

// GetBisectSession retrieves a bisect session by ID
func (s *TraceDiffService) GetBisectSession(ctx context.Context, sessionID uuid.UUID) (*domain.BisectSession, error) {
	session, ok := s.sessions[sessionID]
	if !ok {
		return nil, fmt.Errorf("bisect session not found: %s", sessionID)
	}
	return session, nil
}

// SubmitBisectVerdict processes a verdict for the current bisect step
func (s *TraceDiffService) SubmitBisectVerdict(ctx context.Context, sessionID uuid.UUID, input *domain.BisectVerdictInput) (*domain.BisectSession, error) {
	session, ok := s.sessions[sessionID]
	if !ok {
		return nil, fmt.Errorf("bisect session not found: %s", sessionID)
	}
	if session.Status != domain.BisectStatusActive {
		return nil, fmt.Errorf("bisect session is not active")
	}

	step := domain.BisectStep{
		StepNumber:  len(session.Steps) + 1,
		TraceID:     session.TraceHistory[session.CurrentIndex],
		TraceIndex:  session.CurrentIndex,
		Verdict:     input.Verdict,
		Timestamp:   time.Now(),
	}
	session.Steps = append(session.Steps, step)

	switch input.Verdict {
	case "good":
		session.LowIndex = session.CurrentIndex
	case "bad":
		session.HighIndex = session.CurrentIndex
	case "skip":
		// Move slightly to avoid being stuck
		if session.CurrentIndex+1 < session.HighIndex {
			session.CurrentIndex++
		} else if session.CurrentIndex-1 > session.LowIndex {
			session.CurrentIndex--
		}
	}

	// Check if bisect is complete
	if session.HighIndex-session.LowIndex <= 1 {
		session.Status = domain.BisectStatusCompleted
		now := time.Now()
		session.CompletedAt = &now
		session.RegressionSpan = session.TraceHistory[session.HighIndex]

		s.logger.Info("bisect session completed",
			zap.String("sessionId", session.ID.String()),
			zap.String("regressionTrace", session.RegressionSpan),
			zap.Int("steps", len(session.Steps)),
		)
	} else {
		session.CurrentIndex = (session.LowIndex + session.HighIndex) / 2
	}

	return session, nil
}

// GetBisectResult returns the final result of a completed bisect
func (s *TraceDiffService) GetBisectResult(ctx context.Context, sessionID uuid.UUID) (*domain.BisectResult, error) {
	session, ok := s.sessions[sessionID]
	if !ok {
		return nil, fmt.Errorf("bisect session not found: %s", sessionID)
	}
	if session.Status != domain.BisectStatusCompleted {
		return nil, fmt.Errorf("bisect session is not completed")
	}

	result := &domain.BisectResult{
		SessionID:       session.ID,
		RegressionTrace: session.TraceHistory[session.HighIndex],
		PreviousTrace:   session.TraceHistory[session.LowIndex],
		StepsTaken:      len(session.Steps),
		RegressionSpan:  session.RegressionSpan,
	}

	// Auto-diff the regression boundary
	diff, err := s.DiffTraces(ctx, session.ProjectID, &domain.TraceDiffInput{
		LeftTraceID:  result.PreviousTrace,
		RightTraceID: result.RegressionTrace,
	})
	if err == nil {
		result.Diff = diff
	}

	return result, nil
}

// ListBisectSessions lists active bisect sessions for a project
func (s *TraceDiffService) ListBisectSessions(ctx context.Context, projectID uuid.UUID) ([]domain.BisectSession, error) {
	var sessions []domain.BisectSession
	for _, session := range s.sessions {
		if session.ProjectID == projectID {
			sessions = append(sessions, *session)
		}
	}
	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].CreatedAt.After(sessions[j].CreatedAt)
	})
	return sessions, nil
}
