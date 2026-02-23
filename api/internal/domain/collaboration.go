package domain

import (
	"time"

	"github.com/google/uuid"
)

// CollaborationEventType represents the type of collaboration event
type CollaborationEventType string

const (
	CollaborationEventCursorMove        CollaborationEventType = "cursor_move"
	CollaborationEventAnnotationAdd     CollaborationEventType = "annotation_add"
	CollaborationEventAnnotationResolve CollaborationEventType = "annotation_resolve"
	CollaborationEventUserJoin          CollaborationEventType = "user_join"
	CollaborationEventUserLeave         CollaborationEventType = "user_leave"
)

// IsValid checks if the collaboration event type is valid
func (t CollaborationEventType) IsValid() bool {
	switch t {
	case CollaborationEventCursorMove, CollaborationEventAnnotationAdd, CollaborationEventAnnotationResolve, CollaborationEventUserJoin, CollaborationEventUserLeave:
		return true
	}
	return false
}

// UserPresence represents a user's presence on a trace
type UserPresence struct {
	UserID         uuid.UUID       `json:"userId"`
	UserName       string          `json:"userName"`
	TraceID        string          `json:"traceId"`
	ProjectID      uuid.UUID       `json:"projectId"`
	CursorPosition *CursorPosition `json:"cursorPosition,omitempty"`
	LastSeen       time.Time       `json:"lastSeen"`
}

// CursorPosition represents a user's cursor position within a trace
type CursorPosition struct {
	EventID      string `json:"eventId"`
	ScrollOffset int    `json:"scrollOffset"`
}

// TraceAnnotation represents an annotation on a trace event
type TraceAnnotation struct {
	ID         uuid.UUID  `json:"id"`
	ProjectID  uuid.UUID  `json:"projectId"`
	TraceID    string     `json:"traceId"`
	EventID    string     `json:"eventId"`
	UserID     uuid.UUID  `json:"userId"`
	UserName   string     `json:"userName"`
	Content    string     `json:"content"`
	ResolvedAt *time.Time `json:"resolvedAt,omitempty"`
	CreatedAt  time.Time  `json:"createdAt"`
}

// SharedSession represents a shared collaboration session on a trace
type SharedSession struct {
	ID           uuid.UUID   `json:"id"`
	ProjectID    uuid.UUID   `json:"projectId"`
	TraceID      string      `json:"traceId"`
	CreatedBy    uuid.UUID   `json:"createdBy"`
	Participants []uuid.UUID `json:"participants"`
	ExpiresAt    time.Time   `json:"expiresAt"`
	CreatedAt    time.Time   `json:"createdAt"`
}

// CollaborationEvent represents a real-time collaboration event
type CollaborationEvent struct {
	Type      CollaborationEventType `json:"type"`
	UserID    uuid.UUID              `json:"userId"`
	Payload   any                    `json:"payload"`
	Timestamp time.Time              `json:"timestamp"`
}

// DiscussionThread represents a threaded discussion on a trace observation
type DiscussionThread struct {
	ID             uuid.UUID       `json:"id"`
	TraceID        uuid.UUID       `json:"traceId"`
	ObservationID  *uuid.UUID      `json:"observationId,omitempty"`
	Title          string          `json:"title"`
	Status         string          `json:"status"` // open, resolved, archived
	CreatedBy      uuid.UUID       `json:"createdBy"`
	CreatedByName  string          `json:"createdByName"`
	Messages       []ThreadMessage `json:"messages"`
	ParticipantIDs []uuid.UUID     `json:"participantIds"`
	Tags           []string        `json:"tags,omitempty"`
	CreatedAt      time.Time       `json:"createdAt"`
	UpdatedAt      time.Time       `json:"updatedAt"`
}

// ThreadMessage represents a message in a discussion thread
type ThreadMessage struct {
	ID        uuid.UUID   `json:"id"`
	ThreadID  uuid.UUID   `json:"threadId"`
	UserID    uuid.UUID   `json:"userId"`
	UserName  string      `json:"userName"`
	Content   string      `json:"content"`
	Mentions  []uuid.UUID `json:"mentions,omitempty"`
	CreatedAt time.Time   `json:"createdAt"`
	EditedAt  *time.Time  `json:"editedAt,omitempty"`
}

// EvalQueue represents a shared evaluation queue for team collaboration
type EvalQueue struct {
	ID          uuid.UUID         `json:"id"`
	ProjectID   uuid.UUID         `json:"projectId"`
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	Assignees   []uuid.UUID       `json:"assignees"`
	TraceIDs    []uuid.UUID       `json:"traceIds"`
	Status      string            `json:"status"` // active, completed, paused
	Progress    EvalQueueProgress `json:"progress"`
	CreatedAt   time.Time         `json:"createdAt"`
	UpdatedAt   time.Time         `json:"updatedAt"`
}

// EvalQueueProgress tracks progress of an evaluation queue
type EvalQueueProgress struct {
	Total      int `json:"total"`
	Completed  int `json:"completed"`
	InProgress int `json:"inProgress"`
	Pending    int `json:"pending"`
}

// DiscussionInput for creating threads
type DiscussionInput struct {
	TraceID        uuid.UUID  `json:"traceId" validate:"required"`
	ObservationID  *uuid.UUID `json:"observationId,omitempty"`
	Title          string     `json:"title" validate:"required"`
	InitialMessage string     `json:"initialMessage" validate:"required"`
}

// EvalQueueInput for creating evaluation queues
type EvalQueueInput struct {
	Name        string      `json:"name" validate:"required"`
	Description string      `json:"description,omitempty"`
	Assignees   []uuid.UUID `json:"assignees"`
	TraceIDs    []uuid.UUID `json:"traceIds" validate:"required,min=1"`
}
