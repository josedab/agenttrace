# Changelog

All notable changes to AgentTrace will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- OpenTelemetry Collector support for exporting traces to external backends (Jaeger, Grafana Tempo, Datadog, Honeycomb, New Relic)
- OTLP receiver for ingesting traces from OpenTelemetry-instrumented applications
- Architecture Decision Records (ADRs) documenting major architectural choices
- Anomaly detection for latency, cost, and error rate monitoring
- A/B testing framework for comparing agent configurations
- Community prompt library with versioning and forking
- Jupyter notebook extension for data science workflows
- VS Code extension for inline trace viewing
- Grafana datasource plugin for enterprise monitoring integration
- E2E API test suite for integration testing
- Pre-commit hooks for local linting
- Coverage thresholds in CI (60% minimum)
- Troubleshooting guide in documentation
- GraphQL API documentation
- Hello world examples for Python, TypeScript, and Go

### Changed
- Improved documentation structure with better navigation

### Fixed
- Various documentation typos and broken links

## [0.1.0] - 2024-01-15

### Added
- Initial release of AgentTrace
- Core tracing functionality with traces, spans, generations, and events
- Multi-language SDK support (Python, TypeScript, Go, CLI)
- PostgreSQL for transactional data storage
- ClickHouse for trace analytics and time-series data
- REST API for trace ingestion
- GraphQL API for complex queries
- JWT and API key authentication
- Cost tracking for 400+ LLM models
- Latency analysis and performance monitoring
- Session management for grouping related traces
- Scoring system for trace evaluation
- Prompt management with versioning and labels
- Dataset creation and experiment running
- LLM-as-judge evaluation framework
- Human annotation interface
- Git linking for trace-to-commit correlation
- Code checkpoints for state snapshots
- File operation tracking
- Terminal command logging
- CI/CD integration (GitHub Actions, GitLab CI)
- Docker Compose deployment
- Kubernetes Helm chart
- Web dashboard with Next.js 15
- Real-time trace streaming
- Export functionality (JSON, CSV)
- Webhook notifications

### Security
- API key hashing with bcrypt
- JWT token validation
- Rate limiting middleware
- Input validation and sanitization

---

## Release Types

- **Major (X.0.0)**: Breaking API changes, major architectural shifts
- **Minor (0.X.0)**: New features, backwards-compatible additions
- **Patch (0.0.X)**: Bug fixes, security patches, documentation updates

## How to Update This Changelog

When contributing to AgentTrace:

1. Add your changes under the `[Unreleased]` section
2. Use the appropriate category:
   - `Added` for new features
   - `Changed` for changes in existing functionality
   - `Deprecated` for soon-to-be removed features
   - `Removed` for now removed features
   - `Fixed` for bug fixes
   - `Security` for vulnerability fixes
3. Write entries from the user's perspective
4. Include relevant issue/PR numbers where applicable

Example entry:
```markdown
### Added
- Support for Claude 3.5 Sonnet model pricing (#123)
```

[Unreleased]: https://github.com/agenttrace/agenttrace/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/agenttrace/agenttrace/releases/tag/v0.1.0
