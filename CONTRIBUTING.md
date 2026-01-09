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
- [Community](#community)

## Code of Conduct

By participating in this project, you agree to maintain a respectful and inclusive environment. We expect all contributors to:

- Be respectful and considerate in all interactions
- Welcome newcomers and help them get started
- Focus on constructive feedback
- Accept responsibility for mistakes and learn from them

## Getting Started

### Prerequisites

- **Go 1.21+** - Backend development
- **Node.js 18+** - Frontend and TypeScript SDK development
- **Python 3.9+** - Python SDK development
- **Docker & Docker Compose** - Running local services
- **Make** - Build automation

### Quick Start

1. **Fork the repository** on GitHub

2. **Clone your fork**:
   ```bash
   git clone https://github.com/YOUR_USERNAME/agenttrace.git
   cd agenttrace
   ```

3. **Set up the development environment**:
   ```bash
   # Start required services (PostgreSQL, ClickHouse, Redis, MinIO)
   cd deploy
   cp .env.example .env
   docker compose -f docker-compose.dev.yml up -d

   # Wait for services to be healthy
   docker compose -f docker-compose.dev.yml ps
   ```

4. **Run database migrations**:
   ```bash
   cd ../api
   make migrate-pg-up
   make migrate-ch-up
   ```

5. **Start the backend**:
   ```bash
   cd api
   go run cmd/server/main.go
   ```

6. **Start the frontend** (in a new terminal):
   ```bash
   cd web
   npm install
   npm run dev
   ```

7. **Access the application** at http://localhost:3000

## Development Setup

### Backend (Go)

```bash
cd api

# Install dependencies
go mod download

# Run the server
go run cmd/server/main.go

# Run the background worker
go run cmd/worker/main.go

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
npm run typecheck

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
├── extensions/           # IDE extensions
│   ├── vscode/
│   └── jetbrains/
└── examples/             # Example projects
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
- `develop` - Integration branch for features
- `feature/*` - Feature branches
- `fix/*` - Bug fix branches
- `docs/*` - Documentation branches

Create your branch from `develop`:
```bash
git checkout develop
git pull origin develop
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

1. Submit your PR against the `develop` branch
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
