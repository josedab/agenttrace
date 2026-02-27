package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
)

var autoInstrumentCmd = &cobra.Command{
	Use:   "auto-instrument [flags] -- command [args...]",
	Short: "Auto-detect frameworks and instrument with zero configuration",
	Long: `Automatically detect AI frameworks (LangChain, CrewAI, OpenAI, Anthropic, etc.)
and enable instrumentation with zero configuration.

This command wraps your program with automatic framework detection using
sitecustomize.py (Python) or --require hooks (Node.js).

Examples:
  # Python auto-instrumentation
  agenttrace auto-instrument -- python agent.py

  # Node.js auto-instrumentation
  agenttrace auto-instrument -- node agent.js

  # With explicit API key
  agenttrace auto-instrument --api-key sk-at-xxx -- python agent.py`,
	Args:                  cobra.MinimumNArgs(1),
	DisableFlagsInUseLine: true,
	RunE:                  runAutoInstrument,
}

func init() {
	rootCmd.AddCommand(autoInstrumentCmd)
}

func runAutoInstrument(cmd *cobra.Command, args []string) error {
	// Find the command separator
	dashIdx := -1
	for i, arg := range os.Args {
		if arg == "--" {
			dashIdx = i
			break
		}
	}

	var cmdArgs []string
	if dashIdx >= 0 && dashIdx+1 < len(os.Args) {
		cmdArgs = os.Args[dashIdx+1:]
	} else {
		cmdArgs = args
	}

	if len(cmdArgs) == 0 {
		return fmt.Errorf("no command specified after --")
	}

	binary := cmdArgs[0]
	env := os.Environ()

	// Set AgentTrace environment variables
	resolvedKey := apiKey
	if resolvedKey == "" {
		resolvedKey = os.Getenv("AGENTTRACE_API_KEY")
	}
	if resolvedKey != "" {
		env = setEnv(env, "AGENTTRACE_API_KEY", resolvedKey)
	}

	resolvedHost := host
	if resolvedHost != "" {
		env = setEnv(env, "AGENTTRACE_HOST", resolvedHost)
	}

	// Detect runtime and configure auto-instrumentation
	switch detectRuntime(binary) {
	case "python":
		env = configurePythonAutoInstrument(env)
	case "node":
		env = configureNodeAutoInstrument(env)
	default:
		if verbose {
			fmt.Fprintf(os.Stderr, "[agenttrace] Unknown runtime for %q, setting env vars only\n", binary)
		}
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "[agenttrace] Auto-instrumenting: %s\n", strings.Join(cmdArgs, " "))
	}

	// Execute the wrapped command
	execPath, err := exec.LookPath(binary)
	if err != nil {
		return fmt.Errorf("command not found: %s", binary)
	}

	return syscall.Exec(execPath, cmdArgs, env)
}

func detectRuntime(binary string) string {
	base := filepath.Base(binary)
	switch {
	case strings.HasPrefix(base, "python"):
		return "python"
	case base == "node" || base == "npx" || base == "tsx" || base == "ts-node":
		return "node"
	default:
		return "unknown"
	}
}

func configurePythonAutoInstrument(env []string) []string {
	// Find the agenttrace hooks directory for sitecustomize.py
	// This adds the hooks directory to PYTHONPATH so sitecustomize.py runs at startup
	hooksDir := findPythonHooksDir()
	if hooksDir != "" {
		env = prependEnv(env, "PYTHONPATH", hooksDir)
		if verbose {
			fmt.Fprintf(os.Stderr, "[agenttrace] Python auto-instrument: added %s to PYTHONPATH\n", hooksDir)
		}
	}
	env = setEnv(env, "AGENTTRACE_AUTO_INSTRUMENT", "1")
	return env
}

func configureNodeAutoInstrument(env []string) []string {
	// For Node.js, use --require hook or NODE_OPTIONS
	requirePath := findNodeRequireHook()
	if requirePath != "" {
		existingOpts := ""
		for _, e := range env {
			if strings.HasPrefix(e, "NODE_OPTIONS=") {
				existingOpts = strings.TrimPrefix(e, "NODE_OPTIONS=")
				break
			}
		}
		nodeOpts := fmt.Sprintf("--require %s", requirePath)
		if existingOpts != "" {
			nodeOpts = existingOpts + " " + nodeOpts
		}
		env = setEnv(env, "NODE_OPTIONS", nodeOpts)
		if verbose {
			fmt.Fprintf(os.Stderr, "[agenttrace] Node.js auto-instrument: set NODE_OPTIONS=%s\n", nodeOpts)
		}
	}
	env = setEnv(env, "AGENTTRACE_AUTO_INSTRUMENT", "1")
	return env
}

func findPythonHooksDir() string {
	// Try common locations for the agenttrace hooks directory
	candidates := []string{
		"./venv/lib/python*/site-packages/agenttrace/hooks",
		"./.venv/lib/python*/site-packages/agenttrace/hooks",
	}
	for _, pattern := range candidates {
		matches, err := filepath.Glob(pattern)
		if err == nil && len(matches) > 0 {
			return matches[0]
		}
	}
	return ""
}

func findNodeRequireHook() string {
	candidates := []string{
		"./node_modules/@agenttrace/sdk/dist/auto-instrument.js",
		"./node_modules/agenttrace/dist/auto-instrument.js",
	}
	for _, path := range candidates {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	return ""
}

func setEnv(env []string, key, value string) []string {
	prefix := key + "="
	for i, e := range env {
		if strings.HasPrefix(e, prefix) {
			env[i] = prefix + value
			return env
		}
	}
	return append(env, prefix+value)
}

func prependEnv(env []string, key, value string) []string {
	prefix := key + "="
	for i, e := range env {
		if strings.HasPrefix(e, prefix) {
			existing := strings.TrimPrefix(e, prefix)
			env[i] = prefix + value + string(os.PathListSeparator) + existing
			return env
		}
	}
	return append(env, prefix+value)
}
