---
title: Local and Private Mode
description: Enforce no-egress behavior and inspect effective privacy capabilities.
---

# Local and Private Mode

Set both variables to enable local/private behavior:

```bash
PRIVACY_NO_EGRESS=true
PRIVACY_REDACTION_ENABLED=true
```

No-egress mode refuses every runtime-created outbound surface before any network call is made. A refused request answers `422 Unprocessable Entity` and performs zero outbound calls.

| Capability | Blocked behavior |
| --- | --- |
| `webhookDelivery` | Slack, Discord, and generic webhook delivery, including team digests |
| `githubReporting` | GitHub PR comments and commit statuses |
| `otelExport` | OTLP exporter creation, connection tests, span export, OTel/OTel-bridge export destinations, and enabling bridge export |
| `federation` | Federation peer registration and federated queries against remote instances |
| `remoteExport` | Federation export destinations and export jobs that target a caller-supplied remote destination |
| `warehouseSync` | Warehouse connection creation, connection tests, and syncs |
| `remoteImport` | Migrations whose source is a remote platform DSN |
| `externalModelProviders` | Model provider calls; features that use a model fall back to their deterministic local behavior |
| `sentry` | Sentry telemetry, refused at startup rather than at runtime |

Local behavior is unaffected: local trace storage, redacted share links, the Langfuse **JSON export** import (`json-export`), and export jobs that write to the deployment's own object storage all remain available.

AgentTrace also refuses startup when no-egress conflicts with enabled external providers. Conflicts include:

- `GITHUB_REPORTING_ENABLED=true`
- `OTEL_EXPORTER_ENABLED=true`
- `SENTRY_ENABLED=true`
- any configured external evaluation/model API key (`EVAL_API_KEY`)
- configured Google or GitHub OAuth
- disabled deterministic redaction

This validation prevents a deployment from claiming private operation while still enabling egress.

## Capability endpoint

```bash
curl "$AGENTTRACE_HOST/api/public/privacy/capabilities" \
  -H "Authorization: Bearer $AGENTTRACE_API_KEY"
```

The Privacy Center renders this response directly. Each key in `capabilities` maps to a capability that is enforced in code by a shared outbound guard, so the report and the runtime behavior cannot drift apart.
