# AgentTrace JetBrains Plugin

IntelliJ-based plugin for viewing and managing AgentTrace traces directly in your IDE. Compatible with IntelliJ IDEA, PyCharm, GoLand, WebStorm, and other JetBrains IDEs.

## Features

- **Trace Explorer** — Browse traces, spans, and generations in a dedicated tool window
- **Session Management** — View and filter traces by session
- **Inline Annotations** — See trace metadata alongside your code
- **Git Linking** — Automatic correlation between traces and commits
- **Cost Overview** — View LLM costs per trace directly in the IDE

## Prerequisites

- JetBrains IDE 2023.3 or later
- AgentTrace instance running (local or remote)
- API key from AgentTrace dashboard (Settings → API Keys)

## Installation

### From JetBrains Marketplace

1. Open **Settings/Preferences → Plugins → Marketplace**
2. Search for "AgentTrace"
3. Click **Install** and restart your IDE

### From Source

```bash
cd extensions/jetbrains
./gradlew buildPlugin
```

The plugin ZIP will be in `build/distributions/`. Install via **Settings → Plugins → ⚙️ → Install Plugin from Disk**.

## Configuration

1. Open **Settings/Preferences → Tools → AgentTrace**
2. Enter your AgentTrace host URL (e.g., `http://localhost:8080`)
3. Enter your API key
4. Click **Test Connection** to verify

## Usage

### Trace Explorer

Open the **AgentTrace** tool window from the right sidebar. Browse recent traces, filter by name or session, and click to view details.

### Viewing Trace Details

Click any trace to see:
- Observation tree (spans, generations, events)
- Latency waterfall
- Input/output for each observation
- Scores and metadata
- Cost breakdown

## Development

```bash
# Build
./gradlew buildPlugin

# Run in sandbox IDE
./gradlew runIde

# Run tests
./gradlew test
```

## License

MIT License — see [LICENSE](../../LICENSE) for details.
