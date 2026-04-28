package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	apperrors "github.com/agenttrace/agenttrace/api/internal/pkg/errors"
)

var gitCommitSHA = regexp.MustCompile(`^[0-9a-fA-F]{7,64}$`)

// GitHubReportingConfig keeps GitHub optional for core usage.
type GitHubReportingConfig struct {
	Enabled bool
	APIURL  string
	Token   string
}

// GitHubReportingRepository defines project-scoped repository lookup.
type GitHubReportingRepository interface {
	GetProjectRepository(
		ctx context.Context,
		projectID, repositoryID uuid.UUID,
	) (*domain.GitHubRepository, error)
}

// GitHubReportClient delivers rendered reports.
type GitHubReportClient interface {
	CreateIssueComment(
		ctx context.Context,
		repository string,
		pullRequestNumber int,
		body string,
	) error
	CreateCommitStatus(
		ctx context.Context,
		repository, commitSHA, state, description string,
	) error
}

// OutcomeOverviewProvider supplies canonical outcome data.
type OutcomeOverviewProvider interface {
	GetOverview(
		ctx context.Context,
		projectID uuid.UUID,
		from, to time.Time,
	) (*domain.OutcomeOverview, error)
}

// GitHubOutcomeReporter renders and optionally delivers outcome reports.
type GitHubOutcomeReporter struct {
	repository GitHubReportingRepository
	outcomes   OutcomeOverviewProvider
	client     GitHubReportClient
	enabled    bool
	egress     OutboundGuard
	clock      func() time.Time
}

// NewGitHubOutcomeReporter creates an optional GitHub reporter.
func NewGitHubOutcomeReporter(
	repository GitHubReportingRepository,
	outcomes OutcomeOverviewProvider,
	config GitHubReportingConfig,
	guards ...OutboundGuard,
) *GitHubOutcomeReporter {
	var client GitHubReportClient
	if config.Enabled && config.Token != "" {
		client = NewGitHubRESTClient(config.APIURL, config.Token)
	}
	var policy OutboundGuard
	if len(guards) > 0 {
		policy = guards[0]
	}
	return &GitHubOutcomeReporter{
		repository: repository,
		outcomes:   outcomes,
		client:     client,
		enabled:    config.Enabled,
		egress:     policy,
		clock:      time.Now,
	}
}

// Deliver creates a PR comment or commit status when GitHub reporting is configured.
func (s *GitHubOutcomeReporter) Deliver(
	ctx context.Context,
	projectID uuid.UUID,
	input domain.GitHubOutcomeReportInput,
) (*domain.GitHubOutcomeReportResult, error) {
	if err := RequireOutbound(s.egress, EgressGitHub); err != nil {
		return nil, err
	}
	if !s.enabled || s.client == nil {
		return nil, apperrors.Unprocessable(
			"GitHub reporting is not configured; core outcome analytics remains available",
		)
	}
	if input.RepositoryID == uuid.Nil {
		return nil, apperrors.Validation("repositoryId is required")
	}
	if input.PullRequestNumber <= 0 && strings.TrimSpace(input.CommitSHA) == "" {
		return nil, apperrors.Validation("commitSha or pullRequestNumber is required")
	}

	repository, err := s.repository.GetProjectRepository(
		ctx,
		projectID,
		input.RepositoryID,
	)
	if err != nil {
		return nil, err
	}
	from, to, err := outcomeReportWindow(s.clock().UTC(), input.Window)
	if err != nil {
		return nil, err
	}
	overview, err := s.outcomes.GetOverview(ctx, projectID, from, to)
	if err != nil {
		return nil, err
	}
	digest := BuildOutcomeDigest(overview)
	rendered := RenderOutcomeDigestMarkdown(digest)

	target := ""
	if input.PullRequestNumber > 0 {
		if err := s.client.CreateIssueComment(
			ctx,
			repository.RepoFullName,
			input.PullRequestNumber,
			rendered,
		); err != nil {
			return nil, fmt.Errorf("deliver GitHub pull request report: %w", err)
		}
		target = fmt.Sprintf("pull_request:%d", input.PullRequestNumber)
	} else {
		if !gitCommitSHA.MatchString(input.CommitSHA) {
			return nil, apperrors.Validation("commitSha must be a 7-64 character hexadecimal SHA")
		}
		state := githubCommitStatusState(overview)
		description := digest.Summary
		if len(description) > 140 {
			description = description[:137] + "..."
		}
		if err := s.client.CreateCommitStatus(
			ctx,
			repository.RepoFullName,
			input.CommitSHA,
			state,
			description,
		); err != nil {
			return nil, fmt.Errorf("deliver GitHub commit report: %w", err)
		}
		target = "commit:" + input.CommitSHA
	}

	return &domain.GitHubOutcomeReportResult{
		RepositoryID: input.RepositoryID,
		Target:       target,
		DeliveredAt:  s.clock().UTC(),
	}, nil
}

// githubCommitStatusState maps an outcome overview onto a GitHub commit status
// state ("success", "failure", or "pending").
//
//   - Any observed failure or regression signal is a hard "failure".
//   - When there is no run data at all (success rate unavailable), or when runs
//     are still in progress and none have completed yet, the state stays
//     "pending" rather than prematurely reporting success.
//   - Otherwise, with at least one completed run and no failures, it is
//     "success".
func githubCommitStatusState(overview *domain.OutcomeOverview) string {
	runs := overview.Runs
	if runs.Failed > 0 || overview.SourceControl.RegressionSignals > 0 {
		return "failure"
	}
	// Use the explicit terminal counters rather than deriving completion from
	// Total, which may be absent or temporarily inconsistent while aggregation
	// is in progress. No completed outcome (including an entirely empty window)
	// remains pending.
	if !runs.SuccessRate.Available || (runs.Successful == 0 && runs.Failed == 0) {
		return "pending"
	}
	return "success"
}

func outcomeReportWindow(
	now time.Time,
	window string,
) (from, to time.Time, resultErr error) {
	switch window {
	case "", "7d":
		return now.AddDate(0, 0, -7), now, nil
	case "24h":
		return now.Add(-24 * time.Hour), now, nil
	case "30d":
		return now.AddDate(0, 0, -30), now, nil
	default:
		return time.Time{}, time.Time{}, apperrors.Validation(
			"window must be 24h, 7d, or 30d",
		)
	}
}

// GitHubRESTClient is a minimal GitHub delivery client.
type GitHubRESTClient struct {
	apiURL string
	token  string
	client *http.Client
}

// NewGitHubRESTClient creates a GitHub REST client.
func NewGitHubRESTClient(apiURL, token string) *GitHubRESTClient {
	if strings.TrimSpace(apiURL) == "" {
		apiURL = "https://api.github.com"
	}
	return &GitHubRESTClient{
		apiURL: strings.TrimRight(apiURL, "/"),
		token:  token,
		client: &http.Client{Timeout: 15 * time.Second},
	}
}

// CreateIssueComment posts a report to a pull request conversation.
func (c *GitHubRESTClient) CreateIssueComment(
	ctx context.Context,
	repository string,
	pullRequestNumber int,
	body string,
) error {
	path, err := githubRepositoryPath(repository)
	if err != nil {
		return err
	}
	return c.post(ctx, path+"/issues/"+strconv.Itoa(pullRequestNumber)+"/comments", map[string]string{
		"body": body,
	})
}

// CreateCommitStatus posts a bounded workflow status.
func (c *GitHubRESTClient) CreateCommitStatus(
	ctx context.Context,
	repository, commitSHA, state, description string,
) error {
	path, err := githubRepositoryPath(repository)
	if err != nil {
		return err
	}
	if strings.TrimSpace(commitSHA) == "" {
		return apperrors.Validation("commit SHA is required")
	}
	if !gitCommitSHA.MatchString(commitSHA) {
		return apperrors.Validation("commit SHA must be hexadecimal")
	}
	return c.post(ctx, path+"/statuses/"+url.PathEscape(commitSHA), map[string]string{
		"state":       state,
		"description": description,
		"context":     "AgentTrace/outcomes",
	})
}

func (c *GitHubRESTClient) post(
	ctx context.Context,
	path string,
	payload interface{},
) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.apiURL+path,
		bytes.NewReader(body),
	)
	if err != nil {
		return err
	}
	request.Header.Set("Authorization", "Bearer "+c.token)
	request.Header.Set("Accept", "application/vnd.github+json")
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	response, err := c.client.Do(request)
	if err != nil {
		return err
	}
	responseBody, readErr := io.ReadAll(io.LimitReader(response.Body, 4096))
	closeErr := response.Body.Close()
	if readErr != nil {
		return fmt.Errorf("read GitHub API response: %w", readErr)
	}
	if closeErr != nil {
		return fmt.Errorf("close GitHub API response: %w", closeErr)
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf(
			"GitHub API returned %d: %s",
			response.StatusCode,
			strings.TrimSpace(string(responseBody)),
		)
	}
	return nil
}

func githubRepositoryPath(repository string) (string, error) {
	owner, name, found := strings.Cut(repository, "/")
	if !found || owner == "" || name == "" || strings.Contains(name, "/") {
		return "", apperrors.Validation("invalid GitHub repository name")
	}
	return "/repos/" + url.PathEscape(owner) + "/" + url.PathEscape(name), nil
}
