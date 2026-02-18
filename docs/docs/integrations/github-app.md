---
sidebar_position: 7
title: GitHub App Integration
description: Automatically link pull requests and commits to AgentTrace traces with the AgentTrace GitHub App.
---

# GitHub App Integration

The AgentTrace GitHub App automatically links pull requests and commits to traces, and posts trace summaries directly in PR comments.

## Installation

1. Go to [github.com/apps/agenttrace](https://github.com/apps/agenttrace)
2. Click **Install**
3. Select the repositories you want to connect
4. Authorize the requested permissions

## Quick Start

1. Install the GitHub App (see above)
2. In the AgentTrace dashboard, go to **Settings** > **Integrations** > **GitHub**
3. Click **Connect GitHub** and authorize the OAuth flow
4. Configure your webhook URL:
   - **Cloud**: automatically configured during OAuth
   - **Self-hosted**: set the webhook URL to `https://<your-host>/api/v1/webhooks/github`
5. Select which projects to link to each repository

## Features

### Auto-Link PRs to Traces

When a pull request is opened or updated, the GitHub App:

- Detects traces associated with the PR's branch and commit SHAs
- Links matching traces to the PR in the AgentTrace dashboard
- Provides a direct link from the PR to the trace timeline

### Trace Summary in PR Comments

The app posts a comment on each PR with:

- Number of traces generated during the PR lifecycle
- Total cost and token usage across all linked traces
- Latency summary (mean, P95) for linked traces
- Direct links to individual traces in the AgentTrace UI

Comments are updated automatically as new traces are recorded.

### Commit Linking

Every push event is processed to:

- Link the commit SHA to any traces generated in CI or locally
- Show trace counts per commit in the AgentTrace dashboard
- Enable "View Traces" navigation from the commit view

## Webhook Configuration

### Cloud-Hosted

No manual webhook setup is needed. The OAuth flow configures the webhook automatically.

### Self-Hosted

1. In the AgentTrace dashboard, go to **Settings** > **Integrations** > **GitHub**
2. Copy the webhook URL: `https://<your-host>/api/v1/webhooks/github`
3. Copy the webhook secret
4. In your GitHub App settings, set:
   - **Webhook URL**: the URL from step 2
   - **Webhook Secret**: the secret from step 3
   - **Content type**: `application/json`

### Required Events

The app listens for these webhook events:

| Event | Purpose |
|-------|---------|
| `push` | Link commits to traces |
| `pull_request` | Link PRs and post trace comments |
| `check_suite` | Update trace status on CI completion |

## Permissions

The GitHub App requires the following permissions:

| Permission | Access | Reason |
|------------|--------|--------|
| Pull requests | Read & Write | Post trace summary comments |
| Contents | Read | Read commit and branch info |
| Checks | Read | Monitor CI run status |
| Metadata | Read | Repository metadata |

## Troubleshooting

- **App not posting comments**: Verify the app is installed on the repository and has **Pull requests: Read & Write** permission.
- **Traces not linking**: Ensure `AGENTTRACE_PROJECT_ID` matches the project linked to the repository in **Settings** > **Integrations**.
- **Webhook delivery failures**: Check the webhook delivery log in your GitHub App settings under **Advanced**. Confirm the webhook URL is reachable and the secret matches.
