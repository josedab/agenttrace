---
slug: /
sidebar_position: 1
---

# Introduction

AgentTrace is a comprehensive observability platform designed specifically for AI coding agents. It provides deep insights into how AI agents interact with codebases, make decisions, and execute tasks.

## Why AgentTrace?

Modern AI coding agents like Claude Code, Cursor, and GitHub Copilot Workspace are revolutionizing software development. However, understanding what these agents are doing, why they make certain decisions, and how to improve their performance is challenging without proper observability.

AgentTrace solves this by providing:

- **Complete Trace Visibility**: See every step your AI agent takes, from initial prompt to final code change
- **Cost & Performance Analytics**: Track token usage, latency, and costs across all your AI operations
- **Prompt Management**: Version, test, and optimize your prompts in one place
- **Evaluation Framework**: Score agent outputs with LLM-as-judge and human annotation
- **Git Integration**: Automatically link traces to commits and pull requests
- **CI/CD Integration**: Track agent performance across your development pipeline

## Key Features

### For Developers

- **Real-time Tracing**: Watch your AI agents work in real-time with detailed execution traces
- **IDE Extensions**: VS Code and JetBrains plugins for in-editor observability
- **CLI Wrapper**: Wrap any CLI tool to capture AI agent activity

### For Teams

- **Prompt Library**: Centralized prompt management with versioning and A/B testing
- **Datasets & Experiments**: Test prompts against curated datasets
- **Collaboration**: Share traces, prompts, and insights across your team

### For Organizations

- **Enterprise SSO**: SAML 2.0 and OIDC support for seamless authentication
- **Audit Logging**: Comprehensive audit trails for compliance
- **Role-Based Access**: Fine-grained permissions for teams and projects

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                         Your AI Agent                           │
│            (Claude Code, Cursor, Custom Agent, etc.)            │
└─────────────────────────────────────────────────────────────────┘
                              │
                   AgentTrace SDK / CLI
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                    AgentTrace Platform                          │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────────┐ │
│  │   Tracing   │  │   Prompts   │  │     Evaluation          │ │
│  │   Engine    │  │   Manager   │  │     Framework           │ │
│  └─────────────┘  └─────────────┘  └─────────────────────────┘ │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────────┐ │
│  │   Cost      │  │   Dataset   │  │     CI/CD               │ │
│  │   Tracking  │  │   Runner    │  │     Integration         │ │
│  └─────────────┘  └─────────────┘  └─────────────────────────┘ │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                    AgentTrace Dashboard                         │
│       Traces │ Sessions │ Prompts │ Datasets │ Analytics       │
└─────────────────────────────────────────────────────────────────┘
```

## Langfuse Compatibility

AgentTrace is designed to be compatible with the Langfuse API, making migration seamless. If you're already using Langfuse, you can:

1. Point your existing SDKs to AgentTrace with a simple configuration change
2. Use the same API endpoints and data formats
3. Migrate historical data using our import tools

## Getting Started

Ready to start? Check out our [Quickstart Guide](/getting-started/quickstart) to instrument your first AI agent in under 5 minutes.

## Community

- [GitHub](https://github.com/agenttrace/agenttrace) - Star us and contribute
- [Discord](https://discord.gg/agenttrace) - Join our community
- [Twitter](https://twitter.com/agenttrace) - Follow for updates
