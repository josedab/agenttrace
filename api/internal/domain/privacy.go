package domain

import (
	"time"

	"github.com/google/uuid"
)

// PIIType represents the type of PII detected
type PIIType string

const (
	PIISSN        PIIType = "ssn"
	PIICreditCard PIIType = "credit_card"
	PIIEmail      PIIType = "email"
	PIIPhone      PIIType = "phone"
	PIIName       PIIType = "name"
	PIIAddress    PIIType = "address"
)

// SensitivityLevel represents the sensitivity level for PII detection
type SensitivityLevel string

const (
	SensitivityLow    SensitivityLevel = "low"
	SensitivityMedium SensitivityLevel = "medium"
	SensitivityHigh   SensitivityLevel = "high"
)

// DataResidency represents the data residency region
type DataResidency string

const (
	ResidencyUS     DataResidency = "us"
	ResidencyEU     DataResidency = "eu"
	ResidencyAPAC   DataResidency = "apac"
	ResidencyGlobal DataResidency = "global"
)

// DeletionStatus represents the status of a data deletion request
type DeletionStatus string

const (
	DeletionPending    DeletionStatus = "pending"
	DeletionProcessing DeletionStatus = "processing"
	DeletionCompleted  DeletionStatus = "completed"
)

// DeletionRequestType represents the type of data deletion request
type DeletionRequestType string

const (
	DeletionTypeTrace DeletionRequestType = "trace"
	DeletionTypeUser  DeletionRequestType = "user"
	DeletionTypeAll   DeletionRequestType = "all"
)

// PIIDetectionResult represents the result of a PII scan
type PIIDetectionResult struct {
	TotalScanned    int          `json:"totalScanned"`
	PIIFound        int          `json:"piiFound"`
	Findings        []PIIFinding `json:"findings"`
	RedactionApplied bool        `json:"redactionApplied"`
}

// PIIFinding represents a single PII finding
type PIIFinding struct {
	Type       PIIType `json:"type"`
	Value      string  `json:"value"`
	Location   string  `json:"location"`
	Confidence float64 `json:"confidence"`
	Redacted   bool    `json:"redacted"`
}

// PIIConfig represents the PII detection configuration
type PIIConfig struct {
	ID               uuid.UUID        `json:"id"`
	ProjectID        uuid.UUID        `json:"projectId"`
	Enabled          bool             `json:"enabled"`
	SensitivityLevel SensitivityLevel `json:"sensitivityLevel"`
	AutoRedact       bool             `json:"autoRedact"`
	DataResidency    DataResidency    `json:"dataResidency"`
	RetentionDays    int              `json:"retentionDays"`
	CreatedAt        time.Time        `json:"createdAt"`
}

// PIIConfigInput represents input for updating PII configuration
type PIIConfigInput struct {
	Enabled          *bool             `json:"enabled,omitempty"`
	SensitivityLevel *SensitivityLevel `json:"sensitivityLevel,omitempty"`
	AutoRedact       *bool             `json:"autoRedact,omitempty"`
	DataResidency    *DataResidency    `json:"dataResidency,omitempty"`
	RetentionDays    *int              `json:"retentionDays,omitempty"`
}

// DataDeletionRequest represents a data deletion request
type DataDeletionRequest struct {
	ID          uuid.UUID           `json:"id"`
	ProjectID   uuid.UUID           `json:"projectId"`
	RequestType DeletionRequestType `json:"requestType"`
	TargetID    string              `json:"targetId"`
	Status      DeletionStatus      `json:"status"`
	RequestedAt time.Time           `json:"requestedAt"`
	CompletedAt *time.Time          `json:"completedAt,omitempty"`
}

// DeletionRequestInput represents input for creating a deletion request
type DeletionRequestInput struct {
	RequestType DeletionRequestType `json:"requestType" validate:"required"`
	TargetID    string              `json:"targetId" validate:"required"`
}
