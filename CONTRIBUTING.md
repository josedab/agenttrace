# Contributing to AgentTrace

Thank you for your interest in contributing to AgentTrace! This document provides guidelines and information for contributors.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Project Structure](#project-structure)
- [Making Contributions](#making-contributions)
- [Code Style Guidelines](#code-style-guidelines)
- [Testing](#testing)
- [Pull Request Process](#pull-request-process)
- [Issue Guidelines](#issue-guidelines)
- [Troubleshooting](#troubleshooting)
- [Community](#community)

## Code of Conduct

By participating in this project, you agree to abide by our [Code of Conduct](CODE_OF_CONDUCT.md).

## Getting Started

### Prerequisites

- **Go 1.25.12+** - Backend development
- **Node.js 20+** - Frontend and TypeScript SDK development
- **Python 3.9+** - Python SDK development
- **Docker with Docker Compose v2** - Running the core or full application
- **Make** - Build automation

### Quick Start

Default tests do not require PostgreSQL, ClickHouse, Redis, MinIO, or any other deployed service. Run `make test` after installing the language dependencies for the components you are changing.

For application development, choose the smallest stack that covers your work:

| Workflow | Services | Commands |
|----------|----------|----------|
| Tests only | None | `make test` |
| Core API and dashboard | PostgreSQL, ClickHouse | `make setup && make dev` |
| Workers, async exports, distributed rate limiting | PostgreSQL, ClickHouse, Redis, MinIO | `make setup-full && make dev-full` |

1. **Fork the repository** on GitHub

2. **Clone your fork**:
   ```bash
   git clone https://github.com/YOUR_USERNAME/agenttrace.git
   cd agenttrace
   ```

3. **Install pre-commit hooks**:
   ```bash
   pip install pre-commit
   pre-commit install
   ```

4. **Set up the minimal development environment**:
   ```bash
   make doctor
   make setup
   ```

5. **Start the API and frontend**:
   ```bash
   make dev
   ```

6. **Access the application** at http://localhost:3000

## Development Setup

### Development with Hot Reload

For a faster feedback loop during backend development, use `make dev-hot`, which runs the minimal stack with [air](https://github.com/air-verse/air) for automatic Go rebuilds:

```bash
# Install air (one-time)
go install github.com/air-verse/air@latest

# Start with hot-reload (API auto-rebuilds on .go file changes)
make dev-hot
```

This starts both the Go API server (with hot-reload) and the Next.js dev server. The frontend already supports hot-reload via Next.js Fast Refresh.

Use `make dev-hot-full` when the feature also needs Redis, MinIO, or the background worker.

### Backend (Go)

```bash
cd api

# Install dependencies
go mod download

# Run the server against the minimal stack
make run-core

# Run the server with Redis and MinIO enabled
make run-full

# Run the background worker
make run-worker

# Run tests
make test

# Run linter
make lint

# Generate GraphQL code
make generate
```

### Frontend (Next.js)

```bash
cd web

# Install dependencies
npm install

# Run development server
npm run dev

# Run tests
npm test

# Run linter
npm run lint

# Type check
npm run type-check

# Build for production
npm run build
```

### Python SDK

```bash
cd sdk/python

# Create virtual environment
python -m venv venv
source venv/bin/activate  # or `venv\Scripts\activate` on Windows

# Install in development mode
pip install -e ".[dev]"

# Run tests
pytest

# Run linter
ruff check .

# Type check
mypy agenttrace
```

### TypeScript SDK

```bash
cd sdk/typescript

# Install dependencies
npm install

# Run tests
npm test

# Run linter
npm run lint

# Build
npm run build
```

### Go SDK

```bash
cd sdk/go

# Run tests
go test -v ./...

# Run with race detection
go test -race ./...
```

## Project Structure

```
agenttrace/
├── api/                    # Go backend
│   ├── cmd/               # Entry points
│   │   ├── server/       # HTTP server
│   │   └── worker/       # Background worker
│   ├── internal/          # Internal packages
│   │   ├── domain/       # Domain models
│   │   ├── handler/      # HTTP handlers
│   │   ├── service/      # Business logic
│   │   ├── repository/   # Data access
│   │   ├── graphql/      # GraphQL resolvers
│   │   ├── middleware/   # HTTP middleware
│   │   └── worker/       # Background jobs
│   ├── migrations/        # Database migrations
│   │   ├── postgres/
│   │   └── clickhouse/
│   └── schema/            # GraphQL schema
├── web/                   # Next.js frontend
│   ├── app/              # App router pages
│   ├── components/       # React components
│   ├── hooks/            # Custom hooks
│   └── lib/              # Utilities
├── sdk/                   # Language SDKs
│   ├── python/           # Python SDK
│   ├── typescript/       # TypeScript SDK
│   ├── go/               # Go SDK
│   └── cli/              # CLI wrapper
├── docs/                  # Documentation (Docusaurus)
├── deploy/               # Deployment configs
├── examples/             # Example projects
├── extensions/           # IDE extensions (VS Code, JetBrains)
├── actions/              # GitHub Actions (reusable workflows)
├── ci/                   # CI pipeline configuration
├── scripts/              # Development and build scripts
└── mobile/               # Mobile app (experimental)
```

## Making Contributions

### Types of Contributions

We welcome many types of contributions:

- **Bug fixes** - Fix issues and improve stability
- **Features** - Add new functionality
- **Documentation** - Improve docs, add examples
- **Tests** - Increase test coverage
- **Performance** - Optimize code and queries
- **Integrations** - Add support for new frameworks/tools

### Before You Start

1. **Check existing issues** - Someone may already be working on it
2. **Open an issue first** for major changes to discuss the approach
3. **Keep changes focused** - One feature/fix per PR

### Branching Strategy

- `main` - Stable release branch
- `feature/*` - Feature branches
- `fix/*` - Bug fix branches
- `docs/*` - Documentation branches

Create your branch from `main`:
```bash
git checkout main
git pull origin main
git checkout -b feature/your-feature-name
```

## Code Style Guidelines

### Go (Backend)

- Follow [Effective Go](https://golang.org/doc/effective_go) guidelines
- Use `gofmt` for formatting (automatically enforced)
- Follow [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- Use meaningful variable names
- Add comments for exported functions
- Handle errors explicitly - don't ignore them

```go
// Good
func (s *TraceService) GetTrace(ctx context.Context, id string) (*domain.Trace, error) {
    if id == "" {
        return nil, ErrInvalidTraceID
    }
    return s.repo.FindByID(ctx, id)
}

// Avoid
func (s *TraceService) GetTrace(ctx context.Context, id string) (*domain.Trace, error) {
    return s.repo.FindByID(ctx, id) // Missing validation
}
```

### TypeScript (Frontend & SDK)

- Use TypeScript strict mode
- Prefer functional components with hooks
- Use meaningful component and variable names
- Export types alongside implementations

```typescript
// Good
interface TraceListProps {
  projectId: string;
  limit?: number;
}

export function TraceList({ projectId, limit = 50 }: TraceListProps) {
  const { data, isLoading } = useTraces(projectId, { limit });
  // ...
}

// Avoid
export function TraceList(props: any) {
  // Missing types
}
```

### Python (SDK)

- Follow [PEP 8](https://pep8.org/) style guide
- Use type hints for all public functions
- Use docstrings for public modules, classes, and functions
- Prefer dataclasses or Pydantic for data structures

```python
# Good
def create_trace(
    name: str,
    metadata: dict[str, Any] | None = None,
) -> Trace:
    """Create a new trace.

    Args:
        name: The name of the trace.
        metadata: Optional metadata to attach to the trace.

    Returns:
        The created Trace object.
    """
    ...

# Avoid
def create_trace(name, metadata=None):
    # Missing types and docstring
    ...
```

## Testing

### Writing Tests

- Write tests for all new features and bug fixes
- Aim for meaningful coverage, not just high percentages
- Test edge cases and error conditions
- Use table-driven tests in Go where appropriate

### Running Tests

The default test commands do not require deployed databases, caches, queues, or object storage. Database integration tests are opt-in and skip unless their test environment variables are configured.

```bash
# Backend
cd api && make test

# Frontend
cd web && npm test

# Python SDK
cd sdk/python && pytest

# TypeScript SDK
cd sdk/typescript && npm test

# Go SDK
cd sdk/go && go test ./...
```

### Test Categories

- **Unit tests** - Test individual functions/components
- **Integration tests** - Test interactions between components
- **E2E tests** - Test complete user workflows

## Pull Request Process

### Before Submitting

1. **Ensure tests pass**: All existing and new tests must pass
2. **Run linters**: Fix all linting errors
3. **Update documentation**: Add/update docs for your changes
4. **Write a clear description**: Explain what and why

### PR Title Convention

Use conventional commit format:
- `feat: Add LangChain integration`
- `fix: Resolve trace duplication issue`
- `docs: Update SDK quickstart guide`
- `test: Add integration tests for prompts`
- `refactor: Simplify ingestion pipeline`
- `chore: Update dependencies`

### PR Description Template

```markdown
## Summary
Brief description of the changes.

## Changes
- Change 1
- Change 2

## Testing
How to test these changes.

## Related Issues
Fixes #123
```

### Review Process

1. Submit your PR against the `main` branch
2. Automated checks will run (lint, test, build)
3. A maintainer will review your code
4. Address any feedback
5. Once approved, a maintainer will merge your PR

## Issue Guidelines

### Bug Reports

Include:
- Clear description of the bug
- Steps to reproduce
- Expected vs actual behavior
- Environment details (OS, versions)
- Relevant logs or screenshots

### Feature Requests

Include:
- Clear description of the feature
- Use case / problem it solves
- Proposed solution (if any)
- Alternatives considered

### Issue Labels

| Label | Description |
|-------|-------------|
| `bug` | Something isn't working |
| `feature` | New feature request |
| `docs` | Documentation improvements |
| `good first issue` | Good for newcomers |
| `help wanted` | Extra attention needed |
| `sdk/python` | Python SDK related |
| `sdk/typescript` | TypeScript SDK related |
| `sdk/go` | Go SDK related |
| `frontend` | Web frontend related |
| `backend` | Go backend related |

## Troubleshooting

### Port Conflicts

For the minimal stack, check ports 3000, 5432, 8080, 8123, and 9000. The full stack also uses 6379, 9001, and 9002.
```bash
# Find what's using a port
lsof -i :8080

# Kill the process or change the port in deploy/.env
```

### Database Migrations

Migration targets use a pinned `golang-migrate` runner through `go run`; no global migration CLI or host-installed ClickHouse client is required.

```bash
make migrate-pg-up
make migrate-ch-up
```

### Docker Memory Issues

ClickHouse requires significant memory. If containers crash:
1. Increase Docker Desktop memory to at least 4GB (Settings → Resources)
2. Restart Docker Desktop
3. Run `make docker-up` again

### Node Version Mismatch

This project requires Node.js 20+. Check `.nvmrc` for the exact version:
```bash
# If using nvm
nvm use

# Check your version
node --version
```

### Services Not Starting

Run `make doctor` to verify all prerequisites are installed, then:
```bash
# Check Docker container status
docker compose -f deploy/docker-compose.dev.yml ps

# View container logs
docker compose -f deploy/docker-compose.dev.yml logs

# Include Redis and MinIO when debugging the full stack
docker compose -f deploy/docker-compose.dev.yml --profile full ps
```

## Community

### Getting Help

- **GitHub Discussions** - Ask questions, share ideas
- **Discord** - Real-time chat with the community
- **GitHub Issues** - Report bugs, request features

### Stay Updated

- Watch the repository for updates
- Follow our blog for announcements
- Join our Discord for community discussions

## Recognition

Contributors are recognized in:
- Release notes
- Contributors list in README
- Special badges for significant contributions

Thank you for contributing to AgentTrace!
