// Package wrapper provides git auto-linking functionality.
package wrapper

import (
	"os/exec"
	"strings"

	agenttrace "github.com/agenttrace/agenttrace-go"
)

// GitLinker manages git links during command execution.
type GitLinker struct {
	trace         *agenttrace.Trace
	initialCommit string
	initialBranch string
}

// NewGitLinker creates a new git linker.
func NewGitLinker(trace *agenttrace.Trace) *GitLinker {
	linker := &GitLinker{
		trace: trace,
	}
	linker.captureInitialState()
	return linker
}

// LinkStart creates a git link for the start of execution.
func (l *GitLinker) LinkStart() *agenttrace.GitLinkInfo {
	return l.trace.GitLink(&agenttrace.GitLinkOptions{
		Type:       agenttrace.GitLinkTypeStart,
		AutoDetect: true,
	})
}

// LinkEnd creates a git link for the end of execution.
func (l *GitLinker) LinkEnd() *agenttrace.GitLinkInfo {
	currentCommit := l.getCurrentCommit()
	if currentCommit != l.initialCommit {
		// A new commit was made during execution
		return l.trace.GitLink(&agenttrace.GitLinkOptions{
			Type:       agenttrace.GitLinkTypeCommit,
			AutoDetect: true,
		})
	}
	return nil
}

// LinkBranchChange creates a git link when branch changes.
func (l *GitLinker) LinkBranchChange() *agenttrace.GitLinkInfo {
	currentBranch := l.getCurrentBranch()
	if currentBranch != l.initialBranch && currentBranch != "" {
		return l.trace.GitLink(&agenttrace.GitLinkOptions{
			Type:       agenttrace.GitLinkTypeBranch,
			AutoDetect: true,
		})
	}
	return nil
}

// CheckAndLink checks for git changes and creates appropriate links.
func (l *GitLinker) CheckAndLink() []*agenttrace.GitLinkInfo {
	var links []*agenttrace.GitLinkInfo

	// Check for branch change
	if link := l.LinkBranchChange(); link != nil {
		links = append(links, link)
		l.initialBranch = l.getCurrentBranch()
	}

	// Check for new commit
	currentCommit := l.getCurrentCommit()
	if currentCommit != l.initialCommit && currentCommit != "" {
		link := l.trace.GitLink(&agenttrace.GitLinkOptions{
			Type:       agenttrace.GitLinkTypeCommit,
			AutoDetect: true,
		})
		links = append(links, link)
		l.initialCommit = currentCommit
	}

	return links
}

// HasUncommittedChanges returns true if there are uncommitted changes.
func (l *GitLinker) HasUncommittedChanges() bool {
	out, err := exec.Command("git", "status", "--porcelain").Output()
	if err != nil {
		return false
	}
	return len(strings.TrimSpace(string(out))) > 0
}

// GetDiffSummary returns a summary of current changes.
func (l *GitLinker) GetDiffSummary() map[string]any {
	summary := make(map[string]any)

	// Get staged files
	if out, err := exec.Command("git", "diff", "--cached", "--name-only").Output(); err == nil {
		files := strings.Split(strings.TrimSpace(string(out)), "\n")
		if len(files) > 0 && files[0] != "" {
			summary["staged_files"] = files
		}
	}

	// Get modified files
	if out, err := exec.Command("git", "diff", "--name-only").Output(); err == nil {
		files := strings.Split(strings.TrimSpace(string(out)), "\n")
		if len(files) > 0 && files[0] != "" {
			summary["modified_files"] = files
		}
	}

	// Get untracked files
	if out, err := exec.Command("git", "ls-files", "--others", "--exclude-standard").Output(); err == nil {
		files := strings.Split(strings.TrimSpace(string(out)), "\n")
		if len(files) > 0 && files[0] != "" {
			summary["untracked_files"] = files
		}
	}

	return summary
}

func (l *GitLinker) captureInitialState() {
	l.initialCommit = l.getCurrentCommit()
	l.initialBranch = l.getCurrentBranch()
}

func (l *GitLinker) getCurrentCommit() string {
	if out, err := exec.Command("git", "rev-parse", "HEAD").Output(); err == nil {
		return strings.TrimSpace(string(out))
	}
	return ""
}

func (l *GitLinker) getCurrentBranch() string {
	if out, err := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD").Output(); err == nil {
		return strings.TrimSpace(string(out))
	}
	return ""
}
