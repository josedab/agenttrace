package domain

import (
	"encoding/json"

	"github.com/google/uuid"
)

// LangfuseExport is the documented JSON export subset accepted by AgentTrace.
type LangfuseExport struct {
	Traces       []LangfuseTrace       `json:"traces,omitempty"`
	Observations []LangfuseObservation `json:"observations,omitempty"`
	Scores       []LangfuseScore       `json:"scores,omitempty"`
	Prompts      []LangfusePrompt      `json:"prompts,omitempty"`
}

// LangfuseTrace maps the supported trace export fields.
type LangfuseTrace struct {
	ID        string          `json:"id"`
	Name      string          `json:"name,omitempty"`
	Timestamp string          `json:"timestamp,omitempty"`
	StartTime string          `json:"startTime,omitempty"`
	EndTime   string          `json:"endTime,omitempty"`
	Input     json.RawMessage `json:"input,omitempty"`
	Output    json.RawMessage `json:"output,omitempty"`
	Metadata  json.RawMessage `json:"metadata,omitempty"`
	UserID    string          `json:"userId,omitempty"`
	SessionID string          `json:"sessionId,omitempty"`
	Tags      []string        `json:"tags,omitempty"`
}

// LangfuseObservation maps spans, generations, and events.
type LangfuseObservation struct {
	ID                  string          `json:"id"`
	TraceID             string          `json:"traceId"`
	ParentObservationID string          `json:"parentObservationId,omitempty"`
	Type                string          `json:"type"`
	Name                string          `json:"name,omitempty"`
	StartTime           string          `json:"startTime,omitempty"`
	EndTime             string          `json:"endTime,omitempty"`
	Input               json.RawMessage `json:"input,omitempty"`
	Output              json.RawMessage `json:"output,omitempty"`
	Metadata            json.RawMessage `json:"metadata,omitempty"`
	Model               string          `json:"model,omitempty"`
	ModelParameters     json.RawMessage `json:"modelParameters,omitempty"`
	Usage               json.RawMessage `json:"usage,omitempty"`
	Level               string          `json:"level,omitempty"`
	StatusMessage       string          `json:"statusMessage,omitempty"`
}

// LangfuseScore maps numeric and categorical scores.
type LangfuseScore struct {
	ID            string   `json:"id"`
	TraceID       string   `json:"traceId"`
	ObservationID string   `json:"observationId,omitempty"`
	Name          string   `json:"name"`
	Value         *float64 `json:"value,omitempty"`
	StringValue   string   `json:"stringValue,omitempty"`
	Comment       string   `json:"comment,omitempty"`
	DataType      string   `json:"dataType,omitempty"`
}

// LangfusePrompt maps one immutable prompt version.
type LangfusePrompt struct {
	ID            string          `json:"id"`
	Name          string          `json:"name"`
	Type          string          `json:"type,omitempty"`
	Version       int             `json:"version,omitempty"`
	Prompt        json.RawMessage `json:"prompt,omitempty"`
	Content       string          `json:"content,omitempty"`
	Config        json.RawMessage `json:"config,omitempty"`
	Labels        []string        `json:"labels,omitempty"`
	CommitMessage string          `json:"commitMessage,omitempty"`
}

// LangfuseImportBatch imports a bounded resumable batch.
type LangfuseImportBatch struct {
	JobID       uuid.UUID      `json:"jobId"`
	Fingerprint string         `json:"fingerprint"`
	DryRun      bool           `json:"dryRun"`
	TotalItems  int64          `json:"totalItems"`
	FinalBatch  bool           `json:"finalBatch"`
	Records     LangfuseExport `json:"records"`
}
