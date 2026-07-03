# Changelog

All notable changes to AgentTrace will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Project-scoped agent outcome analytics across traces, git links, CI runs, pull requests, costs, models, and agents
- Safe checkpoint-aware replay plans with deterministic recorded-generation comparison
- Versioned Eval Hub packages with private, organization, and public visibility, provenance, forks, and idempotent runs
- `agenttrace init`, project-aware read-only MCP tools, and resumable Langfuse JSON export import
- Expiring, revocable, server-redacted trace and replay share links
- Optional GitHub outcome reports and SSRF-safe Slack, Discord, and generic webhook team digests
- Enforceable local/private no-egress mode with effective capability reporting
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
- No-egress mode now refuses every runtime-created outbound surface, including OpenTelemetry destinations and exports, federation peers and queries, warehouse connections/tests/syncs, remote export destinations, and remote-source migrations
- Replay plan execution claims a plan with an atomic conditional transition, refuses concurrent execution with a conflict, and recovers executions abandoned for more than 15 minutes
- Eval Hub persists a durable run before external execution, resolves idempotency-key races by returning the first run, and rolls back partially materialized dataset forks
- Team digest delivery validates every webhook before the first send, reports honest per-webhook partial results, and suppresses immediate duplicate sends with a delivery key
- Langfuse imports derive deterministic identifiers, so a retry after a failed ledger write rewrites the same records instead of duplicating them
- The CLI MCP server binds to loopback only, prints the bound address, applies read/write/idle timeouts, and serializes shared trace state
- Consolidated production navigation around Trace & Replay, Eval Hub, Prompts, Cost Center, and Collaboration
- Replaced hardcoded replay and marketplace paths with repository-backed project data
- Improved documentation structure with better navigation (#90)

### Fixed
- Sensitive-data redaction no longer collapses whitespace and newlines and replaces only credential-bearing URL substrings
- Handlers no longer continue with an empty project or user identity when the request context is missing one
- Benchmarks can no longer be published or forked as Eval Hub packages, which previously crossed project ownership
- Eval Hub and outcome digest interfaces expose labelled controls, accessible clipboard failures, and stable run idempotency across retries and double-clicks
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
