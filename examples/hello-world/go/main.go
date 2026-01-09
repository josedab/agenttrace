/*
AgentTrace Hello World Example (Go)

This is the simplest possible example to get started with AgentTrace.
It demonstrates creating a trace, adding spans, and recording a generation.

Prerequisites:

	go mod tidy
	export AGENTTRACE_API_KEY="your-api-key"

Run:

	go run main.go
*/
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/agenttrace/agenttrace-go"
)

func main() {
	ctx := context.Background()

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

	fmt.Println("Creating a trace...")

	// Create a trace - the root container for an operation
	trace := client.Trace(ctx, agenttrace.TraceParams{
		Name: "hello-world",
		Metadata: map[string]interface{}{
			"language": "go",
			"example":  "hello-world",
		},
		Input: map[string]interface{}{
			"question": "What is AgentTrace?",
		},
	})

	// Step 1: Add a span for preprocessing
	fmt.Println("  Adding preprocessing span...")
	preprocessSpan := trace.Span(ctx, agenttrace.SpanParams{
		Name: "preprocess-input",
		Metadata: map[string]interface{}{
			"step": 1,
		},
		Input: map[string]interface{}{
			"raw_question": "What is AgentTrace?",
		},
	})
	time.Sleep(100 * time.Millisecond) // Simulate some work
	preprocessSpan.End(agenttrace.EndParams{
		Output: map[string]interface{}{
			"processed_question": "Explain AgentTrace",
		},
	})

	// Step 2: Record an LLM generation (simulated)
	fmt.Println("  Recording LLM generation...")
	generation := trace.Generation(ctx, agenttrace.GenerationParams{
		Name:  "llm-call",
		Model: "gpt-4",
		ModelParameters: map[string]interface{}{
			"temperature": 0.7,
			"max_tokens":  500,
		},
		Input: []map[string]interface{}{
			{"role": "user", "content": "Explain AgentTrace"},
		},
		Metadata: map[string]interface{}{
			"step": 2,
		},
	})
	time.Sleep(200 * time.Millisecond) // Simulate LLM latency

	// Complete the generation with output and usage
	llmResponse := "AgentTrace is an open-source observability platform for AI coding agents. " +
		"It helps you trace, debug, and monitor autonomous AI agents."

	generation.End(agenttrace.GenerationEndParams{
		Output: map[string]interface{}{
			"role":    "assistant",
			"content": llmResponse,
		},
		Usage: &agenttrace.Usage{
			InputTokens:  10,
			OutputTokens: 35,
			TotalTokens:  45,
		},
	})

	// Step 3: Add a span for postprocessing
	fmt.Println("  Adding postprocessing span...")
	postprocessSpan := trace.Span(ctx, agenttrace.SpanParams{
		Name: "postprocess-output",
		Metadata: map[string]interface{}{
			"step": 3,
		},
		Input: map[string]interface{}{
			"raw_response": llmResponse,
		},
	})
	time.Sleep(50 * time.Millisecond) // Simulate formatting
	finalOutput := fmt.Sprintf("Answer: %s", llmResponse)
	postprocessSpan.End(agenttrace.EndParams{
		Output: map[string]interface{}{
			"formatted_response": finalOutput,
		},
	})

	// Complete the trace
	trace.Update(agenttrace.UpdateParams{
		Output: map[string]interface{}{
			"answer": finalOutput,
		},
		Metadata: map[string]interface{}{
			"status": "success",
		},
	})

	fmt.Println("  Flushing data to AgentTrace...")

	// Ensure all data is sent before exiting
	if err := client.Flush(ctx); err != nil {
		fmt.Printf("Failed to flush: %v\n", err)
	}

	fmt.Println("\nDone! View your trace at: http://localhost:3000/traces")
	fmt.Println("Look for a trace named 'hello-world'")
}
