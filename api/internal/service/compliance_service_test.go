package service

import (
	"context"
	"crypto/sha256"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

// MockComplianceRepository is a mock implementation of the ComplianceRepository
type MockComplianceRepository struct {
	mock.Mock
}

func (m *MockComplianceRepository) SaveRecord(ctx context.Context, record *domain.ComplianceRecord) error {
	args := m.Called(ctx, record)
	return args.Error(0)
}

func (m *MockComplianceRepository) GetRecord(ctx context.Context, projectID uuid.UUID) (*domain.ComplianceRecord, error) {
	args := m.Called(ctx, projectID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ComplianceRecord), args.Error(1)
}

func (m *MockComplianceRepository) ListRecords(ctx context.Context, projectID uuid.UUID) ([]domain.ComplianceRecord, error) {
	args := m.Called(ctx, projectID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.ComplianceRecord), args.Error(1)
}

func (m *MockComplianceRepository) SaveAuditEntry(ctx context.Context, entry *domain.ImmutableAuditEntry) error {
	args := m.Called(ctx, entry)
	return args.Error(0)
}

func (m *MockComplianceRepository) ListAuditEntries(ctx context.Context, projectID uuid.UUID, start, end time.Time) ([]domain.ImmutableAuditEntry, error) {
	args := m.Called(ctx, projectID, start, end)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.ImmutableAuditEntry), args.Error(1)
}

func (m *MockComplianceRepository) SaveAssessment(ctx context.Context, assessment *domain.ConformityAssessment) error {
	args := m.Called(ctx, assessment)
	return args.Error(0)
}

func (m *MockComplianceRepository) GetAssessment(ctx context.Context, assessmentID uuid.UUID) (*domain.ConformityAssessment, error) {
	args := m.Called(ctx, assessmentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ConformityAssessment), args.Error(1)
}

func TestImmutableAuditEntry_ComputeHash(t *testing.T) {
	t.Run("produces consistent SHA-256 hash", func(t *testing.T) {
		ts := time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC)
		entry := &domain.ImmutableAuditEntry{
			EntryType:    "trace_created",
			Actor:        "user@example.com",
			Action:       "CREATE",
			Details:      `{"traceId":"abc-123"}`,
			PreviousHash: "",
			Timestamp:    ts,
		}

		hash1 := entry.ComputeHash()
		hash2 := entry.ComputeHash()

		assert.Equal(t, hash1, hash2, "same entry should produce the same hash")
		assert.Len(t, hash1, 64, "SHA-256 hex string should be 64 characters")

		// Verify manually
		data := entry.EntryType + entry.Actor + entry.Action + entry.Details + entry.PreviousHash + ts.UTC().Format(time.RFC3339Nano)
		expected := sha256.Sum256([]byte(data))
		assert.Equal(t, fmt.Sprintf("%x", expected), hash1)
	})

	t.Run("different entries produce different hashes", func(t *testing.T) {
		ts := time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC)

		entry1 := &domain.ImmutableAuditEntry{
			EntryType: "trace_created",
			Actor:     "user@example.com",
			Action:    "CREATE",
			Details:   `{"traceId":"abc-123"}`,
			Timestamp: ts,
		}

		entry2 := &domain.ImmutableAuditEntry{
			EntryType: "trace_deleted",
			Actor:     "admin@example.com",
			Action:    "DELETE",
			Details:   `{"traceId":"abc-456"}`,
			Timestamp: ts,
		}

		assert.NotEqual(t, entry1.ComputeHash(), entry2.ComputeHash())
	})

	t.Run("hash changes when previous hash changes", func(t *testing.T) {
		ts := time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC)

		entry := &domain.ImmutableAuditEntry{
			EntryType:    "trace_created",
			Actor:        "user@example.com",
			Action:       "CREATE",
			Details:      `{}`,
			PreviousHash: "",
			Timestamp:    ts,
		}
		hashWithoutPrev := entry.ComputeHash()

		entry.PreviousHash = "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890"
		hashWithPrev := entry.ComputeHash()

		assert.NotEqual(t, hashWithoutPrev, hashWithPrev)
	})
}

func TestImmutableAuditEntry_HashChain(t *testing.T) {
	t.Run("consecutive entries form a valid chain", func(t *testing.T) {
		ts1 := time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC)
		entry1 := &domain.ImmutableAuditEntry{
			ID:           uuid.New(),
			EntryType:    "trace_created",
			Actor:        "user@example.com",
			Action:       "CREATE",
			Details:      `{"traceId":"trace-1"}`,
			PreviousHash: "",
			Timestamp:    ts1,
		}
		entry1.Hash = entry1.ComputeHash()

		ts2 := time.Date(2025, 1, 15, 10, 5, 0, 0, time.UTC)
		entry2 := &domain.ImmutableAuditEntry{
			ID:           uuid.New(),
			EntryType:    "trace_updated",
			Actor:        "user@example.com",
			Action:       "UPDATE",
			Details:      `{"traceId":"trace-1"}`,
			PreviousHash: entry1.Hash,
			Timestamp:    ts2,
		}
		entry2.Hash = entry2.ComputeHash()

		ts3 := time.Date(2025, 1, 15, 10, 10, 0, 0, time.UTC)
		entry3 := &domain.ImmutableAuditEntry{
			ID:           uuid.New(),
			EntryType:    "trace_deleted",
			Actor:        "admin@example.com",
			Action:       "DELETE",
			Details:      `{"traceId":"trace-1"}`,
			PreviousHash: entry2.Hash,
			Timestamp:    ts3,
		}
		entry3.Hash = entry3.ComputeHash()

		// Verify chain links
		assert.Equal(t, "", entry1.PreviousHash, "first entry should have empty previous hash")
		assert.Equal(t, entry1.Hash, entry2.PreviousHash, "entry2.PreviousHash should equal entry1.Hash")
		assert.Equal(t, entry2.Hash, entry3.PreviousHash, "entry3.PreviousHash should equal entry2.Hash")

		// Verify hashes are self-consistent
		assert.Equal(t, entry1.Hash, entry1.ComputeHash())
		assert.Equal(t, entry2.Hash, entry2.ComputeHash())
		assert.Equal(t, entry3.Hash, entry3.ComputeHash())

		// Verify all hashes are unique
		assert.NotEqual(t, entry1.Hash, entry2.Hash)
		assert.NotEqual(t, entry2.Hash, entry3.Hash)
		assert.NotEqual(t, entry1.Hash, entry3.Hash)
	})

	t.Run("tampering with an entry breaks the chain", func(t *testing.T) {
		ts1 := time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC)
		entry1 := &domain.ImmutableAuditEntry{
			EntryType:    "trace_created",
			Actor:        "user@example.com",
			Action:       "CREATE",
			Details:      `{"traceId":"trace-1"}`,
			PreviousHash: "",
			Timestamp:    ts1,
		}
		entry1.Hash = entry1.ComputeHash()
		originalHash := entry1.Hash

		// Tamper with the entry
		entry1.Details = `{"traceId":"trace-TAMPERED"}`

		// The stored hash no longer matches the computed hash
		assert.NotEqual(t, originalHash, entry1.ComputeHash())
	})
}

func TestComplianceService_RecordImmutableAuditEntry(t *testing.T) {
	t.Run("creates first entry with empty previous hash", func(t *testing.T) {
		repo := new(MockComplianceRepository)
		svc := NewComplianceService(zap.NewNop(), repo, nil)

		projectID := uuid.New()
		ctx := context.Background()

		// No existing entries
		repo.On("ListAuditEntries", mock.Anything, projectID, mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time")).
			Return([]domain.ImmutableAuditEntry{}, nil)

		repo.On("SaveAuditEntry", mock.Anything, mock.AnythingOfType("*domain.ImmutableAuditEntry")).
			Return(nil)

		entry := &domain.ImmutableAuditEntry{
			EntryType: "trace_created",
			Actor:     "user@example.com",
			Action:    "CREATE",
			Details:   `{"traceId":"trace-1"}`,
		}

		result, err := svc.RecordImmutableAuditEntry(ctx, projectID, entry)

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "", result.PreviousHash)
		assert.NotEmpty(t, result.Hash)
		assert.Equal(t, projectID, result.ProjectID)
		assert.Equal(t, result.Hash, result.ComputeHash(), "stored hash should match computed hash")
	})

	t.Run("chains to previous entry hash", func(t *testing.T) {
		repo := new(MockComplianceRepository)
		svc := NewComplianceService(zap.NewNop(), repo, nil)

		projectID := uuid.New()
		ctx := context.Background()

		previousEntry := domain.ImmutableAuditEntry{
			ID:        uuid.New(),
			ProjectID: projectID,
			Hash:      "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
		}

		repo.On("ListAuditEntries", mock.Anything, projectID, mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time")).
			Return([]domain.ImmutableAuditEntry{previousEntry}, nil)

		repo.On("SaveAuditEntry", mock.Anything, mock.AnythingOfType("*domain.ImmutableAuditEntry")).
			Return(nil)

		entry := &domain.ImmutableAuditEntry{
			EntryType: "trace_updated",
			Actor:     "user@example.com",
			Action:    "UPDATE",
			Details:   `{"traceId":"trace-1"}`,
		}

		result, err := svc.RecordImmutableAuditEntry(ctx, projectID, entry)

		require.NoError(t, err)
		assert.Equal(t, previousEntry.Hash, result.PreviousHash)
		assert.NotEmpty(t, result.Hash)
		assert.NotEqual(t, previousEntry.Hash, result.Hash)
		assert.Equal(t, result.Hash, result.ComputeHash())
	})

	t.Run("returns error when listing entries fails", func(t *testing.T) {
		repo := new(MockComplianceRepository)
		svc := NewComplianceService(zap.NewNop(), repo, nil)

		projectID := uuid.New()
		ctx := context.Background()

		repo.On("ListAuditEntries", mock.Anything, projectID, mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time")).
			Return(nil, fmt.Errorf("database error"))

		entry := &domain.ImmutableAuditEntry{
			EntryType: "trace_created",
			Actor:     "user@example.com",
			Action:    "CREATE",
		}

		result, err := svc.RecordImmutableAuditEntry(ctx, projectID, entry)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to list audit entries")
	})

	t.Run("returns error when saving entry fails", func(t *testing.T) {
		repo := new(MockComplianceRepository)
		svc := NewComplianceService(zap.NewNop(), repo, nil)

		projectID := uuid.New()
		ctx := context.Background()

		repo.On("ListAuditEntries", mock.Anything, projectID, mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time")).
			Return([]domain.ImmutableAuditEntry{}, nil)

		repo.On("SaveAuditEntry", mock.Anything, mock.AnythingOfType("*domain.ImmutableAuditEntry")).
			Return(fmt.Errorf("write error"))

		entry := &domain.ImmutableAuditEntry{
			EntryType: "trace_created",
			Actor:     "user@example.com",
			Action:    "CREATE",
		}

		result, err := svc.RecordImmutableAuditEntry(ctx, projectID, entry)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to save immutable audit entry")
	})
}
