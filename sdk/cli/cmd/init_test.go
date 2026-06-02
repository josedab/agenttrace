package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitDetectsPythonAndWritesSafeConfig(t *testing.T) {
	directory := t.TempDir()
	require.NoError(t, os.WriteFile(
		filepath.Join(directory, "requirements.txt"),
		[]byte("langchain==0.3.0\nopenai==1.0.0\n"),
		0o644,
	))

	project, err := detectProject(directory)
	require.NoError(t, err)
	assert.Equal(t, "python", project.Runtime)
	assert.Equal(t, []string{"langchain", "openai"}, project.Frameworks)

	config := renderAgentTraceConfig(project)
	assert.Contains(t, config, "runtime: python")
	assert.Contains(t, config, "key_env: AGENTTRACE_API_KEY")
	assert.NotContains(t, config, "sk-at-")

	path := filepath.Join(directory, agentTraceConfigName)
	require.NoError(t, writeConfigFile(path, []byte(config), false))
	written, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, config, string(written))
}

func TestInitDetectsNodeFrameworks(t *testing.T) {
	directory := t.TempDir()
	require.NoError(t, os.WriteFile(
		filepath.Join(directory, "package.json"),
		[]byte(`{"dependencies":{"@anthropic-ai/sdk":"1.0.0","ai":"4.0.0"}}`),
		0o644,
	))

	project, err := detectProject(directory)

	require.NoError(t, err)
	assert.Equal(t, "node", project.Runtime)
	assert.Equal(t, []string{"anthropic", "vercel-ai"}, project.Frameworks)
}

func TestInitDoesNotOverwriteWithoutForce(t *testing.T) {
	path := filepath.Join(t.TempDir(), agentTraceConfigName)
	require.NoError(t, os.WriteFile(path, []byte("existing"), 0o644))

	err := writeConfigFile(path, []byte("replacement"), false)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "--force")
	content, readErr := os.ReadFile(path)
	require.NoError(t, readErr)
	assert.Equal(t, "existing", string(content))

	require.NoError(t, writeConfigFile(path, []byte("replacement"), true))
	content, readErr = os.ReadFile(path)
	require.NoError(t, readErr)
	assert.Equal(t, "replacement", string(content))
}

func TestInitRefusesSymlinkOverwrite(t *testing.T) {
	directory := t.TempDir()
	target := filepath.Join(directory, "target")
	require.NoError(t, os.WriteFile(target, []byte("secret"), 0o600))
	link := filepath.Join(directory, agentTraceConfigName)
	require.NoError(t, os.Symlink(target, link))

	err := writeConfigFile(link, []byte("replacement"), true)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "symlink")
	content, readErr := os.ReadFile(target)
	require.NoError(t, readErr)
	assert.Equal(t, "secret", string(content))
}

func TestInitCommandPrintsNextSteps(t *testing.T) {
	directory := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(directory, "go.mod"), []byte("module example.com/app"), 0o644))
	originalDirectory, originalForce := initDirectory, initForce
	defer func() {
		initDirectory, initForce = originalDirectory, originalForce
	}()
	initDirectory, initForce = directory, false
	output := new(bytes.Buffer)
	initCmd.SetOut(output)

	err := runInit(initCmd, nil)

	require.NoError(t, err)
	assert.Contains(t, output.String(), "Runtime: go")
	assert.Contains(t, output.String(), "Export AGENTTRACE_API_KEY")
	assert.Contains(t, output.String(), "agenttrace wrap -- go run .")
}
