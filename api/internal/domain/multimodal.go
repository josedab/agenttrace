package domain

import (
	"time"

	"github.com/google/uuid"
)

// TraceAttachment represents a multi-modal attachment to a trace
type TraceAttachment struct {
	ID            uuid.UUID         `json:"id"`
	TraceID       uuid.UUID         `json:"traceId"`
	ObservationID uuid.UUID         `json:"observationId"`
	Type          string            `json:"type"` // image, screenshot, audio, video, document
	Filename      string            `json:"filename"`
	MimeType      string            `json:"mimeType"`
	SizeBytes     int64             `json:"sizeBytes"`
	URL           string            `json:"url"`
	Metadata      map[string]string `json:"metadata"`
	UploadedAt    time.Time         `json:"uploadedAt"`
}

// MultiModalTrace represents a trace with its multi-modal attachments
type MultiModalTrace struct {
	TraceID     uuid.UUID         `json:"traceId"`
	Attachments []TraceAttachment `json:"attachments"`
	Summary     MultiModalSummary `json:"summary"`
}

// MultiModalSummary provides summary statistics for multi-modal attachments
type MultiModalSummary struct {
	TotalAttachments int            `json:"totalAttachments"`
	ByType           map[string]int `json:"byType"`
	TotalSizeBytes   int64          `json:"totalSizeBytes"`
}

// AttachmentInput represents input for registering a new attachment
type AttachmentInput struct {
	TraceID       uuid.UUID `json:"traceId"`
	ObservationID uuid.UUID `json:"observationId"`
	Type          string    `json:"type"`
	Filename      string    `json:"filename"`
	MimeType      string    `json:"mimeType"`
	SizeBytes     int64     `json:"sizeBytes"`
}

// AttachmentFilter represents filters for listing attachments
type AttachmentFilter struct {
	TraceID string `json:"traceId"`
	Type    string `json:"type"`
	MinSize int64  `json:"minSize"`
	MaxSize int64  `json:"maxSize"`
}
