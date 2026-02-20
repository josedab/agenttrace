---
sidebar_position: 3
title: "Role-Based Access Control"
description: "RBAC with Owner, Admin, and Member roles, plus per-project permissions."
---

# Role-Based Access Control (RBAC)

AgentTrace Enterprise provides role-based access control to manage who can access and modify resources across your organization and individual projects.

## Organization Roles

Every user in an organization is assigned one of three roles:

| Role | Description |
|------|-------------|
| **Owner** | Full control over the organization. Can manage billing, delete the org, and transfer ownership. Only one owner per organization. |
| **Admin** | Can manage members, projects, SSO, and organization settings. Cannot delete the organization or manage billing. |
| **Member** | Can access assigned projects and use the platform features. Cannot manage organization settings. |

### Permission Matrix

| Action | Owner | Admin | Member |
|--------|:-----:|:-----:|:------:|
| View organization settings | ✅ | ✅ | ❌ |
| Manage billing | ✅ | ❌ | ❌ |
| Invite/remove members | ✅ | ✅ | ❌ |
| Change member roles | ✅ | ✅ | ❌ |
| Create projects | ✅ | ✅ | ❌ |
| Delete projects | ✅ | ✅ | ❌ |
| Configure SSO | ✅ | ✅ | ❌ |
| View audit logs | ✅ | ✅ | ❌ |
| Manage API keys | ✅ | ✅ | Per-project |
| Access projects | ✅ | ✅ | Assigned only |

## Project-Level Permissions

Within each project, members can be assigned granular roles:

| Project Role | Description |
|-------------|-------------|
| **Project Admin** | Full control over the project — settings, members, API keys, prompts, datasets |
| **Editor** | Can create and modify traces, prompts, datasets, and evaluations |
| **Viewer** | Read-only access to all project data |

### Project Permission Matrix

| Action | Project Admin | Editor | Viewer |
|--------|:------------:|:------:|:------:|
| View traces | ✅ | ✅ | ✅ |
| View prompts | ✅ | ✅ | ✅ |
| View datasets | ✅ | ✅ | ✅ |
| Create/edit prompts | ✅ | ✅ | ❌ |
| Manage prompt labels | ✅ | ✅ | ❌ |
| Create/edit datasets | ✅ | ✅ | ❌ |
| Run evaluations | ✅ | ✅ | ❌ |
| Manage project API keys | ✅ | ❌ | ❌ |
| Add/remove project members | ✅ | ❌ | ❌ |
| Delete project data | ✅ | ❌ | ❌ |
| Change project settings | ✅ | ❌ | ❌ |

## Managing Roles

### Assign Organization Role

#### Dashboard

1. Go to **Settings > Organization > Members**
2. Click the role dropdown next to a member
3. Select the new role

#### API

```bash
curl -X PATCH "https://api.agenttrace.io/v1/organizations/:orgId/members/:userId" \
  -H "Authorization: Bearer at-your-api-key" \
  -H "Content-Type: application/json" \
  -d '{"role": "admin"}'
```

### Assign Project Role

#### Dashboard

1. Navigate to your project
2. Go to **Project Settings > Members**
3. Click **Add Member** or update an existing member's role

#### API

```bash
curl -X POST "https://api.agenttrace.io/v1/projects/:projectId/members" \
  -H "Authorization: Bearer at-your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "userId": "user_xyz",
    "role": "editor"
  }'
```

## API Key Scoping

API keys inherit the permissions of their scope:

| Key Type | Scope | Use Case |
|----------|-------|----------|
| Organization key | All projects | CI/CD pipelines, admin scripts |
| Project key | Single project | SDK integration, application code |
| Read-only key | Single project (read) | Dashboards, monitoring |

```bash
# Create a project-scoped API key
curl -X POST "https://api.agenttrace.io/v1/projects/:projectId/api-keys" \
  -H "Authorization: Bearer at-your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "production-sdk",
    "scopes": ["traces:write", "prompts:read"]
  }'
```

## SSO Role Mapping

When using SSO, map identity provider groups to AgentTrace roles:

```json
{
  "groupMapping": {
    "agenttrace-admins": "admin",
    "agenttrace-editors": "member",
    "engineering": "member"
  },
  "defaultRole": "member"
}
```

Users are automatically assigned roles based on their IdP group membership.

## Best Practices

1. **Principle of least privilege** — assign the minimum role needed for each user's work
2. **Use project-level roles** — grant broad org access sparingly; use per-project roles instead
3. **Scope API keys narrowly** — use project-scoped keys with minimal permissions
4. **Audit role changes** — review the [audit log](./audit-logs.md) for role change events
5. **Map SSO groups** — automate role assignment through IdP group mapping

## Related

- [Audit Logs](./audit-logs.md) — track role changes and access events
- [SSO](./sso.md) — configure identity provider integration
- [Compliance](./compliance.md) — data access controls for regulatory compliance
