package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"agenttrace/internal/domain"
	"agenttrace/internal/repository/postgres"
)

// MockAuditRepository is a mock implementation of the Audit repository
type MockAuditRepository struct {
	mock.Mock
}

func (m *MockAuditRepository) CreateAuditLog(ctx context.Context, input *domain.AuditLogInput) (*domain.AuditLog, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.AuditLog), args.Error(1)
}

func (m *MockAuditRepository) GetAuditLog(ctx context.Context, orgID, logID uuid.UUID) (*domain.AuditLog, error) {
	args := m.Called(ctx, orgID, logID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.AuditLog), args.Error(1)
}

func (m *MockAuditRepository) ListAuditLogs(ctx context.Context, filter *domain.AuditLogFilter) (*domain.AuditLogList, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.AuditLogList), args.Error(1)
}

func (m *MockAuditRepository) GetAuditSummary(ctx context.Context, orgID uuid.UUID, period string) (*domain.AuditSummary, error) {
	args := m.Called(ctx, orgID, period)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.AuditSummary), args.Error(1)
}

func (m *MockAuditRepository) GetRetentionPolicy(ctx context.Context, orgID uuid.UUID) (*domain.AuditRetentionPolicy, error) {
	args := m.Called(ctx, orgID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.AuditRetentionPolicy), args.Error(1)
}

func (m *MockAuditRepository) UpsertRetentionPolicy(ctx context.Context, policy *domain.AuditRetentionPolicy) error {
	args := m.Called(ctx, policy)
	return args.Error(0)
}

func (m *MockAuditRepository) ApplyRetentionPolicy(ctx context.Context, orgID uuid.UUID) (int64, error) {
	args := m.Called(ctx, orgID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockAuditRepository) CreateExportJob(ctx context.Context, orgID uuid.UUID, requestedBy *uuid.UUID, filter *domain.AuditLogFilter, format string, compress bool) (*postgres.AuditExportJob, error) {
	args := m.Called(ctx, orgID, requestedBy, filter, format, compress)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*postgres.AuditExportJob), args.Error(1)
}

func (m *MockAuditRepository) GetExportJob(ctx context.Context, jobID uuid.UUID) (*postgres.AuditExportJob, error) {
	args := m.Called(ctx, jobID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*postgres.AuditExportJob), args.Error(1)
}

func (m *MockAuditRepository) ListExportJobs(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]postgres.AuditExportJob, error) {
	args := m.Called(ctx, orgID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]postgres.AuditExportJob), args.Error(1)
}

func TestAuditService_Log(t *testing.T) {
	t.Run("creates audit log entry", func(t *testing.T) {
		auditRepo := new(MockAuditRepository)

		orgID := uuid.New()
		actorID := uuid.New()
		resourceID := uuid.New()

		input := &domain.AuditLogInput{
			OrganizationID: orgID,
			ActorID:        &actorID,
			ActorEmail:     "admin@example.com",
			ActorType:      "user",
			Action:         domain.AuditActionUserCreated,
			ResourceType:   domain.AuditResourceUser,
			ResourceID:     &resourceID,
			ResourceName:   "newuser@example.com",
			Description:    "User newuser@example.com was created",
		}

		expectedLog := &domain.AuditLog{
			ID:             uuid.New(),
			OrganizationID: orgID,
			ActorID:        &actorID,
			ActorEmail:     "admin@example.com",
			ActorType:      "user",
			Action:         domain.AuditActionUserCreated,
			ResourceType:   domain.AuditResourceUser,
			ResourceID:     &resourceID,
			ResourceName:   "newuser@example.com",
			Description:    "User newuser@example.com was created",
			CreatedAt:      time.Now(),
		}

		auditRepo.On("CreateAuditLog", mock.Anything, input).Return(expectedLog, nil)

		result, err := auditRepo.CreateAuditLog(context.Background(), input)

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, domain.AuditActionUserCreated, result.Action)
		assert.Equal(t, "admin@example.com", result.ActorEmail)
	})
}

func TestAuditService_LogAction(t *testing.T) {
	t.Run("logs action with minimal parameters", func(t *testing.T) {
		auditRepo := new(MockAuditRepository)

		orgID := uuid.New()
		actorID := uuid.New()

		auditRepo.On("CreateAuditLog", mock.Anything, mock.AnythingOfType("*domain.AuditLogInput")).Return(&domain.AuditLog{}, nil)

		input := &domain.AuditLogInput{
			OrganizationID: orgID,
			ActorID:        &actorID,
			ActorEmail:     "user@example.com",
			ActorType:      "user",
			Action:         domain.AuditActionLogin,
			ResourceType:   domain.AuditResourceUser,
			ResourceID:     &actorID,
			ResourceName:   "user@example.com",
			Description:    "User user@example.com logged in",
		}

		_, err := auditRepo.CreateAuditLog(context.Background(), input)
		assert.NoError(t, err)
	})
}

func TestAuditService_LogWithContext(t *testing.T) {
	t.Run("logs action with request context", func(t *testing.T) {
		auditRepo := new(MockAuditRepository)

		orgID := uuid.New()
		actorID := uuid.New()

		auditRepo.On("CreateAuditLog", mock.Anything, mock.MatchedBy(func(input *domain.AuditLogInput) bool {
			return input.IPAddress == "192.168.1.1" &&
				input.UserAgent == "Mozilla/5.0" &&
				input.RequestID == "req-123"
		})).Return(&domain.AuditLog{}, nil)

		input := &domain.AuditLogInput{
			OrganizationID: orgID,
			ActorID:        &actorID,
			ActorEmail:     "user@example.com",
			ActorType:      "user",
			Action:         domain.AuditActionLogin,
			ResourceType:   domain.AuditResourceUser,
			IPAddress:      "192.168.1.1",
			UserAgent:      "Mozilla/5.0",
			RequestID:      "req-123",
		}

		_, err := auditRepo.CreateAuditLog(context.Background(), input)
		assert.NoError(t, err)
	})
}

func TestAuditService_LogWithChanges(t *testing.T) {
	t.Run("logs action with before/after changes", func(t *testing.T) {
		auditRepo := new(MockAuditRepository)

		orgID := uuid.New()
		actorID := uuid.New()
		targetID := uuid.New()

		before := map[string]any{"role": "member"}
		after := map[string]any{"role": "admin"}

		auditRepo.On("CreateAuditLog", mock.Anything, mock.MatchedBy(func(input *domain.AuditLogInput) bool {
			return input.Changes != nil &&
				input.Changes.Before["role"] == "member" &&
				input.Changes.After["role"] == "admin"
		})).Return(&domain.AuditLog{}, nil)

		input := &domain.AuditLogInput{
			OrganizationID: orgID,
			ActorID:        &actorID,
			ActorEmail:     "admin@example.com",
			ActorType:      "user",
			Action:         domain.AuditActionUserRoleChanged,
			ResourceType:   domain.AuditResourceUser,
			ResourceID:     &targetID,
			Changes: &domain.AuditChanges{
				Before: before,
				After:  after,
			},
		}

		_, err := auditRepo.CreateAuditLog(context.Background(), input)
		assert.NoError(t, err)
	})
}

func TestAuditService_GetAuditLog(t *testing.T) {
	t.Run("retrieves audit log by ID", func(t *testing.T) {
		auditRepo := new(MockAuditRepository)

		orgID := uuid.New()
		logID := uuid.New()

		expectedLog := &domain.AuditLog{
			ID:             logID,
			OrganizationID: orgID,
			Action:         domain.AuditActionLogin,
			CreatedAt:      time.Now(),
		}

		auditRepo.On("GetAuditLog", mock.Anything, orgID, logID).Return(expectedLog, nil)

		result, err := auditRepo.GetAuditLog(context.Background(), orgID, logID)

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, logID, result.ID)
	})

	t.Run("returns nil for non-existent log", func(t *testing.T) {
		auditRepo := new(MockAuditRepository)

		orgID := uuid.New()
		logID := uuid.New()

		auditRepo.On("GetAuditLog", mock.Anything, orgID, logID).Return(nil, nil)

		result, err := auditRepo.GetAuditLog(context.Background(), orgID, logID)

		assert.NoError(t, err)
		assert.Nil(t, result)
	})
}

func TestAuditService_ListAuditLogs(t *testing.T) {
	t.Run("lists audit logs with filters", func(t *testing.T) {
		auditRepo := new(MockAuditRepository)

		orgID := uuid.New()
		filter := &domain.AuditLogFilter{
			OrganizationID: &orgID,
			Actions:        []domain.AuditAction{domain.AuditActionLogin, domain.AuditActionLogout},
			Limit:          50,
			Offset:         0,
		}

		expectedResult := &domain.AuditLogList{
			Data: []domain.AuditLog{
				{ID: uuid.New(), Action: domain.AuditActionLogin},
				{ID: uuid.New(), Action: domain.AuditActionLogout},
			},
			TotalCount: 2,
			HasMore:    false,
		}

		auditRepo.On("ListAuditLogs", mock.Anything, filter).Return(expectedResult, nil)

		result, err := auditRepo.ListAuditLogs(context.Background(), filter)

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Data, 2)
		assert.Equal(t, int64(2), result.TotalCount)
	})

	t.Run("filters by time range", func(t *testing.T) {
		auditRepo := new(MockAuditRepository)

		orgID := uuid.New()
		startTime := time.Now().Add(-24 * time.Hour)
		endTime := time.Now()

		filter := &domain.AuditLogFilter{
			OrganizationID: &orgID,
			StartTime:      &startTime,
			EndTime:        &endTime,
		}

		auditRepo.On("ListAuditLogs", mock.Anything, filter).Return(&domain.AuditLogList{Data: []domain.AuditLog{}}, nil)

		result, err := auditRepo.ListAuditLogs(context.Background(), filter)

		require.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("filters by actor", func(t *testing.T) {
		auditRepo := new(MockAuditRepository)

		orgID := uuid.New()
		actorID := uuid.New()

		filter := &domain.AuditLogFilter{
			OrganizationID: &orgID,
			ActorID:        &actorID,
		}

		auditRepo.On("ListAuditLogs", mock.Anything, filter).Return(&domain.AuditLogList{Data: []domain.AuditLog{}}, nil)

		result, err := auditRepo.ListAuditLogs(context.Background(), filter)

		require.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("filters by resource", func(t *testing.T) {
		auditRepo := new(MockAuditRepository)

		orgID := uuid.New()
		resourceType := domain.AuditResourceProject
		resourceID := uuid.New()

		filter := &domain.AuditLogFilter{
			OrganizationID: &orgID,
			ResourceType:   &resourceType,
			ResourceID:     &resourceID,
		}

		auditRepo.On("ListAuditLogs", mock.Anything, filter).Return(&domain.AuditLogList{Data: []domain.AuditLog{}}, nil)

		result, err := auditRepo.ListAuditLogs(context.Background(), filter)

		require.NoError(t, err)
		assert.NotNil(t, result)
	})
}

func TestAuditService_GetAuditSummary(t *testing.T) {
	t.Run("returns summary for 24h period", func(t *testing.T) {
		auditRepo := new(MockAuditRepository)

		orgID := uuid.New()
		expectedSummary := &domain.AuditSummary{
			TotalEvents: 150,
			ActionCounts: map[domain.AuditAction]int64{
				domain.AuditActionLogin:  100,
				domain.AuditActionLogout: 50,
			},
			TopActors: []domain.ActorCount{
				{ActorEmail: "admin@example.com", Count: 80},
				{ActorEmail: "user@example.com", Count: 70},
			},
		}

		auditRepo.On("GetAuditSummary", mock.Anything, orgID, "24h").Return(expectedSummary, nil)

		result, err := auditRepo.GetAuditSummary(context.Background(), orgID, "24h")

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, int64(150), result.TotalEvents)
		assert.Equal(t, int64(100), result.ActionCounts[domain.AuditActionLogin])
	})

	t.Run("returns summary for 7d period", func(t *testing.T) {
		auditRepo := new(MockAuditRepository)

		orgID := uuid.New()
		auditRepo.On("GetAuditSummary", mock.Anything, orgID, "7d").Return(&domain.AuditSummary{TotalEvents: 1000}, nil)

		result, err := auditRepo.GetAuditSummary(context.Background(), orgID, "7d")

		require.NoError(t, err)
		assert.Equal(t, int64(1000), result.TotalEvents)
	})
}

func TestAuditService_RetentionPolicy(t *testing.T) {
	t.Run("returns default policy when none exists", func(t *testing.T) {
		auditRepo := new(MockAuditRepository)

		orgID := uuid.New()
		auditRepo.On("GetRetentionPolicy", mock.Anything, orgID).Return(nil, nil)

		result, err := auditRepo.GetRetentionPolicy(context.Background(), orgID)

		require.NoError(t, err)
		// Service returns default policy (365 days) when nil
		assert.Nil(t, result)
	})

	t.Run("returns custom retention policy", func(t *testing.T) {
		auditRepo := new(MockAuditRepository)

		orgID := uuid.New()
		policy := &domain.AuditRetentionPolicy{
			OrganizationID: orgID,
			RetentionDays:  90,
			Enabled:        true,
		}

		auditRepo.On("GetRetentionPolicy", mock.Anything, orgID).Return(policy, nil)

		result, err := auditRepo.GetRetentionPolicy(context.Background(), orgID)

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 90, result.RetentionDays)
		assert.True(t, result.Enabled)
	})

	t.Run("sets retention policy", func(t *testing.T) {
		auditRepo := new(MockAuditRepository)

		orgID := uuid.New()
		policy := &domain.AuditRetentionPolicy{
			OrganizationID: orgID,
			RetentionDays:  180,
			Enabled:        true,
		}

		auditRepo.On("UpsertRetentionPolicy", mock.Anything, policy).Return(nil)

		err := auditRepo.UpsertRetentionPolicy(context.Background(), policy)
		assert.NoError(t, err)
	})

	t.Run("applies retention policy", func(t *testing.T) {
		auditRepo := new(MockAuditRepository)

		orgID := uuid.New()
		auditRepo.On("ApplyRetentionPolicy", mock.Anything, orgID).Return(int64(500), nil)

		deletedCount, err := auditRepo.ApplyRetentionPolicy(context.Background(), orgID)

		require.NoError(t, err)
		assert.Equal(t, int64(500), deletedCount)
	})
}

func TestAuditService_ExportJobs(t *testing.T) {
	t.Run("creates export job", func(t *testing.T) {
		auditRepo := new(MockAuditRepository)

		orgID := uuid.New()
		requestedBy := uuid.New()
		filter := &domain.AuditLogFilter{OrganizationID: &orgID}

		expectedJob := &postgres.AuditExportJob{
			ID:          uuid.New(),
			Status:      "pending",
			Format:      "json",
			Compress:    true,
			RequestedBy: &requestedBy,
		}

		auditRepo.On("CreateExportJob", mock.Anything, orgID, &requestedBy, filter, "json", true).Return(expectedJob, nil)

		result, err := auditRepo.CreateExportJob(context.Background(), orgID, &requestedBy, filter, "json", true)

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "pending", result.Status)
		assert.Equal(t, "json", result.Format)
	})

	t.Run("gets export job by ID", func(t *testing.T) {
		auditRepo := new(MockAuditRepository)

		jobID := uuid.New()
		expectedJob := &postgres.AuditExportJob{
			ID:     jobID,
			Status: "completed",
		}

		auditRepo.On("GetExportJob", mock.Anything, jobID).Return(expectedJob, nil)

		result, err := auditRepo.GetExportJob(context.Background(), jobID)

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "completed", result.Status)
	})

	t.Run("lists export jobs", func(t *testing.T) {
		auditRepo := new(MockAuditRepository)

		orgID := uuid.New()
		jobs := []postgres.AuditExportJob{
			{ID: uuid.New(), Status: "completed"},
			{ID: uuid.New(), Status: "pending"},
		}

		auditRepo.On("ListExportJobs", mock.Anything, orgID, 10, 0).Return(jobs, nil)

		result, err := auditRepo.ListExportJobs(context.Background(), orgID, 10, 0)

		require.NoError(t, err)
		assert.Len(t, result, 2)
	})
}

func TestAuditService_ConvenienceMethods(t *testing.T) {
	t.Run("LogLogin creates correct audit entry", func(t *testing.T) {
		auditRepo := new(MockAuditRepository)

		auditRepo.On("CreateAuditLog", mock.Anything, mock.MatchedBy(func(input *domain.AuditLogInput) bool {
			return input.Action == domain.AuditActionLogin &&
				input.ActorType == "user" &&
				input.IPAddress == "192.168.1.1"
		})).Return(&domain.AuditLog{}, nil)

		orgID := uuid.New()
		userID := uuid.New()

		input := &domain.AuditLogInput{
			OrganizationID: orgID,
			ActorID:        &userID,
			ActorEmail:     "user@example.com",
			ActorType:      "user",
			Action:         domain.AuditActionLogin,
			ResourceType:   domain.AuditResourceUser,
			ResourceID:     &userID,
			IPAddress:      "192.168.1.1",
			UserAgent:      "Mozilla/5.0",
		}

		_, err := auditRepo.CreateAuditLog(context.Background(), input)
		assert.NoError(t, err)
	})

	t.Run("LogLoginFailed creates correct audit entry", func(t *testing.T) {
		auditRepo := new(MockAuditRepository)

		auditRepo.On("CreateAuditLog", mock.Anything, mock.MatchedBy(func(input *domain.AuditLogInput) bool {
			return input.Action == domain.AuditActionLoginFailed &&
				input.ActorID == nil // No user ID for failed login
		})).Return(&domain.AuditLog{}, nil)

		orgID := uuid.New()

		input := &domain.AuditLogInput{
			OrganizationID: orgID,
			ActorID:        nil,
			ActorEmail:     "user@example.com",
			ActorType:      "user",
			Action:         domain.AuditActionLoginFailed,
			ResourceType:   domain.AuditResourceUser,
			IPAddress:      "192.168.1.1",
		}

		_, err := auditRepo.CreateAuditLog(context.Background(), input)
		assert.NoError(t, err)
	})

	t.Run("LogSSOLogin includes provider metadata", func(t *testing.T) {
		auditRepo := new(MockAuditRepository)

		auditRepo.On("CreateAuditLog", mock.Anything, mock.MatchedBy(func(input *domain.AuditLogInput) bool {
			return input.Action == domain.AuditActionSSOLogin &&
				input.Metadata != nil &&
				input.Metadata["provider"] == "oidc"
		})).Return(&domain.AuditLog{}, nil)

		orgID := uuid.New()
		userID := uuid.New()

		input := &domain.AuditLogInput{
			OrganizationID: orgID,
			ActorID:        &userID,
			ActorEmail:     "user@example.com",
			ActorType:      "user",
			Action:         domain.AuditActionSSOLogin,
			ResourceType:   domain.AuditResourceUser,
			Metadata:       map[string]any{"provider": "oidc"},
		}

		_, err := auditRepo.CreateAuditLog(context.Background(), input)
		assert.NoError(t, err)
	})

	t.Run("LogAPIKeyCreated creates correct audit entry", func(t *testing.T) {
		auditRepo := new(MockAuditRepository)

		auditRepo.On("CreateAuditLog", mock.Anything, mock.MatchedBy(func(input *domain.AuditLogInput) bool {
			return input.Action == domain.AuditActionAPIKeyCreated &&
				input.ResourceType == domain.AuditResourceAPIKey
		})).Return(&domain.AuditLog{}, nil)

		orgID := uuid.New()
		actorID := uuid.New()
		keyID := uuid.New()

		input := &domain.AuditLogInput{
			OrganizationID: orgID,
			ActorID:        &actorID,
			ActorEmail:     "admin@example.com",
			ActorType:      "user",
			Action:         domain.AuditActionAPIKeyCreated,
			ResourceType:   domain.AuditResourceAPIKey,
			ResourceID:     &keyID,
			ResourceName:   "Production API Key",
		}

		_, err := auditRepo.CreateAuditLog(context.Background(), input)
		assert.NoError(t, err)
	})

	t.Run("LogUserRoleChanged includes before/after changes", func(t *testing.T) {
		auditRepo := new(MockAuditRepository)

		auditRepo.On("CreateAuditLog", mock.Anything, mock.MatchedBy(func(input *domain.AuditLogInput) bool {
			return input.Action == domain.AuditActionUserRoleChanged &&
				input.Changes != nil &&
				input.Changes.Before["role"] == "member" &&
				input.Changes.After["role"] == "admin"
		})).Return(&domain.AuditLog{}, nil)

		orgID := uuid.New()
		actorID := uuid.New()
		targetID := uuid.New()

		input := &domain.AuditLogInput{
			OrganizationID: orgID,
			ActorID:        &actorID,
			ActorEmail:     "admin@example.com",
			ActorType:      "user",
			Action:         domain.AuditActionUserRoleChanged,
			ResourceType:   domain.AuditResourceUser,
			ResourceID:     &targetID,
			Changes: &domain.AuditChanges{
				Before: map[string]any{"role": "member"},
				After:  map[string]any{"role": "admin"},
			},
		}

		_, err := auditRepo.CreateAuditLog(context.Background(), input)
		assert.NoError(t, err)
	})

	t.Run("LogProjectCreated creates correct audit entry", func(t *testing.T) {
		auditRepo := new(MockAuditRepository)

		auditRepo.On("CreateAuditLog", mock.Anything, mock.MatchedBy(func(input *domain.AuditLogInput) bool {
			return input.Action == domain.AuditActionProjectCreated &&
				input.ResourceType == domain.AuditResourceProject &&
				input.ResourceName == "My Project"
		})).Return(&domain.AuditLog{}, nil)

		orgID := uuid.New()
		actorID := uuid.New()
		projectID := uuid.New()

		input := &domain.AuditLogInput{
			OrganizationID: orgID,
			ActorID:        &actorID,
			ActorEmail:     "admin@example.com",
			ActorType:      "user",
			Action:         domain.AuditActionProjectCreated,
			ResourceType:   domain.AuditResourceProject,
			ResourceID:     &projectID,
			ResourceName:   "My Project",
		}

		_, err := auditRepo.CreateAuditLog(context.Background(), input)
		assert.NoError(t, err)
	})

	t.Run("LogSSOConfigured creates correct audit entry", func(t *testing.T) {
		auditRepo := new(MockAuditRepository)

		auditRepo.On("CreateAuditLog", mock.Anything, mock.MatchedBy(func(input *domain.AuditLogInput) bool {
			return input.Action == domain.AuditActionSSOConfigured &&
				input.ResourceType == domain.AuditResourceSSO &&
				input.Metadata["provider"] == "saml"
		})).Return(&domain.AuditLog{}, nil)

		orgID := uuid.New()
		actorID := uuid.New()

		input := &domain.AuditLogInput{
			OrganizationID: orgID,
			ActorID:        &actorID,
			ActorEmail:     "admin@example.com",
			ActorType:      "user",
			Action:         domain.AuditActionSSOConfigured,
			ResourceType:   domain.AuditResourceSSO,
			ResourceID:     &orgID,
			Metadata:       map[string]any{"provider": "saml"},
		}

		_, err := auditRepo.CreateAuditLog(context.Background(), input)
		assert.NoError(t, err)
	})

	t.Run("LogDataExported includes export metadata", func(t *testing.T) {
		auditRepo := new(MockAuditRepository)

		auditRepo.On("CreateAuditLog", mock.Anything, mock.MatchedBy(func(input *domain.AuditLogInput) bool {
			return input.Action == domain.AuditActionDataExported &&
				input.Metadata["recordCount"] == 1000
		})).Return(&domain.AuditLog{}, nil)

		orgID := uuid.New()
		actorID := uuid.New()

		input := &domain.AuditLogInput{
			OrganizationID: orgID,
			ActorID:        &actorID,
			ActorEmail:     "admin@example.com",
			ActorType:      "user",
			Action:         domain.AuditActionDataExported,
			ResourceType:   domain.AuditResourceTrace,
			Metadata:       map[string]any{"dataType": "traces", "recordCount": 1000},
		}

		_, err := auditRepo.CreateAuditLog(context.Background(), input)
		assert.NoError(t, err)
	})
}

func TestAuditService_SecurityEvents(t *testing.T) {
	t.Run("filters security-related events", func(t *testing.T) {
		auditRepo := new(MockAuditRepository)

		orgID := uuid.New()
		since := time.Now().Add(-24 * time.Hour)

		securityActions := []domain.AuditAction{
			domain.AuditActionLogin,
			domain.AuditActionLogout,
			domain.AuditActionLoginFailed,
			domain.AuditActionSSOLogin,
			domain.AuditActionAPIKeyUsed,
			domain.AuditActionAPIKeyCreated,
			domain.AuditActionAPIKeyRevoked,
			domain.AuditActionUserCreated,
			domain.AuditActionUserDeleted,
			domain.AuditActionUserRoleChanged,
			domain.AuditActionSSOConfigured,
			domain.AuditActionSSOEnabled,
			domain.AuditActionSSODisabled,
		}

		filter := &domain.AuditLogFilter{
			OrganizationID: &orgID,
			Actions:        securityActions,
			StartTime:      &since,
			Limit:          100,
		}

		auditRepo.On("ListAuditLogs", mock.Anything, filter).Return(&domain.AuditLogList{
			Data: []domain.AuditLog{
				{Action: domain.AuditActionLogin},
				{Action: domain.AuditActionLoginFailed},
			},
		}, nil)

		result, err := auditRepo.ListAuditLogs(context.Background(), filter)

		require.NoError(t, err)
		assert.Len(t, result.Data, 2)
	})
}

func TestAuditService_ActivityTimeline(t *testing.T) {
	t.Run("returns activity for resource", func(t *testing.T) {
		auditRepo := new(MockAuditRepository)

		orgID := uuid.New()
		resourceType := domain.AuditResourceProject
		resourceID := uuid.New()

		filter := &domain.AuditLogFilter{
			OrganizationID: &orgID,
			ResourceType:   &resourceType,
			ResourceID:     &resourceID,
			Limit:          10,
		}

		auditRepo.On("ListAuditLogs", mock.Anything, filter).Return(&domain.AuditLogList{
			Data: []domain.AuditLog{
				{Action: domain.AuditActionProjectCreated},
				{Action: domain.AuditActionProjectUpdated},
			},
		}, nil)

		result, err := auditRepo.ListAuditLogs(context.Background(), filter)

		require.NoError(t, err)
		assert.Len(t, result.Data, 2)
	})
}

func TestAuditActions(t *testing.T) {
	t.Run("all audit actions are defined", func(t *testing.T) {
		actions := []domain.AuditAction{
			domain.AuditActionLogin,
			domain.AuditActionLogout,
			domain.AuditActionLoginFailed,
			domain.AuditActionSSOLogin,
			domain.AuditActionAPIKeyUsed,
			domain.AuditActionAPIKeyCreated,
			domain.AuditActionAPIKeyRevoked,
			domain.AuditActionUserCreated,
			domain.AuditActionUserUpdated,
			domain.AuditActionUserDeleted,
			domain.AuditActionUserInvited,
			domain.AuditActionUserRoleChanged,
			domain.AuditActionProjectCreated,
			domain.AuditActionProjectUpdated,
			domain.AuditActionProjectDeleted,
			domain.AuditActionSSOConfigured,
			domain.AuditActionSSOEnabled,
			domain.AuditActionSSODisabled,
			domain.AuditActionDataExported,
			domain.AuditActionSettingsChanged,
		}

		for _, action := range actions {
			assert.NotEmpty(t, string(action))
		}
	})
}

func TestAuditResourceTypes(t *testing.T) {
	t.Run("all resource types are defined", func(t *testing.T) {
		resourceTypes := []domain.AuditResourceType{
			domain.AuditResourceUser,
			domain.AuditResourceProject,
			domain.AuditResourceAPIKey,
			domain.AuditResourceSSO,
			domain.AuditResourceSettings,
			domain.AuditResourceTrace,
			domain.AuditResourceDataset,
			domain.AuditResourcePrompt,
			domain.AuditResourceEvaluator,
		}

		for _, rt := range resourceTypes {
			assert.NotEmpty(t, string(rt))
		}
	})
}
