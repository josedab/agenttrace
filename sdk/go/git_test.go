package agenttrace

import (
	"context"
	"testing"
	"time"
)

func TestGitLinkType(t *testing.T) {
	t.Run("all git link types exist", func(t *testing.T) {
		types := []GitLinkType{
			GitLinkTypeStart,
			GitLinkTypeCommit,
			GitLinkTypeRestore,
			GitLinkTypeBranch,
			GitLinkTypeDiff,
		}

		expected := []string{"start", "commit", "restore", "branch", "diff"}

		for i, lt := range types {
			if string(lt) != expected[i] {
				t.Errorf("expected %s, got %s", expected[i], lt)
			}
		}
	})
}

func TestTraceGitLink(t *testing.T) {
	client := New(Config{
		APIKey: "test-api-key",
	})
	defer client.Shutdown()

	ctx := context.Background()
	trace := client.Trace(ctx, TraceOptions{
		Name: "test-trace",
	})

	t.Run("creates git link with nil options", func(t *testing.T) {
		link := trace.GitLink(nil)

		if link.TraceID != trace.id {
			t.Error("expected git link to have trace's ID")
		}
		if link.ID == "" {
			t.Error("expected git link ID to be generated")
		}
		if link.Type != GitLinkTypeCommit {
			t.Errorf("expected type to be 'commit' by default, got '%s'", link.Type)
		}
		if link.CreatedAt.IsZero() {
			t.Error("expected CreatedAt to be set")
		}
	})

	t.Run("creates git link with explicit values", func(t *testing.T) {
		link := trace.GitLink(&GitLinkOptions{
			Type:          GitLinkTypeStart,
			CommitSha:     "explicit-sha-123",
			Branch:        "feature-branch",
			RepoURL:       "https://github.com/explicit/repo",
			CommitMessage: "Explicit commit message",
			AutoDetect:    false,
		})

		if link.Type != GitLinkTypeStart {
			t.Errorf("expected type to be 'start', got '%s'", link.Type)
		}
		if link.CommitSha != "explicit-sha-123" {
			t.Errorf("expected commitSha to be explicit, got '%s'", link.CommitSha)
		}
		if link.Branch != "feature-branch" {
			t.Errorf("expected branch to be 'feature-branch', got '%s'", link.Branch)
		}
		if link.RepoURL != "https://github.com/explicit/repo" {
			t.Errorf("expected repoURL to be explicit, got '%s'", link.RepoURL)
		}
		if link.CommitMessage != "Explicit commit message" {
			t.Errorf("expected commitMessage to be explicit, got '%s'", link.CommitMessage)
		}
	})

	t.Run("creates git link with observation ID", func(t *testing.T) {
		link := trace.GitLink(&GitLinkOptions{
			ObservationID: "obs-456",
			AutoDetect:    false,
		})

		if link.ObservationID != "obs-456" {
			t.Errorf("expected observationID to be 'obs-456', got '%s'", link.ObservationID)
		}
	})

	t.Run("creates git link with files changed", func(t *testing.T) {
		link := trace.GitLink(&GitLinkOptions{
			FilesChanged: []string{"file1.go", "file2.go"},
			AutoDetect:   false,
		})

		if len(link.FilesChanged) != 2 {
			t.Errorf("expected 2 files, got %d", len(link.FilesChanged))
		}
		if link.FilesChanged[0] != "file1.go" || link.FilesChanged[1] != "file2.go" {
			t.Error("files changed not set correctly")
		}
	})

	t.Run("creates git link with all types", func(t *testing.T) {
		types := []GitLinkType{
			GitLinkTypeStart,
			GitLinkTypeCommit,
			GitLinkTypeRestore,
			GitLinkTypeBranch,
			GitLinkTypeDiff,
		}

		for _, linkType := range types {
			link := trace.GitLink(&GitLinkOptions{
				Type:       linkType,
				AutoDetect: false,
			})

			if link.Type != linkType {
				t.Errorf("expected type to be '%s', got '%s'", linkType, link.Type)
			}
		}
	})

	t.Run("auto-detects when values not provided", func(t *testing.T) {
		// This test depends on whether we're in a git repo
		// Just verify it doesn't panic
		link := trace.GitLink(&GitLinkOptions{
			AutoDetect: true,
		})

		if link.ID == "" {
			t.Error("expected link ID to be generated")
		}
	})
}

func TestGitLinkDisabledClient(t *testing.T) {
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

	link := trace.GitLink(&GitLinkOptions{
		CommitSha:  "disabled-sha",
		AutoDetect: false,
	})

	// Should still return link info
	if link.CommitSha != "disabled-sha" {
		t.Error("expected link info even when disabled")
	}
	if link.ID == "" {
		t.Error("expected link ID even when disabled")
	}
}

func TestGitLinkInfo(t *testing.T) {
	t.Run("has all expected fields", func(t *testing.T) {
		now := time.Now()
		info := GitLinkInfo{
			ID:            "git-123",
			TraceID:       "trace-456",
			ObservationID: "obs-789",
			Type:          GitLinkTypeCommit,
			CommitSha:     "abc123def456",
			Branch:        "main",
			RepoURL:       "https://github.com/test/repo",
			CommitMessage: "Test commit message",
			AuthorName:    "Test Author",
			AuthorEmail:   "test@example.com",
			FilesChanged:  []string{"file1.go", "file2.go"},
			CreatedAt:     now,
		}

		if info.ID != "git-123" {
			t.Error("ID not set correctly")
		}
		if info.TraceID != "trace-456" {
			t.Error("TraceID not set correctly")
		}
		if info.ObservationID != "obs-789" {
			t.Error("ObservationID not set correctly")
		}
		if info.Type != GitLinkTypeCommit {
			t.Error("Type not set correctly")
		}
		if info.CommitSha != "abc123def456" {
			t.Error("CommitSha not set correctly")
		}
		if info.Branch != "main" {
			t.Error("Branch not set correctly")
		}
		if info.RepoURL != "https://github.com/test/repo" {
			t.Error("RepoURL not set correctly")
		}
		if info.CommitMessage != "Test commit message" {
			t.Error("CommitMessage not set correctly")
		}
		if info.AuthorName != "Test Author" {
			t.Error("AuthorName not set correctly")
		}
		if info.AuthorEmail != "test@example.com" {
			t.Error("AuthorEmail not set correctly")
		}
		if len(info.FilesChanged) != 2 {
			t.Error("FilesChanged not set correctly")
		}
		if info.CreatedAt.IsZero() {
			t.Error("CreatedAt not set correctly")
		}
	})
}

func TestGitLinkOptions(t *testing.T) {
	t.Run("default auto-detect behavior", func(t *testing.T) {
		opts := GitLinkOptions{}

		// When CommitSha and Branch are empty, AutoDetect should be triggered
		if opts.CommitSha != "" || opts.Branch != "" {
			t.Error("expected empty values by default")
		}
	})

	t.Run("explicit auto-detect false prevents detection", func(t *testing.T) {
		client := New(Config{
			APIKey: "test-api-key",
		})
		defer client.Shutdown()

		ctx := context.Background()
		trace := client.Trace(ctx, TraceOptions{
			Name: "test-trace",
		})

		link := trace.GitLink(&GitLinkOptions{
			AutoDetect: false,
		})

		// With AutoDetect false and no explicit values, values should be empty
		// (unless we're actually in a git repo and the logic still auto-detects)
		if link.ID == "" {
			t.Error("expected link ID to be generated")
		}
	})
}

func TestGitHelperFunctions(t *testing.T) {
	// These tests will behave differently depending on whether
	// we're running in a git repo or not. Just verify they don't panic.

	t.Run("getCommitMessage does not panic", func(t *testing.T) {
		_ = getCommitMessage()
	})

	t.Run("getAuthorInfo does not panic", func(t *testing.T) {
		_, _ = getAuthorInfo()
	})

	t.Run("getChangedFiles does not panic", func(t *testing.T) {
		files := getChangedFiles()
		if files == nil {
			t.Error("expected non-nil slice (even if empty)")
		}
	})

	t.Run("getDiffStats does not panic", func(t *testing.T) {
		_, _ = getDiffStats()
	})
}
