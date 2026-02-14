---
sidebar_position: 4
title: "Prompt Playground"
description: "Interactive prompt playground for testing prompts with different variables and models, with side-by-side comparison."
---

# Prompt Playground

The AgentTrace Prompt Playground is an interactive environment for testing, iterating, and comparing prompts before deploying them to production.

## Overview

The playground lets you:

- **Test prompts** with different variable values in real time
- **Compare outputs** side-by-side across models and prompt versions
- **Iterate quickly** without deploying code changes
- **Save results** as evaluation datasets for regression testing

## Accessing the Playground

1. Navigate to **Prompts** in the dashboard
2. Select a prompt and click **Open in Playground**
3. Or go directly to **Playground** in the sidebar to start from scratch

## Interface

The playground is divided into three panels:

### Prompt Editor (Left)

Write or load a prompt template. Variables are highlighted automatically:

```
You are a {{role}} assistant.

Analyze the following {{language}} code and provide:
1. A summary of what the code does
2. Any potential bugs
3. Suggestions for improvement

Code:
{{code}}
```

Detected variables appear as input fields below the editor.

### Variables Panel (Center)

Fill in variable values for testing:

| Variable | Value |
|----------|-------|
| `role` | `senior code reviewer` |
| `language` | `Python` |
| `code` | `def fib(n): return fib(n-1) + fib(n-2)` |

You can save variable sets as **presets** for repeated testing.

### Output Panel (Right)

View the compiled prompt and LLM response. The output shows:

- **Compiled prompt** — the final text sent to the model
- **Model response** — the LLM output
- **Metadata** — token usage, latency, and cost

## Comparing Outputs

### Side-by-Side Comparison

Click **Add Comparison** to run the same prompt against multiple configurations:

| Configuration | Model | Version | Result |
|---------------|-------|---------|--------|
| Baseline | GPT-4 | v2 (production) | ... |
| Candidate A | GPT-4 | v3 (staging) | ... |
| Candidate B | GPT-4o | v3 (staging) | ... |

This is especially useful for:

- **Model comparison** — test the same prompt across different LLMs
- **Version comparison** — compare outputs between prompt versions
- **Variable sensitivity** — see how different inputs affect output quality

### Diff View

Toggle diff view to highlight differences between two outputs, making it easy to spot regressions or improvements.

## Model Configuration

Configure the model settings for each test run:

| Setting | Description | Default |
|---------|-------------|---------|
| Model | LLM model to use | `gpt-4` |
| Temperature | Randomness (0-2) | `1.0` |
| Max Tokens | Maximum output length | `1024` |
| Top P | Nucleus sampling | `1.0` |
| Stop Sequences | Tokens that stop generation | `[]` |

### Supported Models

The playground supports any model available through your configured API keys:

- **OpenAI**: GPT-4, GPT-4o, GPT-3.5 Turbo
- **Anthropic**: Claude 3.5 Sonnet, Claude 3 Opus, Claude 3 Haiku
- **Google**: Gemini Pro, Gemini Ultra
- **Custom**: Any OpenAI-compatible endpoint

## Saving Results

### Save as Version

After iterating in the playground, save the prompt as a new version:

1. Click **Save as New Version**
2. Add a commit message
3. Optionally assign a label

### Export to Dataset

Save playground inputs and outputs as evaluation datasets:

1. Click **Export to Dataset**
2. Select or create a target dataset
3. The variable values become inputs, LLM responses become expected outputs

This creates a regression test you can run automatically on future prompt changes.

## Chat Mode

For `CHAT` type prompts, the playground provides a conversation interface:

1. Define the system message and initial user message in the template
2. Continue the conversation interactively
3. Each turn is recorded and can be exported

```json
[
  {"role": "system", "content": "You are a {{role}} assistant."},
  {"role": "user", "content": "{{initial_question}}"}
]
```

## Keyboard Shortcuts

| Shortcut | Action |
|----------|--------|
| `Ctrl/Cmd + Enter` | Run prompt |
| `Ctrl/Cmd + S` | Save as new version |
| `Ctrl/Cmd + D` | Toggle diff view |
| `Ctrl/Cmd + Shift + C` | Add comparison column |

## Best Practices

1. **Start with production** — load the current production version as your baseline
2. **Test edge cases** — use variable presets for common edge cases
3. **Compare before promoting** — always compare against the current production version
4. **Save datasets** — export interesting test cases for automated evaluation
5. **Document findings** — use commit messages when saving versions

## Related

- [Overview](./overview.md) — prompt management concepts
- [Variables](./variables.md) — template variable syntax
- [Versioning](./versioning.md) — version management
