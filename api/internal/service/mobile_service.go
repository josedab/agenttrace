package service

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

// MobileService manages mobile companion app functionality
type MobileService struct {
	logger        *zap.Logger
	mu            sync.RWMutex
	devices       map[uuid.UUID]*domain.MobileDevice
	notifications map[uuid.UUID]*domain.PushNotification
}

// NewMobileService creates a new mobile service
func NewMobileService(logger *zap.Logger) *MobileService {
	return &MobileService{
		logger:        logger,
		devices:       make(map[uuid.UUID]*domain.MobileDevice),
		notifications: make(map[uuid.UUID]*domain.PushNotification),
	}
}

// RegisterDevice registers a new mobile device
func (s *MobileService) RegisterDevice(ctx context.Context, userID uuid.UUID, input *domain.MobileDeviceInput) (*domain.MobileDevice, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	device := &domain.MobileDevice{
		ID:         uuid.New(),
		UserID:     userID,
		Platform:   input.Platform,
		PushToken:  input.PushToken,
		LastActive: time.Now(),
		CreatedAt:  time.Now(),
	}

	s.devices[device.ID] = device
	s.logger.Info("registered mobile device",
		zap.String("id", device.ID.String()),
		zap.String("platform", string(input.Platform)),
	)
	return device, nil
}

// GetDashboard returns the mobile dashboard summary
func (s *MobileService) GetDashboard(ctx context.Context, projectID uuid.UUID) (*domain.MobileDashboard, error) {
	now := time.Now()
	return &domain.MobileDashboard{
		ActiveAlerts:   3,
		PendingReviews: 7,
		TodayCost:      24.50,
		TodayTraces:    1847,
		RecentActivity: []domain.MobileActivityItem{
			{Type: "alert", Title: "High Error Rate Detected", Description: "Error rate exceeded 5% threshold on code-review-agent", Timestamp: now.Add(-10 * time.Minute), ActionRequired: true},
			{Type: "deployment", Title: "Agent v2.1.0 Deployed", Description: "test-runner-agent updated to version 2.1.0", Timestamp: now.Add(-45 * time.Minute), ActionRequired: false},
			{Type: "cost", Title: "Daily Budget 80% Used", Description: "Project has consumed $24.50 of $30 daily budget", Timestamp: now.Add(-1 * time.Hour), ActionRequired: true},
			{Type: "review", Title: "New Trace Review", Description: "Trace #4521 flagged for manual review", Timestamp: now.Add(-2 * time.Hour), ActionRequired: true},
			{Type: "performance", Title: "Latency Improvement", Description: "Average latency reduced by 15% after config update", Timestamp: now.Add(-3 * time.Hour), ActionRequired: false},
		},
	}, nil
}

// ListNotifications lists notifications for a user
func (s *MobileService) ListNotifications(ctx context.Context, userID uuid.UUID) ([]domain.PushNotification, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []domain.PushNotification
	for _, n := range s.notifications {
		if n.UserID == userID {
			result = append(result, *n)
		}
	}

	if len(result) == 0 {
		now := time.Now()
		sentAt := now.Add(-30 * time.Minute)
		result = []domain.PushNotification{
			{ID: uuid.New(), UserID: userID, Title: "High Error Rate", Body: "Error rate exceeded threshold on code-review-agent", Data: map[string]string{"type": "alert", "agentName": "code-review-agent"}, Sent: true, SentAt: &sentAt},
			{ID: uuid.New(), UserID: userID, Title: "Budget Warning", Body: "Daily budget is 80% consumed", Data: map[string]string{"type": "cost", "percentage": "80"}, Sent: true, SentAt: &sentAt},
			{ID: uuid.New(), UserID: userID, Title: "Review Required", Body: "Trace #4521 needs manual review", Data: map[string]string{"type": "review", "traceId": "4521"}, Sent: false, SentAt: nil},
		}
	}
	return result, nil
}

// SendPushNotification sends a push notification
func (s *MobileService) SendPushNotification(ctx context.Context, notification *domain.PushNotification) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	notification.ID = uuid.New()
	now := time.Now()
	notification.Sent = true
	notification.SentAt = &now
	s.notifications[notification.ID] = notification

	s.logger.Info("sent push notification",
		zap.String("id", notification.ID.String()),
		zap.String("title", notification.Title),
	)
	return nil
}
