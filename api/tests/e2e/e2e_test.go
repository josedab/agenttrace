//go:build e2e
// +build e2e

package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// E2ETestSuite runs end-to-end API tests against a running AgentTrace instance
type E2ETestSuite struct {
	suite.Suite
	baseURL string
	apiKey  string
	client  *http.Client
}

func TestE2ESuite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E tests in short mode")
	}
	suite.Run(t, new(E2ETestSuite))
}

func (s *E2ETestSuite) SetupSuite() {
	s.baseURL = os.Getenv("AGENTTRACE_API_URL")
	if s.baseURL == "" {
		s.baseURL = "http://localhost:8080"
	}

	s.apiKey = os.Getenv("AGENTTRACE_API_KEY")
	if s.apiKey == "" {
		s.T().Fatal("AGENTTRACE_API_KEY environment variable is required")
	}

	s.client = &http.Client{
		Timeout: 30 * time.Second,
	}

	// Wait for API to be ready
	s.waitForAPI()
}

func (s *E2ETestSuite) waitForAPI() {
	maxAttempts := 30
	for i := 0; i < maxAttempts; i++ {
		resp, err := s.client.Get(s.baseURL + "/health")
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			return
		}
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(1 * time.Second)
	}
	s.T().Fatal("API failed to become ready within timeout")
}

// ============ HELPER METHODS ============

func (s *E2ETestSuite) doRequest(method, path string, body interface{}) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequest(method, s.baseURL+path, bodyReader)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", "application/json")

	return s.client.Do(req)
}

func (s *E2ETestSuite) parseResponse(resp *http.Response, v interface{}) {
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	require.NoError(s.T(), err)

	if v != nil {
		err = json.Unmarshal(body, v)
		require.NoError(s.T(), err, "Failed to parse response: %s", string(body))
	}
}

// ============ HEALTH CHECK TESTS ============

func (s *E2ETestSuite) TestHealthEndpoint() {
	resp, err := s.client.Get(s.baseURL + "/health")
	require.NoError(s.T(), err)
	defer resp.Body.Close()

	assert.Equal(s.T(), http.StatusOK, resp.StatusCode)

	var result map[string]string
	s.parseResponse(resp, &result)
	assert.Equal(s.T(), "ok", result["status"])
}

// ============ TRACE TESTS ============

func (s *E2ETestSuite) TestTraceLifecycle() {
	// Create a trace
	traceInput := map[string]interface{}{
		"name":   "e2e-test-trace",
		"input":  map[string]string{"query": "test input"},
		"userId": "e2e-test-user",
		"tags":   []string{"e2e", "test"},
		"metadata": map[string]interface{}{
			"environment": "test",
		},
	}

	resp, err := s.doRequest("POST", "/api/public/traces", traceInput)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusCreated, resp.StatusCode)

	var createResult map[string]interface{}
	s.parseResponse(resp, &createResult)
	traceID := createResult["id"].(string)
	assert.NotEmpty(s.T(), traceID)

	// Get the trace
	resp, err = s.doRequest("GET", "/api/public/traces/"+traceID, nil)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusOK, resp.StatusCode)

	var getResult map[string]interface{}
	s.parseResponse(resp, &getResult)
	assert.Equal(s.T(), traceID, getResult["id"])
	assert.Equal(s.T(), "e2e-test-trace", getResult["name"])

	// Update the trace
	updateInput := map[string]interface{}{
		"output": map[string]string{"result": "test output"},
	}

	resp, err = s.doRequest("PATCH", "/api/public/traces/"+traceID, updateInput)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusOK, resp.StatusCode)

	// List traces with filter
	resp, err = s.doRequest("GET", "/api/public/traces?userId=e2e-test-user", nil)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusOK, resp.StatusCode)

	var listResult map[string]interface{}
	s.parseResponse(resp, &listResult)
	data := listResult["data"].([]interface{})
	assert.GreaterOrEqual(s.T(), len(data), 1)
}

// ============ OBSERVATION TESTS ============

func (s *E2ETestSuite) TestObservationLifecycle() {
	// First create a trace
	traceResp, err := s.doRequest("POST", "/api/public/traces", map[string]interface{}{
		"name": "e2e-observation-test-trace",
	})
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusCreated, traceResp.StatusCode)

	var traceResult map[string]interface{}
	s.parseResponse(traceResp, &traceResult)
	traceID := traceResult["id"].(string)

	// Create a span observation
	spanInput := map[string]interface{}{
		"traceId": traceID,
		"type":    "SPAN",
		"name":    "e2e-test-span",
		"input":   map[string]string{"data": "span input"},
	}

	resp, err := s.doRequest("POST", "/api/public/observations", spanInput)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusCreated, resp.StatusCode)

	var spanResult map[string]interface{}
	s.parseResponse(resp, &spanResult)
	spanID := spanResult["id"].(string)
	assert.NotEmpty(s.T(), spanID)

	// Create a generation observation
	genInput := map[string]interface{}{
		"traceId":             traceID,
		"parentObservationId": spanID,
		"type":                "GENERATION",
		"name":                "e2e-test-generation",
		"model":               "gpt-4",
		"input": []map[string]string{
			{"role": "user", "content": "Hello"},
		},
		"output": map[string]string{
			"role": "assistant", "content": "Hi there!",
		},
		"usage": map[string]int{
			"promptTokens":     10,
			"completionTokens": 5,
			"totalTokens":      15,
		},
	}

	resp, err = s.doRequest("POST", "/api/public/observations", genInput)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusCreated, resp.StatusCode)

	var genResult map[string]interface{}
	s.parseResponse(resp, &genResult)
	genID := genResult["id"].(string)
	assert.NotEmpty(s.T(), genID)

	// Get observation
	resp, err = s.doRequest("GET", "/api/public/observations/"+genID, nil)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusOK, resp.StatusCode)

	var getResult map[string]interface{}
	s.parseResponse(resp, &getResult)
	assert.Equal(s.T(), "gpt-4", getResult["model"])

	// List observations for trace
	resp, err = s.doRequest("GET", "/api/public/observations?traceId="+traceID, nil)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusOK, resp.StatusCode)

	var listResult map[string]interface{}
	s.parseResponse(resp, &listResult)
	data := listResult["data"].([]interface{})
	assert.Equal(s.T(), 2, len(data)) // span + generation
}

// ============ SCORE TESTS ============

func (s *E2ETestSuite) TestScoreLifecycle() {
	// Create a trace first
	traceResp, err := s.doRequest("POST", "/api/public/traces", map[string]interface{}{
		"name": "e2e-score-test-trace",
	})
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusCreated, traceResp.StatusCode)

	var traceResult map[string]interface{}
	s.parseResponse(traceResp, &traceResult)
	traceID := traceResult["id"].(string)

	// Create a numeric score
	scoreInput := map[string]interface{}{
		"traceId": traceID,
		"name":    "quality",
		"value":   0.95,
		"comment": "High quality response",
	}

	resp, err := s.doRequest("POST", "/api/public/scores", scoreInput)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusCreated, resp.StatusCode)

	var scoreResult map[string]interface{}
	s.parseResponse(resp, &scoreResult)
	assert.NotEmpty(s.T(), scoreResult["id"])
	assert.Equal(s.T(), 0.95, scoreResult["value"])

	// Create a categorical score
	catScoreInput := map[string]interface{}{
		"traceId":     traceID,
		"name":        "sentiment",
		"stringValue": "positive",
	}

	resp, err = s.doRequest("POST", "/api/public/scores", catScoreInput)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusCreated, resp.StatusCode)

	// List scores for trace
	resp, err = s.doRequest("GET", "/api/public/scores?traceId="+traceID, nil)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusOK, resp.StatusCode)

	var listResult map[string]interface{}
	s.parseResponse(resp, &listResult)
	data := listResult["data"].([]interface{})
	assert.Equal(s.T(), 2, len(data))
}

// ============ PROMPT TESTS ============

func (s *E2ETestSuite) TestPromptLifecycle() {
	promptName := fmt.Sprintf("e2e-test-prompt-%d", time.Now().UnixNano())

	// Create a text prompt
	promptInput := map[string]interface{}{
		"name":        promptName,
		"type":        "text",
		"prompt":      "You are a helpful assistant. Answer the following: {{question}}",
		"description": "E2E test prompt",
		"labels":      []string{"latest"},
	}

	resp, err := s.doRequest("POST", "/api/public/prompts", promptInput)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusCreated, resp.StatusCode)

	var createResult map[string]interface{}
	s.parseResponse(resp, &createResult)
	assert.Equal(s.T(), promptName, createResult["name"])
	assert.Equal(s.T(), float64(1), createResult["version"])

	// Get prompt by name
	resp, err = s.doRequest("GET", "/api/public/prompts/"+promptName, nil)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusOK, resp.StatusCode)

	// Update prompt (creates new version)
	updateInput := map[string]interface{}{
		"prompt": "You are an expert assistant. Please answer: {{question}}",
	}

	resp, err = s.doRequest("PUT", "/api/public/prompts/"+promptName, updateInput)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusOK, resp.StatusCode)

	var updateResult map[string]interface{}
	s.parseResponse(resp, &updateResult)
	assert.Equal(s.T(), float64(2), updateResult["version"])

	// Compile prompt with variables
	compileInput := map[string]interface{}{
		"variables": map[string]string{
			"question": "What is the capital of France?",
		},
	}

	resp, err = s.doRequest("POST", "/api/public/prompts/"+promptName+"/compile", compileInput)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusOK, resp.StatusCode)

	var compileResult map[string]interface{}
	s.parseResponse(resp, &compileResult)
	assert.Contains(s.T(), compileResult["compiledPrompt"], "What is the capital of France?")

	// List prompts
	resp, err = s.doRequest("GET", "/api/public/prompts", nil)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusOK, resp.StatusCode)
}

// ============ DATASET TESTS ============

func (s *E2ETestSuite) TestDatasetLifecycle() {
	datasetName := fmt.Sprintf("e2e-test-dataset-%d", time.Now().UnixNano())

	// Create dataset
	datasetInput := map[string]interface{}{
		"name":        datasetName,
		"description": "E2E test dataset",
	}

	resp, err := s.doRequest("POST", "/api/public/datasets", datasetInput)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusCreated, resp.StatusCode)

	var createResult map[string]interface{}
	s.parseResponse(resp, &createResult)
	datasetID := createResult["id"].(string)
	assert.NotEmpty(s.T(), datasetID)

	// Add items to dataset
	itemInput := map[string]interface{}{
		"input":          map[string]string{"question": "What is 2+2?"},
		"expectedOutput": map[string]string{"answer": "4"},
	}

	resp, err = s.doRequest("POST", "/api/public/datasets/"+datasetID+"/items", itemInput)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusCreated, resp.StatusCode)

	var itemResult map[string]interface{}
	s.parseResponse(resp, &itemResult)
	itemID := itemResult["id"].(string)
	assert.NotEmpty(s.T(), itemID)

	// Create experiment run
	runInput := map[string]interface{}{
		"name":        "e2e-test-run",
		"description": "E2E experiment run",
	}

	resp, err = s.doRequest("POST", "/api/public/datasets/"+datasetID+"/runs", runInput)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusCreated, resp.StatusCode)

	var runResult map[string]interface{}
	s.parseResponse(resp, &runResult)
	runID := runResult["id"].(string)
	assert.NotEmpty(s.T(), runID)

	// Get dataset
	resp, err = s.doRequest("GET", "/api/public/datasets/"+datasetID, nil)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusOK, resp.StatusCode)

	// List datasets
	resp, err = s.doRequest("GET", "/api/public/datasets", nil)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusOK, resp.StatusCode)
}

// ============ BATCH INGESTION TESTS ============

func (s *E2ETestSuite) TestBatchIngestion() {
	// Langfuse-compatible batch ingestion
	batchInput := map[string]interface{}{
		"batch": []map[string]interface{}{
			{
				"type": "trace-create",
				"body": map[string]interface{}{
					"id":   "e2e-batch-trace-1",
					"name": "batch-test-trace",
				},
			},
			{
				"type": "observation-create",
				"body": map[string]interface{}{
					"id":      "e2e-batch-span-1",
					"traceId": "e2e-batch-trace-1",
					"type":    "SPAN",
					"name":    "batch-test-span",
				},
			},
			{
				"type": "observation-create",
				"body": map[string]interface{}{
					"id":                  "e2e-batch-gen-1",
					"traceId":             "e2e-batch-trace-1",
					"parentObservationId": "e2e-batch-span-1",
					"type":                "GENERATION",
					"name":                "batch-test-generation",
					"model":               "gpt-3.5-turbo",
				},
			},
			{
				"type": "score-create",
				"body": map[string]interface{}{
					"traceId": "e2e-batch-trace-1",
					"name":    "quality",
					"value":   0.8,
				},
			},
		},
	}

	resp, err := s.doRequest("POST", "/api/public/ingestion", batchInput)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	s.parseResponse(resp, &result)
	successes := result["successes"].([]interface{})
	assert.Equal(s.T(), 4, len(successes))

	// Verify trace was created
	time.Sleep(500 * time.Millisecond) // Wait for async processing

	resp, err = s.doRequest("GET", "/api/public/traces/e2e-batch-trace-1", nil)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusOK, resp.StatusCode)
}

// ============ CHECKPOINT TESTS ============

func (s *E2ETestSuite) TestCheckpointLifecycle() {
	// Create trace first
	traceResp, err := s.doRequest("POST", "/api/public/traces", map[string]interface{}{
		"name": "e2e-checkpoint-test-trace",
	})
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusCreated, traceResp.StatusCode)

	var traceResult map[string]interface{}
	s.parseResponse(traceResp, &traceResult)
	traceID := traceResult["id"].(string)

	// Create checkpoint
	checkpointInput := map[string]interface{}{
		"traceId":      traceID,
		"name":         "e2e-test-checkpoint",
		"storageType":  "inline",
		"fileManifest": []string{"main.go", "utils.go"},
		"gitBranch":    "main",
		"gitCommitSha": "abc123def456",
	}

	resp, err := s.doRequest("POST", "/api/public/checkpoints", checkpointInput)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusCreated, resp.StatusCode)

	var cpResult map[string]interface{}
	s.parseResponse(resp, &cpResult)
	cpID := cpResult["id"].(string)
	assert.NotEmpty(s.T(), cpID)

	// Get checkpoint
	resp, err = s.doRequest("GET", "/api/public/checkpoints/"+cpID, nil)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusOK, resp.StatusCode)

	// List checkpoints
	resp, err = s.doRequest("GET", "/api/public/checkpoints?traceId="+traceID, nil)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusOK, resp.StatusCode)
}

// ============ GIT LINK TESTS ============

func (s *E2ETestSuite) TestGitLinkLifecycle() {
	// Create trace first
	traceResp, err := s.doRequest("POST", "/api/public/traces", map[string]interface{}{
		"name": "e2e-gitlink-test-trace",
	})
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusCreated, traceResp.StatusCode)

	var traceResult map[string]interface{}
	s.parseResponse(traceResp, &traceResult)
	traceID := traceResult["id"].(string)

	// Create git link
	gitLinkInput := map[string]interface{}{
		"traceId":     traceID,
		"repository":  "github.com/agenttrace/agenttrace",
		"commitSha":   "abc123def456",
		"branch":      "main",
		"prNumber":    42,
		"prTitle":     "Add new feature",
		"changedFiles": []string{"main.go", "utils.go"},
	}

	resp, err := s.doRequest("POST", "/api/public/git-links", gitLinkInput)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusCreated, resp.StatusCode)

	var linkResult map[string]interface{}
	s.parseResponse(resp, &linkResult)
	assert.NotEmpty(s.T(), linkResult["id"])
}

// ============ ERROR HANDLING TESTS ============

func (s *E2ETestSuite) TestUnauthorizedAccess() {
	req, err := http.NewRequest("GET", s.baseURL+"/api/public/traces", nil)
	require.NoError(s.T(), err)
	// No auth header

	resp, err := s.client.Do(req)
	require.NoError(s.T(), err)
	defer resp.Body.Close()

	assert.Equal(s.T(), http.StatusUnauthorized, resp.StatusCode)
}

func (s *E2ETestSuite) TestInvalidAPIKey() {
	req, err := http.NewRequest("GET", s.baseURL+"/api/public/traces", nil)
	require.NoError(s.T(), err)
	req.Header.Set("Authorization", "Bearer invalid-key")

	resp, err := s.client.Do(req)
	require.NoError(s.T(), err)
	defer resp.Body.Close()

	assert.Equal(s.T(), http.StatusUnauthorized, resp.StatusCode)
}

func (s *E2ETestSuite) TestNotFound() {
	resp, err := s.doRequest("GET", "/api/public/traces/nonexistent-trace-id", nil)
	require.NoError(s.T(), err)
	defer resp.Body.Close()

	assert.Equal(s.T(), http.StatusNotFound, resp.StatusCode)
}

func (s *E2ETestSuite) TestInvalidInput() {
	// Missing required field
	invalidInput := map[string]interface{}{
		// name is missing
	}

	resp, err := s.doRequest("POST", "/api/public/traces", invalidInput)
	require.NoError(s.T(), err)
	defer resp.Body.Close()

	// Should either succeed with generated name or return 400
	assert.True(s.T(), resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusBadRequest)
}

// ============ PAGINATION TESTS ============

func (s *E2ETestSuite) TestTracePagination() {
	// Create multiple traces
	for i := 0; i < 5; i++ {
		_, err := s.doRequest("POST", "/api/public/traces", map[string]interface{}{
			"name":   fmt.Sprintf("e2e-pagination-trace-%d", i),
			"userId": "e2e-pagination-user",
		})
		require.NoError(s.T(), err)
	}

	// Get first page
	resp, err := s.doRequest("GET", "/api/public/traces?userId=e2e-pagination-user&limit=2", nil)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusOK, resp.StatusCode)

	var page1 map[string]interface{}
	s.parseResponse(resp, &page1)
	data := page1["data"].([]interface{})
	assert.Equal(s.T(), 2, len(data))

	meta := page1["meta"].(map[string]interface{})
	assert.True(s.T(), meta["hasMore"].(bool))
	nextCursor := meta["nextCursor"].(string)
	assert.NotEmpty(s.T(), nextCursor)

	// Get second page
	resp, err = s.doRequest("GET", "/api/public/traces?userId=e2e-pagination-user&limit=2&cursor="+nextCursor, nil)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusOK, resp.StatusCode)

	var page2 map[string]interface{}
	s.parseResponse(resp, &page2)
	data2 := page2["data"].([]interface{})
	assert.GreaterOrEqual(s.T(), len(data2), 1)
}
