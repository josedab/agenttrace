package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

const agentTraceConfigName = ".agenttrace.yaml"

var (
	initDirectory string
	initForce     bool
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize AgentTrace in an existing project",
	Long: `Detect the current project runtime and supported AI frameworks, then create a
minimal .agenttrace.yaml configuration. Existing files are never overwritten
unless --force is supplied. API keys are referenced by environment variable and
are never written to disk.`,
	RunE: runInit,
}

func configureInitCommand() {
	initCmd.Flags().StringVarP(&initDirectory, "dir", "d", ".", "Project directory to initialize")
	initCmd.Flags().BoolVarP(&initForce, "force", "f", false, "Replace an existing generated config")
}

type detectedProject struct {
	Name       string
	Runtime    string
	Frameworks []string
	Command    string
}

func runInit(cmd *cobra.Command, _ []string) error {
	directory, err := filepath.Abs(initDirectory)
	if err != nil {
		return fmt.Errorf("resolve project directory: %w", err)
	}
	info, err := os.Stat(directory)
	if err != nil {
		return fmt.Errorf("inspect project directory: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("project path is not a directory: %s", directory)
	}

	project, err := detectProject(directory)
	if err != nil {
		return err
	}
	configPath := filepath.Join(directory, agentTraceConfigName)
	config := renderAgentTraceConfig(project)
	if err := writeConfigFile(configPath, []byte(config), initForce); err != nil {
		return err
	}

	out := cmd.OutOrStdout()
	fmt.Fprintf(out, "Initialized AgentTrace in %s\n", directory)
	fmt.Fprintf(out, "  Runtime: %s\n", project.Runtime)
	if len(project.Frameworks) > 0 {
		fmt.Fprintf(out, "  Frameworks: %s\n", strings.Join(project.Frameworks, ", "))
	} else {
		fmt.Fprintln(out, "  Frameworks: none detected")
	}
	fmt.Fprintf(out, "  Config: %s\n\n", configPath)
	fmt.Fprintln(out, "Next steps:")
	fmt.Fprintln(out, "  1. Export AGENTTRACE_API_KEY=sk-at-...")
	fmt.Fprintf(out, "  2. %s\n", project.Command)
	fmt.Fprintln(out, "  3. Run agenttrace doctor")
	return nil
}

func detectProject(directory string) (detectedProject, error) {
	project := detectedProject{
		Name:    filepath.Base(directory),
		Runtime: "generic",
		Command: "agenttrace wrap -- <your-agent-command>",
	}

	switch {
	case regularFileExists(filepath.Join(directory, "pyproject.toml")) ||
		regularFileExists(filepath.Join(directory, "requirements.txt")):
		project.Runtime = "python"
		project.Command = "agenttrace auto-instrument -- python <your-agent-entrypoint>.py"
		project.Frameworks = detectPythonFrameworks(directory)
	case regularFileExists(filepath.Join(directory, "package.json")):
		project.Runtime = "node"
		project.Command = "agenttrace auto-instrument -- npm run <your-agent-script>"
		frameworks, err := detectNodeFrameworks(filepath.Join(directory, "package.json"))
		if err != nil {
			return detectedProject{}, err
		}
		project.Frameworks = frameworks
	case regularFileExists(filepath.Join(directory, "go.mod")):
		project.Runtime = "go"
		project.Command = "agenttrace wrap -- go run ."
		project.Frameworks = detectGoFrameworks(filepath.Join(directory, "go.mod"))
	}

	sort.Strings(project.Frameworks)
	return project, nil
}

func detectPythonFrameworks(directory string) []string {
	var content strings.Builder
	for _, name := range []string{"pyproject.toml", "requirements.txt"} {
		data, err := os.ReadFile(filepath.Join(directory, name))
		if err == nil {
			content.Write(data)
			content.WriteByte('\n')
		}
	}
	return detectFrameworkNames(strings.ToLower(content.String()), map[string][]string{
		"anthropic":   {"anthropic"},
		"autogen":     {"autogen", "pyautogen"},
		"crewai":      {"crewai"},
		"langchain":   {"langchain"},
		"llama-index": {"llama-index", "llama_index"},
		"openai":      {"openai"},
	})
}

func detectNodeFrameworks(packageJSONPath string) ([]string, error) {
	data, err := os.ReadFile(packageJSONPath)
	if err != nil {
		return nil, fmt.Errorf("read package.json: %w", err)
	}
	var manifest struct {
		Dependencies    map[string]string `json:"dependencies"`
		DevDependencies map[string]string `json:"devDependencies"`
	}
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("parse package.json: %w", err)
	}
	names := make([]string, 0, len(manifest.Dependencies)+len(manifest.DevDependencies))
	for name := range manifest.Dependencies {
		names = append(names, strings.ToLower(name))
	}
	for name := range manifest.DevDependencies {
		names = append(names, strings.ToLower(name))
	}
	return detectFrameworkNames("\n"+strings.Join(names, "\n")+"\n", map[string][]string{
		"anthropic": {"@anthropic-ai/sdk"},
		"langchain": {"langchain", "@langchain/"},
		"openai":    {"openai"},
		"vercel-ai": {"\nai\n", "@ai-sdk/"},
	}), nil
}

func detectGoFrameworks(goModPath string) []string {
	data, err := os.ReadFile(goModPath)
	if err != nil {
		return nil
	}
	return detectFrameworkNames(strings.ToLower(string(data)), map[string][]string{
		"anthropic":   {"github.com/anthropics/"},
		"langchaingo": {"github.com/tmc/langchaingo"},
		"openai":      {"github.com/sashabaranov/go-openai", "github.com/openai/openai-go"},
	})
}

func detectFrameworkNames(content string, candidates map[string][]string) []string {
	result := make([]string, 0)
	for name, markers := range candidates {
		for _, marker := range markers {
			if strings.Contains(content, marker) {
				result = append(result, name)
				break
			}
		}
	}
	return result
}

func renderAgentTraceConfig(project detectedProject) string {
	var builder strings.Builder
	builder.WriteString("version: 1\n")
	fmt.Fprintf(&builder, "project:\n  name: %q\n", project.Name)
	builder.WriteString("api:\n")
	builder.WriteString("  host_env: AGENTTRACE_HOST\n")
	builder.WriteString("  key_env: AGENTTRACE_API_KEY\n")
	builder.WriteString("instrumentation:\n")
	fmt.Fprintf(&builder, "  runtime: %s\n", project.Runtime)
	if len(project.Frameworks) == 0 {
		builder.WriteString("  frameworks: []\n")
	} else {
		builder.WriteString("  frameworks:\n")
		for _, framework := range project.Frameworks {
			fmt.Fprintf(&builder, "    - %s\n", framework)
		}
	}
	builder.WriteString("privacy:\n")
	builder.WriteString("  redact_secrets: true\n")
	builder.WriteString("  no_egress: false\n")
	return builder.String()
}

func writeConfigFile(path string, content []byte, force bool) error {
	if info, err := os.Lstat(path); err == nil {
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("refusing to overwrite symlink: %s", path)
		}
		if !force {
			return fmt.Errorf("%s already exists; use --force to replace it", path)
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("inspect existing config: %w", err)
	}

	temp, err := os.CreateTemp(filepath.Dir(path), ".agenttrace-*.tmp")
	if err != nil {
		return fmt.Errorf("create temporary config: %w", err)
	}
	tempPath := temp.Name()
	removeTemp := true
	defer func() {
		_ = temp.Close()
		if removeTemp {
			_ = os.Remove(tempPath)
		}
	}()

	if err := temp.Chmod(0o644); err != nil {
		return fmt.Errorf("set config permissions: %w", err)
	}
	if _, err := io.Copy(temp, strings.NewReader(string(content))); err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	if err := temp.Sync(); err != nil {
		return fmt.Errorf("sync config: %w", err)
	}
	if err := temp.Close(); err != nil {
		return fmt.Errorf("close config: %w", err)
	}
	if err := os.Rename(tempPath, path); err != nil {
		return fmt.Errorf("install config: %w", err)
	}
	removeTemp = false
	return nil
}

func regularFileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
