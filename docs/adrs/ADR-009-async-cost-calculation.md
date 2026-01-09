# ADR-009: Asynchronous Cost Calculation

## Status

Accepted

## Context

AgentTrace tracks costs for LLM API calls across 400+ model/provider combinations. Cost calculation involves:

1. **Token counting**: Input and output tokens for each generation
2. **Model identification**: Mapping model names to pricing tiers
3. **Price lookup**: Current pricing per 1K/1M tokens (input vs output)
4. **Currency conversion**: Some providers price in different currencies
5. **Aggregation**: Roll up costs to trace, session, and project levels

Challenges:
- Prices change frequently (OpenAI updates monthly)
- Model names have variants (`gpt-4`, `gpt-4-0613`, `gpt-4-turbo-preview`)
- Context-dependent pricing (different prices for >8K context)
- Ingestion latency is critical (agents waiting for trace confirmation)

### Alternatives Considered

1. **Synchronous calculation during ingestion**
   - Pros: Immediate accurate costs, simpler architecture
   - Cons: Blocks ingestion, price lookup adds latency

2. **Client-side calculation**
   - Pros: No server overhead
   - Cons: Inconsistent across SDKs, clients may have stale prices

3. **Estimated costs, no updates**
   - Pros: Fast, simple
   - Cons: Inaccurate over time as prices change

4. **Asynchronous calculation** (chosen)
   - Pros: Fast ingestion, accurate costs, retroactive updates
   - Cons: Eventually consistent costs, more complex architecture

## Decision

We calculate costs **asynchronously after ingestion** via background workers:

### Architecture

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│ Trace       │     │ ClickHouse  │     │  Redis      │     │   Worker    │
│ Ingestion   │────►│ (raw data)  │────►│  (queue)    │────►│  (cost)     │
│ API         │     │ cost = 0    │     │             │     │             │
└─────────────┘     └─────────────┘     └─────────────┘     └─────────────┘
                                                                   │
                                                                   ▼
                                                          ┌─────────────┐
                                                          │ ClickHouse  │
                                                          │ (updated)   │
                                                          │ cost = 0.05 │
                                                          └─────────────┘
```

### Ingestion Flow

```go
// trace_handler.go
func (h *TraceHandler) IngestObservation(c *fiber.Ctx) error {
    // 1. Parse and validate
    var input domain.ObservationInput
    c.BodyParser(&input)

    // 2. Create observation with zero cost
    observation := &domain.Observation{
        ID:              uuid.New(),
        Type:            input.Type,
        Model:           input.Model,
        InputTokens:     input.Usage.InputTokens,
        OutputTokens:    input.Usage.OutputTokens,
        Cost:            0,  // Placeholder
        CostCalculated:  false,
    }

    // 3. Insert to ClickHouse (fast path)
    h.observationRepo.Create(c.Context(), observation)

    // 4. Queue cost calculation (async)
    h.costQueue.Enqueue(CostCalculationTask{
        ObservationID: observation.ID,
    })

    return c.Status(201).JSON(observation)
}
```

### Cost Calculation Worker

```go
// cost_worker.go
func (w *CostWorker) HandleCostCalculation(ctx context.Context, task *asynq.Task) error {
    var payload CostCalculationTask
    json.Unmarshal(task.Payload(), &payload)

    // 1. Fetch observation
    obs, err := w.observationRepo.GetByID(ctx, payload.ObservationID)
    if err != nil {
        return err
    }

    // 2. Look up pricing
    price, err := w.priceService.GetPrice(obs.Model, obs.Provider)
    if err != nil {
        // Unknown model - log and skip
        w.logger.Warn("unknown model", zap.String("model", obs.Model))
        return nil
    }

    // 3. Calculate cost
    cost := w.calculateCost(obs, price)

    // 4. Update observation
    return w.observationRepo.UpdateCost(ctx, obs.ID, cost)
}

func (w *CostWorker) calculateCost(obs *domain.Observation, price *ModelPrice) float64 {
    inputCost := float64(obs.InputTokens) / 1_000_000 * price.InputPer1M
    outputCost := float64(obs.OutputTokens) / 1_000_000 * price.OutputPer1M
    return inputCost + outputCost
}
```

### Pricing Service

```go
// price_service.go
type ModelPrice struct {
    Model           string
    Provider        string
    InputPer1M      float64  // USD per 1M input tokens
    OutputPer1M     float64  // USD per 1M output tokens
    ContextWindow   int      // For context-dependent pricing
    EffectiveDate   time.Time
}

func (s *PriceService) GetPrice(model, provider string) (*ModelPrice, error) {
    // 1. Check cache
    if price, ok := s.cache.Get(model); ok {
        return price, nil
    }

    // 2. Normalize model name (gpt-4-0613 -> gpt-4)
    normalizedModel := s.normalizeModelName(model)

    // 3. Look up in database
    price, err := s.priceRepo.GetByModel(normalizedModel, provider)
    if err != nil {
        return nil, err
    }

    // 4. Cache for future lookups
    s.cache.Set(model, price, 1*time.Hour)

    return price, nil
}
```

## Consequences

### Positive

- **Fast ingestion**: No blocking on price lookups during trace ingestion
- **Retroactive updates**: Can recalculate costs when prices change
- **Flexible pricing**: Supports complex pricing (context windows, tiers)
- **Resilience**: Price lookup failures don't break ingestion
- **Batch efficiency**: Worker can process costs in batches

### Negative

- **Eventually consistent**: Dashboard shows $0 briefly after ingestion
- **Complexity**: Separate worker, queue, and update path
- **State management**: Must track which observations need cost updates
- **Race conditions**: UI may read stale data during calculation

### Neutral

- Materialized views update after cost calculation completes
- Unknown models tracked for manual pricing additions
- Cost recalculation can be triggered manually

## Pricing Database Schema

```sql
CREATE TABLE model_prices (
    id UUID PRIMARY KEY,
    provider VARCHAR(50) NOT NULL,      -- openai, anthropic, google, etc.
    model_pattern VARCHAR(100) NOT NULL, -- gpt-4*, claude-3-opus*, etc.
    input_per_1m DECIMAL(10, 6) NOT NULL,
    output_per_1m DECIMAL(10, 6) NOT NULL,
    context_threshold INT,               -- For context-dependent pricing
    input_per_1m_above_threshold DECIMAL(10, 6),
    effective_from TIMESTAMP NOT NULL,
    effective_to TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Example data
INSERT INTO model_prices (provider, model_pattern, input_per_1m, output_per_1m, effective_from)
VALUES
    ('openai', 'gpt-4o', 2.50, 10.00, '2024-01-01'),
    ('openai', 'gpt-4-turbo', 10.00, 30.00, '2024-01-01'),
    ('anthropic', 'claude-3-opus', 15.00, 75.00, '2024-01-01'),
    ('anthropic', 'claude-3-sonnet', 3.00, 15.00, '2024-01-01');
```

## Cost Aggregation

```sql
-- Materialized view for trace-level costs
CREATE MATERIALIZED VIEW trace_costs_mv TO trace_costs AS
SELECT
    trace_id,
    project_id,
    sum(cost) as total_cost,
    sum(input_tokens) as total_input_tokens,
    sum(output_tokens) as total_output_tokens,
    max(updated_at) as last_calculated_at
FROM observations
WHERE cost_calculated = true
GROUP BY trace_id, project_id;
```

## Dashboard Handling

```typescript
// Show pending state when cost not yet calculated
function TraceCost({ trace }: { trace: Trace }) {
  if (!trace.costCalculated) {
    return <Skeleton className="w-16 h-4" />;
  }
  return <span>${trace.totalCost.toFixed(4)}</span>;
}
```

## References

- [OpenAI Pricing](https://openai.com/pricing)
- [Anthropic Pricing](https://www.anthropic.com/pricing)
- [LiteLLM Pricing Database](https://github.com/BerriAI/litellm/blob/main/model_prices_and_context_window.json)
