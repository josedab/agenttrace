package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	agenttrace "github.com/agenttrace/agenttrace-go"
	"github.com/spf13/cobra"
)

var (
	mcpPort      int
	mcpHost      string
	mcpProjectID string
)

// MCP server timeouts bound how long a single connection may occupy the server.
const (
	mcpReadHeaderTimeout = 5 * time.Second
	mcpReadTimeout       = 30 * time.Second
	mcpWriteTimeout      = 60 * time.Second
	mcpIdleTimeout       = 120 * time.Second
)

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Start an MCP server for IDE integration",
	Long: `Start a Model Context Protocol (MCP) server for IDE integration.

This allows IDEs and AI assistants to interact with AgentTrace
for tracing, prompt management, and more.

The server holds an API key and exposes project data, so it binds to the
loopback interface only. Remote access must be provided by an authenticated
tunnel or reverse proxy that you control.

Example:
  agenttrace mcp --port 8080`,
	RunE: runMCP,
}

func init() {
	mcpCmd.Flags().IntVar(&mcpPort, "port", 8765, "Port to run the MCP server on")
	mcpCmd.Flags().StringVar(
		&mcpHost,
		"host",
		defaultMCPHost,
		"Loopback address to bind the MCP server to",
	)
	mcpCmd.Flags().StringVar(
		&mcpProjectID,
		"project-id",
		"",
		"Optional project ID header (API keys remain scoped to their own project)",
	)
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
	apiClient, err := newMCPAPIClient(host, apiKey, mcpProjectID)
	if err != nil {
		return err
	}

	// Create MCP server
	server := &MCPServer{
		client:    client,
		apiClient: apiClient,
	}

	address, err := mcpListenAddress(mcpHost, mcpPort)
	if err != nil {
		return err
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

	fmt.Printf("Starting MCP server on %s\n", mcpDisplayURL(address))
	fmt.Println("Available tools:")
	fmt.Println("  - agenttrace_trace_start: Start a new trace")
	fmt.Println("  - agenttrace_trace_end: End the current trace")
	fmt.Println("  - agenttrace_generation: Log an LLM generation")
	fmt.Println("  - agenttrace_score: Submit a score")
	fmt.Println("  - agenttrace_prompt_get: Fetch a prompt")
	fmt.Println("  - agenttrace_trace_search: Search project traces (read-only)")
	fmt.Println("  - agenttrace_trace_get: Get a redacted trace summary (read-only)")
	fmt.Println("  - agenttrace_evaluation_summary: Summarize project evaluations (read-only)")

	httpServer := newMCPHTTPServer(address, mux)
	return httpServer.ListenAndServe()
}

func newMCPHTTPServer(address string, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:              address,
		Handler:           handler,
		ReadHeaderTimeout: mcpReadHeaderTimeout,
		ReadTimeout:       mcpReadTimeout,
		WriteTimeout:      mcpWriteTimeout,
		IdleTimeout:       mcpIdleTimeout,
	}
}

// defaultMCPHost keeps the MCP server reachable only from this machine.
const defaultMCPHost = "127.0.0.1"

// mcpListenAddress builds the listen address and refuses any non-loopback bind.
// The server carries an API key and answers unauthenticated requests, so
// exposing it on a routable interface would hand project data to the network.
func mcpListenAddress(host string, port int) (string, error) {
	if port < 1 || port > 65535 {
		return "", fmt.Errorf("--port must be between 1 and 65535")
	}
	if host == "" {
		host = defaultMCPHost
	}
	ip := net.ParseIP(host)
	if ip == nil {
		if !strings.EqualFold(host, "localhost") {
			return "", fmt.Errorf(
				"--host must be a loopback address; use a tunnel with authentication for remote access",
			)
		}
		ip = net.ParseIP(defaultMCPHost)
	}
	if !ip.IsLoopback() {
		return "", fmt.Errorf(
			"--host %s is not a loopback address; the MCP server serves unauthenticated local requests",
			host,
		)
	}
	return net.JoinHostPort(ip.String(), strconv.Itoa(port)), nil
}

// mcpDisplayURL renders the address that was actually bound.
func mcpDisplayURL(address string) string {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return "http://" + address
	}
	return "http://" + net.JoinHostPort(host, port)
}

// MCPServer handles MCP protocol requests.
// Tool calls arrive on independent HTTP connections, so the active trace is
// guarded; without it two concurrent calls could observe a torn state.
type MCPServer struct {
	client    *agenttrace.Client
	apiClient *mcpAPIClient

	traceMu      sync.Mutex
	currentTrace *agenttrace.Trace
}

type mcpAPIClient struct {
	baseURL   string
	apiKey    string
	projectID string
	client    *http.Client
}

func newMCPAPIClient(baseURL, apiKey, projectID string) (*mcpAPIClient, error) {
	parsed, err := url.Parse(strings.TrimRight(baseURL, "/"))
	if err != nil ||
		(parsed.Scheme != "http" && parsed.Scheme != "https") ||
		parsed.Host == "" ||
		parsed.User != nil {
		return nil, fmt.Errorf("invalid AgentTrace API host")
	}
	return &mcpAPIClient{
		baseURL:   parsed.String(),
		apiKey:    apiKey,
		projectID: projectID,
		client:    &http.Client{Timeout: 15 * time.Second},
	}, nil
}

func (c *mcpAPIClient) getJSON(ctx context.Context, path string, destination any) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return err
	}
	request.Header.Set("Authorization", "Bearer "+c.apiKey)
	if c.projectID != "" {
		request.Header.Set("X-Project-ID", c.projectID)
	}

	response, err := c.client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(response.Body, 4096))
		return fmt.Errorf("AgentTrace API returned %d: %s", response.StatusCode, strings.TrimSpace(string(body)))
	}
	return json.NewDecoder(response.Body).Decode(destination)
}

// MCPCapabilities represents server capabilities
type MCPCapabilities struct {
	Tools     bool `json:"tools"`
	Resources bool `json:"resources"`
	Prompts   bool `json:"prompts"`
}

// MCPTool represents an MCP tool
type MCPTool struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema map[string]any `json:"inputSchema"`
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
		{
			Name:        "agenttrace_trace_search",
			Description: "Search traces in the API-key scoped AgentTrace project without modifying data",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"query": map[string]any{
						"type":        "string",
						"description": "Optional full-text trace query",
					},
					"limit": map[string]any{
						"type":        "integer",
						"description": "Maximum results (1-50)",
					},
					"from": map[string]any{
						"type":        "string",
						"description": "Optional RFC3339 lower time bound",
					},
					"to": map[string]any{
						"type":        "string",
						"description": "Optional RFC3339 upper time bound",
					},
				},
			},
		},
		{
			Name:        "agenttrace_trace_get",
			Description: "Get a source-free trace summary from the API-key scoped project",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"trace_id": map[string]any{
						"type":        "string",
						"description": "Trace ID",
					},
				},
				"required": []string{"trace_id"},
			},
		},
		{
			Name:        "agenttrace_evaluation_summary",
			Description: "Summarize evaluators and Eval Hub run states for the scoped project",
			InputSchema: map[string]any{
				"type":       "object",
				"properties": map[string]any{},
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
		result, isError = s.toolTraceStart(r.Context(), req.Arguments)
	case "agenttrace_trace_end":
		result, isError = s.toolTraceEnd(req.Arguments)
	case "agenttrace_generation":
		result, isError = s.toolGeneration(req.Arguments)
	case "agenttrace_score":
		result, isError = s.toolScore(req.Arguments)
	case "agenttrace_prompt_get":
		result, isError = s.toolPromptGet(r.Context(), req.Arguments)
	case "agenttrace_trace_search":
		result, isError = s.toolTraceSearch(r.Context(), req.Arguments)
	case "agenttrace_trace_get":
		result, isError = s.toolTraceGet(r.Context(), req.Arguments)
	case "agenttrace_evaluation_summary":
		result, isError = s.toolEvaluationSummary(r.Context(), req.Arguments)
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

func (s *MCPServer) toolTraceStart(
	ctx context.Context,
	args map[string]any,
) (string, bool) {
	name, _ := args["name"].(string)
	if name == "" {
		return "name is required", true
	}
	if s.client == nil {
		return "AgentTrace client is not configured", true
	}

	userID, _ := args["user_id"].(string)
	sessionID, _ := args["session_id"].(string)

	s.traceMu.Lock()
	defer s.traceMu.Unlock()
	if s.currentTrace != nil {
		return "An active trace already exists. End it before starting another trace.", true
	}
	trace := s.client.Trace(ctx, agenttrace.TraceOptions{
		Name:      name,
		UserID:    userID,
		SessionID: sessionID,
	})
	s.currentTrace = trace

	return fmt.Sprintf("Trace started: %s (ID: %s)", name, trace.ID()), false
}

func (s *MCPServer) toolTraceEnd(args map[string]any) (string, bool) {
	s.traceMu.Lock()
	defer s.traceMu.Unlock()
	trace := s.currentTrace
	if trace == nil {
		return "No active trace", true
	}

	output, _ := args["output"].(string)

	trace.End(&agenttrace.TraceEndOptions{Output: output})
	s.currentTrace = nil

	return fmt.Sprintf("Trace ended: %s", trace.ID()), false
}

func (s *MCPServer) toolGeneration(args map[string]any) (string, bool) {
	s.traceMu.Lock()
	defer s.traceMu.Unlock()
	trace := s.currentTrace
	if trace == nil {
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

	gen := trace.Generation(agenttrace.GenerationOptions{
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
	s.traceMu.Lock()
	defer s.traceMu.Unlock()
	trace := s.currentTrace
	if trace == nil {
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

	trace.Score(name, value, &agenttrace.ScoreAddOptions{
		Comment: comment,
	})

	return fmt.Sprintf("Score submitted: %s = %.2f", name, value), false
}

func (s *MCPServer) toolPromptGet(ctx context.Context, args map[string]any) (string, bool) {
	name, _ := args["name"].(string)
	if name == "" {
		return "name is required", true
	}
	if s.apiClient == nil {
		return "AgentTrace API client is not configured", true
	}

	params := url.Values{}
	if version := getInt(args, "version"); version > 0 {
		params.Set("version", strconv.Itoa(version))
	}
	if label := getString(args, "label"); label != "" {
		params.Set("label", label)
	}
	path := "/api/public/prompts/" + url.PathEscape(name)
	if encoded := params.Encode(); encoded != "" {
		path += "?" + encoded
	}
	var response struct {
		LatestVersion *struct {
			Content string `json:"content"`
		} `json:"latestVersion"`
	}
	if err := s.apiClient.getJSON(ctx, path, &response); err != nil {
		return fmt.Sprintf("Failed to get prompt: %v", err), true
	}
	if response.LatestVersion == nil {
		return "Prompt has no available version", true
	}

	variables, _ := args["variables"].(map[string]any)
	compiled := response.LatestVersion.Content
	for key, value := range variables {
		compiled = strings.ReplaceAll(compiled, "{{"+key+"}}", fmt.Sprint(value))
	}
	return compiled, false
}

func (s *MCPServer) toolTraceSearch(ctx context.Context, args map[string]any) (string, bool) {
	if s.apiClient == nil {
		return "AgentTrace API client is not configured", true
	}
	limit := getInt(args, "limit")
	if limit == 0 {
		limit = 20
	}
	if limit < 1 || limit > 50 {
		return "limit must be between 1 and 50", true
	}

	params := url.Values{}
	params.Set("limit", strconv.Itoa(limit))
	if query := getString(args, "query"); query != "" {
		params.Set("q", query)
	}
	if from := getString(args, "from"); from != "" {
		if _, err := time.Parse(time.RFC3339, from); err != nil {
			return "from must be RFC3339", true
		}
		params.Set("fromTimestamp", from)
	}
	if to := getString(args, "to"); to != "" {
		if _, err := time.Parse(time.RFC3339, to); err != nil {
			return "to must be RFC3339", true
		}
		params.Set("toTimestamp", to)
	}

	var response map[string]any
	path := "/api/public/traces"
	if params.Get("q") != "" {
		path = "/api/public/traces/search"
	}
	if err := s.apiClient.getJSON(ctx, path+"?"+params.Encode(), &response); err != nil {
		return fmt.Sprintf("Failed to search traces: %v", err), true
	}
	return prettyJSON(sanitizeMCPReadResponse(response)), false
}

func (s *MCPServer) toolTraceGet(ctx context.Context, args map[string]any) (string, bool) {
	traceID := getString(args, "trace_id")
	if traceID == "" {
		return "trace_id is required", true
	}
	if s.apiClient == nil {
		return "AgentTrace API client is not configured", true
	}

	var response map[string]any
	if err := s.apiClient.getJSON(
		ctx,
		"/api/public/traces/"+url.PathEscape(traceID),
		&response,
	); err != nil {
		return fmt.Sprintf("Failed to get trace: %v", err), true
	}
	return prettyJSON(sanitizeMCPReadResponse(response)), false
}

func (s *MCPServer) toolEvaluationSummary(ctx context.Context, _ map[string]any) (string, bool) {
	if s.apiClient == nil {
		return "AgentTrace API client is not configured", true
	}

	var evaluatorResponse struct {
		Evaluators []struct {
			Enabled bool `json:"enabled"`
		} `json:"evaluators"`
	}
	if err := s.apiClient.getJSON(ctx, "/api/public/evaluators?limit=100", &evaluatorResponse); err != nil {
		return fmt.Sprintf("Failed to list evaluators: %v", err), true
	}
	var runResponse struct {
		Runs []struct {
			Status string `json:"status"`
		} `json:"runs"`
	}
	if err := s.apiClient.getJSON(ctx, "/api/public/eval-hub/runs?limit=100", &runResponse); err != nil {
		return fmt.Sprintf("Failed to list evaluation runs: %v", err), true
	}

	activeEvaluators := 0
	for _, evaluator := range evaluatorResponse.Evaluators {
		if evaluator.Enabled {
			activeEvaluators++
		}
	}
	runStatuses := make(map[string]int)
	for _, run := range runResponse.Runs {
		runStatuses[run.Status]++
	}
	return prettyJSON(map[string]any{
		"evaluators": map[string]int{
			"total":  len(evaluatorResponse.Evaluators),
			"active": activeEvaluators,
		},
		"runs": map[string]any{
			"total":    len(runResponse.Runs),
			"byStatus": runStatuses,
		},
	}), false
}

func sanitizeMCPReadResponse(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		result := make(map[string]any, len(typed))
		for key, item := range typed {
			switch strings.ToLower(key) {
			case "input", "output", "metadata", "stdout", "stderr", "command",
				"workingdir", "workingdirectory", "diff", "contentbefore",
				"contentafter", "filesnapshot", "storagepath":
				continue
			default:
				result[key] = sanitizeMCPReadResponse(item)
			}
		}
		return result
	case []any:
		result := make([]any, 0, len(typed))
		for _, item := range typed {
			result = append(result, sanitizeMCPReadResponse(item))
		}
		return result
	default:
		return typed
	}
}

func prettyJSON(value any) string {
	var buffer bytes.Buffer
	encoder := json.NewEncoder(&buffer)
	encoder.SetIndent("", "  ")
	encoder.SetEscapeHTML(true)
	if err := encoder.Encode(value); err != nil {
		return "{}"
	}
	return strings.TrimSpace(buffer.String())
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
