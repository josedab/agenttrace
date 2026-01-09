---
sidebar_position: 1
---

# GitHub Actions Integration

Integrate AgentTrace into your GitHub Actions workflows for automatic trace collection and CI run tracking.

## Quick Start

Add to your workflow:

```yaml
name: AI Agent Pipeline
on: [push, pull_request]

jobs:
  run-agent:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Setup AgentTrace
        uses: agenttrace/setup-agenttrace@v1
        with:
          api-key: ${{ secrets.AGENTTRACE_API_KEY }}
          project-id: ${{ secrets.AGENTTRACE_PROJECT_ID }}

      - name: Run AI Agent
        run: python run_agent.py
```

## Features

- **Automatic environment setup**: Sets all required environment variables
- **CI run tracking**: Creates and updates CI run records
- **Git commit linking**: Automatically links commits to traces
- **SDK installation**: Optionally installs Python/TypeScript/Go SDK

## Configuration

### Inputs

| Input | Required | Default | Description |
|-------|----------|---------|-------------|
| `api-key` | Yes | - | AgentTrace API key |
| `project-id` | Yes | - | Project ID |
| `api-url` | No | Cloud URL | API URL (for self-hosted) |
| `install-sdk` | No | `none` | SDK to install |
| `link-commits` | No | `true` | Auto-link git commits |
| `create-ci-run` | No | `true` | Create CI run record |

### Outputs

| Output | Description |
|--------|-------------|
| `ci-run-id` | Created CI run UUID |
| `session-id` | Session ID for grouping |

## Examples

### With Python SDK

```yaml
- name: Setup AgentTrace
  uses: agenttrace/setup-agenttrace@v1
  with:
    api-key: ${{ secrets.AGENTTRACE_API_KEY }}
    project-id: ${{ secrets.AGENTTRACE_PROJECT_ID }}
    install-sdk: python

- name: Run Agent
  run: |
    python -c "
    from agenttrace import AgentTrace
    at = AgentTrace()
    with at.trace('ci-agent'):
        # Your agent code
        pass
    "
```

### Complete CI Run Tracking

```yaml
- name: Setup AgentTrace
  id: agenttrace
  uses: agenttrace/setup-agenttrace@v1
  with:
    api-key: ${{ secrets.AGENTTRACE_API_KEY }}
    project-id: ${{ secrets.AGENTTRACE_PROJECT_ID }}

- name: Run Tests
  run: pytest tests/

- name: Complete CI Run
  if: always()
  run: |
    STATUS="${{ job.status }}"
    if [ "$STATUS" == "success" ]; then
      AGENTTRACE_STATUS="completed"
    else
      AGENTTRACE_STATUS="failed"
    fi

    curl -X PATCH "$AGENTTRACE_API_URL/v1/ci-runs/$AGENTTRACE_CI_RUN_UUID" \
      -H "Authorization: Bearer $AGENTTRACE_API_KEY" \
      -H "Content-Type: application/json" \
      -d "{\"status\": \"$AGENTTRACE_STATUS\"}"
```

## Environment Variables

The action sets these variables:

| Variable | Description |
|----------|-------------|
| `AGENTTRACE_API_KEY` | API key |
| `AGENTTRACE_PROJECT_ID` | Project ID |
| `AGENTTRACE_SESSION_ID` | Run ID |
| `AGENTTRACE_CI_RUN_UUID` | CI run UUID |
| `AGENTTRACE_CI_PROVIDER` | `github_actions` |
| `AGENTTRACE_CI_SHA` | Commit SHA |
| `AGENTTRACE_CI_REF` | Git ref |
