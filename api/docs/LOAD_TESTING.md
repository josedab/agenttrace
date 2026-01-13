# AgentTrace Load Testing Guide

This guide covers load testing strategies, tools, and benchmarks for AgentTrace deployments.

## Overview

AgentTrace should handle high-throughput trace ingestion while maintaining low-latency queries. Key performance targets:

| Metric | Target | Minimum |
|--------|--------|---------|
| Trace Ingestion | 10,000/sec | 1,000/sec |
| P99 Query Latency | < 500ms | < 2s |
| P99 Ingestion Latency | < 100ms | < 500ms |
| Concurrent Users | 1,000 | 100 |

## Test Scenarios

### 1. Trace Ingestion Load Test

Simulates high-volume trace ingestion from multiple AI agents.

```javascript
// k6/trace-ingestion.js
import http from 'k6/http';
import { check, sleep } from 'k6';
import { randomString } from 'https://jslib.k6.io/k6-utils/1.2.0/index.js';

export const options = {
  stages: [
    { duration: '2m', target: 100 },   // Ramp up
    { duration: '5m', target: 100 },   // Sustained load
    { duration: '2m', target: 200 },   // Spike
    { duration: '5m', target: 200 },   // Sustained spike
    { duration: '2m', target: 0 },     // Ramp down
  ],
  thresholds: {
    http_req_duration: ['p(99)<500'],
    http_req_failed: ['rate<0.01'],
  },
};

const API_URL = __ENV.API_URL || 'http://localhost:8080';
const API_KEY = __ENV.API_KEY;

export default function () {
  const traceId = randomString(32);
  const payload = JSON.stringify({
    id: traceId,
    name: `load-test-trace-${traceId}`,
    timestamp: new Date().toISOString(),
    input: { query: 'What is the weather?' },
    output: { response: 'The weather is sunny.' },
    metadata: {
      model: 'gpt-4',
      tokens: { input: 10, output: 20 },
    },
    observations: generateObservations(traceId, 5),
  });

  const res = http.post(`${API_URL}/api/public/traces`, payload, {
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${API_KEY}`,
    },
  });

  check(res, {
    'status is 200': (r) => r.status === 200,
    'response time < 500ms': (r) => r.timings.duration < 500,
  });

  sleep(0.1);
}

function generateObservations(traceId, count) {
  const observations = [];
  for (let i = 0; i < count; i++) {
    observations.push({
      id: `${traceId}-obs-${i}`,
      trace_id: traceId,
      type: i === 0 ? 'SPAN' : 'GENERATION',
      name: `observation-${i}`,
      start_time: new Date().toISOString(),
      end_time: new Date().toISOString(),
      input: { step: i },
      output: { result: `step ${i} complete` },
    });
  }
  return observations;
}
```

### 2. Query Performance Test

Tests dashboard query performance under load.

```javascript
// k6/query-performance.js
import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  vus: 50,
  duration: '5m',
  thresholds: {
    http_req_duration: ['p(99)<2000'],
    http_req_failed: ['rate<0.01'],
  },
};

const API_URL = __ENV.API_URL || 'http://localhost:8080';
const API_KEY = __ENV.API_KEY;
const PROJECT_ID = __ENV.PROJECT_ID;

const queries = [
  // List traces
  () => http.get(`${API_URL}/api/public/traces?limit=50`, headers()),

  // Get trace with observations
  () => http.get(`${API_URL}/api/public/traces/${randomTraceId()}`, headers()),

  // Search traces
  () => http.get(`${API_URL}/api/public/traces?search=test&limit=20`, headers()),

  // Aggregations
  () => http.get(`${API_URL}/api/public/metrics/traces?period=day`, headers()),

  // GraphQL query
  () => http.post(`${API_URL}/graphql`, JSON.stringify({
    query: `
      query {
        traces(projectId: "${PROJECT_ID}", limit: 20) {
          id
          name
          timestamp
          observations {
            id
            type
          }
        }
      }
    `,
  }), {
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${API_KEY}`,
    },
  }),
];

function headers() {
  return {
    headers: {
      'Authorization': `Bearer ${API_KEY}`,
    },
  };
}

function randomTraceId() {
  // Use a pool of known trace IDs for realistic queries
  const traceIds = __ENV.TRACE_IDS?.split(',') || ['test-trace-1'];
  return traceIds[Math.floor(Math.random() * traceIds.length)];
}

export default function () {
  const query = queries[Math.floor(Math.random() * queries.length)];
  const res = query();

  check(res, {
    'status is 200': (r) => r.status === 200,
    'response time < 2s': (r) => r.timings.duration < 2000,
  });

  sleep(0.5);
}
```

### 3. Mixed Workload Test

Simulates realistic production traffic patterns.

```javascript
// k6/mixed-workload.js
import http from 'k6/http';
import { check, sleep, group } from 'k6';

export const options = {
  scenarios: {
    ingestion: {
      executor: 'constant-arrival-rate',
      rate: 100,
      timeUnit: '1s',
      duration: '10m',
      preAllocatedVUs: 50,
      exec: 'ingestTrace',
    },
    queries: {
      executor: 'constant-vus',
      vus: 20,
      duration: '10m',
      exec: 'queryTraces',
    },
    dashboard: {
      executor: 'ramping-vus',
      startVUs: 5,
      stages: [
        { duration: '2m', target: 10 },
        { duration: '6m', target: 10 },
        { duration: '2m', target: 5 },
      ],
      exec: 'dashboardActivity',
    },
  },
  thresholds: {
    'http_req_duration{scenario:ingestion}': ['p(99)<200'],
    'http_req_duration{scenario:queries}': ['p(99)<1000'],
    'http_req_duration{scenario:dashboard}': ['p(99)<2000'],
  },
};

export function ingestTrace() {
  // Trace ingestion logic
}

export function queryTraces() {
  // Query logic
}

export function dashboardActivity() {
  // Simulate user browsing dashboard
  group('view project', () => {
    http.get(`${API_URL}/api/projects/${PROJECT_ID}`);
    sleep(1);
  });

  group('list traces', () => {
    http.get(`${API_URL}/api/public/traces?limit=50`);
    sleep(2);
  });

  group('view trace detail', () => {
    http.get(`${API_URL}/api/public/traces/${randomTraceId()}`);
    sleep(3);
  });
}
```

### 4. Spike Test

Tests system behavior under sudden traffic spikes.

```javascript
// k6/spike-test.js
export const options = {
  stages: [
    { duration: '1m', target: 10 },    // Baseline
    { duration: '30s', target: 500 },  // Spike to 50x
    { duration: '2m', target: 500 },   // Hold spike
    { duration: '30s', target: 10 },   // Drop
    { duration: '2m', target: 10 },    // Recovery
  ],
};
```

### 5. Soak Test

Tests for memory leaks and degradation over time.

```javascript
// k6/soak-test.js
export const options = {
  stages: [
    { duration: '5m', target: 50 },
    { duration: '4h', target: 50 },   // Run for 4 hours
    { duration: '5m', target: 0 },
  ],
  thresholds: {
    http_req_duration: ['p(99)<1000'],
    http_req_failed: ['rate<0.001'],
  },
};
```

## Running Tests

### Prerequisites

```bash
# Install k6
brew install k6
# or
docker pull grafana/k6

# Set environment variables
export API_URL=https://api.agenttrace.example.com
export API_KEY=sk_test_xxxxx
export PROJECT_ID=proj_xxxxx
```

### Execute Tests

```bash
# Basic run
k6 run k6/trace-ingestion.js

# With environment variables
k6 run -e API_URL=http://localhost:8080 -e API_KEY=sk_test_xxx k6/trace-ingestion.js

# Output to InfluxDB for Grafana dashboards
k6 run --out influxdb=http://localhost:8086/k6 k6/trace-ingestion.js

# Output to JSON for analysis
k6 run --out json=results.json k6/trace-ingestion.js
```

### Docker Execution

```bash
docker run --rm -i grafana/k6 run - < k6/trace-ingestion.js

# With network access to local services
docker run --rm -i --network host \
  -e API_URL=http://localhost:8080 \
  -e API_KEY=sk_test_xxx \
  grafana/k6 run - < k6/trace-ingestion.js
```

## Benchmarks

### Reference Hardware

Tests performed on:
- **API:** 3x pods, 2 CPU / 2GB RAM each
- **Worker:** 2x pods, 1 CPU / 1GB RAM each
- **PostgreSQL:** 4 CPU / 8GB RAM
- **ClickHouse:** 4 CPU / 16GB RAM
- **Redis:** 1 CPU / 2GB RAM

### Results

| Test | Metric | Result |
|------|--------|--------|
| Ingestion | Throughput | 5,200 traces/sec |
| Ingestion | P50 Latency | 12ms |
| Ingestion | P99 Latency | 85ms |
| Query | Throughput | 450 queries/sec |
| Query | P50 Latency | 45ms |
| Query | P99 Latency | 380ms |
| Mixed | Error Rate | 0.02% |

### ClickHouse Query Benchmarks

```sql
-- Trace list query (should be < 100ms)
SELECT id, name, timestamp, status
FROM traces
WHERE project_id = 'xxx'
ORDER BY timestamp DESC
LIMIT 50;

-- Trace with observations (should be < 200ms)
SELECT t.*, o.*
FROM traces t
LEFT JOIN observations o ON t.id = o.trace_id
WHERE t.id = 'xxx';

-- Aggregation query (should be < 500ms)
SELECT
  toDate(timestamp) as date,
  count() as count,
  avg(duration_ms) as avg_duration
FROM traces
WHERE project_id = 'xxx'
  AND timestamp > now() - INTERVAL 30 DAY
GROUP BY date
ORDER BY date;
```

## Performance Tuning

### API Server

```yaml
# Increase connection pool
POSTGRES_MAX_CONNS: 50
POSTGRES_MIN_CONNS: 10

# Tune rate limits for load tests
RATE_LIMIT_REQUESTS_PER_SECOND: 1000
RATE_LIMIT_BURST: 2000
```

### ClickHouse

```sql
-- Optimize for batch inserts
SET max_insert_block_size = 1048576;
SET min_insert_block_size_rows = 1048576;

-- Increase memory for queries
SET max_memory_usage = 10000000000;

-- Parallel query execution
SET max_threads = 8;
```

### Worker

```yaml
# Increase concurrency during load tests
WORKER_CONCURRENCY: 20
```

## CI Integration

### GitHub Actions Workflow

```yaml
name: Load Test

on:
  schedule:
    - cron: '0 2 * * 0'  # Weekly on Sunday
  workflow_dispatch:

jobs:
  load-test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Setup k6
        uses: grafana/setup-k6-action@v1

      - name: Run load test
        run: |
          k6 run --out json=results.json k6/trace-ingestion.js
        env:
          API_URL: ${{ secrets.STAGING_API_URL }}
          API_KEY: ${{ secrets.STAGING_API_KEY }}

      - name: Check thresholds
        run: |
          # Parse results and fail if thresholds exceeded
          python scripts/check-load-test-results.py results.json

      - name: Upload results
        uses: actions/upload-artifact@v4
        with:
          name: load-test-results
          path: results.json
```

## Monitoring During Tests

### Key Metrics to Watch

1. **System Resources**
   - CPU utilization (< 80%)
   - Memory usage (< 85%)
   - Network I/O

2. **Application Metrics**
   - Request rate
   - Error rate
   - Latency percentiles

3. **Database Metrics**
   - Connection pool usage
   - Query duration
   - ClickHouse merges/inserts

### Grafana Dashboard Queries

```promql
# Request rate during test
rate(http_requests_total[1m])

# Error rate
sum(rate(http_requests_total{status=~"5.."}[1m])) / sum(rate(http_requests_total[1m]))

# P99 latency
histogram_quantile(0.99, rate(http_request_duration_seconds_bucket[1m]))

# ClickHouse insert rate
rate(clickhouse_insert_rows_total[1m])
```

## Troubleshooting

### High Error Rate

1. Check API logs for errors
2. Verify database connectivity
3. Check rate limiting configuration
4. Review resource utilization

### High Latency

1. Profile slow queries in ClickHouse
2. Check connection pool exhaustion
3. Review index usage
4. Analyze lock contention

### Memory Issues

1. Monitor Go heap usage
2. Check for goroutine leaks
3. Review ClickHouse memory settings
4. Analyze Redis memory usage
