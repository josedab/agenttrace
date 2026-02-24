package domain

import (
	"time"

	"github.com/google/uuid"
)

// ReviewStatus represents the status of a trace review
type ReviewStatus string

const (
	ReviewStatusPending      ReviewStatus = "pending"
	ReviewStatusInReview     ReviewStatus = "in_review"
	ReviewStatusApproved     ReviewStatus = "approved"
	ReviewStatusRejected     ReviewStatus = "rejected"
	ReviewStatusNeedsChanges ReviewStatus = "needs_changes"
)

// IsValid checks if the review status is valid
func (s ReviewStatus) IsValid() bool {
	switch s {
	case ReviewStatusPending, ReviewStatusInReview, ReviewStatusApproved, ReviewStatusRejected, ReviewStatusNeedsChanges:
		return true
	}
	return false
}

// ActivityType represents the type of activity in the collaboration feed
type ActivityType string

const (
	ActivityTypeTraceReviewed    ActivityType = "trace_reviewed"
	ActivityTypePromptDeployed   ActivityType = "prompt_deployed"
	ActivityTypeEvalCompleted    ActivityType = "eval_completed"
	ActivityTypeAnnotationAdded  ActivityType = "annotation_added"
	ActivityTypeScoreSubmitted   ActivityType = "score_submitted"
)

// IsValid checks if the activity type is valid
func (t ActivityType) IsValid() bool {
	switch t {
	case ActivityTypeTraceReviewed, ActivityTypePromptDeployed, ActivityTypeEvalCompleted, ActivityTypeAnnotationAdded, ActivityTypeScoreSubmitted:
		return true
	}
	return false
}

// ReviewQueue represents a queue of traces awaiting review
type ReviewQueue struct {
	ID             uuid.UUID              `json:"id"`
	ProjectID      uuid.UUID              `json:"projectId"`
	Name           string                 `json:"name"`
	Description    string                 `json:"description,omitempty"`
	Filters        map[string]interface{} `json:"filters,omitempty"`
	AssignedTo     []uuid.UUID            `json:"assignedTo,omitempty"`
	PendingCount   int                    `json:"pendingCount"`
	CompletedCount int                    `json:"completedCount"`
	CreatedAt      time.Time              `json:"createdAt"`
}

// ReviewAssignment represents a review assignment for a specific trace
type ReviewAssignment struct {
	ID          uuid.UUID    `json:"id"`
	QueueID     uuid.UUID    `json:"queueId"`
	TraceID     uuid.UUID    `json:"traceId"`
	AssignedTo  uuid.UUID    `json:"assignedTo"`
	Status      ReviewStatus `json:"status"`
	Feedback    string       `json:"feedback,omitempty"`
	Score       *float64     `json:"score,omitempty"`
	AssignedAt  time.Time    `json:"assignedAt"`
	CompletedAt *time.Time   `json:"completedAt,omitempty"`
}

// QualityStandard represents a set of quality rules enforced on a project
type QualityStandard struct {
	ID             uuid.UUID     `json:"id"`
	ProjectID      uuid.UUID     `json:"projectId"`
	Name           string        `json:"name"`
	Enabled        bool          `json:"enabled"`
	Rules          []QualityRule `json:"rules"`
	EnforceOnDeploy bool         `json:"enforceOnDeploy"`
	CreatedAt      time.Time     `json:"createdAt"`
}

// QualityRule represents a single quality rule within a standard
type QualityRule struct {
	Metric      string  `json:"metric"`
	Operator    string  `json:"operator"`
	Threshold   float64 `json:"threshold"`
	Description string  `json:"description,omitempty"`
}

// ActivityFeedItem represents a single item in the team activity feed
type ActivityFeedItem struct {
	ID           uuid.UUID              `json:"id"`
	ProjectID    uuid.UUID              `json:"projectId"`
	Type         ActivityType           `json:"type"`
	UserID       uuid.UUID              `json:"userId"`
	UserName     string                 `json:"userName"`
	Description  string                 `json:"description"`
	ResourceID   string                 `json:"resourceId"`
	ResourceType string                 `json:"resourceType"`
	Timestamp    time.Time              `json:"timestamp"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}
