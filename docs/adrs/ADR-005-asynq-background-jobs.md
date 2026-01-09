# ADR-005: Asynq for Background Job Processing

## Status

Accepted

## Context

AgentTrace requires background job processing for several critical operations:

1. **Cost calculation**: Computing token costs after trace ingestion (async to not block ingestion)
2. **Evaluations**: Running LLM-as-judge evaluations on traces
3. **Exports**: Sending traces to external systems (OpenTelemetry, webhooks)
4. **Anomaly detection**: Running ML models to detect unusual patterns
5. **Cleanup**: Data retention enforcement and old trace deletion
6. **Notifications**: Sending alerts via email, Slack, PagerDuty

These operations must:
- Not block the HTTP request/response cycle
- Support retries with backoff
- Handle failures gracefully (dead letter queues)
- Scale horizontally with multiple workers

### Alternatives Considered

1. **Apache Kafka**
   - Pros: Industry standard, excellent throughput, event sourcing support
   - Cons: Complex operations, JVM dependency, overkill for our scale

2. **RabbitMQ**
   - Pros: Mature, rich routing, multiple protocols
   - Cons: Separate service to manage, Erlang runtime, complex clustering

3. **AWS SQS / Google Cloud Tasks**
   - Pros: Managed service, no operations
   - Cons: Vendor lock-in, not self-hostable, latency for self-hosted

4. **Go channels + goroutines**
   - Pros: No dependencies, simple
   - Cons: No persistence, lost jobs on restart, no distributed workers

5. **Asynq (Redis-backed)** (chosen)
   - Pros: Go-native, simple ops (just Redis), built-in retries, monitoring UI
   - Cons: Redis single point of failure, limited vs Kafka at extreme scale

## Decision

We use **Asynq** as our background job processing system, backed by Redis.

### Architecture

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│  API Server │────►│    Redis    │◄────│   Worker    │
│  (Producer) │     │   (Queue)   │     │  (Consumer) │
└─────────────┘     └─────────────┘     └─────────────┘
                           │
                    ┌──────┴──────┐
                    │  Asynqmon   │
                    │  (Web UI)   │
                    └─────────────┘
```

### Task Types

```go
const (
    TypeCostCalculation    = "cost:calculate"
    TypeEvaluation         = "eval:run"
    TypeOTelExport         = "otel:export"
    TypeWebhookDelivery    = "webhook:deliver"
    TypeAnomalyDetection   = "anomaly:detect"
    TypeDataCleanup        = "cleanup:traces"
    TypeNotification       = "notify:send"
)
```

### Worker Configuration

```go
// Separate worker process from API server
srv := asynq.NewServer(
    asynq.RedisClientOpt{Addr: redisAddr},
    asynq.Config{
        Concurrency: 10,
        Queues: map[string]int{
            "critical": 6,  // Cost calculation, notifications
            "default":  3,  // Evaluations, exports
            "low":      1,  // Cleanup, anomaly detection
        },
        RetryDelayFunc: asynq.DefaultRetryDelayFunc,
    },
)

mux := asynq.NewServeMux()
mux.HandleFunc(TypeCostCalculation, handleCostCalculation)
mux.HandleFunc(TypeEvaluation, handleEvaluation)
// ... register other handlers
```

### Task Creation

```go
// Enqueue a cost calculation task
task := asynq.NewTask(TypeCostCalculation, payload,
    asynq.MaxRetry(3),
    asynq.Timeout(5*time.Minute),
    asynq.Queue("critical"),
)
client.Enqueue(task)
```

## Consequences

### Positive

- **Go-native**: No FFI or external client libraries; Asynq is pure Go
- **Operational simplicity**: Redis is already in our stack for caching
- **Built-in features**: Retries, deadlines, unique tasks, periodic tasks, middleware
- **Monitoring**: Asynqmon provides web UI for queue inspection
- **Horizontal scaling**: Add more worker replicas to increase throughput
- **Graceful shutdown**: Workers finish in-progress tasks before terminating

### Negative

- **Redis dependency**: Redis becomes critical infrastructure (needs persistence, replication)
- **Scale ceiling**: Redis throughput limits apply (~100k ops/sec per instance)
- **No event sourcing**: Unlike Kafka, can't replay job history
- **Single consumer model**: Each job processed by one worker (no fan-out)

### Neutral

- Asynq has good but not extensive ecosystem compared to Kafka/RabbitMQ
- Monitoring requires deploying Asynqmon separately
- Redis Cluster supported but adds operational complexity

## Operational Considerations

### Redis Configuration

```yaml
# docker-compose.yml
redis:
  image: redis:7-alpine
  command: redis-server --appendonly yes  # Enable AOF persistence
  volumes:
    - redis_data:/data
```

### Worker Deployment

```yaml
# Separate container from API
worker:
  build: ./api
  command: ./worker
  environment:
    - REDIS_ADDR=redis:6379
  deploy:
    replicas: 3  # Scale workers independently
```

### Queue Priorities

| Queue | Priority | Use Cases |
|-------|----------|-----------|
| critical | 6 | Cost calculation, notifications, alerts |
| default | 3 | Evaluations, exports, webhooks |
| low | 1 | Cleanup, anomaly detection, batch jobs |

### Retry Strategy

```go
// Exponential backoff with jitter
asynq.RetryDelayFunc(func(n int, err error, task *asynq.Task) time.Duration {
    return time.Duration(math.Pow(2, float64(n))) * time.Second
})
```

## References

- [Asynq Documentation](https://github.com/hibiken/asynq)
- [Asynqmon Web UI](https://github.com/hibiken/asynqmon)
- [Redis Persistence](https://redis.io/docs/management/persistence/)
