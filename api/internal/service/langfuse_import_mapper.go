package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	apperrors "github.com/agenttrace/agenttrace/api/internal/pkg/errors"
)

func (s *LangfuseImportService) importTrace(
	ctx context.Context,
	projectID uuid.UUID,
	source domain.LangfuseTrace,
	dryRun bool,
) (string, error) {
	if source.ID == "" {
		return "", apperrors.Validation("trace id is required")
	}
	traceID := normalizeExternalID(source.ID, 32)
	startTime, err := parseLangfuseTime(firstNonEmpty(source.StartTime, source.Timestamp))
	if err != nil {
		return "", fmt.Errorf("invalid trace start time: %w", err)
	}
	var endTime *time.Time
	if source.EndTime != "" {
		parsed, err := parseLangfuseTime(source.EndTime)
		if err != nil {
			return "", fmt.Errorf("invalid trace end time: %w", err)
		}
		endTime = &parsed
	}
	input := &domain.TraceInput{
		ID:        traceID,
		Name:      source.Name,
		UserID:    source.UserID,
		SessionID: source.SessionID,
		Tags:      source.Tags,
		Metadata:  rawJSONValue(source.Metadata),
		Input:     rawJSONValue(source.Input),
		Output:    rawJSONValue(source.Output),
		StartTime: &startTime,
		EndTime:   endTime,
	}
	if dryRun {
		return traceID, nil
	}
	trace, err := s.ingestion.IngestTrace(ctx, projectID, input)
	if err != nil {
		return "", err
	}
	return trace.ID, nil
}

func (s *LangfuseImportService) importObservation(
	ctx context.Context,
	projectID uuid.UUID,
	source domain.LangfuseObservation,
	dryRun bool,
) (string, error) {
	if source.ID == "" || source.TraceID == "" {
		return "", apperrors.Validation("observation id and traceId are required")
	}
	observationID := normalizeExternalID(source.ID, 16)
	traceID := normalizeExternalID(source.TraceID, 32)
	startTime, err := parseLangfuseTime(source.StartTime)
	if err != nil {
		return "", fmt.Errorf("invalid observation start time: %w", err)
	}
	var endTime *time.Time
	if source.EndTime != "" {
		parsed, err := parseLangfuseTime(source.EndTime)
		if err != nil {
			return "", fmt.Errorf("invalid observation end time: %w", err)
		}
		endTime = &parsed
	}
	observationType := domain.ObservationType(strings.ToUpper(source.Type))
	if !observationType.IsValid() {
		return "", apperrors.Validation("observation type must be SPAN, GENERATION, or EVENT")
	}
	level := domain.Level(strings.ToUpper(source.Level))
	if source.Level == "" || !level.IsValid() {
		level = domain.LevelDefault
	}
	var parentID *string
	if source.ParentObservationID != "" {
		normalized := normalizeExternalID(source.ParentObservationID, 16)
		parentID = &normalized
	}
	input := &domain.ObservationInput{
		ID:                  &observationID,
		TraceID:             &traceID,
		ParentObservationID: parentID,
		Type:                &observationType,
		Name:                optionalString(source.Name),
		Level:               &level,
		StatusMessage:       optionalString(source.StatusMessage),
		Metadata:            rawJSONValue(source.Metadata),
		StartTime:           &startTime,
		EndTime:             endTime,
		Input:               rawJSONValue(source.Input),
		Output:              rawJSONValue(source.Output),
		Model:               optionalString(source.Model),
		ModelParameters:     rawJSONValue(source.ModelParameters),
		Usage:               rawJSONValue(source.Usage),
	}
	// Parse token usage up front so invalid counts are rejected regardless of
	// dry-run mode. GenerationInput.Usage (an outer *UsageDetailsInput) shadows
	// the embedded ObservationInput.Usage that IngestGeneration ignores, so the
	// parsed value must be assigned to the outer field explicitly.
	usage, err := mapLangfuseUsage(source.Usage)
	if err != nil {
		return "", err
	}
	if dryRun {
		return observationID, nil
	}
	var observation *domain.Observation
	if observationType == domain.ObservationTypeGeneration {
		observation, err = s.ingestion.IngestGeneration(ctx, projectID, &domain.GenerationInput{
			ObservationInput: *input,
			Model:            source.Model,
			ModelParameters:  input.ModelParameters,
			Usage:            usage,
		})
	} else {
		observation, err = s.ingestion.IngestObservation(ctx, projectID, input)
	}
	if err != nil {
		return "", err
	}
	return observation.ID, nil
}

// langfuseImportNamespace anchors deterministic identifiers derived from
// Langfuse source identifiers.
var langfuseImportNamespace = uuid.NewSHA1(
	uuid.NameSpaceURL,
	[]byte("https://agenttrace.dev/langfuse-import"),
)

// deterministicImportID derives a stable UUID for an imported record so a retry
// after a partial failure replaces the same row instead of duplicating it.
func deterministicImportID(projectID uuid.UUID, sourceType, sourceID string) uuid.UUID {
	return uuid.NewSHA1(
		langfuseImportNamespace,
		[]byte(projectID.String()+"/"+sourceType+"/"+sourceID),
	)
}

func (s *LangfuseImportService) importScore(
	ctx context.Context,
	projectID uuid.UUID,
	source domain.LangfuseScore,
	dryRun bool,
) (string, error) {
	if source.ID == "" || source.TraceID == "" || source.Name == "" {
		return "", apperrors.Validation("score id, traceId, and name are required")
	}
	traceID := normalizeExternalID(source.TraceID, 32)
	var observationID *string
	if source.ObservationID != "" {
		normalized := normalizeExternalID(source.ObservationID, 16)
		observationID = &normalized
	}
	var dataType domain.ScoreDataType
	if source.Value != nil {
		dataType = domain.ScoreDataTypeNumeric
	}
	if strings.EqualFold(source.DataType, "BOOLEAN") && source.Value != nil {
		dataType = domain.ScoreDataTypeBoolean
	}
	var stringValue *string
	if source.StringValue != "" {
		stringValue = &source.StringValue
	}
	scoreID := deterministicImportID(projectID, "score", source.ID)
	input := &domain.ScoreInput{
		ID:            &scoreID,
		TraceID:       traceID,
		ObservationID: observationID,
		Name:          source.Name,
		Source:        domain.ScoreSourceAPI,
		DataType:      dataType,
		Value:         source.Value,
		StringValue:   stringValue,
		Comment:       optionalString(source.Comment),
	}
	if dryRun {
		return scoreID.String(), nil
	}
	score, err := s.scores.Create(ctx, projectID, input)
	if err != nil {
		return "", err
	}
	return score.ID.String(), nil
}

func (s *LangfuseImportService) importPrompt(
	ctx context.Context,
	projectID, actorID uuid.UUID,
	source domain.LangfusePrompt,
	dryRun bool,
) (string, error) {
	if source.ID == "" || source.Name == "" {
		return "", apperrors.Validation("prompt id and name are required")
	}
	content := source.Content
	if content == "" {
		value := rawJSONValue(source.Prompt)
		switch typed := value.(type) {
		case string:
			content = typed
		case nil:
		default:
			encoded, marshalErr := json.Marshal(typed)
			if marshalErr != nil {
				return "", fmt.Errorf("encode prompt content: %w", marshalErr)
			}
			content = string(encoded)
		}
	}
	if content == "" {
		return "", apperrors.Validation("prompt content is required")
	}
	promptType := domain.PromptTypeText
	if strings.EqualFold(source.Type, string(domain.PromptTypeChat)) {
		promptType = domain.PromptTypeChat
	}
	if dryRun {
		return normalizeExternalID(source.ID, 32), nil
	}

	existing, err := s.prompts.GetByName(ctx, projectID, source.Name)
	if err == nil {
		// A retried import must not append a second identical version.
		if existing.LatestVersion != nil && existing.LatestVersion.Content == content {
			return existing.LatestVersion.ID.String(), nil
		}
		version, versionErr := s.prompts.CreateVersion(
			ctx,
			existing.ID,
			&domain.PromptVersionInput{
				Content:       content,
				Config:        rawJSONValue(source.Config),
				Labels:        source.Labels,
				CommitMessage: optionalString(source.CommitMessage),
			},
			actorID,
		)
		if versionErr != nil {
			return "", versionErr
		}
		return version.ID.String(), nil
	}
	if !apperrors.IsNotFound(err) {
		return "", err
	}
	prompt, err := s.prompts.Create(ctx, projectID, &domain.PromptInput{
		Name:    source.Name,
		Type:    promptType,
		Content: content,
		Config:  rawJSONValue(source.Config),
		Labels:  source.Labels,
	}, actorID)
	if err != nil {
		return "", err
	}
	return prompt.ID.String(), nil
}

func normalizeExternalID(value string, length int) string {
	normalized := strings.ToLower(strings.ReplaceAll(strings.TrimSpace(value), "-", ""))
	if len(normalized) == length {
		if _, err := hex.DecodeString(normalized); err == nil {
			return normalized
		}
	}
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])[:length]
}

func parseLangfuseTime(value string) (time.Time, error) {
	if value == "" {
		return time.Time{}, apperrors.Validation("timestamp is required")
	}
	for _, layout := range []string{time.RFC3339Nano, time.RFC3339} {
		if parsed, err := time.Parse(layout, value); err == nil {
			return parsed.UTC(), nil
		}
	}
	if milliseconds, err := strconv.ParseInt(value, 10, 64); err == nil {
		return time.UnixMilli(milliseconds).UTC(), nil
	}
	return time.Time{}, apperrors.Validation("unsupported timestamp format")
}

func rawJSONValue(value json.RawMessage) interface{} {
	if len(value) == 0 || string(value) == "null" {
		return nil
	}
	var result interface{}
	if err := json.Unmarshal(value, &result); err != nil {
		return string(value)
	}
	return result
}

func optionalString(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

func checksumValue(value interface{}) string {
	encoded, err := json.Marshal(value)
	if err != nil {
		encoded = []byte(fmt.Sprintf("%#v", value))
	}
	checksum := sha256.Sum256(encoded)
	return hex.EncodeToString(checksum[:])
}

// mapLangfuseUsage converts a Langfuse usage payload into the outer
// GenerationInput.Usage shape. Langfuse has shipped several token field naming
// conventions over time, so all of the following are accepted, in order of
// precedence per dimension:
//
//	input tokens : input, inputTokens, promptTokens
//	output tokens: output, outputTokens, completionTokens
//	total tokens : total, totalTokens
//
// A nil or empty payload yields (nil, nil). Any negative count is rejected so a
// malformed export surfaces as a validation error rather than silently
// underflowing the unsigned usage counters downstream.
func mapLangfuseUsage(raw json.RawMessage) (*domain.UsageDetailsInput, error) {
	trimmed := strings.TrimSpace(string(raw))
	if trimmed == "" || trimmed == "null" {
		return nil, nil
	}

	var payload struct {
		Input  *int64 `json:"input"`
		Output *int64 `json:"output"`
		Total  *int64 `json:"total"`

		InputTokens  *int64 `json:"inputTokens"`
		OutputTokens *int64 `json:"outputTokens"`
		TotalTokens  *int64 `json:"totalTokens"`

		PromptTokens     *int64 `json:"promptTokens"`
		CompletionTokens *int64 `json:"completionTokens"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, apperrors.Validation("invalid observation usage: " + err.Error())
	}

	for _, v := range []*int64{
		payload.Input, payload.Output, payload.Total,
		payload.InputTokens, payload.OutputTokens, payload.TotalTokens,
		payload.PromptTokens, payload.CompletionTokens,
	} {
		if v != nil && *v < 0 {
			return nil, apperrors.Validation("observation usage token counts must not be negative")
		}
	}

	inputTokens := firstNonNilInt64(payload.Input, payload.InputTokens, payload.PromptTokens)
	outputTokens := firstNonNilInt64(payload.Output, payload.OutputTokens, payload.CompletionTokens)
	totalTokens := firstNonNilInt64(payload.Total, payload.TotalTokens)

	if inputTokens == nil && outputTokens == nil && totalTokens == nil {
		return nil, nil
	}

	return &domain.UsageDetailsInput{
		InputTokens:  inputTokens,
		OutputTokens: outputTokens,
		TotalTokens:  totalTokens,
	}, nil
}

// firstNonNilInt64 returns the first non-nil pointer from the provided
// candidates, or nil when they are all absent.
func firstNonNilInt64(candidates ...*int64) *int64 {
	for _, c := range candidates {
		if c != nil {
			return c
		}
	}
	return nil
}
