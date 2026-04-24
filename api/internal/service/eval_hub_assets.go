package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	apperrors "github.com/agenttrace/agenttrace/api/internal/pkg/errors"
)

const maxPublishedDatasetItems = 5000

// EvalHubAssetSnapshot is an immutable source snapshot ready for publication.
type EvalHubAssetSnapshot struct {
	Name        string
	Description string
	Manifest    json.RawMessage
}

// EvalHubAssetRunOutcome is the adapter result persisted as an Eval Hub run.
type EvalHubAssetRunOutcome struct {
	Status            domain.EvalHubRunStatus
	DatasetRunID      *uuid.UUID
	ExperimentID      *uuid.UUID
	Result            json.RawMessage
	CapabilityMessage string
	CompletedAt       *time.Time
}

// EvalHubAssetManager bridges existing evaluator, dataset, prompt, experiment, and benchmark domains.
type EvalHubAssetManager interface {
	Snapshot(
		ctx context.Context,
		projectID uuid.UUID,
		kind domain.EvalHubAssetKind,
		sourceResourceID uuid.UUID,
	) (*EvalHubAssetSnapshot, error)
	Fork(
		ctx context.Context,
		projectID, userID uuid.UUID,
		kind domain.EvalHubAssetKind,
		name string,
		manifest json.RawMessage,
	) (uuid.UUID, json.RawMessage, error)
	Run(
		ctx context.Context,
		projectID, userID uuid.UUID,
		kind domain.EvalHubAssetKind,
		sourceResourceID uuid.UUID,
		manifest json.RawMessage,
		input domain.EvalHubRunInput,
	) (*EvalHubAssetRunOutcome, error)
}

type datasetAssetManifest struct {
	Dataset domain.Dataset       `json:"dataset"`
	Items   []domain.DatasetItem `json:"items"`
}

type promptAssetManifest struct {
	Prompt  domain.Prompt        `json:"prompt"`
	Version domain.PromptVersion `json:"version"`
}

// DefaultEvalHubAssetManager adapts existing domain services without creating parallel stores.
// Benchmarks are intentionally absent: they carry no owning project, so they
// cannot be packaged without crossing tenant boundaries.
type DefaultEvalHubAssetManager struct {
	datasets    *DatasetService
	evaluators  *EvalService
	prompts     *PromptService
	experiments *ExperimentService
}

// NewDefaultEvalHubAssetManager creates the canonical asset adapter.
func NewDefaultEvalHubAssetManager(
	datasets *DatasetService,
	evaluators *EvalService,
	prompts *PromptService,
	experiments *ExperimentService,
) *DefaultEvalHubAssetManager {
	return &DefaultEvalHubAssetManager{
		datasets:    datasets,
		evaluators:  evaluators,
		prompts:     prompts,
		experiments: experiments,
	}
}

// Snapshot loads and project-validates an existing asset.
func (m *DefaultEvalHubAssetManager) Snapshot(
	ctx context.Context,
	projectID uuid.UUID,
	kind domain.EvalHubAssetKind,
	sourceResourceID uuid.UUID,
) (*EvalHubAssetSnapshot, error) {
	switch kind {
	case domain.EvalHubDataset:
		dataset, err := m.datasets.GetForProject(ctx, projectID, sourceResourceID)
		if err != nil {
			return nil, err
		}
		items, total, err := m.datasets.ListItems(
			ctx,
			&domain.DatasetItemFilter{DatasetID: dataset.ID},
			maxPublishedDatasetItems+1,
			0,
		)
		if err != nil {
			return nil, err
		}
		if total > maxPublishedDatasetItems {
			return nil, apperrors.Validation(
				fmt.Sprintf("datasets with more than %d items cannot be published", maxPublishedDatasetItems),
			)
		}
		for index := range items {
			items[index].SourceTraceID = nil
			items[index].SourceObservationID = nil
		}
		return marshalAssetSnapshot(dataset.Name, dataset.Description, datasetAssetManifest{
			Dataset: *dataset,
			Items:   items,
		})

	case domain.EvalHubEvaluator:
		evaluator, err := m.evaluators.Get(ctx, sourceResourceID)
		if err != nil {
			return nil, err
		}
		if evaluator.ProjectID != projectID {
			return nil, apperrors.NotFound("evaluator")
		}
		return marshalAssetSnapshot(evaluator.Name, evaluator.Description, evaluator)

	case domain.EvalHubPrompt:
		prompt, err := m.prompts.GetForProject(ctx, projectID, sourceResourceID)
		if err != nil {
			return nil, err
		}
		if prompt.LatestVersion == nil {
			return nil, apperrors.Validation("prompt has no published version")
		}
		return marshalAssetSnapshot(prompt.Name, prompt.Description, promptAssetManifest{
			Prompt:  *prompt,
			Version: *prompt.LatestVersion,
		})

	case domain.EvalHubExperiment:
		experiment, err := m.experiments.GetByID(ctx, sourceResourceID)
		if err != nil {
			return nil, err
		}
		if experiment.ProjectID != projectID {
			return nil, apperrors.NotFound("experiment")
		}
		return marshalAssetSnapshot(experiment.Name, experiment.Description, experiment)

	case domain.EvalHubBenchmark:
		return nil, errBenchmarkPackagingUnsupported()

	default:
		return nil, apperrors.Validation("unsupported Eval Hub asset kind")
	}
}

// errBenchmarkPackagingUnsupported explains why benchmarks cannot be packaged.
// Benchmarks are stored without an owning project and reference a dataset and
// evaluators from the project that created them, so publishing or forking one
// would move another project's resource identifiers across tenant boundaries.
func errBenchmarkPackagingUnsupported() error {
	return apperrors.Unprocessable(
		"benchmark packages are not supported because benchmarks are not project-owned; " +
			"publish the benchmark dataset and evaluators instead",
	)
}

// Fork materializes an asset into the target project when the underlying domain supports it.
func (m *DefaultEvalHubAssetManager) Fork(
	ctx context.Context,
	projectID, userID uuid.UUID,
	kind domain.EvalHubAssetKind,
	name string,
	manifest json.RawMessage,
) (uuid.UUID, json.RawMessage, error) {
	switch kind {
	case domain.EvalHubDataset:
		return m.forkDataset(ctx, projectID, name, manifest)

	case domain.EvalHubEvaluator:
		var source domain.Evaluator
		if err := json.Unmarshal(manifest, &source); err != nil {
			return uuid.Nil, nil, fmt.Errorf("decode evaluator package: %w", err)
		}
		description := source.Description
		evaluatorType := source.Type
		scoreDataType := source.ScoreDataType
		samplingRate := source.SamplingRate
		enabled := source.Enabled
		created, err := m.evaluators.Create(ctx, projectID, &domain.EvaluatorInput{
			Name:            name,
			Description:     &description,
			Type:            &evaluatorType,
			Config:          parseAssetJSON(source.Config),
			PromptTemplate:  source.PromptTemplate,
			Variables:       source.Variables,
			TargetFilter:    parseAssetJSON(source.TargetFilter),
			SamplingRate:    &samplingRate,
			ScoreName:       source.ScoreName,
			ScoreDataType:   &scoreDataType,
			ScoreCategories: source.ScoreCategories,
			Enabled:         &enabled,
		}, userID)
		if err != nil {
			return uuid.Nil, nil, err
		}
		updated, err := json.Marshal(created)
		return created.ID, updated, err

	case domain.EvalHubPrompt:
		var source promptAssetManifest
		if err := json.Unmarshal(manifest, &source); err != nil {
			return uuid.Nil, nil, fmt.Errorf("decode prompt package: %w", err)
		}
		description := source.Prompt.Description
		created, err := m.prompts.Create(ctx, projectID, &domain.PromptInput{
			Name:        name,
			Type:        source.Prompt.Type,
			Description: &description,
			Tags:        source.Prompt.Tags,
			Content:     source.Version.Content,
			Config:      parseAssetJSON(source.Version.Config),
			Labels:      source.Version.Labels,
		}, userID)
		if err != nil {
			return uuid.Nil, nil, err
		}
		refreshed, err := m.prompts.GetForProject(ctx, projectID, created.ID)
		if err != nil {
			return uuid.Nil, nil, err
		}
		if refreshed.LatestVersion == nil {
			return uuid.Nil, nil, fmt.Errorf("forked prompt %s has no version", created.ID)
		}
		updated, err := json.Marshal(promptAssetManifest{
			Prompt:  *refreshed,
			Version: *refreshed.LatestVersion,
		})
		return created.ID, updated, err

	case domain.EvalHubExperiment:
		var source domain.Experiment
		if err := json.Unmarshal(manifest, &source); err != nil {
			return uuid.Nil, nil, fmt.Errorf("decode experiment package: %w", err)
		}
		variants := make([]domain.VariantInput, 0, len(source.Variants))
		for _, variant := range source.Variants {
			variants = append(variants, domain.VariantInput{
				Name:        variant.Name,
				Description: variant.Description,
				Weight:      variant.Weight,
				IsControl:   variant.IsControl,
				Config:      variant.Config,
			})
		}
		created, err := m.experiments.CreateExperiment(ctx, projectID, userID, &domain.ExperimentInput{
			Name:            name,
			Description:     source.Description,
			Variants:        variants,
			TargetMetric:    source.TargetMetric,
			TargetGoal:      source.TargetGoal,
			TrafficPercent:  source.TrafficPercent,
			TraceNameFilter: source.TraceNameFilter,
			UserIDFilter:    source.UserIDFilter,
			MetadataFilters: source.MetadataFilters,
			MinDuration:     source.MinDuration,
			MinSamples:      source.MinSamples,
		})
		if err != nil {
			return uuid.Nil, nil, err
		}
		updated, err := json.Marshal(created)
		return created.ID, updated, err

	case domain.EvalHubBenchmark:
		return uuid.Nil, nil, errBenchmarkPackagingUnsupported()

	default:
		return uuid.Nil, nil, apperrors.Validation("unsupported Eval Hub asset kind")
	}
}

// forkDataset materializes a project-owned copy of a packaged dataset.
// Every item is decoded and validated before anything is created, and a partial
// materialization is rolled back, so a fork never yields a package that points
// at an incomplete dataset.
func (m *DefaultEvalHubAssetManager) forkDataset(
	ctx context.Context,
	projectID uuid.UUID,
	name string,
	manifest json.RawMessage,
) (uuid.UUID, json.RawMessage, error) {
	var source datasetAssetManifest
	if err := json.Unmarshal(manifest, &source); err != nil {
		return uuid.Nil, nil, fmt.Errorf("decode dataset package: %w", err)
	}

	inputs := make([]*domain.DatasetItemInput, 0, len(source.Items))
	for index, item := range source.Items {
		itemInput, err := decodeDatasetItem(item)
		if err != nil {
			return uuid.Nil, nil, apperrors.Validation(
				fmt.Sprintf("dataset package item %d is invalid: %s", index, err.Error()),
			)
		}
		inputs = append(inputs, itemInput)
	}

	created, err := m.datasets.Create(ctx, projectID, &domain.DatasetInput{
		Name:        name,
		Description: source.Dataset.Description,
		Metadata: map[string]interface{}{
			"evalHubFork": true,
		},
	})
	if err != nil {
		return uuid.Nil, nil, err
	}

	forkedItems := make([]domain.DatasetItem, 0, len(inputs))
	for _, itemInput := range inputs {
		createdItem, itemErr := m.datasets.AddItem(ctx, created.ID, itemInput)
		if itemErr != nil {
			return uuid.Nil, nil, m.rollbackDatasetFork(
				ctx,
				projectID,
				created.ID,
				fmt.Errorf("materialize dataset item: %w", itemErr),
			)
		}
		forkedItems = append(forkedItems, *createdItem)
	}

	updated, err := json.Marshal(datasetAssetManifest{Dataset: *created, Items: forkedItems})
	if err != nil {
		return uuid.Nil, nil, m.rollbackDatasetFork(
			ctx,
			projectID,
			created.ID,
			fmt.Errorf("encode forked dataset: %w", err),
		)
	}
	return created.ID, updated, nil
}

// rollbackDatasetFork removes a partially materialized fork and reports whether
// the cleanup itself failed, so no silent orphan is left behind.
func (m *DefaultEvalHubAssetManager) rollbackDatasetFork(
	ctx context.Context,
	projectID, datasetID uuid.UUID,
	cause error,
) error {
	if err := m.datasets.DeleteForProject(ctx, projectID, datasetID); err != nil {
		return fmt.Errorf("%w; partial dataset %s could not be removed: %v", cause, datasetID, err)
	}
	return cause
}

// decodeDatasetItem converts a packaged item into a creation input.
func decodeDatasetItem(item domain.DatasetItem) (*domain.DatasetItemInput, error) {
	var input interface{}
	if item.Input != "" {
		if err := json.Unmarshal([]byte(item.Input), &input); err != nil {
			input = item.Input
		}
	}
	if input == nil {
		return nil, fmt.Errorf("input is required")
	}

	var expected interface{}
	if item.ExpectedOutput != nil {
		if err := json.Unmarshal([]byte(*item.ExpectedOutput), &expected); err != nil {
			expected = *item.ExpectedOutput
		}
	}
	return &domain.DatasetItemInput{Input: input, ExpectedOutput: expected}, nil
}

// Run delegates to the real project-scoped execution surface when available.
func (m *DefaultEvalHubAssetManager) Run(
	ctx context.Context,
	projectID, _ uuid.UUID,
	kind domain.EvalHubAssetKind,
	sourceResourceID uuid.UUID,
	manifest json.RawMessage,
	input domain.EvalHubRunInput,
) (*EvalHubAssetRunOutcome, error) {
	now := time.Now().UTC()
	switch kind {
	case domain.EvalHubDataset:
		if _, err := m.datasets.GetForProject(ctx, projectID, sourceResourceID); err != nil {
			return nil, err
		}
		runName := input.Name
		if runName == "" {
			runName = "Eval Hub run " + now.Format("2006-01-02 15:04")
		}
		run, err := m.datasets.CreateRun(ctx, sourceResourceID, &domain.DatasetRunInput{
			Name:        runName,
			Description: "Created from a versioned Eval Hub package",
			Metadata: map[string]interface{}{
				"evalHub": true,
			},
		})
		if err != nil {
			return nil, err
		}
		return &EvalHubAssetRunOutcome{
			Status:            domain.EvalHubRunReady,
			DatasetRunID:      &run.ID,
			CapabilityMessage: "Dataset run created; attach generated traces and scores through the existing dataset run API.",
		}, nil

	case domain.EvalHubPrompt:
		var source promptAssetManifest
		if err := json.Unmarshal(manifest, &source); err != nil {
			return nil, fmt.Errorf("decode prompt package: %w", err)
		}
		if missing := domain.ValidateVariables(source.Version.Content, input.Variables); len(missing) > 0 {
			return nil, apperrors.Validation(fmt.Sprintf("missing prompt variables: %v", missing))
		}
		result, err := json.Marshal(map[string]interface{}{
			"promptName": source.Prompt.Name,
			"compiled":   domain.CompilePrompt(source.Version.Content, input.Variables),
			"variables":  input.Variables,
		})
		if err != nil {
			return nil, err
		}
		return &EvalHubAssetRunOutcome{
			Status:      domain.EvalHubRunCompleted,
			Result:      result,
			CompletedAt: &now,
		}, nil

	case domain.EvalHubExperiment:
		experiment, err := m.experiments.GetByID(ctx, sourceResourceID)
		if err != nil {
			return nil, err
		}
		if experiment.ProjectID != projectID {
			return nil, apperrors.NotFound("experiment")
		}
		if err := m.experiments.StartExperiment(ctx, experiment); err != nil {
			return nil, err
		}
		return &EvalHubAssetRunOutcome{
			Status:       domain.EvalHubRunRunning,
			ExperimentID: &experiment.ID,
		}, nil

	case domain.EvalHubEvaluator:
		return &EvalHubAssetRunOutcome{
			Status:            domain.EvalHubRunUnsupported,
			CapabilityMessage: "Evaluator packages require a target trace or observation; use the existing evaluator trigger API after forking.",
			CompletedAt:       &now,
		}, nil

	case domain.EvalHubBenchmark:
		return &EvalHubAssetRunOutcome{
			Status:            domain.EvalHubRunUnsupported,
			CapabilityMessage: "Benchmark packages are not supported; use the benchmark submission API directly.",
			CompletedAt:       &now,
		}, nil

	default:
		return nil, apperrors.Validation("unsupported Eval Hub asset kind")
	}
}

func marshalAssetSnapshot(
	name, description string,
	value interface{},
) (*EvalHubAssetSnapshot, error) {
	manifest, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("encode Eval Hub asset: %w", err)
	}
	return &EvalHubAssetSnapshot{
		Name:        name,
		Description: description,
		Manifest:    manifest,
	}, nil
}

func parseAssetJSON(value string) interface{} {
	if value == "" {
		return nil
	}
	var parsed interface{}
	if err := json.Unmarshal([]byte(value), &parsed); err != nil {
		return value
	}
	return parsed
}
