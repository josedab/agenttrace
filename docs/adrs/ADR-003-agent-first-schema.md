# ADR-003: Agent-First Schema Design

## Status

Accepted

## Context

AgentTrace targets AI coding agents (Claude Code, GitHub Copilot Workspace, Cursor, Aider) rather than general LLM applications. These agents have unique observability requirements that generic LLM tracing platforms don't address:

1. **Git integration**: Agents make commits; understanding which trace led to which code change is critical for debugging
2. **Code checkpoints**: Ability to snapshot and restore code state during agent sessions
3. **File operations**: Tracking which files were read, written, or modified
4. **Terminal commands**: Logging shell commands executed by agents
5. **CI/CD correlation**: Linking traces to CI pipeline runs

Existing platforms like Langfuse treat these as generic metadata, losing the ability to query and correlate them effectively.

### Alternatives Considered

1. **Generic metadata fields**
   - Pros: Flexible, no schema changes needed
   - Cons: No type safety, poor query performance, no relationships

2. **JSON columns with partial indexing**
   - Pros: Semi-structured flexibility
   - Cons: Complex queries, inconsistent data, limited aggregations

3. **First-class entities with dedicated tables** (chosen)
   - Pros: Type safety, efficient queries, relationships, clear semantics
   - Cons: Schema rigidity, migration overhead, more tables to maintain

## Decision

We implement **agent-specific features as first-class entities** in the database schema:

### Core Agent Entities

```sql
-- Git Links: Bi-directional trace ↔ commit correlation
CREATE TABLE git_links (
    id UUID PRIMARY KEY,
    trace_id UUID NOT NULL,
    project_id UUID NOT NULL,
    commit_sha String NOT NULL,
    repository_url String,
    branch String,
    commit_message String,
    author String,
    committed_at DateTime64(3),
    link_type Enum('caused_by', 'resulted_in', 'related_to')
);

-- Checkpoints: Code state snapshots during agent sessions
CREATE TABLE checkpoints (
    id UUID PRIMARY KEY,
    trace_id UUID NOT NULL,
    observation_id UUID,
    project_id UUID NOT NULL,
    name String NOT NULL,
    description String,
    snapshot_type Enum('full', 'incremental'),
    storage_path String NOT NULL,  -- MinIO path
    file_count UInt32,
    total_size_bytes UInt64,
    created_at DateTime64(3)
);

-- File Operations: Read/write/modify tracking
CREATE TABLE file_operations (
    id UUID PRIMARY KEY,
    trace_id UUID NOT NULL,
    observation_id UUID,
    project_id UUID NOT NULL,
    operation_type Enum('read', 'write', 'modify', 'delete', 'create'),
    file_path String NOT NULL,
    content_hash String,  -- SHA-256 of content
    size_bytes UInt64,
    diff String,  -- For modifications
    timestamp DateTime64(3)
);

-- Terminal Commands: Shell execution logging
CREATE TABLE terminal_commands (
    id UUID PRIMARY KEY,
    trace_id UUID NOT NULL,
    observation_id UUID,
    project_id UUID NOT NULL,
    command String NOT NULL,
    working_directory String,
    exit_code Int32,
    stdout String,
    stderr String,
    duration_ms UInt64,
    started_at DateTime64(3),
    completed_at DateTime64(3)
);

-- CI Runs: Pipeline correlation
CREATE TABLE ci_runs (
    id UUID PRIMARY KEY,
    trace_id UUID,
    project_id UUID NOT NULL,
    provider Enum('github_actions', 'gitlab_ci', 'jenkins', 'circleci'),
    run_id String NOT NULL,
    workflow_name String,
    job_name String,
    status Enum('pending', 'running', 'success', 'failure', 'cancelled'),
    started_at DateTime64(3),
    completed_at DateTime64(3),
    url String
);
```

### SDK Integration

Each SDK provides dedicated methods for agent features:

```python
# Python SDK
from agenttrace import git, checkpoint, fileops, terminal

# Link a commit to the current trace
git.link_commit(sha="abc123", link_type="resulted_in")

# Create a code checkpoint
checkpoint.create(name="before-refactor", description="Pre-refactor state")

# Track file operations (automatic with @observe)
fileops.track_write("/src/main.py", content)

# Log terminal commands
terminal.log("npm test", exit_code=0, stdout=output)
```

## Consequences

### Positive

- **Queryability**: Can answer "which traces modified this file?" or "what commits came from this session?"
- **Type safety**: Strongly typed fields prevent data quality issues
- **Performance**: Dedicated indexes enable fast queries on agent-specific data
- **Correlation**: Bi-directional linking between traces, commits, and CI runs
- **Differentiation**: Unique capabilities not available in Langfuse or LangSmith
- **Debugging power**: Replay agent sessions with full context (code changes, commands, files)

### Negative

- **Schema complexity**: More tables to maintain and migrate
- **Storage overhead**: Storing file contents and command outputs can be expensive
- **SDK complexity**: More methods to implement and document across all languages
- **Migration risk**: Schema changes require careful migration planning
- **Learning curve**: Users must understand agent-specific concepts

### Neutral

- Storage of file contents is optional (can store hashes only)
- Terminal output can be truncated for large outputs
- Checkpoints use object storage (MinIO) to avoid bloating the database

## Use Cases Enabled

1. **"What code did this agent session produce?"**
   - Query git_links for all commits associated with a trace

2. **"Roll back to before the agent broke things"**
   - Restore from checkpoint created at session start

3. **"Why did the agent modify this file?"**
   - Trace file_operations back to the triggering observation

4. **"Did the CI pass after the agent's changes?"**
   - Correlate ci_runs with git_links from the same trace

5. **"What commands did the agent run?"**
   - Query terminal_commands with full stdout/stderr

## References

- [Claude Code Architecture](https://docs.anthropic.com/en/docs/claude-code)
- [Cursor Agent Patterns](https://cursor.sh/docs)
- [Aider Git Integration](https://aider.chat/docs/git.html)
