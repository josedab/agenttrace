package domain

import (
	"time"

	"github.com/google/uuid"
)

// MobilePlatform represents the mobile platform type
type MobilePlatform string

const (
	PlatformIOS     MobilePlatform = "ios"
	PlatformAndroid MobilePlatform = "android"
)

// MobileDevice represents a registered mobile device
type MobileDevice struct {
	ID         uuid.UUID      `json:"id"`
	UserID     uuid.UUID      `json:"userId"`
	Platform   MobilePlatform `json:"platform"`
	PushToken  string         `json:"pushToken"`
	LastActive time.Time      `json:"lastActive"`
	CreatedAt  time.Time      `json:"createdAt"`
}

// PushNotification represents a push notification
type PushNotification struct {
	ID     uuid.UUID         `json:"id"`
	UserID uuid.UUID         `json:"userId"`
	Title  string            `json:"title"`
	Body   string            `json:"body"`
	Data   map[string]string `json:"data,omitempty"`
	Sent   bool              `json:"sent"`
	SentAt *time.Time        `json:"sentAt,omitempty"`
}

// MobileDeviceInput represents input for registering a mobile device
type MobileDeviceInput struct {
	Platform  MobilePlatform `json:"platform" validate:"required"`
	PushToken string         `json:"pushToken" validate:"required"`
}

// MobileDashboard represents the mobile dashboard summary
type MobileDashboard struct {
	ActiveAlerts   int                  `json:"activeAlerts"`
	PendingReviews int                  `json:"pendingReviews"`
	TodayCost      float64              `json:"todayCost"`
	TodayTraces    int                  `json:"todayTraces"`
	RecentActivity []MobileActivityItem `json:"recentActivity"`
}

// MobileActivityItem represents a recent activity item
type MobileActivityItem struct {
	Type           string    `json:"type"`
	Title          string    `json:"title"`
	Description    string    `json:"description"`
	Timestamp      time.Time `json:"timestamp"`
	ActionRequired bool      `json:"actionRequired"`
}
