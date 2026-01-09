package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"agenttrace/internal/domain"
	"agenttrace/internal/repository/postgres"
)

type AuditService struct {
	auditRepo *postgres.AuditRepository
}

func NewAuditService(auditRepo *postgres.AuditRepository) *AuditService {
	return &AuditService{
		auditRepo: auditRepo,
	}
}

// Log creates a new audit log entry
func (s *AuditService) Log(ctx context.Context, input *domain.AuditLogInput) (*domain.AuditLog, error) {
	return s.auditRepo.CreateAuditLog(ctx, input)
}

// LogAction is a convenience method for logging with minimal parameters
func (s *AuditService) LogAction(
	ctx context.Context,
	orgID uuid.UUID,
	actorID *uuid.UUID,
	actorEmail string,
	actorType string,
	action domain.AuditAction,
	resourceType domain.AuditResourceType,
	resourceID *uuid.UUID,
	resourceName string,
	description string,
) error {
	input := &domain.AuditLogInput{
		OrganizationID: orgID,
		ActorID:        actorID,
		ActorEmail:     actorEmail,
		ActorType:      actorType,
		Action:         action,
		ResourceType:   resourceType,
		ResourceID:     resourceID,
		ResourceName:   resourceName,
		Description:    description,
	}

	_, err := s.auditRepo.CreateAuditLog(ctx, input)
	return err
}

// LogWithContext logs an action with request context (IP, user agent, etc.)
func (s *AuditService) LogWithContext(
	ctx context.Context,
	orgID uuid.UUID,
	actorID *uuid.UUID,
	actorEmail string,
	actorType string,
	action domain.AuditAction,
	resourceType domain.AuditResourceType,
	resourceID *uuid.UUID,
	resourceName string,
	description string,
	ipAddress string,
	userAgent string,
	requestID string,
	sessionID string,
) error {
	input := &domain.AuditLogInput{
		OrganizationID: orgID,
		ActorID:        actorID,
		ActorEmail:     actorEmail,
		ActorType:      actorType,
		Action:         action,
		ResourceType:   resourceType,
		ResourceID:     resourceID,
		ResourceName:   resourceName,
		Description:    description,
		IPAddress:      ipAddress,
		UserAgent:      userAgent,
		RequestID:      requestID,
		SessionID:      sessionID,
	}

	_, err := s.auditRepo.CreateAuditLog(ctx, input)
	return err
}

// LogWithChanges logs an action that includes before/after changes
func (s *AuditService) LogWithChanges(
	ctx context.Context,
	orgID uuid.UUID,
	actorID *uuid.UUID,
	actorEmail string,
	actorType string,
	action domain.AuditAction,
	resourceType domain.AuditResourceType,
	resourceID *uuid.UUID,
	resourceName string,
	description string,
	before map[string]any,
	after map[string]any,
) error {
	input := &domain.AuditLogInput{
		OrganizationID: orgID,
		ActorID:        actorID,
		ActorEmail:     actorEmail,
		ActorType:      actorType,
		Action:         action,
		ResourceType:   resourceType,
		ResourceID:     resourceID,
		ResourceName:   resourceName,
		Description:    description,
		Changes: &domain.AuditChanges{
			Before: before,
			After:  after,
		},
	}

	_, err := s.auditRepo.CreateAuditLog(ctx, input)
	return err
}

// GetAuditLog retrieves a single audit log entry
func (s *AuditService) GetAuditLog(ctx context.Context, orgID, logID uuid.UUID) (*domain.AuditLog, error) {
	return s.auditRepo.GetAuditLog(ctx, orgID, logID)
}

// ListAuditLogs retrieves audit logs with filtering
func (s *AuditService) ListAuditLogs(ctx context.Context, filter *domain.AuditLogFilter) (*domain.AuditLogList, error) {
	return s.auditRepo.ListAuditLogs(ctx, filter)
}

// GetAuditSummary returns aggregated audit statistics
func (s *AuditService) GetAuditSummary(ctx context.Context, orgID uuid.UUID, period string) (*domain.AuditSummary, error) {
	return s.auditRepo.GetAuditSummary(ctx, orgID, period)
}

// Retention Policy methods

func (s *AuditService) GetRetentionPolicy(ctx context.Context, orgID uuid.UUID) (*domain.AuditRetentionPolicy, error) {
	policy, err := s.auditRepo.GetRetentionPolicy(ctx, orgID)
	if err != nil {
		return nil, err
	}

	// Return default policy if none exists
	if policy == nil {
		return &domain.AuditRetentionPolicy{
			OrganizationID: orgID,
			RetentionDays:  365, // Default to 1 year
			Enabled:        true,
		}, nil
	}

	return policy, nil
}

func (s *AuditService) SetRetentionPolicy(ctx context.Context, orgID uuid.UUID, retentionDays int, enabled bool) (*domain.AuditRetentionPolicy, error) {
	policy := &domain.AuditRetentionPolicy{
		OrganizationID: orgID,
		RetentionDays:  retentionDays,
		Enabled:        enabled,
	}

	if err := s.auditRepo.UpsertRetentionPolicy(ctx, policy); err != nil {
		return nil, err
	}

	return policy, nil
}

func (s *AuditService) ApplyRetentionPolicy(ctx context.Context, orgID uuid.UUID) (int64, error) {
	return s.auditRepo.ApplyRetentionPolicy(ctx, orgID)
}

// Export methods

func (s *AuditService) CreateExportJob(ctx context.Context, orgID uuid.UUID, requestedBy *uuid.UUID, filter *domain.AuditLogFilter, format string, compress bool) (*postgres.AuditExportJob, error) {
	return s.auditRepo.CreateExportJob(ctx, orgID, requestedBy, filter, format, compress)
}

func (s *AuditService) GetExportJob(ctx context.Context, jobID uuid.UUID) (*postgres.AuditExportJob, error) {
	return s.auditRepo.GetExportJob(ctx, jobID)
}

func (s *AuditService) ListExportJobs(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]postgres.AuditExportJob, error) {
	return s.auditRepo.ListExportJobs(ctx, orgID, limit, offset)
}

// Convenience methods for common audit actions

func (s *AuditService) LogLogin(ctx context.Context, orgID uuid.UUID, userID uuid.UUID, email, ipAddress, userAgent string) error {
	return s.LogWithContext(ctx, orgID, &userID, email, "user", domain.AuditActionLogin,
		domain.AuditResourceUser, &userID, email, fmt.Sprintf("User %s logged in", email),
		ipAddress, userAgent, "", "")
}

func (s *AuditService) LogLoginFailed(ctx context.Context, orgID uuid.UUID, email, ipAddress, userAgent, reason string) error {
	return s.LogWithContext(ctx, orgID, nil, email, "user", domain.AuditActionLoginFailed,
		domain.AuditResourceUser, nil, email, fmt.Sprintf("Failed login attempt for %s: %s", email, reason),
		ipAddress, userAgent, "", "")
}

func (s *AuditService) LogLogout(ctx context.Context, orgID uuid.UUID, userID uuid.UUID, email string) error {
	return s.LogAction(ctx, orgID, &userID, email, "user", domain.AuditActionLogout,
		domain.AuditResourceUser, &userID, email, fmt.Sprintf("User %s logged out", email))
}

func (s *AuditService) LogSSOLogin(ctx context.Context, orgID uuid.UUID, userID uuid.UUID, email, provider, ipAddress, userAgent string) error {
	input := &domain.AuditLogInput{
		OrganizationID: orgID,
		ActorID:        &userID,
		ActorEmail:     email,
		ActorType:      "user",
		Action:         domain.AuditActionSSOLogin,
		ResourceType:   domain.AuditResourceUser,
		ResourceID:     &userID,
		ResourceName:   email,
		Description:    fmt.Sprintf("User %s logged in via SSO (%s)", email, provider),
		Metadata:       map[string]any{"provider": provider},
		IPAddress:      ipAddress,
		UserAgent:      userAgent,
	}
	_, err := s.auditRepo.CreateAuditLog(ctx, input)
	return err
}

func (s *AuditService) LogAPIKeyUsed(ctx context.Context, orgID uuid.UUID, keyID uuid.UUID, keyName string, ipAddress, userAgent string) error {
	return s.LogWithContext(ctx, orgID, nil, keyName, "api_key", domain.AuditActionAPIKeyUsed,
		domain.AuditResourceAPIKey, &keyID, keyName, fmt.Sprintf("API key '%s' was used", keyName),
		ipAddress, userAgent, "", "")
}

func (s *AuditService) LogAPIKeyCreated(ctx context.Context, orgID uuid.UUID, actorID uuid.UUID, actorEmail string, keyID uuid.UUID, keyName string) error {
	return s.LogAction(ctx, orgID, &actorID, actorEmail, "user", domain.AuditActionAPIKeyCreated,
		domain.AuditResourceAPIKey, &keyID, keyName, fmt.Sprintf("API key '%s' was created", keyName))
}

func (s *AuditService) LogAPIKeyRevoked(ctx context.Context, orgID uuid.UUID, actorID uuid.UUID, actorEmail string, keyID uuid.UUID, keyName string) error {
	return s.LogAction(ctx, orgID, &actorID, actorEmail, "user", domain.AuditActionAPIKeyRevoked,
		domain.AuditResourceAPIKey, &keyID, keyName, fmt.Sprintf("API key '%s' was revoked", keyName))
}

func (s *AuditService) LogUserCreated(ctx context.Context, orgID uuid.UUID, actorID uuid.UUID, actorEmail string, newUserID uuid.UUID, newUserEmail string) error {
	return s.LogAction(ctx, orgID, &actorID, actorEmail, "user", domain.AuditActionUserCreated,
		domain.AuditResourceUser, &newUserID, newUserEmail, fmt.Sprintf("User %s was created", newUserEmail))
}

func (s *AuditService) LogUserUpdated(ctx context.Context, orgID uuid.UUID, actorID uuid.UUID, actorEmail string, targetUserID uuid.UUID, targetEmail string, before, after map[string]any) error {
	return s.LogWithChanges(ctx, orgID, &actorID, actorEmail, "user", domain.AuditActionUserUpdated,
		domain.AuditResourceUser, &targetUserID, targetEmail, fmt.Sprintf("User %s was updated", targetEmail), before, after)
}

func (s *AuditService) LogUserDeleted(ctx context.Context, orgID uuid.UUID, actorID uuid.UUID, actorEmail string, deletedUserID uuid.UUID, deletedUserEmail string) error {
	return s.LogAction(ctx, orgID, &actorID, actorEmail, "user", domain.AuditActionUserDeleted,
		domain.AuditResourceUser, &deletedUserID, deletedUserEmail, fmt.Sprintf("User %s was deleted", deletedUserEmail))
}

func (s *AuditService) LogUserInvited(ctx context.Context, orgID uuid.UUID, actorID uuid.UUID, actorEmail string, invitedEmail string) error {
	return s.LogAction(ctx, orgID, &actorID, actorEmail, "user", domain.AuditActionUserInvited,
		domain.AuditResourceUser, nil, invitedEmail, fmt.Sprintf("User %s was invited to join", invitedEmail))
}

func (s *AuditService) LogUserRoleChanged(ctx context.Context, orgID uuid.UUID, actorID uuid.UUID, actorEmail string, targetUserID uuid.UUID, targetEmail, oldRole, newRole string) error {
	return s.LogWithChanges(ctx, orgID, &actorID, actorEmail, "user", domain.AuditActionUserRoleChanged,
		domain.AuditResourceUser, &targetUserID, targetEmail,
		fmt.Sprintf("User %s role changed from %s to %s", targetEmail, oldRole, newRole),
		map[string]any{"role": oldRole}, map[string]any{"role": newRole})
}

func (s *AuditService) LogProjectCreated(ctx context.Context, orgID uuid.UUID, actorID uuid.UUID, actorEmail string, projectID uuid.UUID, projectName string) error {
	return s.LogAction(ctx, orgID, &actorID, actorEmail, "user", domain.AuditActionProjectCreated,
		domain.AuditResourceProject, &projectID, projectName, fmt.Sprintf("Project '%s' was created", projectName))
}

func (s *AuditService) LogProjectUpdated(ctx context.Context, orgID uuid.UUID, actorID uuid.UUID, actorEmail string, projectID uuid.UUID, projectName string, before, after map[string]any) error {
	return s.LogWithChanges(ctx, orgID, &actorID, actorEmail, "user", domain.AuditActionProjectUpdated,
		domain.AuditResourceProject, &projectID, projectName, fmt.Sprintf("Project '%s' was updated", projectName), before, after)
}

func (s *AuditService) LogProjectDeleted(ctx context.Context, orgID uuid.UUID, actorID uuid.UUID, actorEmail string, projectID uuid.UUID, projectName string) error {
	return s.LogAction(ctx, orgID, &actorID, actorEmail, "user", domain.AuditActionProjectDeleted,
		domain.AuditResourceProject, &projectID, projectName, fmt.Sprintf("Project '%s' was deleted", projectName))
}

func (s *AuditService) LogSSOConfigured(ctx context.Context, orgID uuid.UUID, actorID uuid.UUID, actorEmail, provider string) error {
	input := &domain.AuditLogInput{
		OrganizationID: orgID,
		ActorID:        &actorID,
		ActorEmail:     actorEmail,
		ActorType:      "user",
		Action:         domain.AuditActionSSOConfigured,
		ResourceType:   domain.AuditResourceSSO,
		ResourceID:     &orgID,
		ResourceName:   provider,
		Description:    fmt.Sprintf("SSO was configured with provider: %s", provider),
		Metadata:       map[string]any{"provider": provider},
	}
	_, err := s.auditRepo.CreateAuditLog(ctx, input)
	return err
}

func (s *AuditService) LogSSOEnabled(ctx context.Context, orgID uuid.UUID, actorID uuid.UUID, actorEmail string) error {
	return s.LogAction(ctx, orgID, &actorID, actorEmail, "user", domain.AuditActionSSOEnabled,
		domain.AuditResourceSSO, &orgID, "SSO", "SSO was enabled for the organization")
}

func (s *AuditService) LogSSODisabled(ctx context.Context, orgID uuid.UUID, actorID uuid.UUID, actorEmail string) error {
	return s.LogAction(ctx, orgID, &actorID, actorEmail, "user", domain.AuditActionSSODisabled,
		domain.AuditResourceSSO, &orgID, "SSO", "SSO was disabled for the organization")
}

func (s *AuditService) LogDataExported(ctx context.Context, orgID uuid.UUID, actorID uuid.UUID, actorEmail, dataType string, recordCount int) error {
	input := &domain.AuditLogInput{
		OrganizationID: orgID,
		ActorID:        &actorID,
		ActorEmail:     actorEmail,
		ActorType:      "user",
		Action:         domain.AuditActionDataExported,
		ResourceType:   domain.AuditResourceType(dataType),
		ResourceName:   dataType,
		Description:    fmt.Sprintf("Exported %d %s records", recordCount, dataType),
		Metadata:       map[string]any{"dataType": dataType, "recordCount": recordCount},
	}
	_, err := s.auditRepo.CreateAuditLog(ctx, input)
	return err
}

func (s *AuditService) LogSettingsChanged(ctx context.Context, orgID uuid.UUID, actorID uuid.UUID, actorEmail, settingName string, before, after any) error {
	return s.LogWithChanges(ctx, orgID, &actorID, actorEmail, "user", domain.AuditActionSettingsChanged,
		domain.AuditResourceSettings, &orgID, settingName, fmt.Sprintf("Setting '%s' was changed", settingName),
		map[string]any{settingName: before}, map[string]any{settingName: after})
}

// Cleanup expired sessions
func (s *AuditService) CleanupExpiredData(ctx context.Context) error {
	// This would be called by a background worker
	// to clean up old audit logs based on retention policies
	return nil
}

// GetActivityTimeline returns recent activity for a user or resource
func (s *AuditService) GetActivityTimeline(ctx context.Context, orgID uuid.UUID, resourceType *domain.AuditResourceType, resourceID *uuid.UUID, limit int) ([]domain.AuditLog, error) {
	filter := &domain.AuditLogFilter{
		OrganizationID: &orgID,
		ResourceType:   resourceType,
		ResourceID:     resourceID,
		Limit:          limit,
	}

	result, err := s.auditRepo.ListAuditLogs(ctx, filter)
	if err != nil {
		return nil, err
	}

	return result.Data, nil
}

// GetSecurityEvents returns security-related audit events
func (s *AuditService) GetSecurityEvents(ctx context.Context, orgID uuid.UUID, since time.Time, limit int) ([]domain.AuditLog, error) {
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
		Limit:          limit,
	}

	result, err := s.auditRepo.ListAuditLogs(ctx, filter)
	if err != nil {
		return nil, err
	}

	return result.Data, nil
}
