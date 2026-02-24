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

func TestNewComplianceReportService(t *testing.T) {
	svc := NewComplianceReportService(zap.NewNop())
	assert.NotNil(t, svc)
}

func TestComplianceReportService_Generate(t *testing.T) {
	svc := NewComplianceReportService(zap.NewNop())
	ctx := context.Background()
	projectID := uuid.New()

	t.Run("generates EU AI Act report", func(t *testing.T) {
		input := &domain.GenerateReportInput{Template: domain.ReportEUAIAct, Title: "Q1 Compliance"}
		report, err := svc.GenerateReport(ctx, projectID, input)
		require.NoError(t, err)
		assert.Equal(t, "ready", report.Status)
		assert.Equal(t, domain.ReportEUAIAct, report.Template)
		assert.Greater(t, len(report.Sections), 0)
		assert.Greater(t, report.Score, 0.0)
		assert.LessOrEqual(t, report.Score, 100.0)

		for _, section := range report.Sections {
			assert.NotEmpty(t, section.Title)
			assert.NotEmpty(t, section.Status)
		}
	})

	t.Run("generates SOC 2 report", func(t *testing.T) {
		input := &domain.GenerateReportInput{Template: domain.ReportSOC2}
		report, err := svc.GenerateReport(ctx, projectID, input)
		require.NoError(t, err)
		assert.Equal(t, domain.ReportSOC2, report.Template)
		assert.Greater(t, len(report.Sections), 0)
	})

	t.Run("lists reports", func(t *testing.T) {
		reports := svc.ListReports(ctx, projectID)
		assert.GreaterOrEqual(t, len(reports), 2)
	})

	t.Run("gets report by ID", func(t *testing.T) {
		reports := svc.ListReports(ctx, projectID)
		require.Greater(t, len(reports), 0)

		report, err := svc.GetReport(ctx, reports[0].ID)
		require.NoError(t, err)
		assert.NotNil(t, report)
	})
}

func TestComplianceReportService_Templates(t *testing.T) {
	svc := NewComplianceReportService(zap.NewNop())
	ctx := context.Background()

	templates := svc.GetTemplates(ctx)
	assert.GreaterOrEqual(t, len(templates), 2)
}
