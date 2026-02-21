package domain

import (
	"time"

	"github.com/google/uuid"
)

// TenantPlan represents the subscription plan for a tenant
type TenantPlan string

const (
	TenantPlanFree       TenantPlan = "FREE"
	TenantPlanPro        TenantPlan = "PRO"
	TenantPlanEnterprise TenantPlan = "ENTERPRISE"
)

// Tenant represents a multi-tenant account
type Tenant struct {
	ID           uuid.UUID    `json:"id"`
	Name         string       `json:"name"`
	Slug         string       `json:"slug"`
	Plan         TenantPlan   `json:"plan"`
	UsageLimits  TenantLimits `json:"usageLimits"`
	CurrentUsage TenantUsage  `json:"currentUsage"`
	CreatedAt    time.Time    `json:"createdAt"`
}

// TenantLimits defines usage limits for a tenant's plan
type TenantLimits struct {
	MaxTracesPerMonth int   `json:"maxTracesPerMonth"`
	MaxProjects       int   `json:"maxProjects"`
	MaxUsers          int   `json:"maxUsers"`
	MaxRetentionDays  int   `json:"maxRetentionDays"`
	MaxStorageGB      int64 `json:"maxStorageGB"`
}

// TenantUsage tracks current resource usage for a tenant
type TenantUsage struct {
	TracesThisMonth int       `json:"tracesThisMonth"`
	ProjectCount    int       `json:"projectCount"`
	UserCount       int       `json:"userCount"`
	StorageUsedGB   float64   `json:"storageUsedGB"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

// UsageEvent represents a single usage event for metering
type UsageEvent struct {
	TenantID  uuid.UUID      `json:"tenantId"`
	EventType UsageEventType `json:"eventType"`
	Value     int64          `json:"value"`
	Timestamp time.Time      `json:"timestamp"`
}

// UsageEventType represents the type of usage event
type UsageEventType string

const (
	UsageEventTraceIngested UsageEventType = "trace_ingested"
	UsageEventStorageUsed   UsageEventType = "storage_used"
)
