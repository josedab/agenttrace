package agenttrace

import (
	"context"
	"testing"
	"time"
)

func TestTraceTerminalCmd(t *testing.T) {
	client := New(Config{
		APIKey: "test-api-key",
	})
	defer client.Shutdown()

	ctx := context.Background()
	trace := client.Trace(ctx, TraceOptions{
		Name: "test-trace",
	})

	t.Run("tracks command with minimal options", func(t *testing.T) {
		cmd := trace.TerminalCmd(TerminalCommandOptions{
			Command: "npm",
		})

		if cmd.Command != "npm" {
			t.Errorf("expected command to be 'npm', got '%s'", cmd.Command)
		}
		if cmd.TraceID != trace.id {
			t.Error("expected terminal cmd to have trace's ID")
		}
		if cmd.ID == "" {
			t.Error("expected terminal cmd ID to be generated")
		}
		if len(cmd.Args) != 0 {
			t.Error("expected empty args by default")
		}
		if cmd.WorkingDirectory == "" {
			t.Error("expected working directory to be set")
		}
	})

	t.Run("tracks command with args", func(t *testing.T) {
		cmd := trace.TerminalCmd(TerminalCommandOptions{
			Command: "npm",
			Args:    []string{"install", "--save", "lodash"},
		})

		if cmd.Command != "npm" {
			t.Errorf("expected command to be 'npm', got '%s'", cmd.Command)
		}
		if len(cmd.Args) != 3 {
			t.Errorf("expected 3 args, got %d", len(cmd.Args))
		}
		if cmd.Args[0] != "install" || cmd.Args[1] != "--save" || cmd.Args[2] != "lodash" {
			t.Error("args not set correctly")
		}
	})

	t.Run("tracks command with output", func(t *testing.T) {
		cmd := trace.TerminalCmd(TerminalCommandOptions{
			Command:  "echo",
			Args:     []string{"hello"},
			Stdout:   "hello\n",
			Stderr:   "",
			ExitCode: 0,
		})

		if cmd.Stdout != "hello\n" {
			t.Errorf("expected stdout to be 'hello\\n', got '%s'", cmd.Stdout)
		}
		if cmd.Stderr != "" {
			t.Errorf("expected stderr to be empty, got '%s'", cmd.Stderr)
		}
		if cmd.ExitCode != 0 {
			t.Errorf("expected exitCode to be 0, got %d", cmd.ExitCode)
		}
		if cmd.Success != true {
			t.Error("expected success to be true")
		}
	})

	t.Run("tracks failed command", func(t *testing.T) {
		cmd := trace.TerminalCmd(TerminalCommandOptions{
			Command:  "false",
			ExitCode: 1,
			Stderr:   "Error occurred",
		})

		if cmd.ExitCode != 1 {
			t.Errorf("expected exitCode to be 1, got %d", cmd.ExitCode)
		}
		if cmd.Success != false {
			t.Error("expected success to be false")
		}
		if cmd.Stderr != "Error occurred" {
			t.Error("expected stderr to be set")
		}
	})

	t.Run("tracks with custom working directory", func(t *testing.T) {
		cmd := trace.TerminalCmd(TerminalCommandOptions{
			Command:          "ls",
			WorkingDirectory: "/custom/path",
		})

		if cmd.WorkingDirectory != "/custom/path" {
			t.Errorf("expected workingDirectory to be '/custom/path', got '%s'", cmd.WorkingDirectory)
		}
	})

	t.Run("tracks with observation ID", func(t *testing.T) {
		cmd := trace.TerminalCmd(TerminalCommandOptions{
			Command:       "test",
			ObservationID: "obs-456",
		})

		if cmd.ObservationID != "obs-456" {
			t.Errorf("expected observationID to be 'obs-456', got '%s'", cmd.ObservationID)
		}
	})

	t.Run("calculates duration", func(t *testing.T) {
		startedAt := time.Now().Add(-5 * time.Second)
		completedAt := time.Now()

		cmd := trace.TerminalCmd(TerminalCommandOptions{
			Command:     "long-command",
			StartedAt:   &startedAt,
			CompletedAt: &completedAt,
		})

		if cmd.DurationMs < 4900 || cmd.DurationMs > 5100 {
			t.Errorf("expected duration to be ~5000ms, got %d", cmd.DurationMs)
		}
	})

	t.Run("tracks timed out command", func(t *testing.T) {
		cmd := trace.TerminalCmd(TerminalCommandOptions{
			Command:  "slow-command",
			TimedOut: true,
			Killed:   true,
			ExitCode: -1,
		})

		if cmd.ExitCode != -1 {
			t.Error("expected exitCode to be -1 for killed process")
		}
	})

	t.Run("tracks truncated output", func(t *testing.T) {
		cmd := trace.TerminalCmd(TerminalCommandOptions{
			Command:         "verbose-command",
			Stdout:          "truncated...",
			StdoutTruncated: true,
			Stderr:          "errors...",
			StderrTruncated: true,
		})

		if cmd.Stdout != "truncated..." {
			t.Error("expected truncated stdout")
		}
	})

	t.Run("handles nil args", func(t *testing.T) {
		cmd := trace.TerminalCmd(TerminalCommandOptions{
			Command: "test",
			Args:    nil,
		})

		if cmd.Args == nil {
			t.Error("expected args to be initialized to empty slice")
		}
		if len(cmd.Args) != 0 {
			t.Error("expected args to be empty")
		}
	})
}

func TestTraceRunCmd(t *testing.T) {
	client := New(Config{
		APIKey: "test-api-key",
	})
	defer client.Shutdown()

	ctx := context.Background()
	trace := client.Trace(ctx, TraceOptions{
		Name: "test-trace",
	})

	t.Run("runs echo command", func(t *testing.T) {
		result := trace.RunCmd(ctx, "echo", &RunCommandOptions{
			Args: []string{"hello"},
		})

		if result.ExitCode != 0 {
			t.Errorf("expected exitCode to be 0, got %d", result.ExitCode)
		}
		if result.Stdout != "hello\n" {
			t.Errorf("expected stdout to be 'hello\\n', got '%s'", result.Stdout)
		}
		if result.Info == nil {
			t.Error("expected info to be set")
		}
		if result.Info.Command != "echo" {
			t.Errorf("expected command to be 'echo', got '%s'", result.Info.Command)
		}
	})

	t.Run("handles non-zero exit code", func(t *testing.T) {
		result := trace.RunCmd(ctx, "false", nil)

		if result.ExitCode == 0 {
			t.Error("expected non-zero exit code")
		}
		if result.Info.Success != false {
			t.Error("expected success to be false")
		}
	})

	t.Run("handles nonexistent command", func(t *testing.T) {
		result := trace.RunCmd(ctx, "nonexistent-command-12345", nil)

		if result.ExitCode != -1 {
			t.Errorf("expected exitCode to be -1 for failed command, got %d", result.ExitCode)
		}
	})

	t.Run("respects timeout", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
		defer cancel()

		result := trace.RunCmd(ctx, "sleep", &RunCommandOptions{
			Args:    []string{"10"},
			Timeout: 100 * time.Millisecond,
		})

		// Command should be killed due to timeout
		if result.ExitCode == 0 {
			t.Error("expected command to fail due to timeout")
		}
	})

	t.Run("uses custom working directory", func(t *testing.T) {
		result := trace.RunCmd(ctx, "pwd", &RunCommandOptions{
			WorkingDirectory: "/tmp",
		})

		if result.ExitCode != 0 {
			t.Errorf("expected exitCode to be 0, got %d", result.ExitCode)
		}
		if result.Info.WorkingDirectory != "/tmp" {
			t.Errorf("expected workingDirectory to be '/tmp', got '%s'", result.Info.WorkingDirectory)
		}
	})

	t.Run("passes environment variables", func(t *testing.T) {
		result := trace.RunCmd(ctx, "sh", &RunCommandOptions{
			Args: []string{"-c", "echo $TEST_VAR"},
			Env:  map[string]string{"TEST_VAR": "test_value"},
		})

		if result.ExitCode != 0 {
			t.Errorf("expected exitCode to be 0, got %d", result.ExitCode)
		}
		if result.Stdout != "test_value\n" {
			t.Errorf("expected stdout to contain env var value, got '%s'", result.Stdout)
		}
	})

	t.Run("truncates long output", func(t *testing.T) {
		result := trace.RunCmd(ctx, "sh", &RunCommandOptions{
			Args:           []string{"-c", "yes | head -1000"},
			MaxOutputBytes: 100,
		})

		if len(result.Stdout) > 100 {
			t.Errorf("expected stdout to be truncated to 100 bytes, got %d", len(result.Stdout))
		}
	})
}

func TestTerminalCmdDisabledClient(t *testing.T) {
	enabled := false
	client := New(Config{
		APIKey:  "test-api-key",
		Enabled: &enabled,
	})
	defer client.Shutdown()

	ctx := context.Background()
	trace := client.Trace(ctx, TraceOptions{
		Name: "test-trace",
	})

	cmd := trace.TerminalCmd(TerminalCommandOptions{
		Command: "npm",
		Args:    []string{"test"},
	})

	// Should still return command info
	if cmd.Command != "npm" {
		t.Error("expected command info even when disabled")
	}
	if cmd.ID == "" {
		t.Error("expected command ID even when disabled")
	}
}

func TestTerminalCommandInfo(t *testing.T) {
	t.Run("has all expected fields", func(t *testing.T) {
		now := time.Now()
		info := TerminalCommandInfo{
			ID:               "cmd-123",
			TraceID:          "trace-456",
			ObservationID:    "obs-789",
			Command:          "npm",
			Args:             []string{"test"},
			WorkingDirectory: "/path/to/project",
			ExitCode:         0,
			Stdout:           "output",
			Stderr:           "",
			Success:          true,
			DurationMs:       1000,
			StartedAt:        now,
			CompletedAt:      now.Add(time.Second),
		}

		if info.ID != "cmd-123" {
			t.Error("ID not set correctly")
		}
		if info.TraceID != "trace-456" {
			t.Error("TraceID not set correctly")
		}
		if info.ObservationID != "obs-789" {
			t.Error("ObservationID not set correctly")
		}
		if info.Command != "npm" {
			t.Error("Command not set correctly")
		}
		if len(info.Args) != 1 || info.Args[0] != "test" {
			t.Error("Args not set correctly")
		}
		if info.WorkingDirectory != "/path/to/project" {
			t.Error("WorkingDirectory not set correctly")
		}
		if info.ExitCode != 0 {
			t.Error("ExitCode not set correctly")
		}
		if info.Stdout != "output" {
			t.Error("Stdout not set correctly")
		}
		if info.Stderr != "" {
			t.Error("Stderr not set correctly")
		}
		if info.Success != true {
			t.Error("Success not set correctly")
		}
		if info.DurationMs != 1000 {
			t.Error("DurationMs not set correctly")
		}
	})
}

func TestRunCommandOptions(t *testing.T) {
	t.Run("has sensible defaults when nil", func(t *testing.T) {
		client := New(Config{
			APIKey: "test-api-key",
		})
		defer client.Shutdown()

		ctx := context.Background()
		trace := client.Trace(ctx, TraceOptions{
			Name: "test-trace",
		})

		// Should not panic with nil options
		result := trace.RunCmd(ctx, "echo", nil)

		if result.Info.Command != "echo" {
			t.Error("expected command to be set")
		}
	})
}
