package domain

import (
	"time"

	"github.com/google/uuid"
)

// OnboardingStep represents a step in the cloud onboarding process
type OnboardingStep string

const (
	OnboardingStepAccountCreated OnboardingStep = "account_created"
	OnboardingStepProjectCreated OnboardingStep = "project_created"
	OnboardingStepSDKInstalled   OnboardingStep = "sdk_installed"
	OnboardingStepFirstTrace     OnboardingStep = "first_trace"
	OnboardingStepFirstEval      OnboardingStep = "first_eval"
)

// IsValid checks if the onboarding step is valid
func (s OnboardingStep) IsValid() bool {
	switch s {
	case OnboardingStepAccountCreated, OnboardingStepProjectCreated, OnboardingStepSDKInstalled, OnboardingStepFirstTrace, OnboardingStepFirstEval:
		return true
	}
	return false
}

// PlanTier represents the subscription plan tier
type PlanTier string

const (
	PlanTierFree       PlanTier = "free"
	PlanTierTeam       PlanTier = "team"
	PlanTierEnterprise PlanTier = "enterprise"
)

// IsValid checks if the plan tier is valid
func (p PlanTier) IsValid() bool {
	switch p {
	case PlanTierFree, PlanTierTeam, PlanTierEnterprise:
		return true
	}
	return false
}

// CloudOnboarding represents the onboarding state for a cloud tenant
type CloudOnboarding struct {
	ID          uuid.UUID              `json:"id"`
	TenantID    uuid.UUID              `json:"tenantId"`
	Steps       []OnboardingStepStatus `json:"steps"`
	CurrentStep OnboardingStep         `json:"currentStep"`
	SDKLanguage string                 `json:"sdkLanguage,omitempty"`
	Framework   string                 `json:"framework,omitempty"`
	CompletedAt *time.Time             `json:"completedAt,omitempty"`
	CreatedAt   time.Time              `json:"createdAt"`
}

// OnboardingStepStatus represents the completion status of an onboarding step
type OnboardingStepStatus struct {
	Step        OnboardingStep `json:"step"`
	Completed   bool           `json:"completed"`
	CompletedAt *time.Time     `json:"completedAt,omitempty"`
}

// QuickstartConfig represents the generated quickstart configuration for a new project
type QuickstartConfig struct {
	Language  string `json:"language"`
	Framework string `json:"framework"`
	APIKey    string `json:"apiKey"`
	ProjectID string `json:"projectId"`
	Host      string `json:"host"`
}

// UsageMeter represents the current usage metrics for a tenant
type UsageMeter struct {
	TenantID          uuid.UUID `json:"tenantId"`
	Period            string    `json:"period"`
	TracesUsed        int64     `json:"tracesUsed"`
	TracesLimit       int64     `json:"tracesLimit"`
	StorageUsedBytes  int64     `json:"storageUsedBytes"`
	StorageLimitBytes int64     `json:"storageLimitBytes"`
	APICallsUsed      int64     `json:"apiCallsUsed"`
	APICallsLimit     int64     `json:"apiCallsLimit"`
	PercentUsed       float64   `json:"percentUsed"`
}
