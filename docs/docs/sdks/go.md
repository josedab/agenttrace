---
sidebar_position: 4
---

# Go SDK

The AgentTrace Go SDK provides observability for Go-based AI agents with idiomatic Go patterns.

## Installation

```bash
go get github.com/agenttrace/agenttrace-go
```

**Requirements**: Go 1.21+

## Quick Start

```go
package main

import (
    "context"

    agenttrace "github.com/agenttrace/agenttrace-go"
)

func main() {
    // Initialize the client
    client := agenttrace.New(agenttrace.Config{
        APIKey: "sk-at-...",
        Host:   "https://api.agenttrace.io",
    })
    defer client.Shutdown()

    ctx := context.Background()

    // Create a trace
    trace := client.Trace(ctx, agenttrace.TraceOptions{
        Name: "my-agent",
    })

    // Create a generation (LLM call)
    gen := trace.Generation(agenttrace.GenerationOptions{
        Name:  "llm-call",
        Model: "gpt-4",
        Input: map[string]any{"query": "Hello"},
    })

    // ... make LLM call ...

    gen.End(&agenttrace.GenerationEndOptions{
        Output: "Hi there!",
        Usage: &agenttrace.UsageDetails{
            InputTokens:  10,
            OutputTokens: 5,
        },
    })

    trace.End(nil)

    // Flush ensures events are sent
    client.Flush()
}
```

## Configuration

### Config Options

```go
client := agenttrace.New(agenttrace.Config{
    // APIKey is required for authentication
    APIKey: "sk-at-...",

    // Host is the API URL (default: "https://api.agenttrace.io")
    Host: "https://api.agenttrace.io",

    // PublicKey for client-side usage (optional)
    PublicKey: "pk-...",

    // ProjectID override (optional)
    ProjectID: "your-project-id",

    // Enabled controls tracing (default: true)
    Enabled: ptr(true),

    // FlushAt is events per batch (default: 20)
    FlushAt: 20,

    // FlushInterval is duration between flushes (default: 5s)
    FlushInterval: 5 * time.Second,

    // MaxRetries for failed requests (default: 3)
    MaxRetries: 3,

    // Timeout for requests (default: 10s)
    Timeout: 10 * time.Second,
})

// Helper for boolean pointers
func ptr(b bool) *bool { return &b }
```

### Environment Variables

```bash
export AGENTTRACE_API_KEY="sk-at-..."
export AGENTTRACE_HOST="https://api.agenttrace.io"
export AGENTTRACE_PROJECT_ID="your-project-id"
```

## Core Concepts

### Traces

Traces represent complete execution paths:

```go
trace := client.Trace(ctx, agenttrace.TraceOptions{
    Name:      "code-review-agent",
    ID:        "custom-id",           // Optional: auto-generated if empty
    UserID:    "user-123",
    SessionID: "session-456",
    Metadata: map[string]any{
        "version": "1.0",
    },
    Tags:   []string{"production", "code-review"},
    Input:  map[string]any{"code": "function add(a, b) { return a + b; }"},
    Public: false,
})

// Do work...

trace.End(&agenttrace.TraceEndOptions{
    Output: map[string]any{"result": "Code looks good!"},
})
```

### Spans

Spans represent non-LLM operations:

```go
span := trace.Span(agenttrace.SpanOptions{
    Name: "parse-document",
    ParentObservationID: "", // Optional: for nested spans
    Metadata: map[string]any{
        "documentId": "doc-123",
    },
    Input: document,
    Level: "DEFAULT", // DEBUG, DEFAULT, WARNING, ERROR
})

// Do work...
result := parseDocument(document)

span.End(&agenttrace.SpanEndOptions{
    Output: result,
})
```

### Generations

Generations track LLM calls:

```go
gen := trace.Generation(agenttrace.GenerationOptions{
    Name:  "gpt-4-completion",
    Model: "gpt-4",
    ModelParameters: map[string]any{
        "temperature": 0.7,
        "max_tokens": 1000,
    },
    Input: messages,
    Level: "DEFAULT",
})

// Make LLM call
response, err := openai.ChatCompletion(messages)

gen.End(&agenttrace.GenerationEndOptions{
    Output: response.Content,
    Model:  "gpt-4-0125-preview", // Actual model used
    Usage: &agenttrace.UsageDetails{
        InputTokens:  response.Usage.PromptTokens,
        OutputTokens: response.Usage.CompletionTokens,
        TotalTokens:  response.Usage.TotalTokens,
    },
})
```

## Context-Based Tracing

### Context Functions

The SDK provides context-aware tracing functions:

```go
import agenttrace "github.com/agenttrace/agenttrace-go"

func main() {
    ctx := context.Background()

    // Start a trace and get context
    trace, ctx := agenttrace.StartTrace(ctx, agenttrace.TraceOptions{
        Name: "my-workflow",
    })
    defer trace.End(nil)

    // Start a span within the trace
    span, ctx := agenttrace.StartSpan(ctx, agenttrace.SpanOptions{
        Name: "step-1",
    })
    defer span.End(nil)

    // Start a generation within the trace
    gen, ctx := agenttrace.StartGeneration(ctx, agenttrace.GenerationOptions{
        Name:  "llm-call",
        Model: "gpt-4",
    })
    defer gen.End(nil)
}
```

### Context Retrieval

```go
// Get current trace from context
trace := agenttrace.GetCurrentTrace(ctx)
if trace != nil {
    fmt.Println("Trace ID:", trace.ID())
}

// Get current span from context
span := agenttrace.GetCurrentSpan(ctx)

// Get current generation from context
gen := agenttrace.GetCurrentGeneration(ctx)

// Get global client
client := agenttrace.GetGlobalClient()
```

### Context Helpers

```go
// Add trace to context
ctx = agenttrace.WithTrace(ctx, trace)

// Add span to context
ctx = agenttrace.WithSpan(ctx, span)

// Add generation to context
ctx = agenttrace.WithGeneration(ctx, gen)
```

## Nested Traces

Create hierarchical traces for complex workflows:

```go
trace := client.Trace(ctx, agenttrace.TraceOptions{
    Name: "document-processor",
})

// First step
extractSpan := trace.Span(agenttrace.SpanOptions{
    Name: "extract-text",
})
text := extractText(document)
extractSpan.End(&agenttrace.SpanEndOptions{Output: text})

// Nested operations
analyzeSpan := trace.Span(agenttrace.SpanOptions{
    Name: "analyze-content",
})

// Nested generation under analyze span
sentimentGen := trace.Generation(agenttrace.GenerationOptions{
    Name:                "sentiment-analysis",
    Model:               "gpt-4",
    ParentObservationID: analyzeSpan.ID(), // Link to parent
    Input:               map[string]any{"text": text},
})
sentiment := analyzeSentiment(text)
sentimentGen.End(&agenttrace.GenerationEndOptions{Output: sentiment})

analyzeSpan.End(&agenttrace.SpanEndOptions{
    Output: map[string]any{"sentiment": sentiment},
})

trace.End(&agenttrace.TraceEndOptions{
    Output: map[string]any{"text": text, "sentiment": sentiment},
})
```

## Agent Features

### Checkpoints

Create state snapshots during agent execution:

```go
trace := client.Trace(ctx, agenttrace.TraceOptions{
    Name: "code-editor-agent",
})

// Checkpoint before making changes
beforeCp := trace.Checkpoint(agenttrace.CheckpointOptions{
    Name:           "before-edit",
    Type:           agenttrace.CheckpointTypeManual,
    Description:    "State before code modification",
    Files:          []string{"src/main.go", "src/utils.go"},
    IncludeGitInfo: true,
})

fmt.Println("Checkpoint ID:", beforeCp.ID)
fmt.Println("Git Commit:", beforeCp.GitCommitSha)

// Make changes
editFiles()

// Checkpoint after changes
afterCp := trace.Checkpoint(agenttrace.CheckpointOptions{
    Name:  "after-edit",
    Type:  agenttrace.CheckpointTypeMilestone,
    Files: []string{"src/main.go", "src/utils.go"},
})

trace.End(nil)
```

#### Checkpoint Types

```go
const (
    CheckpointTypeManual    CheckpointType = "manual"     // User-initiated
    CheckpointTypeAuto      CheckpointType = "auto"       // Automatic
    CheckpointTypeToolCall  CheckpointType = "tool_call"  // Before/after tool
    CheckpointTypeError     CheckpointType = "error"      // On error
    CheckpointTypeMilestone CheckpointType = "milestone"  // Progress point
    CheckpointTypeRestore   CheckpointType = "restore"    // State restoration
)
```

### Git Linking

Associate traces with git commits:

```go
trace := client.Trace(ctx, agenttrace.TraceOptions{
    Name: "feature-implementation",
})

// Auto-detect git info (default)
link := trace.GitLink(nil)

// Or with explicit values
link := trace.GitLink(&agenttrace.GitLinkOptions{
    Type:          agenttrace.GitLinkTypeCommit,
    CommitSha:     "abc123def456",
    Branch:        "feature/new-api",
    RepoURL:       "https://github.com/org/repo",
    CommitMessage: "Add new API endpoint",
    FilesChanged:  []string{"api.go", "routes.go"},
    AutoDetect:    false, // Disable auto-detection
})

fmt.Println("Link ID:", link.ID)
fmt.Println("Author:", link.AuthorName, link.AuthorEmail)
```

#### Git Link Types

```go
const (
    GitLinkTypeStart   GitLinkType = "start"   // Beginning of work
    GitLinkTypeCommit  GitLinkType = "commit"  // Specific commit
    GitLinkTypeRestore GitLinkType = "restore" // Restored state
    GitLinkTypeBranch  GitLinkType = "branch"  // Branch switch
    GitLinkTypeDiff    GitLinkType = "diff"    // Current diff
)
```

### File Operations

Track file read/write operations:

```go
// Track a file update
op := trace.FileOp(agenttrace.FileOperationOptions{
    Operation:     agenttrace.FileOpUpdate,
    FilePath:      "src/main.go",
    ContentBefore: oldContent,
    ContentAfter:  newContent,
    LinesAdded:    intPtr(15),
    LinesRemoved:  intPtr(8),
    ToolName:      "edit-file",
    Reason:        "Refactoring for clarity",
})

// Track file creation
trace.FileOp(agenttrace.FileOperationOptions{
    Operation:    agenttrace.FileOpCreate,
    FilePath:     "src/utils/helpers.go",
    ContentAfter: newFileContent,
})

// Track file deletion
trace.FileOp(agenttrace.FileOperationOptions{
    Operation: agenttrace.FileOpDelete,
    FilePath:  "src/old_file.go",
})

// Helper for int pointers
func intPtr(i int) *int { return &i }
```

#### Operation Types

```go
const (
    FileOpCreate FileOperationType = "create" // New file
    FileOpRead   FileOperationType = "read"   // File read
    FileOpUpdate FileOperationType = "update" // File modified
    FileOpDelete FileOperationType = "delete" // File deleted
    FileOpRename FileOperationType = "rename" // File renamed
    FileOpCopy   FileOperationType = "copy"   // File copied
    FileOpMove   FileOperationType = "move"   // File moved
    FileOpChmod  FileOperationType = "chmod"  // Permissions changed
)
```

### Terminal Commands

Track shell command execution:

```go
// Manual tracking
cmd := trace.TerminalCmd(agenttrace.TerminalCommandOptions{
    Command:          "npm",
    Args:             []string{"test"},
    ExitCode:         0,
    Stdout:           "All tests passed (42 tests)",
    Stderr:           "",
    WorkingDirectory: "/project",
    ToolName:         "run-tests",
    Reason:           "Verify changes before commit",
})

// Run and track automatically
result := trace.RunCmd(ctx, "npm", &agenttrace.RunCommandOptions{
    Args:             []string{"run", "build"},
    WorkingDirectory: "/project",
    Timeout:          60 * time.Second,
    MaxOutputBytes:   100000,
    Env: map[string]string{
        "NODE_ENV": "production",
    },
})

fmt.Println("Exit code:", result.ExitCode)
fmt.Println("Stdout:", result.Stdout)
fmt.Println("Command ID:", result.Info.ID)
```

## Prompts

### Fetching Prompts

```go
// Get latest version
prompt, err := agenttrace.GetPrompt(agenttrace.GetPromptOptions{
    Name: "code-review",
})

// Get specific version
version := 2
prompt, err := agenttrace.GetPrompt(agenttrace.GetPromptOptions{
    Name:    "code-review",
    Version: &version,
})

// Get by label
prompt, err := agenttrace.GetPrompt(agenttrace.GetPromptOptions{
    Name:  "code-review",
    Label: "production",
})

// With fallback
prompt, err := agenttrace.GetPrompt(agenttrace.GetPromptOptions{
    Name:     "code-review",
    Fallback: "Review the following code for issues:\n{{code}}",
})
```

### Compiling Prompts

```go
prompt, _ := agenttrace.GetPrompt(agenttrace.GetPromptOptions{
    Name: "code-review",
})

// Compile with variables
compiled := prompt.Compile(map[string]any{
    "language": "Go",
    "code":     "func add(a, b int) int { return a + b }",
})

// Get variable names
variables := prompt.GetVariables() // []string{"language", "code"}
```

### Chat Prompts

```go
prompt, _ := agenttrace.GetPrompt(agenttrace.GetPromptOptions{
    Name: "chat-template",
})

// Prompt content:
// system: You are a helpful {{role}}.
// user: {{question}}

messages := prompt.CompileChat(map[string]any{
    "role":     "coding assistant",
    "question": "How do I sort a slice?",
})

// Result:
// []ChatMessage{
//     {Role: "system", Content: "You are a helpful coding assistant."},
//     {Role: "user", Content: "How do I sort a slice?"},
// }
```

### Cache Management

```go
// Set cache TTL (default: 1 minute)
agenttrace.SetPromptCacheTTL(2 * time.Minute)

// Clear all cached prompts
agenttrace.ClearPromptCache()

// Invalidate specific prompt
agenttrace.InvalidatePrompt("code-review")
```

## Scoring

### Score a Trace

```go
// Score by trace ID
client.Score(agenttrace.ScoreOptions{
    TraceID:  "trace-123",
    Name:     "quality",
    Value:    0.95,
    DataType: "NUMERIC",
    Comment:  "Excellent response quality",
})

// Score within a trace
trace.Score("accuracy", 0.92, &agenttrace.ScoreAddOptions{
    DataType: "NUMERIC",
    Comment:  "High accuracy",
})
```

### Score Types

```go
// Numeric score
client.Score(agenttrace.ScoreOptions{
    TraceID:  traceID,
    Name:     "quality",
    Value:    0.95,
    DataType: "NUMERIC",
})

// Boolean score
client.Score(agenttrace.ScoreOptions{
    TraceID:  traceID,
    Name:     "correct",
    Value:    true,
    DataType: "BOOLEAN",
})

// Categorical score
client.Score(agenttrace.ScoreOptions{
    TraceID:  traceID,
    Name:     "rating",
    Value:    "excellent",
    DataType: "CATEGORICAL",
})
```

## HTTP Middleware

### Standard Library

```go
import (
    agenttrace "github.com/agenttrace/agenttrace-go"
    "github.com/agenttrace/agenttrace-go/middleware"
)

func main() {
    client := agenttrace.New(agenttrace.Config{
        APIKey: "sk-at-...",
    })
    defer client.Shutdown()

    // Create middleware
    traceMiddleware := middleware.HTTP(&middleware.HTTPMiddlewareConfig{
        TraceName: func(r *http.Request) string {
            return r.Method + " " + r.URL.Path
        },
        SkipPaths: []string{"/health", "/metrics"},
        ExtractUserID: func(r *http.Request) string {
            return r.Header.Get("X-User-ID")
        },
        ExtractSessionID: func(r *http.Request) string {
            return r.Header.Get("X-Session-ID")
        },
    })

    // Apply middleware
    http.Handle("/api/", traceMiddleware(myHandler))
    http.ListenAndServe(":8080", nil)
}
```

### With Router (Chi, Gorilla, etc.)

```go
import (
    "github.com/go-chi/chi/v5"
    "github.com/agenttrace/agenttrace-go/middleware"
)

func main() {
    r := chi.NewRouter()

    // Apply tracing middleware
    r.Use(middleware.HTTP(nil)) // Use defaults

    r.Get("/api/users", listUsers)
    r.Post("/api/users", createUser)

    http.ListenAndServe(":8080", r)
}
```

## Update Traces

Modify traces after creation:

```go
trace := client.Trace(ctx, agenttrace.TraceOptions{
    Name: "my-task",
})

// Update properties
newName := "updated-name"
trace.Update(agenttrace.TraceUpdateOptions{
    Name: &newName,
    Metadata: map[string]any{
        "step": 2,
    },
    Tags:   []string{"updated", "important"},
    Output: map[string]any{"partial": "result"},
})

// Continue work...

trace.End(&agenttrace.TraceEndOptions{
    Output: map[string]any{"final": "result"},
})
```

## Flushing and Shutdown

```go
// Flush pending events immediately
client.Flush()

// Shutdown with flush (recommended)
defer client.Shutdown()

// The client also auto-flushes:
// - Every FlushInterval (default: 5 seconds)
// - When FlushAt events accumulate (default: 20)
```

## Disabled Mode

Disable tracing for testing:

```go
enabled := false
client := agenttrace.New(agenttrace.Config{
    APIKey:  "sk-at-...",
    Enabled: &enabled,
})

// Check if enabled
if client.Enabled() {
    fmt.Println("Tracing is enabled")
}
```

## Error Handling

```go
trace := client.Trace(ctx, agenttrace.TraceOptions{
    Name: "risky-task",
})

result, err := riskyOperation()
if err != nil {
    trace.Update(agenttrace.TraceUpdateOptions{
        Output: map[string]any{
            "error": err.Error(),
        },
        Metadata: map[string]any{
            "errorType": fmt.Sprintf("%T", err),
        },
    })
    trace.End(nil)
    return err
}

trace.End(&agenttrace.TraceEndOptions{
    Output: result,
})
```

## Go Types Reference

### Config

```go
type Config struct {
    APIKey        string
    Host          string
    PublicKey     string
    ProjectID     string
    Enabled       *bool
    FlushAt       int
    FlushInterval time.Duration
    MaxRetries    int
    Timeout       time.Duration
}
```

### TraceOptions

```go
type TraceOptions struct {
    Name      string
    ID        string
    UserID    string
    SessionID string
    Metadata  map[string]any
    Tags      []string
    Input     any
    Public    bool
}
```

### SpanOptions

```go
type SpanOptions struct {
    Name                string
    ID                  string
    ParentObservationID string
    Metadata            map[string]any
    Input               any
    Level               string // DEBUG, DEFAULT, WARNING, ERROR
}
```

### GenerationOptions

```go
type GenerationOptions struct {
    Name                string
    ID                  string
    ParentObservationID string
    Model               string
    ModelParameters     map[string]any
    Input               any
    Metadata            map[string]any
    Level               string
}

type UsageDetails struct {
    InputTokens  int
    OutputTokens int
    TotalTokens  int
}
```

### Agent Feature Types

```go
// Checkpoints
type CheckpointOptions struct {
    Name           string
    Type           CheckpointType
    ObservationID  string
    Description    string
    Files          []string
    IncludeGitInfo bool
}

type CheckpointInfo struct {
    ID             string
    Name           string
    Type           CheckpointType
    TraceID        string
    ObservationID  string
    GitCommitSha   string
    GitBranch      string
    FilesChanged   []string
    TotalFiles     int
    TotalSizeBytes int64
    CreatedAt      time.Time
}

// Git Links
type GitLinkOptions struct {
    Type          GitLinkType
    ObservationID string
    CommitSha     string
    Branch        string
    RepoURL       string
    CommitMessage string
    FilesChanged  []string
    AutoDetect    bool
}

type GitLinkInfo struct {
    ID            string
    TraceID       string
    ObservationID string
    Type          GitLinkType
    CommitSha     string
    Branch        string
    RepoURL       string
    CommitMessage string
    AuthorName    string
    AuthorEmail   string
    FilesChanged  []string
    CreatedAt     time.Time
}

// File Operations
type FileOperationOptions struct {
    Operation     FileOperationType
    FilePath      string
    ObservationID string
    NewPath       string
    ContentBefore string
    ContentAfter  string
    LinesAdded    *int
    LinesRemoved  *int
    ToolName      string
    Reason        string
    StartedAt     *time.Time
    CompletedAt   *time.Time
    Success       *bool
    ErrorMessage  string
}

type FileOperationInfo struct {
    ID            string
    TraceID       string
    Operation     FileOperationType
    FilePath      string
    FileSize      int64
    LinesAdded    int
    LinesRemoved  int
    Success       bool
    DurationMs    int64
    StartedAt     time.Time
    CompletedAt   time.Time
}

// Terminal Commands
type TerminalCommandOptions struct {
    Command          string
    Args             []string
    ObservationID    string
    WorkingDirectory string
    ExitCode         int
    Stdout           string
    Stderr           string
    ToolName         string
    Reason           string
    StartedAt        *time.Time
    CompletedAt      *time.Time
}

type RunCommandOptions struct {
    Args             []string
    WorkingDirectory string
    Env              map[string]string
    Timeout          time.Duration
    MaxOutputBytes   int
}

type RunCommandResult struct {
    Info     *TerminalCommandInfo
    ExitCode int
    Stdout   string
    Stderr   string
}
```

### Prompts

```go
type GetPromptOptions struct {
    Name     string
    Version  *int
    Label    string
    Fallback string
    CacheTTL time.Duration
}

type PromptVersion struct {
    ID        string
    Version   int
    Prompt    string
    Config    map[string]any
    Labels    []string
    CreatedAt string
}

type ChatMessage struct {
    Role    string // system, user, assistant, function
    Content string
}
```
