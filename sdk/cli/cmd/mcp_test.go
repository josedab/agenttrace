package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	agenttrace "github.com/agenttrace/agenttrace-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMCPServerCapabilities(t *testing.T) {
	server := &MCPServer{}

	req := httptest.NewRequest(http.MethodGet, "/mcp/capabilities", nil)
	w := httptest.NewRecorder()

	server.handleCapabilities(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var caps MCPCapabilities
	err := json.Unmarshal(w.Body.Bytes(), &caps)
	require.NoError(t, err)
	assert.True(t, caps.Tools)
	assert.False(t, caps.Resources)
	assert.True(t, caps.Prompts)
}

func TestMCPServerToolsList(t *testing.T) {
	server := &MCPServer{}

	req := httptest.NewRequest(http.MethodGet, "/mcp/tools/list", nil)
	w := httptest.NewRecorder()

	server.handleToolsList(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var result map[string][]MCPTool
	err := json.Unmarshal(w.Body.Bytes(), &result)
	require.NoError(t, err)

	tools := result["tools"]
	require.Len(t, tools, 8)

	toolNames := make([]string, 0, len(tools))
	for _, tool := range tools {
		toolNames = append(toolNames, tool.Name)
	}
	assert.Contains(t, toolNames, "agenttrace_trace_start")
	assert.Contains(t, toolNames, "agenttrace_trace_end")
	assert.Contains(t, toolNames, "agenttrace_generation")
	assert.Contains(t, toolNames, "agenttrace_score")
	assert.Contains(t, toolNames, "agenttrace_prompt_get")
	assert.Contains(t, toolNames, "agenttrace_trace_search")
	assert.Contains(t, toolNames, "agenttrace_trace_get")
	assert.Contains(t, toolNames, "agenttrace_evaluation_summary")
}

func TestMCPServerToolsCall(t *testing.T) {
	t.Run("rejects non-POST methods", func(t *testing.T) {
		server := &MCPServer{}
		req := httptest.NewRequest(http.MethodGet, "/mcp/tools/call", nil)
		w := httptest.NewRecorder()

		server.handleToolsCall(w, req)

		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})

	t.Run("rejects invalid JSON body", func(t *testing.T) {
		server := &MCPServer{}
		req := httptest.NewRequest(http.MethodPost, "/mcp/tools/call",
			bytes.NewBufferString("not-json"))
		w := httptest.NewRecorder()

		server.handleToolsCall(w, req)

		var resp MCPToolCallResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.True(t, resp.IsError)
		assert.Contains(t, resp.Content[0].Text, "Invalid request")
	})

	t.Run("returns error for unknown tool", func(t *testing.T) {
		server := &MCPServer{}
		body, _ := json.Marshal(MCPToolCallRequest{
			Name:      "unknown_tool",
			Arguments: map[string]any{},
		})
		req := httptest.NewRequest(http.MethodPost, "/mcp/tools/call",
			bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		server.handleToolsCall(w, req)

		var resp MCPToolCallResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.True(t, resp.IsError)
		assert.Contains(t, resp.Content[0].Text, "Unknown tool")
	})
}

func TestMCPToolTraceStart(t *testing.T) {
	t.Run("requires name argument", func(t *testing.T) {
		server := &MCPServer{}
		result, isError := server.toolTraceStart(t.Context(), map[string]any{})
		assert.True(t, isError)
		assert.Contains(t, result, "name is required")
	})

	t.Run("requires name to be non-empty", func(t *testing.T) {
		server := &MCPServer{}
		result, isError := server.toolTraceStart(t.Context(), map[string]any{"name": ""})
		assert.True(t, isError)
		assert.Contains(t, result, "name is required")
	})
}

func TestMCPToolTraceEnd(t *testing.T) {
	t.Run("returns error when no active trace", func(t *testing.T) {
		server := &MCPServer{}
		result, isError := server.toolTraceEnd(map[string]any{})
		assert.True(t, isError)
		assert.Contains(t, result, "No active trace")
	})
}

func TestMCPToolGeneration(t *testing.T) {
	t.Run("requires active trace", func(t *testing.T) {
		server := &MCPServer{}
		result, isError := server.toolGeneration(map[string]any{
			"name":  "test-gen",
			"model": "gpt-4",
		})
		assert.True(t, isError)
		assert.Contains(t, result, "No active trace")
	})

	t.Run("requires name argument", func(t *testing.T) {
		server := &MCPServer{currentTrace: nil}
		_, isError := server.toolGeneration(map[string]any{
			"model": "gpt-4",
		})
		assert.True(t, isError)
	})
}

func TestMCPToolScore(t *testing.T) {
	t.Run("requires active trace", func(t *testing.T) {
		server := &MCPServer{}
		result, isError := server.toolScore(map[string]any{
			"name":  "quality",
			"value": 0.9,
		})
		assert.True(t, isError)
		assert.Contains(t, result, "No active trace")
	})

	t.Run("requires active trace even with missing name", func(t *testing.T) {
		server := &MCPServer{}
		result, isError := server.toolScore(map[string]any{
			"value": 0.9,
		})
		assert.True(t, isError)
		assert.Contains(t, result, "No active trace")
	})

	t.Run("requires active trace even with missing value", func(t *testing.T) {
		server := &MCPServer{}
		result, isError := server.toolScore(map[string]any{
			"name": "quality",
		})
		assert.True(t, isError)
		assert.Contains(t, result, "No active trace")
	})
}

func TestMCPToolPromptGet(t *testing.T) {
	t.Run("requires name argument", func(t *testing.T) {
		server := &MCPServer{}
		result, isError := server.toolPromptGet(t.Context(), map[string]any{})
		assert.True(t, isError)
		assert.Contains(t, result, "name is required")
	})
}

func TestMCPReadOnlyTraceTools(t *testing.T) {
	t.Run("gets a source-free trace summary with local credentials", func(t *testing.T) {
		apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/api/public/traces/trace-1", r.URL.Path)
			assert.Equal(t, "Bearer local-key", r.Header.Get("Authorization"))
			assert.Equal(t, "project-1", r.Header.Get("X-Project-ID"))
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{
				"id":"trace-1",
				"name":"agent run",
				"input":"secret prompt",
				"output":"secret output",
				"metadata":{"token":"secret"},
				"totalCost":0.12
			}`))
		}))
		defer apiServer.Close()

		apiClient, err := newMCPAPIClient(apiServer.URL, "local-key", "project-1")
		require.NoError(t, err)
		apiClient.client = apiServer.Client()
		server := &MCPServer{apiClient: apiClient}

		result, isError := server.toolTraceGet(t.Context(), map[string]any{"trace_id": "trace-1"})

		assert.False(t, isError)
		assert.Contains(t, result, "agent run")
		assert.Contains(t, result, "totalCost")
		assert.NotContains(t, result, "secret prompt")
		assert.NotContains(t, result, "secret output")
		assert.NotContains(t, result, `"metadata"`)
	})

	t.Run("searches within bounded query parameters", func(t *testing.T) {
		apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/api/public/traces/search", r.URL.Path)
			assert.Equal(t, "fix tests", r.URL.Query().Get("q"))
			assert.Equal(t, "5", r.URL.Query().Get("limit"))
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"traces":[{"id":"trace-1","name":"fix tests"}]}`))
		}))
		defer apiServer.Close()

		apiClient, err := newMCPAPIClient(apiServer.URL, "local-key", "")
		require.NoError(t, err)
		apiClient.client = apiServer.Client()
		server := &MCPServer{apiClient: apiClient}

		result, isError := server.toolTraceSearch(t.Context(), map[string]any{
			"query": "fix tests",
			"limit": float64(5),
		})

		assert.False(t, isError)
		assert.Contains(t, result, "trace-1")
	})
}

func TestMCPToolEvaluationSummary(t *testing.T) {
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/public/evaluators":
			_, _ = w.Write([]byte(`{"evaluators":[{"enabled":true},{"enabled":false}]}`))
		case "/api/public/eval-hub/runs":
			_, _ = w.Write([]byte(`{"runs":[{"status":"completed"},{"status":"unsupported"}]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer apiServer.Close()

	apiClient, err := newMCPAPIClient(apiServer.URL, "local-key", "")
	require.NoError(t, err)
	apiClient.client = apiServer.Client()
	server := &MCPServer{apiClient: apiClient}

	result, isError := server.toolEvaluationSummary(t.Context(), nil)

	assert.False(t, isError)
	assert.Contains(t, result, `"active": 1`)
	assert.Contains(t, result, `"completed": 1`)
	assert.Contains(t, result, `"unsupported": 1`)
}

func TestNewMCPAPIClientRejectsInvalidHost(t *testing.T) {
	_, err := newMCPAPIClient("file:///tmp/socket", "key", "")
	require.Error(t, err)
}

func TestHelperFunctions(t *testing.T) {
	t.Run("getString returns value for existing key", func(t *testing.T) {
		m := map[string]any{"key": "value"}
		assert.Equal(t, "value", getString(m, "key"))
	})

	t.Run("getString returns empty for missing key", func(t *testing.T) {
		m := map[string]any{}
		assert.Empty(t, getString(m, "key"))
	})

	t.Run("getString returns empty for non-string value", func(t *testing.T) {
		m := map[string]any{"key": 123}
		assert.Empty(t, getString(m, "key"))
	})

	t.Run("getInt returns value for existing key", func(t *testing.T) {
		m := map[string]any{"key": float64(42)}
		assert.Equal(t, 42, getInt(m, "key"))
	})

	t.Run("getInt returns 0 for missing key", func(t *testing.T) {
		m := map[string]any{}
		assert.Equal(t, 0, getInt(m, "key"))
	})

	t.Run("getInt returns 0 for non-numeric value", func(t *testing.T) {
		m := map[string]any{"key": "not-a-number"}
		assert.Equal(t, 0, getInt(m, "key"))
	})
}

func TestMCPListenAddressBindsLoopbackOnly(t *testing.T) {
	t.Run("defaults to the loopback interface", func(t *testing.T) {
		address, err := mcpListenAddress("", 8765)
		require.NoError(t, err)
		assert.Equal(t, "127.0.0.1:8765", address)
		assert.Equal(t, "http://127.0.0.1:8765", mcpDisplayURL(address))
	})

	t.Run("accepts explicit loopback addresses", func(t *testing.T) {
		address, err := mcpListenAddress("localhost", 9000)
		require.NoError(t, err)
		assert.Equal(t, "127.0.0.1:9000", address)

		address, err = mcpListenAddress("::1", 9000)
		require.NoError(t, err)
		assert.Equal(t, "[::1]:9000", address)
		assert.Equal(t, "http://[::1]:9000", mcpDisplayURL(address))
	})

	t.Run("refuses routable interfaces", func(t *testing.T) {
		for _, host := range []string{"0.0.0.0", "192.168.1.10", "example.com", "::"} {
			_, err := mcpListenAddress(host, 8765)
			require.Error(t, err, "host %s must be refused", host)
		}
	})

	t.Run("validates the port range", func(t *testing.T) {
		_, err := mcpListenAddress("127.0.0.1", 0)
		require.Error(t, err)
		_, err = mcpListenAddress("127.0.0.1", 70000)
		require.Error(t, err)
	})
}

func TestMCPServerTimeoutsAreBounded(t *testing.T) {
	handler := http.NewServeMux()
	server := newMCPHTTPServer("127.0.0.1:8765", handler)

	assert.Equal(t, "127.0.0.1:8765", server.Addr)
	assert.Same(t, handler, server.Handler)
	assert.Positive(t, server.ReadHeaderTimeout)
	assert.Positive(t, server.ReadTimeout)
	assert.Positive(t, server.WriteTimeout)
	assert.Positive(t, server.IdleTimeout)
	assert.LessOrEqual(t, server.ReadHeaderTimeout, server.ReadTimeout)
}

func TestMCPTraceStartSerializesConcurrentRequests(t *testing.T) {
	enabled := false
	client := agenttrace.New(agenttrace.Config{
		APIKey:  "local-test-key",
		Host:    "http://127.0.0.1",
		Enabled: &enabled,
	})
	t.Cleanup(client.Shutdown)
	server := &MCPServer{client: client}

	start := make(chan struct{})
	results := make(chan bool, 16)
	var wg sync.WaitGroup
	for index := range 16 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			_, isError := server.toolTraceStart(t.Context(), map[string]any{
				"name": "trace-" + string(rune('a'+index)),
			})
			results <- isError
		}()
	}
	close(start)
	wg.Wait()
	close(results)

	successes := 0
	for isError := range results {
		if !isError {
			successes++
		}
	}
	assert.Equal(t, 1, successes)

	_, isError := server.toolTraceEnd(map[string]any{})
	assert.False(t, isError)
	_, isError = server.toolTraceEnd(map[string]any{})
	assert.True(t, isError)
}

func TestMCPAPIClientPropagatesRequestCancellation(t *testing.T) {
	requestStarted := make(chan struct{})
	apiServer := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		close(requestStarted)
		<-r.Context().Done()
	}))
	defer apiServer.Close()

	apiClient, err := newMCPAPIClient(apiServer.URL, "local-key", "")
	require.NoError(t, err)
	apiClient.client = apiServer.Client()

	ctx, cancel := context.WithCancel(context.Background())
	result := make(chan error, 1)
	go func() {
		var destination map[string]any
		result <- apiClient.getJSON(ctx, "/blocked", &destination)
	}()

	select {
	case <-requestStarted:
	case <-time.After(time.Second):
		t.Fatal("request did not reach the server")
	}
	cancel()

	select {
	case requestErr := <-result:
		require.Error(t, requestErr)
		assert.ErrorIs(t, requestErr, context.Canceled)
	case <-time.After(time.Second):
		t.Fatal("canceled request did not return")
	}
}
