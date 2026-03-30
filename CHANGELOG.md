# Changelog

All notable changes to AgentTrace will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- OpenTelemetry Collector support for exporting traces to external backends (Jaeger, Grafana Tempo, Datadog, Honeycomb, New Relic) (#87)
- OTLP receiver for ingesting traces from OpenTelemetry-instrumented applications (#87)
- Architecture Decision Records (ADRs) documenting major architectural choices (#92)
- Anomaly detection for latency, cost, and error rate monitoring (#78)
- A/B testing framework for comparing agent configurations (#80)
- Community prompt library with versioning and forking (#83)
- Jupyter notebook extension for data science workflows (#95)
- VS Code extension for inline trace viewing (#72)
- Grafana datasource plugin for enterprise monitoring integration (#88)
- E2E API test suite for integration testing (#65)
- Pre-commit hooks for local linting (#66)
- Coverage thresholds in CI (60% minimum) (#67)
- Troubleshooting guide in documentation (#91)
- GraphQL API documentation (#70)
- Hello world examples for Python, TypeScript, and Go (#69)

### Changed
- Improved documentation structure with better navigation (#90)

### Fixed
- Various documentation typos and broken links (#93)

## [0.1.0] - 2024-01-15

### Added
- Initial release of AgentTrace (#1)
- Core tracing functionality with traces, spans, generations, and events (#2)
- Multi-language SDK support (Python, TypeScript, Go, CLI) (#3)
- PostgreSQL for transactional data storage (#4)
- ClickHouse for trace analytics and time-series data (#4)
- REST API for trace ingestion (#5)
- GraphQL API for complex queries (#6)
- JWT and API key authentication (#7)
- Cost tracking for 400+ LLM models (#12)
- Latency analysis and performance monitoring (#13)
- Session management for grouping related traces (#14)
- Scoring system for trace evaluation (#15)
- Prompt management with versioning and labels (#18)
- Dataset creation and experiment running (#20)
- LLM-as-judge evaluation framework (#22)
- Human annotation interface (#23)
- Git linking for trace-to-commit correlation (#25)
- Code checkpoints for state snapshots (#26)
- File operation tracking (#27)
- Terminal command logging (#28)
- CI/CD integration (GitHub Actions, GitLab CI) (#30)
- Docker Compose deployment (#8)
- Kubernetes Helm chart (#35)
- Web dashboard with Next.js 15 (#9)
- Real-time trace streaming (#38)
- Export functionality (JSON, CSV) (#40)
- Webhook notifications (#42)

### Security
- API key hashing with bcrypt (#7)
- JWT token validation (#7)
- Rate limiting middleware (#45)
- Input validation and sanitization (#46)

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
