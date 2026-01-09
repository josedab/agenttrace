package agenttrace

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/google/uuid"
)

// CheckpointType represents the type of checkpoint.
type CheckpointType string

const (
	CheckpointTypeManual   CheckpointType = "manual"
	CheckpointTypeAuto     CheckpointType = "auto"
	CheckpointTypeToolCall CheckpointType = "tool_call"
	CheckpointTypeError    CheckpointType = "error"
	CheckpointTypeMilestone CheckpointType = "milestone"
	CheckpointTypeRestore  CheckpointType = "restore"
)

// CheckpointOptions holds options for creating a checkpoint.
type CheckpointOptions struct {
	Name           string
	Type           CheckpointType
	ObservationID  string
	Description    string
	Files          []string
	IncludeGitInfo bool
}

// CheckpointInfo contains information about a created checkpoint.
type CheckpointInfo struct {
	ID             string
	Name           string
	Type           CheckpointType
	TraceID        string
	ObservationID  string
	GitCommitSha   string
	GitBranch      string
	FilesChanged   []string
	TotalFiles     int
	TotalSizeBytes int64
	CreatedAt      time.Time
}

// GitInfo contains git repository information.
type GitInfo struct {
	CommitSha string
	Branch    string
	RepoURL   string
}

// Checkpoint creates a checkpoint for this trace.
func (t *Trace) Checkpoint(opts CheckpointOptions) *CheckpointInfo {
	checkpointID := uuid.New().String()
	now := time.Now().UTC()

	if opts.Type == "" {
		opts.Type = CheckpointTypeManual
	}

	// Include git info by default
	if opts.IncludeGitInfo || opts.IncludeGitInfo == false && len(opts.Files) == 0 {
		opts.IncludeGitInfo = true
	}

	var gitInfo GitInfo
	if opts.IncludeGitInfo {
		gitInfo = getGitInfo()
	}

	// Calculate file info
	filesChanged := opts.Files
	if filesChanged == nil {
		filesChanged = []string{}
	}
	totalFiles := len(filesChanged)
	var totalSizeBytes int64

	filesSnapshot := make(map[string]map[string]interface{})
	for _, filePath := range filesChanged {
		if info, err := os.Stat(filePath); err == nil {
			totalSizeBytes += info.Size()

			if f, err := os.Open(filePath); err == nil {
				h := sha256.New()
				io.Copy(h, f)
				f.Close()
				hash := hex.EncodeToString(h.Sum(nil))
				filesSnapshot[filePath] = map[string]interface{}{
					"size": info.Size(),
					"hash": hash,
				}
			}
		}
	}

	// Send to API
	if t.client.Enabled() {
		t.client.addEvent(map[string]any{
			"type": "checkpoint-create",
			"body": map[string]any{
				"id":             checkpointID,
				"traceId":        t.id,
				"observationId":  opts.ObservationID,
				"name":           opts.Name,
				"description":    opts.Description,
				"type":           string(opts.Type),
				"gitCommitSha":   gitInfo.CommitSha,
				"gitBranch":      gitInfo.Branch,
				"gitRepoUrl":     gitInfo.RepoURL,
				"filesSnapshot":  filesSnapshot,
				"filesChanged":   filesChanged,
				"totalFiles":     totalFiles,
				"totalSizeBytes": totalSizeBytes,
				"timestamp":      now.Format(time.RFC3339Nano),
			},
		})
	}

	return &CheckpointInfo{
		ID:             checkpointID,
		Name:           opts.Name,
		Type:           opts.Type,
		TraceID:        t.id,
		ObservationID:  opts.ObservationID,
		GitCommitSha:   gitInfo.CommitSha,
		GitBranch:      gitInfo.Branch,
		FilesChanged:   filesChanged,
		TotalFiles:     totalFiles,
		TotalSizeBytes: totalSizeBytes,
		CreatedAt:      now,
	}
}

func getGitInfo() GitInfo {
	var info GitInfo

	if out, err := exec.Command("git", "rev-parse", "HEAD").Output(); err == nil {
		info.CommitSha = strings.TrimSpace(string(out))
	}

	if out, err := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD").Output(); err == nil {
		info.Branch = strings.TrimSpace(string(out))
	}

	if out, err := exec.Command("git", "config", "--get", "remote.origin.url").Output(); err == nil {
		info.RepoURL = strings.TrimSpace(string(out))
	}

	return info
}
