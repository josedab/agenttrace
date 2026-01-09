/**
 * AgentTrace Hello World Example (TypeScript)
 *
 * This is the simplest possible example to get started with AgentTrace.
 * It demonstrates creating a trace, adding spans, and recording a generation.
 *
 * Prerequisites:
 *     npm install
 *     export AGENTTRACE_API_KEY="your-api-key"
 *
 * Run:
 *     npx ts-node main.ts
 */

import { AgentTrace } from "@agenttrace/sdk";

async function sleep(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

async function main(): Promise<void> {
  // Initialize the AgentTrace client
  const client = new AgentTrace({
    apiKey: process.env.AGENTTRACE_API_KEY,
    host: process.env.AGENTTRACE_HOST || "http://localhost:8080",
  });

  console.log("Creating a trace...");

  // Create a trace - the root container for an operation
  const trace = client.trace({
    name: "hello-world",
    metadata: { language: "typescript", example: "hello-world" },
    input: { question: "What is AgentTrace?" },
  });

  // Step 1: Add a span for preprocessing
  console.log("  Adding preprocessing span...");
  const preprocessSpan = trace.span({
    name: "preprocess-input",
    metadata: { step: 1 },
    input: { raw_question: "What is AgentTrace?" },
  });
  await sleep(100); // Simulate some work
  preprocessSpan.end({ output: { processed_question: "Explain AgentTrace" } });

  // Step 2: Record an LLM generation (simulated)
  console.log("  Recording LLM generation...");
  const generation = trace.generation({
    name: "llm-call",
    model: "gpt-4",
    modelParameters: { temperature: 0.7, max_tokens: 500 },
    input: [{ role: "user", content: "Explain AgentTrace" }],
    metadata: { step: 2 },
  });
  await sleep(200); // Simulate LLM latency

  // Complete the generation with output and usage
  const llmResponse =
    "AgentTrace is an open-source observability platform for AI coding agents. " +
    "It helps you trace, debug, and monitor autonomous AI agents.";

  generation.end({
    output: { role: "assistant", content: llmResponse },
    usage: {
      inputTokens: 10,
      outputTokens: 35,
      totalTokens: 45,
    },
  });

  // Step 3: Add a span for postprocessing
  console.log("  Adding postprocessing span...");
  const postprocessSpan = trace.span({
    name: "postprocess-output",
    metadata: { step: 3 },
    input: { raw_response: llmResponse },
  });
  await sleep(50); // Simulate formatting
  const finalOutput = `Answer: ${llmResponse}`;
  postprocessSpan.end({ output: { formatted_response: finalOutput } });

  // Complete the trace
  trace.update({
    output: { answer: finalOutput },
    metadata: { status: "success" },
  });

  console.log("  Flushing data to AgentTrace...");

  // Ensure all data is sent before exiting
  await client.flush();
  await client.shutdown();

  console.log("\nDone! View your trace at: http://localhost:3000/traces");
  console.log("Look for a trace named 'hello-world'");
}

main().catch(console.error);
