package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	agenttrace "github.com/agenttrace/agenttrace-go"
	"github.com/agenttrace/agenttrace-cli/internal/wrapper"
	"github.com/spf13/cobra"
)

var (
	traceName     string
	traceUserID   string
	sessionID     string
	traceTags     []string
	watchFiles    bool
	gitAutoLink   bool
	captureStdout bool
	captureStderr bool
	checkpoints   bool
)

var wrapCmd = &cobra.Command{
	Use:   "wrap [flags] -- command [args...]",
	Short: "Wrap a command and trace its execution",
	Long: `Wrap a command-line tool and automatically trace its execution.

Features:
  - Automatic trace creation for command execution
  - Optional file change detection and checkpointing
  - Git integration for linking traces to commits
  - Stdout/stderr capture

Examples:
  # Basic wrapping
  agenttrace wrap -- python agent.py

  # With custom trace name
  agenttrace wrap --name "code-review" -- ./review.sh

  # With file watching and checkpoints
  agenttrace wrap --watch --checkpoints -- npm run dev

  # With git auto-linking
  agenttrace wrap --git -- git commit -m "fix bug"`,
	Args: cobra.MinimumNArgs(1),
	RunE: runWrap,
}

func init() {
	wrapCmd.Flags().StringVar(&traceName, "name", "", "Custom trace name (defaults to command name)")
	wrapCmd.Flags().StringVar(&traceUserID, "user-id", "", "User ID for the trace")
	wrapCmd.Flags().StringVar(&sessionID, "session-id", "", "Session ID for the trace")
	wrapCmd.Flags().StringSliceVar(&traceTags, "tags", nil, "Tags for the trace")
	wrapCmd.Flags().BoolVar(&watchFiles, "watch", false, "Watch for file changes")
	wrapCmd.Flags().BoolVar(&gitAutoLink, "git", false, "Auto-link to git commits")
	wrapCmd.Flags().BoolVar(&captureStdout, "capture-stdout", true, "Capture stdout")
	wrapCmd.Flags().BoolVar(&captureStderr, "capture-stderr", true, "Capture stderr")
	wrapCmd.Flags().BoolVar(&checkpoints, "checkpoints", false, "Create checkpoints on file changes")
}

func runWrap(cmd *cobra.Command, args []string) error {
	apiKey := getAPIKey()
	if apiKey == "" {
		return fmt.Errorf("API key required. Set --api-key or AGENTTRACE_API_KEY")
	}

	// Initialize AgentTrace client
	client := agenttrace.New(agenttrace.Config{
		APIKey: apiKey,
		Host:   host,
	})
	defer client.Shutdown()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle signals
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		cancel()
	}()

	// Create wrapper
	w := wrapper.New(wrapper.Config{
		Command:       args[0],
		Args:          args[1:],
		Client:        client,
		TraceName:     traceName,
		UserID:        traceUserID,
		SessionID:     sessionID,
		Tags:          traceTags,
		WatchFiles:    watchFiles,
		GitAutoLink:   gitAutoLink,
		CaptureStdout: captureStdout,
		CaptureStderr: captureStderr,
		Checkpoints:   checkpoints,
		Verbose:       verbose,
	})

	logVerbose("Starting trace for: %s", strings.Join(args, " "))

	// Run the wrapped command
	exitCode, err := w.Run(ctx)
	if err != nil {
		return err
	}

	// Flush before exit
	client.Flush()

	if exitCode != 0 {
		os.Exit(exitCode)
	}

	return nil
}

// Simple wrapper run for basic cases
func runSimpleWrap(ctx context.Context, args []string, client *agenttrace.Client) (int, error) {
	name := traceName
	if name == "" {
		name = args[0]
	}

	// Create trace
	trace := client.Trace(ctx, agenttrace.TraceOptions{
		Name:      name,
		UserID:    traceUserID,
		SessionID: sessionID,
		Tags:      traceTags,
		Metadata: map[string]any{
			"command": strings.Join(args, " "),
		},
	})

	// Create the command
	execCmd := exec.CommandContext(ctx, args[0], args[1:]...)
	execCmd.Stdin = os.Stdin
	execCmd.Stdout = os.Stdout
	execCmd.Stderr = os.Stderr

	// Run the command
	startTime := time.Now()
	err := execCmd.Run()
	duration := time.Since(startTime)

	// Get exit code
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = 1
		}
	}

	// End trace
	trace.Update(agenttrace.TraceUpdateOptions{
		Metadata: map[string]any{
			"exit_code":   exitCode,
			"duration_ms": duration.Milliseconds(),
		},
	})

	output := map[string]any{
		"exit_code":   exitCode,
		"duration_ms": duration.Milliseconds(),
	}

	if err != nil && exitCode != 0 {
		output["error"] = err.Error()
	}

	trace.End(&agenttrace.TraceEndOptions{Output: output})

	return exitCode, nil
}
