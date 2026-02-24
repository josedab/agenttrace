package domain

import (
	"time"
)

// CollabAnnotation represents a collaborative annotation on a trace
type CollabAnnotation struct {
	ID        string             `json:"id"`
	TraceID   string             `json:"traceId"`
	ProjectID string             `json:"projectId"`
	SpanID    string             `json:"spanId,omitempty"`
	UserID    string             `json:"userId"`
	UserName  string             `json:"userName"`
	Content   string             `json:"content"`
	Position  AnnotationPosition `json:"position"`
	Thread    []AnnotationReply  `json:"thread"`
	Resolved  bool               `json:"resolved"`
	CreatedAt time.Time          `json:"createdAt"`
}

// AnnotationPosition describes where in a trace the annotation is placed
type AnnotationPosition struct {
	SpanID     string `json:"spanId"`
	LineNumber *int   `json:"lineNumber,omitempty"`
	Timestamp  string `json:"timestamp,omitempty"`
}

// AnnotationReply represents a reply in an annotation thread
type AnnotationReply struct {
	ID        string    `json:"id"`
	UserID    string    `json:"userId"`
	UserName  string    `json:"userName"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"createdAt"`
}

// AnnotationPresence represents who is actively viewing a trace
type AnnotationPresence struct {
	TraceID     string         `json:"traceId"`
	ActiveUsers []PresenceUser `json:"activeUsers"`
}

// PresenceUser represents a user currently viewing a trace
type PresenceUser struct {
	UserID         string    `json:"userId"`
	UserName       string    `json:"userName"`
	CursorPosition string    `json:"cursorPosition,omitempty"`
	LastActive     time.Time `json:"lastActive"`
}

// AnnotationInput represents input for creating an annotation
type AnnotationInput struct {
	TraceID  string             `json:"traceId" validate:"required"`
	SpanID   string             `json:"spanId,omitempty"`
	Content  string             `json:"content" validate:"required"`
	Position AnnotationPosition `json:"position"`
}

// ReplyInput represents input for adding a reply to an annotation
type ReplyInput struct {
	Content string `json:"content" validate:"required"`
}
