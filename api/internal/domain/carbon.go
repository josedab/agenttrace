package domain

// CarbonFootprint represents the energy and carbon footprint for a project
type CarbonFootprint struct {
	ProjectID     string                 `json:"projectId"`
	TotalEnergyKWh float64              `json:"totalEnergyKWh"`
	TotalCO2Grams float64               `json:"totalCO2Grams"`
	ByModel       map[string]ModelCarbon `json:"byModel"`
	Period        CarbonDateRange        `json:"period"`
}

// ModelCarbon represents carbon metrics for a specific model
type ModelCarbon struct {
	Model          string  `json:"model"`
	Traces         int     `json:"traces"`
	EnergyKWh      float64 `json:"energyKWh"`
	CO2Grams       float64 `json:"co2Grams"`
	EnergyPerToken float64 `json:"energyPerToken"`
}

// CarbonDateRange represents a time period for carbon tracking
type CarbonDateRange struct {
	From string `json:"from"`
	To   string `json:"to"`
}

// CarbonConfig represents the carbon tracking configuration for a project
type CarbonConfig struct {
	ProjectID              string   `json:"projectId"`
	Enabled                bool     `json:"enabled"`
	Region                 string   `json:"region"` // us-east, eu-west, ap-southeast
	GridIntensityGCO2PerKWh float64 `json:"gridIntensityGCO2PerKWh"`
	CarbonBudgetKg         *float64 `json:"carbonBudgetKg,omitempty"`
	ReportingEnabled       bool     `json:"reportingEnabled"`
}

// CarbonSuggestion represents a suggestion for reducing carbon footprint
type CarbonSuggestion struct {
	CurrentModel   string  `json:"currentModel"`
	SuggestedModel string  `json:"suggestedModel"`
	CO2Reduction   float64 `json:"co2Reduction"`
	QualityImpact  string  `json:"qualityImpact"`
}

// CarbonConfigInput represents input for updating carbon configuration
type CarbonConfigInput struct {
	Enabled                *bool    `json:"enabled,omitempty"`
	Region                 string   `json:"region,omitempty"`
	GridIntensityGCO2PerKWh *float64 `json:"gridIntensityGCO2PerKWh,omitempty"`
	CarbonBudgetKg         *float64 `json:"carbonBudgetKg,omitempty"`
	ReportingEnabled       *bool    `json:"reportingEnabled,omitempty"`
}
