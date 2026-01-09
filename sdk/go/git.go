package agenttrace

import (
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

// GitLinkType represents the type of git link.
type GitLinkType string

const (
	GitLinkTypeStart   GitLinkType = "start"
	GitLinkTypeCommit  GitLinkType = "commit"
	GitLinkTypeRestore GitLinkType = "restore"
	GitLinkTypeBranch  GitLinkType = "branch"
	GitLinkTypeDiff    GitLinkType = "diff"
)

// GitLinkOptions holds options for creating a git link.
type GitLinkOptions struct {
	Type          GitLinkType
	ObservationID string
	CommitSha     string
	Branch        string
	RepoURL       string
	CommitMessage string
	FilesChanged  []string
	AutoDetect    bool
}

// GitLinkInfo contains information about a created git link.
type GitLinkInfo struct {
	ID            string
	TraceID       string
	ObservationID string
	Type          GitLinkType
	CommitSha     string
	Branch        string
	RepoURL       string
	CommitMessage string
	AuthorName    string
	AuthorEmail   string
	FilesChanged  []string
	CreatedAt     time.Time
}

// GitLink creates a git link for this trace.
func (t *Trace) GitLink(opts *GitLinkOptions) *GitLinkInfo {
	if opts == nil {
		opts = &GitLinkOptions{}
	}

	linkID := uuid.New().String()
	now := time.Now().UTC()

	if opts.Type == "" {
		opts.Type = GitLinkTypeCommit
	}

	// Auto-detect by default
	autoDetect := opts.AutoDetect || (opts.CommitSha == "" && opts.Branch == "")

	var commitSha, branch, repoURL, commitMessage string
	var authorName, authorEmail string
	var filesChanged []string

	if autoDetect {
		gitInfo := getGitInfo()
		if opts.CommitSha == "" {
			commitSha = gitInfo.CommitSha
		} else {
			commitSha = opts.CommitSha
		}
		if opts.Branch == "" {
			branch = gitInfo.Branch
		} else {
			branch = opts.Branch
		}
		if opts.RepoURL == "" {
			repoURL = gitInfo.RepoURL
		} else {
			repoURL = opts.RepoURL
		}

		if opts.CommitMessage == "" {
			commitMessage = getCommitMessage()
		} else {
			commitMessage = opts.CommitMessage
		}

		authorName, authorEmail = getAuthorInfo()

		if opts.FilesChanged == nil {
			filesChanged = getChangedFiles()
		} else {
			filesChanged = opts.FilesChanged
		}
	} else {
		commitSha = opts.CommitSha
		branch = opts.Branch
		repoURL = opts.RepoURL
		commitMessage = opts.CommitMessage
		filesChanged = opts.FilesChanged
	}

	if filesChanged == nil {
		filesChanged = []string{}
	}

	// Get diff stats
	additions, deletions := getDiffStats()

	// Send to API
	if t.client.Enabled() {
		t.client.addEvent(map[string]any{
			"type": "git-link-create",
			"body": map[string]any{
				"id":            linkID,
				"traceId":       t.id,
				"observationId": opts.ObservationID,
				"linkType":      string(opts.Type),
				"commitSha":     commitSha,
				"branch":        branch,
				"repoUrl":       repoURL,
				"commitMessage": commitMessage,
				"authorName":    authorName,
				"authorEmail":   authorEmail,
				"filesChanged":  filesChanged,
				"additions":     additions,
				"deletions":     deletions,
				"timestamp":     now.Format(time.RFC3339Nano),
			},
		})
	}

	return &GitLinkInfo{
		ID:            linkID,
		TraceID:       t.id,
		ObservationID: opts.ObservationID,
		Type:          opts.Type,
		CommitSha:     commitSha,
		Branch:        branch,
		RepoURL:       repoURL,
		CommitMessage: commitMessage,
		AuthorName:    authorName,
		AuthorEmail:   authorEmail,
		FilesChanged:  filesChanged,
		CreatedAt:     now,
	}
}

func getCommitMessage() string {
	if out, err := exec.Command("git", "log", "-1", "--format=%s").Output(); err == nil {
		return strings.TrimSpace(string(out))
	}
	return ""
}

func getAuthorInfo() (name, email string) {
	if out, err := exec.Command("git", "log", "-1", "--format=%an").Output(); err == nil {
		name = strings.TrimSpace(string(out))
	}
	if out, err := exec.Command("git", "log", "-1", "--format=%ae").Output(); err == nil {
		email = strings.TrimSpace(string(out))
	}
	return
}

func getChangedFiles() []string {
	files := []string{}
	seen := make(map[string]bool)

	// Staged files
	if out, err := exec.Command("git", "diff", "--cached", "--name-only").Output(); err == nil {
		for _, f := range strings.Split(strings.TrimSpace(string(out)), "\n") {
			if f != "" && !seen[f] {
				files = append(files, f)
				seen[f] = true
			}
		}
	}

	// Unstaged files
	if out, err := exec.Command("git", "diff", "--name-only").Output(); err == nil {
		for _, f := range strings.Split(strings.TrimSpace(string(out)), "\n") {
			if f != "" && !seen[f] {
				files = append(files, f)
				seen[f] = true
			}
		}
	}

	// Untracked files
	if out, err := exec.Command("git", "ls-files", "--others", "--exclude-standard").Output(); err == nil {
		for _, f := range strings.Split(strings.TrimSpace(string(out)), "\n") {
			if f != "" && !seen[f] {
				files = append(files, f)
				seen[f] = true
			}
		}
	}

	return files
}

func getDiffStats() (additions, deletions int) {
	if out, err := exec.Command("git", "diff", "--shortstat").Output(); err == nil {
		output := string(out)
		// Parse: "2 files changed, 10 insertions(+), 5 deletions(-)"
		insertRe := regexp.MustCompile(`(\d+) insertion`)
		deleteRe := regexp.MustCompile(`(\d+) deletion`)

		if match := insertRe.FindStringSubmatch(output); len(match) > 1 {
			additions, _ = strconv.Atoi(match[1])
		}
		if match := deleteRe.FindStringSubmatch(output); len(match) > 1 {
			deletions, _ = strconv.Atoi(match[1])
		}
	}
	return
}
