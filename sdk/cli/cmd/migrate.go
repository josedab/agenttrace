package cmd

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

var (
	migrateSource      string
	migrateSourceDSN   string
	migrateSourceFile  string
	migrateHost        string
	migrateDryRun      bool
	migrateIncremental bool
	migrateBatchSize   int
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Migrate data from another observability platform",
	Long: `Migrate traces, prompts, datasets, and scores from another platform to AgentTrace.

Supported sources:
  langfuse    - Migrate from a Langfuse instance

Examples:
  # Validate connection to Langfuse
  agenttrace migrate --source langfuse --validate --source-dsn "postgres://user:pass@host:5432/langfuse"

  # Dry run (preview what will be migrated)
  agenttrace migrate --source langfuse --source-dsn "postgres://..." --dry-run

  # Full migration
  agenttrace migrate --source langfuse --source-dsn "postgres://..."

  # Incremental migration (only new data since last migration)
  agenttrace migrate --source langfuse --source-dsn "postgres://..." --incremental`,
	RunE: runMigrate,
}

var migrateValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate connection to the migration source",
	RunE:  runMigrateValidate,
}

func init() {
	migrateCmd.PersistentFlags().StringVar(&migrateSource, "source", "langfuse", "Source platform (langfuse)")
	migrateCmd.PersistentFlags().StringVar(&migrateSourceDSN, "source-dsn", "", "Source database connection string")
	migrateCmd.PersistentFlags().StringVar(&migrateSourceFile, "source-file", "", "Path to a Langfuse JSON export")
	migrateCmd.PersistentFlags().StringVar(&migrateHost, "target-host", "", "AgentTrace API host (or set AGENTTRACE_API_URL)")
	migrateCmd.PersistentFlags().BoolVar(&migrateDryRun, "dry-run", false, "Preview migration without writing data")
	migrateCmd.PersistentFlags().BoolVar(&migrateIncremental, "incremental", false, "Only migrate data added since last migration")
	migrateCmd.PersistentFlags().IntVar(&migrateBatchSize, "batch-size", 100, "JSON import batch size (1-500)")

	migrateCmd.AddCommand(migrateValidateCmd)
	rootCmd.AddCommand(migrateCmd)
}

func runMigrate(cmd *cobra.Command, args []string) error {
	key := getAPIKey()
	if key == "" {
		return fmt.Errorf("API key required. Set --api-key or AGENTTRACE_API_KEY")
	}

	if migrateSource == "langfuse" {
		sourceFile := resolvedLangfuseSourceFile()
		if sourceFile == "" {
			return fmt.Errorf("--source-file is required for Langfuse JSON migration")
		}
		return runLangfuseJSONMigration(sourceFile, getMigrateHost(), key)
	}

	if migrateSourceDSN == "" {
		return fmt.Errorf("--source-dsn is required for source %s", migrateSource)
	}

	targetHost := getMigrateHost()

	fmt.Printf("🔄 Starting migration from %s...\n", migrateSource)
	fmt.Printf("   Source:      %s\n", migrateSource)
	fmt.Printf("   Source DSN:  %s\n", maskDSN(migrateSourceDSN))
	fmt.Printf("   Target:      %s\n", targetHost)
	fmt.Printf("   Dry Run:     %v\n", migrateDryRun)
	fmt.Printf("   Incremental: %v\n", migrateIncremental)
	fmt.Println()

	// Start migration via API
	body := map[string]any{
		"source": migrateSource,
		"config": map[string]any{
			"sourceDSN":       migrateSourceDSN,
			"dryRun":          migrateDryRun,
			"incrementalMode": migrateIncremental,
			"includeTraces":   true,
			"includePrompts":  true,
			"includeDatasets": true,
			"includeScores":   true,
		},
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", targetHost+"/api/public/migrations", strings.NewReader(string(jsonBody)))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+key)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to start migration: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("migration API returned status %d", resp.StatusCode)
	}

	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	jobID, _ := result["id"].(string)
	fmt.Printf("✅ Migration started (job: %s)\n", jobID)

	if migrateDryRun {
		fmt.Println("   This is a dry run — no data was written.")
		return nil
	}

	// Poll for progress
	fmt.Println("📊 Monitoring progress...")
	return pollMigrationProgress(client, targetHost, key, jobID)
}

func runMigrateValidate(cmd *cobra.Command, args []string) error {
	if migrateSource == "langfuse" {
		sourceFile := resolvedLangfuseSourceFile()
		if sourceFile == "" {
			return fmt.Errorf("--source-file is required for Langfuse JSON validation")
		}
		export, _, err := readLangfuseExport(sourceFile)
		if err != nil {
			return err
		}
		fmt.Printf(
			"✅ Langfuse JSON export is valid (%d traces, %d observations, %d scores, %d prompts)\n",
			len(export.Traces),
			len(export.Observations),
			len(export.Scores),
			len(export.Prompts),
		)
		return nil
	}

	key := getAPIKey()
	if key == "" {
		return fmt.Errorf("API key required. Set --api-key or AGENTTRACE_API_KEY")
	}

	if migrateSourceDSN == "" {
		return fmt.Errorf("--source-dsn is required")
	}

	targetHost := getMigrateHost()
	fmt.Printf("🔍 Validating %s connection...\n", migrateSource)

	body := map[string]any{
		"source": migrateSource,
		"dsn":    migrateSourceDSN,
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", targetHost+"/api/public/migrations/validate", strings.NewReader(string(jsonBody)))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+key)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}
	defer resp.Body.Close()

	var result map[string]any
	json.NewDecoder(resp.Body).Decode(&result)

	if valid, _ := result["valid"].(bool); valid {
		fmt.Println("✅ Connection valid! Ready to migrate.")
		if msg, ok := result["message"].(string); ok {
			fmt.Printf("   %s\n", msg)
		}
	} else {
		msg, _ := result["message"].(string)
		fmt.Printf("❌ Validation failed: %s\n", msg)
		return fmt.Errorf("validation failed")
	}

	return nil
}

type langfuseExportFile struct {
	Traces       []json.RawMessage `json:"traces"`
	Observations []json.RawMessage `json:"observations"`
	Scores       []json.RawMessage `json:"scores"`
	Prompts      []json.RawMessage `json:"prompts"`
}

type langfuseBatchItem struct {
	kind string
	data json.RawMessage
}

func runLangfuseJSONMigration(sourceFile, targetHost, key string) error {
	export, fingerprint, err := readLangfuseExport(sourceFile)
	if err != nil {
		return err
	}
	if migrateBatchSize < 1 || migrateBatchSize > 500 {
		return fmt.Errorf("--batch-size must be between 1 and 500")
	}

	items := flattenLangfuseExport(export)
	jobID := uuid.NewSHA1(
		uuid.NameSpaceURL,
		[]byte(fingerprint+"|"+strings.TrimRight(targetHost, "/")+"|"+strconv.FormatBool(migrateDryRun)),
	)
	fmt.Printf("🔄 Importing Langfuse JSON export...\n")
	fmt.Printf("   Source:      %s\n", filepath.Base(sourceFile))
	fmt.Printf("   Fingerprint: %s\n", fingerprint[:12])
	fmt.Printf("   Job:         %s\n", jobID)
	fmt.Printf("   Records:     %d\n", len(items))
	fmt.Printf("   Dry Run:     %v\n\n", migrateDryRun)

	client := &http.Client{Timeout: 60 * time.Second}
	var lastJob struct {
		Status   string `json:"status"`
		Progress struct {
			TotalItems      int64 `json:"totalItems"`
			ProcessedItems  int64 `json:"processedItems"`
			TracesMigrated  int64 `json:"tracesMigrated"`
			PromptsMigrated int64 `json:"promptsMigrated"`
			ScoresMigrated  int64 `json:"scoresMigrated"`
		} `json:"progress"`
		Errors []string `json:"errors"`
	}

	for start := 0; start < len(items); start += migrateBatchSize {
		end := start + migrateBatchSize
		if end > len(items) {
			end = len(items)
		}
		records := batchLangfuseItems(items[start:end])
		payload := map[string]any{
			"jobId":       jobID,
			"fingerprint": fingerprint,
			"dryRun":      migrateDryRun,
			"totalItems":  len(items),
			"finalBatch":  end == len(items),
			"records":     records,
		}
		body, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("encode Langfuse import batch: %w", err)
		}
		request, err := http.NewRequest(
			http.MethodPost,
			strings.TrimRight(targetHost, "/")+"/api/public/migrations/langfuse/import",
			strings.NewReader(string(body)),
		)
		if err != nil {
			return fmt.Errorf("create Langfuse import request: %w", err)
		}
		request.Header.Set("Authorization", "Bearer "+key)
		request.Header.Set("Content-Type", "application/json")

		response, err := client.Do(request)
		if err != nil {
			return fmt.Errorf(
				"send Langfuse import batch %d-%d (safe to rerun): %w",
				start,
				end,
				err,
			)
		}
		if response.StatusCode < 200 || response.StatusCode >= 300 {
			message, _ := io.ReadAll(io.LimitReader(response.Body, 4096))
			response.Body.Close()
			return fmt.Errorf(
				"Langfuse import API returned %d: %s",
				response.StatusCode,
				strings.TrimSpace(string(message)),
			)
		}
		if err := json.NewDecoder(response.Body).Decode(&lastJob); err != nil {
			response.Body.Close()
			return fmt.Errorf("decode Langfuse import response: %w", err)
		}
		response.Body.Close()
		fmt.Printf("\r   Progress: %d/%d", lastJob.Progress.ProcessedItems, len(items))
	}

	fmt.Println()
	if lastJob.Status == "FAILED" {
		fmt.Println("❌ Import completed with errors:")
		for _, item := range lastJob.Errors {
			fmt.Printf("   - %s\n", item)
		}
		return fmt.Errorf("Langfuse import completed with errors")
	}
	if migrateDryRun {
		fmt.Println("✅ Dry run completed; no records were written.")
	} else {
		fmt.Println("✅ Langfuse import completed.")
	}
	fmt.Printf("   Traces: %d · Prompts: %d · Scores: %d\n",
		lastJob.Progress.TracesMigrated,
		lastJob.Progress.PromptsMigrated,
		lastJob.Progress.ScoresMigrated,
	)
	return nil
}

func readLangfuseExport(path string) (langfuseExportFile, string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return langfuseExportFile{}, "", fmt.Errorf("inspect Langfuse export: %w", err)
	}
	if !info.Mode().IsRegular() {
		return langfuseExportFile{}, "", fmt.Errorf("Langfuse export must be a regular file")
	}
	if info.Size() > 100*1024*1024 {
		return langfuseExportFile{}, "", fmt.Errorf("Langfuse export exceeds the 100 MiB limit")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return langfuseExportFile{}, "", fmt.Errorf("read Langfuse export: %w", err)
	}
	var export langfuseExportFile
	if err := json.Unmarshal(data, &export); err != nil {
		return langfuseExportFile{}, "", fmt.Errorf("parse Langfuse export JSON: %w", err)
	}
	if len(flattenLangfuseExport(export)) == 0 {
		return langfuseExportFile{}, "", fmt.Errorf("Langfuse export contains no supported records")
	}
	checksum := sha256.Sum256(data)
	return export, fmt.Sprintf("%x", checksum[:]), nil
}

func flattenLangfuseExport(export langfuseExportFile) []langfuseBatchItem {
	items := make([]langfuseBatchItem, 0,
		len(export.Traces)+len(export.Observations)+len(export.Scores)+len(export.Prompts),
	)
	appendItems := func(kind string, values []json.RawMessage) {
		for _, value := range values {
			items = append(items, langfuseBatchItem{kind: kind, data: value})
		}
	}
	appendItems("traces", export.Traces)
	appendItems("observations", export.Observations)
	appendItems("scores", export.Scores)
	appendItems("prompts", export.Prompts)
	return items
}

func batchLangfuseItems(items []langfuseBatchItem) map[string][]json.RawMessage {
	result := map[string][]json.RawMessage{
		"traces":       {},
		"observations": {},
		"scores":       {},
		"prompts":      {},
	}
	for _, item := range items {
		result[item.kind] = append(result[item.kind], item.data)
	}
	return result
}

func resolvedLangfuseSourceFile() string {
	if migrateSourceFile != "" {
		return migrateSourceFile
	}
	if migrateSourceDSN != "" {
		if info, err := os.Stat(migrateSourceDSN); err == nil && info.Mode().IsRegular() {
			return migrateSourceDSN
		}
	}
	return ""
}

func pollMigrationProgress(client *http.Client, targetHost, key, jobID string) error {
	for {
		time.Sleep(2 * time.Second)

		req, _ := http.NewRequest("GET", targetHost+"/api/public/migrations/"+jobID, nil)
		req.Header.Set("Authorization", "Bearer "+key)

		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("   ⚠ Failed to check progress: %v\n", err)
			continue
		}

		var job map[string]any
		json.NewDecoder(resp.Body).Decode(&job)
		resp.Body.Close()

		status, _ := job["status"].(string)
		progress, _ := job["progress"].(map[string]any)

		if progress != nil {
			processed, _ := progress["processedItems"].(float64)
			total, _ := progress["totalItems"].(float64)
			if total > 0 {
				pct := (processed / total) * 100
				fmt.Printf("\r   Progress: %.0f/%.0f (%.1f%%)", processed, total, pct)
			}
		}

		switch status {
		case "COMPLETED":
			fmt.Println()
			fmt.Println("✅ Migration completed successfully!")
			printMigrationSummary(progress)
			return nil
		case "FAILED":
			fmt.Println()
			errs, _ := job["errors"].([]any)
			fmt.Println("❌ Migration failed!")
			for _, e := range errs {
				fmt.Printf("   - %v\n", e)
			}
			return fmt.Errorf("migration failed")
		}
	}
}

func printMigrationSummary(progress map[string]any) {
	if progress == nil {
		return
	}
	fmt.Println("   Summary:")
	if v, ok := progress["tracesMigrated"].(float64); ok && v > 0 {
		fmt.Printf("   - Traces:   %.0f\n", v)
	}
	if v, ok := progress["promptsMigrated"].(float64); ok && v > 0 {
		fmt.Printf("   - Prompts:  %.0f\n", v)
	}
	if v, ok := progress["datasetsMigrated"].(float64); ok && v > 0 {
		fmt.Printf("   - Datasets: %.0f\n", v)
	}
	if v, ok := progress["scoresMigrated"].(float64); ok && v > 0 {
		fmt.Printf("   - Scores:   %.0f\n", v)
	}
}

func getMigrateHost() string {
	if migrateHost != "" {
		return migrateHost
	}
	if h := os.Getenv("AGENTTRACE_API_URL"); h != "" {
		return h
	}
	return host
}

func maskDSN(dsn string) string {
	// Mask password in DSN for display
	if idx := strings.Index(dsn, "://"); idx >= 0 {
		rest := dsn[idx+3:]
		if atIdx := strings.Index(rest, "@"); atIdx >= 0 {
			if colonIdx := strings.Index(rest[:atIdx], ":"); colonIdx >= 0 {
				return dsn[:idx+3] + rest[:colonIdx+1] + "****@" + rest[atIdx+1:]
			}
		}
	}
	return dsn
}
