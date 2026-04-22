package postgres

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type migrationLockContextKey struct{}

type migrationLockTxStub struct {
	execErr     error
	commitErr   error
	rollbackErr error

	execSQL           string
	execArgs          []any
	commitCtx         context.Context
	commitCtxErr      error
	commitValue       any
	commitDeadline    time.Time
	commitHasDeadline bool
	rollbackCtx       context.Context
	commitCalls       int
	rollbackCalls     int
}

func (s *migrationLockTxStub) Begin(context.Context) (pgx.Tx, error) {
	panic("unexpected nested transaction")
}

func (s *migrationLockTxStub) Commit(ctx context.Context) error {
	s.commitCalls++
	s.commitCtx = ctx
	s.commitCtxErr = ctx.Err()
	s.commitValue = ctx.Value(migrationLockContextKey{})
	s.commitDeadline, s.commitHasDeadline = ctx.Deadline()
	return s.commitErr
}

func (s *migrationLockTxStub) Rollback(ctx context.Context) error {
	s.rollbackCalls++
	s.rollbackCtx = ctx
	return s.rollbackErr
}

func (s *migrationLockTxStub) CopyFrom(
	context.Context,
	pgx.Identifier,
	[]string,
	pgx.CopyFromSource,
) (int64, error) {
	panic("unexpected CopyFrom")
}

func (s *migrationLockTxStub) SendBatch(context.Context, *pgx.Batch) pgx.BatchResults {
	panic("unexpected SendBatch")
}

func (s *migrationLockTxStub) LargeObjects() pgx.LargeObjects {
	panic("unexpected LargeObjects")
}

func (s *migrationLockTxStub) Prepare(
	context.Context,
	string,
	string,
) (*pgconn.StatementDescription, error) {
	panic("unexpected Prepare")
}

func (s *migrationLockTxStub) Exec(
	_ context.Context,
	sql string,
	arguments ...any,
) (pgconn.CommandTag, error) {
	s.execSQL = sql
	s.execArgs = append([]any(nil), arguments...)
	return pgconn.CommandTag{}, s.execErr
}

func (s *migrationLockTxStub) Query(
	context.Context,
	string,
	...any,
) (pgx.Rows, error) {
	panic("unexpected Query")
}

func (s *migrationLockTxStub) QueryRow(
	context.Context,
	string,
	...any,
) pgx.Row {
	panic("unexpected QueryRow")
}

func (s *migrationLockTxStub) Conn() *pgx.Conn {
	return nil
}

func TestWithMigrationBatchLockCommitsWithDetachedContext(t *testing.T) {
	projectID := uuid.New()
	jobID := uuid.New()
	tx := &migrationLockTxStub{}
	ctx, cancel := context.WithCancel(context.WithValue(
		context.Background(),
		migrationLockContextKey{},
		"request-value",
	))

	err := withMigrationBatchLock(
		ctx,
		func(context.Context) (pgx.Tx, error) { return tx, nil },
		projectID,
		jobID,
		func(lockCtx context.Context) error {
			assert.Equal(t, "request-value", lockCtx.Value(migrationLockContextKey{}))
			cancel()
			return nil
		},
	)

	require.NoError(t, err)
	assert.Contains(t, tx.execSQL, "pg_advisory_xact_lock")
	assert.Equal(t, []any{
		"langfuse-import:" + projectID.String(),
		jobID.String(),
	}, tx.execArgs)
	require.NotNil(t, tx.commitCtx)
	assert.NoError(t, tx.commitCtxErr, "commit context must be detached from request cancellation")
	assert.Equal(t, "request-value", tx.commitValue)
	require.True(t, tx.commitHasDeadline)
	assert.LessOrEqual(t, time.Until(tx.commitDeadline), batchLockReleaseTimeout)
	assert.Equal(t, 1, tx.commitCalls)
	assert.Zero(t, tx.rollbackCalls)
}

func TestWithMigrationBatchLockPropagatesAcquisitionAndReleaseErrors(t *testing.T) {
	acquireErr := errors.New("lock unavailable")
	releaseErr := errors.New("rollback failed")
	tx := &migrationLockTxStub{
		execErr:     acquireErr,
		rollbackErr: releaseErr,
	}

	err := withMigrationBatchLock(
		context.Background(),
		func(context.Context) (pgx.Tx, error) { return tx, nil },
		uuid.New(),
		uuid.New(),
		func(context.Context) error {
			t.Fatal("callback must not run when lock acquisition fails")
			return nil
		},
	)

	require.Error(t, err)
	assert.ErrorIs(t, err, acquireErr)
	assert.ErrorIs(t, err, releaseErr)
	assert.Zero(t, tx.commitCalls)
	assert.Equal(t, 1, tx.rollbackCalls)
}

func TestWithMigrationBatchLockPropagatesCallbackAndReleaseErrors(t *testing.T) {
	callbackErr := errors.New("batch failed")
	releaseErr := errors.New("rollback failed")
	tx := &migrationLockTxStub{rollbackErr: releaseErr}

	err := withMigrationBatchLock(
		context.Background(),
		func(context.Context) (pgx.Tx, error) { return tx, nil },
		uuid.New(),
		uuid.New(),
		func(context.Context) error { return callbackErr },
	)

	require.Error(t, err)
	assert.ErrorIs(t, err, callbackErr)
	assert.ErrorIs(t, err, releaseErr)
	assert.Zero(t, tx.commitCalls)
	assert.Equal(t, 1, tx.rollbackCalls)
}

func TestWithMigrationBatchLockPropagatesCommitError(t *testing.T) {
	releaseErr := errors.New("commit failed")
	tx := &migrationLockTxStub{
		commitErr:   releaseErr,
		rollbackErr: pgx.ErrTxClosed,
	}

	err := withMigrationBatchLock(
		context.Background(),
		func(context.Context) (pgx.Tx, error) { return tx, nil },
		uuid.New(),
		uuid.New(),
		func(context.Context) error { return nil },
	)

	require.Error(t, err)
	assert.ErrorIs(t, err, releaseErr)
	assert.Equal(t, 1, tx.commitCalls)
	assert.Equal(t, 1, tx.rollbackCalls, "rollback is attempted after a failed commit")
}
