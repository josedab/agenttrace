/*
AgentTrace Go Agent Workflow Example

This example demonstrates building a multi-step agent workflow in Go,
including parallel operations, error handling, and checkpointing.

Prerequisites:

	go mod tidy
	export AGENTTRACE_API_KEY="your-api-key"

Run:

	go run main.go
*/
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"sync"
	"time"

	"github.com/agenttrace/agenttrace-go"
)

// Document represents a document to be processed
type Document struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

// ProcessingResult represents the result of document processing
type ProcessingResult struct {
	DocumentID string  `json:"documentId"`
	Summary    string  `json:"summary"`
	Sentiment  string  `json:"sentiment"`
	Keywords   []string `json:"keywords"`
	Score      float64 `json:"score"`
}

// MockLLMCall simulates an LLM API call
func MockLLMCall(ctx context.Context, prompt string) (string, int, int, error) {
	// Simulate latency
	time.Sleep(time.Duration(100+rand.Intn(200)) * time.Millisecond)

	// Simulate occasional failures
	if rand.Float32() < 0.1 {
		return "", 0, 0, fmt.Errorf("simulated LLM API error")
	}

	return fmt.Sprintf("Response to: %s...", prompt[:min(50, len(prompt))]),
		rand.Intn(100) + 50,
		rand.Intn(200) + 100,
		nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func main() {
	ctx := context.Background()
	rand.Seed(time.Now().UnixNano())

	// Initialize the AgentTrace client
	apiKey := os.Getenv("AGENTTRACE_API_KEY")
	host := os.Getenv("AGENTTRACE_HOST")
	if host == "" {
		host = "http://localhost:8080"
	}

	client, err := agenttrace.NewClient(
		agenttrace.WithAPIKey(apiKey),
		agenttrace.WithHost(host),
	)
	if err != nil {
		fmt.Printf("Failed to create client: %v\n", err)
		os.Exit(1)
	}
	defer client.Shutdown(ctx)

	// Sample documents to process
	documents := []Document{
		{ID: "doc-1", Title: "Introduction to Go", Content: "Go is a statically typed programming language designed at Google..."},
		{ID: "doc-2", Title: "AI in 2024", Content: "Artificial intelligence continues to advance rapidly with new developments..."},
		{ID: "doc-3", Title: "Cloud Computing", Content: "Cloud computing provides on-demand access to computing resources..."},
	}

	fmt.Printf("Processing %d documents...\n\n", len(documents))

	// Create the main workflow trace
	trace := client.Trace(ctx, agenttrace.TraceParams{
		Name: "document-processing-workflow",
		Metadata: map[string]interface{}{
			"language":       "go",
			"example":        "agent-workflow",
			"documentCount":  len(documents),
		},
		Input: map[string]interface{}{
			"documents": documents,
		},
		Tags: []string{"batch", "parallel", "production"},
	})

	// Phase 1: Parallel document analysis
	fmt.Println("Phase 1: Analyzing documents in parallel...")
	analysisSpan := trace.Span(ctx, agenttrace.SpanParams{
		Name: "parallel-analysis",
		Metadata: map[string]interface{}{
			"phase":        1,
			"parallelism":  len(documents),
		},
	})

	results := make([]ProcessingResult, len(documents))
	errors := make([]error, len(documents))
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

	// Check for errors
	successCount := 0
	failureCount := 0
	for i, err := range errors {
		if err != nil {
			fmt.Printf("  ✗ Document %s failed: %v\n", documents[i].ID, err)
			failureCount++
		} else {
			fmt.Printf("  ✓ Document %s processed successfully\n", documents[i].ID)
			successCount++
		}
	}

	analysisSpan.End(agenttrace.EndParams{
		Output: map[string]interface{}{
			"successCount": successCount,
			"failureCount": failureCount,
		},
	})

	// Phase 2: Aggregate results
	fmt.Println("\nPhase 2: Aggregating results...")
	aggregateSpan := trace.Span(ctx, agenttrace.SpanParams{
		Name: "aggregate-results",
		Metadata: map[string]interface{}{
			"phase": 2,
		},
		Input: map[string]interface{}{
			"resultCount": successCount,
		},
	})

	// Calculate aggregate statistics
	var totalScore float64
	allKeywords := make(map[string]int)
	sentimentCounts := make(map[string]int)

	for i, result := range results {
		if errors[i] == nil {
			totalScore += result.Score
			sentimentCounts[result.Sentiment]++
			for _, kw := range result.Keywords {
				allKeywords[kw]++
			}
		}
	}

	avgScore := 0.0
	if successCount > 0 {
		avgScore = totalScore / float64(successCount)
	}

	aggregateSpan.End(agenttrace.EndParams{
		Output: map[string]interface{}{
			"averageScore":    avgScore,
			"sentimentCounts": sentimentCounts,
			"uniqueKeywords":  len(allKeywords),
		},
	})

	// Phase 3: Generate summary report
	fmt.Println("\nPhase 3: Generating summary report...")
	summaryGen := trace.Generation(ctx, agenttrace.GenerationParams{
		Name:  "generate-summary-report",
		Model: "gpt-4",
		ModelParameters: map[string]interface{}{
			"temperature": 0.3,
		},
		Input: []map[string]interface{}{
			{
				"role":    "system",
				"content": "You are a document analyst. Generate a summary report.",
			},
			{
				"role":    "user",
				"content": fmt.Sprintf("Summarize these %d processed documents with avg score %.2f", successCount, avgScore),
			},
		},
		Metadata: map[string]interface{}{
			"phase": 3,
		},
	})

	summaryResponse, inputTokens, outputTokens, err := MockLLMCall(ctx,
		fmt.Sprintf("Summarize %d documents", successCount))

	if err != nil {
		summaryGen.End(agenttrace.GenerationEndParams{
			Output: map[string]interface{}{
				"error": err.Error(),
			},
			Level: "error",
			StatusMessage: err.Error(),
		})
		fmt.Printf("  ✗ Summary generation failed: %v\n", err)
	} else {
		summaryGen.End(agenttrace.GenerationEndParams{
			Output: map[string]interface{}{
				"role":    "assistant",
				"content": summaryResponse,
			},
			Usage: &agenttrace.Usage{
				InputTokens:  inputTokens,
				OutputTokens: outputTokens,
				TotalTokens:  inputTokens + outputTokens,
			},
		})
		fmt.Printf("  ✓ Summary report generated\n")
	}

	// Create a checkpoint of the final state
	fmt.Println("\nCreating checkpoint...")
	finalState := map[string]interface{}{
		"results":         results,
		"averageScore":    avgScore,
		"sentimentCounts": sentimentCounts,
		"summary":         summaryResponse,
	}
	finalStateJSON, _ := json.Marshal(finalState)

	trace.Event(ctx, agenttrace.EventParams{
		Name: "checkpoint-created",
		Input: map[string]interface{}{
			"checkpointData": string(finalStateJSON),
			"timestamp":      time.Now().Format(time.RFC3339),
		},
	})

	// Complete the trace
	trace.Update(agenttrace.UpdateParams{
		Output: map[string]interface{}{
			"processedDocuments": successCount,
			"failedDocuments":    failureCount,
			"averageScore":       avgScore,
			"topKeywords":        getTopKeywords(allKeywords, 5),
		},
		Metadata: map[string]interface{}{
			"status":   "completed",
			"duration": "estimated",
		},
	})

	// Add a score for the workflow
	trace.Score(ctx, agenttrace.ScoreParams{
		Name:  "workflow_success_rate",
		Value: float64(successCount) / float64(len(documents)),
		Comment: fmt.Sprintf("%d of %d documents processed successfully", successCount, len(documents)),
	})

	// Flush all data
	fmt.Println("\nFlushing data to AgentTrace...")
	if err := client.Flush(ctx); err != nil {
		fmt.Printf("Failed to flush: %v\n", err)
	}

	fmt.Println("\n=== Summary ===")
	fmt.Printf("Documents processed: %d/%d\n", successCount, len(documents))
	fmt.Printf("Average score: %.2f\n", avgScore)
	fmt.Printf("Unique keywords found: %d\n", len(allKeywords))
	fmt.Println("\nDone! View your trace at: http://localhost:3000/traces")
}

// processDocument processes a single document with multiple LLM calls
func processDocument(ctx context.Context, trace *agenttrace.Trace, doc Document) (ProcessingResult, error) {
	result := ProcessingResult{
		DocumentID: doc.ID,
	}

	// Create a span for this document's processing
	docSpan := trace.Span(ctx, agenttrace.SpanParams{
		Name: fmt.Sprintf("process-document-%s", doc.ID),
		Metadata: map[string]interface{}{
			"documentId": doc.ID,
			"title":      doc.Title,
		},
		Input: map[string]interface{}{
			"content": doc.Content,
		},
	})

	// Step 1: Summarize
	summaryGen := trace.Generation(ctx, agenttrace.GenerationParams{
		Name:  fmt.Sprintf("summarize-%s", doc.ID),
		Model: "gpt-3.5-turbo",
		ModelParameters: map[string]interface{}{
			"temperature": 0.5,
			"max_tokens":  150,
		},
		Input: []map[string]interface{}{
			{"role": "user", "content": fmt.Sprintf("Summarize: %s", doc.Content)},
		},
		ParentObservationID: docSpan.ID(),
	})

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
	result.Summary = summary
	summaryGen.End(agenttrace.GenerationEndParams{
		Output: map[string]interface{}{"summary": summary},
		Usage:  &agenttrace.Usage{InputTokens: inTok, OutputTokens: outTok, TotalTokens: inTok + outTok},
	})

	// Step 2: Sentiment analysis
	sentimentGen := trace.Generation(ctx, agenttrace.GenerationParams{
		Name:  fmt.Sprintf("sentiment-%s", doc.ID),
		Model: "gpt-3.5-turbo",
		Input: []map[string]interface{}{
			{"role": "user", "content": fmt.Sprintf("Analyze sentiment: %s", doc.Content)},
		},
		ParentObservationID: docSpan.ID(),
	})

	sentiments := []string{"positive", "negative", "neutral"}
	result.Sentiment = sentiments[rand.Intn(len(sentiments))]
	sentimentGen.End(agenttrace.GenerationEndParams{
		Output: map[string]interface{}{"sentiment": result.Sentiment},
		Usage:  &agenttrace.Usage{InputTokens: 20, OutputTokens: 5, TotalTokens: 25},
	})

	// Step 3: Extract keywords
	keywordGen := trace.Generation(ctx, agenttrace.GenerationParams{
		Name:  fmt.Sprintf("keywords-%s", doc.ID),
		Model: "gpt-3.5-turbo",
		Input: []map[string]interface{}{
			{"role": "user", "content": fmt.Sprintf("Extract keywords: %s", doc.Content)},
		},
		ParentObservationID: docSpan.ID(),
	})

	// Simulate keyword extraction
	allKeywords := []string{"technology", "programming", "cloud", "AI", "development", "data", "software"}
	numKeywords := 2 + rand.Intn(3)
	result.Keywords = make([]string, numKeywords)
	for i := 0; i < numKeywords; i++ {
		result.Keywords[i] = allKeywords[rand.Intn(len(allKeywords))]
	}
	keywordGen.End(agenttrace.GenerationEndParams{
		Output: map[string]interface{}{"keywords": result.Keywords},
		Usage:  &agenttrace.Usage{InputTokens: 30, OutputTokens: 10, TotalTokens: 40},
	})

	// Calculate final score
	result.Score = 0.5 + rand.Float64()*0.5

	// End document span
	docSpan.End(agenttrace.EndParams{
		Output: map[string]interface{}{
			"summary":   result.Summary,
			"sentiment": result.Sentiment,
			"keywords":  result.Keywords,
			"score":     result.Score,
		},
	})

	return result, nil
}

// getTopKeywords returns the top N keywords by frequency
func getTopKeywords(keywords map[string]int, n int) []string {
	type kv struct {
		Key   string
		Value int
	}
	var sorted []kv
	for k, v := range keywords {
		sorted = append(sorted, kv{k, v})
	}

	// Simple bubble sort for small N
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[j].Value > sorted[i].Value {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	result := make([]string, 0, n)
	for i := 0; i < min(n, len(sorted)); i++ {
		result = append(result, sorted[i].Key)
	}
	return result
}
