# Jupyter Integration

AgentTrace provides a JupyterLab extension for trace visualization and auto-instrumentation directly within your notebooks.

## Installation

```bash
pip install agenttrace-jupyter
```

## Configuration

Set your API key as an environment variable before starting JupyterLab:

```bash
export AGENTTRACE_API_KEY=your-api-key
jupyter lab
```

Or configure within JupyterLab using the sidebar panel.

## Features

### Sidebar Panel

The AgentTrace sidebar shows:

- **Connection Status** - Whether AgentTrace is configured
- **Session Metrics** - Cell executions, LLM calls, tokens, and costs
- **Recent Traces** - Click to view details
- **Auto-Trace Toggle** - Enable/disable automatic tracing

Access the sidebar by clicking "AgentTrace" in the notebook toolbar.

### Auto-Instrumentation

When enabled, the extension automatically traces:

- Cell executions containing LLM calls
- AgentTrace SDK decorators (`@observe`, `@generation`)
- Direct API calls to OpenAI, Anthropic, etc.

### Real-Time Metrics

Track your notebook session:

| Metric | Description |
|--------|-------------|
| Cell Executions | Total cells executed |
| LLM Calls | Number of LLM API calls |
| Total Tokens | Aggregate token usage |
| Total Cost | Estimated API costs |

## Usage Examples

### Basic Tracing

```python
from agenttrace import observe, generation

@observe()
def analyze_data(data):
    """This function will be automatically traced."""
    with generation(name="analysis", model="gpt-4") as gen:
        prompt = f"Analyze this data: {data}"
        response = call_llm(prompt)
        gen.update(
            input={"prompt": prompt},
            output=response,
            usage={"total_tokens": 150}
        )
    return response

# Run in a cell - trace appears in sidebar
result = analyze_data(my_dataframe.describe())
```

### Manual Tracing

```python
from agenttrace import AgentTrace

client = AgentTrace()

# Create a trace for complex operations
with client.trace(name="data-pipeline") as trace:
    # Step 1: Data loading
    with trace.span(name="load-data"):
        data = load_data()

    # Step 2: LLM analysis
    with trace.generation(name="analyze", model="claude-3-5-sonnet-20241022") as gen:
        analysis = analyze_with_llm(data)
        gen.update(output=analysis)

    # Step 3: Visualization
    with trace.span(name="visualize"):
        create_charts(analysis)
```

### Viewing Inline Traces

After execution, traces appear in the sidebar. Click a trace to see:

```
Timeline:
├─ data-pipeline (2.5s)
│  ├─ load-data (0.3s)
│  ├─ analyze [GPT-4] (1.8s) - $0.02
│  └─ visualize (0.4s)

Metrics:
- Total tokens: 1,250
- Cost: $0.02
- Duration: 2.5s
```

## Configuration Options

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `AGENTTRACE_API_KEY` | Your API key | Required |
| `AGENTTRACE_BASE_URL` | API endpoint | `https://api.agenttrace.io` |
| `AGENTTRACE_PROJECT_ID` | Default project | None |
| `AGENTTRACE_AUTO_TRACE` | Enable auto-tracing | `true` |

### In-Notebook Configuration

```python
import os

# Configure at runtime
os.environ["AGENTTRACE_API_KEY"] = "your-key"
os.environ["AGENTTRACE_AUTO_TRACE"] = "true"

# Or via the extension API
from agenttrace_jupyter import configure

configure(
    api_key="your-key",
    auto_trace=True,
    project_id="my-project"
)
```

## Tips for Notebooks

### Organizing Traces

Use meaningful names for your traces:

```python
@observe(name="feature-engineering")
def prepare_features(df):
    ...

@observe(name="model-training")
def train_model(X, y):
    ...

@observe(name="evaluation")
def evaluate(model, X_test, y_test):
    ...
```

### Cost Tracking

Monitor costs across your notebook session:

```python
# The sidebar shows cumulative costs
# For programmatic access:
from agenttrace_jupyter import get_session_metrics

metrics = get_session_metrics()
print(f"Session cost: ${metrics['total_cost']:.4f}")
print(f"Total tokens: {metrics['total_tokens']}")
```

### Exporting Traces

Export traces for sharing or analysis:

```python
from agenttrace import AgentTrace

client = AgentTrace()

# Get trace by ID (from sidebar)
trace = client.get_trace("trace-id")

# Export to dict
trace_data = trace.to_dict()

# Save to file
import json
with open("trace_export.json", "w") as f:
    json.dump(trace_data, f, indent=2)
```

## Troubleshooting

### Extension Not Loading

1. Verify installation:
   ```bash
   jupyter labextension list
   # Should show @agenttrace/jupyter
   ```

2. Rebuild if needed:
   ```bash
   jupyter lab build
   ```

### Not Configured Error

Ensure your API key is set:

```bash
echo $AGENTTRACE_API_KEY
# Should show your key
```

### Traces Not Appearing

1. Check auto-trace is enabled in the sidebar
2. Verify your code uses AgentTrace SDK decorators
3. Check the browser console for errors

### Performance Issues

For large notebooks with many traces:

```python
# Disable auto-trace for performance
import os
os.environ["AGENTTRACE_AUTO_TRACE"] = "false"

# Manually trace important cells only
from agenttrace import observe

@observe()
def important_function():
    ...
```

## Development

Install in development mode:

```bash
cd extensions/jupyter
pip install -e ".[test]"
jupyter labextension develop . --overwrite
jlpm watch
```

Run tests:

```bash
pytest
```

## API Reference

### Python API

```python
from agenttrace_jupyter import (
    configure,           # Configure the extension
    get_session_metrics, # Get current session metrics
    get_recent_traces,   # Get list of recent traces
    refresh_traces,      # Refresh trace list
)
```

### REST Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/agenttrace/config` | GET | Get configuration |
| `/agenttrace/config` | POST | Update configuration |
| `/agenttrace/traces` | GET | List recent traces |
| `/agenttrace/traces` | POST | Create trace |
| `/agenttrace/metrics` | GET | Get session metrics |
