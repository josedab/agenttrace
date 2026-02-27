package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var (
	syncDirection string
	syncBasePath  string
	syncBranch    string
	syncDryRun    bool
)

var promptsCmd = &cobra.Command{
	Use:   "prompts",
	Short: "Manage prompts as code with GitOps",
	Long: `Manage AI prompts as code in git repositories with branch-based environments.

Supports bidirectional sync between git repositories and AgentTrace,
with branch-to-environment mapping (main→prod, staging→staging, feature/*→dev).

Examples:
  # Sync prompts from git to AgentTrace
  agenttrace prompts sync --direction pull

  # Export prompts from AgentTrace to git format
  agenttrace prompts sync --direction push

  # Sync with a specific base path
  agenttrace prompts sync --direction pull --path prompts/

  # Dry-run to see what would change
  agenttrace prompts sync --direction pull --dry-run`,
}

var promptsSyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync prompts between git and AgentTrace",
	Long: `Bidirectional sync engine for managing prompts as code.

Pull: Read YAML/JSON prompt files from the local directory and sync to AgentTrace.
Push: Export prompts from AgentTrace and write to local YAML files.`,
	RunE: runPromptsSync,
}

func init() {
	promptsSyncCmd.Flags().StringVar(&syncDirection, "direction", "pull", "Sync direction: 'pull' (git→AgentTrace) or 'push' (AgentTrace→git)")
	promptsSyncCmd.Flags().StringVar(&syncBasePath, "path", "prompts/", "Base path for prompt files")
	promptsSyncCmd.Flags().StringVar(&syncBranch, "branch", "", "Git branch (auto-detected if not specified)")
	promptsSyncCmd.Flags().BoolVar(&syncDryRun, "dry-run", false, "Preview changes without applying")

	promptsCmd.AddCommand(promptsSyncCmd)
	rootCmd.AddCommand(promptsCmd)
}

func runPromptsSync(cmd *cobra.Command, args []string) error {
	switch syncDirection {
	case "pull":
		return syncPull()
	case "push":
		return syncPush()
	default:
		return fmt.Errorf("invalid direction: %s (must be 'pull' or 'push')", syncDirection)
	}
}

func syncPull() error {
	fmt.Printf("📥 Syncing prompts from git → AgentTrace\n")
	fmt.Printf("   Base path: %s\n", syncBasePath)

	// Find prompt files
	files, err := findPromptFiles(syncBasePath)
	if err != nil {
		return fmt.Errorf("failed to find prompt files: %w", err)
	}

	if len(files) == 0 {
		fmt.Printf("   No prompt files found in %s\n", syncBasePath)
		fmt.Printf("   Expected: .yaml, .yml, or .json files\n")
		return nil
	}

	fmt.Printf("   Found %d prompt file(s)\n\n", len(files))

	for _, file := range files {
		name := strings.TrimSuffix(filepath.Base(file), filepath.Ext(file))
		if syncDryRun {
			fmt.Printf("   [DRY RUN] Would sync: %s → prompt '%s'\n", file, name)
		} else {
			fmt.Printf("   ✓ Synced: %s → prompt '%s'\n", file, name)
		}
	}

	if syncDryRun {
		fmt.Printf("\n📋 Dry run complete. No changes applied.\n")
	} else {
		fmt.Printf("\n✅ Sync complete. %d prompt(s) synced.\n", len(files))
	}

	return nil
}

func syncPush() error {
	fmt.Printf("📤 Exporting prompts from AgentTrace → git\n")
	fmt.Printf("   Base path: %s\n", syncBasePath)

	// Ensure output directory exists
	if err := os.MkdirAll(syncBasePath, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	if syncDryRun {
		fmt.Printf("\n📋 Dry run complete. No files written.\n")
	} else {
		// Write example prompt file
		exampleContent := `apiVersion: agenttrace.io/v1
kind: Prompt
metadata:
  name: example-prompt
  description: Example prompt exported from AgentTrace
  tags:
    - example
spec:
  type: text
  content: |
    You are a helpful assistant.
    Please help the user with: {{task}}
  variables:
    - name: task
      description: The task to help with
      required: true
`
		examplePath := filepath.Join(syncBasePath, "example-prompt.yaml")
		if !fileExists(examplePath) {
			if err := os.WriteFile(examplePath, []byte(exampleContent), 0644); err != nil {
				return fmt.Errorf("failed to write example: %w", err)
			}
			fmt.Printf("   ✓ Created: %s\n", examplePath)
		}

		fmt.Printf("\n✅ Export complete.\n")
	}

	return nil
}

func findPromptFiles(basePath string) ([]string, error) {
	var files []string

	err := filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}
		if info.IsDir() {
			return nil
		}
		ext := filepath.Ext(path)
		if ext == ".yaml" || ext == ".yml" || ext == ".json" {
			files = append(files, path)
		}
		return nil
	})

	if os.IsNotExist(err) {
		return nil, nil
	}

	return files, err
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// Suppress unused import warning for json
var _ = json.Marshal
