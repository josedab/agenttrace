package domain

import (
	"time"

	"github.com/google/uuid"
)

// BudgetAutoAction represents the automatic action taken when a budget threshold is exceeded
type BudgetAutoAction string

const (
	BudgetAutoActionNone        BudgetAutoAction = "NONE"
	BudgetAutoActionAlertOnly   BudgetAutoAction = "ALERT_ONLY"
	BudgetAutoActionThrottle    BudgetAutoAction = "THROTTLE"
	BudgetAutoActionSwitchModel BudgetAutoAction = "SWITCH_MODEL"
	BudgetAutoActionBlock       BudgetAutoAction = "BLOCK"
)

// IsValid checks if the budget auto action is valid
func (a BudgetAutoAction) IsValid() bool {
	switch a {
	case BudgetAutoActionNone, BudgetAutoActionAlertOnly, BudgetAutoActionThrottle, BudgetAutoActionSwitchModel, BudgetAutoActionBlock:
		return true
	}
	return false
}

// CostBudget represents a cost budget configuration for a project
type CostBudget struct {
	ID                    uuid.UUID         `json:"id"`
	ProjectID             uuid.UUID         `json:"projectId"`
	Name                  string            `json:"name"`
	MonthlyLimitCents     int64             `json:"monthlyLimitCents"`
	AlertThresholds       []BudgetThreshold `json:"alertThresholds"`
	AutoAction            BudgetAutoAction  `json:"autoAction"`
	CurrentSpendCents     int64             `json:"currentSpendCents"`
	ForecastedSpendCents  int64             `json:"forecastedSpendCents"`
	ForecastExhaustionDate *time.Time       `json:"forecastExhaustionDate,omitempty"`
	Enabled               bool              `json:"enabled"`
	CreatedAt             time.Time         `json:"createdAt"`
	UpdatedAt             time.Time         `json:"updatedAt"`
}

// BudgetThreshold represents a notification threshold within a cost budget
type BudgetThreshold struct {
	Percent   int    `json:"percent"` // e.g. 50, 80, 100
	NotifyVia string `json:"notifyVia"` // webhook, email, slack
	Notified  bool   `json:"notified"`
}

// BudgetForecast represents a cost forecast for a project
type BudgetForecast struct {
	ProjectID           uuid.UUID         `json:"projectId"`
	CurrentDailyRate    float64           `json:"currentDailyRate"`
	ProjectedMonthlyTotal float64         `json:"projectedMonthlyTotal"`
	ExhaustionDate      *time.Time        `json:"exhaustionDate,omitempty"`
	DataPointsDays      int               `json:"dataPointsDays"`
	Confidence          float64           `json:"confidence"`
	DailySpend          []DailySpendPoint `json:"dailySpend"`
}

// DailySpendPoint represents spend data for a single day
type DailySpendPoint struct {
	Date        string `json:"date"` // YYYY-MM-DD
	AmountCents int64  `json:"amountCents"`
	TraceCount  int    `json:"traceCount"`
}

// CostBudgetInput represents input for creating or updating a cost budget
type CostBudgetInput struct {
	Name              string            `json:"name"`
	MonthlyLimitCents int64             `json:"monthlyLimitCents"`
	AlertThresholds   []BudgetThreshold `json:"alertThresholds"`
	AutoAction        BudgetAutoAction  `json:"autoAction"`
}
