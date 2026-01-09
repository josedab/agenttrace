package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	// Version is set at build time
	Version = "0.1.0"

	// Global flags
	apiKey  string
	host    string
	verbose bool
)

var rootCmd = &cobra.Command{
	Use:   "agenttrace",
	Short: "AgentTrace CLI - Observability for AI Coding Agents",
	Long: `AgentTrace CLI provides tools for tracing and observing AI coding agents.

Commands:
  wrap    - Wrap a command and trace its execution
  mcp     - Start an MCP server for IDE integration

Example:
  agenttrace wrap -- python agent.py
  agenttrace wrap --name "my-agent" -- npm run dev
  agenttrace mcp --port 8080`,
	Version: Version,
}

func init() {
	rootCmd.PersistentFlags().StringVar(&apiKey, "api-key", "", "AgentTrace API key (or set AGENTTRACE_API_KEY)")
	rootCmd.PersistentFlags().StringVar(&host, "host", "https://api.agenttrace.io", "AgentTrace API host")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")

	// Add subcommands
	rootCmd.AddCommand(wrapCmd)
	rootCmd.AddCommand(mcpCmd)
}

// Execute runs the CLI
func Execute() error {
	return rootCmd.Execute()
}

// getAPIKey returns the API key from flag or environment
func getAPIKey() string {
	if apiKey != "" {
		return apiKey
	}
	return os.Getenv("AGENTTRACE_API_KEY")
}

// logVerbose logs a message if verbose mode is enabled
func logVerbose(format string, args ...any) {
	if verbose {
		fmt.Fprintf(os.Stderr, "[agenttrace] "+format+"\n", args...)
	}
}
