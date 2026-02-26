package main

import (
	"github.com/gofiber/fiber/v2"
)

// registerPromptsRoutes registers prompt, dataset, and data pipeline routes
func registerPromptsRoutes(public fiber.Router, h *Handlers) {
	// Prompts
	public.Get("/prompts", h.Prompts.ListPrompts)
	public.Get("/prompts/:name", h.Prompts.GetPrompt)
	public.Post("/prompts", h.Prompts.CreatePrompt)
	public.Put("/prompts/:name", h.Prompts.UpdatePrompt)
	public.Delete("/prompts/:name", h.Prompts.DeletePrompt)
	public.Get("/prompts/:name/versions", h.Prompts.ListVersions)
	public.Post("/prompts/:name/labels", h.Prompts.SetLabel)
	public.Delete("/prompts/:name/labels/:label", h.Prompts.RemoveLabel)
	public.Post("/prompts/:name/compile", h.Prompts.CompilePrompt)

	// Datasets
	public.Get("/datasets", h.Datasets.ListDatasets)
	public.Get("/datasets/:id", h.Datasets.GetDataset)
	public.Post("/datasets", h.Datasets.CreateDataset)
	public.Put("/datasets/:id", h.Datasets.UpdateDataset)
	public.Delete("/datasets/:id", h.Datasets.DeleteDataset)
	public.Get("/datasets/:id/items", h.Datasets.ListItems)
	public.Post("/datasets/:id/items", h.Datasets.CreateItem)
	public.Put("/datasets/:datasetId/items/:id", h.Datasets.UpdateItem)
	public.Delete("/datasets/:datasetId/items/:id", h.Datasets.DeleteItem)
	public.Get("/datasets/:id/runs", h.Datasets.ListRuns)
	public.Post("/datasets/:id/runs", h.Datasets.CreateRun)
	public.Get("/datasets/:datasetId/runs/:id", h.Datasets.GetRun)
	public.Post("/datasets/:datasetId/runs/:id/items", h.Datasets.AddRunItem)

	// Export endpoints
	public.Post("/export/data", h.Export.ExportData)
	public.Post("/export/dataset", h.Export.ExportDataset)

	// Import endpoints
	public.Post("/import/dataset", h.Import.ImportDataset)
	public.Post("/import/dataset/csv", h.Import.ImportDatasetCSV)
	public.Post("/import/dataset/openai-finetune", h.Import.ImportOpenAIFinetune)
	public.Post("/import/prompt", h.Import.ImportPrompt)

	// Prompt Regression Testing in CI
	public.Get("/prompt-ci/baselines", h.PromptCI.ListBaselines)
	public.Post("/prompt-ci/baselines", h.PromptCI.CreateBaseline)
	public.Get("/prompt-ci/baselines/:baselineId", h.PromptCI.GetBaseline)
	public.Post("/prompt-ci/compare", h.PromptCI.RunComparison)
	public.Get("/prompt-ci/runs", h.PromptCI.ListRuns)
	public.Post("/prompt-ci/gates", h.PromptCI.CreateGateConfig)
	public.Get("/prompt-ci/gates", h.PromptCI.ListGateConfigs)
	public.Post("/prompt-ci/gates/evaluate", h.PromptCI.EvaluateGate)

	// Continuous Prompt Optimization
	public.Post("/prompt-optimization", h.PromptOptimization.StartOptimization)
	public.Get("/prompt-optimization/:optimizationId", h.PromptOptimization.GetOptimization)
	public.Get("/prompt-optimization", h.PromptOptimization.ListOptimizations)
	public.Get("/prompt-optimization/config", h.PromptOptimization.GetOptConfig)
	public.Put("/prompt-optimization/config", h.PromptOptimization.UpdateOptConfig)
	public.Post("/prompt-optimization/variants/:variantId/approve", h.PromptOptimization.ApproveVariant)
	public.Post("/prompt-optimization/variants/:variantId/reject", h.PromptOptimization.RejectVariant)

	// Prompt Caching
	public.Get("/prompt-cache/analyze", h.PromptCache.Analyze)
	public.Get("/prompt-cache/config", h.PromptCache.GetConfig)
	public.Put("/prompt-cache/config", h.PromptCache.UpdateConfig)
	public.Get("/prompt-cache/stats", h.PromptCache.GetStats)
	public.Post("/prompt-cache/invalidate", h.PromptCache.Invalidate)

	// Synthetic Data
	public.Post("/synthetic-data/generate", h.SyntheticData.Generate)
	public.Get("/synthetic-data/datasets", h.SyntheticData.List)
	public.Get("/synthetic-data/datasets/:datasetId", h.SyntheticData.Get)
	public.Get("/synthetic-data/stats", h.SyntheticData.GetStats)

	// Training Pipeline
	public.Get("/training/datasets", h.TrainingPipeline.ListDatasets)
	public.Post("/training/datasets", h.TrainingPipeline.CreateDataset)
	public.Post("/training/datasets/:datasetId/export", h.TrainingPipeline.ExportDataset)
	public.Get("/training/failure-patterns", h.TrainingPipeline.DetectFailures)
}
