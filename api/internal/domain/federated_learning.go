package domain

import (
	"time"

	"github.com/google/uuid"
)

// FederationRing represents a federated learning ring
type FederationRing struct {
	ID                 uuid.UUID          `json:"id"`
	Name               string             `json:"name"`
	Participants       int                `json:"participants"`
	Status             string             `json:"status"` // active, paused
	PrivacyLevel       string             `json:"privacyLevel"` // strict, moderate, relaxed
	AggregatedInsights []FederatedInsight `json:"aggregatedInsights"`
	CreatedAt          time.Time          `json:"createdAt"`
}

// FederatedInsight represents an insight derived from federated learning
type FederatedInsight struct {
	ID               uuid.UUID `json:"id"`
	RingID           uuid.UUID `json:"ringId"`
	Category         string    `json:"category"` // prompt_optimization, model_selection, cost_reduction, error_pattern
	Insight          string    `json:"insight"`
	Confidence       float64   `json:"confidence"`
	ContributingOrgs int       `json:"contributingOrgs"`
	GeneratedAt      time.Time `json:"generatedAt"`
}

// FederationJoinInput represents input for joining a federation ring
type FederationJoinInput struct {
	RingName       string   `json:"ringName"`
	DataCategories []string `json:"dataCategories"`
	PrivacyLevel   string   `json:"privacyLevel"`
}

// FederationConfig represents the federation configuration for a project
type FederationConfig struct {
	ProjectID                  uuid.UUID   `json:"projectId"`
	Enabled                    bool        `json:"enabled"`
	ParticipatingRings         []uuid.UUID `json:"participatingRings"`
	SharingCategories          []string    `json:"sharingCategories"`
	DifferentialPrivacyEpsilon float64     `json:"differentialPrivacyEpsilon"`
}
