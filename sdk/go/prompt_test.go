package agenttrace

import (
	"sort"
	"testing"
)

func TestPromptVersion_Compile(t *testing.T) {
	t.Run("compiles with double brace variables", func(t *testing.T) {
		prompt := &PromptVersion{
			Prompt: "Hello, {{name}}! Welcome to {{place}}.",
		}

		result := prompt.Compile(map[string]any{
			"name":  "Alice",
			"place": "AgentTrace",
		})

		expected := "Hello, Alice! Welcome to AgentTrace."
		if result != expected {
			t.Errorf("expected '%s', got '%s'", expected, result)
		}
	})

	t.Run("compiles with single brace variables", func(t *testing.T) {
		prompt := &PromptVersion{
			Prompt: "Hello, {name}!",
		}

		result := prompt.Compile(map[string]any{
			"name": "Bob",
		})

		expected := "Hello, Bob!"
		if result != expected {
			t.Errorf("expected '%s', got '%s'", expected, result)
		}
	})

	t.Run("compiles with mixed brace styles", func(t *testing.T) {
		prompt := &PromptVersion{
			Prompt: "Hello, {{name}}! Your score is {score}.",
		}

		result := prompt.Compile(map[string]any{
			"name":  "Charlie",
			"score": 95,
		})

		expected := "Hello, Charlie! Your score is 95."
		if result != expected {
			t.Errorf("expected '%s', got '%s'", expected, result)
		}
	})

	t.Run("handles multiple occurrences of same variable", func(t *testing.T) {
		prompt := &PromptVersion{
			Prompt: "{{name}} said hello. {{name}} is happy.",
		}

		result := prompt.Compile(map[string]any{
			"name": "Eve",
		})

		expected := "Eve said hello. Eve is happy."
		if result != expected {
			t.Errorf("expected '%s', got '%s'", expected, result)
		}
	})

	t.Run("leaves unmatched variables as-is", func(t *testing.T) {
		prompt := &PromptVersion{
			Prompt: "Hello, {{name}}! Your {{missing}} is here.",
		}

		result := prompt.Compile(map[string]any{
			"name": "Frank",
		})

		expected := "Hello, Frank! Your {{missing}} is here."
		if result != expected {
			t.Errorf("expected '%s', got '%s'", expected, result)
		}
	})
}

func TestPromptVersion_CompileChat(t *testing.T) {
	t.Run("parses simple chat format", func(t *testing.T) {
		prompt := &PromptVersion{
			Prompt: "system: You are a helpful assistant.\nuser: Hello, {{name}}!",
		}

		messages := prompt.CompileChat(map[string]any{
			"name": "Grace",
		})

		if len(messages) != 2 {
			t.Fatalf("expected 2 messages, got %d", len(messages))
		}

		if messages[0].Role != "system" {
			t.Errorf("expected role 'system', got '%s'", messages[0].Role)
		}
		if messages[0].Content != "You are a helpful assistant." {
			t.Errorf("unexpected content: '%s'", messages[0].Content)
		}

		if messages[1].Role != "user" {
			t.Errorf("expected role 'user', got '%s'", messages[1].Role)
		}
		if messages[1].Content != "Hello, Grace!" {
			t.Errorf("unexpected content: '%s'", messages[1].Content)
		}
	})

	t.Run("handles multi-line content", func(t *testing.T) {
		prompt := &PromptVersion{
			Prompt: "system: You are a helpful assistant.\nBe concise.\nuser: Tell me about {{topic}}.",
		}

		messages := prompt.CompileChat(map[string]any{
			"topic": "Go",
		})

		if len(messages) != 2 {
			t.Fatalf("expected 2 messages, got %d", len(messages))
		}

		expected := "You are a helpful assistant.\nBe concise."
		if messages[0].Content != expected {
			t.Errorf("expected '%s', got '%s'", expected, messages[0].Content)
		}
	})

	t.Run("supports all role types", func(t *testing.T) {
		prompt := &PromptVersion{
			Prompt: "system: System message\nuser: User message\nassistant: Assistant message\nfunction: Function result",
		}

		messages := prompt.CompileChat(nil)

		if len(messages) != 4 {
			t.Fatalf("expected 4 messages, got %d", len(messages))
		}

		roles := []string{messages[0].Role, messages[1].Role, messages[2].Role, messages[3].Role}
		expected := []string{"system", "user", "assistant", "function"}

		for i, role := range roles {
			if role != expected[i] {
				t.Errorf("expected role '%s' at index %d, got '%s'", expected[i], i, role)
			}
		}
	})

	t.Run("is case-insensitive for roles", func(t *testing.T) {
		prompt := &PromptVersion{
			Prompt: "System: System message\nUSER: User message\nAssistant: Assistant message",
		}

		messages := prompt.CompileChat(nil)

		if len(messages) != 3 {
			t.Fatalf("expected 3 messages, got %d", len(messages))
		}

		// All roles should be lowercase
		for _, msg := range messages {
			if msg.Role != "system" && msg.Role != "user" && msg.Role != "assistant" {
				t.Errorf("expected lowercase role, got '%s'", msg.Role)
			}
		}
	})
}

func TestPromptVersion_GetVariables(t *testing.T) {
	t.Run("extracts double brace variables", func(t *testing.T) {
		prompt := &PromptVersion{
			Prompt: "Hello, {{name}}! Welcome to {{place}}.",
		}

		variables := prompt.GetVariables()
		sort.Strings(variables)

		expected := []string{"name", "place"}
		sort.Strings(expected)

		if len(variables) != len(expected) {
			t.Fatalf("expected %d variables, got %d", len(expected), len(variables))
		}

		for i, v := range variables {
			if v != expected[i] {
				t.Errorf("expected '%s', got '%s'", expected[i], v)
			}
		}
	})

	t.Run("extracts single brace variables", func(t *testing.T) {
		prompt := &PromptVersion{
			Prompt: "Score: {score}, Level: {level}",
		}

		variables := prompt.GetVariables()

		if len(variables) != 2 {
			t.Fatalf("expected 2 variables, got %d", len(variables))
		}
	})

	t.Run("deduplicates variables", func(t *testing.T) {
		prompt := &PromptVersion{
			Prompt: "{{name}} said hello. {{name}} is {{name}}.",
		}

		variables := prompt.GetVariables()

		if len(variables) != 1 {
			t.Errorf("expected 1 variable, got %d", len(variables))
		}
		if variables[0] != "name" {
			t.Errorf("expected 'name', got '%s'", variables[0])
		}
	})

	t.Run("returns empty slice for no variables", func(t *testing.T) {
		prompt := &PromptVersion{
			Prompt: "Hello, world!",
		}

		variables := prompt.GetVariables()

		if len(variables) != 0 {
			t.Errorf("expected empty slice, got %d variables", len(variables))
		}
	})
}

func TestPromptCache(t *testing.T) {
	t.Run("ClearPromptCache clears the cache", func(t *testing.T) {
		ClearPromptCache()
		// No panic means success
	})

	t.Run("SetPromptCacheTTL sets TTL", func(t *testing.T) {
		SetPromptCacheTTL(30 * 1000 * 1000 * 1000) // 30 seconds in nanoseconds
		// No panic means success
	})

	t.Run("InvalidatePrompt removes entries", func(t *testing.T) {
		InvalidatePrompt("test-prompt")
		// No panic means success
	})
}
