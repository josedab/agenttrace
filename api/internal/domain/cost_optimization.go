package domain

import (
	"time"

	"github.com/google/uuid"
)

// CostRecommendationStatus represents the status of a cost recommendation
type CostRecommendationStatus string

const (
	CostRecommendationPending   CostRecommendationStatus = "PENDING"
	CostRecommendationApplied   CostRecommendationStatus = "APPLIED"
	CostRecommendationDismissed CostRecommendationStatus = "DISMISSED"
)

// CostRecommendation represents a suggestion to switch to a cheaper model
type CostRecommendation struct {
	ID                      uuid.UUID                `json:"id"`
	ProjectID               uuid.UUID                `json:"projectId"`
	CurrentModel            string                   `json:"currentModel"`
	RecommendedModel        string                   `json:"recommendedModel"`
	TraceCount              int                      `json:"traceCount"`
	EstimatedSavingsPerMonth float64                 `json:"estimatedSavingsPerMonth"`
	QualityImpactEstimate   float64                  `json:"qualityImpactEstimate"`
	Confidence              float64                  `json:"confidence"`
	Status                  CostRecommendationStatus `json:"status"`
	CreatedAt               time.Time                `json:"createdAt"`
}

// CostAnalysis represents a complete cost analysis for a project
type CostAnalysis struct {
	ProjectID         uuid.UUID            `json:"projectId"`
	TotalCostPeriod   float64              `json:"totalCostPeriod"`
	ModelBreakdown    []ModelCostEntry     `json:"modelBreakdown"`
	Recommendations   []CostRecommendation `json:"recommendations"`
	PotentialSavings  float64              `json:"potentialSavings"`
}

// ModelCostEntry represents per-model cost statistics
type ModelCostEntry struct {
	Model           string  `json:"model"`
	TraceCount      int     `json:"traceCount"`
	TotalCost       float64 `json:"totalCost"`
	AvgCostPerTrace float64 `json:"avgCostPerTrace"`
}

// DateRange represents a time range for analysis queries
type DateRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// CostAutopilotConfig represents autopilot configuration
type CostAutopilotConfig struct {
	ID               uuid.UUID `json:"id"`
	ProjectID        uuid.UUID `json:"projectId"`
	Enabled          bool      `json:"enabled"`
	MaxBudgetDaily   float64   `json:"maxBudgetDaily"`
	MaxBudgetMonthly float64   `json:"maxBudgetMonthly"`
	OptimizationLevel string   `json:"optimizationLevel"` // conservative, balanced, aggressive
	AutoApply        bool      `json:"autoApply"`
	NotifyOnSave     bool      `json:"notifyOnSave"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
}

// CostForecast represents a cost projection
type CostForecast struct {
	ProjectID             uuid.UUID         `json:"projectId"`
	CurrentDailyCost      float64           `json:"currentDailyCost"`
	ProjectedDaily        float64           `json:"projectedDailyCost"`
	ProjectedMonthly      float64           `json:"projectedMonthlyCost"`
	ProjectedYearly       float64           `json:"projectedYearlyCost"`
	ConfidenceInterval    [2]float64        `json:"confidenceInterval"` // 95% CI
	DailyProjections      []DailyProjection `json:"dailyProjections"`
	OptimizationPotential float64           `json:"optimizationPotential"` // percent savings possible
	BudgetStatus          string            `json:"budgetStatus"`         // within, warning, exceeded
}

// DailyProjection represents a single day forecast
type DailyProjection struct {
	Date      time.Time `json:"date"`
	Projected float64   `json:"projected"`
	Low       float64   `json:"low"`
	High      float64   `json:"high"`
}

// PromptOptimization represents a prompt compression suggestion
type PromptOptimization struct {
	ID               uuid.UUID `json:"id"`
	TraceID          uuid.UUID `json:"traceId"`
	ObservationID    uuid.UUID `json:"observationId"`
	OriginalTokens   int       `json:"originalTokens"`
	OptimizedTokens  int       `json:"optimizedTokens"`
	SavingsPercent   float64   `json:"savingsPercent"`
	OriginalPrompt   string    `json:"originalPrompt,omitempty"`
	OptimizedPrompt  string    `json:"optimizedPrompt,omitempty"`
	Technique        string    `json:"technique"` // compression, deduplication, caching, model_routing
	EstimatedSavings float64   `json:"estimatedMonthlySavings"`
	QualityImpact    string    `json:"qualityImpact"` // none, minimal, moderate
}

// CostReport represents a comprehensive cost report
type CostReport struct {
	ProjectID          uuid.UUID            `json:"projectId"`
	Period             DateRange            `json:"period"`
	TotalCost          float64              `json:"totalCost"`
	TotalTokens        int64                `json:"totalTokens"`
	TraceCount         int                  `json:"traceCount"`
	CostByModel        []ModelCostEntry     `json:"costByModel"`
	CostByDay          []DailyCostEntry     `json:"costByDay"`
	TopExpensiveTraces []TraceCostEntry     `json:"topExpensiveTraces"`
	Recommendations    []CostRecommendation `json:"recommendations"`
	Optimizations      []PromptOptimization `json:"optimizations"`
	Forecast           CostForecast         `json:"forecast"`
	ROI                ROICalculation       `json:"roi"`
}

// DailyCostEntry represents cost for a single day
type DailyCostEntry struct {
	Date   time.Time `json:"date"`
	Cost   float64   `json:"cost"`
	Tokens int64     `json:"tokens"`
	Traces int       `json:"traces"`
}

// TraceCostEntry represents cost for a single trace
type TraceCostEntry struct {
	TraceID   uuid.UUID `json:"traceId"`
	TraceName string    `json:"traceName"`
	Cost      float64   `json:"cost"`
	Tokens    int       `json:"tokens"`
	Model     string    `json:"model"`
}

// ROICalculation shows the return on investment from optimizations
type ROICalculation struct {
	TotalSavings         float64 `json:"totalSavings"`
	OptimizationsApplied int     `json:"optimizationsApplied"`
	CostBefore           float64 `json:"costBefore"`
	CostAfter            float64 `json:"costAfter"`
	SavingsPercent       float64 `json:"savingsPercent"`
}

// AutopilotConfigInput for updating autopilot settings
type AutopilotConfigInput struct {
	Enabled           *bool    `json:"enabled,omitempty"`
	MaxBudgetDaily    *float64 `json:"maxBudgetDaily,omitempty"`
	MaxBudgetMonthly  *float64 `json:"maxBudgetMonthly,omitempty"`
	OptimizationLevel string   `json:"optimizationLevel,omitempty"`
	AutoApply         *bool    `json:"autoApply,omitempty"`
}

// CostHotspot identifies a high-cost area in trace data
type CostHotspot struct {
	ID              uuid.UUID            `json:"id"`
	ProjectID       uuid.UUID            `json:"projectId"`
	Category        string               `json:"category"`        // model, prompt, tool, agent
	Name            string               `json:"name"`
	TotalCostUSD    float64              `json:"totalCostUsd"`
	TraceCount      int                  `json:"traceCount"`
	AvgCostPerTrace float64              `json:"avgCostPerTrace"`
	Trend           string               `json:"trend"`           // increasing, stable, decreasing
	TrendPct        float64              `json:"trendPercent"`
	TopModels       []ModelCostBreakdown `json:"topModels,omitempty"`
	PercentOfTotal  float64              `json:"percentOfTotal"`
	Severity        string               `json:"severity,omitempty"` // low, medium, high, critical
}

// ModelCostBreakdown shows cost per model
type ModelCostBreakdown struct {
	Model      string  `json:"model"`
	CostUSD    float64 `json:"costUsd"`
	TraceCount int     `json:"traceCount"`
	AvgTokens  int     `json:"avgTokens"`
	Percentage float64 `json:"percentage"`
}

// CachingStrategy represents a recommended caching optimization
type CachingStrategy struct {
	ID              uuid.UUID `json:"id"`
	Type            string    `json:"type"` // prompt_cache, semantic_cache, response_cache
	Description     string    `json:"description"`
	EstimatedSaving float64   `json:"estimatedMonthlySaving"`
	HitRateEstimate float64   `json:"hitRateEstimate"`
	Implementation  string    `json:"implementation"`
	Complexity      string    `json:"complexity"` // low, medium, high
}

// ModelRoutingSuggestion recommends using different models for different task types
type ModelRoutingSuggestion struct {
	TaskType        string  `json:"taskType"`
	CurrentModel    string  `json:"currentModel"`
	SuggestedModel  string  `json:"suggestedModel"`
	CostReduction   float64 `json:"costReductionPercent"`
	QualityImpact   float64 `json:"qualityImpactPercent"`
	Confidence      float64 `json:"confidence"`
	SampleSize      int     `json:"sampleSize"`
}

// CostAutopilotReport represents the autopilot's comprehensive analysis
type CostAutopilotReport struct {
	ProjectID          uuid.UUID                `json:"projectId"`
	Hotspots           []CostHotspot            `json:"hotspots"`
	CachingStrategies  []CachingStrategy        `json:"cachingStrategies"`
	ModelRouting       []ModelRoutingSuggestion  `json:"modelRouting"`
	BudgetAlerts       []BudgetAlert            `json:"budgetAlerts"`
	TotalSavingsPotential float64               `json:"totalSavingsPotential"`
	GeneratedAt        time.Time                `json:"generatedAt"`
}

// BudgetAlert represents a budget threshold alert
type BudgetAlert struct {
	ID          uuid.UUID `json:"id"`
	Type        string    `json:"type"` // warning, exceeded, projected_exceed
	Message     string    `json:"message"`
	CurrentCost float64   `json:"currentCost"`
	BudgetLimit float64   `json:"budgetLimit"`
	Percentage  float64   `json:"percentage"`
	CreatedAt   time.Time `json:"createdAt"`
}

// CostAutopilotRule represents an automation rule for cost optimization
type CostAutopilotRule struct {
	ID             uuid.UUID         `json:"id"`
	ProjectID      uuid.UUID         `json:"projectId"`
	Name           string            `json:"name"`
	RuleType       string            `json:"ruleType"` // model_downgrade, cache_enable, rate_limit, budget_alert
	Condition      CostRuleCondition `json:"condition"`
	Action         CostRuleAction    `json:"action"`
	Enabled        bool              `json:"enabled"`
	ExecutionCount int               `json:"executionCount"`
	LastExecuted   *time.Time        `json:"lastExecuted,omitempty"`
	CreatedAt      time.Time         `json:"createdAt"`
}

// CostRuleCondition defines when a rule triggers
type CostRuleCondition struct {
	Metric        string  `json:"metric"`   // daily_cost, per_trace_cost, token_usage, error_rate
	Operator      string  `json:"operator"` // gt, lt, gte, lte, eq
	Threshold     float64 `json:"threshold"`
	WindowMinutes int     `json:"windowMinutes"`
}

// CostRuleAction defines what happens when a rule triggers
type CostRuleAction struct {
	ActionType    string            `json:"actionType"` // switch_model, enable_cache, send_alert, throttle
	Parameters    map[string]string `json:"parameters"`
	FallbackModel string           `json:"fallbackModel,omitempty"`
}

// CostAutopilotRuleInput represents input for creating an autopilot rule
type CostAutopilotRuleInput struct {
	Name      string            `json:"name" validate:"required"`
	RuleType  string            `json:"ruleType" validate:"required"`
	Condition CostRuleCondition `json:"condition" validate:"required"`
	Action    CostRuleAction    `json:"action" validate:"required"`
}

// AutopilotCostPrediction represents a future cost prediction with confidence
type AutopilotCostPrediction struct {
	Date            time.Time `json:"date"`
	PredictedCost   float64   `json:"predictedCost"`
	LowerBound      float64   `json:"lowerBound"`
	UpperBound      float64   `json:"upperBound"`
	Confidence      float64   `json:"confidence"`
	BudgetRemaining float64   `json:"budgetRemaining,omitempty"`
	OverrunRisk     float64   `json:"overrunRisk"` // 0.0-1.0
}

// CostAutopilotDashboard provides comprehensive cost autopilot overview
type CostAutopilotDashboard struct {
	CurrentMonthCost  float64                    `json:"currentMonthCost"`
	MonthlyBudget     float64                    `json:"monthlyBudget"`
	BudgetUtilization float64                    `json:"budgetUtilization"`
	ProjectedOverrun  float64                    `json:"projectedOverrun"`
	Hotspots          []CostHotspot              `json:"hotspots"`
	Predictions       []AutopilotCostPrediction  `json:"predictions"`
	ActiveRules       int                        `json:"activeRules"`
	SavingsThisMonth  float64                    `json:"savingsThisMonth"`
	Recommendations   []CostRecommendation       `json:"recommendations"`
}
