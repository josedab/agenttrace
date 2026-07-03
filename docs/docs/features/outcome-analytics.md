---
title: Agent Outcome Analytics
description: Project-scoped trace, git, CI, pull request, and cost outcome analytics.
---

# Agent Outcome Analytics

AgentTrace correlates recorded agent runs with linked commits and CI runs. The **Collaboration** workspace and `/analytics/outcomes` dashboard use only stored project data; missing inputs are reported as unavailable rather than estimated.

## Metrics

- completed agent-run success rate
- CI pass, failure, cancellation, and in-progress counts
- linked commits and pull requests
- failing-CI regression signals and commit-message revert signals
- total cost and cost per successful outcome
- model and `agent_name` breakdowns when attribution exists

Success means a trace has completed and is not at `ERROR` level. A regression signal means a failing CI run is linked to a recorded commit. A revert signal means a linked commit message starts with `revert`; it is an indicator, not a claim that a specific agent change was reverted.

## API

```bash
curl "$AGENTTRACE_HOST/api/public/outcomes?window=30d" \
  -H "Authorization: Bearer $AGENTTRACE_API_KEY"
```

Supported windows are `24h`, `7d`, `30d`, and `90d`. Explicit RFC3339 `from` and `to` values are also accepted.

Generate reusable report content:

```bash
curl "$AGENTTRACE_HOST/api/public/outcomes/digest?window=7d&format=markdown" \
  -H "Authorization: Bearer $AGENTTRACE_API_KEY"
```

The same digest renderer is used by optional GitHub, Slack, Discord, and generic webhook delivery.

## Optional delivery

GitHub reporting is disabled unless `GITHUB_REPORTING_ENABLED=true` and `GITHUB_REPORT_TOKEN` are configured. Webhook delivery uses project-owned webhook IDs and permits only public HTTPS destinations. Neither integration is required for outcome analytics.
