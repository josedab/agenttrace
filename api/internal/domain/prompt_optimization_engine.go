package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// PromptTuningStrategy represents the algorithm used for automated prompt tuning.
type PromptTuningStrategy string

const (
	TuningStrategyDSPyBootstrap    PromptTuningStrategy = "dspy_bootstrap"
	TuningStrategyDSPyMIPRO        PromptTuningStrategy = "dspy_mipro"
	TuningStrategyGradientDescent  PromptTuningStrategy = "gradient_descent"
	TuningStrategyRandomSearch     PromptTuningStrategy = "random_search"
	TuningStrategyBayesian         PromptTuningStrategy = "bayesian"
)

// IsValid checks if the tuning strategy is valid.
func (s PromptTuningStrategy) IsValid() bool {
	switch s {
	case TuningStrategyDSPyBootstrap, TuningStrategyDSPyMIPRO, TuningStrategyGradientDescent, TuningStrategyRandomSearch, TuningStrategyBayesian:
		return true
	}
	return false
}

// ABTestStatus represents the lifecycle status of an A/B test.
type ABTestStatus string

const (
	ABTestStatusDraft     ABTestStatus = "draft"
	ABTestStatusRunning   ABTestStatus = "running"
	ABTestStatusPaused    ABTestStatus = "paused"
	ABTestStatusCompleted ABTestStatus = "completed"
	ABTestStatusCancelled ABTestStatus = "cancelled"
)

// IsValid checks if the A/B test status is valid.
func (s ABTestStatus) IsValid() bool {
	switch s {
	case ABTestStatusDraft, ABTestStatusRunning, ABTestStatusPaused, ABTestStatusCompleted, ABTestStatusCancelled:
		return true
	}
	return false
}

// PromptTuningJob represents an automated prompt tuning run.
type PromptTuningJob struct {
	ID                  uuid.UUID            `json:"id"`
	ProjectID           uuid.UUID            `json:"projectId"`
	PromptID            uuid.UUID            `json:"promptId"`
	Strategy            PromptTuningStrategy `json:"strategy"`
	ObjectiveMetric     string               `json:"objectiveMetric"`
	ConstraintMetrics   map[string]float64   `json:"constraintMetrics,omitempty"`
	BaselinePrompt      string               `json:"baselinePrompt"`
	DatasetID           *uuid.UUID           `json:"datasetId,omitempty"`
	Iterations          int                  `json:"iterations"`
	CompletedIterations int                  `json:"completedIterations"`
	BestScore           float64              `json:"bestScore"`
	BaselineScore       float64              `json:"baselineScore"`
	ImprovementPct      float64              `json:"improvementPct"`
	GeneratedVariants   []TuningVariant      `json:"generatedVariants,omitempty"`
	Status              OptimizationStatus   `json:"status"`
	CreatedAt           time.Time            `json:"createdAt"`
	CompletedAt         *time.Time           `json:"completedAt,omitempty"`
}

// TuningVariant represents a single prompt variant generated during a tuning job.
type TuningVariant struct {
	ID             uuid.UUID            `json:"id"`
	Iteration      int                  `json:"iteration"`
	PromptText     string               `json:"promptText"`
	Score          float64              `json:"score"`
	Metrics        map[string]float64   `json:"metrics,omitempty"`
	Rationale      string               `json:"rationale"`
	Demonstrations []TuningDemonstration `json:"demonstrations,omitempty"`
	IsSelected     bool                 `json:"isSelected"`
}

// TuningDemonstration represents a few-shot example used in DSPy bootstrap tuning.
type TuningDemonstration struct {
	Input       string `json:"input"`
	Output      string `json:"output"`
	Explanation string `json:"explanation"`
}

// ABTest represents an A/B test comparing prompt variants.
type ABTest struct {
	ID               uuid.UUID          `json:"id"`
	ProjectID        uuid.UUID          `json:"projectId"`
	PromptID         uuid.UUID          `json:"promptId"`
	Name             string             `json:"name"`
	Status           ABTestStatus       `json:"status"`
	ControlPrompt    ABTestArm          `json:"controlPrompt"`
	TreatmentPrompts []ABTestArm        `json:"treatmentPrompts,omitempty"`
	TrafficSplit     map[string]float64 `json:"trafficSplit,omitempty"`
	MinSampleSize    int                `json:"minSampleSize"`
	CurrentSampleSize int               `json:"currentSampleSize"`
	ConfidenceLevel  float64            `json:"confidenceLevel"`
	PrimaryMetric    string             `json:"primaryMetric"`
	Results          *ABTestResults     `json:"results,omitempty"`
	CreatedAt        time.Time          `json:"createdAt"`
	UpdatedAt        time.Time          `json:"updatedAt"`
	CompletedAt      *time.Time         `json:"completedAt,omitempty"`
}

// ABTestArm represents a single arm (control or treatment) in an A/B test.
type ABTestArm struct {
	ID             uuid.UUID          `json:"id"`
	Name           string             `json:"name"`
	PromptText     string             `json:"promptText"`
	SampleSize     int                `json:"sampleSize"`
	Metrics        map[string]float64 `json:"metrics,omitempty"`
	ConversionRate float64            `json:"conversionRate"`
}

// ABTestResults holds the statistical results of a completed A/B test.
type ABTestResults struct {
	WinnerID                *uuid.UUID `json:"winnerId,omitempty"`
	WinnerName              string     `json:"winnerName"`
	StatisticallySignificant bool      `json:"statisticallySignificant"`
	PValue                  float64    `json:"pValue"`
	ConfidenceInterval      []float64  `json:"confidenceInterval,omitempty"`
	LiftPct                 float64    `json:"liftPct"`
	Recommendation          string     `json:"recommendation"`
}

// PromptTuningInput represents the input for creating a new prompt tuning job.
type PromptTuningInput struct {
	PromptID          uuid.UUID            `json:"promptId" validate:"required"`
	Strategy          PromptTuningStrategy `json:"strategy" validate:"required"`
	ObjectiveMetric   string               `json:"objectiveMetric" validate:"required"`
	ConstraintMetrics map[string]float64   `json:"constraintMetrics,omitempty"`
	BaselinePrompt    string               `json:"baselinePrompt" validate:"required"`
	Iterations        int                  `json:"iterations,omitempty"`
	DatasetID         *uuid.UUID           `json:"datasetId,omitempty"`
}

// Validate validates the PromptTuningInput fields.
func (i *PromptTuningInput) Validate() error {
	if i.PromptID == uuid.Nil {
		return fmt.Errorf("promptId is required")
	}
	if !i.Strategy.IsValid() {
		return fmt.Errorf("invalid tuning strategy: %s", i.Strategy)
	}
	if i.ObjectiveMetric == "" {
		return fmt.Errorf("objectiveMetric is required")
	}
	if i.BaselinePrompt == "" {
		return fmt.Errorf("baselinePrompt is required")
	}
	if i.Iterations <= 0 {
		i.Iterations = 10
	}
	return nil
}

// ABTestInput represents the input for creating a new A/B test.
type ABTestInput struct {
	PromptID             uuid.UUID          `json:"promptId" validate:"required"`
	Name                 string             `json:"name" validate:"required,min=1,max=200"`
	ControlPromptText    string             `json:"controlPromptText" validate:"required"`
	TreatmentPromptTexts []string           `json:"treatmentPromptTexts" validate:"required,min=1"`
	TrafficSplit         map[string]float64 `json:"trafficSplit,omitempty"`
	MinSampleSize        int                `json:"minSampleSize,omitempty"`
	ConfidenceLevel      float64            `json:"confidenceLevel,omitempty"`
	PrimaryMetric        string             `json:"primaryMetric" validate:"required"`
}

// Validate validates the ABTestInput fields.
func (i *ABTestInput) Validate() error {
	if i.PromptID == uuid.Nil {
		return fmt.Errorf("promptId is required")
	}
	if i.Name == "" || len(i.Name) > 200 {
		return fmt.Errorf("name is required and must be between 1 and 200 characters")
	}
	if i.ControlPromptText == "" {
		return fmt.Errorf("controlPromptText is required")
	}
	if len(i.TreatmentPromptTexts) == 0 {
		return fmt.Errorf("at least one treatmentPromptText is required")
	}
	if i.PrimaryMetric == "" {
		return fmt.Errorf("primaryMetric is required")
	}
	return nil
}

// PromptOptimizationDashboard provides a summary view of tuning jobs and A/B tests.
type PromptOptimizationDashboard struct {
	TotalTuningJobs   int               `json:"totalTuningJobs"`
	ActiveJobs        int               `json:"activeJobs"`
	CompletedJobs     int               `json:"completedJobs"`
	TotalABTests      int               `json:"totalABTests"`
	ActiveABTests     int               `json:"activeABTests"`
	AvgImprovementPct float64           `json:"avgImprovementPct"`
	RecentJobs        []PromptTuningJob `json:"recentJobs,omitempty"`
	ActiveTests       []ABTest          `json:"activeTests,omitempty"`
	TopOptimizations  []PromptTuningJob `json:"topOptimizations,omitempty"`
}
