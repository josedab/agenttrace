# AgentTrace Regression Gate

GitHub Action that runs regression tests against your AI agent's prompt/model changes and blocks PRs on quality drops.

## Quick Start

```yaml
name: Agent Quality Gate
on:
  pull_request:
    branches: [main]

jobs:
  regression-test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Run regression gate
        uses: agenttrace/agenttrace/actions/regression-gate@main
        with:
          api-key: ${{ secrets.AGENTTRACE_API_KEY }}
          test-ids: "test-1,test-2,test-3"
```

## Inputs

| Input | Required | Default | Description |
|-------|----------|---------|-------------|
| `api-key` | ✅ | — | AgentTrace API key |
| `api-url` | — | `https://api.agenttrace.io` | AgentTrace API URL (for self-hosted) |
| `test-ids` | ✅ | — | Comma-separated regression test IDs |
| `fail-on-regression` | — | `true` | Whether to fail the workflow on quality drop |
| `comment-on-pr` | — | `true` | Post results as a PR comment |

## Outputs

| Output | Description |
|--------|-------------|
| `passed` | `true` if all tests passed |
| `results-json` | Full results as JSON |
| `summary` | Human-readable summary |

## How It Works

1. Calls the AgentTrace regression gate API with your test IDs
2. Each test runs its dataset against the current agent configuration
3. Compares scores against the baseline thresholds
4. Posts a markdown table as a PR comment
5. Fails the workflow if any metric drops below threshold

## Setting Up Regression Tests

1. Go to AgentTrace → Regression Tests → Create Test
2. Select a baseline dataset and evaluators
3. Set quality thresholds (e.g., "accuracy ≥ 0.85")
4. Copy the test ID into your workflow

## Self-Hosted

```yaml
- uses: agenttrace/agenttrace/actions/regression-gate@main
  with:
    api-key: ${{ secrets.AGENTTRACE_API_KEY }}
    api-url: "https://your-instance.example.com"
    test-ids: "test-1"
```
