package main

import (
	"go.uber.org/zap"

	chrepo "github.com/agenttrace/agenttrace/api/internal/repository/clickhouse"
	pgrepo "github.com/agenttrace/agenttrace/api/internal/repository/postgres"
)

// Repositories holds all repository instances
type Repositories struct {
	// ClickHouse repositories (time-series data)
	Trace           *chrepo.TraceRepository
	Observation     *chrepo.ObservationRepository
	Score           *chrepo.ScoreRepository
	Session         *chrepo.SessionRepository
	Checkpoint      *chrepo.CheckpointRepository
	GitLink         *chrepo.GitLinkRepository
	FileOperation   *chrepo.FileOperationRepository
	TerminalCommand *chrepo.TerminalCommandRepository
	CIRun           *chrepo.CIRunRepository
	SkillProfile    *chrepo.SkillProfileRepository
	PredictiveCostCH *chrepo.PredictiveCostRepository

	// PostgreSQL repositories (relational data)
	User               *pgrepo.UserRepository
	Org                *pgrepo.OrgRepository
	Project            *pgrepo.ProjectRepository
	APIKey             *pgrepo.APIKeyRepository
	Prompt             *pgrepo.PromptRepository
	Dataset            *pgrepo.DatasetRepository
	Evaluator          *pgrepo.EvaluatorRepository
	Webhook            *pgrepo.WebhookRepository
	Experiment         *pgrepo.ExperimentRepository
	Guardrail          *pgrepo.GuardrailRepository
	Regression         *pgrepo.RegressionRepository
	Benchmark          *pgrepo.BenchmarkRepository
	Collaboration      *pgrepo.CollaborationRepository
	Migration          *pgrepo.MigrationRepository
	Tenant             *pgrepo.TenantRepository
	DebugSession       *pgrepo.DebugSessionRepository
	CostRecommendation *pgrepo.CostRecommendationRepository
	DiffAnalysis       *pgrepo.DiffAnalysisRepository
	Federation         *pgrepo.FederationRepository
	Intervention       *pgrepo.InterventionRepository
	Discussion         *pgrepo.DiscussionRepository
	AlertChannel       *pgrepo.AlertChannelRepository
	CostAutopilot      *pgrepo.CostAutopilotRepository
	RCA                *pgrepo.RCARepository
	AgentVersion       *pgrepo.AgentVersionRepository
	Plugin             *pgrepo.PluginRepository
	Privacy            *pgrepo.PrivacyRepository
	PredictiveCost     *pgrepo.PredictiveCostRepository
}

// initRepositories initializes all repositories
func initRepositories(dbs *Databases, logger *zap.Logger) *Repositories {
	return &Repositories{
		// ClickHouse repositories
		Trace:           chrepo.NewTraceRepository(dbs.ClickHouse, logger),
		Observation:     chrepo.NewObservationRepository(dbs.ClickHouse, logger),
		Score:           chrepo.NewScoreRepository(dbs.ClickHouse, logger),
		Session:         chrepo.NewSessionRepository(dbs.ClickHouse),
		Checkpoint:      chrepo.NewCheckpointRepository(dbs.ClickHouse),
		GitLink:         chrepo.NewGitLinkRepository(dbs.ClickHouse),
		FileOperation:   chrepo.NewFileOperationRepository(dbs.ClickHouse),
		TerminalCommand: chrepo.NewTerminalCommandRepository(dbs.ClickHouse),
		CIRun:           chrepo.NewCIRunRepository(dbs.ClickHouse),
		SkillProfile:    chrepo.NewSkillProfileRepository(dbs.ClickHouse),
		PredictiveCostCH: chrepo.NewPredictiveCostRepository(dbs.ClickHouse),

		// PostgreSQL repositories
		User:       pgrepo.NewUserRepository(dbs.Postgres),
		Org:        pgrepo.NewOrgRepository(dbs.Postgres),
		Project:    pgrepo.NewProjectRepository(dbs.Postgres),
		APIKey:     pgrepo.NewAPIKeyRepository(dbs.Postgres),
		Prompt:     pgrepo.NewPromptRepository(dbs.Postgres),
		Dataset:    pgrepo.NewDatasetRepository(dbs.Postgres),
		Evaluator:  pgrepo.NewEvaluatorRepository(dbs.Postgres),
		Webhook:    pgrepo.NewWebhookRepository(dbs.Postgres),
		Experiment: pgrepo.NewExperimentRepository(dbs.Postgres),
		Guardrail:          pgrepo.NewGuardrailRepository(dbs.Postgres),
		Regression:         pgrepo.NewRegressionRepository(dbs.Postgres),
		Benchmark:          pgrepo.NewBenchmarkRepository(dbs.Postgres),
		Collaboration:      pgrepo.NewCollaborationRepository(dbs.Postgres),
		Migration:          pgrepo.NewMigrationRepository(dbs.Postgres),
		Tenant:             pgrepo.NewTenantRepository(dbs.Postgres),
		DebugSession:       pgrepo.NewDebugSessionRepository(dbs.Postgres),
		CostRecommendation: pgrepo.NewCostRecommendationRepository(dbs.Postgres),
		DiffAnalysis:       pgrepo.NewDiffAnalysisRepository(dbs.Postgres),
		Federation:         pgrepo.NewFederationRepository(dbs.Postgres),
		Intervention:       pgrepo.NewInterventionRepository(dbs.Postgres),
		Discussion:         pgrepo.NewDiscussionRepository(dbs.Postgres),
		AlertChannel:       pgrepo.NewAlertChannelRepository(dbs.Postgres),
		CostAutopilot:      pgrepo.NewCostAutopilotRepository(dbs.Postgres),
		RCA:                pgrepo.NewRCARepository(dbs.Postgres),
		AgentVersion:       pgrepo.NewAgentVersionRepository(dbs.Postgres),
		Plugin:             pgrepo.NewPluginRepository(dbs.Postgres),
		Privacy:            pgrepo.NewPrivacyRepository(dbs.Postgres),
		PredictiveCost:     pgrepo.NewPredictiveCostRepository(dbs.Postgres),
	}
}
