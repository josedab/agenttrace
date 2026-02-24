package service

import (
	"context"
	"sync"

	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

// Model energy coefficients (Wh per token)
var modelEnergyCoefficients = map[string]float64{
	"gpt-4":        0.005,
	"gpt-4-turbo":  0.0045,
	"gpt-3.5":      0.002,
	"claude":        0.004,
	"claude-3-opus": 0.0048,
	"claude-3-sonnet": 0.003,
	"claude-3-haiku":  0.0015,
	"gemini-pro":      0.003,
	"llama-70b":       0.0035,
	"llama-13b":       0.001,
	"mistral-7b":      0.0008,
}

// Default grid intensity by region (gCO2/kWh)
var regionGridIntensity = map[string]float64{
	"us-east":      380.0,
	"eu-west":      250.0,
	"ap-southeast": 500.0,
}

// CarbonService manages energy and carbon tracking
type CarbonService struct {
	logger  *zap.Logger
	mu      sync.RWMutex
	configs map[string]*domain.CarbonConfig // projectID -> config
}

// NewCarbonService creates a new carbon tracking service
func NewCarbonService(logger *zap.Logger) *CarbonService {
	return &CarbonService{
		logger:  logger,
		configs: make(map[string]*domain.CarbonConfig),
	}
}

// GetFootprint calculates the carbon footprint for a project in a given period
func (s *CarbonService) GetFootprint(ctx context.Context, projectID string, period domain.CarbonDateRange) (*domain.CarbonFootprint, error) {
	config := s.getOrDefaultConfig(projectID)

	gridIntensity := config.GridIntensityGCO2PerKWh
	if gridIntensity == 0 {
		gridIntensity = regionGridIntensity[config.Region]
	}

	// Simulate footprint data based on model energy coefficients
	byModel := make(map[string]domain.ModelCarbon)
	var totalEnergy, totalCO2 float64

	mockModels := []struct {
		model  string
		traces int
		tokens int
	}{
		{"gpt-4", 150, 450000},
		{"claude-3-sonnet", 200, 380000},
		{"gpt-3.5", 500, 1200000},
		{"claude-3-haiku", 300, 900000},
	}

	for _, m := range mockModels {
		coefficient, ok := modelEnergyCoefficients[m.model]
		if !ok {
			coefficient = 0.003 // default
		}

		energyWh := float64(m.tokens) * coefficient
		energyKWh := energyWh / 1000.0
		co2Grams := energyKWh * gridIntensity

		byModel[m.model] = domain.ModelCarbon{
			Model:          m.model,
			Traces:         m.traces,
			EnergyKWh:      energyKWh,
			CO2Grams:       co2Grams,
			EnergyPerToken: coefficient,
		}

		totalEnergy += energyKWh
		totalCO2 += co2Grams
	}

	return &domain.CarbonFootprint{
		ProjectID:      projectID,
		TotalEnergyKWh: totalEnergy,
		TotalCO2Grams:  totalCO2,
		ByModel:        byModel,
		Period:         period,
	}, nil
}

// GetConfig returns the carbon configuration for a project
func (s *CarbonService) GetConfig(ctx context.Context, projectID string) (*domain.CarbonConfig, error) {
	return s.getOrDefaultConfig(projectID), nil
}

// UpdateConfig updates the carbon configuration for a project
func (s *CarbonService) UpdateConfig(ctx context.Context, projectID string, input *domain.CarbonConfigInput) (*domain.CarbonConfig, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	config := s.getOrDefaultConfigLocked(projectID)

	if input.Enabled != nil {
		config.Enabled = *input.Enabled
	}
	if input.Region != "" {
		config.Region = input.Region
	}
	if input.GridIntensityGCO2PerKWh != nil {
		config.GridIntensityGCO2PerKWh = *input.GridIntensityGCO2PerKWh
	}
	if input.CarbonBudgetKg != nil {
		config.CarbonBudgetKg = input.CarbonBudgetKg
	}
	if input.ReportingEnabled != nil {
		config.ReportingEnabled = *input.ReportingEnabled
	}

	s.configs[projectID] = config
	s.logger.Info("updated carbon config", zap.String("projectId", projectID))
	return config, nil
}

// GetSuggestions returns model replacement suggestions to reduce carbon footprint
func (s *CarbonService) GetSuggestions(ctx context.Context, projectID string) ([]domain.CarbonSuggestion, error) {
	return []domain.CarbonSuggestion{
		{
			CurrentModel:   "gpt-4",
			SuggestedModel: "claude-3-sonnet",
			CO2Reduction:   40.0,
			QualityImpact:  "minimal",
		},
		{
			CurrentModel:   "gpt-4",
			SuggestedModel: "gpt-3.5",
			CO2Reduction:   60.0,
			QualityImpact:  "moderate",
		},
		{
			CurrentModel:   "claude-3-opus",
			SuggestedModel: "claude-3-haiku",
			CO2Reduction:   68.75,
			QualityImpact:  "moderate",
		},
		{
			CurrentModel:   "llama-70b",
			SuggestedModel: "mistral-7b",
			CO2Reduction:   77.1,
			QualityImpact:  "significant",
		},
	}, nil
}

func (s *CarbonService) getOrDefaultConfig(projectID string) *domain.CarbonConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.getOrDefaultConfigLocked(projectID)
}

func (s *CarbonService) getOrDefaultConfigLocked(projectID string) *domain.CarbonConfig {
	if config, ok := s.configs[projectID]; ok {
		return config
	}
	return &domain.CarbonConfig{
		ProjectID:              projectID,
		Enabled:                true,
		Region:                 "us-east",
		GridIntensityGCO2PerKWh: regionGridIntensity["us-east"],
		ReportingEnabled:       true,
	}
}
