---
sidebar_position: 2
title: "Audit Logs"
description: "Comprehensive audit logging for compliance — tracks user actions, API key usage, and configuration changes."
---

# Audit Logs

AgentTrace Enterprise provides comprehensive audit logging that tracks every significant action in your organization. Audit logs are essential for compliance, security investigations, and operational visibility.

## What Is Logged

### User Actions

| Event | Description |
|-------|-------------|
| `user.login` | User signed in (method: password, SSO, OAuth) |
| `user.logout` | User signed out |
| `user.login_failed` | Failed login attempt |
| `user.created` | New user account created |
| `user.updated` | User profile or role updated |
| `user.deleted` | User account deleted |
| `user.invited` | User invited to organization |

### API Key Events

| Event | Description |
|-------|-------------|
| `api_key.created` | New API key generated |
| `api_key.deleted` | API key revoked |
| `api_key.used` | API key used for authentication (sampled) |
| `api_key.rotated` | API key rotated |

### Project & Resource Actions

| Event | Description |
|-------|-------------|
| `project.created` | New project created |
| `project.updated` | Project settings changed |
| `project.deleted` | Project deleted |
| `prompt.created` | New prompt created |
| `prompt.version_created` | New prompt version published |
| `prompt.label_changed` | Prompt label moved to different version |
| `dataset.created` | New evaluation dataset created |
| `dataset.deleted` | Dataset deleted |

### Configuration Changes

| Event | Description |
|-------|-------------|
| `org.settings_updated` | Organization settings changed |
| `org.sso_configured` | SSO configuration updated |
| `org.sso_enabled` | SSO enabled for organization |
| `org.sso_disabled` | SSO disabled |
| `org.member_role_changed` | Member role updated |
| `org.retention_policy_changed` | Data retention policy updated |

## Viewing Audit Logs

### Dashboard

1. Navigate to **Settings > Organization > Audit Logs**
2. Use filters to narrow results by:
   - **Date range** — last 24 hours, 7 days, 30 days, or custom
   - **Actor** — specific user or API key
   - **Event type** — category of action
   - **Resource** — specific project or prompt

### API

```bash
# List recent audit log entries
curl "https://api.agenttrace.io/v1/organizations/:orgId/audit-logs" \
  -H "Authorization: Bearer at-your-api-key"

# Filter by event type and date range
curl "https://api.agenttrace.io/v1/organizations/:orgId/audit-logs?\
event=user.login&\
from=2024-01-01T00:00:00Z&\
to=2024-01-31T23:59:59Z" \
  -H "Authorization: Bearer at-your-api-key"
```

### Response Format

```json
{
  "data": [
    {
      "id": "evt_abc123",
      "timestamp": "2024-01-15T14:30:00Z",
      "event": "prompt.label_changed",
      "actor": {
        "type": "user",
        "id": "user_xyz",
        "email": "alice@company.com",
        "name": "Alice Smith"
      },
      "resource": {
        "type": "prompt",
        "id": "prompt_456",
        "name": "code-review"
      },
      "details": {
        "label": "production",
        "fromVersion": 2,
        "toVersion": 3
      },
      "ipAddress": "203.0.113.42",
      "userAgent": "Mozilla/5.0..."
    }
  ],
  "totalCount": 1247,
  "hasMore": true
}
```

## Export

### CSV Export

Export audit logs for offline analysis or compliance reporting:

```bash
curl "https://api.agenttrace.io/v1/organizations/:orgId/audit-logs/export?\
format=csv&\
from=2024-01-01T00:00:00Z&\
to=2024-01-31T23:59:59Z" \
  -H "Authorization: Bearer at-your-api-key" \
  -o audit_logs_january.csv
```

### SIEM Integration

Forward audit logs to your SIEM (Splunk, Datadog, Elasticsearch) via webhook:

```bash
curl -X POST "https://api.agenttrace.io/v1/organizations/:orgId/audit-logs/webhooks" \
  -H "Authorization: Bearer at-your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://your-siem.com/api/events",
    "events": ["user.login", "user.login_failed", "api_key.created"],
    "headers": {"Authorization": "Bearer your-siem-token"}
  }'
```

## Retention

Audit logs are retained for a configurable period:

| Plan | Default Retention | Maximum |
|------|-------------------|---------|
| Cloud | 90 days | 1 year |
| Enterprise | 1 year | Unlimited |
| Self-Hosted | Unlimited | Unlimited |

## Best Practices

1. **Review regularly** — check audit logs weekly for unexpected activity
2. **Set up alerts** — configure webhooks for critical events like `user.login_failed` or `api_key.created`
3. **Export for compliance** — regularly export logs to long-term storage for audit requirements
4. **Limit admin access** — fewer admins means fewer high-privilege events to review

## Related

- [RBAC](./rbac.md) — role-based access control
- [Compliance](./compliance.md) — data retention and GDPR
- [SSO](./sso.md) — single sign-on configuration
