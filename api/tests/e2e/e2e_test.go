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
	"strings"
	"testing"
	"time"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/google/uuid"
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
	require.Equal(s.T(), http.StatusOK, resp.StatusCode)

	var result struct {
		Status  string            `json:"status"`
		Version string            `json:"version"`
		Checks  map[string]string `json:"checks"`
	}
	s.parseResponse(resp, &result)
	assert.Equal(s.T(), "healthy", result.Status)
	assert.NotEmpty(s.T(), result.Version)
	assert.Equal(s.T(), "healthy", result.Checks["postgres"])
	assert.Equal(s.T(), "healthy", result.Checks["clickhouse"])
	assert.Equal(s.T(), "healthy", result.Checks["redis"])
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
	require.Equal(s.T(), http.StatusCreated, resp.StatusCode)

	var createResult domain.Trace
	s.parseResponse(resp, &createResult)
	traceID := createResult.ID
	require.NotEmpty(s.T(), traceID)
	assert.Equal(s.T(), domain.LevelDefault, createResult.Level)
	assert.JSONEq(s.T(), `{"query":"test input"}`, createResult.Input)
	assert.JSONEq(s.T(), `{"environment":"test"}`, createResult.Metadata)

	// Get the trace
	resp, err = s.doRequest("GET", "/api/public/traces/"+traceID, nil)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusOK, resp.StatusCode)

	var getResult domain.Trace
	s.parseResponse(resp, &getResult)
	assert.Equal(s.T(), traceID, getResult.ID)
	assert.Equal(s.T(), "e2e-test-trace", getResult.Name)
	assert.Equal(s.T(), "e2e-test-user", getResult.UserID)
	assert.ElementsMatch(s.T(), []string{"e2e", "test"}, getResult.Tags)

	// Update the trace
	updateInput := map[string]interface{}{
		"output": map[string]string{"result": "test output"},
	}

	resp, err = s.doRequest("PATCH", "/api/public/traces/"+traceID, updateInput)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusOK, resp.StatusCode)

	var updateResult domain.Trace
	s.parseResponse(resp, &updateResult)
	assert.JSONEq(s.T(), `{"result":"test output"}`, updateResult.Output)

	// List traces with filter
	resp, err = s.doRequest("GET", "/api/public/traces?userId=e2e-test-user", nil)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusOK, resp.StatusCode)

	var listResult domain.TraceList
	s.parseResponse(resp, &listResult)
	assert.GreaterOrEqual(s.T(), listResult.TotalCount, int64(1))
	assert.Condition(s.T(), func() bool {
		for _, trace := range listResult.Traces {
			if trace.ID == traceID {
				return true
			}
		}
		return false
	}, "created trace should be present in the filtered list")
}

// ============ OBSERVATION TESTS ============

func (s *E2ETestSuite) TestObservationLifecycle() {
	// First create a trace
	traceResp, err := s.doRequest("POST", "/api/public/traces", map[string]interface{}{
		"name": "e2e-observation-test-trace",
	})
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusCreated, traceResp.StatusCode)

	var traceResult domain.Trace
	s.parseResponse(traceResp, &traceResult)
	traceID := traceResult.ID

	// Create a span observation
	spanInput := map[string]interface{}{
		"traceId": traceID,
		"name":    "e2e-test-span",
		"input":   map[string]string{"data": "span input"},
	}

	resp, err := s.doRequest("POST", "/api/public/spans", spanInput)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusCreated, resp.StatusCode)

	var spanResult domain.Observation
	s.parseResponse(resp, &spanResult)
	spanID := spanResult.ID
	require.NotEmpty(s.T(), spanID)
	assert.Equal(s.T(), domain.ObservationTypeSpan, spanResult.Type)
	assert.JSONEq(s.T(), `{"data":"span input"}`, spanResult.Input)

	// Create a generation observation
	genInput := map[string]interface{}{
		"traceId":             traceID,
		"parentObservationId": spanID,
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

	resp, err = s.doRequest("POST", "/api/public/generations", genInput)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusCreated, resp.StatusCode)

	var genResult domain.Observation
	s.parseResponse(resp, &genResult)
	genID := genResult.ID
	require.NotEmpty(s.T(), genID)
	assert.Equal(s.T(), domain.ObservationTypeGeneration, genResult.Type)
	assert.Equal(s.T(), "gpt-4", genResult.Model)
	assert.Equal(s.T(), uint64(10), genResult.UsageDetails.InputTokens)
	assert.Equal(s.T(), uint64(5), genResult.UsageDetails.OutputTokens)
	assert.Equal(s.T(), uint64(15), genResult.UsageDetails.TotalTokens)

	// List observations for trace
	resp, err = s.doRequest("GET", "/api/public/traces/"+traceID+"/observations", nil)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusOK, resp.StatusCode)

	var tree domain.ObservationTree
	s.parseResponse(resp, &tree)
	require.NotNil(s.T(), tree.Observation)
	assert.Equal(s.T(), spanID, tree.Observation.ID)
	require.Len(s.T(), tree.Children, 1)
	require.NotNil(s.T(), tree.Children[0].Observation)
	assert.Equal(s.T(), genID, tree.Children[0].Observation.ID)
	assert.Equal(s.T(), "gpt-4", tree.Children[0].Observation.Model)
}

// ============ SCORE TESTS ============

func (s *E2ETestSuite) TestScoreLifecycle() {
	// Create a trace first
	traceResp, err := s.doRequest("POST", "/api/public/traces", map[string]interface{}{
		"name": "e2e-score-test-trace",
	})
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusCreated, traceResp.StatusCode)

	var traceResult domain.Trace
	s.parseResponse(traceResp, &traceResult)
	traceID := traceResult.ID

	// Create a numeric score
	scoreInput := map[string]interface{}{
		"traceId": traceID,
		"name":    "quality",
		"value":   0.95,
		"comment": "High quality response",
	}

	resp, err := s.doRequest("POST", "/api/public/scores", scoreInput)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusCreated, resp.StatusCode)

	var scoreResult domain.Score
	s.parseResponse(resp, &scoreResult)
	require.NotEqual(s.T(), uuid.Nil, scoreResult.ID)
	require.NotNil(s.T(), scoreResult.Value)
	assert.Equal(s.T(), 0.95, *scoreResult.Value)
	assert.Equal(s.T(), domain.ScoreSourceAPI, scoreResult.Source)
	assert.Equal(s.T(), domain.ScoreDataTypeNumeric, scoreResult.DataType)

	// Create a categorical score
	catScoreInput := map[string]interface{}{
		"traceId":     traceID,
		"name":        "sentiment",
		"stringValue": "positive",
	}

	resp, err = s.doRequest("POST", "/api/public/scores", catScoreInput)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusCreated, resp.StatusCode)

	var categoricalScore domain.Score
	s.parseResponse(resp, &categoricalScore)
	require.NotNil(s.T(), categoricalScore.StringValue)
	assert.Equal(s.T(), "positive", *categoricalScore.StringValue)
	assert.Equal(s.T(), domain.ScoreDataTypeCategorical, categoricalScore.DataType)

	// Get score by ID
	resp, err = s.doRequest("GET", "/api/public/scores/"+scoreResult.ID.String(), nil)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusOK, resp.StatusCode)

	var persistedScore domain.Score
	s.parseResponse(resp, &persistedScore)
	assert.Equal(s.T(), scoreResult.ID, persistedScore.ID)
	assert.Equal(s.T(), "quality", persistedScore.Name)

	// List scores for trace
	resp, err = s.doRequest("GET", "/api/public/scores?traceId="+traceID, nil)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusOK, resp.StatusCode)

	var listResult domain.ScoreList
	s.parseResponse(resp, &listResult)
	assert.Equal(s.T(), int64(2), listResult.TotalCount)
	assert.Len(s.T(), listResult.Scores, 2)
	assert.False(s.T(), listResult.HasMore)
}

// ============ PROMPT TESTS ============

func (s *E2ETestSuite) TestPromptLifecycle() {
	promptName := fmt.Sprintf("e2e-test-prompt-%d", time.Now().UnixNano())

	// Create a text prompt
	promptInput := map[string]interface{}{
		"name":        promptName,
		"type":        "text",
		"content":     "You are a helpful assistant. Answer the following: {{question}}",
		"description": "E2E test prompt",
		"labels":      []string{"production"},
	}

	resp, err := s.doRequest("POST", "/api/public/prompts", promptInput)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusCreated, resp.StatusCode)

	var createResult domain.Prompt
	s.parseResponse(resp, &createResult)
	assert.Equal(s.T(), promptName, createResult.Name)
	require.NotNil(s.T(), createResult.LatestVersion)
	assert.Equal(s.T(), 1, createResult.LatestVersion.Version)
	assert.Equal(s.T(), promptInput["content"], createResult.LatestVersion.Content)
	assert.Equal(s.T(), []string{"production"}, createResult.LatestVersion.Labels)
	assert.Equal(s.T(), "{}", createResult.LatestVersion.Config)

	// Get prompt by name
	resp, err = s.doRequest("GET", "/api/public/prompts/"+promptName, nil)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusOK, resp.StatusCode)

	var getResult domain.Prompt
	s.parseResponse(resp, &getResult)
	require.NotNil(s.T(), getResult.LatestVersion)
	assert.Equal(s.T(), 1, getResult.LatestVersion.Version)

	// Create a new prompt version
	updateInput := map[string]interface{}{
		"content":       "You are an expert assistant. Please answer: {{question}}",
		"labels":        []string{"latest"},
		"commitMessage": "Improve system instruction",
	}

	resp, err = s.doRequest("POST", "/api/public/prompts/"+promptName+"/versions", updateInput)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusCreated, resp.StatusCode)

	var updateResult domain.PromptVersion
	s.parseResponse(resp, &updateResult)
	assert.Equal(s.T(), 2, updateResult.Version)
	assert.Equal(s.T(), updateInput["content"], updateResult.Content)
	assert.Equal(s.T(), []string{"latest"}, updateResult.Labels)

	// Compile prompt with variables
	compileInput := map[string]interface{}{
		"variables": map[string]string{
			"question": "What is the capital of France?",
		},
		"version": 2,
	}

	resp, err = s.doRequest("POST", "/api/public/prompts/"+promptName+"/compile", compileInput)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusOK, resp.StatusCode)

	var compileResult struct {
		Version   int               `json:"version"`
		Compiled  string            `json:"compiled"`
		Variables map[string]string `json:"variables"`
	}
	s.parseResponse(resp, &compileResult)
	assert.Equal(s.T(), 2, compileResult.Version)
	assert.Contains(s.T(), compileResult.Compiled, "What is the capital of France?")
	assert.Equal(s.T(), "What is the capital of France?", compileResult.Variables["question"])

	// List prompts
	resp, err = s.doRequest("GET", "/api/public/prompts", nil)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusOK, resp.StatusCode)

	var listResult domain.PromptList
	s.parseResponse(resp, &listResult)
	assert.Condition(s.T(), func() bool {
		for _, prompt := range listResult.Prompts {
			if prompt.Name == promptName {
				return true
			}
		}
		return false
	}, "created prompt should be present in the prompt list")
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
	require.Equal(s.T(), http.StatusCreated, resp.StatusCode)

	var createResult domain.Dataset
	s.parseResponse(resp, &createResult)
	datasetID := createResult.ID
	require.NotEqual(s.T(), uuid.Nil, datasetID)
	assert.Equal(s.T(), "{}", createResult.Metadata)

	// Add items to dataset
	itemInput := map[string]interface{}{
		"input":          map[string]string{"question": "What is 2+2?"},
		"expectedOutput": map[string]string{"answer": "4"},
	}

	resp, err = s.doRequest("POST", "/api/public/datasets/"+datasetID.String()+"/items", itemInput)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusCreated, resp.StatusCode)

	var itemResult domain.DatasetItem
	s.parseResponse(resp, &itemResult)
	itemID := itemResult.ID
	require.NotEqual(s.T(), uuid.Nil, itemID)
	assert.JSONEq(s.T(), `{"question":"What is 2+2?"}`, itemResult.Input)
	require.NotNil(s.T(), itemResult.ExpectedOutput)
	assert.JSONEq(s.T(), `{"answer":"4"}`, *itemResult.ExpectedOutput)
	assert.Equal(s.T(), "{}", itemResult.Metadata)

	// Create experiment run
	runInput := map[string]interface{}{
		"name":        "e2e-test-run",
		"description": "E2E experiment run",
	}

	resp, err = s.doRequest("POST", "/api/public/datasets/"+datasetID.String()+"/runs", runInput)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusCreated, resp.StatusCode)

	var runResult domain.DatasetRun
	s.parseResponse(resp, &runResult)
	require.NotEqual(s.T(), uuid.Nil, runResult.ID)
	assert.Equal(s.T(), "{}", runResult.Metadata)

	// Get dataset
	resp, err = s.doRequest("GET", "/api/public/datasets/"+datasetID.String(), nil)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusOK, resp.StatusCode)

	var persisted domain.Dataset
	s.parseResponse(resp, &persisted)
	assert.Equal(s.T(), datasetID, persisted.ID)
	assert.Equal(s.T(), datasetName, persisted.Name)

	// List datasets
	resp, err = s.doRequest("GET", "/api/public/datasets", nil)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusOK, resp.StatusCode)

	var listResult domain.DatasetList
	s.parseResponse(resp, &listResult)
	assert.Condition(s.T(), func() bool {
		for _, dataset := range listResult.Datasets {
			if dataset.ID == datasetID {
				return true
			}
		}
		return false
	}, "created dataset should be present in the dataset list")

	resp, err = s.doRequest("GET", "/api/public/datasets/"+datasetID.String()+"/items", nil)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusOK, resp.StatusCode)

	var items struct {
		Data       []domain.DatasetItem `json:"data"`
		TotalCount int64                `json:"totalCount"`
	}
	s.parseResponse(resp, &items)
	assert.Equal(s.T(), int64(1), items.TotalCount)
	require.Len(s.T(), items.Data, 1)
	assert.Equal(s.T(), itemID, items.Data[0].ID)
}

// ============ BATCH INGESTION TESTS ============

func (s *E2ETestSuite) TestBatchIngestion() {
	traceID := strings.ReplaceAll(uuid.NewString(), "-", "")
	spanID := traceID[:16]
	generationID := traceID[16:]

	// Langfuse-compatible batch ingestion
	batchInput := map[string]interface{}{
		"batch": []map[string]interface{}{
			{
				"id":   "trace-event",
				"type": "trace-create",
				"body": map[string]interface{}{
					"id":   traceID,
					"name": "batch-test-trace",
				},
			},
			{
				"id":   "span-event",
				"type": "span-create",
				"body": map[string]interface{}{
					"id":      spanID,
					"traceId": traceID,
					"name":    "batch-test-span",
				},
			},
			{
				"id":   "generation-event",
				"type": "generation-create",
				"body": map[string]interface{}{
					"id":                  generationID,
					"traceId":             traceID,
					"parentObservationId": spanID,
					"name":                "batch-test-generation",
					"model":               "gpt-3.5-turbo",
				},
			},
			{
				"id":   "score-event",
				"type": "score-create",
				"body": map[string]interface{}{
					"traceId": traceID,
					"name":    "quality",
					"value":   0.8,
				},
			},
		},
	}

	resp, err := s.doRequest("POST", "/api/public/ingestion", batchInput)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusOK, resp.StatusCode)

	var result struct {
		Successes []string                 `json:"successes"`
		Errors    []map[string]interface{} `json:"errors"`
	}
	s.parseResponse(resp, &result)
	assert.ElementsMatch(s.T(), []string{"trace-event", "span-event", "generation-event", "score-event"}, result.Successes)
	assert.Empty(s.T(), result.Errors)

	// Verify trace was created
	resp, err = s.doRequest("GET", "/api/public/traces/"+traceID, nil)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusOK, resp.StatusCode)

	var trace domain.Trace
	s.parseResponse(resp, &trace)
	assert.Equal(s.T(), "batch-test-trace", trace.Name)

	resp, err = s.doRequest("GET", "/api/public/traces/"+traceID+"/observations", nil)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusOK, resp.StatusCode)

	var tree domain.ObservationTree
	s.parseResponse(resp, &tree)
	require.NotNil(s.T(), tree.Observation)
	assert.Equal(s.T(), spanID, tree.Observation.ID)
	require.Len(s.T(), tree.Children, 1)
	assert.Equal(s.T(), generationID, tree.Children[0].Observation.ID)
}

// ============ CHECKPOINT TESTS ============

func (s *E2ETestSuite) TestCheckpointLifecycle() {
	// Create trace first
	traceResp, err := s.doRequest("POST", "/api/public/traces", map[string]interface{}{
		"name": "e2e-checkpoint-test-trace",
	})
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusCreated, traceResp.StatusCode)

	var traceResult domain.Trace
	s.parseResponse(traceResp, &traceResult)
	traceID := traceResult.ID

	// Create checkpoint
	checkpointInput := map[string]interface{}{
		"traceId":        traceID,
		"name":           "e2e-test-checkpoint",
		"type":           "manual",
		"filesSnapshot":  `{"main.go":{"hash":"abc"},"utils.go":{"hash":"def"}}`,
		"filesChanged":   []string{"main.go", "utils.go"},
		"totalFiles":     2,
		"totalSizeBytes": 256,
		"gitBranch":      "main",
		"gitCommitSha":   "abc123def456",
	}

	resp, err := s.doRequest("POST", "/api/public/checkpoints", checkpointInput)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusCreated, resp.StatusCode)

	var cpResult domain.Checkpoint
	s.parseResponse(resp, &cpResult)
	cpID := cpResult.ID
	require.NotEqual(s.T(), uuid.Nil, cpID)
	assert.Equal(s.T(), domain.CheckpointTypeManual, cpResult.Type)
	assert.Equal(s.T(), []string{"main.go", "utils.go"}, cpResult.FilesChanged)
	assert.Equal(s.T(), uint32(2), cpResult.TotalFiles)
	assert.Equal(s.T(), uint64(256), cpResult.TotalSizeBytes)

	// Get checkpoint
	resp, err = s.doRequest("GET", "/api/public/checkpoints/"+cpID.String(), nil)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusOK, resp.StatusCode)

	var persisted domain.Checkpoint
	s.parseResponse(resp, &persisted)
	assert.Equal(s.T(), cpID, persisted.ID)
	assert.Equal(s.T(), traceID, persisted.TraceID)
	assert.Equal(s.T(), "abc123def456", persisted.GitCommitSha)

	// List checkpoints
	resp, err = s.doRequest("GET", "/api/public/checkpoints?traceId="+traceID, nil)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusOK, resp.StatusCode)

	var listResult struct {
		Data       []domain.Checkpoint `json:"data"`
		TotalCount int64               `json:"totalCount"`
		HasMore    bool                `json:"hasMore"`
	}
	s.parseResponse(resp, &listResult)
	assert.Equal(s.T(), int64(1), listResult.TotalCount)
	require.Len(s.T(), listResult.Data, 1)
	assert.Equal(s.T(), cpID, listResult.Data[0].ID)
	assert.False(s.T(), listResult.HasMore)
}

// ============ GIT LINK TESTS ============

func (s *E2ETestSuite) TestGitLinkLifecycle() {
	// Create trace first
	traceResp, err := s.doRequest("POST", "/api/public/traces", map[string]interface{}{
		"name": "e2e-gitlink-test-trace",
	})
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusCreated, traceResp.StatusCode)

	var traceResult domain.Trace
	s.parseResponse(traceResp, &traceResult)
	traceID := traceResult.ID

	// Create git link
	gitLinkInput := map[string]interface{}{
		"traceId":       traceID,
		"repoUrl":       "https://github.com/agenttrace/agenttrace",
		"commitSha":     "abc123def456",
		"branch":        "main",
		"commitMessage": "Add new feature",
		"filesModified": []string{"main.go", "utils.go"},
		"additions":     42,
		"deletions":     3,
		"linkType":      "current",
	}

	resp, err := s.doRequest("POST", "/api/public/git-links", gitLinkInput)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusCreated, resp.StatusCode)

	var linkResult domain.GitLink
	s.parseResponse(resp, &linkResult)
	require.NotEqual(s.T(), uuid.Nil, linkResult.ID)
	assert.Equal(s.T(), domain.GitLinkTypeCurrent, linkResult.LinkType)
	assert.Equal(s.T(), uint32(2), linkResult.FilesChangedCount)
	assert.Equal(s.T(), []string{"main.go", "utils.go"}, linkResult.FilesModified)
	assert.Equal(s.T(), uint32(42), linkResult.Additions)

	resp, err = s.doRequest("GET", "/api/public/git-links/"+linkResult.ID.String(), nil)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusOK, resp.StatusCode)

	var persisted domain.GitLink
	s.parseResponse(resp, &persisted)
	assert.Equal(s.T(), linkResult.ID, persisted.ID)
	assert.Equal(s.T(), traceID, persisted.TraceID)
	assert.Equal(s.T(), "https://github.com/agenttrace/agenttrace", persisted.RepoURL)

	resp, err = s.doRequest("GET", "/api/public/git-links?traceId="+traceID, nil)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusOK, resp.StatusCode)

	var listResult struct {
		Data       []domain.GitLink `json:"data"`
		TotalCount int64            `json:"totalCount"`
		HasMore    bool             `json:"hasMore"`
	}
	s.parseResponse(resp, &listResult)
	assert.Equal(s.T(), int64(1), listResult.TotalCount)
	require.Len(s.T(), listResult.Data, 1)
	assert.Equal(s.T(), linkResult.ID, listResult.Data[0].ID)
	assert.False(s.T(), listResult.HasMore)
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

func (s *E2ETestSuite) TestTraceDefaults() {
	resp, err := s.doRequest("POST", "/api/public/traces", map[string]interface{}{})
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusCreated, resp.StatusCode)

	var trace domain.Trace
	s.parseResponse(resp, &trace)
	assert.NotEmpty(s.T(), trace.ID)
	assert.Empty(s.T(), trace.Name)
	assert.Equal(s.T(), domain.LevelDefault, trace.Level)
	assert.Equal(s.T(), "{}", trace.Metadata)
	assert.Empty(s.T(), trace.Input)
	assert.Empty(s.T(), trace.Output)
}

// ============ PAGINATION TESTS ============

func (s *E2ETestSuite) TestTracePagination() {
	userID := "e2e-pagination-" + uuid.NewString()
	createdIDs := make(map[string]struct{}, 5)

	// Create multiple traces
	for i := 0; i < 5; i++ {
		resp, err := s.doRequest("POST", "/api/public/traces", map[string]interface{}{
			"name":   fmt.Sprintf("e2e-pagination-trace-%d", i),
			"userId": userID,
		})
		require.NoError(s.T(), err)
		require.Equal(s.T(), http.StatusCreated, resp.StatusCode)

		var trace domain.Trace
		s.parseResponse(resp, &trace)
		createdIDs[trace.ID] = struct{}{}
	}

	// Get first page
	resp, err := s.doRequest("GET", "/api/public/traces?userId="+userID+"&limit=2&offset=0", nil)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusOK, resp.StatusCode)

	var page1 domain.TraceList
	s.parseResponse(resp, &page1)
	assert.Equal(s.T(), int64(5), page1.TotalCount)
	assert.Len(s.T(), page1.Traces, 2)
	assert.True(s.T(), page1.HasMore)

	// Get second page
	resp, err = s.doRequest("GET", "/api/public/traces?userId="+userID+"&limit=2&offset=2", nil)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusOK, resp.StatusCode)

	var page2 domain.TraceList
	s.parseResponse(resp, &page2)
	assert.Equal(s.T(), int64(5), page2.TotalCount)
	assert.Len(s.T(), page2.Traces, 2)
	assert.True(s.T(), page2.HasMore)

	page1IDs := map[string]struct{}{
		page1.Traces[0].ID: {},
		page1.Traces[1].ID: {},
	}
	for _, trace := range page2.Traces {
		_, duplicated := page1IDs[trace.ID]
		assert.False(s.T(), duplicated, "pages should not overlap")
	}

	// Get final page
	resp, err = s.doRequest("GET", "/api/public/traces?userId="+userID+"&limit=2&offset=4", nil)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusOK, resp.StatusCode)

	var page3 domain.TraceList
	s.parseResponse(resp, &page3)
	assert.Equal(s.T(), int64(5), page3.TotalCount)
	require.Len(s.T(), page3.Traces, 1)
	assert.False(s.T(), page3.HasMore)

	for _, page := range [][]domain.Trace{page1.Traces, page2.Traces, page3.Traces} {
		for _, trace := range page {
			_, created := createdIDs[trace.ID]
			assert.True(s.T(), created, "paginated result should contain only traces created by this test")
		}
	}
}

// ================================================================
// Next-Gen Feature E2E Tests
// ================================================================

// TestStreamingEndpoints tests the real-time streaming API
func (s *E2ETestSuite) TestStreamingEndpoints() {
	// GET /api/public/streams
	resp, err := s.doRequest("GET", "/api/public/streams", nil)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusOK, resp.StatusCode)

	var streamsResp struct {
		Streams []json.RawMessage `json:"streams"`
		Count   int               `json:"count"`
	}
	s.parseResponse(resp, &streamsResp)
	assert.NotNil(s.T(), streamsResp.Streams)
	assert.Equal(s.T(), len(streamsResp.Streams), streamsResp.Count)
}

// TestDiffIntelligenceLifecycle tests the diff analysis API
func (s *E2ETestSuite) TestDiffIntelligenceLifecycle() {
	// POST /api/public/diff-analysis - Create analysis
	input := map[string]interface{}{
		"traceId": uuid.New().String(),
		"fileChanges": []map[string]interface{}{
			{
				"filePath":     "main.go",
				"operation":    "modify",
				"contentAfter": "package main\nfunc main() {}\n",
				"diff":         "+func main() {}\n",
			},
		},
	}

	resp, err := s.doRequest("POST", "/api/public/diff-analysis", input)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusCreated, resp.StatusCode)

	var analysis map[string]interface{}
	s.parseResponse(resp, &analysis)

	analysisID, ok := analysis["id"].(string)
	require.True(s.T(), ok, "analysis should have an id")
	assert.Equal(s.T(), "completed", analysis["status"])
	assert.NotNil(s.T(), analysis["overallScore"])

	// GET /api/public/diff-analysis/:id
	resp, err = s.doRequest("GET", "/api/public/diff-analysis/"+analysisID, nil)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusOK, resp.StatusCode)
	defer resp.Body.Close()

	// GET /api/public/diff-analysis
	resp, err = s.doRequest("GET", "/api/public/diff-analysis?limit=10", nil)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusOK, resp.StatusCode)
	defer resp.Body.Close()

	// GET /api/public/diff-analysis/trend
	resp, err = s.doRequest("GET", "/api/public/diff-analysis/trend?days=30", nil)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusOK, resp.StatusCode)
	defer resp.Body.Close()
}

// TestGuardrailPlaybooks tests guardrail playbook endpoints
func (s *E2ETestSuite) TestGuardrailPlaybooks() {
	// GET /api/public/guardrails/templates
	resp, err := s.doRequest("GET", "/api/public/guardrails/templates", nil)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusOK, resp.StatusCode)

	var templatesResp map[string]interface{}
	s.parseResponse(resp, &templatesResp)
	templates, ok := templatesResp["templates"].([]interface{})
	require.True(s.T(), ok)
	assert.Greater(s.T(), len(templates), 0, "should have at least one template")

	// POST /api/public/guardrails/playbooks
	playbook := map[string]interface{}{
		"name":     "test-playbook",
		"template": "production-safe",
	}
	resp, err = s.doRequest("POST", "/api/public/guardrails/playbooks", playbook)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusCreated, resp.StatusCode)

	var playbookResp map[string]interface{}
	s.parseResponse(resp, &playbookResp)
	assert.Equal(s.T(), "test-playbook", playbookResp["name"])
	assert.True(s.T(), playbookResp["enabled"].(bool))
}

// TestCollaborationDiscussions tests discussion thread endpoints
func (s *E2ETestSuite) TestCollaborationDiscussions() {
	// POST /api/public/collaboration/discussions
	input := map[string]interface{}{
		"traceId":        uuid.New().String(),
		"title":          "Test Discussion",
		"initialMessage": "This is a test discussion thread",
	}
	resp, err := s.doRequest("POST", "/api/public/collaboration/discussions", input)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusCreated, resp.StatusCode)

	var thread map[string]interface{}
	s.parseResponse(resp, &thread)
	assert.Equal(s.T(), "Test Discussion", thread["title"])
	assert.Equal(s.T(), "open", thread["status"])

	// POST /api/public/collaboration/eval-queues
	queueInput := map[string]interface{}{
		"name":     "test-eval-queue",
		"traceIds": []string{uuid.New().String()},
	}
	resp, err = s.doRequest("POST", "/api/public/collaboration/eval-queues", queueInput)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusCreated, resp.StatusCode)
	defer resp.Body.Close()
}

// TestAnomalyDashboard tests anomaly detection endpoints
func (s *E2ETestSuite) TestAnomalyDashboard() {
	// GET /api/public/anomaly/dashboard
	resp, err := s.doRequest("GET", "/api/public/anomaly/dashboard", nil)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusOK, resp.StatusCode)

	var dashboard map[string]interface{}
	s.parseResponse(resp, &dashboard)
	assert.NotNil(s.T(), dashboard["healthScore"])

	// POST /api/public/anomaly/channels
	channel := map[string]interface{}{
		"name": "slack-alerts",
		"type": "slack",
		"config": map[string]string{
			"webhookUrl": "https://hooks.slack.com/test",
		},
	}
	resp, err = s.doRequest("POST", "/api/public/anomaly/channels", channel)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusCreated, resp.StatusCode)
	defer resp.Body.Close()

	// GET /api/public/anomaly/anomalies/:id/root-cause (test with random ID)
	resp, err = s.doRequest("GET", "/api/public/anomaly/anomalies/"+uuid.New().String()+"/root-cause", nil)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusOK, resp.StatusCode)
	defer resp.Body.Close()
}

// TestBenchmarkLifecycle tests benchmark endpoints
func (s *E2ETestSuite) TestBenchmarkLifecycle() {
	datasetName := "e2e-benchmark-dataset-" + uuid.NewString()
	resp, err := s.doRequest("POST", "/api/public/datasets", map[string]interface{}{
		"name": datasetName,
	})
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusCreated, resp.StatusCode)

	var dataset domain.Dataset
	s.parseResponse(resp, &dataset)
	require.NotEqual(s.T(), uuid.Nil, dataset.ID)

	// POST /api/public/benchmarks
	benchmark := map[string]interface{}{
		"name":        "code-gen-benchmark-" + uuid.NewString(),
		"description": "Test benchmark for code generation",
		"category":    "code_generation",
		"datasetId":   dataset.ID.String(),
		"metrics": []map[string]interface{}{
			{"name": "accuracy", "weight": 0.5, "higherIsBetter": true},
			{"name": "speed", "weight": 0.3, "higherIsBetter": false},
			{"name": "cost", "weight": 0.2, "higherIsBetter": false},
		},
		"isPublic": true,
	}
	resp, err = s.doRequest("POST", "/api/public/benchmarks", benchmark)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusCreated, resp.StatusCode)

	var created domain.Benchmark
	s.parseResponse(resp, &created)
	require.NotEqual(s.T(), uuid.Nil, created.ID)
	assert.Equal(s.T(), dataset.ID, created.DatasetID)
	assert.Empty(s.T(), created.EvaluatorIDs)
	assert.Len(s.T(), created.Metrics, 3)
	assert.True(s.T(), created.IsPublic)

	// GET /api/public/benchmarks
	resp, err = s.doRequest("GET", "/api/public/benchmarks", nil)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusOK, resp.StatusCode)

	var benchmarks []domain.Benchmark
	s.parseResponse(resp, &benchmarks)
	assert.Condition(s.T(), func() bool {
		for _, listed := range benchmarks {
			if listed.ID == created.ID {
				return true
			}
		}
		return false
	}, "created benchmark should be present in the benchmark list")

	resp, err = s.doRequest("GET", "/api/public/benchmarks/"+created.ID.String(), nil)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusOK, resp.StatusCode)

	var persisted domain.Benchmark
	s.parseResponse(resp, &persisted)
	assert.Equal(s.T(), created.ID, persisted.ID)
	assert.Equal(s.T(), dataset.ID, persisted.DatasetID)
}

// TestCostOptimizerAutopilot tests cost optimization endpoints
func (s *E2ETestSuite) TestCostOptimizerAutopilot() {
	// GET /api/public/cost-optimizer/forecast
	resp, err := s.doRequest("GET", "/api/public/cost-optimizer/forecast", nil)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusOK, resp.StatusCode)

	var forecast map[string]interface{}
	s.parseResponse(resp, &forecast)
	assert.NotNil(s.T(), forecast["projectedMonthlyCost"])

	// POST /api/public/cost-optimizer/autopilot
	config := map[string]interface{}{
		"enabled":           true,
		"maxBudgetDaily":    50.0,
		"maxBudgetMonthly":  1000.0,
		"optimizationLevel": "balanced",
	}
	resp, err = s.doRequest("POST", "/api/public/cost-optimizer/autopilot", config)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusCreated, resp.StatusCode)

	var autopilot map[string]interface{}
	s.parseResponse(resp, &autopilot)
	assert.True(s.T(), autopilot["enabled"].(bool))
	assert.Equal(s.T(), "balanced", autopilot["optimizationLevel"])

	// POST /api/public/cost-optimizer/report
	resp, err = s.doRequest("POST", "/api/public/cost-optimizer/report", map[string]interface{}{})
	require.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusOK, resp.StatusCode)
	defer resp.Body.Close()
}

// TestReplayReproduction tests trace reproduction endpoints
func (s *E2ETestSuite) TestReplayReproduction() {
	// Create a trace first
	traceResp, err := s.doRequest("POST", "/api/public/traces", map[string]interface{}{
		"name": "repro-test-trace",
	})
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusCreated, traceResp.StatusCode)

	var traceResult domain.Trace
	s.parseResponse(traceResp, &traceResult)
	traceID := traceResult.ID
	require.NotEmpty(s.T(), traceID)

	// POST /api/public/traces/:traceId/reproduce
	reproInput := map[string]interface{}{
		"format": "python",
		"config": map[string]interface{}{
			"includeEnvironment": true,
			"deterministicMode":  true,
		},
	}
	resp, err := s.doRequest("POST", "/api/public/traces/"+traceID+"/reproduce", reproInput)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusCreated, resp.StatusCode)

	var script domain.ReproductionScript
	s.parseResponse(resp, &script)
	assert.Equal(s.T(), domain.ReproFormatPython, script.Format)
	assert.Equal(s.T(), "python", script.Language)
	assert.True(s.T(), script.Config.IncludeEnvironment)
	assert.True(s.T(), script.Config.DeterministicMode)
	assert.Contains(s.T(), script.Script, "repro-test-trace")
	assert.Contains(s.T(), script.Script, "Reproduction complete")
}

// TestFederationLifecycle tests federation endpoints
func (s *E2ETestSuite) TestFederationLifecycle() {
	// POST /api/public/federation/peers
	peer := map[string]interface{}{
		"name": "staging-cluster",
		"url":  "https://staging.agenttrace.example.com",
	}
	resp, err := s.doRequest("POST", "/api/public/federation/peers", peer)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusCreated, resp.StatusCode)

	var peerResp map[string]interface{}
	s.parseResponse(resp, &peerResp)
	peerID, _ := peerResp["id"].(string)
	require.NotEmpty(s.T(), peerID)

	// GET /api/public/federation/peers
	resp, err = s.doRequest("GET", "/api/public/federation/peers", nil)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusOK, resp.StatusCode)
	defer resp.Body.Close()

	// POST /api/public/federation/destinations
	dest := map[string]interface{}{
		"name":     "datadog-export",
		"type":     "datadog",
		"endpoint": "https://trace.agent.datadoghq.com",
		"protocol": "grpc",
	}
	resp, err = s.doRequest("POST", "/api/public/federation/destinations", dest)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusCreated, resp.StatusCode)
	defer resp.Body.Close()

	// GET /api/public/federation/destinations
	resp, err = s.doRequest("GET", "/api/public/federation/destinations", nil)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusOK, resp.StatusCode)
	defer resp.Body.Close()

	// POST /api/public/federation/query
	query := map[string]interface{}{
		"query": "high-cost traces",
	}
	resp, err = s.doRequest("POST", "/api/public/federation/query", query)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusOK, resp.StatusCode)
	defer resp.Body.Close()

	// DELETE /api/public/federation/peers/:id
	resp, err = s.doRequest("DELETE", "/api/public/federation/peers/"+peerID, nil)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusOK, resp.StatusCode)
	defer resp.Body.Close()
}
