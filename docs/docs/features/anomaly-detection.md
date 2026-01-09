# Anomaly Detection

AgentTrace provides intelligent anomaly detection to help you identify unusual behavior in your AI agents before they become problems. The system uses statistical methods to establish baselines and detect deviations.

## Overview

The anomaly detection system monitors key metrics from your traces:

- **Latency** - Response time anomalies
- **Cost** - Unexpected cost spikes
- **Error Rate** - Unusual error patterns
- **Token Usage** - Abnormal token consumption
- **Custom Metrics** - Your own defined metrics

## Detection Methods

AgentTrace supports multiple statistical detection algorithms:

### Z-Score Detection

Uses standard deviation to identify outliers. A Z-score above the threshold indicates an anomaly.

```json
{
  "method": "z_score",
  "config": {
    "zScoreThreshold": 3.0,  // 3 standard deviations
    "minSamples": 30,
    "lookbackHours": 24
  }
}
```

**Best for:** Normally distributed data with consistent patterns.

### IQR (Interquartile Range)

Uses quartiles to identify outliers. More robust to non-normal distributions.

```json
{
  "method": "iqr",
  "config": {
    "iqrMultiplier": 1.5,  // Standard IQR multiplier
    "minSamples": 30,
    "lookbackHours": 24
  }
}
```

**Best for:** Skewed data or data with occasional legitimate outliers.

### MAD (Median Absolute Deviation)

Robust statistical method using median instead of mean.

```json
{
  "method": "mad",
  "config": {
    "madThreshold": 3.0,
    "minSamples": 30,
    "lookbackHours": 24
  }
}
```

**Best for:** Data with extreme outliers that would skew mean-based methods.

### Moving Average

Compares current value against a rolling average.

```json
{
  "method": "moving_average",
  "config": {
    "windowSize": 10,      // Number of data points
    "deviation": 0.2,      // 20% deviation threshold
    "minSamples": 10
  }
}
```

**Best for:** Detecting sudden changes from recent behavior.

### Exponential Moving Average (EMA)

Weighted moving average that gives more importance to recent values.

```json
{
  "method": "exponential_ema",
  "config": {
    "alpha": 0.3,          // Smoothing factor (0-1)
    "deviation": 0.2,      // 20% deviation threshold
    "minSamples": 10
  }
}
```

**Best for:** Time series with trends where recent data is more relevant.

### Static Thresholds

Simple upper/lower bound checking.

```json
{
  "method": "threshold",
  "config": {
    "minThreshold": 0,
    "maxThreshold": 5000   // e.g., max 5 seconds latency
  }
}
```

**Best for:** When you have known acceptable ranges.

## Creating Detection Rules

### Via API

```bash
curl -X POST https://your-agenttrace.com/api/public/anomaly/rules \
  -H "Authorization: Bearer $AGENTTRACE_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "High Latency Detection",
    "type": "latency",
    "method": "z_score",
    "config": {
      "zScoreThreshold": 2.5,
      "minSamples": 50,
      "lookbackHours": 48
    },
    "severity": "high",
    "alertChannels": ["webhook-id-1", "webhook-id-2"],
    "cooldownMinutes": 30,
    "traceNameFilter": "production-*"
  }'
```

### Rule Configuration Options

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | Human-readable rule name |
| `type` | enum | `latency`, `cost`, `error_rate`, `tokens`, `custom` |
| `method` | enum | Detection algorithm to use |
| `config` | object | Method-specific configuration |
| `severity` | enum | `low`, `medium`, `high`, `critical` |
| `alertChannels` | array | Webhook IDs for notifications |
| `cooldownMinutes` | int | Minutes between repeated alerts |
| `traceNameFilter` | string | Glob pattern for trace filtering |
| `metadataFilters` | object | Key-value filters on trace metadata |
| `enabled` | bool | Whether rule is active |

## Alerts and Notifications

When an anomaly is detected, AgentTrace can send alerts to configured webhooks (Slack, Discord, PagerDuty, etc.).

### Alert Lifecycle

1. **Active** - Anomaly detected, alert triggered
2. **Acknowledged** - Team member has seen the alert
3. **Resolved** - Issue has been addressed
4. **Suppressed** - Alert temporarily silenced

### Acknowledging Alerts

```bash
curl -X POST https://your-agenttrace.com/api/public/anomaly/alerts/{id}/acknowledge \
  -H "Authorization: Bearer $AGENTTRACE_API_KEY"
```

### Resolving Alerts

```bash
curl -X POST https://your-agenttrace.com/api/public/anomaly/alerts/{id}/resolve \
  -H "Authorization: Bearer $AGENTTRACE_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "note": "Root cause: API rate limiting. Implemented retry logic."
  }'
```

### Adding Notes

```bash
curl -X POST https://your-agenttrace.com/api/public/anomaly/alerts/{id}/notes \
  -H "Authorization: Bearer $AGENTTRACE_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "content": "Investigating increased latency in embedding model."
  }'
```

## Testing Rules

Before deploying a rule, test it against sample data:

```bash
curl -X POST https://your-agenttrace.com/api/public/anomaly/rules/test \
  -H "Authorization: Bearer $AGENTTRACE_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "type": "latency",
    "method": "z_score",
    "config": {
      "zScoreThreshold": 3.0,
      "minSamples": 30
    },
    "testValue": 5000,
    "historicalData": [1200, 1300, 1100, 1250, 1400, 1150, 1350, ...]
  }'
```

Response:

```json
{
  "isAnomaly": true,
  "score": 4.2,
  "threshold": 3.0,
  "expected": 1250,
  "testValue": 5000,
  "description": "Value 5000.00 is 4.2 standard deviations above mean (1250.00)",
  "severity": "high",
  "baselineStats": {
    "mean": 1250,
    "stdDev": 100,
    "median": 1250,
    "p95": 1400,
    "p99": 1450,
    "min": 1100,
    "max": 1400
  }
}
```

## Viewing Anomaly Statistics

Get an overview of anomalies in your project:

```bash
curl https://your-agenttrace.com/api/public/anomaly/stats \
  -H "Authorization: Bearer $AGENTTRACE_API_KEY" \
  -G -d "startTime=2024-01-01T00:00:00Z" -d "endTime=2024-01-08T00:00:00Z"
```

Response:

```json
{
  "projectId": "...",
  "period": {
    "start": "2024-01-01T00:00:00Z",
    "end": "2024-01-08T00:00:00Z"
  },
  "totalAnomalies": 47,
  "activeAlerts": 3,
  "bySeverity": {
    "low": 20,
    "medium": 15,
    "high": 10,
    "critical": 2
  },
  "byType": {
    "latency": 25,
    "cost": 12,
    "error_rate": 10
  },
  "topAffectedTraces": [
    {"traceName": "document-processing", "count": 15},
    {"traceName": "chat-completion", "count": 12}
  ]
}
```

## Best Practices

### Choosing Detection Methods

| Use Case | Recommended Method |
|----------|-------------------|
| General purpose monitoring | Z-Score |
| Cost monitoring (often skewed) | IQR or MAD |
| Real-time alerting | Moving Average |
| Trending metrics | EMA |
| SLA enforcement | Static Threshold |

### Setting Thresholds

1. **Start conservative** - Use higher thresholds initially to reduce false positives
2. **Tune based on feedback** - Lower thresholds as you understand your data
3. **Consider severity** - Use lower thresholds for critical alerting vs informational

### Cooldown Configuration

- **High-volume traces**: Use longer cooldowns (30-60 min) to prevent alert fatigue
- **Critical systems**: Use shorter cooldowns (5-15 min) for faster response
- **Batch processing**: Align cooldown with batch intervals

### Filtering Strategies

Use trace name filters to create targeted rules:

```json
{
  "name": "Production Latency",
  "traceNameFilter": "prod-*",
  "metadataFilters": {
    "environment": "production",
    "tier": "critical"
  }
}
```

## Integration with Notifications

Anomaly alerts integrate with the [notification system](/integrations/notifications):

1. Create webhooks for your notification channels
2. Reference webhook IDs in your anomaly rules
3. Receive formatted alerts with anomaly details

Example Slack notification:

```
🚨 High Latency Anomaly Detected: Production API

Value 5000ms is 4.2 standard deviations above mean (1250ms)

Details:
- Severity: high
- Current Value: 5000ms
- Expected Value: 1250ms
- Deviation: +300%
- Trace: document-processing

Detected at: 2024-01-15T14:30:00Z
```

## API Reference

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/public/anomaly/rules` | GET | List detection rules |
| `/api/public/anomaly/rules` | POST | Create detection rule |
| `/api/public/anomaly/rules/{id}` | GET | Get rule details |
| `/api/public/anomaly/rules/{id}` | PATCH | Update rule |
| `/api/public/anomaly/rules/{id}` | DELETE | Delete rule |
| `/api/public/anomaly/rules/{id}/toggle` | POST | Enable/disable rule |
| `/api/public/anomaly/rules/test` | POST | Test rule configuration |
| `/api/public/anomaly/anomalies` | GET | List detected anomalies |
| `/api/public/anomaly/anomalies/{id}` | GET | Get anomaly details |
| `/api/public/anomaly/alerts` | GET | List alerts |
| `/api/public/anomaly/alerts/{id}` | GET | Get alert details |
| `/api/public/anomaly/alerts/{id}/acknowledge` | POST | Acknowledge alert |
| `/api/public/anomaly/alerts/{id}/resolve` | POST | Resolve alert |
| `/api/public/anomaly/alerts/{id}/notes` | POST | Add note to alert |
| `/api/public/anomaly/stats` | GET | Get anomaly statistics |
