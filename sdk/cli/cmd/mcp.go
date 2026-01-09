package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"

	agenttrace "github.com/agenttrace/agenttrace-go"
	"github.com/spf13/cobra"
)

var (
	mcpPort int
)

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Start an MCP server for IDE integration",
	Long: `Start a Model Context Protocol (MCP) server for IDE integration.

This allows IDEs and AI assistants to interact with AgentTrace
for tracing, prompt management, and more.

Example:
  agenttrace mcp --port 8080`,
	RunE: runMCP,
}

func init() {
	mcpCmd.Flags().IntVar(&mcpPort, "port", 8765, "Port to run the MCP server on")
}

func runMCP(cmd *cobra.Command, args []string) error {
	apiKey := getAPIKey()
	if apiKey == "" {
		return fmt.Errorf("API key required. Set --api-key or AGENTTRACE_API_KEY")
	}

	// Initialize AgentTrace client
	client := agenttrace.New(agenttrace.Config{
		APIKey: apiKey,
		Host:   host,
	})
	defer client.Shutdown()

	// Create MCP server
	server := &MCPServer{
		client: client,
	}

	mux := http.NewServeMux()

	// MCP protocol endpoints
	mux.HandleFunc("/mcp/capabilities", server.handleCapabilities)
	mux.HandleFunc("/mcp/tools/list", server.handleToolsList)
	mux.HandleFunc("/mcp/tools/call", server.handleToolsCall)

	// Health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	addr := fmt.Sprintf(":%d", mcpPort)
	fmt.Printf("Starting MCP server on http://localhost%s\n", addr)
	fmt.Println("Available tools:")
	fmt.Println("  - agenttrace_trace_start: Start a new trace")
	fmt.Println("  - agenttrace_trace_end: End the current trace")
	fmt.Println("  - agenttrace_generation: Log an LLM generation")
	fmt.Println("  - agenttrace_score: Submit a score")
	fmt.Println("  - agenttrace_prompt_get: Fetch a prompt")

	return http.ListenAndServe(addr, mux)
}

// MCPServer handles MCP protocol requests
type MCPServer struct {
	client       *agenttrace.Client
	currentTrace *agenttrace.Trace
}

// MCPCapabilities represents server capabilities
type MCPCapabilities struct {
	Tools     bool `json:"tools"`
	Resources bool `json:"resources"`
	Prompts   bool `json:"prompts"`
}

// MCPTool represents an MCP tool
type MCPTool struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	InputSchema map[string]any    `json:"inputSchema"`
}

// MCPToolCallRequest represents a tool call request
type MCPToolCallRequest struct {
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments"`
}

// MCPToolCallResponse represents a tool call response
type MCPToolCallResponse struct {
	Content []MCPContent `json:"content"`
	IsError bool         `json:"isError,omitempty"`
}

// MCPContent represents content in a response
type MCPContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

func (s *MCPServer) handleCapabilities(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(MCPCapabilities{
		Tools:     true,
		Resources: false,
		Prompts:   true,
	})
}

func (s *MCPServer) handleToolsList(w http.ResponseWriter, r *http.Request) {
	tools := []MCPTool{
		{
			Name:        "agenttrace_trace_start",
			Description: "Start a new trace for tracking agent execution",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"name": map[string]any{
						"type":        "string",
						"description": "Name of the trace",
					},
					"user_id": map[string]any{
						"type":        "string",
						"description": "User ID",
					},
					"session_id": map[string]any{
						"type":        "string",
						"description": "Session ID",
					},
				},
				"required": []string{"name"},
			},
		},
		{
			Name:        "agenttrace_trace_end",
			Description: "End the current trace",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"output": map[string]any{
						"type":        "string",
						"description": "Final output of the trace",
					},
				},
			},
		},
		{
			Name:        "agenttrace_generation",
			Description: "Log an LLM generation/completion",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"name": map[string]any{
						"type":        "string",
						"description": "Name of the generation",
					},
					"model": map[string]any{
						"type":        "string",
						"description": "Model name",
					},
					"input": map[string]any{
						"type":        "string",
						"description": "Input prompt",
					},
					"output": map[string]any{
						"type":        "string",
						"description": "Model output",
					},
					"input_tokens": map[string]any{
						"type":        "integer",
						"description": "Number of input tokens",
					},
					"output_tokens": map[string]any{
						"type":        "integer",
						"description": "Number of output tokens",
					},
				},
				"required": []string{"name", "model"},
			},
		},
		{
			Name:        "agenttrace_score",
			Description: "Submit a score for the current trace",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"name": map[string]any{
						"type":        "string",
						"description": "Score name",
					},
					"value": map[string]any{
						"type":        "number",
						"description": "Score value (0-1)",
					},
					"comment": map[string]any{
						"type":        "string",
						"description": "Optional comment",
					},
				},
				"required": []string{"name", "value"},
			},
		},
		{
			Name:        "agenttrace_prompt_get",
			Description: "Fetch a prompt by name",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"name": map[string]any{
						"type":        "string",
						"description": "Prompt name",
					},
					"version": map[string]any{
						"type":        "integer",
						"description": "Optional version number",
					},
					"label": map[string]any{
						"type":        "string",
						"description": "Optional label (e.g., 'production')",
					},
					"variables": map[string]any{
						"type":        "object",
						"description": "Variables to compile the prompt with",
					},
				},
				"required": []string{"name"},
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"tools": tools})
}

func (s *MCPServer) handleToolsCall(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req MCPToolCallRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.respondError(w, fmt.Sprintf("Invalid request: %v", err))
		return
	}

	var result string
	var isError bool

	switch req.Name {
	case "agenttrace_trace_start":
		result, isError = s.toolTraceStart(req.Arguments)
	case "agenttrace_trace_end":
		result, isError = s.toolTraceEnd(req.Arguments)
	case "agenttrace_generation":
		result, isError = s.toolGeneration(req.Arguments)
	case "agenttrace_score":
		result, isError = s.toolScore(req.Arguments)
	case "agenttrace_prompt_get":
		result, isError = s.toolPromptGet(req.Arguments)
	default:
		result = fmt.Sprintf("Unknown tool: %s", req.Name)
		isError = true
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(MCPToolCallResponse{
		Content: []MCPContent{{Type: "text", Text: result}},
		IsError: isError,
	})
}

func (s *MCPServer) respondError(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(MCPToolCallResponse{
		Content: []MCPContent{{Type: "text", Text: msg}},
		IsError: true,
	})
}

func (s *MCPServer) toolTraceStart(args map[string]any) (string, bool) {
	name, _ := args["name"].(string)
	if name == "" {
		return "name is required", true
	}

	userID, _ := args["user_id"].(string)
	sessionID, _ := args["session_id"].(string)

	s.currentTrace = s.client.Trace(nil, agenttrace.TraceOptions{
		Name:      name,
		UserID:    userID,
		SessionID: sessionID,
	})

	return fmt.Sprintf("Trace started: %s (ID: %s)", name, s.currentTrace.ID()), false
}

func (s *MCPServer) toolTraceEnd(args map[string]any) (string, bool) {
	if s.currentTrace == nil {
		return "No active trace", true
	}

	output, _ := args["output"].(string)

	s.currentTrace.End(&agenttrace.TraceEndOptions{Output: output})
	traceID := s.currentTrace.ID()
	s.currentTrace = nil

	return fmt.Sprintf("Trace ended: %s", traceID), false
}

func (s *MCPServer) toolGeneration(args map[string]any) (string, bool) {
	if s.currentTrace == nil {
		return "No active trace. Start a trace first.", true
	}

	name, _ := args["name"].(string)
	if name == "" {
		return "name is required", true
	}

	model, _ := args["model"].(string)
	input, _ := args["input"].(string)
	output, _ := args["output"].(string)
	inputTokens, _ := args["input_tokens"].(float64)
	outputTokens, _ := args["output_tokens"].(float64)

	gen := s.currentTrace.Generation(agenttrace.GenerationOptions{
		Name:  name,
		Model: model,
		Input: input,
	})

	var usage *agenttrace.UsageDetails
	if inputTokens > 0 || outputTokens > 0 {
		usage = &agenttrace.UsageDetails{
			InputTokens:  int(inputTokens),
			OutputTokens: int(outputTokens),
			TotalTokens:  int(inputTokens + outputTokens),
		}
	}

	gen.End(&agenttrace.GenerationEndOptions{
		Output: output,
		Usage:  usage,
	})

	return fmt.Sprintf("Generation logged: %s", gen.ID()), false
}

func (s *MCPServer) toolScore(args map[string]any) (string, bool) {
	if s.currentTrace == nil {
		return "No active trace. Start a trace first.", true
	}

	name, _ := args["name"].(string)
	if name == "" {
		return "name is required", true
	}

	value, ok := args["value"].(float64)
	if !ok {
		return "value is required and must be a number", true
	}

	comment, _ := args["comment"].(string)

	s.currentTrace.Score(name, value, &agenttrace.ScoreAddOptions{
		Comment: comment,
	})

	return fmt.Sprintf("Score submitted: %s = %.2f", name, value), false
}

func (s *MCPServer) toolPromptGet(args map[string]any) (string, bool) {
	name, _ := args["name"].(string)
	if name == "" {
		return "name is required", true
	}

	opts := agenttrace.GetPromptOptions{Name: name}

	if version, ok := args["version"].(float64); ok {
		v := int(version)
		opts.Version = &v
	}
	if label, ok := args["label"].(string); ok {
		opts.Label = label
	}

	prompt, err := agenttrace.GetPrompt(opts)
	if err != nil {
		return fmt.Sprintf("Failed to get prompt: %v", err), true
	}

	// Compile with variables if provided
	variables, _ := args["variables"].(map[string]any)
	if variables != nil {
		compiled := prompt.Compile(variables)
		return compiled, false
	}

	return prompt.Prompt, false
}

func getString(m map[string]any, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func getInt(m map[string]any, key string) int {
	if v, ok := m[key].(float64); ok {
		return int(v)
	}
	return 0
}
