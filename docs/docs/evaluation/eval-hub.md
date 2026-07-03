---
title: Eval Hub
description: One workspace for evaluators, datasets, prompts, experiments, benchmarks, and packages.
---

# Eval Hub

Eval Hub consolidates the former evaluator, dataset, playground, benchmark, and marketplace surfaces under `/evals`.

## Versioned packages

A package references one existing project asset:

- dataset
- evaluator
- prompt
- experiment

Benchmarks cannot be packaged. A benchmark row has no owning project and points at a dataset and evaluators that belong to the project that created it, so publishing or forking one would move another project's resources across a tenant boundary. Publish or fork the benchmark's dataset and evaluators instead; the API answers `422 Unprocessable Entity` for benchmark packages.

Publishing creates an immutable manifest and SHA-256 checksum. Publishing the same package again creates the next version. Visibility can be:

- `private`: owner project only
- `organization`: projects in the same organization
- `public`: discoverable by authenticated projects

## Forking and provenance

Public and organization packages must be forked before execution in another project. Forks preserve the source package and version and materialize project-owned assets where supported.

Dataset forks validate every packaged item before anything is created. If materialization fails partway, the partially created dataset is deleted and the fork fails, so a package never points at incomplete data. When cleanup itself fails, the error names the dataset that was left behind.

Dataset package snapshots omit source trace and observation links. Dataset publication is bounded to 5,000 items.

## Runs

- dataset packages create an existing dataset run and wait for traces/scores through the normal dataset API
- prompt packages compile deterministically with supplied variables
- experiment packages start the existing experiment
- evaluator packages return an explicit prerequisite message when a target trace or observation is required

A run row is persisted before any dataset run or experiment is started, so an execution that crashes leaves a durable record instead of invisible work. The run is then updated with the outcome, and a failed execution is stored with status `failed` and a safe message.

Run requests accept an idempotency key. The key is unique per project: a repeated request, including two concurrent requests that race, returns the run that was created first and executes the work exactly once. The web client reuses one key per package version while the run is `ready` or `running`, and after a failed response, so a double-click or retry cannot start a second active run. It releases the key after a terminal response so a later deliberate run is new work.

## API

```text
GET  /api/public/eval-hub/packages
POST /api/public/eval-hub/packages
GET  /api/public/eval-hub/packages/:packageId
POST /api/public/eval-hub/packages/:packageId/fork
POST /api/public/eval-hub/packages/:packageId/runs
GET  /api/public/eval-hub/runs
GET  /api/public/eval-hub/runs/:runId
```

Former `/eval-marketplace` web routes redirect to Eval Hub. Dataset marketplace API routes remain as compatibility aliases.
