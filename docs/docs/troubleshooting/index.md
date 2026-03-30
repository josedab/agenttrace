---
sidebar_position: 50
---

# Troubleshooting

Common issues and solutions when using AgentTrace, organized by component.

- **[Connection Issues](./connection.md)** — Traces not appearing, connection refused, SSL errors
- **[Authentication](./authentication.md)** — 401/403 errors, API key issues
- **[SDK Issues](./sdk.md)** — Python, TypeScript, and Go SDK problems
- **[Integrations](./integrations.md)** — OpenAI, LangChain, Anthropic integration issues
- **[Data Issues](./data.md)** — Missing tokens, incorrect costs, truncated payloads
- **[Performance](./performance.md)** — High latency, memory usage
- **[Deployment](./deployment.md)** — Docker, ClickHouse, PostgreSQL issues
- **[UI Issues](./ui.md)** — Dashboard and graph rendering problems

## Getting Help

If you're still stuck:

1. **Check logs**:
   ```bash
   # API logs
   docker compose logs api

   # Worker logs
   docker compose logs worker

   # All logs
   docker compose logs -f
   ```

2. **Enable debug mode**:
   ```bash
   export AGENTTRACE_DEBUG=true
   ```

3. **Search existing issues**: [GitHub Issues](https://github.com/agenttrace/agenttrace/issues)

4. **Join the community**: [Discord](https://discord.gg/agenttrace)

5. **File a bug report** with:
   - AgentTrace version
   - SDK version
   - Reproduction steps
   - Error messages and logs
