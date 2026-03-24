package cmd

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// validPythonPackageName matches safe Python package names (lowercase, underscores only).
var validPythonPackageName = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Diagnose AgentTrace setup and connectivity",
	Long: `Run diagnostic checks to verify your AgentTrace setup:
  - API key configuration
  - Server connectivity
  - SDK installation (Python, Node.js, Go)
  - Framework detection
  - Runtime environment

Examples:
  agenttrace doctor
  agenttrace doctor --verbose`,
	RunE: runDoctor,
}

func init() {
	rootCmd.AddCommand(doctorCmd)
}

type checkResult struct {
	name    string
	status  string // "pass", "warn", "fail"
	message string
}

func runDoctor(cmd *cobra.Command, args []string) error {
	fmt.Println("🩺 AgentTrace Doctor")
	fmt.Println(strings.Repeat("=", 50))
	fmt.Println()

	var results []checkResult

	// Check 1: API Key
	results = append(results, checkAPIKey())

	// Check 2: Server connectivity
	results = append(results, checkServerConnectivity())

	// Check 3: Python SDK
	results = append(results, checkPythonSDK())

	// Check 4: Node.js SDK
	results = append(results, checkNodeSDK())

	// Check 5: Go SDK
	results = append(results, checkGoSDK())

	// Check 6: Python frameworks
	results = append(results, checkPythonFrameworks()...)

	// Check 7: Runtime environment
	results = append(results, checkRuntimeEnvironment())

	// Print results
	passCount, warnCount, failCount := 0, 0, 0
	for _, r := range results {
		icon := "✅"
		switch r.status {
		case "warn":
			icon = "⚠️"
			warnCount++
		case "fail":
			icon = "❌"
			failCount++
		default:
			passCount++
		}
		fmt.Printf("  %s %s: %s\n", icon, r.name, r.message)
	}

	fmt.Println()
	fmt.Printf("Results: %d passed, %d warnings, %d failed\n", passCount, warnCount, failCount)

	if failCount > 0 {
		fmt.Println("\n💡 Run with --verbose for more details on failures.")
		return fmt.Errorf("%d check(s) failed", failCount)
	}

	if passCount > 0 && failCount == 0 {
		fmt.Println("\n🎉 AgentTrace is ready! Time-to-first-trace: run your agent with tracing enabled.")
	}

	return nil
}

func checkAPIKey() checkResult {
	key := apiKey
	if key == "" {
		key = os.Getenv("AGENTTRACE_API_KEY")
	}

	if key == "" {
		return checkResult{
			name:    "API Key",
			status:  "warn",
			message: "Not configured (set AGENTTRACE_API_KEY or use --api-key)",
		}
	}

	if !strings.HasPrefix(key, "sk-at-") {
		return checkResult{
			name:    "API Key",
			status:  "warn",
			message: fmt.Sprintf("Key format unusual (prefix: %s...)", key[:min(6, len(key))]),
		}
	}

	return checkResult{
		name:    "API Key",
		status:  "pass",
		message: fmt.Sprintf("Configured (%s...)", key[:min(10, len(key))]),
	}
}

func checkServerConnectivity() checkResult {
	resolvedHost := host
	if resolvedHost == "" {
		resolvedHost = os.Getenv("AGENTTRACE_HOST")
	}
	if resolvedHost == "" {
		resolvedHost = "https://api.agenttrace.io"
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(resolvedHost + "/api/public/health")
	if err != nil {
		return checkResult{
			name:    "Server Connectivity",
			status:  "fail",
			message: fmt.Sprintf("Cannot reach %s (%v)", resolvedHost, err),
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		return checkResult{
			name:    "Server Connectivity",
			status:  "pass",
			message: fmt.Sprintf("Connected to %s", resolvedHost),
		}
	}

	return checkResult{
		name:    "Server Connectivity",
		status:  "warn",
		message: fmt.Sprintf("Server returned HTTP %d", resp.StatusCode),
	}
}

func checkPythonSDK() checkResult {
	out, err := exec.Command("python3", "-c", "import agenttrace; print(agenttrace.__version__)").Output()
	if err != nil {
		out2, err2 := exec.Command("python", "-c", "import agenttrace; print(agenttrace.__version__)").Output()
		if err2 != nil {
			return checkResult{
				name:    "Python SDK",
				status:  "warn",
				message: "Not installed (pip install agenttrace)",
			}
		}
		out = out2
	}

	version := strings.TrimSpace(string(out))
	return checkResult{
		name:    "Python SDK",
		status:  "pass",
		message: fmt.Sprintf("Installed (v%s)", version),
	}
}

func checkNodeSDK() checkResult {
	out, err := exec.Command("node", "-e", "try{console.log(require('agenttrace').version)}catch(e){console.log('not-found')}").Output()
	if err != nil {
		return checkResult{
			name:    "Node.js SDK",
			status:  "warn",
			message: "Node.js not available or SDK not installed",
		}
	}

	version := strings.TrimSpace(string(out))
	if version == "not-found" || version == "" {
		return checkResult{
			name:    "Node.js SDK",
			status:  "warn",
			message: "Not installed (npm install agenttrace)",
		}
	}

	return checkResult{
		name:    "Node.js SDK",
		status:  "pass",
		message: fmt.Sprintf("Installed (v%s)", version),
	}
}

func checkGoSDK() checkResult {
	// Check if agenttrace-go is in go.sum or go.mod in current dir
	if _, err := os.Stat("go.mod"); err != nil {
		return checkResult{
			name:    "Go SDK",
			status:  "warn",
			message: "Not a Go project (no go.mod)",
		}
	}

	data, err := os.ReadFile("go.sum")
	if err != nil {
		return checkResult{
			name:    "Go SDK",
			status:  "warn",
			message: "Cannot read go.sum",
		}
	}

	if strings.Contains(string(data), "github.com/agenttrace/agenttrace-go") {
		return checkResult{
			name:    "Go SDK",
			status:  "pass",
			message: "Installed (found in go.sum)",
		}
	}

	return checkResult{
		name:    "Go SDK",
		status:  "warn",
		message: "Not installed (go get github.com/agenttrace/agenttrace-go)",
	}
}

func checkPythonFrameworks() []checkResult {
	var results []checkResult

	frameworks := map[string]string{
		"openai":     "OpenAI",
		"anthropic":  "Anthropic",
		"langchain":  "LangChain",
		"llama_index": "LlamaIndex",
		"crewai":     "CrewAI",
		"autogen":    "AutoGen",
	}

	for pkg, name := range frameworks {
		if !validPythonPackageName.MatchString(pkg) {
			continue
		}
		out, err := exec.Command("python3", "-c", fmt.Sprintf("import %s; print(%s.__version__)", pkg, pkg)).Output()
		if err != nil {
			continue
		}
		version := strings.TrimSpace(string(out))
		results = append(results, checkResult{
			name:    fmt.Sprintf("Framework: %s", name),
			status:  "pass",
			message: fmt.Sprintf("Detected v%s (auto-instrumentable)", version),
		})
	}

	if len(results) == 0 {
		results = append(results, checkResult{
			name:    "AI Frameworks",
			status:  "warn",
			message: "No supported frameworks detected",
		})
	}

	return results
}

func checkRuntimeEnvironment() checkResult {
	return checkResult{
		name:   "Runtime Environment",
		status: "pass",
		message: fmt.Sprintf("OS=%s ARCH=%s Go=%s", runtime.GOOS, runtime.GOARCH, runtime.Version()),
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
