package domain

import (
	"time"
)

// SyntheticDataset represents a generated synthetic dataset
type SyntheticDataset struct {
	ID         string    `json:"id"`
	ProjectID  string    `json:"projectId"`
	Name       string    `json:"name"`
	Type       string    `json:"type"` // code_files, api_responses, terminal_output, adversarial
	ItemCount  int       `json:"itemCount"`
	Language   string    `json:"language,omitempty"`
	Difficulty string    `json:"difficulty"` // easy, medium, hard, adversarial
	Status     string    `json:"status"`     // generating, ready
	CreatedAt  time.Time `json:"createdAt"`
}

// SyntheticItem represents a single item in a synthetic dataset
type SyntheticItem struct {
	ID             string   `json:"id"`
	DatasetID      string   `json:"datasetId"`
	Input          string   `json:"input"`
	ExpectedOutput string   `json:"expectedOutput"`
	Difficulty     string   `json:"difficulty"`
	Tags           []string `json:"tags"`
}

// GenerateInput represents input for generating a synthetic dataset
type GenerateInput struct {
	Name             string `json:"name" validate:"required"`
	Type             string `json:"type" validate:"required"`
	Count            int    `json:"count" validate:"required,min=1"`
	Language         string `json:"language,omitempty"`
	Difficulty       string `json:"difficulty" validate:"required"`
	AdversarialFocus string `json:"adversarialFocus,omitempty"`
}

// SyntheticStats provides statistics about synthetic datasets for a project
type SyntheticStats struct {
	TotalDatasets int            `json:"totalDatasets"`
	TotalItems    int            `json:"totalItems"`
	ByType        map[string]int `json:"byType"`
	ByDifficulty  map[string]int `json:"byDifficulty"`
}
