---
sidebar_position: 5
title: JetBrains Plugin
description: View and manage AgentTrace sessions in IntelliJ IDEA, PyCharm, GoLand, and other JetBrains IDEs.
---

# JetBrains Plugin

The AgentTrace plugin for JetBrains IDEs (IntelliJ IDEA, PyCharm, GoLand, WebStorm, and others) provides a Trace Explorer tool window, session management, and inline annotations.

## Installation

### From JetBrains Marketplace

1. Open **Settings** (`Ctrl+Alt+S` / `Cmd+,`)
2. Go to **Plugins** > **Marketplace**
3. Search for **"AgentTrace"**
4. Click **Install** and restart the IDE

### Manual Installation

1. Download the plugin `.zip` from [GitHub Releases](https://github.com/agenttrace/agenttrace/releases)
2. Go to **Settings** > **Plugins** > **⚙️** > **Install Plugin from Disk…**
3. Select the downloaded file and restart

## Quick Start

1. Open **Settings** > **Tools** > **AgentTrace**
2. Enter your API key (starts with `sk-at-`)
3. Select your project
4. Open the **AgentTrace** tool window from **View** > **Tool Windows** > **AgentTrace**

## Features

### Trace Explorer Tool Window

Browse and inspect traces without leaving your IDE:

- View recent traces with status indicators
- Expand a trace to see its span tree
- Inspect inputs, outputs, and metadata for each span
- Filter by name, status, model, or date range
- Double-click a trace to open the detail view

### Session Management

Organize traces by session:

- List active and recent sessions
- View all traces within a session
- Track session-level cost and duration
- Navigate between sessions with keyboard shortcuts

### Inline Annotations

See trace context directly in your code:

- Gutter icons on instrumented functions
- Hover to see the latest trace name, latency, and cost
- Click to jump to the full trace in the tool window
- Annotations update automatically as new traces arrive

## Configuration

Configure the plugin in **Settings** > **Tools** > **AgentTrace**:

| Setting | Default | Description |
|---------|---------|-------------|
| API Key | — | AgentTrace API key |
| Project ID | — | Default project ID |
| API URL | Cloud URL | Custom URL for self-hosted instances |
| Show Annotations | `true` | Display inline gutter annotations |
| Auto-Refresh Interval | `30s` | Trace list refresh interval |

## Supported IDEs

The plugin is compatible with all IntelliJ Platform 2023.1+ IDEs:

| IDE | Supported |
|-----|-----------|
| IntelliJ IDEA (Community & Ultimate) | ✅ |
| PyCharm (Community & Professional) | ✅ |
| GoLand | ✅ |
| WebStorm | ✅ |
| CLion | ✅ |
| Rider | ✅ |

## Troubleshooting

- **Tool window not visible**: Go to **View** > **Tool Windows** > **AgentTrace**. If missing, verify the plugin is enabled under **Settings** > **Plugins** > **Installed**.
- **Connection errors**: Check your API key and API URL in **Settings** > **Tools** > **AgentTrace**. Ensure the URL is reachable from your network.
- **Annotations not showing**: Confirm **Show Annotations** is enabled in settings. Annotations require at least one trace linked to the current project.
