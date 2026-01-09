# AgentTrace for VS Code

AI Agent Observability directly in your editor. View traces, monitor costs, and debug AI agent executions without leaving VS Code.

## Features

### Trace Explorer

View recent traces directly in VS Code:
- See trace status, duration, cost, and token usage
- Expand traces to view observations (spans, generations, events)
- Click to view detailed trace information
- Filter traces by status, time range, or search

### Session Management

Track conversation sessions:
- View all active sessions
- See trace counts and total costs per session
- Navigate between session traces

### Git Integration

Correlate traces with your code changes:
- View traces by git commit
- Link traces to specific commits
- See which code changes triggered which agent behaviors

### Cost Monitoring

Track your LLM spend:
- Real-time cost display in the status bar
- Daily, weekly, and monthly cost summaries
- Cost breakdown by model

### File-Based Trace Filtering

Find traces related to specific files:
- Filter traces by the currently open file
- See which agent runs modified which files

## Installation

### From VS Code Marketplace

1. Open VS Code
2. Go to Extensions (Ctrl+Shift+X / Cmd+Shift+X)
3. Search for "AgentTrace"
4. Click Install

### Manual Installation

1. Download the `.vsix` file from the [releases page](https://github.com/agenttrace/agenttrace/releases)
2. In VS Code, go to Extensions
3. Click the "..." menu and select "Install from VSIX..."
4. Select the downloaded file

## Configuration

After installation, configure the extension:

1. Open Command Palette (Ctrl+Shift+P / Cmd+Shift+P)
2. Run "AgentTrace: Configure"
3. Set your API Key and Project ID

Or configure via settings:

```json
{
  "agenttrace.apiUrl": "https://api.agenttrace.io",
  "agenttrace.dashboardUrl": "https://app.agenttrace.io",
  "agenttrace.apiKey": "your-api-key",
  "agenttrace.projectId": "your-project-id",
  "agenttrace.autoRefresh": true,
  "agenttrace.refreshInterval": 30,
  "agenttrace.showStatusBarItem": true,
  "agenttrace.enableFileDecorations": true,
  "agenttrace.maxTraces": 50
}
```

### Self-Hosted Setup

If you're running AgentTrace locally or self-hosted:

```json
{
  "agenttrace.apiUrl": "http://localhost:8080",
  "agenttrace.dashboardUrl": "http://localhost:3000"
}
```

## Commands

Access these commands from the Command Palette (Ctrl+Shift+P / Cmd+Shift+P):

| Command | Description |
|---------|-------------|
| AgentTrace: Refresh Traces | Manually refresh the traces list |
| AgentTrace: View Trace Details | View detailed information about a trace |
| AgentTrace: Open in Browser | Open the selected trace in the web dashboard |
| AgentTrace: Configure | Configure API key and settings |
| AgentTrace: Link Current Commit | Link the current git commit to a trace |
| AgentTrace: Create Checkpoint | Create a checkpoint for a trace |
| AgentTrace: Show Cost Summary | View your cost breakdown |
| AgentTrace: Filter by File | Find traces related to the current file |
| AgentTrace: Copy Trace ID | Copy a trace ID to clipboard |

## Views

The extension adds an AgentTrace icon to the Activity Bar with three views:

### Recent Traces

Shows the most recent traces for your project. Each trace displays:
- Name and status (completed, running, error)
- Duration and cost
- Token count
- Git commit (if linked)

Expand a trace to see its observations:
- **Generations** (purple): LLM calls with model, tokens, and cost
- **Spans** (blue): Custom spans for functions or operations
- **Events** (orange): Point-in-time events

### Sessions

Groups traces by session ID. Useful for tracking multi-turn conversations.

### Git History

Shows traces correlated with git commits. Useful for understanding which code changes affected agent behavior.

## Status Bar

When enabled, the status bar shows:
- Today's total cost
- Click to view detailed cost breakdown by time period and model

## File Decorations

When enabled, files that are referenced in recent traces will show a trace indicator in the explorer.

## Tips

### Quick Access to Traces

- Click any trace in the tree view to see a detailed breakdown
- Right-click for additional options (open in browser, copy ID)

### Debugging Failed Traces

1. Look for traces with the red error icon
2. Expand to see which observation failed
3. Click to view full error details
4. Use "Open in Browser" for the complete waterfall view

### Cost Management

- Keep the status bar visible to monitor spend
- Use the cost summary command to see trends
- Filter traces by cost to find expensive operations

### Git Workflow

1. Make code changes
2. Run your agent
3. Link the trace to your current commit
4. Later, view traces grouped by commit to see how changes affected behavior

## Development

### Building from Source

```bash
cd extensions/vscode
npm install
npm run compile
```

### Running in Development

1. Open the `extensions/vscode` folder in VS Code
2. Press F5 to launch Extension Development Host
3. The extension will be loaded in the new window

### Packaging

```bash
npm run package
npx vsce package
```

## Troubleshooting

### Extension Not Loading

- Check the Output panel (View > Output) and select "AgentTrace" from the dropdown
- Verify your API key is correct
- Ensure you have network access to the API URL

### Traces Not Appearing

- Click the refresh button in the Traces view
- Check that your Project ID is correct
- Verify traces exist in the web dashboard

### High Memory Usage

- Reduce `agenttrace.maxTraces` setting
- Disable `agenttrace.autoRefresh` if not needed
- Increase `agenttrace.refreshInterval`

## Feedback & Issues

- [Report bugs](https://github.com/agenttrace/agenttrace/issues)
- [Request features](https://github.com/agenttrace/agenttrace/discussions)

## License

MIT License - see [LICENSE](../../LICENSE) for details.
