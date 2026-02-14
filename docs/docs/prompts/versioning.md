---
sidebar_position: 2
title: "Prompt Versioning"
description: "How prompt versioning works in AgentTrace — immutable versions, pinning, and version history."
---

# Prompt Versioning

Every change to a prompt in AgentTrace creates a new immutable version. This gives you a complete history of every iteration, the ability to pin deployments to specific versions, and instant rollback when needed.

## How Versioning Works

When you create a prompt, it starts at **version 1**. Each subsequent edit — whether through the dashboard, API, or SDK — increments the version number automatically.

```
Prompt: "code-review"
├── Version 1  (created Jan 10)  ← original
├── Version 2  (created Jan 14)  ← improved system prompt
├── Version 3  (created Jan 18)  ← added output format
└── Version 4  (created Jan 22)  ← current latest
```

### Immutability

Versions are **immutable** once created. You cannot edit an existing version — you can only create a new one. This ensures:

- **Reproducibility** — you can always recreate the exact prompt used in any past generation
- **Audit trail** — every change is recorded with a timestamp and optional commit message
- **Safe rollback** — revert to any previous version by moving a label

## Creating New Versions

### Via API

```bash
curl -X POST "https://api.agenttrace.io/v1/prompts/code-review/versions" \
  -H "Authorization: Bearer at-your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "content": "Review the following {{language}} code for bugs:\n\n{{code}}",
    "labels": ["staging"],
    "commitMessage": "Added language parameter for better context"
  }'
```

### Via Dashboard

1. Navigate to **Prompts** and select your prompt
2. Click **Edit** to open the editor
3. Make your changes and click **Save as New Version**
4. Optionally add a commit message describing the change

## Pinning Versions

You can fetch a prompt pinned to a specific version, ensuring your application always uses the exact same prompt text regardless of newer versions.

### Pin in SDK

```python
# Always use version 3, regardless of what "latest" or labels point to
prompt = client.get_prompt("code-review", version=3)
```

```typescript
const prompt = await getPrompt({ name: 'code-review', version: 3 });
```

### Pin in API

```bash
curl "https://api.agenttrace.io/v1/prompts/code-review?version=3" \
  -H "Authorization: Bearer at-your-api-key"
```

## Using "Latest"

When you fetch a prompt without specifying a version or label, AgentTrace returns the version pointed to by the `production` label. If no `production` label exists, it returns the most recently created version.

```python
# Returns production-labeled version, or latest if no production label
prompt = client.get_prompt("code-review")
```

## Version History

### Listing Versions

```bash
curl "https://api.agenttrace.io/v1/prompts/code-review/versions" \
  -H "Authorization: Bearer at-your-api-key"
```

Response:

```json
{
  "data": [
    {
      "version": 4,
      "content": "Review the following {{language}} code...",
      "labels": ["staging"],
      "variables": ["language", "code"],
      "commitMessage": "Added language parameter",
      "createdAt": "2024-01-22T10:00:00Z"
    },
    {
      "version": 3,
      "content": "Review the following code...",
      "labels": ["production"],
      "variables": ["code"],
      "commitMessage": "Added structured output format",
      "createdAt": "2024-01-18T10:00:00Z"
    }
  ],
  "totalCount": 4
}
```

### Dashboard View

The version history in the dashboard shows:

- **Diff view** — see exactly what changed between versions
- **Labels** — which labels point to each version
- **Metrics** — usage count and quality scores per version
- **Commit messages** — why each change was made

## Comparing Versions

Use the dashboard's comparison view to see side-by-side diffs between any two versions. This is especially useful when debugging prompt regressions.

## Best Practices

1. **Always include commit messages** — they make the version history meaningful
2. **Don't skip versions** — if a version is bad, create a new one rather than trying to "fix" it
3. **Pin versions in production code** — use labels rather than `latest` to avoid unexpected changes
4. **Review version metrics** — check quality scores before promoting a version

## Related

- [Labels](./labels.md) — use labels to manage version promotion
- [Playground](./playground.md) — test versions before promoting
- [Overview](./overview.md) — prompt management concepts
