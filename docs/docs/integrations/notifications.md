# Notifications & Webhooks

AgentTrace supports sending notifications to various platforms when important events occur, such as trace errors, cost threshold breaches, or daily summaries.

## Supported Platforms

### Slack

Send notifications to Slack channels using incoming webhooks.

1. **Create a Slack App**
   - Go to [Slack API](https://api.slack.com/apps)
   - Click "Create New App" → "From scratch"
   - Name your app and select your workspace

2. **Enable Incoming Webhooks**
   - Go to "Incoming Webhooks" in the sidebar
   - Toggle "Activate Incoming Webhooks" to On
   - Click "Add New Webhook to Workspace"
   - Select the channel for notifications

3. **Configure in AgentTrace**
   ```bash
   curl -X POST https://api.agenttrace.io/api/public/webhooks \
     -H "Authorization: Bearer $AGENTTRACE_API_KEY" \
     -H "Content-Type: application/json" \
     -d '{
       "type": "slack",
       "name": "Production Alerts",
       "url": "https://hooks.slack.com/services/T.../B.../...",
       "events": ["trace.error", "trace.cost_threshold"],
       "isEnabled": true,
       "costThreshold": 0.50
     }'
   ```

### Discord

Send notifications to Discord channels using webhooks.

1. **Create a Discord Webhook**
   - Open your Discord server settings
   - Go to Integrations → Webhooks
   - Click "New Webhook"
   - Copy the webhook URL

2. **Configure in AgentTrace**
   ```bash
   curl -X POST https://api.agenttrace.io/api/public/webhooks \
     -H "Authorization: Bearer $AGENTTRACE_API_KEY" \
     -H "Content-Type: application/json" \
     -d '{
       "type": "discord",
       "name": "Dev Notifications",
       "url": "https://discord.com/api/webhooks/...",
       "events": ["trace.error", "daily.cost_report"],
       "isEnabled": true
     }'
   ```

### Microsoft Teams

Send notifications to Microsoft Teams channels.

1. **Create an Incoming Webhook**
   - In your Teams channel, click "..." → "Connectors"
   - Find "Incoming Webhook" and click "Configure"
   - Name your webhook and copy the URL

2. **Configure in AgentTrace**
   ```bash
   curl -X POST https://api.agenttrace.io/api/public/webhooks \
     -H "Authorization: Bearer $AGENTTRACE_API_KEY" \
     -H "Content-Type: application/json" \
     -d '{
       "type": "msteams",
       "name": "Team Alerts",
       "url": "https://outlook.office.com/webhook/...",
       "events": ["trace.error"],
       "isEnabled": true
     }'
   ```

### PagerDuty

Send alerts to PagerDuty for on-call escalation.

1. **Create a PagerDuty Integration**
   - Go to Services → Service Directory
   - Select your service or create a new one
   - Go to Integrations → Add Integration
   - Select "Events API v2"
   - Copy the Integration Key

2. **Configure in AgentTrace**
   ```bash
   curl -X POST https://api.agenttrace.io/api/public/webhooks \
     -H "Authorization: Bearer $AGENTTRACE_API_KEY" \
     -H "Content-Type: application/json" \
     -d '{
       "type": "pagerduty",
       "name": "Critical Alerts",
       "url": "https://events.pagerduty.com/v2/enqueue",
       "events": ["trace.error", "anomaly.detected"],
       "isEnabled": true,
       "headers": {
         "X-Routing-Key": "your-integration-key"
       }
     }'
   ```

### Generic Webhooks

Send JSON payloads to any HTTP endpoint.

```bash
curl -X POST https://api.agenttrace.io/api/public/webhooks \
  -H "Authorization: Bearer $AGENTTRACE_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "type": "generic",
    "name": "Custom Integration",
    "url": "https://your-api.example.com/webhooks/agenttrace",
    "secret": "your-webhook-secret",
    "events": ["trace.error", "trace.cost_threshold", "eval.failed"],
    "isEnabled": true,
    "headers": {
      "X-Custom-Header": "custom-value"
    }
  }'
```

## Event Types

| Event | Description |
|-------|-------------|
| `trace.error` | A trace completed with an error status |
| `trace.cost_threshold` | A trace's cost exceeded the configured threshold |
| `trace.latency_threshold` | A trace's latency exceeded the configured threshold |
| `daily.cost_report` | Daily summary of costs and usage |
| `eval.failed` | An evaluation failed to complete |
| `eval.score_low` | An evaluation score fell below the threshold |
| `anomaly.detected` | An anomaly was detected in trace patterns |
| `team.digest` | Project outcome digest rendered from real trace, git, CI, and cost data |

## Deliver a Team Digest

Create Slack, Discord, or generic webhooks first, then deliver one canonical report to selected project-owned webhook IDs:

```bash
curl -X POST "$AGENTTRACE_HOST/api/public/outcomes/digest/deliver" \
  -H "Authorization: Bearer $AGENTTRACE_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "window": "7d",
    "webhookIds": [
      "00000000-0000-0000-0000-000000000001"
    ]
  }'
```

### Validation happens before any send

Every requested webhook is resolved and checked before the first delivery: it must exist in the project, be enabled, use a supported type (`slack`, `discord`, `generic`), and have a usable public HTTPS destination. If any check fails, the request is rejected with `4xx` and **no** webhook receives the digest. Privacy no-egress mode is rejected the same way, with `422`, before any send.

### Partial results are reported honestly

Once delivery processing starts, a failing channel no longer aborts the request. Each non-duplicate webhook whose duplicate-safety lookup succeeds is attempted, every resulting outcome is persisted where storage is available, and the response summarizes the outcome:

```json
{
  "deliveries": [{ "webhookId": "...", "success": true, "deliveryKey": "9f1c..." }],
  "succeeded": 2,
  "failed": 1,
  "duplicates": 0,
  "deliveryKey": "9f1c..."
}
```

`failed` counts channels that rejected the delivery, could not complete the duplicate-safety lookup, or whose delivery record could not be persisted. A duplicate-safety lookup failure skips that send rather than risking a duplicate. `triggerUpdateFailures` counts deliveries that succeeded while their bookkeeping update failed; those are **not** reported as failures.

### Duplicate suppression

Each digest carries a `deliveryKey` derived from the project, the reporting window, and the rendered content. Before sending, AgentTrace atomically claims `(webhookId, deliveryKey)` in PostgreSQL across API instances. If the same key was already delivered within five minutes, the recorded delivery is returned; if another request is still sending it, the retry is reported as an in-progress duplicate. In both cases it is counted in `duplicates` and is not sent again. Claims expire after five minutes so a process crash does not block delivery indefinitely.

Duplicate webhook IDs in one request are ignored.

## Thresholds

Configure thresholds to trigger notifications:

```json
{
  "costThreshold": 1.00,      // USD per trace
  "latencyThreshold": 5000,   // milliseconds
  "scoreThreshold": 0.7       // 0-1 scale (triggers below this value)
}
```

## Rate Limiting

Prevent notification spam with rate limiting:

```json
{
  "rateLimitPerHour": 10  // Maximum notifications per hour
}
```

## Webhook Security

Webhook destinations must use public HTTPS without embedded credentials. AgentTrace resolves every destination before connecting, rejects private, loopback, link-local, and metadata-service IPs, pins the dial to validated public addresses, and does not follow redirects.

When `PRIVACY_NO_EGRESS=true`, webhook delivery is blocked at runtime.

### Signature Verification

AgentTrace signs webhook payloads using HMAC-SHA256. Verify signatures to ensure requests are authentic:

```python
import hmac
import hashlib

def verify_signature(payload: bytes, signature: str, secret: str) -> bool:
    expected = 'sha256=' + hmac.new(
        secret.encode(),
        payload,
        hashlib.sha256
    ).hexdigest()
    return hmac.compare_digest(signature, expected)

# In your webhook handler:
signature = request.headers.get('X-AgentTrace-Signature')
if not verify_signature(request.data, signature, WEBHOOK_SECRET):
    return 'Invalid signature', 401
```

### Request Headers

AgentTrace includes these headers with every webhook request:

| Header | Description |
|--------|-------------|
| `Content-Type` | `application/json` |
| `User-Agent` | `AgentTrace-Webhook/1.0` |
| `X-AgentTrace-Signature` | HMAC-SHA256 signature (if secret configured) |
| `X-AgentTrace-Event` | The event type that triggered this webhook |
| `X-AgentTrace-Delivery-ID` | Unique ID for this delivery attempt |

## Testing Webhooks

Test your webhook configuration:

```bash
curl -X POST https://api.agenttrace.io/api/public/webhooks/{id}/test \
  -H "Authorization: Bearer $AGENTTRACE_API_KEY"
```

## Webhook Delivery History

View delivery attempts and failures:

```bash
curl https://api.agenttrace.io/api/public/webhooks/{id}/deliveries \
  -H "Authorization: Bearer $AGENTTRACE_API_KEY"
```

Response:
```json
{
  "deliveries": [
    {
      "id": "del_123",
      "eventType": "trace.error",
      "statusCode": 200,
      "success": true,
      "duration": 245,
      "createdAt": "2024-01-15T10:30:00Z"
    },
    {
      "id": "del_124",
      "eventType": "trace.cost_threshold",
      "statusCode": 500,
      "success": false,
      "error": "Internal server error",
      "retryCount": 2,
      "createdAt": "2024-01-15T11:00:00Z"
    }
  ],
  "totalCount": 2,
  "hasMore": false
}
```

## Retry Policy

Failed webhook deliveries are automatically retried:

- **Maximum Retries**: 3
- **Retry Delays**: 1 minute, 5 minutes, 30 minutes
- **Retryable Errors**: 5xx status codes, timeouts, connection errors
- **Non-Retryable**: 4xx status codes (except 429)

## Payload Examples

### Slack Message (trace.error)

```json
{
  "attachments": [
    {
      "color": "#dc3545",
      "title": "Trace Error Alert",
      "title_link": "https://app.agenttrace.io/traces/tr_abc123",
      "text": "Trace 'process-document' failed with error:\n```API rate limit exceeded```",
      "fields": [
        {"title": "Trace ID", "value": "tr_abc123", "short": true},
        {"title": "Project", "value": "my-project", "short": true}
      ],
      "footer": "AgentTrace",
      "ts": 1705312200
    }
  ]
}
```

### Discord Message (daily.cost_report)

```json
{
  "username": "AgentTrace",
  "embeds": [
    {
      "title": "Daily Cost Report",
      "description": "Daily summary for 2024-01-15:\n- Total Cost: $45.23\n- Total Traces: 1,234",
      "color": 1564103,
      "timestamp": "2024-01-16T00:00:00Z",
      "footer": {"text": "AgentTrace"},
      "fields": [
        {"name": "Project", "value": "production", "inline": true}
      ]
    }
  ]
}
```

### Generic Webhook Payload

```json
{
  "id": "evt_abc123",
  "eventType": "trace.error",
  "timestamp": "2024-01-15T10:30:00Z",
  "projectId": "proj_xyz",
  "data": {
    "traceId": "tr_abc123",
    "traceName": "process-document",
    "error": "API rate limit exceeded",
    "latencyMs": 15234,
    "cost": 0.0045,
    "model": "gpt-4"
  }
}
```

## Best Practices

1. **Use specific events**: Subscribe only to events you care about
2. **Set appropriate thresholds**: Start with conservative thresholds and adjust
3. **Enable rate limiting**: Prevent alert fatigue during incidents
4. **Verify signatures**: Always verify webhook signatures in production
5. **Handle retries gracefully**: Make your webhook handlers idempotent
6. **Monitor delivery rates**: Review the delivery history periodically
