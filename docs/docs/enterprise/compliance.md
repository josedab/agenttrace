---
sidebar_position: 4
title: "Compliance"
description: "Compliance features including data retention policies, GDPR data export/deletion, and SOC 2 readiness."
---

# Compliance

AgentTrace Enterprise includes compliance features to help your organization meet regulatory requirements including GDPR, SOC 2, HIPAA, and internal data governance policies.

## Data Retention Policies

Configure automatic data retention to manage storage costs and comply with data minimization requirements.

### Setting Retention Policies

#### Dashboard

1. Navigate to **Settings > Organization > Data Retention**
2. Set retention periods for each data type
3. Click **Save**

#### API

```bash
curl -X PUT "https://api.agenttrace.io/v1/organizations/:orgId/retention-policy" \
  -H "Authorization: Bearer at-your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "traces": {"retentionDays": 90},
    "observations": {"retentionDays": 90},
    "scores": {"retentionDays": 180},
    "auditLogs": {"retentionDays": 365},
    "prompts": {"retentionDays": null}
  }'
```

Setting `retentionDays` to `null` means data is retained indefinitely.

### Retention Defaults

| Data Type | Default | Minimum | Notes |
|-----------|---------|---------|-------|
| Traces | 90 days | 7 days | Includes all observation data |
| Scores | 180 days | 7 days | Evaluation scores and annotations |
| Audit logs | 365 days | 90 days | Compliance requirement |
| Prompts | Indefinite | — | Versions are never auto-deleted |
| User data | Indefinite | — | Deleted on account removal |

### How Retention Works

Data past its retention period is soft-deleted first (marked for deletion), then permanently removed within 24 hours by a background job. This provides a safety window for recovery.

## GDPR Compliance

AgentTrace helps you comply with GDPR requirements for personal data handling.

### Data Subject Access Request (DSAR)

Export all data associated with a user:

```bash
curl -X POST "https://api.agenttrace.io/v1/organizations/:orgId/gdpr/export" \
  -H "Authorization: Bearer at-your-api-key" \
  -H "Content-Type: application/json" \
  -d '{"email": "user@example.com"}'
```

Response includes a download URL for a ZIP archive containing:

- User profile data
- All traces initiated by the user
- Prompt versions created by the user
- Audit log entries for the user
- API keys created by the user (metadata only, not secrets)

### Right to Erasure (Data Deletion)

Delete all personal data for a user:

```bash
curl -X POST "https://api.agenttrace.io/v1/organizations/:orgId/gdpr/delete" \
  -H "Authorization: Bearer at-your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "confirmDeletion": true
  }'
```

This action:

- Deletes the user account and profile
- Anonymizes traces (replaces user identifiers with a hash)
- Removes the user from audit log actor fields
- Revokes all API keys
- Is **irreversible**

### Data Processing Agreements

AgentTrace provides a DPA (Data Processing Agreement) for cloud customers. Contact [enterprise@agenttrace.io](mailto:enterprise@agenttrace.io) to obtain a signed DPA.

## SOC 2 Readiness

AgentTrace is designed with SOC 2 controls in mind across Trust Service Criteria.

### Security Controls

| Control | Implementation |
|---------|---------------|
| Encryption at rest | AES-256 for all stored data |
| Encryption in transit | TLS 1.2+ for all connections |
| Authentication | SSO, MFA support, API key rotation |
| Authorization | RBAC with per-project permissions |
| Audit logging | All significant actions logged |
| Vulnerability management | Regular dependency scanning, security patches |

### Availability Controls

| Control | Implementation |
|---------|---------------|
| Redundancy | Multi-replica deployments, database replication |
| Backups | Automated daily backups with point-in-time recovery |
| Monitoring | Health checks, Prometheus metrics, alerting |
| Disaster recovery | Documented backup/restore procedures |

### Confidentiality Controls

| Control | Implementation |
|---------|---------------|
| Data classification | Traces, prompts, and metadata are classified by sensitivity |
| Access controls | RBAC, API key scoping, network policies |
| Data retention | Configurable policies with automatic enforcement |
| Secure deletion | Cryptographic erasure of deleted data |

## Self-Hosted Compliance

For organizations with strict data residency requirements, self-hosting AgentTrace ensures:

- **Data never leaves your infrastructure** — all data stays in your cloud account or on-premises
- **Network isolation** — deploy in air-gapped environments with no external connectivity
- **Custom encryption** — bring your own encryption keys
- **Audit everything** — full access to all logs and data stores

## Compliance Reports

Generate compliance reports from the dashboard:

1. Navigate to **Settings > Organization > Compliance**
2. Select report type:
   - **Data inventory** — all data types, locations, and retention periods
   - **Access review** — all users, roles, and last access times
   - **Audit summary** — aggregated audit log statistics
3. Click **Generate Report**

## Best Practices

1. **Set retention policies early** — configure before ingesting data at scale
2. **Regular access reviews** — review user roles quarterly
3. **Test GDPR workflows** — practice data export and deletion in staging
4. **Document your DPA** — maintain a record of data processing activities
5. **Monitor audit logs** — set up alerts for compliance-relevant events

## Related

- [Audit Logs](./audit-logs.md) — detailed audit logging
- [RBAC](./rbac.md) — access control configuration
- [SSO](./sso.md) — identity provider integration
- [Backup & Restore](../self-hosting/backup.md) — data protection procedures
