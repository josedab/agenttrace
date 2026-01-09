# AgentTrace PRD v3
## Product Requirements Document - Complete Langfuse Parity

**Version:** 3.0  
**Date:** January 2026  
**Author:** Jose Baena  
**Status:** Draft

---

## Table of Contents

1. [Executive Summary](#1-executive-summary)
2. [System Architecture](#2-system-architecture)
3. [Data Model](#3-data-model)
4. [ClickHouse Schema Design](#4-clickhouse-schema-design)
5. [Feature Requirements](#5-feature-requirements)
6. [API Design](#6-api-design)
7. [SDK Design](#7-sdk-design)
8. [Business Model](#8-business-model)
9. [Development Timeline](#9-development-timeline)
10. [Risks and Mitigations](#10-risks-and-mitigations)
11. [Open Questions](#11-open-questions)
12. [Appendices](#12-appendices)

---

## 1. Executive Summary

### 1.1 Vision

**"Langfuse for Coding Agents"** - AgentTrace is an open-source observability platform purpose-built for AI coding agents like Claude Code, GitHub Copilot Workspace, Cursor, and Aider. While Langfuse excels at general LLM application observability, AgentTrace extends this foundation with first-class support for the unique needs of autonomous coding agents: git correlation, code checkpoints, terminal command logging, file operation tracking, and CI/CD attribution.

### 1.2 Strategic Position

AgentTrace occupies the intersection of two rapidly growing markets:

- **LLM Observability** ($500M+ TAM, 40% CAGR) - Currently served by Langfuse, LangSmith, Datadog LLM Observability
- **AI-Assisted Development** ($10B+ TAM by 2027) - Rapidly expanding with GitHub Copilot (1.8M+ paid subscribers), Cursor, Claude Code

No existing solution adequately addresses the observability needs of autonomous coding agents. AgentTrace fills this gap.

### 1.3 Architecture Decision

AgentTrace follows Langfuse's proven architecture pattern with enhancements for efficiency:

| Component | Langfuse | AgentTrace | Rationale |
|-----------|----------|------------|-----------|
| Frontend | Next.js 15 | Next.js 15 | Proven, SSR, Server Components |
| Backend | Node.js/tRPC | Go + Fiber | 10x throughput, lower latency |
| Auth | NextAuth | NextAuth | Mature, flexible providers |
| Trace Storage | ClickHouse (v3) | ClickHouse | OLAP optimized, proven at scale |
| Metadata | PostgreSQL | PostgreSQL | ACID for users, projects, prompts |
| Queue | Redis + BullMQ | Redis + Asynq | Go-native, simpler, performant |
| Cache | Redis | Redis | Standard, proven |
| Object Storage | S3/Blob | S3/MinIO | Checkpoints, exports, multi-modal |

### 1.4 Key Differentiators

**Full Langfuse Parity:**
- Tracing & observations (spans, generations, events)
- Sessions and user tracking
- Prompt management with version control
- Cost tracking with 400+ model support
- Evaluation (LLM-as-Judge, annotation queues)
- Datasets & experiments
- Custom dashboards & metrics
- Role-based access control
- Self-hosting with Docker

**AgentTrace Unique Features:**
- Git correlation (bi-directional trace ↔ commit linking)
- Code checkpoints (state snapshots, rollback capability)
- Terminal command logging (CLI agent observability)
- File operation tracking (filesystem visibility)
- CI/CD attribution (test failure → agent trace linking)
- Multi-agent SDK (Python, TypeScript, Go, CLI wrapper)

---

## 2. System Architecture

### 2.1 High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              AgentTrace Platform                             │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │                        app.agenttrace.dev                           │   │
│  │                         (Next.js 15 App)                            │   │
│  │  ┌──────────────┐ ┌──────────────┐ ┌──────────────┐ ┌────────────┐ │   │
│  │  │   Dashboard  │ │    Traces    │ │   Prompts    │ │   Evals    │ │   │
│  │  │     UI       │ │   Explorer   │ │   Manager    │ │  & Scores  │ │   │
│  │  └──────────────┘ └──────────────┘ └──────────────┘ └────────────┘ │   │
│  │  ┌──────────────┐ ┌──────────────┐ ┌──────────────┐ ┌────────────┐ │   │
│  │  │  Playground  │ │   Datasets   │ │    Users     │ │  Settings  │ │   │
│  │  │              │ │ & Experiments│ │   & RBAC     │ │            │ │   │
│  │  └──────────────┘ └──────────────┘ └──────────────┘ └────────────┘ │   │
│  │  ┌───────────────────────────────────────────────────────────────┐ │   │
│  │  │                    BFF API Routes (/api/*)                    │ │   │
│  │  │              NextAuth │ GraphQL Proxy │ Aggregation           │ │   │
│  │  └───────────────────────────────────────────────────────────────┘ │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                      │                                      │
│                                      ▼                                      │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │                        api.agenttrace.dev                           │   │
│  │                          (Go Backend)                               │   │
│  │  ┌──────────────────────────────────────────────────────────────┐  │   │
│  │  │                      Fiber HTTP Server                        │  │   │
│  │  │  ┌────────────┐ ┌────────────┐ ┌────────────┐ ┌────────────┐ │  │   │
│  │  │  │   OTLP     │ │  GraphQL   │ │  REST API  │ │  Webhooks  │ │  │   │
│  │  │  │  Ingestion │ │   Query    │ │  (Public)  │ │            │ │  │   │
│  │  │  │  /v1/traces│ │  /graphql  │ │  /api/v1/* │ │  /webhooks │ │  │   │
│  │  │  └────────────┘ └────────────┘ └────────────┘ └────────────┘ │  │   │
│  │  └──────────────────────────────────────────────────────────────┘  │   │
│  │  ┌──────────────────────────────────────────────────────────────┐  │   │
│  │  │                    Asynq Worker Pool                          │  │   │
│  │  │  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────────────┐ │  │   │
│  │  │  │  Cost    │ │  Eval    │ │  Export  │ │    Checkpoint    │ │  │   │
│  │  │  │  Calc    │ │  Runner  │ │  Jobs    │ │    Processing    │ │  │   │
│  │  │  └──────────┘ └──────────┘ └──────────┘ └──────────────────┘ │  │   │
│  │  └──────────────────────────────────────────────────────────────┘  │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                      │                                      │
│          ┌───────────────────────────┼───────────────────────────┐         │
│          ▼                           ▼                           ▼         │
│  ┌──────────────┐           ┌──────────────┐           ┌──────────────┐   │
│  │  ClickHouse  │           │  PostgreSQL  │           │    Redis     │   │
│  │              │           │              │           │              │   │
│  │  - Traces    │           │  - Users     │           │  - Queue     │   │
│  │  - Spans     │           │  - Projects  │           │  - Cache     │   │
│  │  - Scores    │           │  - Prompts   │           │  - Sessions  │   │
│  │  - Metrics   │           │  - Datasets  │           │  - Rate Limit│   │
│  │  - Events    │           │  - API Keys  │           │              │   │
│  └──────────────┘           └──────────────┘           └──────────────┘   │
│          │                                                      │          │
│          └──────────────────────────┬───────────────────────────┘          │
│                                     ▼                                      │
│                            ┌──────────────┐                                │
│                            │   S3/MinIO   │                                │
│                            │              │                                │
│                            │ - Checkpoints│                                │
│                            │ - File Snaps │                                │
│                            │ - Exports    │                                │
│                            │ - Multi-modal│                                │
│                            └──────────────┘                                │
└─────────────────────────────────────────────────────────────────────────────┘
```

### 2.2 Service Responsibilities

**Next.js App (app.agenttrace.dev)**
- Server-side rendered dashboard UI
- NextAuth authentication (Email, Google, GitHub, OIDC SSO)
- BFF API routes for UI aggregation
- GraphQL proxy to Go backend
- Static asset serving
- Real-time updates via Server-Sent Events

**Go Backend (api.agenttrace.dev)**
- OTLP trace ingestion (OpenTelemetry protocol)
- Legacy REST batch ingestion (Langfuse compatibility)
- GraphQL query API (gqlgen)
- Public REST API (OpenAPI 3.1)
- Webhook delivery
- Background job processing (Asynq)
- Cost calculation engine
- Evaluation execution
- Export generation

**Asynq Workers**
- Cost calculation for generations
- LLM-as-Judge evaluation execution
- Scheduled exports (S3, GCS, Azure Blob)
- Checkpoint processing and storage
- Metrics aggregation
- Data retention cleanup

### 2.3 Repository Structure

```
agenttrace/
├── api/                          # Go backend
│   ├── cmd/
│   │   ├── server/               # Main API server
│   │   └── worker/               # Asynq worker process
│   ├── internal/
│   │   ├── config/               # Configuration
│   │   ├── handler/              # HTTP handlers
│   │   ├── middleware/           # Auth, rate limiting
│   │   ├── service/              # Business logic
│   │   ├── repository/           # Data access
│   │   ├── graphql/              # GraphQL resolvers
│   │   ├── worker/               # Background jobs
│   │   └── pkg/                  # Shared utilities
│   ├── migrations/               # SQL migrations
│   ├── schema/                   # GraphQL schema
│   └── go.mod
│
├── web/                          # Next.js frontend
│   ├── app/                      # App router pages
│   │   ├── (auth)/               # Auth pages
│   │   ├── (dashboard)/          # Dashboard pages
│   │   ├── api/                  # API routes (BFF)
│   │   └── layout.tsx
│   ├── components/               # React components
│   ├── lib/                      # Utilities
│   ├── hooks/                    # Custom hooks
│   └── package.json
│
├── sdk/                          # Client SDKs
│   ├── python/                   # Python SDK
│   ├── typescript/               # TypeScript SDK
│   ├── go/                       # Go SDK
│   └── cli/                      # CLI wrapper
│
├── deploy/                       # Deployment configs
│   ├── docker-compose.yml        # Local/self-hosted
│   ├── docker-compose.dev.yml    # Development
│   └── k8s/                      # Kubernetes manifests
│
├── helm/                         # Helm chart
│   └── agenttrace/
│
└── docs/                         # Documentation
    ├── getting-started/
    ├── api-reference/
    ├── sdk/
    └── self-hosting/
```

---

## 3. Data Model

### 3.1 Core Entities

#### 3.1.1 Trace

A trace represents a single request or operation in your application. For coding agents, this typically represents one user prompt → agent response cycle.

| Field | Type | Description |
|-------|------|-------------|
| id | UUID | Primary identifier (W3C format: 32 hex chars) |
| project_id | UUID | Parent project |
| name | String | Trace name (e.g., "code-generation", "refactor") |
| user_id | String? | End-user identifier |
| session_id | String? | Session grouping identifier |
| input | JSON? | Trace input (user prompt, context) |
| output | JSON? | Trace output (final response) |
| metadata | JSON? | Custom metadata |
| version | String? | Application version |
| release | String? | Release identifier |
| tags | String[] | Filterable tags |
| environment | String | dev/staging/prod |
| level | Enum | DEBUG, DEFAULT, WARNING, ERROR |
| public | Boolean | Public sharing enabled |
| start_time | Timestamp | Trace start |
| end_time | Timestamp? | Trace end |
| created_at | Timestamp | Record creation |
| updated_at | Timestamp | Last update |

#### 3.1.2 Observation

Observations are the building blocks of traces. They can be spans (generic operations), generations (LLM calls), or events (point-in-time occurrences).

| Field | Type | Description |
|-------|------|-------------|
| id | UUID | Primary identifier (W3C format: 16 hex chars) |
| trace_id | UUID | Parent trace |
| parent_observation_id | UUID? | Parent observation (for nesting) |
| project_id | UUID | Parent project |
| type | Enum | SPAN, GENERATION, EVENT |
| name | String | Observation name |
| start_time | Timestamp | Start time |
| end_time | Timestamp? | End time |
| level | Enum | DEBUG, DEFAULT, WARNING, ERROR |
| status_message | String? | Status or error message |
| input | JSON? | Observation input |
| output | JSON? | Observation output |
| metadata | JSON? | Custom metadata |
| version | String? | Version identifier |

**Generation-Specific Fields:**

| Field | Type | Description |
|-------|------|-------------|
| model | String | Model identifier (e.g., "claude-sonnet-4-20250514") |
| model_parameters | JSON | Temperature, max_tokens, etc. |
| usage_details | JSON | Token counts (input, output, cache_read, cache_creation) |
| cost_details | JSON | Calculated costs (input_cost, output_cost, total_cost) |
| prompt_id | UUID? | Linked prompt (for prompt management) |
| prompt_version | Int? | Linked prompt version |

#### 3.1.3 Session

Sessions group related traces, typically representing a multi-turn conversation or extended interaction.

| Field | Type | Description |
|-------|------|-------------|
| id | String | Session identifier |
| project_id | UUID | Parent project |
| user_id | String? | Associated user |
| created_at | Timestamp | First trace time |
| bookmarked | Boolean | Starred for review |
| public | Boolean | Public sharing enabled |

#### 3.1.4 Score

Scores capture evaluations of traces or observations, from automated evaluators or human feedback.

| Field | Type | Description |
|-------|------|-------------|
| id | UUID | Primary identifier |
| project_id | UUID | Parent project |
| trace_id | UUID | Scored trace |
| observation_id | UUID? | Scored observation (optional) |
| name | String | Score name (e.g., "helpfulness", "accuracy") |
| value | Float? | Numeric score value |
| string_value | String? | Categorical score value |
| data_type | Enum | NUMERIC, BOOLEAN, CATEGORICAL |
| source | Enum | API, EVAL, ANNOTATION |
| comment | String? | Evaluator comment |
| config_id | UUID? | Linked evaluator config |
| author_user_id | UUID? | Human annotator |
| created_at | Timestamp | Score creation |

### 3.2 Prompt Management Entities

#### 3.2.1 Prompt

| Field | Type | Description |
|-------|------|-------------|
| id | UUID | Primary identifier |
| project_id | UUID | Parent project |
| name | String | Prompt name (unique per project) |
| type | Enum | TEXT, CHAT |
| created_at | Timestamp | Creation time |
| updated_at | Timestamp | Last update |

#### 3.2.2 PromptVersion

| Field | Type | Description |
|-------|------|-------------|
| id | UUID | Primary identifier |
| prompt_id | UUID | Parent prompt |
| version | Int | Version number (auto-increment) |
| content | JSON | Prompt content (text or chat messages) |
| config | JSON? | Model config (model, temperature, etc.) |
| labels | String[] | Assigned labels (production, staging, etc.) |
| created_at | Timestamp | Creation time |
| created_by | UUID? | Author user |

### 3.3 Evaluation Entities

#### 3.3.1 Dataset

| Field | Type | Description |
|-------|------|-------------|
| id | UUID | Primary identifier |
| project_id | UUID | Parent project |
| name | String | Dataset name |
| description | String? | Description |
| metadata | JSON? | Custom metadata |
| created_at | Timestamp | Creation time |
| updated_at | Timestamp | Last update |

#### 3.3.2 DatasetItem

| Field | Type | Description |
|-------|------|-------------|
| id | UUID | Primary identifier |
| dataset_id | UUID | Parent dataset |
| input | JSON | Item input |
| expected_output | JSON? | Expected output (ground truth) |
| metadata | JSON? | Item metadata |
| source_trace_id | UUID? | Source trace (if from production) |
| source_observation_id | UUID? | Source observation |
| status | Enum | ACTIVE, ARCHIVED |
| created_at | Timestamp | Creation time |

#### 3.3.3 DatasetRun

| Field | Type | Description |
|-------|------|-------------|
| id | UUID | Primary identifier |
| dataset_id | UUID | Parent dataset |
| name | String | Run name |
| description | String? | Run description |
| metadata | JSON? | Run metadata |
| created_at | Timestamp | Creation time |

#### 3.3.4 DatasetRunItem

| Field | Type | Description |
|-------|------|-------------|
| id | UUID | Primary identifier |
| dataset_run_id | UUID | Parent run |
| dataset_item_id | UUID | Source item |
| trace_id | UUID | Generated trace |
| observation_id | UUID? | Generated observation |
| created_at | Timestamp | Creation time |

#### 3.3.5 Evaluator

| Field | Type | Description |
|-------|------|-------------|
| id | UUID | Primary identifier |
| project_id | UUID | Parent project |
| name | String | Evaluator name |
| type | Enum | LLM_AS_JUDGE, CUSTOM, MANUAL |
| description | String? | Description |
| template_id | String? | Built-in template (hallucination, relevance, etc.) |
| prompt | JSON? | Custom prompt (for LLM-as-Judge) |
| model | String? | Evaluator model |
| output_schema | JSON | Expected output format |
| sampling_rate | Float | Sampling rate (0.0-1.0) |
| target_filter | JSON? | Filter for target traces |
| variable_mapping | JSON? | Map trace fields to prompt variables |
| enabled | Boolean | Active status |
| created_at | Timestamp | Creation time |
| updated_at | Timestamp | Last update |

### 3.4 AgentTrace Unique Entities

#### 3.4.1 Checkpoint

Represents a point-in-time snapshot of code state during agent execution.

| Field | Type | Description |
|-------|------|-------------|
| id | UUID | Primary identifier |
| project_id | UUID | Parent project |
| trace_id | UUID | Parent trace |
| observation_id | UUID? | Associated observation |
| name | String | Checkpoint name |
| description | String? | Description |
| storage_type | Enum | S3, GIT_STASH, INLINE |
| storage_uri | String | Storage location |
| file_manifest | JSON | List of files with checksums |
| git_sha | String? | Git commit at checkpoint time |
| git_branch | String? | Git branch |
| working_dir | String | Working directory path |
| size_bytes | Int64 | Total size |
| created_at | Timestamp | Creation time |

#### 3.4.2 GitLink

Associates traces with git commits, branches, and PRs.

| Field | Type | Description |
|-------|------|-------------|
| id | UUID | Primary identifier |
| project_id | UUID | Parent project |
| trace_id | UUID | Linked trace |
| repository_url | String | Repository URL |
| commit_sha | String | Commit SHA |
| branch | String? | Branch name |
| pr_number | Int? | Pull request number |
| pr_url | String? | Pull request URL |
| commit_message | String? | Commit message |
| author_name | String? | Commit author |
| author_email | String? | Author email |
| files_changed | String[] | Changed files |
| additions | Int? | Lines added |
| deletions | Int? | Lines deleted |
| created_at | Timestamp | Creation time |

#### 3.4.3 FileOperation

Tracks file system operations during agent execution.

| Field | Type | Description |
|-------|------|-------------|
| id | UUID | Primary identifier |
| project_id | UUID | Parent project |
| trace_id | UUID | Parent trace |
| observation_id | UUID? | Associated observation |
| operation | Enum | CREATE, READ, UPDATE, DELETE, MOVE, COPY |
| file_path | String | File path |
| file_type | String? | File extension/type |
| before_content_hash | String? | Content hash before |
| after_content_hash | String? | Content hash after |
| before_size | Int64? | Size before |
| after_size | Int64? | Size after |
| diff_summary | JSON? | Diff summary (lines added/removed) |
| timestamp | Timestamp | Operation time |

#### 3.4.4 TerminalCommand

Logs terminal/shell commands executed by agents.

| Field | Type | Description |
|-------|------|-------------|
| id | UUID | Primary identifier |
| project_id | UUID | Parent project |
| trace_id | UUID | Parent trace |
| observation_id | UUID? | Associated observation |
| command | String | Command executed |
| working_dir | String | Working directory |
| exit_code | Int? | Exit code |
| stdout | String? | Standard output (truncated) |
| stderr | String? | Standard error (truncated) |
| duration_ms | Int | Execution duration |
| timestamp | Timestamp | Execution time |

#### 3.4.5 CIRun

Associates traces with CI/CD pipeline runs.

| Field | Type | Description |
|-------|------|-------------|
| id | UUID | Primary identifier |
| project_id | UUID | Parent project |
| trace_id | UUID | Linked trace |
| provider | Enum | GITHUB_ACTIONS, GITLAB_CI, JENKINS, CIRCLECI |
| run_id | String | CI run identifier |
| run_url | String | CI run URL |
| workflow_name | String? | Workflow/pipeline name |
| job_name | String? | Job name |
| status | Enum | PENDING, RUNNING, SUCCESS, FAILURE, CANCELLED |
| started_at | Timestamp? | Run start time |
| finished_at | Timestamp? | Run finish time |
| trigger_event | String? | Trigger event (push, PR, etc.) |
| created_at | Timestamp | Link creation time |

---

## 4. ClickHouse Schema Design

### 4.1 Traces Table

```sql
CREATE TABLE traces (
    id UUID,
    project_id UUID,
    name String,
    user_id Nullable(String),
    session_id Nullable(String),
    input String DEFAULT '',  -- JSON as string
    output String DEFAULT '',  -- JSON as string
    metadata String DEFAULT '{}',
    version Nullable(String),
    release Nullable(String),
    tags Array(String) DEFAULT [],
    environment String DEFAULT 'production',
    level Enum8('DEBUG' = 0, 'DEFAULT' = 1, 'WARNING' = 2, 'ERROR' = 3) DEFAULT 'DEFAULT',
    public Bool DEFAULT false,
    start_time DateTime64(3),
    end_time Nullable(DateTime64(3)),
    created_at DateTime64(3) DEFAULT now64(3),
    updated_at DateTime64(3) DEFAULT now64(3),
    
    -- Calculated fields
    duration_ms Nullable(Int64) MATERIALIZED if(end_time IS NOT NULL, 
        toInt64((end_time - start_time) * 1000), NULL),
    
    -- Sign for ReplacingMergeTree
    sign Int8 DEFAULT 1
)
ENGINE = ReplacingMergeTree(updated_at)
PARTITION BY toYYYYMM(start_time)
ORDER BY (project_id, start_time, id)
SETTINGS index_granularity = 8192;

-- Secondary indexes
ALTER TABLE traces ADD INDEX idx_user_id user_id TYPE bloom_filter GRANULARITY 1;
ALTER TABLE traces ADD INDEX idx_session_id session_id TYPE bloom_filter GRANULARITY 1;
ALTER TABLE traces ADD INDEX idx_tags tags TYPE bloom_filter GRANULARITY 1;
ALTER TABLE traces ADD INDEX idx_name name TYPE bloom_filter GRANULARITY 1;
```

### 4.2 Observations Table

```sql
CREATE TABLE observations (
    id UUID,
    trace_id UUID,
    parent_observation_id Nullable(UUID),
    project_id UUID,
    type Enum8('SPAN' = 0, 'GENERATION' = 1, 'EVENT' = 2),
    name String,
    start_time DateTime64(3),
    end_time Nullable(DateTime64(3)),
    level Enum8('DEBUG' = 0, 'DEFAULT' = 1, 'WARNING' = 2, 'ERROR' = 3) DEFAULT 'DEFAULT',
    status_message Nullable(String),
    input String DEFAULT '',
    output String DEFAULT '',
    metadata String DEFAULT '{}',
    version Nullable(String),
    
    -- Generation-specific fields
    model Nullable(String),
    model_parameters String DEFAULT '{}',
    usage_details String DEFAULT '{}',
    cost_details String DEFAULT '{}',
    prompt_id Nullable(UUID),
    prompt_version Nullable(Int32),
    
    created_at DateTime64(3) DEFAULT now64(3),
    updated_at DateTime64(3) DEFAULT now64(3),
    
    -- Calculated fields
    duration_ms Nullable(Int64) MATERIALIZED if(end_time IS NOT NULL,
        toInt64((end_time - start_time) * 1000), NULL),
    
    -- Token extraction from usage_details
    input_tokens Nullable(Int64) MATERIALIZED JSONExtractInt(usage_details, 'input'),
    output_tokens Nullable(Int64) MATERIALIZED JSONExtractInt(usage_details, 'output'),
    total_tokens Nullable(Int64) MATERIALIZED 
        JSONExtractInt(usage_details, 'input') + JSONExtractInt(usage_details, 'output'),
    
    -- Cost extraction
    total_cost Nullable(Float64) MATERIALIZED JSONExtractFloat(cost_details, 'total'),
    
    sign Int8 DEFAULT 1
)
ENGINE = ReplacingMergeTree(updated_at)
PARTITION BY toYYYYMM(start_time)
ORDER BY (project_id, trace_id, start_time, id)
SETTINGS index_granularity = 8192;

-- Secondary indexes
ALTER TABLE observations ADD INDEX idx_type type TYPE set(3) GRANULARITY 1;
ALTER TABLE observations ADD INDEX idx_model model TYPE bloom_filter GRANULARITY 1;
ALTER TABLE observations ADD INDEX idx_name name TYPE bloom_filter GRANULARITY 1;
```

### 4.3 Scores Table

```sql
CREATE TABLE scores (
    id UUID,
    project_id UUID,
    trace_id UUID,
    observation_id Nullable(UUID),
    name String,
    value Nullable(Float64),
    string_value Nullable(String),
    data_type Enum8('NUMERIC' = 0, 'BOOLEAN' = 1, 'CATEGORICAL' = 2),
    source Enum8('API' = 0, 'EVAL' = 1, 'ANNOTATION' = 2),
    comment Nullable(String),
    config_id Nullable(UUID),
    author_user_id Nullable(UUID),
    created_at DateTime64(3) DEFAULT now64(3),
    
    sign Int8 DEFAULT 1
)
ENGINE = ReplacingMergeTree(created_at)
PARTITION BY toYYYYMM(created_at)
ORDER BY (project_id, trace_id, name, id)
SETTINGS index_granularity = 8192;
```

### 4.4 AgentTrace Unique Tables

#### Checkpoints

```sql
CREATE TABLE checkpoints (
    id UUID,
    project_id UUID,
    trace_id UUID,
    observation_id Nullable(UUID),
    name String,
    description Nullable(String),
    storage_type Enum8('S3' = 0, 'GIT_STASH' = 1, 'INLINE' = 2),
    storage_uri String,
    file_manifest String DEFAULT '[]',  -- JSON array
    git_sha Nullable(String),
    git_branch Nullable(String),
    working_dir String,
    size_bytes Int64,
    created_at DateTime64(3) DEFAULT now64(3)
)
ENGINE = MergeTree()
PARTITION BY toYYYYMM(created_at)
ORDER BY (project_id, trace_id, created_at, id)
SETTINGS index_granularity = 8192;
```

#### Git Links

```sql
CREATE TABLE git_links (
    id UUID,
    project_id UUID,
    trace_id UUID,
    repository_url String,
    commit_sha String,
    branch Nullable(String),
    pr_number Nullable(Int32),
    pr_url Nullable(String),
    commit_message Nullable(String),
    author_name Nullable(String),
    author_email Nullable(String),
    files_changed Array(String) DEFAULT [],
    additions Nullable(Int32),
    deletions Nullable(Int32),
    created_at DateTime64(3) DEFAULT now64(3)
)
ENGINE = MergeTree()
PARTITION BY toYYYYMM(created_at)
ORDER BY (project_id, trace_id, created_at, id)
SETTINGS index_granularity = 8192;

ALTER TABLE git_links ADD INDEX idx_commit_sha commit_sha TYPE bloom_filter GRANULARITY 1;
ALTER TABLE git_links ADD INDEX idx_branch branch TYPE bloom_filter GRANULARITY 1;
```

#### File Operations

```sql
CREATE TABLE file_operations (
    id UUID,
    project_id UUID,
    trace_id UUID,
    observation_id Nullable(UUID),
    operation Enum8('CREATE' = 0, 'READ' = 1, 'UPDATE' = 2, 'DELETE' = 3, 'MOVE' = 4, 'COPY' = 5),
    file_path String,
    file_type Nullable(String),
    before_content_hash Nullable(String),
    after_content_hash Nullable(String),
    before_size Nullable(Int64),
    after_size Nullable(Int64),
    diff_summary String DEFAULT '{}',
    timestamp DateTime64(3)
)
ENGINE = MergeTree()
PARTITION BY toYYYYMM(timestamp)
ORDER BY (project_id, trace_id, timestamp, id)
SETTINGS index_granularity = 8192;
```

#### Terminal Commands

```sql
CREATE TABLE terminal_commands (
    id UUID,
    project_id UUID,
    trace_id UUID,
    observation_id Nullable(UUID),
    command String,
    working_dir String,
    exit_code Nullable(Int32),
    stdout Nullable(String),
    stderr Nullable(String),
    duration_ms Int32,
    timestamp DateTime64(3)
)
ENGINE = MergeTree()
PARTITION BY toYYYYMM(timestamp)
ORDER BY (project_id, trace_id, timestamp, id)
SETTINGS index_granularity = 8192;
```

#### CI Runs

```sql
CREATE TABLE ci_runs (
    id UUID,
    project_id UUID,
    trace_id UUID,
    provider Enum8('GITHUB_ACTIONS' = 0, 'GITLAB_CI' = 1, 'JENKINS' = 2, 'CIRCLECI' = 3),
    run_id String,
    run_url String,
    workflow_name Nullable(String),
    job_name Nullable(String),
    status Enum8('PENDING' = 0, 'RUNNING' = 1, 'SUCCESS' = 2, 'FAILURE' = 3, 'CANCELLED' = 4),
    started_at Nullable(DateTime64(3)),
    finished_at Nullable(DateTime64(3)),
    trigger_event Nullable(String),
    created_at DateTime64(3) DEFAULT now64(3)
)
ENGINE = ReplacingMergeTree(created_at)
PARTITION BY toYYYYMM(created_at)
ORDER BY (project_id, trace_id, run_id, id)
SETTINGS index_granularity = 8192;
```

### 4.5 Materialized Views

#### Daily Cost Aggregation

```sql
CREATE MATERIALIZED VIEW daily_costs_mv
ENGINE = SummingMergeTree()
PARTITION BY toYYYYMM(date)
ORDER BY (project_id, date, model)
AS SELECT
    project_id,
    toDate(start_time) AS date,
    model,
    count() AS generation_count,
    sum(input_tokens) AS total_input_tokens,
    sum(output_tokens) AS total_output_tokens,
    sum(total_cost) AS total_cost
FROM observations
WHERE type = 'GENERATION' AND model IS NOT NULL
GROUP BY project_id, date, model;
```

---

## 5. Feature Requirements

### 5.1 MVP Features (Q2 2026) - Full Langfuse Parity

#### 5.1.1 Tracing & Observations

**Core Tracing:**
- OTLP ingestion endpoint (/v1/traces) with OpenTelemetry protocol support
- Legacy REST batch ingestion (Langfuse API compatibility)
- W3C Trace Context ID format (32 hex char trace ID, 16 hex char span ID)
- Distributed tracing support with context propagation
- Hierarchical observations: Traces → Spans → Generations → Events
- Automatic parent-child relationship inference

**Trace Features:**
- Sessions for multi-turn conversation grouping
- User ID tracking and per-user analytics
- Environment separation (dev/staging/prod)
- Tags and metadata for filtering
- Log levels (DEBUG, DEFAULT, WARNING, ERROR)
- Public trace sharing with unique URLs

**UI Features:**
- Timeline view with latency breakdown
- Agent graph visualization (DAG view for agentic workflows)
- Trace detail panel with input/output inspection
- Nested observation tree view
- Real-time trace streaming
- Advanced filtering and search

**Multi-Modal Support:**
- Image rendering in trace UI
- Audio playback support
- File attachment viewing
- Content masking/redaction for sensitive data

#### 5.1.2 Prompt Management

**Core Features:**
- Create prompts via UI, SDK, and API
- Text prompts (single string) and Chat prompts (message array)
- Automatic version numbering on updates
- Variable substitution with {{variable}} syntax
- Message placeholders for chat history
- Prompt composability (include other prompts)

**Version Control:**
- Immutable versions with full history
- Labels for deployment (production, staging, latest, custom)
- Protected labels (admin/owner only modification)
- Collaborative editing with user attribution
- Change tracking and audit log

**Deployment:**
- Label-based deployment (no code changes)
- SDK caching (client-side and server-side)
- Cache TTL configuration
- Fetch by label or specific version
- Webhooks for version changes
- GitHub integration for sync

**Configuration:**
- Model settings (model name, temperature, etc.)
- Tool definitions
- Response schemas
- Folders for organization

**Analytics:**
- Performance metrics per prompt version
- Cost comparison across versions
- Latency analysis
- Link prompts to traces for real-world performance

**Playground:**
- Interactive prompt testing
- Variable injection
- Model parameter adjustment
- Side-by-side version comparison

#### 5.1.3 Cost & Usage Tracking

**Model Pricing:**
- Built-in pricing for 400+ models
- Support for all major providers (OpenAI, Anthropic, Google, Azure, AWS Bedrock, etc.)
- Context-dependent pricing tiers (e.g., Claude extended context rates)
- Custom model pricing configuration
- Regular price updates

**Token Tracking:**
- Input tokens
- Output tokens
- Cache read tokens
- Cache creation tokens
- Total tokens

**Cost Calculation:**
- Automatic cost calculation on ingestion
- Background recalculation for updated prices
- Per-trace, per-user, per-session aggregation
- Daily/weekly/monthly summaries

**Usage Analytics:**
- Model usage distribution
- Token consumption trends
- Cost trends over time
- Top users by cost
- Spend alerts and thresholds

#### 5.1.4 Evaluation

**LLM-as-Judge:**
- Built-in evaluator templates:
  - Hallucination detection
  - Relevance scoring
  - Toxicity detection
  - Helpfulness rating
  - Correctness verification
  - Conciseness evaluation
- Custom evaluator prompts
- Configurable sampling rate
- Variable mapping from trace data
- Multi-model support for evaluation

**Annotation Queues:**
- Create queues for human review
- Filter traces for queue
- Assign reviewers
- Scoring interface
- Comment support
- Progress tracking

**Score Types:**
- Numeric scores (0-1, 1-10, etc.)
- Boolean scores (pass/fail)
- Categorical scores (A/B/C, good/bad)
- Multi-dimensional scoring

**User Feedback:**
- Browser SDK for thumbs up/down
- Server-side feedback API
- Feedback to score mapping
- Feedback analytics

**Score Analytics:**
- Distribution charts
- Trend analysis
- Model comparison
- Prompt version comparison
- Statistical significance testing

#### 5.1.5 Datasets & Experiments

**Dataset Management:**
- Create datasets via UI, SDK, API
- Add items manually or from production traces
- Input/expected output pairs
- Metadata per item
- Versioning and archival
- Import/export (CSV, JSON)

**Experiment Runs:**
- Run prompts against datasets
- Track all generated traces
- Compare runs side-by-side
- Automated evaluation on runs
- Progress tracking
- Parallel execution

**Analysis:**
- Aggregate scores per run
- Item-level comparison
- Regression detection
- Export results

#### 5.1.6 Dashboards & Metrics

**Custom Dashboards:**
- Widget-based layout
- Chart types: line, bar, pie, table, number
- Time range selection
- Filter configuration
- Saved views

**Pre-built Dashboards:**
- Latency overview
- Cost breakdown
- Usage statistics
- Error rates
- Model performance

**Metrics API:**
- Daily metrics export
- Integration with PostHog, Mixpanel
- Billing integration support
- Custom aggregations

#### 5.1.7 Administration

**Organizations & Projects:**
- Multi-organization support
- Projects for isolation
- Project-level settings
- Cross-project analytics (org level)

**Role-Based Access Control:**
- Organization roles: Owner, Admin, Member, Viewer
- Project-level role overrides
- Fine-grained permissions
- API key scoping

**Authentication:**
- Email/password
- Google OAuth
- GitHub OAuth
- Enterprise SSO (OIDC/SAML)
- SCIM provisioning

**Audit & Compliance:**
- Audit logs (Enterprise)
- Data retention configuration
- Data deletion support
- API key management

**LLM Connections:**
- Manage API keys for providers
- Azure OpenAI configuration
- AWS Bedrock setup
- Vertex AI integration

#### 5.1.8 Data Platform

**APIs:**
- Public REST API (OpenAPI 3.1 spec)
- GraphQL API for complex queries
- SDK query methods (Python, TypeScript)
- Rate limiting and pagination

**Export:**
- UI export (CSV, JSON)
- Scheduled export to S3/GCS/Azure Blob
- OpenAI fine-tuning format
- Custom export templates

**Integrations:**
- MCP server for AI assistant access
- Webhook delivery
- 50+ framework integrations

### 5.2 Phase 2 Features (Q3 2026) - AgentTrace Differentiation

#### 5.2.1 Git Correlation

**Bi-directional Linking:**
- Associate traces with commits
- Link commits to traces
- PR-level aggregation
- Branch filtering

**Git Metadata:**
- Commit SHA and message
- Author information
- Changed files list
- Diff statistics (additions/deletions)

**UI Features:**
- Git timeline view
- Commit → trace navigation
- Trace → commit navigation
- PR summary with agent metrics

**GitHub Integration:**
- GitHub App for automatic linking
- PR comments with trace summaries
- Status checks integration
- Webhook receivers

#### 5.2.2 Checkpoint & Rollback

**Checkpoint Creation:**
- SDK method: `checkpoint(name, description)`
- Automatic checkpoints at key points
- Configurable checkpoint triggers
- Size limits and quotas

**Storage Options:**
- S3/MinIO for full snapshots
- Git-based diffs for efficiency
- Inline storage for small checkpoints

**Rollback:**
- One-click restore from checkpoint
- Preview before restore
- Rollback confirmation
- Post-rollback trace linking

**File Manifest:**
- Track all files at checkpoint
- Content hashes for verification
- Size and type information
- Diff from previous checkpoint

#### 5.2.3 Terminal Command Logging

**Command Capture:**
- Automatic capture in CLI wrapper
- SDK method for manual logging
- Exit code tracking
- Duration measurement

**Output Handling:**
- Stdout/stderr capture
- Smart truncation for large output
- Sensitive data redaction
- Output search

**UI Features:**
- Command timeline
- Searchable command history
- Output viewer with highlighting
- Error highlighting

#### 5.2.4 File Operation Tracking

**Operation Types:**
- Create, Read, Update, Delete
- Move, Copy
- Directory operations

**Tracking Data:**
- File path and type
- Before/after content hashes
- Size changes
- Diff summaries

**UI Features:**
- File change timeline
- Diff viewer
- File tree visualization
- Filter by operation type

#### 5.2.5 Multi-Agent SDK

**Python SDK:**
- `@observe()` decorator
- `checkpoint()` method
- `git_link()` method
- `file_op()` method
- `terminal_cmd()` method
- Async support
- Context propagation

**TypeScript SDK:**
- OTEL integration
- `startActiveObservation()`
- Checkpoint and git link support
- Browser and Node.js

**Go SDK:**
- Context-based tracing
- Middleware for popular frameworks
- All AgentTrace features

**CLI Wrapper:**
- `agenttrace wrap <command>`
- Automatic trace creation
- Environment variable injection
- MCP server mode

### 5.3 Phase 3 Features (Q4 2026)

#### 5.3.1 CI/CD Integration

**GitHub Actions:**
- Pre-built action
- Automatic trace linking
- Test failure attribution
- PR comments with metrics

**GitLab CI:**
- CI component
- Pipeline integration
- Test coverage correlation

**Other Providers:**
- Jenkins plugin
- CircleCI orb
- Generic webhook integration

#### 5.3.2 IDE Extensions

**VS Code Extension:**
- Inline trace viewing
- Checkpoint quick access
- Agent cost display
- Trace navigation

**JetBrains Plugin:**
- IntelliJ, PyCharm, WebStorm
- Similar features to VS Code

#### 5.3.3 Enterprise Features

**Advanced Security:**
- SOC 2 Type II certification
- ISO 27001 compliance
- HIPAA compliance option
- SSO enforcement
- IP allowlisting

**Governance:**
- Budget allocation per team
- Cost alerts and limits
- Approval workflows
- Usage policies

**Support:**
- Dedicated support
- SLA guarantees
- Professional services
- Training sessions

---

## 6. API Design

### 6.1 Ingestion API

#### OTLP Endpoints

```
POST /v1/traces
Content-Type: application/x-protobuf
Authorization: Bearer <api-key>

# OpenTelemetry Protocol Buffer format
# Standard OTLP trace export
```

```
POST /v1/traces
Content-Type: application/json
Authorization: Bearer <api-key>

# OpenTelemetry JSON format
{
  "resourceSpans": [...]
}
```

#### Legacy REST Ingestion (Langfuse Compatibility)

```
POST /api/public/ingestion
Content-Type: application/json
Authorization: Basic <base64(public_key:secret_key)>

{
  "batch": [
    {
      "type": "trace-create",
      "body": {
        "id": "trace-uuid",
        "name": "my-trace",
        "userId": "user-123",
        "sessionId": "session-456",
        "input": {...},
        "output": {...},
        "metadata": {...}
      }
    },
    {
      "type": "generation-create",
      "body": {
        "traceId": "trace-uuid",
        "name": "llm-call",
        "model": "claude-sonnet-4-20250514",
        "input": [...],
        "output": {...},
        "usage": {
          "input": 1000,
          "output": 500
        }
      }
    }
  ]
}
```

### 6.2 Public API Endpoints

#### Traces

```
GET    /api/public/traces
GET    /api/public/traces/:id
POST   /api/public/traces
PATCH  /api/public/traces/:id
DELETE /api/public/traces/:id
```

#### Observations

```
GET    /api/public/observations
GET    /api/public/observations/:id
POST   /api/public/observations
PATCH  /api/public/observations/:id
```

#### Sessions

```
GET    /api/public/sessions
GET    /api/public/sessions/:id
```

#### Scores

```
GET    /api/public/scores
GET    /api/public/scores/:id
POST   /api/public/scores
PATCH  /api/public/scores/:id
DELETE /api/public/scores/:id
```

#### Prompts

```
GET    /api/public/prompts
GET    /api/public/prompts/:name
POST   /api/public/prompts
GET    /api/public/prompts/:name/versions
GET    /api/public/prompts/:name/versions/:version
```

#### Datasets

```
GET    /api/public/datasets
GET    /api/public/datasets/:id
POST   /api/public/datasets
GET    /api/public/datasets/:id/items
POST   /api/public/datasets/:id/items
GET    /api/public/datasets/:id/runs
POST   /api/public/datasets/:id/runs
```

#### Metrics

```
GET    /api/public/metrics/daily
GET    /api/public/metrics/usage
GET    /api/public/metrics/cost
```

### 6.3 AgentTrace Unique Endpoints

#### Checkpoints

```
GET    /api/public/checkpoints
GET    /api/public/checkpoints/:id
POST   /api/public/checkpoints
GET    /api/public/checkpoints/:id/download
POST   /api/public/checkpoints/:id/restore
```

#### Git Links

```
GET    /api/public/git-links
GET    /api/public/git-links/:id
POST   /api/public/git-links
GET    /api/public/git-links/by-commit/:sha
GET    /api/public/git-links/by-trace/:traceId
```

#### File Operations

```
GET    /api/public/file-operations
GET    /api/public/file-operations/:id
POST   /api/public/file-operations
GET    /api/public/traces/:traceId/file-operations
```

#### Terminal Commands

```
GET    /api/public/terminal-commands
GET    /api/public/terminal-commands/:id
POST   /api/public/terminal-commands
GET    /api/public/traces/:traceId/terminal-commands
```

#### CI Runs

```
GET    /api/public/ci-runs
GET    /api/public/ci-runs/:id
POST   /api/public/ci-runs
PATCH  /api/public/ci-runs/:id
GET    /api/public/traces/:traceId/ci-runs
```

### 6.4 GraphQL API

```graphql
type Query {
  # Traces
  trace(id: ID!): Trace
  traces(filter: TraceFilter, pagination: PaginationInput): TraceConnection!
  
  # Observations
  observation(id: ID!): Observation
  observations(traceId: ID!, filter: ObservationFilter): [Observation!]!
  
  # Sessions
  session(id: ID!): Session
  sessions(filter: SessionFilter, pagination: PaginationInput): SessionConnection!
  
  # Scores
  scores(filter: ScoreFilter, pagination: PaginationInput): ScoreConnection!
  
  # Prompts
  prompt(name: String!, version: Int, label: String): Prompt
  prompts(filter: PromptFilter): [Prompt!]!
  
  # Datasets
  dataset(id: ID!): Dataset
  datasets: [Dataset!]!
  datasetRun(id: ID!): DatasetRun
  
  # AgentTrace Unique
  checkpoint(id: ID!): Checkpoint
  checkpoints(traceId: ID!): [Checkpoint!]!
  gitLink(id: ID!): GitLink
  gitLinks(filter: GitLinkFilter): [GitLink!]!
  fileOperations(traceId: ID!): [FileOperation!]!
  terminalCommands(traceId: ID!): [TerminalCommand!]!
  ciRuns(traceId: ID!): [CIRun!]!
  
  # Metrics
  dailyMetrics(startDate: Date!, endDate: Date!): [DailyMetric!]!
  costByModel(startDate: Date!, endDate: Date!): [ModelCost!]!
}

type Mutation {
  # Traces
  createTrace(input: CreateTraceInput!): Trace!
  updateTrace(id: ID!, input: UpdateTraceInput!): Trace!
  
  # Scores
  createScore(input: CreateScoreInput!): Score!
  
  # Prompts
  createPrompt(input: CreatePromptInput!): PromptVersion!
  updatePromptLabels(name: String!, version: Int!, labels: [String!]!): PromptVersion!
  
  # Datasets
  createDataset(input: CreateDatasetInput!): Dataset!
  createDatasetItem(datasetId: ID!, input: CreateDatasetItemInput!): DatasetItem!
  createDatasetRun(datasetId: ID!, input: CreateDatasetRunInput!): DatasetRun!
  
  # AgentTrace Unique
  createCheckpoint(input: CreateCheckpointInput!): Checkpoint!
  restoreCheckpoint(id: ID!): RestoreResult!
  createGitLink(input: CreateGitLinkInput!): GitLink!
  createFileOperation(input: CreateFileOperationInput!): FileOperation!
  createTerminalCommand(input: CreateTerminalCommandInput!): TerminalCommand!
  createCIRun(input: CreateCIRunInput!): CIRun!
  updateCIRun(id: ID!, input: UpdateCIRunInput!): CIRun!
}
```

### 6.5 BFF API Routes (Next.js)

```typescript
// /app/api/auth/[...nextauth]/route.ts
// NextAuth.js authentication endpoints

// /app/api/graphql/route.ts
// GraphQL proxy to Go backend with auth context

// /app/api/dashboard/route.ts
// Aggregated dashboard data

// /app/api/traces/[id]/timeline/route.ts
// Trace timeline with computed layout

// /app/api/traces/[id]/graph/route.ts
// Agent graph visualization data

// /app/api/prompts/[name]/compile/route.ts
// Server-side prompt compilation with caching

// /app/api/export/route.ts
// Trigger export jobs
```

---

## 7. SDK Design

### 7.1 Python SDK

```python
from agenttrace import AgentTrace, observe, generation, checkpoint, git_link

# Initialize
at = AgentTrace(
    public_key="pk-at-xxx",
    secret_key="sk-at-xxx",
    host="https://api.agenttrace.dev"  # or self-hosted
)

# Decorator-based tracing
@observe(name="code-generation")
def generate_code(prompt: str, context: dict):
    # Automatic trace creation
    
    # Create checkpoint before major changes
    checkpoint(
        name="pre-generation",
        description="State before code generation"
    )
    
    # LLM call with generation tracking
    with generation(name="claude-call", model="claude-sonnet-4-20250514") as gen:
        response = anthropic.messages.create(
            model="claude-sonnet-4-20250514",
            messages=[{"role": "user", "content": prompt}]
        )
        gen.output = response.content
        gen.usage = {
            "input": response.usage.input_tokens,
            "output": response.usage.output_tokens
        }
    
    # Link to git commit
    git_link(
        commit_sha="abc123",
        branch="feature/new-feature",
        repository_url="https://github.com/org/repo"
    )
    
    # Log file operation
    file_op(
        operation="UPDATE",
        file_path="src/main.py",
        before_hash="...",
        after_hash="..."
    )
    
    # Log terminal command
    terminal_cmd(
        command="python -m pytest",
        exit_code=0,
        duration_ms=5000
    )
    
    return response.content

# Prompt management
prompt = at.get_prompt("code-review", label="production")
compiled = prompt.compile(code=user_code, language="python")

# Evaluation
at.score(
    trace_id="...",
    name="code-quality",
    value=0.85,
    comment="Good structure, minor issues"
)

# Dataset management
dataset = at.create_dataset(name="code-review-examples")
dataset.add_item(
    input={"code": "..."},
    expected_output={"review": "..."}
)
```

### 7.2 TypeScript SDK

```typescript
import { AgentTrace, observe, startGeneration, checkpoint } from '@agenttrace/sdk';

// Initialize
const at = new AgentTrace({
  publicKey: 'pk-at-xxx',
  secretKey: 'sk-at-xxx',
  baseUrl: 'https://api.agenttrace.dev'
});

// Function wrapper
const generateCode = observe(
  { name: 'code-generation' },
  async (prompt: string, context: Record<string, any>) => {
    // Create checkpoint
    await checkpoint({
      name: 'pre-generation',
      description: 'State before code generation'
    });
    
    // Start generation
    const gen = startGeneration({
      name: 'claude-call',
      model: 'claude-sonnet-4-20250514',
      input: [{ role: 'user', content: prompt }]
    });
    
    try {
      const response = await anthropic.messages.create({
        model: 'claude-sonnet-4-20250514',
        messages: [{ role: 'user', content: prompt }]
      });
      
      gen.end({
        output: response.content,
        usage: {
          input: response.usage.input_tokens,
          output: response.usage.output_tokens
        }
      });
      
      return response.content;
    } catch (error) {
      gen.end({ error });
      throw error;
    }
  }
);

// Prompt management
const prompt = await at.getPrompt('code-review', { label: 'production' });
const compiled = prompt.compile({ code: userCode, language: 'python' });

// User feedback (browser)
import { AgentTraceBrowser } from '@agenttrace/browser';

const atBrowser = new AgentTraceBrowser({ publicKey: 'pk-at-xxx' });
atBrowser.score({
  traceId: '...',
  name: 'user-feedback',
  value: 1 // thumbs up
});
```

### 7.3 Go SDK

```go
package main

import (
    "context"
    at "github.com/agenttrace/agenttrace-go"
)

func main() {
    // Initialize
    client := at.NewClient(at.Config{
        PublicKey: "pk-at-xxx",
        SecretKey: "sk-at-xxx",
        BaseURL:   "https://api.agenttrace.dev",
    })
    defer client.Shutdown()
    
    // Start trace
    ctx, trace := client.StartTrace(context.Background(), at.TraceConfig{
        Name:   "code-generation",
        UserID: "user-123",
    })
    defer trace.End()
    
    // Create checkpoint
    client.Checkpoint(ctx, at.CheckpointConfig{
        Name:        "pre-generation",
        Description: "State before code generation",
    })
    
    // Start generation
    ctx, gen := client.StartGeneration(ctx, at.GenerationConfig{
        Name:  "claude-call",
        Model: "claude-sonnet-4-20250514",
        Input: messages,
    })
    
    response, err := callClaude(messages)
    if err != nil {
        gen.Error(err)
        return
    }
    
    gen.End(at.GenerationOutput{
        Output: response.Content,
        Usage: at.Usage{
            Input:  response.Usage.InputTokens,
            Output: response.Usage.OutputTokens,
        },
    })
    
    // Git link
    client.GitLink(ctx, at.GitLinkConfig{
        CommitSHA:     "abc123",
        Branch:        "feature/new-feature",
        RepositoryURL: "https://github.com/org/repo",
    })
}

// HTTP middleware
func TracingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        ctx, span := atClient.StartSpan(r.Context(), at.SpanConfig{
            Name: r.URL.Path,
        })
        defer span.End()
        
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

### 7.4 CLI Agent Wrapper

```bash
# Wrap any CLI agent
agenttrace wrap -- claude-code "implement a REST API"

# With configuration
agenttrace wrap \
  --name "claude-code-session" \
  --user-id "developer-123" \
  --tags "refactor,api" \
  -- claude-code "refactor the authentication module"

# Auto-checkpoint on file changes
agenttrace wrap \
  --auto-checkpoint \
  --checkpoint-interval 60s \
  -- aider "add unit tests"

# MCP server mode (for Claude Desktop integration)
agenttrace mcp-server --port 3333

# Git auto-linking
agenttrace wrap \
  --git-auto-link \
  --git-commit-on-complete \
  -- cursor "fix the bug in auth.py"
```

### 7.5 Prompt Management Examples

```python
# Python prompt management

# Create prompt
at.create_prompt(
    name="code-review",
    type="chat",
    prompt=[
        {"role": "system", "content": "You are an expert code reviewer."},
        {"role": "user", "content": "Review this {{language}} code:\n\n{{code}}"}
    ],
    config={
        "model": "claude-sonnet-4-20250514",
        "temperature": 0.3,
        "max_tokens": 2000
    },
    labels=["staging"]
)

# Fetch and use
prompt = at.get_prompt("code-review", label="production")
messages = prompt.compile(language="Python", code=user_code)

# Use with any LLM client
response = anthropic.messages.create(
    model=prompt.config.get("model", "claude-sonnet-4-20250514"),
    messages=messages,
    temperature=prompt.config.get("temperature", 0.7),
    max_tokens=prompt.config.get("max_tokens", 1000)
)
```

### 7.6 Evaluation Examples

```python
# Run evaluation on dataset
from agenttrace import evaluate

results = evaluate(
    dataset_name="code-review-examples",
    prompt_name="code-review",
    evaluators=[
        "hallucination",  # Built-in
        "relevance",      # Built-in
        {                 # Custom
            "name": "code-correctness",
            "prompt": "Does the review correctly identify issues? Score 0-1.",
            "model": "claude-sonnet-4-20250514"
        }
    ],
    model="claude-sonnet-4-20250514",
    run_name="v2-test"
)

print(f"Average hallucination score: {results.scores['hallucination'].mean()}")
print(f"Average relevance score: {results.scores['relevance'].mean()}")
```

---

## 8. Business Model

### 8.1 Pricing Tiers

| Tier | Price | Traces/Month | Features |
|------|-------|--------------|----------|
| **Open Source** | Free | Unlimited | Self-hosted, all features, community support |
| **Cloud Starter** | Free | 50,000 | Cloud hosted, 7-day retention, basic support |
| **Pro** | $49/seat/mo | 500,000 | 30-day retention, SSO, priority support |
| **Team** | $99/seat/mo | 2,000,000 | 90-day retention, advanced analytics, SLA |
| **Enterprise** | $200-500/seat/mo | Custom | Custom retention, HIPAA, dedicated support |

### 8.2 Revenue Projections

| Year | Cloud Customers | Avg Revenue/Customer | ARR |
|------|-----------------|---------------------|-----|
| Year 1 | 100 | $267/mo | $320K |
| Year 2 | 500 | $617/mo | $3.7M |
| Year 3 | 1,500 | $950/mo | $17.1M |

### 8.3 Exit Strategy

**Primary Targets:**
- **GitHub/Microsoft** - Natural fit for Copilot ecosystem, $50-100M acquisition potential
- **Datadog** - Expanding LLM observability, $75-150M based on ARR multiple
- **Arize AI / LangChain** - Strategic acquisition for technology, $30-50M

**Alternative Paths:**
- Series A growth path ($20-50M raise at $100-200M valuation)
- Acqui-hire by major AI lab (Anthropic, OpenAI)
- Strategic acquisition by Atlassian for developer tools integration

### 8.4 Competitive Positioning

| Feature | AgentTrace | Langfuse | LangSmith | Datadog | Arize |
|---------|------------|----------|-----------|---------|-------|
| Open Source | ✅ MIT | ✅ MIT | ❌ | ❌ | ❌ |
| Self-Hosted | ✅ | ✅ | ❌ | ❌ | ✅ |
| Git Correlation | ✅ | ❌ | ❌ | ❌ | ❌ |
| Code Checkpoints | ✅ | ❌ | ❌ | ❌ | ❌ |
| Terminal Logging | ✅ | ❌ | ❌ | ❌ | ❌ |
| CI/CD Integration | ✅ | ❌ | ❌ | ✅ | ❌ |
| Prompt Management | ✅ | ✅ | ✅ | ❌ | ❌ |
| LLM-as-Judge | ✅ | ✅ | ✅ | ❌ | ✅ |
| Datasets/Experiments | ✅ | ✅ | ✅ | ❌ | ✅ |
| ClickHouse Backend | ✅ | ✅ v3 | ❌ | ❌ | ❌ |
| Go Backend | ✅ | ❌ Node | ❌ Python | ❌ | ❌ |

---

## 9. Development Timeline

### 9.1 MVP Phase (Weeks 1-16)

**Weeks 1-2: Foundation**
- Repository setup (Go backend, Next.js frontend)
- Docker Compose development environment
- PostgreSQL schema for metadata
- ClickHouse schema for traces
- Basic CI/CD pipeline

**Weeks 3-4: Core Ingestion**
- OTLP ingestion endpoint
- Legacy REST batch ingestion
- Basic trace/observation storage
- Redis queue setup

**Weeks 5-6: Dashboard Foundation**
- NextAuth authentication
- Project/organization management
- Basic dashboard layout
- Navigation and routing

**Weeks 7-8: Trace Exploration**
- Trace list view with filtering
- Trace detail page
- Timeline view
- Observation tree

**Weeks 9-10: Prompt Management**
- Prompt CRUD operations
- Version control
- Label management
- SDK integration
- Playground (basic)

**Weeks 11-12: Evaluation**
- Score ingestion and storage
- LLM-as-Judge execution
- Annotation queue UI
- Score analytics

**Weeks 13-14: Cost Tracking**
- Model pricing database
- Cost calculation worker
- Usage dashboards
- Spend analytics

**Weeks 15-16: SDK & Launch Prep**
- Python SDK completion
- TypeScript SDK completion
- Documentation site
- Self-hosting guide
- Beta launch preparation

### 9.2 Phase 2 (Weeks 17-28)

**Weeks 17-19: Git Correlation**
- Git link API and storage
- GitHub App integration
- Bi-directional navigation
- PR summary comments

**Weeks 20-22: Checkpoints**
- Checkpoint creation API
- S3/MinIO storage
- Restore functionality
- File manifest tracking

**Weeks 23-25: Terminal & File Logging**
- Terminal command logging
- File operation tracking
- CLI wrapper tool
- UI integration

**Weeks 26-28: Integration Polish**
- SDK refinement
- Performance optimization
- Documentation updates
- Community feedback integration

### 9.3 Phase 3 (Weeks 29-40)

**Weeks 29-32: CI/CD Integration**
- GitHub Actions action
- GitLab CI component
- Test failure attribution
- Pipeline correlation

**Weeks 33-36: IDE Extensions**
- VS Code extension
- JetBrains plugin
- Inline trace viewing
- Quick actions

**Weeks 37-40: Enterprise**
- SOC 2 preparation
- SSO enforcement
- Advanced RBAC
- Audit logging
- Support infrastructure

### 9.4 Technology Stack Summary

**Backend (Go):**
- Go 1.22+
- Fiber/Echo HTTP framework
- gqlgen GraphQL
- Asynq job queue
- sqlx/pgx PostgreSQL
- clickhouse-go
- go-redis
- AWS SDK v2

**Frontend (Next.js):**
- Next.js 15 (App Router)
- React 19
- TypeScript 5
- NextAuth.js
- TanStack Query
- Tailwind CSS
- Radix UI
- Recharts / Tremor
- React Flow (graphs)
- Monaco Editor

**Infrastructure:**
- ClickHouse 24+
- PostgreSQL 16
- Redis 7
- S3/MinIO
- Docker & Docker Compose
- Kubernetes (optional)
- GitHub Actions CI/CD

---

## 10. Risks and Mitigations

### 10.1 Market Risks

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| GitHub builds native Copilot observability | High | High | Focus on open-source, multi-agent support, deep integration |
| Datadog adds agent monitoring | High | Medium | Differentiate on coding-specific features, open-source |
| Langfuse adds coding features | Medium | Medium | Move fast, establish community, unique architecture |

### 10.2 Technical Risks

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| OTEL GenAI conventions change | Medium | Medium | Abstract OTEL layer, quick adaptation |
| ClickHouse licensing changes | Low | High | Document migration path, consider alternatives |
| Next.js + Go complexity | Low | Medium | Clear service boundaries, shared schema |

### 10.3 Business Risks

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| Solo founder execution | Medium | High | Seek co-founder, prioritize ruthlessly |
| Cloud costs at scale | Medium | Medium | Efficient Go backend, ClickHouse optimization |
| Slow enterprise adoption | Medium | Medium | Focus on developer community first |

---

## 11. Open Questions

### 11.1 Brand & Domain
- Trademark search for "AgentTrace" needed
- Domain availability: agenttrace.dev, agenttrace.io?
- Alternative names if conflicts exist

### 11.2 Initial Agent Focus
- Start with Claude Code only (deep integration)?
- Or multi-platform from day one (broader market)?
- Partnership opportunities with agent vendors?

### 11.3 Langfuse API Compatibility
- Full compatibility for easy migration?
- Or clean break with simpler API?
- Cost of maintaining compatibility layer?

### 11.4 Checkpoint Storage
- Full file snapshots (simple, storage-heavy)?
- Git-based diffs (complex, efficient)?
- Hybrid approach?

### 11.5 License Choice
- MIT (maximum adoption)?
- Apache 2.0 (patent protection)?
- BSL (commercial protection)?

### 11.6 Funding Strategy
- Bootstrap initially?
- Pre-seed for faster execution?
- Accelerator programs (Y Combinator, etc.)?

### 11.7 Team Composition
- When to seek co-founder?
- Skills needed (frontend, DevRel, sales)?
- Contractor vs. full-time early hires?

### 11.8 Cloud vs. Self-Hosted Emphasis
- Push self-hosted for community growth?
- Focus on cloud for revenue?
- How to balance both?

---

## 12. Appendices

### 12.1 Go Dependencies

```go
// go.mod
module github.com/agenttrace/agenttrace

go 1.22

require (
    // HTTP Framework
    github.com/gofiber/fiber/v2 v2.52.0
    
    // GraphQL
    github.com/99designs/gqlgen v0.17.45
    
    // Database
    github.com/jackc/pgx/v5 v5.5.5
    github.com/ClickHouse/clickhouse-go/v2 v2.20.0
    
    // Queue
    github.com/hibiken/asynq v0.24.1
    
    // Redis
    github.com/redis/go-redis/v9 v9.5.1
    
    // AWS
    github.com/aws/aws-sdk-go-v2 v1.25.0
    
    // OpenTelemetry
    go.opentelemetry.io/otel v1.24.0
    go.opentelemetry.io/collector v0.96.0
    
    // Auth
    github.com/golang-jwt/jwt/v5 v5.2.1
    
    // Utilities
    github.com/google/uuid v1.6.0
    go.uber.org/zap v1.27.0
    github.com/spf13/viper v1.18.2
)
```

### 12.2 Next.js Dependencies

```json
{
  "dependencies": {
    "next": "15.0.0",
    "react": "19.0.0",
    "react-dom": "19.0.0",
    "next-auth": "5.0.0",
    "@tanstack/react-query": "5.28.0",
    "tailwindcss": "3.4.0",
    "@radix-ui/react-dropdown-menu": "2.0.6",
    "@radix-ui/react-dialog": "1.0.5",
    "@radix-ui/react-tabs": "1.0.4",
    "recharts": "2.12.0",
    "@tremor/react": "3.14.0",
    "reactflow": "11.10.0",
    "@monaco-editor/react": "4.6.0",
    "zod": "3.22.4",
    "date-fns": "3.3.0",
    "lucide-react": "0.350.0"
  },
  "devDependencies": {
    "typescript": "5.4.0",
    "@types/react": "18.2.0",
    "@types/node": "20.11.0",
    "eslint": "8.57.0",
    "prettier": "3.2.0"
  }
}
```

### 12.3 Docker Compose Configuration

```yaml
# docker-compose.yml
version: '3.8'

services:
  web:
    build:
      context: ./web
      dockerfile: Dockerfile
    ports:
      - "3000:3000"
    environment:
      - NEXT_PUBLIC_API_URL=http://api:8080
      - NEXTAUTH_URL=http://localhost:3000
      - NEXTAUTH_SECRET=${NEXTAUTH_SECRET}
      - DATABASE_URL=postgresql://postgres:postgres@postgres:5432/agenttrace
    depends_on:
      - api
      - postgres

  api:
    build:
      context: ./api
      dockerfile: Dockerfile
    ports:
      - "8080:8080"
    environment:
      - DATABASE_URL=postgresql://postgres:postgres@postgres:5432/agenttrace
      - CLICKHOUSE_URL=clickhouse://clickhouse:9000/agenttrace
      - REDIS_URL=redis://redis:6379
      - S3_ENDPOINT=http://minio:9000
      - S3_ACCESS_KEY=${S3_ACCESS_KEY}
      - S3_SECRET_KEY=${S3_SECRET_KEY}
      - S3_BUCKET=agenttrace
    depends_on:
      - postgres
      - clickhouse
      - redis
      - minio

  worker:
    build:
      context: ./api
      dockerfile: Dockerfile.worker
    environment:
      - DATABASE_URL=postgresql://postgres:postgres@postgres:5432/agenttrace
      - CLICKHOUSE_URL=clickhouse://clickhouse:9000/agenttrace
      - REDIS_URL=redis://redis:6379
      - S3_ENDPOINT=http://minio:9000
      - S3_ACCESS_KEY=${S3_ACCESS_KEY}
      - S3_SECRET_KEY=${S3_SECRET_KEY}
    depends_on:
      - redis
      - clickhouse
      - postgres

  postgres:
    image: postgres:16-alpine
    ports:
      - "5432:5432"
    environment:
      - POSTGRES_USER=postgres
      - POSTGRES_PASSWORD=postgres
      - POSTGRES_DB=agenttrace
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./api/migrations:/docker-entrypoint-initdb.d

  clickhouse:
    image: clickhouse/clickhouse-server:24.2
    ports:
      - "8123:8123"
      - "9000:9000"
    environment:
      - CLICKHOUSE_DB=agenttrace
      - CLICKHOUSE_USER=default
      - CLICKHOUSE_PASSWORD=clickhouse
    volumes:
      - clickhouse_data:/var/lib/clickhouse
      - ./api/migrations/clickhouse:/docker-entrypoint-initdb.d

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
    command: redis-server --appendonly yes

  minio:
    image: minio/minio:latest
    ports:
      - "9001:9000"
      - "9002:9001"
    environment:
      - MINIO_ROOT_USER=${S3_ACCESS_KEY:-minioadmin}
      - MINIO_ROOT_PASSWORD=${S3_SECRET_KEY:-minioadmin}
    volumes:
      - minio_data:/data
    command: server /data --console-address ":9001"

volumes:
  postgres_data:
  clickhouse_data:
  redis_data:
  minio_data:
```

---

## Document History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | Jan 2026 | Jose Baena | Initial PRD |
| 2.0 | Jan 2026 | Jose Baena | Go backend architecture |
| 3.0 | Jan 2026 | Jose Baena | Full Langfuse parity, Next.js + Go split architecture |

---

*This document is confidential and intended for internal use only.*
