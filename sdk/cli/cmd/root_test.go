package cmd

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRootCommand(t *testing.T) {
	t.Run("prints help output", func(t *testing.T) {
		buf := new(bytes.Buffer)
		rootCmd.SetOut(buf)
		rootCmd.SetArgs([]string{"--help"})
		err := rootCmd.Execute()
		require.NoError(t, err)
		assert.Contains(t, buf.String(), "AgentTrace CLI")
		assert.Contains(t, buf.String(), "wrap")
		assert.Contains(t, buf.String(), "mcp")
	})

	t.Run("prints version", func(t *testing.T) {
		buf := new(bytes.Buffer)
		rootCmd.SetOut(buf)
		rootCmd.SetArgs([]string{"--version"})
		err := rootCmd.Execute()
		require.NoError(t, err)
		assert.Contains(t, buf.String(), "agenttrace")
	})

	t.Run("has expected persistent flags", func(t *testing.T) {
		f := rootCmd.PersistentFlags()
		assert.NotNil(t, f.Lookup("api-key"))
		assert.NotNil(t, f.Lookup("host"))
		assert.NotNil(t, f.Lookup("verbose"))
	})

	t.Run("has expected subcommands", func(t *testing.T) {
		commands := rootCmd.Commands()
		names := make([]string, 0, len(commands))
		for _, c := range commands {
			names = append(names, c.Name())
		}
		assert.Contains(t, names, "wrap")
		assert.Contains(t, names, "mcp")
	})
}

func TestGetAPIKey(t *testing.T) {
	t.Run("returns flag value when set", func(t *testing.T) {
		original := apiKey
		defer func() { apiKey = original }()

		apiKey = "test-key-from-flag"
		assert.Equal(t, "test-key-from-flag", getAPIKey())
	})

	t.Run("falls back to environment variable", func(t *testing.T) {
		original := apiKey
		defer func() { apiKey = original }()

		apiKey = ""
		t.Setenv("AGENTTRACE_API_KEY", "test-key-from-env")
		assert.Equal(t, "test-key-from-env", getAPIKey())
	})

	t.Run("returns empty when neither set", func(t *testing.T) {
		original := apiKey
		defer func() { apiKey = original }()

		apiKey = ""
		t.Setenv("AGENTTRACE_API_KEY", "")
		assert.Empty(t, getAPIKey())
	})
}

func TestWrapCommand(t *testing.T) {
	t.Run("has expected flags", func(t *testing.T) {
		f := wrapCmd.Flags()
		assert.NotNil(t, f.Lookup("name"))
		assert.NotNil(t, f.Lookup("user-id"))
		assert.NotNil(t, f.Lookup("session-id"))
		assert.NotNil(t, f.Lookup("tags"))
		assert.NotNil(t, f.Lookup("watch"))
		assert.NotNil(t, f.Lookup("git"))
		assert.NotNil(t, f.Lookup("capture-stdout"))
		assert.NotNil(t, f.Lookup("capture-stderr"))
		assert.NotNil(t, f.Lookup("checkpoints"))
	})

	t.Run("requires minimum 1 argument", func(t *testing.T) {
		buf := new(bytes.Buffer)
		rootCmd.SetOut(buf)
		rootCmd.SetErr(buf)
		rootCmd.SetArgs([]string{"wrap"})
		err := rootCmd.Execute()
		assert.Error(t, err)
	})

	t.Run("requires API key", func(t *testing.T) {
		original := apiKey
		defer func() { apiKey = original }()

		apiKey = ""
		t.Setenv("AGENTTRACE_API_KEY", "")

		buf := new(bytes.Buffer)
		rootCmd.SetOut(buf)
		rootCmd.SetErr(buf)
		rootCmd.SetArgs([]string{"wrap", "--", "echo", "hello"})
		err := rootCmd.Execute()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "API key required")
	})
}

func TestMCPCommand(t *testing.T) {
	t.Run("has port flag with default", func(t *testing.T) {
		f := mcpCmd.Flags()
		portFlag := f.Lookup("port")
		require.NotNil(t, portFlag)
		assert.Equal(t, "8765", portFlag.DefValue)
	})

	t.Run("requires API key", func(t *testing.T) {
		original := apiKey
		defer func() { apiKey = original }()

		apiKey = ""
		t.Setenv("AGENTTRACE_API_KEY", "")

		buf := new(bytes.Buffer)
		rootCmd.SetOut(buf)
		rootCmd.SetErr(buf)
		rootCmd.SetArgs([]string{"mcp"})
		err := rootCmd.Execute()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "API key required")
	})
}
