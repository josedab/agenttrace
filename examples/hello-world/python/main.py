"""
AgentTrace Hello World Example (Python)

This is the simplest possible example to get started with AgentTrace.
It demonstrates creating a trace, adding spans, and recording a generation.

Prerequisites:
    pip install agenttrace
    export AGENTTRACE_API_KEY="your-api-key"

Run:
    python main.py
"""

import os
import time
from agenttrace import AgentTrace


def main():
    # Initialize the AgentTrace client
    # It will use AGENTTRACE_API_KEY and AGENTTRACE_HOST from environment
    client = AgentTrace(
        api_key=os.getenv("AGENTTRACE_API_KEY"),
        host=os.getenv("AGENTTRACE_HOST", "http://localhost:8080"),
    )

    print("Creating a trace...")

    # Create a trace - the root container for an operation
    trace = client.trace(
        name="hello-world",
        metadata={"language": "python", "example": "hello-world"},
        input={"question": "What is AgentTrace?"},
    )

    # Step 1: Add a span for preprocessing
    print("  Adding preprocessing span...")
    preprocess_span = trace.span(
        name="preprocess-input",
        metadata={"step": 1},
        input={"raw_question": "What is AgentTrace?"},
    )
    time.sleep(0.1)  # Simulate some work
    preprocess_span.end(output={"processed_question": "Explain AgentTrace"})

    # Step 2: Record an LLM generation (simulated)
    print("  Recording LLM generation...")
    generation = trace.generation(
        name="llm-call",
        model="gpt-4",
        model_parameters={"temperature": 0.7, "max_tokens": 500},
        input=[{"role": "user", "content": "Explain AgentTrace"}],
        metadata={"step": 2},
    )
    time.sleep(0.2)  # Simulate LLM latency

    # Complete the generation with output and usage
    llm_response = (
        "AgentTrace is an open-source observability platform for AI coding agents. "
        "It helps you trace, debug, and monitor autonomous AI agents."
    )
    generation.end(
        output={"role": "assistant", "content": llm_response},
        usage={
            "input_tokens": 10,
            "output_tokens": 35,
            "total_tokens": 45,
        },
    )

    # Step 3: Add a span for postprocessing
    print("  Adding postprocessing span...")
    postprocess_span = trace.span(
        name="postprocess-output",
        metadata={"step": 3},
        input={"raw_response": llm_response},
    )
    time.sleep(0.05)  # Simulate formatting
    final_output = f"Answer: {llm_response}"
    postprocess_span.end(output={"formatted_response": final_output})

    # Complete the trace
    trace.update(
        output={"answer": final_output},
        metadata={"status": "success"},
    )

    print("  Flushing data to AgentTrace...")

    # Ensure all data is sent before exiting
    client.flush()
    client.shutdown()

    print("\nDone! View your trace at: http://localhost:3000/traces")
    print("Look for a trace named 'hello-world'")


if __name__ == "__main__":
    main()
