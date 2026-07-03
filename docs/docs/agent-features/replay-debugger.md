---
title: Time-Travel Replay Debugger
description: Inspect recorded timelines, plan safe replay, and compare replay branches.
---

# Time-Travel Replay Debugger

The replay debugger combines trace observations, file operations, terminal commands, checkpoints, and git links into one ordered timeline.

## Safety model

AgentTrace does **not** execute recorded commands, tools, or file writes on the API host.

Two replay modes are represented:

- `recorded_generation`: deterministically replays recorded model outputs by content hash. It makes no provider call and incurs no provider cost.
- `sandbox`: reported as unavailable unless a separate sandbox executor is configured. AgentTrace does not silently fall back to host execution.

Model, prompt, or temperature overrides require a configured sandbox/model provider and are reported as unsupported otherwise.

## Workflow

1. Open a trace and choose **Open replay debugger**.
2. Inspect the real timeline.
3. Select the trace origin or an authorized checkpoint.
4. Create a replay plan and review its capability report.
5. Execute recorded-generation replay when the plan is ready.
6. Compare original and replay generation counts, tokens, costs, and deterministic output hashes.

Plans transition through `ready`, `running`, `completed`, `failed`, or `unsupported`. Every lookup and every write is scoped to the authenticated project, including checkpoints and replay branches.

### Concurrency and recovery

Execution claims a plan with a single conditional update (`project_id`, `id`, `status = 'ready'` → `running`). Only one request can win that transition, so two concurrent execute calls never run the same plan twice: the loser receives `409 Conflict` naming the current status.

Failed plans can be returned to `ready` with the project-scoped retry endpoint. Running plans are **not** reclaimed based only on age: elapsed time cannot prove that an external sandbox stopped, and automatic takeover could execute the same plan twice. If a process dies mid-execution, verify that its sandbox execution has stopped, then create a replacement plan. Failure and completion updates remain scoped to the project that claimed the plan.

## API

```text
GET  /api/public/traces/:traceId/replay-capabilities
POST /api/public/traces/:traceId/replay-plans
GET  /api/public/replay-plans/:planId
POST /api/public/replay-plans/:planId/execute
POST /api/public/replay-plans/:planId/retry
GET  /api/public/replay-plans/:planId/comparison
```
