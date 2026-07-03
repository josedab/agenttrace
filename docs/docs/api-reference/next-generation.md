---
title: Outcome, Replay, Eval Hub, and Sharing APIs
description: Canonical API routes introduced by the next-generation product consolidation.
---

# Outcome, Replay, Eval Hub, and Sharing APIs

All authenticated routes are project scoped. API keys use their bound project; JWT requests must provide `X-Project-ID` or use the explicit `/api/v1/projects/:projectId/...` routes.

## Outcomes

```text
GET  /api/public/outcomes
GET  /api/public/outcomes/digest
POST /api/public/outcomes/github-report
POST /api/public/outcomes/digest/deliver
```

`digest/deliver` validates every requested webhook before the first send. Once sending begins it never aborts: the `200` body reports `succeeded`, `failed`, `duplicates`, and one delivery record per webhook.

## Replay

```text
GET  /api/public/traces/:traceId/replay
GET  /api/public/traces/:traceId/replay-capabilities
POST /api/public/traces/:traceId/replay-plans
GET  /api/public/replay-plans/:planId
POST /api/public/replay-plans/:planId/execute
POST /api/public/replay-plans/:planId/retry
GET  /api/public/replay-plans/:planId/comparison
```

`execute` claims the plan with one conditional transition, so concurrent calls cannot both run it; the loser receives `409` with the current status. `retry` returns only a failed plan to `ready`; running plans are never reclaimed based solely on age.

## Eval Hub

```text
GET  /api/public/eval-hub/packages
POST /api/public/eval-hub/packages
GET  /api/public/eval-hub/packages/:packageId
POST /api/public/eval-hub/packages/:packageId/fork
POST /api/public/eval-hub/packages/:packageId/runs
GET  /api/public/eval-hub/runs
GET  /api/public/eval-hub/runs/:runId
```

Packages cover datasets, evaluators, prompts, and experiments. Benchmarks are rejected with `422` because they are not project owned. Run requests accept a project-unique `idempotencyKey`; repeating it returns the original run.

## Outbound infrastructure

These routes create or use outbound destinations and are refused with `422` in privacy no-egress mode:

```text
POST /api/public/otel/destinations
POST /api/public/otel-bridge/destinations
POST /api/public/federation/peers
POST /api/public/federation/query
POST /api/public/federation/destinations
POST /api/public/warehouse/connections
POST /api/public/warehouse/connections/:connId/test
POST /api/public/warehouse/connections/:connId/sync
POST /v1/export/data          # only when a remote destination is supplied
POST /v1/export/dataset       # only when a remote destination is supplied
POST /api/public/migrations   # only when the source DSN is not json-export
```

## Redacted sharing

Authenticated creation:

```text
POST   /api/public/traces/:traceId/share-links
POST   /api/public/replay-plans/:planId/share-links
DELETE /api/v1/projects/:projectId/share-links/:linkId
```

Unauthenticated, rate-limited resolution:

```text
GET /api/share/:token
```

Tokens contain 256 bits of randomness. Only their SHA-256 hashes are stored. Public responses omit prompts, outputs, metadata, commands, paths, diffs, repository identifiers, and credentials.

## Langfuse import

```text
POST /api/public/migrations/langfuse/import
```

Use the CLI rather than calling this batched endpoint directly:

```bash
agenttrace migrate --source langfuse --source-file ./langfuse-export.json --dry-run
agenttrace migrate --source langfuse --source-file ./langfuse-export.json
```

Generation usage accepts `input`/`output`/`total`,
`inputTokens`/`outputTokens`/`totalTokens`, and
`promptTokens`/`completionTokens`. Negative token counts are rejected. Batches
for the same migration job are serialized across API instances, so concurrent
uploads cannot overwrite progress or errors.
