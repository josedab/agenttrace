package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

// DebugProtocolMessage represents a message in the debug protocol over WebSocket
type DebugProtocolMessage struct {
	Type      string          `json:"type"`
	SessionID string          `json:"sessionId"`
	Payload   json.RawMessage `json:"payload"`
	Timestamp time.Time       `json:"timestamp"`
}

// Debug protocol message types
const (
	DebugMsgConnect       = "connect"
	DebugMsgDisconnect    = "disconnect"
	DebugMsgSetBreakpoint = "set_breakpoint"
	DebugMsgRemoveBreak   = "remove_breakpoint"
	DebugMsgStepOver      = "step_over"
	DebugMsgStepInto      = "step_into"
	DebugMsgContinue      = "continue"
	DebugMsgPause         = "pause"
	DebugMsgInspectState  = "inspect_state"
	DebugMsgEvalExpr      = "eval_expression"
	DebugMsgBreakpointHit = "breakpoint_hit"
	DebugMsgStateUpdate   = "state_update"
	DebugMsgTraceEvent    = "trace_event"
	DebugMsgCostUpdate    = "cost_update"
	DebugMsgTokenStream   = "token_stream"
	DebugMsgError         = "error"
)

// BreakpointType defines when a breakpoint should trigger
type BreakpointType string

const (
	BreakpointOnToolCall     BreakpointType = "tool_call"
	BreakpointOnCostThreshold BreakpointType = "cost_threshold"
	BreakpointOnPatternMatch BreakpointType = "pattern_match"
	BreakpointOnError        BreakpointType = "error"
	BreakpointOnSpanStart    BreakpointType = "span_start"
	BreakpointOnSpanEnd      BreakpointType = "span_end"
)

// DebugBreakpoint defines a breakpoint in the debug session
type DebugBreakpoint struct {
	ID           string         `json:"id"`
	Type         BreakpointType `json:"type"`
	Condition    string         `json:"condition,omitempty"`
	ToolName     string         `json:"toolName,omitempty"`     // For tool_call breakpoints
	CostLimit    float64        `json:"costLimit,omitempty"`    // For cost_threshold breakpoints
	Pattern      string         `json:"pattern,omitempty"`      // For pattern_match breakpoints
	SpanName     string         `json:"spanName,omitempty"`     // For span breakpoints
	Enabled      bool           `json:"enabled"`
	HitCount     int            `json:"hitCount"`
}

// DebugSessionState represents the full state of a debug session
type DebugSessionState struct {
	SessionID     string                   `json:"sessionId"`
	TraceID       string                   `json:"traceId"`
	Status        domain.DebugSessionStatus `json:"status"`
	CurrentStep   int                      `json:"currentStep"`
	TotalSteps    int                      `json:"totalSteps"`
	Breakpoints   []DebugBreakpoint        `json:"breakpoints"`
	Variables     map[string]interface{}   `json:"variables"`
	CostSoFar     float64                  `json:"costSoFar"`
	TokensSoFar   int                      `json:"tokensSoFar"`
	ElapsedMs     int64                    `json:"elapsedMs"`
	CurrentSpan   *SpanInspection          `json:"currentSpan,omitempty"`
	Timeline      []TimelineEntry          `json:"timeline"`
}

// SpanInspection provides detailed inspection of the current span
type SpanInspection struct {
	SpanID     string                 `json:"spanId"`
	SpanName   string                 `json:"spanName"`
	Type       string                 `json:"type"`
	Input      interface{}            `json:"input,omitempty"`
	Output     interface{}            `json:"output,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	StartTime  time.Time              `json:"startTime"`
	Duration   int64                  `json:"durationMs"`
	Cost       float64                `json:"cost"`
	Tokens     int                    `json:"tokens"`
	ParentID   string                 `json:"parentId,omitempty"`
	Children   []string               `json:"children,omitempty"`
}

// TimelineEntry represents a step in the trace timeline
type TimelineEntry struct {
	Index     int       `json:"index"`
	Timestamp time.Time `json:"timestamp"`
	Type      string    `json:"type"` // span_start, span_end, generation, tool_call, error
	Name      string    `json:"name"`
	Duration  int64     `json:"durationMs,omitempty"`
	Cost      float64   `json:"cost,omitempty"`
	HasError  bool      `json:"hasError"`
}

// TokenStreamData represents streaming token data
type TokenStreamData struct {
	Token    string `json:"token"`
	Index    int    `json:"index"`
	Model    string `json:"model"`
	SpanID   string `json:"spanId"`
}

// DebugSessionManager manages active debug sessions
type DebugSessionManager struct {
	logger   *zap.Logger
	sessions map[string]*activeDebugSession
	mu       sync.RWMutex
}

type activeDebugSession struct {
	state      *DebugSessionState
	sendCh     chan DebugProtocolMessage
	receiveCh  chan DebugProtocolMessage
	doneCh     chan struct{}
	createdAt  time.Time
	lastActive time.Time
}

// NewDebugSessionManager creates a new debug session manager
func NewDebugSessionManager(logger *zap.Logger) *DebugSessionManager {
	return &DebugSessionManager{
		logger:   logger,
		sessions: make(map[string]*activeDebugSession),
	}
}

// CreateDebugSession creates a new interactive debug session
func (m *DebugSessionManager) CreateDebugSession(ctx context.Context, projectID, userID uuid.UUID, traceID string) (*DebugSessionState, error) {
	sessionID := uuid.New().String()

	state := &DebugSessionState{
		SessionID:   sessionID,
		TraceID:     traceID,
		Status:      domain.DebugSessionActive,
		CurrentStep: 0,
		Breakpoints: []DebugBreakpoint{},
		Variables:   make(map[string]interface{}),
		Timeline:    []TimelineEntry{},
	}

	session := &activeDebugSession{
		state:      state,
		sendCh:     make(chan DebugProtocolMessage, 100),
		receiveCh:  make(chan DebugProtocolMessage, 100),
		doneCh:     make(chan struct{}),
		createdAt:  time.Now(),
		lastActive: time.Now(),
	}

	m.mu.Lock()
	m.sessions[sessionID] = session
	m.mu.Unlock()

	m.logger.Info("debug session created",
		zap.String("sessionId", sessionID),
		zap.String("traceId", traceID),
	)

	return state, nil
}

// GetSession retrieves an active debug session
func (m *DebugSessionManager) GetSession(sessionID string) (*DebugSessionState, error) {
	m.mu.RLock()
	session, exists := m.sessions[sessionID]
	m.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("debug session not found: %s", sessionID)
	}

	return session.state, nil
}

// HandleMessage processes an incoming debug protocol message
func (m *DebugSessionManager) HandleMessage(ctx context.Context, msg DebugProtocolMessage) (*DebugProtocolMessage, error) {
	m.mu.RLock()
	session, exists := m.sessions[msg.SessionID]
	m.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("debug session not found: %s", msg.SessionID)
	}

	session.lastActive = time.Now()

	var response *DebugProtocolMessage

	switch msg.Type {
	case DebugMsgSetBreakpoint:
		response = m.handleSetBreakpoint(session, msg)
	case DebugMsgRemoveBreak:
		response = m.handleRemoveBreakpoint(session, msg)
	case DebugMsgStepOver:
		response = m.handleStepOver(session)
	case DebugMsgStepInto:
		response = m.handleStepInto(session)
	case DebugMsgContinue:
		response = m.handleContinue(session)
	case DebugMsgPause:
		response = m.handlePause(session)
	case DebugMsgInspectState:
		response = m.handleInspectState(session)
	default:
		return &DebugProtocolMessage{
			Type:      DebugMsgError,
			SessionID: msg.SessionID,
			Payload:   mustMarshal(map[string]string{"error": "unknown message type: " + msg.Type}),
			Timestamp: time.Now(),
		}, nil
	}

	return response, nil
}

// CloseSession terminates a debug session
func (m *DebugSessionManager) CloseSession(sessionID string) error {
	m.mu.Lock()
	session, exists := m.sessions[sessionID]
	if exists {
		session.state.Status = domain.DebugSessionComplete
		close(session.doneCh)
		delete(m.sessions, sessionID)
	}
	m.mu.Unlock()

	if !exists {
		return fmt.Errorf("debug session not found: %s", sessionID)
	}

	m.logger.Info("debug session closed", zap.String("sessionId", sessionID))
	return nil
}

// ListActiveSessions returns all active debug sessions
func (m *DebugSessionManager) ListActiveSessions() []DebugSessionState {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var sessions []DebugSessionState
	for _, s := range m.sessions {
		sessions = append(sessions, *s.state)
	}
	return sessions
}

// CleanupStaleSessions removes sessions that have been inactive
func (m *DebugSessionManager) CleanupStaleSessions(maxIdle time.Duration) int {
	m.mu.Lock()
	defer m.mu.Unlock()

	cleaned := 0
	cutoff := time.Now().Add(-maxIdle)
	for id, session := range m.sessions {
		if session.lastActive.Before(cutoff) {
			session.state.Status = domain.DebugSessionComplete
			close(session.doneCh)
			delete(m.sessions, id)
			cleaned++
		}
	}

	if cleaned > 0 {
		m.logger.Info("cleaned up stale debug sessions", zap.Int("count", cleaned))
	}

	return cleaned
}

func (m *DebugSessionManager) handleSetBreakpoint(session *activeDebugSession, msg DebugProtocolMessage) *DebugProtocolMessage {
	var bp DebugBreakpoint
	json.Unmarshal(msg.Payload, &bp)

	if bp.ID == "" {
		bp.ID = uuid.New().String()[:8]
	}
	bp.Enabled = true
	bp.HitCount = 0

	session.state.Breakpoints = append(session.state.Breakpoints, bp)

	return &DebugProtocolMessage{
		Type:      DebugMsgStateUpdate,
		SessionID: msg.SessionID,
		Payload:   mustMarshal(session.state),
		Timestamp: time.Now(),
	}
}

func (m *DebugSessionManager) handleRemoveBreakpoint(session *activeDebugSession, msg DebugProtocolMessage) *DebugProtocolMessage {
	var req struct{ ID string `json:"id"` }
	json.Unmarshal(msg.Payload, &req)

	for i, bp := range session.state.Breakpoints {
		if bp.ID == req.ID {
			session.state.Breakpoints = append(session.state.Breakpoints[:i], session.state.Breakpoints[i+1:]...)
			break
		}
	}

	return &DebugProtocolMessage{
		Type:      DebugMsgStateUpdate,
		SessionID: msg.SessionID,
		Payload:   mustMarshal(session.state),
		Timestamp: time.Now(),
	}
}

func (m *DebugSessionManager) handleStepOver(session *activeDebugSession) *DebugProtocolMessage {
	if session.state.CurrentStep < session.state.TotalSteps-1 {
		session.state.CurrentStep++
	}
	session.state.Status = domain.DebugSessionPaused

	return &DebugProtocolMessage{
		Type:      DebugMsgStateUpdate,
		SessionID: session.state.SessionID,
		Payload:   mustMarshal(session.state),
		Timestamp: time.Now(),
	}
}

func (m *DebugSessionManager) handleStepInto(session *activeDebugSession) *DebugProtocolMessage {
	if session.state.CurrentStep < session.state.TotalSteps-1 {
		session.state.CurrentStep++
	}
	session.state.Status = domain.DebugSessionPaused

	return &DebugProtocolMessage{
		Type:      DebugMsgStateUpdate,
		SessionID: session.state.SessionID,
		Payload:   mustMarshal(session.state),
		Timestamp: time.Now(),
	}
}

func (m *DebugSessionManager) handleContinue(session *activeDebugSession) *DebugProtocolMessage {
	session.state.Status = domain.DebugSessionActive

	// Advance to next breakpoint or end
	for session.state.CurrentStep < session.state.TotalSteps-1 {
		session.state.CurrentStep++
		if m.checkBreakpoints(session) {
			session.state.Status = domain.DebugSessionPaused
			break
		}
	}

	return &DebugProtocolMessage{
		Type:      DebugMsgStateUpdate,
		SessionID: session.state.SessionID,
		Payload:   mustMarshal(session.state),
		Timestamp: time.Now(),
	}
}

func (m *DebugSessionManager) handlePause(session *activeDebugSession) *DebugProtocolMessage {
	session.state.Status = domain.DebugSessionPaused

	return &DebugProtocolMessage{
		Type:      DebugMsgStateUpdate,
		SessionID: session.state.SessionID,
		Payload:   mustMarshal(session.state),
		Timestamp: time.Now(),
	}
}

func (m *DebugSessionManager) handleInspectState(session *activeDebugSession) *DebugProtocolMessage {
	return &DebugProtocolMessage{
		Type:      DebugMsgStateUpdate,
		SessionID: session.state.SessionID,
		Payload:   mustMarshal(session.state),
		Timestamp: time.Now(),
	}
}

func (m *DebugSessionManager) checkBreakpoints(session *activeDebugSession) bool {
	if session.state.CurrentStep >= len(session.state.Timeline) {
		return false
	}

	entry := session.state.Timeline[session.state.CurrentStep]

	for i, bp := range session.state.Breakpoints {
		if !bp.Enabled {
			continue
		}

		matched := false
		switch bp.Type {
		case BreakpointOnToolCall:
			matched = entry.Type == "tool_call" && (bp.ToolName == "" || bp.ToolName == entry.Name)
		case BreakpointOnCostThreshold:
			matched = session.state.CostSoFar >= bp.CostLimit
		case BreakpointOnError:
			matched = entry.HasError
		case BreakpointOnSpanStart:
			matched = entry.Type == "span_start" && (bp.SpanName == "" || bp.SpanName == entry.Name)
		case BreakpointOnSpanEnd:
			matched = entry.Type == "span_end" && (bp.SpanName == "" || bp.SpanName == entry.Name)
		}

		if matched {
			session.state.Breakpoints[i].HitCount++
			return true
		}
	}

	return false
}

func mustMarshal(v interface{}) json.RawMessage {
	data, _ := json.Marshal(v)
	return data
}
