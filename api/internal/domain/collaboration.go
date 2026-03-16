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

// TraceReviewRequest represents a formal review request for a trace with approval workflow
type TraceReviewRequest struct {
	ID                uuid.UUID              `json:"id"`
	ProjectID         uuid.UUID              `json:"projectId"`
	TraceID           string                 `json:"traceId"`
	Title             string                 `json:"title"`
	Description       string                 `json:"description,omitempty"`
	RequestedBy       uuid.UUID              `json:"requestedBy"`
	AssignedTo        []uuid.UUID            `json:"assignedTo"`
	Status            ReviewStatus           `json:"status"`
	Priority          string                 `json:"priority"` // low, medium, high, urgent
	Labels            []string               `json:"labels,omitempty"`
	DueAt             *time.Time             `json:"dueAt,omitempty"`
	Comments          []CollabReviewComment  `json:"comments,omitempty"`
	ApprovalCount     int                    `json:"approvalCount"`
	RequiredApprovals int                    `json:"requiredApprovals"`
	CreatedAt         time.Time              `json:"createdAt"`
	UpdatedAt         time.Time              `json:"updatedAt"`
}

// CollabReviewComment represents a threaded comment on a collaborative review
type CollabReviewComment struct {
	ID         uuid.UUID      `json:"id"`
	ReviewID   uuid.UUID      `json:"reviewId"`
	ParentID   *uuid.UUID     `json:"parentId,omitempty"` // for threading
	AuthorID   uuid.UUID      `json:"authorId"`
	AuthorName string         `json:"authorName"`
	Content    string         `json:"content"`
	Mentions   []string       `json:"mentions,omitempty"` // @user references
	SpanID     string         `json:"spanId,omitempty"`   // link to specific span
	Resolved   bool           `json:"resolved"`
	Reactions  map[string]int `json:"reactions,omitempty"` // emoji -> count
	CreatedAt  time.Time      `json:"createdAt"`
	UpdatedAt  time.Time      `json:"updatedAt"`
}

// CollabReviewQueue configures how reviews are assigned and tracked with SLA
type CollabReviewQueue struct {
	ID              uuid.UUID   `json:"id"`
	ProjectID       uuid.UUID   `json:"projectId"`
	Name            string      `json:"name"`
	AssignmentRule  string      `json:"assignmentRule"` // round_robin, load_balanced, manual, random
	Reviewers       []uuid.UUID `json:"reviewers"`
	AutoAssign      bool        `json:"autoAssign"`
	SLAHours        int         `json:"slaHours"`
	EscalationHours int         `json:"escalationHours"`
	PendingCount    int         `json:"pendingCount"`
	AvgReviewTimeH  float64     `json:"avgReviewTimeHours"`
	CreatedAt       time.Time   `json:"createdAt"`
}

// NotificationIntegration represents a Slack/Teams/GitHub integration
type NotificationIntegration struct {
	ID         uuid.UUID `json:"id"`
	ProjectID  uuid.UUID `json:"projectId"`
	Type       string    `json:"type"` // slack, teams, github, email
	Name       string    `json:"name"`
	WebhookURL string    `json:"webhookUrl,omitempty"`
	ChannelID  string    `json:"channelId,omitempty"`
	Enabled    bool      `json:"enabled"`
	Events     []string  `json:"events"` // review_created, comment_added, approved, mentioned
	CreatedAt  time.Time `json:"createdAt"`
}

// CollabTraceReviewInput represents input for creating a collaborative review request
type CollabTraceReviewInput struct {
	TraceID           string      `json:"traceId" validate:"required"`
	Title             string      `json:"title" validate:"required,min=1,max=200"`
	Description       string      `json:"description,omitempty"`
	AssignedTo        []uuid.UUID `json:"assignedTo,omitempty"`
	Priority          string      `json:"priority,omitempty"`
	Labels            []string    `json:"labels,omitempty"`
	RequiredApprovals int         `json:"requiredApprovals,omitempty"`
}

// CollabReviewCommentInput represents input for creating a threaded review comment
type CollabReviewCommentInput struct {
	Content  string     `json:"content" validate:"required,min=1,max=5000"`
	ParentID *uuid.UUID `json:"parentId,omitempty"`
	SpanID   string     `json:"spanId,omitempty"`
	Mentions []string   `json:"mentions,omitempty"`
}

// CollabReviewQueueInput represents input for creating a review queue with assignment rules
type CollabReviewQueueInput struct {
	Name            string      `json:"name" validate:"required"`
	AssignmentRule  string      `json:"assignmentRule,omitempty"`
	Reviewers       []uuid.UUID `json:"reviewers,omitempty"`
	AutoAssign      bool        `json:"autoAssign,omitempty"`
	SLAHours        int         `json:"slaHours,omitempty"`
	EscalationHours int         `json:"escalationHours,omitempty"`
}

// NotificationIntegrationInput represents input for adding an integration
type NotificationIntegrationInput struct {
	Type       string   `json:"type" validate:"required"`
	Name       string   `json:"name" validate:"required"`
	WebhookURL string   `json:"webhookUrl,omitempty"`
	ChannelID  string   `json:"channelId,omitempty"`
	Events     []string `json:"events,omitempty"`
}
