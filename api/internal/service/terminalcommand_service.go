package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

// TerminalCommandRepository defines terminal command repository operations
type TerminalCommandRepository interface {
	Create(ctx context.Context, cmd *domain.TerminalCommand) error
	CreateBatch(ctx context.Context, cmds []*domain.TerminalCommand) error
	GetByTraceID(ctx context.Context, projectID uuid.UUID, traceID string) ([]domain.TerminalCommand, error)
	List(ctx context.Context, filter *domain.TerminalCommandFilter, limit, offset int) (*domain.TerminalCommandList, error)
	GetStats(ctx context.Context, projectID uuid.UUID, traceID *string) (*domain.TerminalCommandStats, error)
}

// TerminalCommandService handles terminal command logging
type TerminalCommandService struct {
	termCmdRepo TerminalCommandRepository
	traceRepo   TraceRepository
}

// NewTerminalCommandService creates a new terminal command service
func NewTerminalCommandService(
	termCmdRepo TerminalCommandRepository,
	traceRepo TraceRepository,
) *TerminalCommandService {
	return &TerminalCommandService{
		termCmdRepo: termCmdRepo,
		traceRepo:   traceRepo,
	}
}

// Log records a terminal command execution
func (s *TerminalCommandService) Log(ctx context.Context, projectID uuid.UUID, input *domain.TerminalCommandInput) (*domain.TerminalCommand, error) {
	// Verify trace exists
	_, err := s.traceRepo.GetByID(ctx, projectID, input.TraceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get trace: %w", err)
	}

	now := time.Now()
	startedAt := now
	if input.StartedAt != nil {
		startedAt = *input.StartedAt
	}

	var completedAt *time.Time
	if input.CompletedAt != nil {
		completedAt = input.CompletedAt
	} else {
		completedAt = &now
	}

	durationMs := uint32(completedAt.Sub(startedAt).Milliseconds())

	var workingDirectory, shell, stdout, stderr string
	var toolName, reason, envVars string

	if input.WorkingDirectory != nil {
		workingDirectory = *input.WorkingDirectory
	}
	if input.Shell != nil {
		shell = *input.Shell
	}
	if input.Stdout != nil {
		stdout = *input.Stdout
	}
	if input.Stderr != nil {
		stderr = *input.Stderr
	}
	if input.ToolName != nil {
		toolName = *input.ToolName
	}
	if input.Reason != nil {
		reason = *input.Reason
	}
	if input.EnvVars != nil {
		envVars = *input.EnvVars
	}

	var exitCode int32
	var maxMemoryBytes uint64
	var cpuTimeMs uint32

	if input.ExitCode != nil {
		exitCode = *input.ExitCode
	}
	if input.MaxMemoryBytes != nil {
		maxMemoryBytes = *input.MaxMemoryBytes
	}
	if input.CPUTimeMs != nil {
		cpuTimeMs = *input.CPUTimeMs
	}

	success := exitCode == 0
	if input.Success != nil {
		success = *input.Success
	}

	var timedOut, killed, stdoutTruncated, stderrTruncated bool
	if input.TimedOut != nil {
		timedOut = *input.TimedOut
	}
	if input.Killed != nil {
		killed = *input.Killed
	}
	if input.StdoutTruncated != nil {
		stdoutTruncated = *input.StdoutTruncated
	}
	if input.StderrTruncated != nil {
		stderrTruncated = *input.StderrTruncated
	}

	var args []string
	if input.Args != nil {
		args = input.Args
	}

	cmd := &domain.TerminalCommand{
		ID:               uuid.New(),
		ProjectID:        projectID,
		TraceID:          input.TraceID,
		ObservationID:    input.ObservationID,
		Command:          input.Command,
		Args:             args,
		WorkingDirectory: workingDirectory,
		Shell:            shell,
		EnvVars:          envVars,
		StartedAt:        startedAt,
		CompletedAt:      completedAt,
		DurationMs:       durationMs,
		ExitCode:         exitCode,
		Stdout:           stdout,
		Stderr:           stderr,
		StdoutTruncated:  stdoutTruncated,
		StderrTruncated:  stderrTruncated,
		Success:          success,
		TimedOut:         timedOut,
		Killed:           killed,
		MaxMemoryBytes:   maxMemoryBytes,
		CPUTimeMs:        cpuTimeMs,
		ToolName:         toolName,
		Reason:           reason,
	}

	if err := s.termCmdRepo.Create(ctx, cmd); err != nil {
		return nil, fmt.Errorf("failed to create terminal command: %w", err)
	}

	return cmd, nil
}

// LogBatch records multiple terminal command executions
func (s *TerminalCommandService) LogBatch(ctx context.Context, projectID uuid.UUID, inputs []*domain.TerminalCommandInput) ([]*domain.TerminalCommand, error) {
	if len(inputs) == 0 {
		return []*domain.TerminalCommand{}, nil
	}

	// Verify trace exists for first input
	_, err := s.traceRepo.GetByID(ctx, projectID, inputs[0].TraceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get trace: %w", err)
	}

	now := time.Now()
	cmds := make([]*domain.TerminalCommand, len(inputs))

	for i, input := range inputs {
		startedAt := now
		if input.StartedAt != nil {
			startedAt = *input.StartedAt
		}

		var completedAt *time.Time
		if input.CompletedAt != nil {
			completedAt = input.CompletedAt
		} else {
			nowCopy := now
			completedAt = &nowCopy
		}

		durationMs := uint32(completedAt.Sub(startedAt).Milliseconds())

		var workingDirectory, shell, stdout, stderr string
		var toolName, reason, envVars string

		if input.WorkingDirectory != nil {
			workingDirectory = *input.WorkingDirectory
		}
		if input.Shell != nil {
			shell = *input.Shell
		}
		if input.Stdout != nil {
			stdout = *input.Stdout
		}
		if input.Stderr != nil {
			stderr = *input.Stderr
		}
		if input.ToolName != nil {
			toolName = *input.ToolName
		}
		if input.Reason != nil {
			reason = *input.Reason
		}
		if input.EnvVars != nil {
			envVars = *input.EnvVars
		}

		var exitCode int32
		var maxMemoryBytes uint64
		var cpuTimeMs uint32

		if input.ExitCode != nil {
			exitCode = *input.ExitCode
		}
		if input.MaxMemoryBytes != nil {
			maxMemoryBytes = *input.MaxMemoryBytes
		}
		if input.CPUTimeMs != nil {
			cpuTimeMs = *input.CPUTimeMs
		}

		success := exitCode == 0
		if input.Success != nil {
			success = *input.Success
		}

		var timedOut, killed, stdoutTruncated, stderrTruncated bool
		if input.TimedOut != nil {
			timedOut = *input.TimedOut
		}
		if input.Killed != nil {
			killed = *input.Killed
		}
		if input.StdoutTruncated != nil {
			stdoutTruncated = *input.StdoutTruncated
		}
		if input.StderrTruncated != nil {
			stderrTruncated = *input.StderrTruncated
		}

		var args []string
		if input.Args != nil {
			args = input.Args
		}

		cmds[i] = &domain.TerminalCommand{
			ID:               uuid.New(),
			ProjectID:        projectID,
			TraceID:          input.TraceID,
			ObservationID:    input.ObservationID,
			Command:          input.Command,
			Args:             args,
			WorkingDirectory: workingDirectory,
			Shell:            shell,
			EnvVars:          envVars,
			StartedAt:        startedAt,
			CompletedAt:      completedAt,
			DurationMs:       durationMs,
			ExitCode:         exitCode,
			Stdout:           stdout,
			Stderr:           stderr,
			StdoutTruncated:  stdoutTruncated,
			StderrTruncated:  stderrTruncated,
			Success:          success,
			TimedOut:         timedOut,
			Killed:           killed,
			MaxMemoryBytes:   maxMemoryBytes,
			CPUTimeMs:        cpuTimeMs,
			ToolName:         toolName,
			Reason:           reason,
		}
	}

	if err := s.termCmdRepo.CreateBatch(ctx, cmds); err != nil {
		return nil, fmt.Errorf("failed to create terminal commands: %w", err)
	}

	return cmds, nil
}

// GetByTraceID retrieves terminal commands for a trace
func (s *TerminalCommandService) GetByTraceID(ctx context.Context, projectID uuid.UUID, traceID string) ([]domain.TerminalCommand, error) {
	return s.termCmdRepo.GetByTraceID(ctx, projectID, traceID)
}

// List retrieves terminal commands with filtering
func (s *TerminalCommandService) List(ctx context.Context, filter *domain.TerminalCommandFilter, limit, offset int) (*domain.TerminalCommandList, error) {
	return s.termCmdRepo.List(ctx, filter, limit, offset)
}

// GetStats retrieves terminal command statistics
func (s *TerminalCommandService) GetStats(ctx context.Context, projectID uuid.UUID, traceID *string) (*domain.TerminalCommandStats, error) {
	return s.termCmdRepo.GetStats(ctx, projectID, traceID)
}
