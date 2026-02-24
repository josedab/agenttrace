package domain

import (
	"time"

	"github.com/google/uuid"
)

type CrossOrgBenchmark struct {
	ID                uuid.UUID `json:"id"`
	Category          string    `json:"category"`
	MetricName        string    `json:"metricName"`
	Percentile        float64   `json:"percentile"`
	IndustryAvg       float64   `json:"industryAvg"`
	IndustryP50       float64   `json:"industryP50"`
	IndustryP90       float64   `json:"industryP90"`
	AnonymousRank     int       `json:"anonymousRank"`
	TotalParticipants int       `json:"totalParticipants"`
	UpdatedAt         time.Time `json:"updatedAt"`
}

type CrossOrgSubmission struct {
	ID          uuid.UUID          `json:"id"`
	ProjectID   uuid.UUID          `json:"projectId"`
	Category    string             `json:"category"`
	Metrics     map[string]float64 `json:"metrics"`
	Anonymized  bool               `json:"anonymized"`
	SubmittedAt time.Time          `json:"submittedAt"`
}

type CrossOrgReport struct {
	ProjectID         uuid.UUID           `json:"projectId"`
	Benchmarks        []CrossOrgBenchmark `json:"benchmarks"`
	OverallPercentile float64             `json:"overallPercentile"`
	StrongAreas       []string            `json:"strongAreas"`
	WeakAreas         []string            `json:"weakAreas"`
}

type CrossOrgSubmissionInput struct {
	Category string             `json:"category" validate:"required"`
	Metrics  map[string]float64 `json:"metrics" validate:"required"`
}
