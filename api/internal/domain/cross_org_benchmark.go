package domain

import (
	"time"

	"github.com/google/uuid"
)

// CrossOrgBenchmark represents an anonymized industry benchmark for a specific
// metric, allowing organizations to compare their AI agent performance against peers.
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

// CrossOrgSubmission represents an organization's anonymized metric submission
// to the cross-organization benchmark pool.
type CrossOrgSubmission struct {
	ID          uuid.UUID          `json:"id"`
	ProjectID   uuid.UUID          `json:"projectId"`
	Category    string             `json:"category"`
	Metrics     map[string]float64 `json:"metrics"`
	Anonymized  bool               `json:"anonymized"`
	SubmittedAt time.Time          `json:"submittedAt"`
}

// CrossOrgReport provides a project's benchmark results compared to industry
// data, highlighting strengths and areas for improvement.
type CrossOrgReport struct {
	ProjectID         uuid.UUID           `json:"projectId"`
	Benchmarks        []CrossOrgBenchmark `json:"benchmarks"`
	OverallPercentile float64             `json:"overallPercentile"`
	StrongAreas       []string            `json:"strongAreas"`
	WeakAreas         []string            `json:"weakAreas"`
}

// CrossOrgSubmissionInput is the input for submitting metrics to the cross-org benchmark.
type CrossOrgSubmissionInput struct {
	Category string             `json:"category" validate:"required"`
	Metrics  map[string]float64 `json:"metrics" validate:"required"`
}
