---
sidebar_position: 5
title: "Template Variables"
description: "Use {{variable}} syntax in prompts for dynamic content. Type-safe compilation with SDK integration."
---

# Template Variables

AgentTrace prompts support template variables using `{{variable_name}}` syntax. Variables let you create reusable prompt templates that are compiled with dynamic values at runtime.

## Syntax

Variables use double curly braces:

```
Hello {{user_name}}! Please review the following {{language}} code:

{{code}}

Focus on: {{review_focus}}
```

### Naming Rules

- Use `snake_case` for variable names: `user_name`, `code_snippet`
- Names must start with a letter and contain only letters, numbers, and underscores
- Names are case-sensitive: `userName` and `username` are different variables

### Automatic Extraction

When you create or update a prompt, AgentTrace automatically extracts all variables:

```json
{
  "content": "Analyze {{code}} in {{language}} for {{issue_type}}",
  "variables": ["code", "language", "issue_type"]
}
```

## Compiling Prompts

Compilation substitutes variable placeholders with actual values.

### Python SDK

```python
prompt = client.get_prompt("code-review")

# Compile with keyword arguments
compiled = prompt.compile(
    language="Python",
    code="def hello(): pass",
    review_focus="security vulnerabilities"
)
print(compiled)
# "Hello! Please review the following Python code:\n\ndef hello(): pass\n\nFocus on: security vulnerabilities"
```

### TypeScript SDK

```typescript
const prompt = await getPrompt({ name: 'code-review' });

const compiled = prompt.compile({
  language: 'Python',
  code: 'def hello(): pass',
  review_focus: 'security vulnerabilities',
});
```

### Go SDK

```go
prompt, _ := agenttrace.GetPrompt(agenttrace.GetPromptOptions{
    Name: "code-review",
})

compiled := prompt.Compile(map[string]any{
    "language":     "Python",
    "code":         "def hello(): pass",
    "review_focus": "security vulnerabilities",
})
```

### Server-Side Compilation

Compile via the API without using an SDK:

```bash
curl -X POST "https://api.agenttrace.io/v1/prompts/code-review/compile" \
  -H "Authorization: Bearer at-your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "variables": {
      "language": "Python",
      "code": "def hello(): pass",
      "review_focus": "security vulnerabilities"
    },
    "label": "production"
  }'
```

## Chat Prompt Variables

For `CHAT` type prompts, variables can appear in any message:

```json
[
  {
    "role": "system",
    "content": "You are a {{role}} assistant specialized in {{domain}}."
  },
  {
    "role": "user",
    "content": "{{user_message}}"
  }
]
```

Compile chat prompts to get a message array:

```python
prompt = client.get_prompt("assistant")
messages = prompt.compile_chat(
    role="senior engineer",
    domain="code review",
    user_message="Review my PR"
)
# [
#   {"role": "system", "content": "You are a senior engineer assistant specialized in code review."},
#   {"role": "user", "content": "Review my PR"}
# ]
```

## Validation

### Missing Variables

If you compile a prompt without providing all required variables, the SDK raises an error:

```python
prompt = client.get_prompt("code-review")
# Prompt expects: language, code, review_focus

compiled = prompt.compile(language="Python")
# Raises: MissingVariableError: Missing required variables: code, review_focus
```

### Extra Variables

Extra variables that don't exist in the template are silently ignored:

```python
compiled = prompt.compile(
    language="Python",
    code="...",
    review_focus="bugs",
    unused_var="this is ignored"
)
```

## Escaping

To include literal `{{` in your prompt without it being treated as a variable, use triple braces:

```
Use the {{{template_syntax}}} format.
```

This renders as: `Use the {{template_syntax}} format.`

## Variable Best Practices

### Use Descriptive Names

```
# Good
{{user_code_snippet}}
{{target_programming_language}}

# Bad
{{c}}
{{l}}
```

### Document Expected Formats

Add descriptions to your prompt metadata to document what each variable expects:

| Variable | Expected Format | Example |
|----------|----------------|---------|
| `language` | Programming language name | `Python`, `TypeScript` |
| `code` | Source code string | `def hello(): ...` |
| `review_focus` | Comma-separated focus areas | `security, performance` |

### Validate Before Compilation

Validate variable values in your application before compiling:

```python
def review_code(code: str, language: str) -> str:
    if not code.strip():
        raise ValueError("Code cannot be empty")
    if language not in SUPPORTED_LANGUAGES:
        raise ValueError(f"Unsupported language: {language}")

    prompt = client.get_prompt("code-review")
    return prompt.compile(code=code, language=language)
```

## Tracing Variable Usage

When you compile a prompt within a traced generation, AgentTrace records the variables used. This lets you:

- Search traces by variable values
- Analyze which inputs produce poor outputs
- Build evaluation datasets from real usage

```python
with client.trace("review-agent") as trace:
    prompt = client.get_prompt("code-review")
    compiled = prompt.compile(language="Python", code=user_code)

    gen = trace.generation(
        name="review",
        model="gpt-4",
        input=compiled,
        prompt_id=prompt.id,
        prompt_version=prompt.version,
    )
```

## Related

- [Overview](./overview.md) — prompt management concepts
- [Playground](./playground.md) — test variables interactively
- [Python SDK](../sdks/python.md) — full SDK reference
- [TypeScript SDK](../sdks/typescript.md) — full SDK reference
