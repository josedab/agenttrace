package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

func TestNewRCAService(t *testing.T) {
	svc := NewRCAService(zap.NewNop())
	assert.NotNil(t, svc)
}

func TestRCAService_AnalyzeTrace(t *testing.T) {
	svc := NewRCAService(zap.NewNop())
	ctx := context.Background()
	projectID := uuid.New()
	traceID := uuid.New()

	report, err := svc.AnalyzeTrace(ctx, projectID, traceID)
	require.NoError(t, err)

	t.Run("report has valid fields", func(t *testing.T) {
		assert.NotEqual(t, uuid.Nil, report.ID)
		assert.Equal(t, projectID, report.ProjectID)
		assert.Equal(t, traceID, report.TraceID)
	})

	t.Run("has category and confidence", func(t *testing.T) {
		assert.NotEmpty(t, string(report.PrimaryCategory))
		assert.GreaterOrEqual(t, report.Confidence, 0.75)
		assert.LessOrEqual(t, report.Confidence, 1.0)
	})

	t.Run("has summary and analysis", func(t *testing.T) {
		assert.NotEmpty(t, report.Summary)
		assert.NotEmpty(t, report.DetailedAnalysis)
	})

	t.Run("has contributing factors", func(t *testing.T) {
		assert.GreaterOrEqual(t, len(report.ContributingFactors), 2)
		for _, f := range report.ContributingFactors {
			assert.NotEmpty(t, string(f.Category))
			assert.NotEmpty(t, f.Description)
			assert.Greater(t, f.Impact, 0.0)
		}
	})

	t.Run("has remediations", func(t *testing.T) {
		assert.Greater(t, len(report.Remediations), 0)
		for _, r := range report.Remediations {
			assert.NotEmpty(t, r.Action)
			assert.NotEmpty(t, r.Description)
			assert.Greater(t, r.Priority, 0)
		}
	})

	t.Run("has similar incidents", func(t *testing.T) {
		assert.Len(t, report.SimilarIncidents, 2)
	})
}

func TestRCAService_GetReport(t *testing.T) {
	svc := NewRCAService(zap.NewNop())
	ctx := context.Background()

	t.Run("existing report", func(t *testing.T) {
		report, _ := svc.AnalyzeTrace(ctx, uuid.New(), uuid.New())
		got, err := svc.GetReport(ctx, report.ID)
		require.NoError(t, err)
		assert.Equal(t, report.ID, got.ID)
	})

	t.Run("not found", func(t *testing.T) {
		_, err := svc.GetReport(ctx, uuid.New())
		assert.Error(t, err)
	})
}

func TestRCAService_ListReports(t *testing.T) {
	svc := NewRCAService(zap.NewNop())
	ctx := context.Background()
	projectA := uuid.New()
	projectB := uuid.New()

	_, _ = svc.AnalyzeTrace(ctx, projectA, uuid.New())
	_, _ = svc.AnalyzeTrace(ctx, projectA, uuid.New())
	_, _ = svc.AnalyzeTrace(ctx, projectB, uuid.New())

	t.Run("filters by project", func(t *testing.T) {
		reportsA, err := svc.ListReports(ctx, projectA)
		require.NoError(t, err)
		assert.Len(t, reportsA, 2)

		reportsB, err := svc.ListReports(ctx, projectB)
		require.NoError(t, err)
		assert.Len(t, reportsB, 1)
	})

	t.Run("empty project returns empty slice", func(t *testing.T) {
		reports, err := svc.ListReports(ctx, uuid.New())
		require.NoError(t, err)
		assert.Empty(t, reports)
	})
}

func TestRCAService_DetectAnomalies(t *testing.T) {
	svc := NewRCAService(zap.NewNop())
	ctx := context.Background()
	projectID := uuid.New()

	t.Run("valid input", func(t *testing.T) {
		input := domain.CorrelatedAnomalyInput{
			AnomalyType:    "latency_spike",
			Title:          "High latency detected",
			Description:    "P99 latency exceeded 5s",
			AffectedTraces: []string{"trace-1", "trace-2"},
			Severity:       "critical",
		}
		anomaly, err := svc.DetectAnomalies(ctx, projectID, input)
		require.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, anomaly.ID)
		assert.Equal(t, projectID, anomaly.ProjectID)
		assert.Equal(t, "latency_spike", anomaly.AnomalyType)
		assert.Equal(t, "critical", anomaly.Severity)
		assert.Equal(t, "High latency detected", anomaly.Title)
		assert.Equal(t, "open", anomaly.Status)
		assert.GreaterOrEqual(t, anomaly.Correlation, 0.7)
		assert.LessOrEqual(t, anomaly.Correlation, 1.0)
		assert.Greater(t, len(anomaly.RootCauses), 0)
		assert.Greater(t, len(anomaly.Remediations), 0)
	})

	t.Run("default severity", func(t *testing.T) {
		input := domain.CorrelatedAnomalyInput{
			AnomalyType: "error_burst",
			Title:       "Error spike",
		}
		anomaly, err := svc.DetectAnomalies(ctx, projectID, input)
		require.NoError(t, err)
		assert.Equal(t, "warning", anomaly.Severity)
	})

	t.Run("nil affected traces defaults to empty slice", func(t *testing.T) {
		input := domain.CorrelatedAnomalyInput{
			AnomalyType: "cost_surge",
			Title:       "Cost increase",
		}
		anomaly, err := svc.DetectAnomalies(ctx, projectID, input)
		require.NoError(t, err)
		assert.NotNil(t, anomaly.AffectedTraces)
		assert.Empty(t, anomaly.AffectedTraces)
	})

	t.Run("missing title returns error", func(t *testing.T) {
		input := domain.CorrelatedAnomalyInput{
			AnomalyType: "latency_spike",
		}
		_, err := svc.DetectAnomalies(ctx, projectID, input)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "title must not be empty")
	})
}

func TestRCAService_GetAnomaly(t *testing.T) {
	svc := NewRCAService(zap.NewNop())
	ctx := context.Background()
	projectID := uuid.New()

	t.Run("existing anomaly", func(t *testing.T) {
		input := domain.CorrelatedAnomalyInput{
			AnomalyType: "latency_spike",
			Title:       "Test anomaly",
		}
		created, _ := svc.DetectAnomalies(ctx, projectID, input)
		got, err := svc.GetAnomaly(ctx, created.ID)
		require.NoError(t, err)
		assert.Equal(t, created.ID, got.ID)
	})

	t.Run("not found", func(t *testing.T) {
		_, err := svc.GetAnomaly(ctx, uuid.New())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "anomaly not found")
	})
}

func TestRCAService_ListAnomalies(t *testing.T) {
	svc := NewRCAService(zap.NewNop())
	ctx := context.Background()
	projectA := uuid.New()
	projectB := uuid.New()

	svc.DetectAnomalies(ctx, projectA, domain.CorrelatedAnomalyInput{AnomalyType: "latency_spike", Title: "A1"})
	svc.DetectAnomalies(ctx, projectA, domain.CorrelatedAnomalyInput{AnomalyType: "error_burst", Title: "A2"})
	svc.DetectAnomalies(ctx, projectB, domain.CorrelatedAnomalyInput{AnomalyType: "cost_surge", Title: "B1"})

	t.Run("filters by project", func(t *testing.T) {
		anomalies, err := svc.ListAnomalies(ctx, projectA)
		require.NoError(t, err)
		assert.Len(t, anomalies, 2)

		anomalies, err = svc.ListAnomalies(ctx, projectB)
		require.NoError(t, err)
		assert.Len(t, anomalies, 1)
	})

	t.Run("empty project returns empty slice", func(t *testing.T) {
		anomalies, err := svc.ListAnomalies(ctx, uuid.New())
		require.NoError(t, err)
		assert.Empty(t, anomalies)
		assert.NotNil(t, anomalies)
	})
}

func TestRCAService_AcknowledgeAnomaly(t *testing.T) {
	svc := NewRCAService(zap.NewNop())
	ctx := context.Background()
	projectID := uuid.New()

	t.Run("sets status to acknowledged", func(t *testing.T) {
		input := domain.CorrelatedAnomalyInput{AnomalyType: "latency_spike", Title: "Ack test"}
		created, _ := svc.DetectAnomalies(ctx, projectID, input)
		assert.Equal(t, "open", created.Status)

		acked, err := svc.AcknowledgeAnomaly(ctx, created.ID)
		require.NoError(t, err)
		assert.Equal(t, "acknowledged", acked.Status)
	})

	t.Run("not found", func(t *testing.T) {
		_, err := svc.AcknowledgeAnomaly(ctx, uuid.New())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "anomaly not found")
	})
}

func TestRCAService_CreateAlertChannel(t *testing.T) {
	svc := NewRCAService(zap.NewNop())
	ctx := context.Background()
	projectID := uuid.New()

	t.Run("valid config", func(t *testing.T) {
		input := domain.DeliveryChannelInput{
			Name:   "Slack Alerts",
			Type:   "slack",
			Config: map[string]string{"webhook_url": "https://hooks.slack.com/test"},
		}
		channel, err := svc.CreateAlertChannel(ctx, projectID, input)
		require.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, channel.ID)
		assert.Equal(t, projectID, channel.ProjectID)
		assert.Equal(t, "Slack Alerts", channel.Name)
		assert.Equal(t, "slack", channel.Type)
		assert.True(t, channel.Enabled)
		assert.Equal(t, "untested", channel.TestStatus)
		assert.False(t, channel.CreatedAt.IsZero())
	})

	t.Run("missing name returns error", func(t *testing.T) {
		input := domain.DeliveryChannelInput{
			Type:   "slack",
			Config: map[string]string{"webhook_url": "https://hooks.slack.com/test"},
		}
		_, err := svc.CreateAlertChannel(ctx, projectID, input)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "name must not be empty")
	})

	t.Run("missing type returns error", func(t *testing.T) {
		input := domain.DeliveryChannelInput{
			Name:   "Test",
			Config: map[string]string{},
		}
		_, err := svc.CreateAlertChannel(ctx, projectID, input)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "type must not be empty")
	})
}

func TestRCAService_ListAlertChannels(t *testing.T) {
	svc := NewRCAService(zap.NewNop())
	ctx := context.Background()
	projectID := uuid.New()

	svc.CreateAlertChannel(ctx, projectID, domain.DeliveryChannelInput{Name: "Ch1", Type: "slack", Config: map[string]string{}})
	svc.CreateAlertChannel(ctx, projectID, domain.DeliveryChannelInput{Name: "Ch2", Type: "email", Config: map[string]string{}})

	t.Run("returns channels for project", func(t *testing.T) {
		channels, err := svc.ListAlertChannels(ctx, projectID)
		require.NoError(t, err)
		assert.Len(t, channels, 2)
	})

	t.Run("empty project returns empty slice", func(t *testing.T) {
		channels, err := svc.ListAlertChannels(ctx, uuid.New())
		require.NoError(t, err)
		assert.Empty(t, channels)
		assert.NotNil(t, channels)
	})
}

func TestRCAService_TestAlertChannel(t *testing.T) {
	svc := NewRCAService(zap.NewNop())
	ctx := context.Background()
	projectID := uuid.New()

	t.Run("returns test result", func(t *testing.T) {
		ch, _ := svc.CreateAlertChannel(ctx, projectID, domain.DeliveryChannelInput{
			Name: "Test Channel", Type: "webhook", Config: map[string]string{"url": "https://example.com"},
		})
		assert.Equal(t, "untested", ch.TestStatus)

		tested, err := svc.TestAlertChannel(ctx, ch.ID)
		require.NoError(t, err)
		assert.Equal(t, "success", tested.TestStatus)
	})

	t.Run("not found", func(t *testing.T) {
		_, err := svc.TestAlertChannel(ctx, uuid.New())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "alert channel not found")
	})
}

func TestRCAService_CreateCorrelationRule(t *testing.T) {
	svc := NewRCAService(zap.NewNop())
	ctx := context.Background()
	projectID := uuid.New()

	t.Run("valid rule", func(t *testing.T) {
		input := domain.CorrelationRuleInput{
			Name:         "Latency + Error Correlation",
			AnomalyTypes: []string{"latency_spike", "error_burst"},
		}
		rule, err := svc.CreateCorrelationRule(ctx, projectID, input)
		require.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, rule.ID)
		assert.Equal(t, projectID, rule.ProjectID)
		assert.Equal(t, "Latency + Error Correlation", rule.Name)
		assert.Equal(t, []string{"latency_spike", "error_burst"}, rule.AnomalyTypes)
		assert.Equal(t, 30, rule.WindowMinutes)
		assert.Equal(t, 0.7, rule.MinCorrelation)
		assert.Equal(t, "warning", rule.Severity)
		assert.True(t, rule.Enabled)
		assert.NotNil(t, rule.Channels)
	})

	t.Run("empty name returns error", func(t *testing.T) {
		input := domain.CorrelationRuleInput{
			AnomalyTypes: []string{"latency_spike"},
		}
		_, err := svc.CreateCorrelationRule(ctx, projectID, input)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "name must not be empty")
	})

	t.Run("empty anomaly types returns error", func(t *testing.T) {
		input := domain.CorrelationRuleInput{
			Name: "Bad Rule",
		}
		_, err := svc.CreateCorrelationRule(ctx, projectID, input)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "anomaly types must not be empty")
	})
}

func TestRCAService_ListCorrelationRules(t *testing.T) {
	svc := NewRCAService(zap.NewNop())
	ctx := context.Background()
	projectID := uuid.New()

	svc.CreateCorrelationRule(ctx, projectID, domain.CorrelationRuleInput{Name: "R1", AnomalyTypes: []string{"latency_spike"}})
	svc.CreateCorrelationRule(ctx, projectID, domain.CorrelationRuleInput{Name: "R2", AnomalyTypes: []string{"error_burst"}})

	t.Run("returns rules for project", func(t *testing.T) {
		rules, err := svc.ListCorrelationRules(ctx, projectID)
		require.NoError(t, err)
		assert.Len(t, rules, 2)
	})

	t.Run("empty project returns empty slice", func(t *testing.T) {
		rules, err := svc.ListCorrelationRules(ctx, uuid.New())
		require.NoError(t, err)
		assert.Empty(t, rules)
		assert.NotNil(t, rules)
	})
}

func TestRCAService_GetAlertDashboardStats(t *testing.T) {
	svc := NewRCAService(zap.NewNop())
	ctx := context.Background()
	projectID := uuid.New()

	// Create anomalies with different statuses and severities
	svc.DetectAnomalies(ctx, projectID, domain.CorrelatedAnomalyInput{AnomalyType: "latency_spike", Title: "Open", Severity: "critical"})
	a2, _ := svc.DetectAnomalies(ctx, projectID, domain.CorrelatedAnomalyInput{AnomalyType: "error_burst", Title: "Acked", Severity: "warning"})
	svc.AcknowledgeAnomaly(ctx, a2.ID)

	t.Run("returns valid stats", func(t *testing.T) {
		stats, err := svc.GetAlertDashboardStats(ctx, projectID)
		require.NoError(t, err)
		assert.Equal(t, 1, stats.OpenAnomalies)
		assert.Equal(t, 1, stats.CriticalAlerts)
		assert.Equal(t, 0, stats.ActiveInvestigations)
		assert.Greater(t, stats.MTTR, 0.0)
		assert.NotEmpty(t, stats.AnomalyTrend)
	})

	t.Run("empty project returns zero stats", func(t *testing.T) {
		stats, err := svc.GetAlertDashboardStats(ctx, uuid.New())
		require.NoError(t, err)
		assert.Equal(t, 0, stats.OpenAnomalies)
		assert.Equal(t, 0, stats.CriticalAlerts)
		assert.Equal(t, 0, stats.ActiveInvestigations)
	})
}

func TestRCAService_CreateInvestigation(t *testing.T) {
	svc := NewRCAService(zap.NewNop())
	ctx := context.Background()
	projectID := uuid.New()
	investigatorID := uuid.New()

	t.Run("valid investigation", func(t *testing.T) {
		anomaly, _ := svc.DetectAnomalies(ctx, projectID, domain.CorrelatedAnomalyInput{AnomalyType: "latency_spike", Title: "Anomaly"})
		inv, err := svc.CreateInvestigation(ctx, projectID, anomaly.ID, "Investigate latency", investigatorID)
		require.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, inv.ID)
		assert.Equal(t, projectID, inv.ProjectID)
		assert.Equal(t, anomaly.ID, inv.AnomalyID)
		assert.Equal(t, "Investigate latency", inv.Title)
		assert.Equal(t, "open", inv.Status)
		assert.Equal(t, investigatorID, inv.InvestigatorID)
		assert.NotNil(t, inv.Findings)
		assert.Empty(t, inv.Findings)
		assert.Len(t, inv.Timeline, 1)
		assert.Equal(t, "created", inv.Timeline[0].Action)
		assert.False(t, inv.CreatedAt.IsZero())
	})

	t.Run("missing title returns error", func(t *testing.T) {
		anomaly, _ := svc.DetectAnomalies(ctx, projectID, domain.CorrelatedAnomalyInput{AnomalyType: "error_burst", Title: "Anomaly2"})
		_, err := svc.CreateInvestigation(ctx, projectID, anomaly.ID, "", investigatorID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "title must not be empty")
	})

	t.Run("anomaly not found", func(t *testing.T) {
		_, err := svc.CreateInvestigation(ctx, projectID, uuid.New(), "Test", investigatorID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "anomaly not found")
	})
}

func TestRCAService_UpdateInvestigation(t *testing.T) {
	svc := NewRCAService(zap.NewNop())
	ctx := context.Background()
	projectID := uuid.New()
	investigatorID := uuid.New()

	t.Run("status transition", func(t *testing.T) {
		anomaly, _ := svc.DetectAnomalies(ctx, projectID, domain.CorrelatedAnomalyInput{AnomalyType: "latency_spike", Title: "Anomaly"})
		inv, _ := svc.CreateInvestigation(ctx, projectID, anomaly.ID, "Test investigation", investigatorID)

		updated, err := svc.UpdateInvestigation(ctx, inv.ID, "investigating", "", "")
		require.NoError(t, err)
		assert.Equal(t, "investigating", updated.Status)
		assert.Len(t, updated.Timeline, 2)
	})

	t.Run("adding findings via root cause and resolution", func(t *testing.T) {
		anomaly, _ := svc.DetectAnomalies(ctx, projectID, domain.CorrelatedAnomalyInput{AnomalyType: "error_burst", Title: "Anomaly2"})
		inv, _ := svc.CreateInvestigation(ctx, projectID, anomaly.ID, "Another investigation", investigatorID)

		updated, err := svc.UpdateInvestigation(ctx, inv.ID, "resolved", "Database connection pool exhaustion", "Increased pool size to 50")
		require.NoError(t, err)
		assert.Equal(t, "resolved", updated.Status)
		assert.Equal(t, "Database connection pool exhaustion", updated.RootCause)
		assert.Equal(t, "Increased pool size to 50", updated.Resolution)
		assert.True(t, updated.UpdatedAt.After(inv.CreatedAt) || updated.UpdatedAt.Equal(inv.CreatedAt))
	})

	t.Run("not found", func(t *testing.T) {
		_, err := svc.UpdateInvestigation(ctx, uuid.New(), "investigating", "", "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "investigation not found")
	})
}

func TestRCAService_ListInvestigations(t *testing.T) {
	svc := NewRCAService(zap.NewNop())
	ctx := context.Background()
	projectID := uuid.New()
	investigatorID := uuid.New()

	anomaly, _ := svc.DetectAnomalies(ctx, projectID, domain.CorrelatedAnomalyInput{AnomalyType: "latency_spike", Title: "Anomaly"})
	svc.CreateInvestigation(ctx, projectID, anomaly.ID, "Inv 1", investigatorID)
	svc.CreateInvestigation(ctx, projectID, anomaly.ID, "Inv 2", investigatorID)

	t.Run("returns investigations for project", func(t *testing.T) {
		investigations, err := svc.ListInvestigations(ctx, projectID)
		require.NoError(t, err)
		assert.Len(t, investigations, 2)
	})

	t.Run("empty project returns empty slice", func(t *testing.T) {
		investigations, err := svc.ListInvestigations(ctx, uuid.New())
		require.NoError(t, err)
		assert.Empty(t, investigations)
		assert.NotNil(t, investigations)
	})
}
