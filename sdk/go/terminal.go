package agenttrace

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"time"

	"github.com/google/uuid"
)

// TerminalCommandOptions holds options for tracking a terminal command.
type TerminalCommandOptions struct {
	Command          string
	Args             []string
	ObservationID    string
	WorkingDirectory string
	Shell            string
	EnvVars          map[string]string
	ExitCode         int
	Stdout           string
	Stderr           string
	StdoutTruncated  bool
	StderrTruncated  bool
	TimedOut         bool
	Killed           bool
	MaxMemoryBytes   int64
	CPUTimeMs        int64
	ToolName         string
	Reason           string
	StartedAt        *time.Time
	CompletedAt      *time.Time
	Success          *bool
}

// TerminalCommandInfo contains information about a tracked terminal command.
type TerminalCommandInfo struct {
	ID               string
	TraceID          string
	ObservationID    string
	Command          string
	Args             []string
	WorkingDirectory string
	ExitCode         int
	Stdout           string
	Stderr           string
	Success          bool
	DurationMs       int64
	StartedAt        time.Time
	CompletedAt      time.Time
}

// RunCommandOptions holds options for running and tracking a command.
type RunCommandOptions struct {
	Args             []string
	ObservationID    string
	WorkingDirectory string
	Env              map[string]string
	Timeout          time.Duration
	ToolName         string
	Reason           string
	MaxOutputBytes   int
}

// RunCommandResult contains the result of running a command.
type RunCommandResult struct {
	Info     *TerminalCommandInfo
	ExitCode int
	Stdout   string
	Stderr   string
}

// TerminalCmd tracks a terminal command for this trace.
func (t *Trace) TerminalCmd(opts TerminalCommandOptions) *TerminalCommandInfo {
	cmdID := uuid.New().String()
	now := time.Now().UTC()

	startedAt := now
	if opts.StartedAt != nil {
		startedAt = *opts.StartedAt
	}

	completedAt := now
	if opts.CompletedAt != nil {
		completedAt = *opts.CompletedAt
	}

	durationMs := completedAt.Sub(startedAt).Milliseconds()

	workingDirectory := opts.WorkingDirectory
	if workingDirectory == "" {
		workingDirectory, _ = os.Getwd()
	}

	args := opts.Args
	if args == nil {
		args = []string{}
	}

	success := opts.ExitCode == 0
	if opts.Success != nil {
		success = *opts.Success
	}

	// Convert env vars to JSON string
	var envVarsStr string
	if opts.EnvVars != nil {
		if data, err := json.Marshal(opts.EnvVars); err == nil {
			envVarsStr = string(data)
		}
	}

	// Send to API
	if t.client.Enabled() {
		t.client.addEvent(map[string]any{
			"type": "terminal-command-create",
			"body": map[string]any{
				"id":               cmdID,
				"traceId":          t.id,
				"observationId":    opts.ObservationID,
				"command":          opts.Command,
				"args":             args,
				"workingDirectory": workingDirectory,
				"shell":            opts.Shell,
				"envVars":          envVarsStr,
				"startedAt":        startedAt.Format(time.RFC3339Nano),
				"completedAt":      completedAt.Format(time.RFC3339Nano),
				"durationMs":       durationMs,
				"exitCode":         opts.ExitCode,
				"stdout":           opts.Stdout,
				"stderr":           opts.Stderr,
				"stdoutTruncated":  opts.StdoutTruncated,
				"stderrTruncated":  opts.StderrTruncated,
				"success":          success,
				"timedOut":         opts.TimedOut,
				"killed":           opts.Killed,
				"maxMemoryBytes":   opts.MaxMemoryBytes,
				"cpuTimeMs":        opts.CPUTimeMs,
				"toolName":         opts.ToolName,
				"reason":           opts.Reason,
			},
		})
	}

	return &TerminalCommandInfo{
		ID:               cmdID,
		TraceID:          t.id,
		ObservationID:    opts.ObservationID,
		Command:          opts.Command,
		Args:             args,
		WorkingDirectory: workingDirectory,
		ExitCode:         opts.ExitCode,
		Stdout:           opts.Stdout,
		Stderr:           opts.Stderr,
		Success:          success,
		DurationMs:       durationMs,
		StartedAt:        startedAt,
		CompletedAt:      completedAt,
	}
}

// RunCmd runs a command and tracks it.
func (t *Trace) RunCmd(ctx context.Context, command string, opts *RunCommandOptions) *RunCommandResult {
	if opts == nil {
		opts = &RunCommandOptions{}
	}

	startedAt := time.Now().UTC()

	workingDirectory := opts.WorkingDirectory
	if workingDirectory == "" {
		workingDirectory, _ = os.Getwd()
	}

	maxOutputBytes := opts.MaxOutputBytes
	if maxOutputBytes == 0 {
		maxOutputBytes = 100000
	}

	args := opts.Args
	if args == nil {
		args = []string{}
	}

	// Create command
	var cmd *exec.Cmd
	if opts.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, opts.Timeout)
		defer cancel()
	}

	cmd = exec.CommandContext(ctx, command, args...)
	cmd.Dir = workingDirectory

	// Set environment
	if opts.Env != nil {
		cmd.Env = os.Environ()
		for k, v := range opts.Env {
			cmd.Env = append(cmd.Env, k+"="+v)
		}
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	completedAt := time.Now().UTC()

	timedOut := ctx.Err() == context.DeadlineExceeded
	killed := cmd.ProcessState != nil && !cmd.ProcessState.Exited()

	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = -1
		}
	}

	stdoutStr := stdout.String()
	stderrStr := stderr.String()

	stdoutTruncated := len(stdoutStr) > maxOutputBytes
	stderrTruncated := len(stderrStr) > maxOutputBytes

	if stdoutTruncated {
		stdoutStr = stdoutStr[:maxOutputBytes]
	}
	if stderrTruncated {
		stderrStr = stderrStr[:maxOutputBytes]
	}

	info := t.TerminalCmd(TerminalCommandOptions{
		Command:          command,
		Args:             args,
		ObservationID:    opts.ObservationID,
		WorkingDirectory: workingDirectory,
		ExitCode:         exitCode,
		Stdout:           stdoutStr,
		Stderr:           stderrStr,
		StdoutTruncated:  stdoutTruncated,
		StderrTruncated:  stderrTruncated,
		TimedOut:         timedOut,
		Killed:           killed,
		ToolName:         opts.ToolName,
		Reason:           opts.Reason,
		StartedAt:        &startedAt,
		CompletedAt:      &completedAt,
	})

	return &RunCommandResult{
		Info:     info,
		ExitCode: exitCode,
		Stdout:   stdoutStr,
		Stderr:   stderrStr,
	}
}
