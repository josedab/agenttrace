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

func TestNewWebhookOrchestrationService(t *testing.T) {
	svc := NewWebhookOrchestrationService(zap.NewNop())
	assert.NotNil(t, svc)
}

func TestWebhookOrchestrationService_RuleCRUD(t *testing.T) {
	svc := NewWebhookOrchestrationService(zap.NewNop())
	ctx := context.Background()
	projectID := uuid.New()

	t.Run("creates rule", func(t *testing.T) {
		input := &domain.WebhookRuleInput{
			Name:         "cost-alert",
			Trigger:      domain.TriggerCostExceeded,
			Action:       domain.ActionSlack,
			ActionConfig: map[string]string{"webhookUrl": "https://hooks.slack.com/test"},
			Condition:    domain.RuleCondition{Threshold: floatPtr(10.0)},
		}
		rule, err := svc.CreateRule(ctx, projectID, input)
		require.NoError(t, err)
		assert.Equal(t, "cost-alert", rule.Name)
		assert.True(t, rule.Enabled)
		assert.Equal(t, 0, rule.FireCount)
	})

	t.Run("lists rules for project", func(t *testing.T) {
		rules := svc.ListRules(ctx, projectID)
		assert.GreaterOrEqual(t, len(rules), 1)
	})

	t.Run("deletes rule", func(t *testing.T) {
		rules := svc.ListRules(ctx, projectID)
		require.Greater(t, len(rules), 0)

		err := svc.DeleteRule(ctx, rules[0].ID)
		require.NoError(t, err)

		remaining := svc.ListRules(ctx, projectID)
		assert.Less(t, len(remaining), len(rules))
	})
}

func TestWebhookOrchestrationService_Templates(t *testing.T) {
	svc := NewWebhookOrchestrationService(zap.NewNop())
	ctx := context.Background()

	templates := svc.GetTemplates(ctx)
	assert.Greater(t, len(templates), 0)

	for _, tmpl := range templates {
		assert.NotEmpty(t, tmpl.Name)
		assert.NotEmpty(t, tmpl.Description)
	}
}

func TestWebhookOrchestrationService_TestRule(t *testing.T) {
	svc := NewWebhookOrchestrationService(zap.NewNop())
	ctx := context.Background()
	projectID := uuid.New()

	input := &domain.WebhookRuleInput{
		Name: "test-rule", Trigger: domain.TriggerErrorDetected, Action: domain.ActionSlack,
		ActionConfig: map[string]string{"webhookUrl": "https://test.com"},
	}
	rule, _ := svc.CreateRule(ctx, projectID, input)

	delivery, err := svc.TestRule(ctx, rule.ID)
	require.NoError(t, err)
	assert.NotNil(t, delivery)
	assert.Equal(t, "success", delivery.Status)
}

