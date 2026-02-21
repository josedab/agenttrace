package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var (
	migrateSource    string
	migrateSourceDSN string
	migrateHost      string
	migrateDryRun    bool
	migrateIncremental bool
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
	migrateCmd.PersistentFlags().StringVar(&migrateHost, "target-host", "", "AgentTrace API host (or set AGENTTRACE_API_URL)")
	migrateCmd.PersistentFlags().BoolVar(&migrateDryRun, "dry-run", false, "Preview migration without writing data")
	migrateCmd.PersistentFlags().BoolVar(&migrateIncremental, "incremental", false, "Only migrate data added since last migration")

	migrateCmd.AddCommand(migrateValidateCmd)
	rootCmd.AddCommand(migrateCmd)
}

func runMigrate(cmd *cobra.Command, args []string) error {
	key := getAPIKey()
	if key == "" {
		return fmt.Errorf("API key required. Set --api-key or AGENTTRACE_API_KEY")
	}

	if migrateSourceDSN == "" {
		return fmt.Errorf("--source-dsn is required")
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
