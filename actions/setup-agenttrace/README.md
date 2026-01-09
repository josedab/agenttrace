# Setup AgentTrace GitHub Action

Integrate AgentTrace observability into your GitHub Actions workflows. This action sets up the AgentTrace environment, creates CI run records, and automatically links git commits to traces.

## Features

- Automatic environment setup with AgentTrace credentials
- CI run tracking and status reporting
- Git commit linking to traces
- Optional SDK installation (Python, TypeScript, Go, or CLI)
- Supports self-hosted AgentTrace instances

## Usage

### Basic Setup

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
        run: |
          # Your AI agent code here
          # Environment variables are automatically set:
          # - AGENTTRACE_API_KEY
          # - AGENTTRACE_PROJECT_ID
          # - AGENTTRACE_API_URL
          # - AGENTTRACE_SESSION_ID
          python run_agent.py
```

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

    with at.trace('ci-agent-run'):
        # Your agent code here
        pass
    "
```

### With TypeScript SDK

```yaml
- name: Setup Node.js
  uses: actions/setup-node@v4
  with:
    node-version: '20'

- name: Setup AgentTrace
  uses: agenttrace/setup-agenttrace@v1
  with:
    api-key: ${{ secrets.AGENTTRACE_API_KEY }}
    project-id: ${{ secrets.AGENTTRACE_PROJECT_ID }}
    install-sdk: typescript

- name: Run Agent
  run: npx ts-node agent.ts
```

### With CLI Wrapper

```yaml
- name: Setup AgentTrace
  uses: agenttrace/setup-agenttrace@v1
  with:
    api-key: ${{ secrets.AGENTTRACE_API_KEY }}
    project-id: ${{ secrets.AGENTTRACE_PROJECT_ID }}
    install-sdk: cli

- name: Run Agent with Tracing
  run: agenttrace wrap -- your-agent-command
```

### Self-Hosted Instance

```yaml
- name: Setup AgentTrace
  uses: agenttrace/setup-agenttrace@v1
  with:
    api-key: ${{ secrets.AGENTTRACE_API_KEY }}
    project-id: ${{ secrets.AGENTTRACE_PROJECT_ID }}
    api-url: https://agenttrace.your-company.com/api
```

### Advanced Configuration

```yaml
- name: Setup AgentTrace
  id: agenttrace
  uses: agenttrace/setup-agenttrace@v1
  with:
    api-key: ${{ secrets.AGENTTRACE_API_KEY }}
    project-id: ${{ secrets.AGENTTRACE_PROJECT_ID }}
    session-id: 'deployment-${{ github.sha }}'
    link-commits: true
    create-ci-run: true
    environment: production
    install-sdk: python

- name: Use CI Run ID
  run: echo "CI Run ID: ${{ steps.agenttrace.outputs.ci-run-id }}"
```

## Inputs

| Input | Description | Required | Default |
|-------|-------------|----------|---------|
| `api-key` | AgentTrace API key | Yes | - |
| `project-id` | AgentTrace project ID | Yes | - |
| `api-url` | AgentTrace API URL | No | `https://api.agenttrace.io` |
| `session-id` | Session ID to group traces | No | `${{ github.run_id }}` |
| `link-commits` | Auto-link commits to traces | No | `true` |
| `create-ci-run` | Create CI run record | No | `true` |
| `ci-provider` | CI provider name | No | `github_actions` |
| `environment` | Deployment environment | No | - |
| `install-sdk` | SDK to install (`python`, `typescript`, `go`, `cli`, `none`) | No | `none` |
| `python-version` | Python version for SDK | No | - |
| `node-version` | Node.js version for SDK | No | - |

## Outputs

| Output | Description |
|--------|-------------|
| `ci-run-id` | The created CI run UUID |
| `session-id` | The session ID used for traces |
| `api-url` | The AgentTrace API URL |

## Environment Variables

The action sets the following environment variables:

| Variable | Description |
|----------|-------------|
| `AGENTTRACE_API_KEY` | Your API key |
| `AGENTTRACE_PROJECT_ID` | Your project ID |
| `AGENTTRACE_API_URL` | API URL |
| `AGENTTRACE_SESSION_ID` | Session ID for grouping |
| `AGENTTRACE_CI_PROVIDER` | CI provider name |
| `AGENTTRACE_CI_RUN_ID` | GitHub run ID |
| `AGENTTRACE_CI_WORKFLOW` | Workflow name |
| `AGENTTRACE_CI_JOB` | Job name |
| `AGENTTRACE_CI_ACTOR` | User who triggered |
| `AGENTTRACE_CI_EVENT` | Event type |
| `AGENTTRACE_CI_REF` | Git ref |
| `AGENTTRACE_CI_SHA` | Commit SHA |
| `AGENTTRACE_CI_REPO` | Repository name |
| `AGENTTRACE_ENVIRONMENT` | Deployment environment (if set) |
| `AGENTTRACE_CI_RUN_UUID` | AgentTrace CI run UUID |

## Completing CI Runs

To update the CI run status when your workflow completes, add a cleanup step:

```yaml
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

## Support

- Documentation: https://docs.agenttrace.io
- Issues: https://github.com/agenttrace/agenttrace/issues
- Discord: https://discord.gg/agenttrace
