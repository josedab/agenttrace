// Package wrapper provides checkpoint functionality for tracing.
package wrapper

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"

	agenttrace "github.com/agenttrace/agenttrace-go"
)

// CheckpointManager manages checkpoints during command execution.
type CheckpointManager struct {
	trace      *agenttrace.Trace
	watchPaths []string
	fileHashes map[string]string
}

// NewCheckpointManager creates a new checkpoint manager.
func NewCheckpointManager(trace *agenttrace.Trace, watchPaths []string) *CheckpointManager {
	return &CheckpointManager{
		trace:      trace,
		watchPaths: watchPaths,
		fileHashes: make(map[string]string),
	}
}

// TakeSnapshot takes a snapshot of the current file state.
func (m *CheckpointManager) TakeSnapshot() {
	for _, path := range m.watchPaths {
		m.hashPath(path)
	}
}

// CheckForChanges checks for file changes and creates checkpoints.
func (m *CheckpointManager) CheckForChanges(name string) *agenttrace.CheckpointInfo {
	changedFiles := []string{}

	for _, path := range m.watchPaths {
		if m.hasChanged(path) {
			changedFiles = append(changedFiles, path)
		}
	}

	if len(changedFiles) == 0 {
		return nil
	}

	return m.trace.Checkpoint(agenttrace.CheckpointOptions{
		Name:           name,
		Type:           agenttrace.CheckpointTypeAuto,
		Files:          changedFiles,
		IncludeGitInfo: true,
	})
}

// CreateManualCheckpoint creates a manual checkpoint.
func (m *CheckpointManager) CreateManualCheckpoint(name, description string, files []string) *agenttrace.CheckpointInfo {
	return m.trace.Checkpoint(agenttrace.CheckpointOptions{
		Name:           name,
		Description:    description,
		Type:           agenttrace.CheckpointTypeManual,
		Files:          files,
		IncludeGitInfo: true,
	})
}

func (m *CheckpointManager) hashPath(path string) {
	filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			return nil
		}
		if hash := m.hashFile(p); hash != "" {
			m.fileHashes[p] = hash
		}
		return nil
	})
}

func (m *CheckpointManager) hasChanged(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}

	if info.IsDir() {
		changed := false
		filepath.Walk(path, func(p string, fi os.FileInfo, err error) error {
			if err != nil || fi.IsDir() {
				return nil
			}
			if m.fileChanged(p) {
				changed = true
			}
			return nil
		})
		return changed
	}

	return m.fileChanged(path)
}

func (m *CheckpointManager) fileChanged(path string) bool {
	newHash := m.hashFile(path)
	if newHash == "" {
		return false
	}

	oldHash, exists := m.fileHashes[path]
	if !exists {
		m.fileHashes[path] = newHash
		return true
	}

	if oldHash != newHash {
		m.fileHashes[path] = newHash
		return true
	}

	return false
}

func (m *CheckpointManager) hashFile(path string) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return ""
	}

	return hex.EncodeToString(h.Sum(nil))
}
