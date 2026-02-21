# AgentTrace Benchmark Suite

Curated benchmark datasets for evaluating AI coding agents. Run any benchmark with a single API call to compare your agent against community baselines.

## Available Benchmarks

### 1. Code Generation (`code-generation`)
Tests the agent's ability to generate correct code from natural language descriptions.
- **20 tasks** ranging from simple functions to complex algorithms
- **Metrics**: correctness (0-1), code quality (0-1), time-to-solution (seconds)
- **Evaluators**: Unit test pass rate + LLM-as-Judge for code quality

### 2. Bug Fixing (`bug-fixing`)
Tests the agent's ability to identify and fix bugs in existing code.
- **15 tasks** with known bugs in Python, TypeScript, and Go
- **Metrics**: fix accuracy (0-1), regression rate (0-1), explanation quality (0-1)
- **Evaluators**: Diff comparison + test suite + LLM review

### 3. Code Refactoring (`refactoring`)
Tests the agent's ability to improve code without changing behavior.
- **10 tasks** with code smells, performance issues, and readability problems
- **Metrics**: behavior preservation (0-1), improvement score (0-1), readability (0-1)
- **Evaluators**: Test equivalence + complexity reduction + LLM review

### 4. Documentation (`documentation`)
Tests the agent's ability to write and improve code documentation.
- **10 tasks** requiring docstrings, README updates, and API docs
- **Metrics**: completeness (0-1), accuracy (0-1), clarity (0-1)
- **Evaluators**: LLM-as-Judge + template compliance

### 5. Multi-Step Tasks (`multi-step`)
Tests the agent's ability to complete complex multi-file, multi-step tasks.
- **5 tasks** requiring planning, execution, and verification
- **Metrics**: completion rate (0-1), step accuracy (0-1), cost efficiency (0-1)
- **Evaluators**: End-to-end test suite + cost analysis

## Quick Start

```bash
# List available benchmarks
curl -H "Authorization: Bearer $API_KEY" \
  https://api.agenttrace.io/api/public/benchmarks

# Submit a benchmark run
curl -X POST -H "Authorization: Bearer $API_KEY" \
  -H "Content-Type: application/json" \
  https://api.agenttrace.io/api/public/benchmarks/{id}/submit \
  -d '{"agentName": "my-agent", "agentVersion": "1.0.0"}'

# View leaderboard
curl -H "Authorization: Bearer $API_KEY" \
  https://api.agenttrace.io/api/public/benchmarks/{id}/leaderboard
```

## SDK Usage

```python
from agenttrace import AgentTrace

client = AgentTrace(api_key="...")

# Run benchmark
result = client.benchmarks.submit(
    benchmark_id="code-generation",
    agent_name="my-agent",
    agent_version="1.0.0",
)

print(f"Overall Score: {result.overall_score}")
print(f"Rank: #{result.rank}")
```
