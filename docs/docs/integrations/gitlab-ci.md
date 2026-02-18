---
sidebar_position: 3
title: GitLab CI Integration
description: Integrate AgentTrace into your GitLab CI/CD pipelines for automatic trace collection and CI run tracking.
---

# GitLab CI Integration

Integrate AgentTrace into your GitLab CI/CD pipelines for automatic trace collection, CI run tracking, and commit linking.

## Quick Start

Include the AgentTrace CI template in your `.gitlab-ci.yml`:

```yaml
include:
  - remote: 'https://raw.githubusercontent.com/agenttrace/agenttrace/main/ci/gitlab/agenttrace.yml'

run-agent:
  extends: .agenttrace-python
  script:
    - python run_agent.py
```

For self-hosted instances, use a local include instead:

```yaml
include:
  - local: 'ci/gitlab/agenttrace.yml'
```

## Configuration

### Required Variables

Set the following in **Settings > CI/CD > Variables** (mark `AGENTTRACE_API_KEY` as masked):

| Variable | Required | Description |
|----------|----------|-------------|
| `AGENTTRACE_API_KEY` | Yes | AgentTrace API key (starts with `sk-at-`) |
| `AGENTTRACE_PROJECT_ID` | Yes | Your AgentTrace project ID |
| `AGENTTRACE_API_URL` | No | Custom API URL for self-hosted deployments |

### Available Templates

The [`ci/gitlab/agenttrace.yml`](https://github.com/agenttrace/agenttrace/blob/main/ci/gitlab/agenttrace.yml) file provides language-specific templates:

| Template | Description |
|----------|-------------|
| `.agenttrace-python` | Python projects — installs `agenttrace` via pip |
| `.agenttrace-node` | Node.js projects — installs `agenttrace` via npm |
| `.agenttrace-go` | Go projects — installs the Go SDK |
| `.agenttrace-cli` | Any project — installs the AgentTrace CLI |

## Features

- **CI run tracking**: Automatically creates and updates CI run records per pipeline
- **Git commit linking**: Links commits to traces via the `/v1/git-links` endpoint
- **Run completion**: Updates the CI run status (`completed` or `failed`) in an `after_script`

## Examples

### Python Agent Pipeline

```yaml
include:
  - remote: 'https://raw.githubusercontent.com/agenttrace/agenttrace/main/ci/gitlab/agenttrace.yml'

agent-pipeline:
  extends: .agenttrace-python
  script:
    - python -c "
      from agenttrace import AgentTrace
      at = AgentTrace()
      with at.trace('gitlab-ci-agent'):
          pass
      "
```

### Custom Setup with CLI

```yaml
include:
  - remote: 'https://raw.githubusercontent.com/agenttrace/agenttrace/main/ci/gitlab/agenttrace.yml'

agent-run:
  extends: .agenttrace-cli
  script:
    - agenttrace run -- python run_agent.py
```

## Environment Variables

The template sets these variables automatically:

| Variable | Source |
|----------|--------|
| `AGENTTRACE_SESSION_ID` | `$CI_PIPELINE_ID` |
| `AGENTTRACE_CI_PROVIDER` | `gitlab_ci` |
| `AGENTTRACE_CI_SHA` | `$CI_COMMIT_SHA` |
| `AGENTTRACE_CI_REF` | `$CI_COMMIT_REF_NAME` |
| `AGENTTRACE_CI_RUN_UUID` | Created at runtime |
