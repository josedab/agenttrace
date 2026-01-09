// Package wrapper provides file operation tracking functionality.
package wrapper

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	agenttrace "github.com/agenttrace/agenttrace-go"
	"github.com/fsnotify/fsnotify"
)

// FileTracker tracks file operations during command execution.
type FileTracker struct {
	trace       *agenttrace.Trace
	watcher     *fsnotify.Watcher
	watchedDirs map[string]bool
	operations  []*agenttrace.FileOperationInfo
}

// NewFileTracker creates a new file tracker.
func NewFileTracker(trace *agenttrace.Trace) (*FileTracker, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	return &FileTracker{
		trace:       trace,
		watcher:     watcher,
		watchedDirs: make(map[string]bool),
		operations:  make([]*agenttrace.FileOperationInfo, 0),
	}, nil
}

// Watch adds a path to watch for file changes.
func (t *FileTracker) Watch(path string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return err
	}

	if info.IsDir() {
		return t.watchDir(absPath)
	}

	return t.watcher.Add(absPath)
}

// Start starts tracking file changes.
func (t *FileTracker) Start(stopCh <-chan struct{}) {
	go func() {
		for {
			select {
			case <-stopCh:
				return
			case event, ok := <-t.watcher.Events:
				if !ok {
					return
				}
				t.handleEvent(event)
			case _, ok := <-t.watcher.Errors:
				if !ok {
					return
				}
				// Log error but continue
			}
		}
	}()
}

// Stop stops the file tracker.
func (t *FileTracker) Stop() {
	t.watcher.Close()
}

// GetOperations returns all tracked file operations.
func (t *FileTracker) GetOperations() []*agenttrace.FileOperationInfo {
	return t.operations
}

func (t *FileTracker) watchDir(dir string) error {
	if t.watchedDirs[dir] {
		return nil
	}

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		// Skip hidden directories
		if info.IsDir() {
			base := filepath.Base(path)
			if strings.HasPrefix(base, ".") && path != dir {
				return filepath.SkipDir
			}

			// Skip common build/dependency directories
			switch base {
			case "node_modules", "__pycache__", "venv", ".venv", "vendor", "dist", "build":
				return filepath.SkipDir
			}

			if err := t.watcher.Add(path); err == nil {
				t.watchedDirs[path] = true
			}
		}

		return nil
	})

	return err
}

func (t *FileTracker) handleEvent(event fsnotify.Event) {
	// Skip temporary files
	base := filepath.Base(event.Name)
	if strings.HasPrefix(base, ".") ||
		strings.HasSuffix(base, ".swp") ||
		strings.HasSuffix(base, ".tmp") ||
		strings.HasSuffix(base, "~") {
		return
	}

	var opType agenttrace.FileOperationType
	switch {
	case event.Op&fsnotify.Create != 0:
		opType = agenttrace.FileOpCreate
	case event.Op&fsnotify.Write != 0:
		opType = agenttrace.FileOpUpdate
	case event.Op&fsnotify.Remove != 0:
		opType = agenttrace.FileOpDelete
	case event.Op&fsnotify.Rename != 0:
		opType = agenttrace.FileOpRename
	case event.Op&fsnotify.Chmod != 0:
		opType = agenttrace.FileOpChmod
	default:
		return
	}

	now := time.Now().UTC()
	info := t.trace.FileOp(agenttrace.FileOperationOptions{
		Operation:   opType,
		FilePath:    event.Name,
		ToolName:    "fsnotify",
		StartedAt:   &now,
		CompletedAt: &now,
	})

	t.operations = append(t.operations, info)

	// If a new directory was created, watch it
	if opType == agenttrace.FileOpCreate {
		if fi, err := os.Stat(event.Name); err == nil && fi.IsDir() {
			t.watchDir(event.Name)
		}
	}
}
