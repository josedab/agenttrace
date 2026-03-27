---
sidebar_position: 4
title: VS Code Extension
description: View and manage AgentTrace sessions directly in Visual Studio Code with inline trace visualization.
---

# VS Code Extension

The AgentTrace VS Code extension brings inline trace viewing, session management, and cost monitoring directly into your editor.

## Installation

1. Open VS Code
2. Go to **Extensions** (`Ctrl+Shift+X` / `Cmd+Shift+X`)
3. Search for **"AgentTrace"**
4. Click **Install**

Or install from the command line:

```bash
code --install-extension agenttrace.agenttrace-vscode
```

## Quick Start

1. Open the command palette (`Ctrl+Shift+P` / `Cmd+Shift+P`)
2. Run **AgentTrace: Sign In**
3. Enter your API key (starts with `sk-at-`)
4. Select your project from the dropdown

The Trace Explorer sidebar will appear in the Activity Bar.

## Features

### Trace Explorer Sidebar

Browse traces directly from the Activity Bar:

- View recent traces sorted by time
- Filter by name, status, or date range
- Click a trace to see its full span tree
- Expand individual spans to view inputs, outputs, and token counts

### Session Management

Group and navigate traces by session:

- View active and recent sessions
- See all traces within a session
- Track session duration and total cost
- Jump between sessions quickly

### Git Linking

Traces are automatically linked to your Git context:

- Current branch and commit SHA are attached to new traces
- View which traces were generated from a specific commit
- Navigate from a trace back to the relevant code change

### Cost Monitoring

Monitor LLM spending in real time:

- Status bar widget showing current session cost
- Per-trace cost breakdown in the Trace Explorer
- Daily and weekly cost summaries in the sidebar
- Configurable cost threshold warnings

## Configuration

Configure the extension in **Settings** (`Ctrl+,` / `Cmd+,`) under **AgentTrace**:

| Setting | Default | Description |
|---------|---------|-------------|
| `agenttrace.apiKey` | — | API key for authentication |
| `agenttrace.projectId` | — | Default project ID |
| `agenttrace.apiUrl` | Cloud URL | API URL for self-hosted instances |
| `agenttrace.costWarningThreshold` | `10.00` | Cost warning threshold in USD |
| `agenttrace.autoLinkGit` | `true` | Auto-link traces to Git context |

## Commands

Open the command palette and type **AgentTrace** to see all available commands:

| Command | Description |
|---------|-------------|
| `AgentTrace: Sign In` | Authenticate with your API key |
| `AgentTrace: Select Project` | Switch active project |
| `AgentTrace: Open Trace` | Open a trace by ID |
| `AgentTrace: Show Session` | View traces for a session |
| `AgentTrace: Show Costs` | Open the cost summary panel |

## Troubleshooting

- **Extension not activating**: Ensure you are signed in and have a valid API key.
- **No traces appearing**: Check that `projectId` is set and the API URL is reachable.
- **Git linking not working**: Verify the workspace is a Git repository with a remote configured.

## Advanced Features

### Keyboard Shortcuts

| Shortcut (Mac / Windows-Linux) | Action |
|-------------------------------|--------|
| `Cmd+Shift+T` / `Ctrl+Shift+T` | Open Trace Explorer |
| `Cmd+Shift+L` / `Ctrl+Shift+L` | Open latest trace |
| `Cmd+Shift+C` / `Ctrl+Shift+C` | Copy current trace ID |

### Inline Annotations

When enabled, the extension shows gutter decorations on instrumented functions:

- **Green dot**: Last trace succeeded
- **Yellow dot**: Last trace had warnings
- **Red dot**: Last trace had errors

Hover over the gutter icon to see trace name, latency, cost, and a link to open the full trace.

Enable or disable inline annotations in settings:

| Setting | Default | Description |
|---------|---------|-------------|
| `agenttrace.showAnnotations` | `true` | Display inline gutter annotations |
| `agenttrace.annotationDetail` | `compact` | Annotation detail level (`compact` or `full`) |

### Trace Diff View

Compare two traces side-by-side to identify regressions or improvements:

1. Open the Trace Explorer sidebar
2. Right-click a trace and select **Compare With…**
3. Select the second trace from the list
4. The diff view highlights differences in span trees, latencies, and costs

### Multi-Root Workspace Support

The extension works with VS Code multi-root workspaces. Each workspace folder can be associated with a different project:

1. Open **Settings** > **Workspace** > **AgentTrace**
2. Configure `agenttrace.projectId` per folder in your `.code-workspace` file

### Status Bar Integration

The status bar widget at the bottom of the editor shows:

- **Session cost**: Running total for the current session
- **Active traces**: Number of in-progress traces
- **Connection status**: Green (connected), yellow (reconnecting), or red (disconnected)

Click the status bar widget to open the cost summary panel.
