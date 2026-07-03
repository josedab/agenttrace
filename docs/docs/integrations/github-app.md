---
sidebar_position: 7
title: GitHub App Integration
description: Automatically link pull requests and commits to AgentTrace traces with the AgentTrace GitHub App.
---

# GitHub App Integration

AgentTrace can ingest GitHub push webhooks to link commits with traces. Outcome reports can be delivered on demand when optional GitHub reporting credentials are configured.

## Prerequisite

Outcome delivery requires a GitHub repository record linked to the AgentTrace project. GitHub is optional; trace ingestion, outcome analytics, and digest rendering work without it.

## Features

### Commit Linking

Every push event is processed to:

- Link the commit SHA to any traces generated in CI or locally
- Show trace counts per commit in the AgentTrace dashboard
- Enable "View Traces" navigation from the commit view

### Outcome Reports

Outcome reporting is optional and never required for core AgentTrace usage.

Configure:

```bash
GITHUB_REPORTING_ENABLED=true
GITHUB_REPORT_TOKEN=...
GITHUB_API_URL=https://api.github.com
```

Then deliver a digest to a pull request:

```bash
curl -X POST "$AGENTTRACE_HOST/api/public/outcomes/github-report" \
  -H "Authorization: Bearer $AGENTTRACE_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "repositoryId": "00000000-0000-0000-0000-000000000001",
    "pullRequestNumber": 42,
    "window": "7d"
  }'
```

Provide `commitSha` instead of `pullRequestNumber` to create an `AgentTrace/outcomes` commit status. Reports use the same real-data digest shown in the dashboard.

If reporting is not configured, the API returns an explicit capability error. In no-egress mode, delivery is blocked.

## Webhook events

Deployments that mount the optional GitHub webhook handler should select only implemented events:

| Event | Purpose |
|-------|---------|
| `installation` | Track app installation state |
| `installation_repositories` | Track repository access changes |
| `push` | Store commit metadata for later trace correlation |

## Permissions

The GitHub App requires the following permissions:

| Permission | Access | Reason |
|------------|--------|--------|
| Issues | Write when reporting is enabled | Post an explicitly requested PR digest |
| Commit statuses | Write when reporting is enabled | Publish `AgentTrace/outcomes` status |
| Contents | Read | Read commit and branch info |
| Metadata | Read | Repository metadata |

## Troubleshooting

- **Report delivery says not configured**: set `GITHUB_REPORTING_ENABLED` and `GITHUB_REPORT_TOKEN`.
- **Traces not linking**: Ensure `AGENTTRACE_PROJECT_ID` matches the project linked to the repository in **Settings** > **Integrations**.
- **Webhook delivery failures**: Check the webhook delivery log in your GitHub App settings under **Advanced**. Confirm the webhook URL is reachable and the secret matches.
