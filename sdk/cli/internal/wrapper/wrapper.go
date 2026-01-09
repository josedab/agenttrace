// Package wrapper provides command wrapping functionality for tracing.
package wrapper

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	agenttrace "github.com/agenttrace/agenttrace-go"
	"github.com/fsnotify/fsnotify"
	"github.com/google/uuid"
)

// Config holds the wrapper configuration.
type Config struct {
	Command       string
	Args          []string
	Client        *agenttrace.Client
	TraceName     string
	UserID        string
	SessionID     string
	Tags          []string
	WatchFiles    bool
	GitAutoLink   bool
	CaptureStdout bool
	CaptureStderr bool
	Checkpoints   bool
	Verbose       bool
}

// Wrapper wraps a command for tracing.
type Wrapper struct {
	config       Config
	trace        *agenttrace.Trace
	watcher      *fsnotify.Watcher
	outputBuffer strings.Builder
	errorBuffer  strings.Builder
	mu           sync.Mutex
}

// New creates a new wrapper.
func New(config Config) *Wrapper {
	return &Wrapper{
		config: config,
	}
}

// Run runs the wrapped command.
func (w *Wrapper) Run(ctx context.Context) (int, error) {
	// Determine trace name
	name := w.config.TraceName
	if name == "" {
		name = w.config.Command
	}

	// Build metadata
	metadata := map[string]any{
		"command": w.config.Command,
		"args":    w.config.Args,
		"cwd":     w.getCwd(),
	}

	// Add git info if enabled
	if w.config.GitAutoLink {
		gitInfo := w.getGitInfo()
		for k, v := range gitInfo {
			metadata["git."+k] = v
		}
	}

	// Create trace
	w.trace = w.config.Client.Trace(ctx, agenttrace.TraceOptions{
		Name:      name,
		UserID:    w.config.UserID,
		SessionID: w.config.SessionID,
		Tags:      w.config.Tags,
		Metadata:  metadata,
		Input: map[string]any{
			"command": strings.Join(append([]string{w.config.Command}, w.config.Args...), " "),
		},
	})

	// Start file watcher if enabled
	if w.config.WatchFiles {
		if err := w.startWatcher(ctx); err != nil {
			w.log("Failed to start file watcher: %v", err)
		}
	}

	// Create command
	cmd := exec.CommandContext(ctx, w.config.Command, w.config.Args...)
	cmd.Stdin = os.Stdin

	// Set up output capture
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return 1, fmt.Errorf("failed to get stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return 1, fmt.Errorf("failed to get stderr pipe: %w", err)
	}

	// Start command
	startTime := time.Now()
	if err := cmd.Start(); err != nil {
		w.endTrace(1, time.Since(startTime), err)
		return 1, fmt.Errorf("failed to start command: %w", err)
	}

	// Capture output
	var wg sync.WaitGroup
	wg.Add(2)

	go w.captureOutput(&wg, stdout, os.Stdout, &w.outputBuffer, w.config.CaptureStdout)
	go w.captureOutput(&wg, stderr, os.Stderr, &w.errorBuffer, w.config.CaptureStderr)

	// Wait for output capture
	wg.Wait()

	// Wait for command
	err = cmd.Wait()
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

	// Stop watcher
	if w.watcher != nil {
		w.watcher.Close()
	}

	// End trace
	w.endTrace(exitCode, duration, err)

	return exitCode, nil
}

func (w *Wrapper) captureOutput(wg *sync.WaitGroup, reader io.Reader, writer io.Writer, buffer *strings.Builder, capture bool) {
	defer wg.Done()

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()

		// Write to original output
		fmt.Fprintln(writer, line)

		// Capture if enabled
		if capture {
			w.mu.Lock()
			buffer.WriteString(line)
			buffer.WriteString("\n")
			w.mu.Unlock()
		}
	}
}

func (w *Wrapper) endTrace(exitCode int, duration time.Duration, err error) {
	output := map[string]any{
		"exit_code":   exitCode,
		"duration_ms": duration.Milliseconds(),
	}

	w.mu.Lock()
	if w.config.CaptureStdout && w.outputBuffer.Len() > 0 {
		// Limit output size
		stdout := w.outputBuffer.String()
		if len(stdout) > 10000 {
			stdout = stdout[:10000] + "... (truncated)"
		}
		output["stdout"] = stdout
	}

	if w.config.CaptureStderr && w.errorBuffer.Len() > 0 {
		stderr := w.errorBuffer.String()
		if len(stderr) > 10000 {
			stderr = stderr[:10000] + "... (truncated)"
		}
		output["stderr"] = stderr
	}
	w.mu.Unlock()

	if err != nil && exitCode != 0 {
		output["error"] = err.Error()
	}

	w.trace.End(&agenttrace.TraceEndOptions{Output: output})
}

func (w *Wrapper) startWatcher(ctx context.Context) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	w.watcher = watcher

	// Watch current directory
	cwd := w.getCwd()
	if err := watcher.Add(cwd); err != nil {
		return err
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&(fsnotify.Write|fsnotify.Create) != 0 {
					w.handleFileChange(event.Name)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				w.log("Watcher error: %v", err)
			}
		}
	}()

	return nil
}

func (w *Wrapper) handleFileChange(path string) {
	// Skip hidden files and common build artifacts
	base := filepath.Base(path)
	if strings.HasPrefix(base, ".") ||
		strings.HasSuffix(base, ".pyc") ||
		strings.HasSuffix(base, ".swp") {
		return
	}

	if w.config.Checkpoints {
		w.createCheckpoint(path)
	}

	w.log("File changed: %s", path)
}

func (w *Wrapper) createCheckpoint(path string) {
	span := w.trace.Span(agenttrace.SpanOptions{
		Name: "file-checkpoint",
		Metadata: map[string]any{
			"file": path,
		},
	})

	span.End(&agenttrace.SpanEndOptions{
		Output: map[string]any{
			"file": path,
			"time": time.Now().UTC().Format(time.RFC3339),
		},
	})
}

func (w *Wrapper) getCwd() string {
	cwd, err := os.Getwd()
	if err != nil {
		return ""
	}
	return cwd
}

func (w *Wrapper) getGitInfo() map[string]string {
	info := make(map[string]string)

	// Get current branch
	if branch, err := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD").Output(); err == nil {
		info["branch"] = strings.TrimSpace(string(branch))
	}

	// Get current commit
	if commit, err := exec.Command("git", "rev-parse", "HEAD").Output(); err == nil {
		info["commit"] = strings.TrimSpace(string(commit))
	}

	// Get repo root
	if root, err := exec.Command("git", "rev-parse", "--show-toplevel").Output(); err == nil {
		info["root"] = strings.TrimSpace(string(root))
	}

	// Check for uncommitted changes
	if status, err := exec.Command("git", "status", "--porcelain").Output(); err == nil {
		info["dirty"] = fmt.Sprintf("%v", len(status) > 0)
	}

	return info
}

func (w *Wrapper) log(format string, args ...any) {
	if w.config.Verbose {
		fmt.Fprintf(os.Stderr, "[agenttrace] "+format+"\n", args...)
	}
}

func generateID() string {
	return uuid.New().String()
}
