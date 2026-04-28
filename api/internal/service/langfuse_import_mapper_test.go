package service

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	apperrors "github.com/agenttrace/agenttrace/api/internal/pkg/errors"
)

func usageInt64Ptr(v int64) *int64 { return &v }

func TestMapLangfuseUsage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		raw     string
		want    *domain.UsageDetailsInput
		wantErr bool
	}{
		{name: "empty", raw: "", want: nil},
		{name: "null", raw: "null", want: nil},
		{name: "no token fields", raw: `{"unit":"TOKENS"}`, want: nil},
		{
			name: "input/output/total",
			raw:  `{"input":10,"output":20,"total":30}`,
			want: &domain.UsageDetailsInput{
				InputTokens:  usageInt64Ptr(10),
				OutputTokens: usageInt64Ptr(20),
				TotalTokens:  usageInt64Ptr(30),
			},
		},
		{
			name: "inputTokens/outputTokens/totalTokens",
			raw:  `{"inputTokens":11,"outputTokens":22,"totalTokens":33}`,
			want: &domain.UsageDetailsInput{
				InputTokens:  usageInt64Ptr(11),
				OutputTokens: usageInt64Ptr(22),
				TotalTokens:  usageInt64Ptr(33),
			},
		},
		{
			name: "promptTokens/completionTokens",
			raw:  `{"promptTokens":5,"completionTokens":7}`,
			want: &domain.UsageDetailsInput{
				InputTokens:  usageInt64Ptr(5),
				OutputTokens: usageInt64Ptr(7),
				TotalTokens:  nil,
			},
		},
		{
			name: "canonical fields take precedence over aliases",
			raw:  `{"input":1,"inputTokens":2,"promptTokens":3,"output":4,"outputTokens":5}`,
			want: &domain.UsageDetailsInput{
				InputTokens:  usageInt64Ptr(1),
				OutputTokens: usageInt64Ptr(4),
				TotalTokens:  nil,
			},
		},
		{name: "negative input rejected", raw: `{"input":-1}`, wantErr: true},
		{name: "negative completionTokens rejected", raw: `{"completionTokens":-5}`, wantErr: true},
		{name: "invalid json rejected", raw: `{"input":`, wantErr: true},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := mapLangfuseUsage(json.RawMessage(tc.raw))
			if tc.wantErr {
				require.Error(t, err)
				assert.True(t, apperrors.IsValidation(err), "expected validation error, got %v", err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

// TestLangfuseImportForwardsGenerationUsage proves the parsed token counts reach
// IngestGeneration through the OUTER GenerationInput.Usage field, guarding
// against the embedded-field shadowing regression.
func TestLangfuseImportForwardsGenerationUsage(t *testing.T) {
	projectID := uuid.New()
	jobID := uuid.New()
	repository := newLangfuseImportRepositoryStub()
	ingestion := &langfuseIngestionStub{}
	service := newLangfuseTestService(repository, ingestion, &langfuseScoreStub{}, &langfusePromptStub{})

	batch := domain.LangfuseImportBatch{
		JobID:       jobID,
		Fingerprint: "0123456789abcdef0123456789abcdef",
		TotalItems:  2,
		FinalBatch:  true,
		Records: domain.LangfuseExport{
			Traces: []domain.LangfuseTrace{{
				ID:        "trace-source-id",
				Name:      "agent",
				StartTime: "2026-07-25T10:00:00Z",
			}},
			Observations: []domain.LangfuseObservation{{
				ID:        "observation-source-id",
				TraceID:   "trace-source-id",
				Type:      "GENERATION",
				StartTime: "2026-07-25T10:00:01Z",
				Model:     "gpt-4.1",
				Usage:     json.RawMessage(`{"input":42,"output":58,"total":100}`),
			}},
		},
	}

	_, err := service.ImportBatch(context.Background(), projectID, uuid.New(), batch)
	require.NoError(t, err)

	require.NotNil(t, ingestion.lastGeneration)
	require.NotNil(t, ingestion.lastGeneration.Usage, "outer GenerationInput.Usage must be populated")
	usage := ingestion.lastGeneration.Usage
	require.NotNil(t, usage.InputTokens)
	require.NotNil(t, usage.OutputTokens)
	require.NotNil(t, usage.TotalTokens)
	assert.Equal(t, int64(42), *usage.InputTokens)
	assert.Equal(t, int64(58), *usage.OutputTokens)
	assert.Equal(t, int64(100), *usage.TotalTokens)

	// The normalized usage counters must survive into the persisted shape.
	normalized := usage.Normalize()
	assert.Equal(t, uint64(42), normalized.InputTokens)
	assert.Equal(t, uint64(58), normalized.OutputTokens)
	assert.Equal(t, uint64(100), normalized.TotalTokens)
}
