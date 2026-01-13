package worker

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/minio/minio-go/v7"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

const (
	// TypeDataExport is the task type for data export
	TypeDataExport = "export:data"
	// TypeDatasetExport is the task type for dataset export
	TypeDatasetExport = "export:dataset"
)

// DataExportPayload is the payload for data export tasks
type DataExportPayload struct {
	JobID       uuid.UUID              `json:"job_id"`
	ProjectID   uuid.UUID              `json:"project_id"`
	UserID      uuid.UUID              `json:"user_id"`
	Type        string                 `json:"type"` // traces, observations, scores
	Format      domain.ExportFormat    `json:"format"`
	Filters     map[string]interface{} `json:"filters"`
	Destination *ExportDestination     `json:"destination,omitempty"`
}

// ExportDestination defines where to send the export
type ExportDestination struct {
	Type   domain.DestinationType `json:"type"`
	Bucket string                 `json:"bucket"`
	Path   string                 `json:"path"`
	Config map[string]string      `json:"config"`
}

// NewDataExportTask creates a data export task
func NewDataExportTask(payload *DataExportPayload) (*asynq.Task, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data export payload: %w", err)
	}
	return asynq.NewTask(TypeDataExport, data, asynq.MaxRetry(3), asynq.Timeout(30*time.Minute)), nil
}

// DatasetExportPayload is the payload for dataset export tasks
type DatasetExportPayload struct {
	JobID       uuid.UUID           `json:"job_id"`
	ProjectID   uuid.UUID           `json:"project_id"`
	DatasetID   uuid.UUID           `json:"dataset_id"`
	UserID      uuid.UUID           `json:"user_id"`
	Format      domain.ExportFormat `json:"format"`
	IncludeRuns bool                `json:"include_runs"`
	Destination *ExportDestination  `json:"destination,omitempty"`
}

// NewDatasetExportTask creates a dataset export task
func NewDatasetExportTask(payload *DatasetExportPayload) (*asynq.Task, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal dataset export payload: %w", err)
	}
	return asynq.NewTask(TypeDatasetExport, data, asynq.MaxRetry(3), asynq.Timeout(30*time.Minute)), nil
}

// ExportWorker handles export tasks
type ExportWorker struct {
	logger         *zap.Logger
	queryService   *service.QueryService
	scoreService   *service.ScoreService
	datasetService *service.DatasetService
	minioClient    *minio.Client
	bucket         string
}

// NewExportWorker creates a new export worker
func NewExportWorker(
	logger *zap.Logger,
	queryService *service.QueryService,
	scoreService *service.ScoreService,
	datasetService *service.DatasetService,
	minioClient *minio.Client,
	bucket string,
) *ExportWorker {
	return &ExportWorker{
		logger:         logger,
		queryService:   queryService,
		scoreService:   scoreService,
		datasetService: datasetService,
		minioClient:    minioClient,
		bucket:         bucket,
	}
}

// ProcessTask processes a data export task
func (w *ExportWorker) ProcessTask(ctx context.Context, t *asynq.Task) error {
	var payload DataExportPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal data export payload: %w", err)
	}

	w.logger.Info("processing data export",
		zap.String("job_id", payload.JobID.String()),
		zap.String("type", payload.Type),
		zap.String("format", string(payload.Format)),
	)

	// Export data based on type
	var data []byte
	var err error
	var filename string

	switch payload.Type {
	case "traces":
		data, err = w.exportTraces(ctx, payload.ProjectID, payload.Filters, payload.Format)
		filename = fmt.Sprintf("traces_%s.%s", time.Now().Format("20060102_150405"), payload.Format)
	case "observations":
		data, err = w.exportObservations(ctx, payload.ProjectID, payload.Filters, payload.Format)
		filename = fmt.Sprintf("observations_%s.%s", time.Now().Format("20060102_150405"), payload.Format)
	case "scores":
		data, err = w.exportScores(ctx, payload.ProjectID, payload.Filters, payload.Format)
		filename = fmt.Sprintf("scores_%s.%s", time.Now().Format("20060102_150405"), payload.Format)
	default:
		return fmt.Errorf("unsupported export type: %s", payload.Type)
	}

	if err != nil {
		return fmt.Errorf("failed to export data: %w", err)
	}

	// Upload to storage
	path := fmt.Sprintf("exports/%s/%s", payload.ProjectID.String(), filename)
	if err := w.uploadToStorage(ctx, path, data, payload.Destination); err != nil {
		return fmt.Errorf("failed to upload export: %w", err)
	}

	w.logger.Info("data export completed",
		zap.String("job_id", payload.JobID.String()),
		zap.String("path", path),
		zap.Int("size", len(data)),
	)

	return nil
}

// ProcessDatasetTask processes a dataset export task
func (w *ExportWorker) ProcessDatasetTask(ctx context.Context, t *asynq.Task) error {
	var payload DatasetExportPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal dataset export payload: %w", err)
	}

	w.logger.Info("processing dataset export",
		zap.String("job_id", payload.JobID.String()),
		zap.String("dataset_id", payload.DatasetID.String()),
		zap.String("format", string(payload.Format)),
	)

	// Get dataset
	dataset, err := w.datasetService.Get(ctx, payload.DatasetID)
	if err != nil {
		return fmt.Errorf("failed to get dataset: %w", err)
	}

	// Get items
	itemFilter := &domain.DatasetItemFilter{DatasetID: payload.DatasetID}
	itemsList, _, err := w.datasetService.ListItems(ctx, itemFilter, 10000, 0)
	if err != nil {
		return fmt.Errorf("failed to get dataset items: %w", err)
	}

	// Convert to pointer slice for export functions
	items := make([]*domain.DatasetItem, len(itemsList))
	for i := range itemsList {
		items[i] = &itemsList[i]
	}

	// Export based on format
	var data []byte
	var filename string

	switch payload.Format {
	case domain.ExportFormatJSON:
		data, err = w.exportDatasetJSON(dataset, items, payload.IncludeRuns)
		filename = fmt.Sprintf("%s_%s.json", dataset.Name, time.Now().Format("20060102_150405"))
	case domain.ExportFormatCSV:
		data, err = w.exportDatasetCSV(items)
		filename = fmt.Sprintf("%s_%s.csv", dataset.Name, time.Now().Format("20060102_150405"))
	case domain.ExportFormatOpenAIFinetune:
		data, err = w.exportDatasetOpenAIFinetune(items)
		filename = fmt.Sprintf("%s_%s.jsonl", dataset.Name, time.Now().Format("20060102_150405"))
	default:
		return fmt.Errorf("unsupported export format: %s", payload.Format)
	}

	if err != nil {
		return fmt.Errorf("failed to export dataset: %w", err)
	}

	// Upload to storage
	path := fmt.Sprintf("exports/%s/datasets/%s", payload.ProjectID.String(), filename)
	if err := w.uploadToStorage(ctx, path, data, payload.Destination); err != nil {
		return fmt.Errorf("failed to upload export: %w", err)
	}

	w.logger.Info("dataset export completed",
		zap.String("job_id", payload.JobID.String()),
		zap.String("path", path),
		zap.Int("items", len(items)),
	)

	return nil
}

// exportTraces exports traces
func (w *ExportWorker) exportTraces(
	ctx context.Context,
	projectID uuid.UUID,
	filters map[string]interface{},
	format domain.ExportFormat,
) ([]byte, error) {
	_ = filters

	// Get traces
	traceList, err := w.queryService.ListTraces(ctx, &domain.TraceFilter{
		ProjectID: projectID,
	}, 10000, 0)
	if err != nil {
		return nil, err
	}

	switch format {
	case domain.ExportFormatJSON:
		return json.Marshal(traceList.Traces)
	case domain.ExportFormatCSV:
		return w.tracesToCSV(traceList.Traces)
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}

// exportObservations exports observations
func (w *ExportWorker) exportObservations(
	ctx context.Context,
	projectID uuid.UUID,
	filters map[string]interface{},
	format domain.ExportFormat,
) ([]byte, error) {
	_ = filters

	// Get observations
	observations, _, err := w.queryService.ListObservations(ctx, &domain.ObservationFilter{
		ProjectID: projectID,
	}, 10000, 0)
	if err != nil {
		return nil, err
	}

	switch format {
	case domain.ExportFormatJSON:
		return json.Marshal(observations)
	case domain.ExportFormatCSV:
		return w.observationsToCSV(observations)
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}

// exportScores exports scores
func (w *ExportWorker) exportScores(
	ctx context.Context,
	projectID uuid.UUID,
	filters map[string]interface{},
	format domain.ExportFormat,
) ([]byte, error) {
	// Build score filter from map
	filter := &domain.ScoreFilter{
		ProjectID: projectID,
	}

	// Apply filters if provided
	if filters != nil {
		if traceID, ok := filters["trace_id"].(string); ok {
			filter.TraceID = &traceID
		}
		if name, ok := filters["name"].(string); ok {
			filter.Name = &name
		}
		if source, ok := filters["source"].(string); ok {
			s := domain.ScoreSource(source)
			filter.Source = &s
		}
	}

	// Set project ID in filter and get scores from service
	filter.ProjectID = projectID
	scoreList, err := w.scoreService.List(ctx, filter, 10000, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to list scores: %w", err)
	}

	switch format {
	case domain.ExportFormatJSON:
		return json.Marshal(scoreList.Scores)
	case domain.ExportFormatCSV:
		return w.scoresToCSV(scoreList.Scores)
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}

// tracesToCSV converts traces to CSV
func (w *ExportWorker) tracesToCSV(traces []domain.Trace) ([]byte, error) {
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// Header
	header := []string{"id", "name", "start_time", "user_id", "session_id", "input", "output", "duration_ms", "total_cost"}
	if err := writer.Write(header); err != nil {
		return nil, err
	}

	// Rows
	for _, t := range traces {
		row := []string{
			t.ID,
			t.Name,
			t.StartTime.Format(time.RFC3339),
			t.UserID,
			t.SessionID,
			t.Input,
			t.Output,
			fmt.Sprintf("%.2f", t.DurationMs),
			fmt.Sprintf("%.6f", t.TotalCost),
		}
		if err := writer.Write(row); err != nil {
			return nil, err
		}
	}

	writer.Flush()
	return buf.Bytes(), writer.Error()
}

// observationsToCSV converts observations to CSV
func (w *ExportWorker) observationsToCSV(observations []domain.Observation) ([]byte, error) {
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// Header
	header := []string{"id", "trace_id", "parent_id", "type", "name", "model", "start_time", "end_time", "input_tokens", "output_tokens", "total_cost"}
	if err := writer.Write(header); err != nil {
		return nil, err
	}

	// Rows
	for _, o := range observations {
		endTime := ""
		if o.EndTime != nil {
			endTime = o.EndTime.Format(time.RFC3339)
		}
		parentID := ""
		if o.ParentObservationID != nil {
			parentID = *o.ParentObservationID
		}
		row := []string{
			o.ID,
			o.TraceID,
			parentID,
			string(o.Type),
			o.Name,
			o.Model,
			o.StartTime.Format(time.RFC3339),
			endTime,
			fmt.Sprintf("%d", o.UsageDetails.InputTokens),
			fmt.Sprintf("%d", o.UsageDetails.OutputTokens),
			fmt.Sprintf("%.6f", o.CostDetails.TotalCost),
		}
		if err := writer.Write(row); err != nil {
			return nil, err
		}
	}

	writer.Flush()
	return buf.Bytes(), writer.Error()
}

// scoresToCSV converts scores to CSV
func (w *ExportWorker) scoresToCSV(scores []domain.Score) ([]byte, error) {
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// Header
	header := []string{"id", "trace_id", "observation_id", "name", "value", "string_value", "data_type", "source", "comment", "created_at"}
	if err := writer.Write(header); err != nil {
		return nil, err
	}

	// Rows
	for _, s := range scores {
		value := ""
		if s.Value != nil {
			value = fmt.Sprintf("%.4f", *s.Value)
		}
		observationID := ""
		if s.ObservationID != nil {
			observationID = *s.ObservationID
		}
		stringValue := ""
		if s.StringValue != nil {
			stringValue = *s.StringValue
		}
		row := []string{
			s.ID.String(),
			s.TraceID,
			observationID,
			s.Name,
			value,
			stringValue,
			string(s.DataType),
			string(s.Source),
			s.Comment,
			s.CreatedAt.Format(time.RFC3339),
		}
		if err := writer.Write(row); err != nil {
			return nil, err
		}
	}

	writer.Flush()
	return buf.Bytes(), writer.Error()
}

// exportDatasetJSON exports dataset as JSON
func (w *ExportWorker) exportDatasetJSON(dataset *domain.Dataset, items []*domain.DatasetItem, includeRuns bool) ([]byte, error) {
	export := map[string]interface{}{
		"dataset": dataset,
		"items":   items,
	}

	if includeRuns {
		runs, _, err := w.datasetService.ListRuns(context.Background(), dataset.ID, 1000, 0)
		if err == nil {
			export["runs"] = runs
		}
	}

	return json.MarshalIndent(export, "", "  ")
}

// exportDatasetCSV exports dataset as CSV
func (w *ExportWorker) exportDatasetCSV(items []*domain.DatasetItem) ([]byte, error) {
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// Header
	header := []string{"id", "input", "expected_output", "metadata", "status", "created_at"}
	if err := writer.Write(header); err != nil {
		return nil, err
	}

	// Rows
	for _, item := range items {
		expectedOutput := ""
		if item.ExpectedOutput != nil {
			expectedOutput = *item.ExpectedOutput
		}
		row := []string{
			item.ID.String(),
			item.Input,
			expectedOutput,
			item.Metadata,
			string(item.Status),
			item.CreatedAt.Format(time.RFC3339),
		}
		if err := writer.Write(row); err != nil {
			return nil, err
		}
	}

	writer.Flush()
	return buf.Bytes(), writer.Error()
}

// exportDatasetOpenAIFinetune exports dataset in OpenAI fine-tune format (JSONL)
func (w *ExportWorker) exportDatasetOpenAIFinetune(items []*domain.DatasetItem) ([]byte, error) {
	var buf bytes.Buffer

	for _, item := range items {
		// Convert to OpenAI messages format
		messages := []map[string]string{}

		// Add system message if present in metadata
		if item.Metadata != "" {
			var metadata map[string]interface{}
			if err := json.Unmarshal([]byte(item.Metadata), &metadata); err == nil {
				if system, ok := metadata["system"].(string); ok {
					messages = append(messages, map[string]string{
						"role":    "system",
						"content": system,
					})
				}
			}
		}

		// Add user message from input
		if item.Input != "" {
			var inputMap map[string]interface{}
			if err := json.Unmarshal([]byte(item.Input), &inputMap); err == nil {
				if content, ok := inputMap["content"].(string); ok {
					messages = append(messages, map[string]string{
						"role":    "user",
						"content": content,
					})
				} else {
					messages = append(messages, map[string]string{
						"role":    "user",
						"content": item.Input,
					})
				}
			} else {
				messages = append(messages, map[string]string{
					"role":    "user",
					"content": item.Input,
				})
			}
		}

		// Add assistant message from expected output
		if item.ExpectedOutput != nil {
			var outputMap map[string]interface{}
			if err := json.Unmarshal([]byte(*item.ExpectedOutput), &outputMap); err == nil {
				if content, ok := outputMap["content"].(string); ok {
					messages = append(messages, map[string]string{
						"role":    "assistant",
						"content": content,
					})
				} else {
					messages = append(messages, map[string]string{
						"role":    "assistant",
						"content": *item.ExpectedOutput,
					})
				}
			} else {
				messages = append(messages, map[string]string{
					"role":    "assistant",
					"content": *item.ExpectedOutput,
				})
			}
		}

		line := map[string]interface{}{
			"messages": messages,
		}

		lineJSON, err := json.Marshal(line)
		if err != nil {
			continue
		}

		buf.Write(lineJSON)
		buf.WriteByte('\n')
	}

	return buf.Bytes(), nil
}

// uploadToStorage uploads data to storage
func (w *ExportWorker) uploadToStorage(ctx context.Context, path string, data []byte, dest *ExportDestination) error {
	bucket := w.bucket
	if dest != nil && dest.Bucket != "" {
		bucket = dest.Bucket
	}
	if dest != nil && dest.Path != "" {
		path = dest.Path
	}

	// Upload to MinIO
	reader := bytes.NewReader(data)
	_, err := w.minioClient.PutObject(ctx, bucket, path, reader, int64(len(data)), minio.PutObjectOptions{
		ContentType: "application/octet-stream",
	})
	if err != nil {
		return fmt.Errorf("failed to upload to MinIO: %w", err)
	}

	return nil
}
