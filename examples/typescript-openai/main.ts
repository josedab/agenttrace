/**
 * AgentTrace TypeScript + OpenAI Example
 *
 * This example demonstrates integrating AgentTrace with OpenAI's API,
 * including streaming responses, function calling, and error handling.
 *
 * Prerequisites:
 *     npm install
 *     export AGENTTRACE_API_KEY="your-api-key"
 *     export OPENAI_API_KEY="your-openai-key"
 *
 * Run:
 *     npx ts-node main.ts
 */

import { AgentTrace } from "@agenttrace/sdk";
import OpenAI from "openai";

// Tool definitions for function calling
const tools: OpenAI.ChatCompletionTool[] = [
  {
    type: "function",
    function: {
      name: "get_weather",
      description: "Get the current weather in a given location",
      parameters: {
        type: "object",
        properties: {
          location: {
            type: "string",
            description: "The city and state, e.g. San Francisco, CA",
          },
          unit: {
            type: "string",
            enum: ["celsius", "fahrenheit"],
          },
        },
        required: ["location"],
      },
    },
  },
  {
    type: "function",
    function: {
      name: "search_web",
      description: "Search the web for information",
      parameters: {
        type: "object",
        properties: {
          query: {
            type: "string",
            description: "The search query",
          },
        },
        required: ["query"],
      },
    },
  },
];

// Mock tool implementations
function getWeather(location: string, unit: string = "fahrenheit"): string {
  // In a real app, this would call a weather API
  const temp = unit === "celsius" ? "22" : "72";
  return JSON.stringify({
    location,
    temperature: temp,
    unit,
    condition: "sunny",
  });
}

function searchWeb(query: string): string {
  // In a real app, this would call a search API
  return JSON.stringify({
    results: [
      { title: `Result 1 for "${query}"`, snippet: "Sample search result..." },
      { title: `Result 2 for "${query}"`, snippet: "Another result..." },
    ],
  });
}

async function main(): Promise<void> {
  // Initialize clients
  const agenttrace = new AgentTrace({
    apiKey: process.env.AGENTTRACE_API_KEY,
    host: process.env.AGENTTRACE_HOST || "http://localhost:8080",
  });

  const openai = new OpenAI({
    apiKey: process.env.OPENAI_API_KEY,
  });

  console.log("Starting OpenAI agent with tool use...\n");

  // Create the main trace
  const trace = agenttrace.trace({
    name: "openai-agent-with-tools",
    metadata: {
      example: "typescript-openai",
      features: ["streaming", "function-calling", "error-handling"],
    },
    input: {
      userQuery:
        "What's the weather like in San Francisco and New York today?",
    },
  });

  try {
    const messages: OpenAI.ChatCompletionMessageParam[] = [
      {
        role: "system",
        content:
          "You are a helpful assistant with access to tools. Use them when needed.",
      },
      {
        role: "user",
        content:
          "What's the weather like in San Francisco and New York today?",
      },
    ];

    let iteration = 0;
    const maxIterations = 5;

    while (iteration < maxIterations) {
      iteration++;
      console.log(`\n--- Iteration ${iteration} ---`);

      // Create generation for this LLM call
      const generation = trace.generation({
        name: `llm-call-${iteration}`,
        model: "gpt-4o-mini",
        modelParameters: { temperature: 0.7 },
        input: messages,
        metadata: { iteration },
      });

      // Make streaming API call
      console.log("Calling OpenAI (streaming)...");

      const stream = await openai.chat.completions.create({
        model: "gpt-4o-mini",
        messages,
        tools,
        stream: true,
      });

      // Collect streamed response
      let content = "";
      let toolCalls: OpenAI.ChatCompletionMessageToolCall[] = [];
      let finishReason: string | null = null;
      let promptTokens = 0;
      let completionTokens = 0;

      for await (const chunk of stream) {
        const delta = chunk.choices[0]?.delta;
        const choice = chunk.choices[0];

        if (delta?.content) {
          content += delta.content;
          process.stdout.write(delta.content);
        }

        if (delta?.tool_calls) {
          for (const toolCall of delta.tool_calls) {
            if (toolCalls[toolCall.index] === undefined) {
              toolCalls[toolCall.index] = {
                id: toolCall.id || "",
                type: "function",
                function: {
                  name: toolCall.function?.name || "",
                  arguments: "",
                },
              };
            }
            if (toolCall.function?.arguments) {
              toolCalls[toolCall.index].function.arguments +=
                toolCall.function.arguments;
            }
          }
        }

        if (choice?.finish_reason) {
          finishReason = choice.finish_reason;
        }

        // Note: usage is typically only in the final chunk
        if (chunk.usage) {
          promptTokens = chunk.usage.prompt_tokens;
          completionTokens = chunk.usage.completion_tokens;
        }
      }

      console.log(""); // New line after streaming

      // End generation with collected data
      generation.end({
        output: {
          content,
          tool_calls:
            toolCalls.length > 0
              ? toolCalls.map((tc) => ({
                  id: tc.id,
                  function: tc.function.name,
                  arguments: tc.function.arguments,
                }))
              : undefined,
          finish_reason: finishReason,
        },
        usage: {
          inputTokens: promptTokens,
          outputTokens: completionTokens,
          totalTokens: promptTokens + completionTokens,
        },
      });

      // If no tool calls, we're done
      if (toolCalls.length === 0) {
        console.log("\n✓ Agent completed without tool calls");
        trace.update({
          output: { finalResponse: content },
          metadata: { iterations: iteration, status: "completed" },
        });
        break;
      }

      // Process tool calls
      console.log(`\nProcessing ${toolCalls.length} tool call(s)...`);

      // Add assistant message with tool calls
      messages.push({
        role: "assistant",
        content: content || null,
        tool_calls: toolCalls,
      });

      // Execute each tool and record spans
      for (const toolCall of toolCalls) {
        const toolSpan = trace.span({
          name: `tool-${toolCall.function.name}`,
          metadata: { toolCallId: toolCall.id },
          input: { arguments: toolCall.function.arguments },
        });

        let result: string;

        try {
          const args = JSON.parse(toolCall.function.arguments);

          switch (toolCall.function.name) {
            case "get_weather":
              console.log(`  → Calling get_weather(${args.location})`);
              result = getWeather(args.location, args.unit);
              break;
            case "search_web":
              console.log(`  → Calling search_web(${args.query})`);
              result = searchWeb(args.query);
              break;
            default:
              result = JSON.stringify({ error: "Unknown tool" });
          }

          toolSpan.end({ output: { result: JSON.parse(result) } });
        } catch (error) {
          result = JSON.stringify({
            error: error instanceof Error ? error.message : "Unknown error",
          });
          toolSpan.end({
            output: { error: result },
            level: "error",
            statusMessage: `Tool execution failed: ${result}`,
          });
        }

        // Add tool result to messages
        messages.push({
          role: "tool",
          tool_call_id: toolCall.id,
          content: result,
        });
      }
    }
  } catch (error) {
    console.error("\n✗ Agent failed:", error);

    // Record the error in the trace
    trace.update({
      output: { error: error instanceof Error ? error.message : "Unknown error" },
      metadata: { status: "error" },
      level: "error",
      statusMessage:
        error instanceof Error ? error.message : "Unknown error occurred",
    });
  } finally {
    // Ensure data is sent
    console.log("\nFlushing data to AgentTrace...");
    await agenttrace.flush();
    await agenttrace.shutdown();
  }

  console.log(
    "\nDone! View your trace at: http://localhost:3000/traces"
  );
}

main().catch(console.error);
