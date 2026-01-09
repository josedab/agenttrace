# AgentTrace Go Agent Workflow Example

This example demonstrates building a multi-step agent workflow in Go, showcasing:

- **Parallel processing** - Process multiple documents concurrently
- **Nested spans** - Organize traces hierarchically per document
- **Multiple LLM calls** - Chain summarization, sentiment, and keyword extraction
- **Error handling** - Graceful handling of failed operations
- **Scoring** - Track workflow success rate as a score
- **Checkpointing** - Record workflow state for debugging

## Prerequisites

1. AgentTrace running locally or remotely:
   ```bash
   docker compose up -d
   ```

2. Environment variables:
   ```bash
   export AGENTTRACE_API_KEY="your-agenttrace-api-key"
   export AGENTTRACE_HOST="http://localhost:8080"
   ```

## Running the Example

```bash
go mod tidy
go run main.go
```

## What It Does

1. Creates a main workflow trace for processing 3 documents
2. **Phase 1**: Processes all documents in parallel, each with:
   - Summarization (LLM call)
   - Sentiment analysis (LLM call)
   - Keyword extraction (LLM call)
3. **Phase 2**: Aggregates results across all documents
4. **Phase 3**: Generates a summary report using an LLM
5. Creates a checkpoint with the final state
6. Adds a success rate score to the trace

## Trace Structure

```
document-processing-workflow (trace)
├── parallel-analysis (span)
│   ├── process-document-doc-1 (span)
│   │   ├── summarize-doc-1 (generation)
│   │   ├── sentiment-doc-1 (generation)
│   │   └── keywords-doc-1 (generation)
│   ├── process-document-doc-2 (span)
│   │   └── ... (parallel with doc-1)
│   └── process-document-doc-3 (span)
│       └── ... (parallel with doc-1)
├── aggregate-results (span)
├── generate-summary-report (generation)
└── checkpoint-created (event)
```

## Key Patterns Demonstrated

### Parallel Processing with Tracing

```go
var wg sync.WaitGroup
for i, doc := range documents {
    wg.Add(1)
    go func(idx int, d Document) {
        defer wg.Done()
        result, err := processDocument(ctx, trace, d)
        results[idx] = result
        errors[idx] = err
    }(i, doc)
}
wg.Wait()
```

### Nested Spans and Generations

```go
// Parent span for document processing
docSpan := trace.Span(ctx, agenttrace.SpanParams{
    Name: fmt.Sprintf("process-document-%s", doc.ID),
})

// Child generation linked to parent span
summaryGen := trace.Generation(ctx, agenttrace.GenerationParams{
    Name: fmt.Sprintf("summarize-%s", doc.ID),
    ParentObservationID: docSpan.ID(),  // Link to parent
})
```

### Error Handling

```go
summary, inTok, outTok, err := MockLLMCall(ctx, doc.Content)
if err != nil {
    summaryGen.End(agenttrace.GenerationEndParams{
        Output: map[string]interface{}{"error": err.Error()},
        Level:  "error",
    })
    docSpan.End(agenttrace.EndParams{
        Output: map[string]interface{}{"error": err.Error()},
        Level:  "error",
    })
    return result, fmt.Errorf("summarization failed: %w", err)
}
```

### Adding Scores

```go
trace.Score(ctx, agenttrace.ScoreParams{
    Name:    "workflow_success_rate",
    Value:   float64(successCount) / float64(len(documents)),
    Comment: fmt.Sprintf("%d of %d documents processed successfully",
        successCount, len(documents)),
})
```

### Events for Checkpointing

```go
trace.Event(ctx, agenttrace.EventParams{
    Name: "checkpoint-created",
    Input: map[string]interface{}{
        "checkpointData": string(finalStateJSON),
        "timestamp":      time.Now().Format(time.RFC3339),
    },
})
```

## Viewing Traces

1. Open http://localhost:3000 in your browser
2. Navigate to **Traces**
3. Find the trace named "document-processing-workflow"
4. Expand to see the parallel document processing
5. Check the **Scores** tab for the success rate

## Customization Ideas

- Increase document count to test parallelism limits
- Add retry logic for failed LLM calls
- Implement real LLM integration (OpenAI, Anthropic)
- Add file-based input for processing real documents
- Implement rate limiting for API calls
