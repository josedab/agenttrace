# AgentTrace GitLab CI/CD Integration

Integrate AgentTrace observability into your GitLab CI/CD pipelines.

## Quick Start

### 1. Set CI/CD Variables

In your GitLab project, go to **Settings > CI/CD > Variables** and add:

| Variable | Description | Masked |
|----------|-------------|--------|
| `AGENTTRACE_API_KEY` | Your AgentTrace API key | Yes |
| `AGENTTRACE_PROJECT_ID` | Your project ID | No |
| `AGENTTRACE_API_URL` | (Optional) Custom API URL | No |

### 2. Include the Template

Add to your `.gitlab-ci.yml`:

```yaml
include:
  - remote: 'https://raw.githubusercontent.com/agenttrace/agenttrace/main/ci/gitlab/agenttrace.yml'
```

For self-hosted, copy the file and use:

```yaml
include:
  - local: 'ci/gitlab/agenttrace.yml'
```

### 3. Use a Template

```yaml
agent-job:
  extends: .agenttrace-python
  script:
    - python run_agent.py
```

## Available Templates

### `.agenttrace-python`

Pre-configured with Python and the AgentTrace Python SDK:

```yaml
run-agent:
  extends: .agenttrace-python
  script:
    - python -c "
      from agenttrace import AgentTrace
      at = AgentTrace()
      with at.trace('ci-agent'):
          # Your agent code
          pass
      "
```

### `.agenttrace-node`

Pre-configured with Node.js and the AgentTrace TypeScript SDK:

```yaml
run-agent:
  extends: .agenttrace-node
  script:
    - npx ts-node agent.ts
```

### `.agenttrace-go`

Pre-configured with Go and the AgentTrace Go SDK:

```yaml
run-agent:
  extends: .agenttrace-go
  script:
    - go run main.go
```

### `.agenttrace-cli`

Pre-configured with the AgentTrace CLI wrapper:

```yaml
run-agent:
  extends: .agenttrace-cli
  script:
    - agenttrace wrap -- your-agent-command
```

## Custom Setup

For more control, use the base templates:

```yaml
include:
  - remote: 'https://raw.githubusercontent.com/agenttrace/agenttrace/main/ci/gitlab/agenttrace.yml'

my-agent-job:
  extends:
    - .agenttrace-setup
    - .agenttrace-complete
  image: your-custom-image:latest
  variables:
    AGENTTRACE_ENVIRONMENT: production
  script:
    - echo "Running with AgentTrace session: $AGENTTRACE_SESSION_ID"
    - your-custom-agent
```

## Environment Variables

The templates automatically set:

| Variable | Value |
|----------|-------|
| `AGENTTRACE_API_KEY` | From CI/CD variables |
| `AGENTTRACE_PROJECT_ID` | From CI/CD variables |
| `AGENTTRACE_API_URL` | API URL |
| `AGENTTRACE_SESSION_ID` | Pipeline ID |
| `AGENTTRACE_CI_PROVIDER` | `gitlab_ci` |
| `AGENTTRACE_CI_RUN_ID` | Pipeline ID |
| `AGENTTRACE_CI_JOB_ID` | Job ID |
| `AGENTTRACE_CI_WORKFLOW` | Project name |
| `AGENTTRACE_CI_JOB` | Job name |
| `AGENTTRACE_CI_ACTOR` | User login |
| `AGENTTRACE_CI_EVENT` | Pipeline source |
| `AGENTTRACE_CI_REF` | Branch/tag name |
| `AGENTTRACE_CI_SHA` | Commit SHA |
| `AGENTTRACE_CI_REPO` | Project path |
| `AGENTTRACE_CI_RUN_UUID` | AgentTrace run UUID |

## Features

### Automatic CI Run Tracking

The templates automatically:
- Create a CI run record when the job starts
- Link the git commit to the project
- Update the CI run status when the job completes

### Multiple Jobs

Each job creates its own CI run:

```yaml
include:
  - remote: 'https://raw.githubusercontent.com/agenttrace/agenttrace/main/ci/gitlab/agenttrace.yml'

build-agent:
  extends: .agenttrace-python
  stage: build
  script:
    - python build.py

test-agent:
  extends: .agenttrace-python
  stage: test
  script:
    - pytest tests/

deploy-agent:
  extends: .agenttrace-python
  stage: deploy
  variables:
    AGENTTRACE_ENVIRONMENT: production
  script:
    - python deploy.py
  only:
    - main
```

## Troubleshooting

### CI Run Not Created

Check that:
1. `AGENTTRACE_API_KEY` is set and valid
2. `AGENTTRACE_PROJECT_ID` is correct
3. `jq` is available in the image

### Link Commit Failed

This is non-blocking. The job will continue even if commit linking fails.

### Custom Image Without curl/jq

Add them in your job:

```yaml
my-job:
  extends: .agenttrace-setup
  image: alpine:latest
  before_script:
    - apk add --no-cache curl jq
    - !reference [.agenttrace-setup, before_script]
```

## Support

- Documentation: https://docs.agenttrace.io
- Issues: https://github.com/agenttrace/agenttrace/issues
