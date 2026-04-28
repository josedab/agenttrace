package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	apperrors "github.com/agenttrace/agenttrace/api/internal/pkg/errors"
)

type githubReportingRepositoryStub struct {
	projectID  uuid.UUID
	repository *domain.GitHubRepository
}

func (r *githubReportingRepositoryStub) GetProjectRepository(
	_ context.Context,
	projectID, repositoryID uuid.UUID,
) (*domain.GitHubRepository, error) {
	if r.repository == nil ||
		r.projectID != projectID ||
		r.repository.ID != repositoryID {
		return nil, apperrors.NotFound("GitHub repository")
	}
	return r.repository, nil
}

type outcomeOverviewProviderStub struct {
	overview *domain.OutcomeOverview
	project  uuid.UUID
}

func (p *outcomeOverviewProviderStub) GetOverview(
	_ context.Context,
	projectID uuid.UUID,
	_, _ time.Time,
) (*domain.OutcomeOverview, error) {
	p.project = projectID
	return p.overview, nil
}

type githubReportClientStub struct {
	commentRepository string
	commentPR         int
	commentBody       string
	statusRepository  string
	statusCommit      string
	statusState       string
}

func (c *githubReportClientStub) CreateIssueComment(
	_ context.Context,
	repository string,
	pullRequestNumber int,
	body string,
) error {
	c.commentRepository = repository
	c.commentPR = pullRequestNumber
	c.commentBody = body
	return nil
}

func (c *githubReportClientStub) CreateCommitStatus(
	_ context.Context,
	repository, commitSHA, state, _ string,
) error {
	c.statusRepository = repository
	c.statusCommit = commitSHA
	c.statusState = state
	return nil
}

func TestGitHubOutcomeReporterFailsExplicitlyWhenUnconfigured(t *testing.T) {
	reporter := NewGitHubOutcomeReporter(
		&githubReportingRepositoryStub{},
		&outcomeOverviewProviderStub{},
		GitHubReportingConfig{},
	)

	_, err := reporter.Deliver(
		context.Background(),
		uuid.New(),
		domain.GitHubOutcomeReportInput{RepositoryID: uuid.New(), CommitSHA: "abc1234"},
	)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "not configured")
}

func TestGitHubOutcomeReporterHonorsNoEgress(t *testing.T) {
	reporter := NewGitHubOutcomeReporter(
		&githubReportingRepositoryStub{},
		&outcomeOverviewProviderStub{},
		GitHubReportingConfig{Enabled: true, Token: "configured"},
		NewEgressPolicy(true, true),
	)

	_, err := reporter.Deliver(
		context.Background(),
		uuid.New(),
		domain.GitHubOutcomeReportInput{RepositoryID: uuid.New(), CommitSHA: "abc1234"},
	)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "no-egress")
}

func TestGitHubOutcomeReporterDeliversPullRequestDigest(t *testing.T) {
	projectID := uuid.New()
	repositoryID := uuid.New()
	repository := &githubReportingRepositoryStub{
		projectID: projectID,
		repository: &domain.GitHubRepository{
			ID:           repositoryID,
			ProjectID:    projectID,
			RepoFullName: "agenttrace/example",
		},
	}
	overview := &domain.OutcomeOverview{
		ProjectID: projectID,
		Period: domain.OutcomePeriod{
			From: time.Date(2026, 7, 18, 0, 0, 0, 0, time.UTC),
			To:   time.Date(2026, 7, 25, 0, 0, 0, 0, time.UTC),
		},
		Runs: domain.OutcomeRunMetrics{
			Total:       10,
			Successful:  8,
			SuccessRate: ratioMetric(8, 10),
		},
		GeneratedAt: time.Date(2026, 7, 25, 0, 0, 0, 0, time.UTC),
	}
	provider := &outcomeOverviewProviderStub{overview: overview}
	client := &githubReportClientStub{}
	reporter := NewGitHubOutcomeReporter(
		repository,
		provider,
		GitHubReportingConfig{Enabled: true, Token: "configured"},
	)
	reporter.client = client
	reporter.clock = func() time.Time {
		return time.Date(2026, 7, 25, 12, 0, 0, 0, time.UTC)
	}

	result, err := reporter.Deliver(
		context.Background(),
		projectID,
		domain.GitHubOutcomeReportInput{
			RepositoryID:      repositoryID,
			PullRequestNumber: 42,
			Window:            "7d",
		},
	)

	require.NoError(t, err)
	assert.Equal(t, projectID, provider.project)
	assert.Equal(t, "agenttrace/example", client.commentRepository)
	assert.Equal(t, 42, client.commentPR)
	assert.Contains(t, client.commentBody, "Agent outcome digest")
	assert.Equal(t, "pull_request:42", result.Target)
}

func TestGitHubOutcomeReporterCommitStatusReflectsFailures(t *testing.T) {
	projectID := uuid.New()
	repositoryID := uuid.New()
	client := &githubReportClientStub{}
	reporter := NewGitHubOutcomeReporter(
		&githubReportingRepositoryStub{
			projectID: projectID,
			repository: &domain.GitHubRepository{
				ID:           repositoryID,
				ProjectID:    projectID,
				RepoFullName: "agenttrace/example",
			},
		},
		&outcomeOverviewProviderStub{overview: &domain.OutcomeOverview{
			Runs: domain.OutcomeRunMetrics{Total: 2, Failed: 1},
			SourceControl: domain.OutcomeSCMMetrics{
				RegressionSignals: 1,
			},
		}},
		GitHubReportingConfig{Enabled: true, Token: "configured"},
	)
	reporter.client = client

	_, err := reporter.Deliver(
		context.Background(),
		projectID,
		domain.GitHubOutcomeReportInput{
			RepositoryID: repositoryID,
			CommitSHA:    "abc1234",
		},
	)

	require.NoError(t, err)
	assert.Equal(t, "failure", client.statusState)
	assert.Equal(t, "abc1234", client.statusCommit)
}

func TestGitHubRESTClientUsesBoundedRepositoryPathAndToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/repos/owner/repo/issues/7/comments", r.URL.Path)
		assert.Equal(t, "Bearer secret-token", r.Header.Get("Authorization"))
		var payload map[string]string
		require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		assert.Equal(t, "report", payload["body"])
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	client := NewGitHubRESTClient(server.URL, "secret-token")
	client.client = server.Client()

	err := client.CreateIssueComment(context.Background(), "owner/repo", 7, "report")

	require.NoError(t, err)
}

func TestGitHubRESTClientRejectsMalformedRepository(t *testing.T) {
	client := NewGitHubRESTClient("https://api.github.com", "token")
	err := client.CreateIssueComment(context.Background(), "owner/repo/extra", 1, "report")
	require.Error(t, err)
}

func TestGitHubCommitStatusState(t *testing.T) {
	t.Parallel()

	availableRate := func(v float64) domain.OutcomeRateMetric {
		return domain.OutcomeRateMetric{Value: &v, Available: true}
	}

	tests := []struct {
		name     string
		overview domain.OutcomeOverview
		want     string
	}{
		{
			name: "no data stays pending",
			overview: domain.OutcomeOverview{
				Runs: domain.OutcomeRunMetrics{
					Total:       0,
					SuccessRate: availableRate(0),
				},
			},
			want: "pending",
		},
		{
			name: "unavailable data stays pending",
			overview: domain.OutcomeOverview{
				Runs: domain.OutcomeRunMetrics{
					Total:       1,
					Successful:  1,
					SuccessRate: domain.OutcomeRateMetric{Available: false},
				},
			},
			want: "pending",
		},
		{
			name: "all in progress stays pending",
			overview: domain.OutcomeOverview{
				Runs: domain.OutcomeRunMetrics{
					Total:       5,
					Successful:  0,
					Failed:      0,
					InProgress:  5,
					SuccessRate: availableRate(0),
				},
			},
			want: "pending",
		},
		{
			name: "completed runs without failures are success",
			overview: domain.OutcomeOverview{
				Runs: domain.OutcomeRunMetrics{
					Total:       5,
					Successful:  4,
					Failed:      0,
					InProgress:  1,
					SuccessRate: availableRate(0.8),
				},
			},
			want: "success",
		},
		{
			name: "failed run is failure",
			overview: domain.OutcomeOverview{
				Runs: domain.OutcomeRunMetrics{
					Total:       5,
					Successful:  3,
					Failed:      2,
					InProgress:  0,
					SuccessRate: availableRate(0.6),
				},
			},
			want: "failure",
		},
		{
			name: "regression signal is failure even with in-progress runs",
			overview: domain.OutcomeOverview{
				Runs: domain.OutcomeRunMetrics{
					Total:       3,
					InProgress:  3,
					SuccessRate: availableRate(0),
				},
				SourceControl: domain.OutcomeSCMMetrics{RegressionSignals: 1},
			},
			want: "failure",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.want, githubCommitStatusState(&tc.overview))
		})
	}
}
