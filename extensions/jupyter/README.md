# AgentTrace Jupyter Extension

A JupyterLab extension for AgentTrace - trace visualization and auto-instrumentation for AI/LLM workflows in notebooks.

## Features

- **Auto-Instrumentation**: Automatically trace cell executions with LLM calls
- **Sidebar Panel**: View traces, metrics, and costs in real-time
- **Trace Visualization**: Timeline view of LLM calls and tool usage
- **Session Metrics**: Track total tokens, costs, and execution counts
- **Integration**: Works with the AgentTrace Python SDK

## Requirements

- JupyterLab >= 4.0.0
- Python >= 3.8
- AgentTrace Python SDK

## Installation

```bash
pip install agenttrace-jupyter
```

Or install from source:

```bash
# Clone the repository
git clone https://github.com/agenttrace/agenttrace.git
cd agenttrace/extensions/jupyter

# Install in development mode
pip install -e ".[test]"

# Link for development
jupyter labextension develop . --overwrite

# Rebuild after changes
jlpm build
```

## Configuration

Set your AgentTrace API key as an environment variable:

```bash
export AGENTTRACE_API_KEY=your-api-key
export AGENTTRACE_BASE_URL=https://api.agenttrace.io  # Optional, defaults to cloud
export AGENTTRACE_PROJECT_ID=your-project-id         # Optional
```

Or configure within JupyterLab by clicking the AgentTrace button in the toolbar.

## Usage

### Viewing the Sidebar

Click the "AgentTrace" button in the notebook toolbar or use the command palette:
1. Press `Ctrl+Shift+C` (or `Cmd+Shift+C` on Mac)
2. Search for "AgentTrace"
3. Select "Show AgentTrace Panel"

### Auto-Tracing

By default, the extension traces all cell executions that include LLM calls. Toggle this in the sidebar panel.

### Using with AgentTrace SDK

```python
from agenttrace import observe, generation

@observe()
def my_agent_function():
    with generation(name="llm-call", model="gpt-4") as gen:
        response = call_openai("Tell me a joke")
        gen.update(output=response)
    return response

# Traces will appear in the sidebar automatically
result = my_agent_function()
```

### Viewing Trace Details

Click on any trace in the sidebar to see:
- Timeline of all spans
- Token usage per call
- Cost breakdown
- Input/output for each LLM call

## Development

### Build

```bash
# Install dependencies
jlpm

# Build TypeScript
jlpm build

# Watch for changes
jlpm watch
```

### Testing

```bash
# Run Python tests
pytest

# Run linting
jlpm lint
```

### Uninstall

```bash
pip uninstall agenttrace-jupyter
```

## Contributing

See the [main AgentTrace contributing guide](../../CONTRIBUTING.md).

## License

MIT License - see [LICENSE](../../LICENSE) for details.
