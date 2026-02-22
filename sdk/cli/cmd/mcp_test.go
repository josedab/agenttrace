package cmd

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

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
	require.Len(t, tools, 5)

	toolNames := make([]string, 0, len(tools))
	for _, tool := range tools {
		toolNames = append(toolNames, tool.Name)
	}
	assert.Contains(t, toolNames, "agenttrace_trace_start")
	assert.Contains(t, toolNames, "agenttrace_trace_end")
	assert.Contains(t, toolNames, "agenttrace_generation")
	assert.Contains(t, toolNames, "agenttrace_score")
	assert.Contains(t, toolNames, "agenttrace_prompt_get")
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
		result, isError := server.toolTraceStart(map[string]any{})
		assert.True(t, isError)
		assert.Contains(t, result, "name is required")
	})

	t.Run("requires name to be non-empty", func(t *testing.T) {
		server := &MCPServer{}
		result, isError := server.toolTraceStart(map[string]any{"name": ""})
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
		result, isError := server.toolPromptGet(map[string]any{})
		assert.True(t, isError)
		assert.Contains(t, result, "name is required")
	})
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
