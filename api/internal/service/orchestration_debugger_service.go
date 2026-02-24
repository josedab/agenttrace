package service

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

// OrchestrationDebuggerService manages multi-agent orchestration debugging sessions
type OrchestrationDebuggerService struct {
	logger   *zap.Logger
	mu       sync.RWMutex
	sessions map[uuid.UUID]*domain.OrchestrationSession
}

// NewOrchestrationDebuggerService creates a new orchestration debugger service
func NewOrchestrationDebuggerService(logger *zap.Logger) *OrchestrationDebuggerService {
	return &OrchestrationDebuggerService{
		logger:   logger,
		sessions: make(map[uuid.UUID]*domain.OrchestrationSession),
	}
}

// CreateSession creates a new orchestration debugging session with mock topology
func (s *OrchestrationDebuggerService) CreateSession(ctx context.Context, projectID uuid.UUID, input *domain.OrchestrationSessionInput) (*domain.OrchestrationSession, error) {
	now := time.Now()
	agents := s.generateMockAgents()
	messages := s.generateMockMessages(agents, now)

	session := &domain.OrchestrationSession{
		ID:          uuid.New(),
		ProjectID:   projectID,
		TraceID:     input.TraceID,
		Agents:      agents,
		Messages:    messages,
		Breakpoints: []domain.AgentBreakpoint{},
		Status:      "running",
		CurrentStep: 0,
		TotalSteps:  len(messages),
		CreatedAt:   now,
	}

	s.mu.Lock()
	s.sessions[session.ID] = session
	s.mu.Unlock()

	s.logger.Info("created orchestration session",
		zap.String("sessionId", session.ID.String()),
		zap.String("traceId", input.TraceID.String()),
		zap.Int("agents", len(agents)),
		zap.Int("messages", len(messages)),
	)

	return session, nil
}

// GetSession retrieves an orchestration session by ID
func (s *OrchestrationDebuggerService) GetSession(ctx context.Context, sessionID uuid.UUID) (*domain.OrchestrationSession, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	session, exists := s.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("session not found")
	}
	return session, nil
}

// ExecuteCommand executes a debug command on a session
func (s *OrchestrationDebuggerService) ExecuteCommand(ctx context.Context, sessionID uuid.UUID, cmd *domain.DebugCommand) (*domain.OrchestrationSession, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	session, exists := s.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("session not found")
	}

	switch cmd.Action {
	case "step":
		steps := cmd.StepCount
		if steps <= 0 {
			steps = 1
		}
		session.CurrentStep += steps
		if session.CurrentStep >= session.TotalSteps {
			session.CurrentStep = session.TotalSteps
			session.Status = "completed"
		}
	case "continue":
		// Check for breakpoints, otherwise run to completion
		hitBreakpoint := false
		for i := session.CurrentStep + 1; i < session.TotalSteps; i++ {
			msg := session.Messages[i]
			for _, bp := range session.Breakpoints {
				if bp.Enabled && (bp.AgentID == msg.ToAgent || bp.AgentID == msg.FromAgent) {
					session.CurrentStep = i
					session.Status = "paused"
					hitBreakpoint = true
					break
				}
			}
			if hitBreakpoint {
				break
			}
		}
		if !hitBreakpoint {
			session.CurrentStep = session.TotalSteps
			session.Status = "completed"
		}
	case "step_over":
		if session.CurrentStep < session.TotalSteps-1 {
			session.CurrentStep++
		} else {
			session.Status = "completed"
		}
	case "inspect":
		// Inspect doesn't change state, just returns current session
	default:
		return nil, fmt.Errorf("unknown action: %s", cmd.Action)
	}

	s.logger.Info("executed debug command",
		zap.String("sessionId", sessionID.String()),
		zap.String("action", cmd.Action),
		zap.Int("currentStep", session.CurrentStep),
	)

	return session, nil
}

// AddBreakpoint adds a breakpoint to an orchestration session
func (s *OrchestrationDebuggerService) AddBreakpoint(ctx context.Context, sessionID uuid.UUID, bp *domain.AgentBreakpoint) (*domain.OrchestrationSession, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	session, exists := s.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("session not found")
	}

	bp.ID = uuid.New().String()
	session.Breakpoints = append(session.Breakpoints, *bp)

	s.logger.Info("added breakpoint",
		zap.String("sessionId", sessionID.String()),
		zap.String("agentId", bp.AgentID),
		zap.String("condition", bp.Condition),
	)

	return session, nil
}

// ListSessions lists all orchestration sessions for a project
func (s *OrchestrationDebuggerService) ListSessions(ctx context.Context, projectID uuid.UUID) ([]domain.OrchestrationSession, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var sessions []domain.OrchestrationSession
	for _, session := range s.sessions {
		if session.ProjectID == projectID {
			sessions = append(sessions, *session)
		}
	}

	if sessions == nil {
		sessions = []domain.OrchestrationSession{}
	}
	return sessions, nil
}

func (s *OrchestrationDebuggerService) generateMockAgents() []domain.OrchestratorAgent {
	return []domain.OrchestratorAgent{
		{
			ID:         "coordinator-1",
			Name:       "TaskCoordinator",
			Type:       domain.AgentNodeCoordinator,
			Model:      "gpt-4o",
			Status:     "active",
			TokensUsed: 2450,
			Cost:       0.0735,
		},
		{
			ID:         "worker-code",
			Name:       "CodeWriter",
			Type:       domain.AgentNodeWorker,
			Model:      "gpt-4o",
			Status:     "idle",
			TokensUsed: 8200,
			Cost:       0.246,
		},
		{
			ID:         "worker-review",
			Name:       "CodeReviewer",
			Type:       domain.AgentNodeWorker,
			Model:      "claude-3.5-sonnet",
			Status:     "idle",
			TokensUsed: 3100,
			Cost:       0.0465,
		},
		{
			ID:     "tool-search",
			Name:   "SearchTool",
			Type:   domain.AgentNodeTool,
			Status: "idle",
		},
	}
}

func (s *OrchestrationDebuggerService) generateMockMessages(agents []domain.OrchestratorAgent, baseTime time.Time) []domain.AgentMessage {
	messages := []domain.AgentMessage{
		{ID: uuid.New().String(), FromAgent: "coordinator-1", ToAgent: "tool-search", Direction: domain.MessageRequest, Content: "Search for existing implementations of authentication middleware", Timestamp: baseTime, LatencyMs: 0, TokenCount: 42, StepIndex: 0},
		{ID: uuid.New().String(), FromAgent: "tool-search", ToAgent: "coordinator-1", Direction: domain.MessageResponse, Content: "Found 3 relevant files: auth.go, middleware.go, jwt.go", Timestamp: baseTime.Add(800 * time.Millisecond), LatencyMs: 800, TokenCount: 85, StepIndex: 1},
		{ID: uuid.New().String(), FromAgent: "coordinator-1", ToAgent: "worker-code", Direction: domain.MessageRequest, Content: "Implement JWT-based authentication middleware using the existing patterns found", Timestamp: baseTime.Add(1 * time.Second), LatencyMs: 200, TokenCount: 156, StepIndex: 2},
		{ID: uuid.New().String(), FromAgent: "worker-code", ToAgent: "coordinator-1", Direction: domain.MessageResponse, Content: "Generated authentication middleware with JWT validation, token refresh, and role-based access control", Timestamp: baseTime.Add(4 * time.Second), LatencyMs: 3000, TokenCount: 1250, StepIndex: 3},
		{ID: uuid.New().String(), FromAgent: "coordinator-1", ToAgent: "worker-review", Direction: domain.MessageRequest, Content: "Review the generated authentication middleware for security vulnerabilities", Timestamp: baseTime.Add(4500 * time.Millisecond), LatencyMs: 500, TokenCount: 1300, StepIndex: 4},
		{ID: uuid.New().String(), FromAgent: "worker-review", ToAgent: "coordinator-1", Direction: domain.MessageResponse, Content: "Review complete. Found 2 issues: missing token expiry validation, hardcoded secret key", Timestamp: baseTime.Add(7 * time.Second), LatencyMs: 2500, TokenCount: 420, StepIndex: 5},
		{ID: uuid.New().String(), FromAgent: "coordinator-1", ToAgent: "worker-code", Direction: domain.MessageRequest, Content: "Fix the security issues: add token expiry validation and use environment variable for secret key", Timestamp: baseTime.Add(7500 * time.Millisecond), LatencyMs: 500, TokenCount: 180, StepIndex: 6},
		{ID: uuid.New().String(), FromAgent: "worker-code", ToAgent: "coordinator-1", Direction: domain.MessageResponse, Content: "Fixed both issues. Token expiry is now validated, secret key loaded from SIGNING_KEY env var", Timestamp: baseTime.Add(10 * time.Second), LatencyMs: 2500, TokenCount: 890, StepIndex: 7},
		{ID: uuid.New().String(), FromAgent: "coordinator-1", ToAgent: "worker-review", Direction: domain.MessageRequest, Content: "Final review of the updated authentication middleware", Timestamp: baseTime.Add(10500 * time.Millisecond), LatencyMs: 500, TokenCount: 950, StepIndex: 8},
		{ID: uuid.New().String(), FromAgent: "worker-review", ToAgent: "coordinator-1", Direction: domain.MessageResponse, Content: "Approved. All security concerns addressed. Code follows project conventions.", Timestamp: baseTime.Add(13 * time.Second), LatencyMs: 2500, TokenCount: 210, StepIndex: 9},
		{ID: uuid.New().String(), FromAgent: "coordinator-1", ToAgent: "worker-code", Direction: domain.MessageBroadcast, Content: "Task complete. Authentication middleware implemented and reviewed.", Timestamp: baseTime.Add(13500 * time.Millisecond), LatencyMs: 500, TokenCount: 65, StepIndex: 10},
	}

	// Add slight random variation to latencies
	for i := range messages {
		if messages[i].LatencyMs > 0 {
			messages[i].LatencyMs += int64(rand.Intn(200)) //nolint:gosec
		}
	}

	return messages
}
