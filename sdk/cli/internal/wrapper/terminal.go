// Package wrapper provides terminal command tracking functionality.
package wrapper

import (
	"context"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	agenttrace "github.com/agenttrace/agenttrace-go"
)

// TerminalTracker tracks terminal commands during execution.
type TerminalTracker struct {
	trace    *agenttrace.Trace
	commands []*agenttrace.TerminalCommandInfo
}

// NewTerminalTracker creates a new terminal tracker.
func NewTerminalTracker(trace *agenttrace.Trace) *TerminalTracker {
	return &TerminalTracker{
		trace:    trace,
		commands: make([]*agenttrace.TerminalCommandInfo, 0),
	}
}

// TrackCommand tracks a command that was executed.
func (t *TerminalTracker) TrackCommand(opts agenttrace.TerminalCommandOptions) *agenttrace.TerminalCommandInfo {
	info := t.trace.TerminalCmd(opts)
	t.commands = append(t.commands, info)
	return info
}

// RunCommand runs and tracks a command.
func (t *TerminalTracker) RunCommand(ctx context.Context, command string, args []string) *agenttrace.RunCommandResult {
	result := t.trace.RunCmd(ctx, command, &agenttrace.RunCommandOptions{
		Args: args,
	})
	t.commands = append(t.commands, result.Info)
	return result
}

// GetCommands returns all tracked commands.
func (t *TerminalTracker) GetCommands() []*agenttrace.TerminalCommandInfo {
	return t.commands
}

// TrackedCommand wraps exec.Cmd with automatic tracking.
type TrackedCommand struct {
	cmd       *exec.Cmd
	tracker   *TerminalTracker
	command   string
	args      []string
	startTime time.Time
	stdout    strings.Builder
	stderr    strings.Builder
}

// NewTrackedCommand creates a new tracked command.
func (t *TerminalTracker) NewTrackedCommand(ctx context.Context, command string, args ...string) *TrackedCommand {
	cmd := exec.CommandContext(ctx, command, args...)
	return &TrackedCommand{
		cmd:     cmd,
		tracker: t,
		command: command,
		args:    args,
	}
}

// SetStdin sets the command's stdin.
func (tc *TrackedCommand) SetStdin(r io.Reader) {
	tc.cmd.Stdin = r
}

// SetDir sets the command's working directory.
func (tc *TrackedCommand) SetDir(dir string) {
	tc.cmd.Dir = dir
}

// SetEnv sets the command's environment.
func (tc *TrackedCommand) SetEnv(env []string) {
	tc.cmd.Env = env
}

// Start starts the command.
func (tc *TrackedCommand) Start() error {
	tc.startTime = time.Now().UTC()
	return tc.cmd.Start()
}

// Wait waits for the command to complete and tracks it.
func (tc *TrackedCommand) Wait() (*agenttrace.TerminalCommandInfo, error) {
	err := tc.cmd.Wait()
	completedAt := time.Now().UTC()

	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = -1
		}
	}

	workingDir := tc.cmd.Dir
	if workingDir == "" {
		workingDir, _ = os.Getwd()
	}

	info := tc.tracker.TrackCommand(agenttrace.TerminalCommandOptions{
		Command:          tc.command,
		Args:             tc.args,
		WorkingDirectory: workingDir,
		ExitCode:         exitCode,
		Stdout:           tc.stdout.String(),
		Stderr:           tc.stderr.String(),
		StartedAt:        &tc.startTime,
		CompletedAt:      &completedAt,
	})

	return info, err
}

// Run starts and waits for the command.
func (tc *TrackedCommand) Run() (*agenttrace.TerminalCommandInfo, error) {
	if err := tc.Start(); err != nil {
		now := time.Now().UTC()
		return tc.tracker.TrackCommand(agenttrace.TerminalCommandOptions{
			Command:     tc.command,
			Args:        tc.args,
			ExitCode:    -1,
			Stderr:      err.Error(),
			StartedAt:   &now,
			CompletedAt: &now,
		}), err
	}
	return tc.Wait()
}

// Cmd returns the underlying exec.Cmd.
func (tc *TrackedCommand) Cmd() *exec.Cmd {
	return tc.cmd
}
