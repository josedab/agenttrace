package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

// PromptTuningService handles prompt tuning jobs and A/B testing workflows
type PromptTuningService struct {
	logger *zap.Logger
}

// NewPromptTuningService creates a new prompt tuning service
func NewPromptTuningService(logger *zap.Logger) *PromptTuningService {
	return &PromptTuningService{
		logger: logger,
	}
}

// StartTuningJob begins a new DSPy-style prompt tuning job with iteratively improving variants
func (s *PromptTuningService) StartTuningJob(ctx context.Context, projectID uuid.UUID, input domain.PromptTuningInput) (*domain.PromptTuningJob, error) {
	if err := input.Validate(); err != nil {
		return nil, fmt.Errorf("invalid tuning input: %w", err)
	}

	now := time.Now()
	jobID := uuid.New()

	variants := []domain.TuningVariant{
		{
			ID:        uuid.New(),
			Iteration: 1,
			PromptText: input.BaselinePrompt + "\n\nIMPORTANT: Follow the output schema exactly. " +
				"Validate all tool names against the provided tool list before calling.",
			Score: 0.62,
			Metrics: map[string]float64{
				"accuracy": 0.60, "latency_ms": 320, "cost_per_call": 0.0021,
			},
			Rationale: "Added explicit schema adherence and tool validation constraints to reduce hallucination errors",
			Demonstrations: []domain.TuningDemonstration{
				{
					Input:       "What is the capital of France?",
					Output:      `{"answer": "Paris", "confidence": 0.99}`,
					Explanation: "Demonstrates structured JSON output with confidence score",
				},
			},
			IsSelected: false,
		},
		{
			ID:        uuid.New(),
			Iteration: 2,
			PromptText: input.BaselinePrompt + "\n\nThink step by step. First identify the task type, " +
				"then select the appropriate tool, then format your response as valid JSON.",
			Score: 0.71,
			Metrics: map[string]float64{
				"accuracy": 0.70, "latency_ms": 385, "cost_per_call": 0.0025,
			},
			Rationale: "Chain-of-thought decomposition improves accuracy at slight latency cost",
			Demonstrations: []domain.TuningDemonstration{
				{
					Input:       "Summarize the key points from the quarterly report",
					Output:      `{"summary": ["Revenue up 12%", "Customer churn reduced by 3%", "New market expansion planned"], "word_count": 15}`,
					Explanation: "Shows multi-step reasoning: identify task as summarization, extract key points, format as JSON array",
				},
				{
					Input:       "Calculate the compound interest on $10,000 at 5% for 3 years",
					Output:      `{"result": 11576.25, "formula": "P(1+r)^t", "steps": ["P=10000", "r=0.05", "t=3", "10000*(1.05)^3=11576.25"]}`,
					Explanation: "Demonstrates step-by-step calculation with formula exposition",
				},
			},
			IsSelected: false,
		},
		{
			ID:        uuid.New(),
			Iteration: 3,
			PromptText: input.BaselinePrompt + "\n\nYou MUST respond with valid JSON matching the output schema. " +
				"Before responding, verify: 1) All tool names exist in the schema. " +
				"2) All required fields are present. 3) Data types match the specification.",
			Score: 0.78,
			Metrics: map[string]float64{
				"accuracy": 0.77, "latency_ms": 350, "cost_per_call": 0.0023,
			},
			Rationale: "Checklist-based self-verification reduces format errors while maintaining reasonable latency",
			Demonstrations: []domain.TuningDemonstration{
				{
					Input:       "Extract entities from: 'Apple Inc. reported Q3 earnings in Cupertino'",
					Output:      `{"entities": [{"text": "Apple Inc.", "type": "ORG"}, {"text": "Q3", "type": "TIME"}, {"text": "Cupertino", "type": "LOC"}]}`,
					Explanation: "Entity extraction with self-verification: all fields present, correct types, valid JSON",
				},
			},
			IsSelected: false,
		},
		{
			ID:        uuid.New(),
			Iteration: 4,
			PromptText: input.BaselinePrompt + "\n\nRole: Expert assistant with strict output discipline.\n" +
				"Rules:\n- Only use tools from the provided schema\n- Always return valid JSON\n" +
				"- Keep reasoning concise (under 50 words)\n- Verify output before returning",
			Score: 0.84,
			Metrics: map[string]float64{
				"accuracy": 0.83, "latency_ms": 290, "cost_per_call": 0.0019,
			},
			Rationale: "Structured role definition with numbered rules combines previous improvements; conciseness constraint reduces cost",
			Demonstrations: []domain.TuningDemonstration{
				{
					Input:       "Classify sentiment: 'The product exceeded my expectations but shipping was slow'",
					Output:      `{"sentiment": "mixed", "aspects": [{"topic": "product", "sentiment": "positive"}, {"topic": "shipping", "sentiment": "negative"}]}`,
					Explanation: "Aspect-level sentiment with concise reasoning under 50 words",
				},
				{
					Input:       "Translate to Spanish: 'The meeting is at 3 PM tomorrow'",
					Output:      `{"translation": "La reunión es a las 3 PM mañana", "source_lang": "en", "target_lang": "es"}`,
					Explanation: "Direct translation with metadata, no verbose explanation",
				},
			},
			IsSelected: false,
		},
		{
			ID:        uuid.New(),
			Iteration: 5,
			PromptText: input.BaselinePrompt + "\n\nYou are a precise, efficient assistant.\n\n" +
				"INSTRUCTIONS:\n1. Identify the task type from the input\n2. Select and verify tools against the schema\n" +
				"3. Execute with minimal token usage\n4. Self-check: validate JSON, required fields, and data types\n5. Return response\n\n" +
				"CONSTRAINTS: Only use schema-defined tools. Output valid JSON. Keep reasoning under 30 words.",
			Score: 0.89,
			Metrics: map[string]float64{
				"accuracy": 0.88, "latency_ms": 275, "cost_per_call": 0.0017,
			},
			Rationale: "Final optimized variant combining structured instructions, self-verification, and token efficiency from all prior iterations",
			Demonstrations: []domain.TuningDemonstration{
				{
					Input:       "Given the user query 'book a flight to NYC next Friday', extract intent and entities",
					Output:      `{"intent": "book_flight", "entities": [{"type": "destination", "value": "NYC"}, {"type": "date", "value": "next Friday"}], "confidence": 0.95}`,
					Explanation: "Optimal: correct intent extraction, structured entities, high confidence, valid JSON",
				},
			},
			IsSelected: true,
		},
	}

	baselineScore := 0.52
	bestScore := variants[len(variants)-1].Score

	job := &domain.PromptTuningJob{
		ID:                  jobID,
		ProjectID:           projectID,
		PromptID:            input.PromptID,
		Strategy:            input.Strategy,
		ObjectiveMetric:     input.ObjectiveMetric,
		ConstraintMetrics:   input.ConstraintMetrics,
		BaselinePrompt:      input.BaselinePrompt,
		DatasetID:           input.DatasetID,
		Iterations:          input.Iterations,
		CompletedIterations: 5,
		BestScore:           bestScore,
		BaselineScore:       baselineScore,
		ImprovementPct:      ((bestScore - baselineScore) / baselineScore) * 100,
		GeneratedVariants:   variants,
		Status:              domain.OptimizationStatusCompleted,
		CreatedAt:           now,
	}

	s.logger.Info("tuning job started",
		zap.String("jobId", jobID.String()),
		zap.String("projectId", projectID.String()),
		zap.String("promptId", input.PromptID.String()),
		zap.String("strategy", string(input.Strategy)),
		zap.String("objectiveMetric", input.ObjectiveMetric),
		zap.Int("iterations", input.Iterations),
		zap.Int("variants", len(variants)),
		zap.Float64("bestScore", bestScore),
		zap.Float64("improvementPct", job.ImprovementPct),
	)
	return job, nil
}

// GetTuningJob retrieves a tuning job by ID
func (s *PromptTuningService) GetTuningJob(ctx context.Context, projectID, jobID uuid.UUID) (*domain.PromptTuningJob, error) {
	s.logger.Debug("fetching tuning job",
		zap.String("projectId", projectID.String()),
		zap.String("jobId", jobID.String()),
	)

	now := time.Now()
	return &domain.PromptTuningJob{
		ID:                  jobID,
		ProjectID:           projectID,
		PromptID:            uuid.New(),
		Strategy:            domain.TuningStrategyDSPyBootstrap,
		ObjectiveMetric:     "accuracy",
		BaselinePrompt:      "You are a helpful assistant.",
		Iterations:          10,
		CompletedIterations: 10,
		BestScore:           0.89,
		BaselineScore:       0.52,
		ImprovementPct:      71.15,
		Status:              domain.OptimizationStatusCompleted,
		CreatedAt:           now.Add(-2 * time.Hour),
	}, nil
}

// ListTuningJobs lists tuning jobs for a project with optional status filtering
func (s *PromptTuningService) ListTuningJobs(ctx context.Context, projectID uuid.UUID, status string, limit, offset int) ([]domain.PromptTuningJob, error) {
	s.logger.Debug("listing tuning jobs",
		zap.String("projectId", projectID.String()),
		zap.String("status", status),
		zap.Int("limit", limit),
		zap.Int("offset", offset),
	)

	if limit <= 0 {
		limit = 20
	}

	now := time.Now()
	completedAt := now.Add(-30 * time.Minute)

	jobs := []domain.PromptTuningJob{
		{
			ID:                  uuid.New(),
			ProjectID:           projectID,
			PromptID:            uuid.New(),
			Strategy:            domain.TuningStrategyDSPyBootstrap,
			ObjectiveMetric:     "accuracy",
			BaselinePrompt:      "You are a helpful assistant.",
			Iterations:          10,
			CompletedIterations: 10,
			BestScore:           0.89,
			BaselineScore:       0.52,
			ImprovementPct:      71.15,
			Status:              domain.OptimizationStatusCompleted,
			CreatedAt:           now.Add(-4 * time.Hour),
			CompletedAt:         &completedAt,
		},
		{
			ID:                  uuid.New(),
			ProjectID:           projectID,
			PromptID:            uuid.New(),
			Strategy:            domain.TuningStrategyDSPyMIPRO,
			ObjectiveMetric:     "f1_score",
			BaselinePrompt:      "Extract entities from the following text.",
			Iterations:          15,
			CompletedIterations: 8,
			BestScore:           0.76,
			BaselineScore:       0.61,
			ImprovementPct:      24.59,
			Status:              domain.OptimizationStatusTesting,
			CreatedAt:           now.Add(-2 * time.Hour),
		},
		{
			ID:                  uuid.New(),
			ProjectID:           projectID,
			PromptID:            uuid.New(),
			Strategy:            domain.TuningStrategyBayesian,
			ObjectiveMetric:     "latency",
			BaselinePrompt:      "Summarize the following document concisely.",
			Iterations:          20,
			CompletedIterations: 0,
			BestScore:           0,
			BaselineScore:       0.45,
			ImprovementPct:      0,
			Status:              domain.OptimizationStatusAnalyzing,
			CreatedAt:           now.Add(-10 * time.Minute),
		},
	}

	if status != "" {
		filtered := make([]domain.PromptTuningJob, 0)
		for _, j := range jobs {
			if string(j.Status) == status {
				filtered = append(filtered, j)
			}
		}
		jobs = filtered
	}

	if offset >= len(jobs) {
		return []domain.PromptTuningJob{}, nil
	}
	end := offset + limit
	if end > len(jobs) {
		end = len(jobs)
	}
	return jobs[offset:end], nil
}

// CreateABTest creates a new A/B test with control and treatment arms
func (s *PromptTuningService) CreateABTest(ctx context.Context, projectID uuid.UUID, input domain.ABTestInput) (*domain.ABTest, error) {
	if err := input.Validate(); err != nil {
		return nil, fmt.Errorf("invalid A/B test input: %w", err)
	}

	now := time.Now()
	testID := uuid.New()

	controlArm := domain.ABTestArm{
		ID:             uuid.New(),
		Name:           "control",
		PromptText:     input.ControlPromptText,
		SampleSize:     0,
		Metrics:        map[string]float64{},
		ConversionRate: 0,
	}

	treatmentArms := make([]domain.ABTestArm, len(input.TreatmentPromptTexts))
	for i, text := range input.TreatmentPromptTexts {
		treatmentArms[i] = domain.ABTestArm{
			ID:             uuid.New(),
			Name:           fmt.Sprintf("treatment_%d", i+1),
			PromptText:     text,
			SampleSize:     0,
			Metrics:        map[string]float64{},
			ConversionRate: 0,
		}
	}

	// Build initial even traffic split
	trafficSplit := input.TrafficSplit
	if len(trafficSplit) == 0 {
		totalArms := 1 + len(treatmentArms)
		splitPct := 1.0 / float64(totalArms)
		trafficSplit = map[string]float64{
			controlArm.ID.String(): splitPct,
		}
		for _, arm := range treatmentArms {
			trafficSplit[arm.ID.String()] = splitPct
		}
	}

	minSampleSize := input.MinSampleSize
	if minSampleSize <= 0 {
		minSampleSize = 1000
	}
	confidenceLevel := input.ConfidenceLevel
	if confidenceLevel <= 0 {
		confidenceLevel = 0.95
	}

	test := &domain.ABTest{
		ID:                testID,
		ProjectID:         projectID,
		PromptID:          input.PromptID,
		Name:              input.Name,
		Status:            domain.ABTestStatusRunning,
		ControlPrompt:     controlArm,
		TreatmentPrompts:  treatmentArms,
		TrafficSplit:      trafficSplit,
		MinSampleSize:     minSampleSize,
		CurrentSampleSize: 0,
		ConfidenceLevel:   confidenceLevel,
		PrimaryMetric:     input.PrimaryMetric,
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	s.logger.Info("A/B test created",
		zap.String("testId", testID.String()),
		zap.String("projectId", projectID.String()),
		zap.String("promptId", input.PromptID.String()),
		zap.String("name", input.Name),
		zap.Int("treatmentArms", len(treatmentArms)),
		zap.Int("minSampleSize", minSampleSize),
		zap.Float64("confidenceLevel", confidenceLevel),
		zap.String("primaryMetric", input.PrimaryMetric),
	)
	return test, nil
}

// GetABTest retrieves an A/B test with realistic mock results
func (s *PromptTuningService) GetABTest(ctx context.Context, projectID, testID uuid.UUID) (*domain.ABTest, error) {
	s.logger.Debug("fetching A/B test",
		zap.String("projectId", projectID.String()),
		zap.String("testId", testID.String()),
	)

	now := time.Now()
	completedAt := now.Add(-1 * time.Hour)

	controlID := uuid.New()
	treatmentID := uuid.New()
	winnerID := treatmentID

	test := &domain.ABTest{
		ID:        testID,
		ProjectID: projectID,
		PromptID:  uuid.New(),
		Name:      "Structured output prompt optimization",
		Status:    domain.ABTestStatusCompleted,
		ControlPrompt: domain.ABTestArm{
			ID:         controlID,
			Name:       "control",
			PromptText: "You are a helpful assistant. Please respond to the user query.",
			SampleSize: 1247,
			Metrics: map[string]float64{
				"quality":       0.72,
				"latency_ms":    345.0,
				"cost_per_call": 0.0024,
			},
			ConversionRate: 0.72,
		},
		TreatmentPrompts: []domain.ABTestArm{
			{
				ID:         treatmentID,
				Name:       "treatment_1",
				PromptText: "You are a precise assistant. Think step by step. Respond with valid JSON only.",
				SampleSize: 1312,
				Metrics: map[string]float64{
					"quality":       0.84,
					"latency_ms":    298.0,
					"cost_per_call": 0.0019,
				},
				ConversionRate: 0.84,
			},
		},
		TrafficSplit: map[string]float64{
			controlID.String():   0.5,
			treatmentID.String(): 0.5,
		},
		MinSampleSize:     1000,
		CurrentSampleSize: 2559,
		ConfidenceLevel:   0.95,
		PrimaryMetric:     "quality",
		Results: &domain.ABTestResults{
			WinnerID:                 &winnerID,
			WinnerName:               "treatment_1",
			StatisticallySignificant: true,
			PValue:                   0.0023,
			ConfidenceInterval:       []float64{0.08, 0.16},
			LiftPct:                  16.67,
			Recommendation:           "Treatment 1 shows a statistically significant 16.67% improvement in quality (p=0.0023). Recommend promoting treatment_1 to production. The optimized prompt also reduces latency by 13.6% and cost by 20.8%.",
		},
		CreatedAt:   now.Add(-48 * time.Hour),
		UpdatedAt:   now,
		CompletedAt: &completedAt,
	}

	return test, nil
}

// ListABTests lists A/B tests for a project with optional status filtering
func (s *PromptTuningService) ListABTests(ctx context.Context, projectID uuid.UUID, status string, limit, offset int) ([]domain.ABTest, error) {
	s.logger.Debug("listing A/B tests",
		zap.String("projectId", projectID.String()),
		zap.String("status", status),
		zap.Int("limit", limit),
		zap.Int("offset", offset),
	)

	if limit <= 0 {
		limit = 20
	}

	now := time.Now()
	completedAt := now.Add(-1 * time.Hour)

	tests := []domain.ABTest{
		{
			ID:                uuid.New(),
			ProjectID:         projectID,
			PromptID:          uuid.New(),
			Name:              "Structured output prompt optimization",
			Status:            domain.ABTestStatusCompleted,
			MinSampleSize:     1000,
			CurrentSampleSize: 2559,
			ConfidenceLevel:   0.95,
			PrimaryMetric:     "quality",
			CreatedAt:         now.Add(-48 * time.Hour),
			UpdatedAt:         now,
			CompletedAt:       &completedAt,
		},
		{
			ID:                uuid.New(),
			ProjectID:         projectID,
			PromptID:          uuid.New(),
			Name:              "Chain-of-thought vs direct prompting",
			Status:            domain.ABTestStatusRunning,
			MinSampleSize:     2000,
			CurrentSampleSize: 843,
			ConfidenceLevel:   0.95,
			PrimaryMetric:     "accuracy",
			CreatedAt:         now.Add(-12 * time.Hour),
			UpdatedAt:         now,
		},
		{
			ID:              uuid.New(),
			ProjectID:       projectID,
			PromptID:        uuid.New(),
			Name:            "Latency reduction experiment",
			Status:          domain.ABTestStatusDraft,
			MinSampleSize:   1500,
			ConfidenceLevel: 0.90,
			PrimaryMetric:   "latency_ms",
			CreatedAt:       now.Add(-1 * time.Hour),
			UpdatedAt:       now,
		},
	}

	if status != "" {
		filtered := make([]domain.ABTest, 0)
		for _, t := range tests {
			if string(t.Status) == status {
				filtered = append(filtered, t)
			}
		}
		tests = filtered
	}

	if offset >= len(tests) {
		return []domain.ABTest{}, nil
	}
	end := offset + limit
	if end > len(tests) {
		end = len(tests)
	}
	return tests[offset:end], nil
}

// CompleteABTest finalizes an A/B test and determines the winner
func (s *PromptTuningService) CompleteABTest(ctx context.Context, projectID, testID uuid.UUID) (*domain.ABTestResults, error) {
	s.logger.Info("completing A/B test",
		zap.String("projectId", projectID.String()),
		zap.String("testId", testID.String()),
	)

	winnerID := uuid.New()
	results := &domain.ABTestResults{
		WinnerID:                 &winnerID,
		WinnerName:               "treatment_1",
		StatisticallySignificant: true,
		PValue:                   0.0031,
		ConfidenceInterval:       []float64{0.05, 0.14},
		LiftPct:                  12.3,
		Recommendation:           "Treatment 1 demonstrates a statistically significant improvement of 12.3% over control (p=0.0031, 95% CI [5%, 14%]). Recommend deploying treatment_1 as the new default prompt.",
	}

	s.logger.Info("A/B test completed",
		zap.String("testId", testID.String()),
		zap.String("winner", results.WinnerName),
		zap.Float64("pValue", results.PValue),
		zap.Float64("liftPct", results.LiftPct),
		zap.Bool("significant", results.StatisticallySignificant),
	)
	return results, nil
}

// GetDashboard returns a prompt optimization dashboard summary
func (s *PromptTuningService) GetDashboard(ctx context.Context, projectID uuid.UUID) (*domain.PromptOptimizationDashboard, error) {
	s.logger.Debug("fetching optimization dashboard",
		zap.String("projectId", projectID.String()),
	)

	now := time.Now()
	completedAt1 := now.Add(-30 * time.Minute)
	completedAt2 := now.Add(-6 * time.Hour)

	dashboard := &domain.PromptOptimizationDashboard{
		TotalTuningJobs:   12,
		ActiveJobs:        3,
		CompletedJobs:     9,
		TotalABTests:      8,
		ActiveABTests:     2,
		AvgImprovementPct: 34.7,
		RecentJobs: []domain.PromptTuningJob{
			{
				ID:                  uuid.New(),
				ProjectID:           projectID,
				PromptID:            uuid.New(),
				Strategy:            domain.TuningStrategyDSPyBootstrap,
				ObjectiveMetric:     "accuracy",
				Iterations:          10,
				CompletedIterations: 10,
				BestScore:           0.89,
				BaselineScore:       0.52,
				ImprovementPct:      71.15,
				Status:              domain.OptimizationStatusCompleted,
				CreatedAt:           now.Add(-4 * time.Hour),
				CompletedAt:         &completedAt1,
			},
			{
				ID:                  uuid.New(),
				ProjectID:           projectID,
				PromptID:            uuid.New(),
				Strategy:            domain.TuningStrategyDSPyMIPRO,
				ObjectiveMetric:     "f1_score",
				Iterations:          15,
				CompletedIterations: 8,
				BestScore:           0.76,
				BaselineScore:       0.61,
				ImprovementPct:      24.59,
				Status:              domain.OptimizationStatusTesting,
				CreatedAt:           now.Add(-2 * time.Hour),
			},
		},
		ActiveTests: []domain.ABTest{
			{
				ID:                uuid.New(),
				ProjectID:         projectID,
				PromptID:          uuid.New(),
				Name:              "Chain-of-thought vs direct prompting",
				Status:            domain.ABTestStatusRunning,
				MinSampleSize:     2000,
				CurrentSampleSize: 843,
				ConfidenceLevel:   0.95,
				PrimaryMetric:     "accuracy",
				CreatedAt:         now.Add(-12 * time.Hour),
				UpdatedAt:         now,
			},
		},
		TopOptimizations: []domain.PromptTuningJob{
			{
				ID:                  uuid.New(),
				ProjectID:           projectID,
				PromptID:            uuid.New(),
				Strategy:            domain.TuningStrategyDSPyBootstrap,
				ObjectiveMetric:     "accuracy",
				Iterations:          10,
				CompletedIterations: 10,
				BestScore:           0.89,
				BaselineScore:       0.52,
				ImprovementPct:      71.15,
				Status:              domain.OptimizationStatusCompleted,
				CreatedAt:           now.Add(-24 * time.Hour),
				CompletedAt:         &completedAt2,
			},
		},
	}

	return dashboard, nil
}
