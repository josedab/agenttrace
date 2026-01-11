package testutil

import (
	"time"

	"github.com/google/uuid"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

// NewTestProject creates a test project with default values.
func NewTestProject(orgID uuid.UUID) *domain.Project {
	return &domain.Project{
		ID:             uuid.New(),
		Name:           "test-project",
		OrganizationID: orgID,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
}

// NewTestOrganization creates a test organization with default values.
func NewTestOrganization() *domain.Organization {
	return &domain.Organization{
		ID:        uuid.New(),
		Name:      "test-org",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// NewTestUser creates a test user with default values.
func NewTestUser(orgID uuid.UUID) *domain.User {
	return &domain.User{
		ID:             uuid.New(),
		Email:          "test@example.com",
		Name:           "Test User",
		OrganizationID: orgID,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
}

// NewTestAPIKey creates a test API key with default values.
func NewTestAPIKey(projectID uuid.UUID) *domain.APIKey {
	return &domain.APIKey{
		ID:        uuid.New(),
		Name:      "test-key",
		ProjectID: projectID,
		CreatedAt: time.Now(),
	}
}

// NewTestTrace creates a test trace with default values.
func NewTestTrace(projectID uuid.UUID) *domain.Trace {
	return &domain.Trace{
		ID:        "trace-" + uuid.New().String()[:8],
		ProjectID: projectID,
		Name:      "test-trace",
		Input:     "test input",
		Output:    "test output",
		StartTime: time.Now(),
		EndTime:   timePtr(time.Now().Add(time.Second)),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// NewTestDataset creates a test dataset with default values.
func NewTestDataset(projectID uuid.UUID) *domain.Dataset {
	return &domain.Dataset{
		ID:          uuid.New(),
		ProjectID:   projectID,
		Name:        "test-dataset",
		Description: "Test dataset description",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// NewTestScore creates a test score with default values.
func NewTestScore(projectID uuid.UUID, traceID string) *domain.Score {
	return &domain.Score{
		ID:        uuid.New(),
		ProjectID: projectID,
		TraceID:   traceID,
		Name:      "test-score",
		Value:     0.85,
		Source:    domain.ScoreSourceAPI,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// timePtr returns a pointer to the given time.
func timePtr(t time.Time) *time.Time {
	return &t
}
