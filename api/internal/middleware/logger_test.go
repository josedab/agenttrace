package middleware

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSanitizeQueryString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		contains string // expected substring in output
		notContains string // must NOT appear in output
	}{
		{
			name:        "api_key is redacted",
			input:       "api_key=secret123&name=test",
			contains:    "%5BREDACTED%5D", // url-encoded [REDACTED]
			notContains: "secret123",
		},
		{
			name:        "token is redacted but name is kept",
			input:       "token=abc&name=test",
			contains:    "name=test",
			notContains: "token=abc",
		},
		{
			name:        "mixed case API_KEY is redacted",
			input:       "API_KEY=myvalue&other=ok",
			notContains: "myvalue",
		},
		{
			name:     "empty query string unchanged",
			input:    "",
			contains: "",
		},
		{
			name:     "malformed query string returned as-is",
			input:    "%;invalid;query",
			contains: "%;invalid;query",
		},
		{
			name:        "multiple sensitive params redacted",
			input:       "api_key=k1&token=t1&password=p1&name=safe",
			notContains: "k1",
		},
		{
			name:        "url-encoded param names",
			input:       "api_key=encoded%20secret&safe=value",
			notContains: "encoded%20secret",
		},
		{
			name:        "password param redacted",
			input:       "password=hunter2",
			notContains: "hunter2",
		},
		{
			name:        "secret param redacted",
			input:       "secret=mysecret",
			notContains: "mysecret",
		},
		{
			name:     "safe params unchanged",
			input:    "page=1&limit=10&sort=name",
			contains: "page=1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeQueryString(tt.input)

			if tt.contains != "" {
				assert.Contains(t, result, tt.contains)
			}
			if tt.notContains != "" {
				assert.NotContains(t, result, tt.notContains)
			}

			// Empty input should return empty
			if tt.input == "" {
				assert.Equal(t, "", result)
			}
		})
	}
}

func TestSanitizeQueryString_RedactedValue(t *testing.T) {
	result := sanitizeQueryString("api_key=secret123")
	// Verify the key is preserved but value is redacted
	assert.Contains(t, result, "api_key")
	assert.Contains(t, result, "REDACTED")
	assert.NotContains(t, result, "secret123")
}

func TestSanitizeQueryString_NoSensitiveParams(t *testing.T) {
	input := "page=1&limit=20"
	result := sanitizeQueryString(input)
	// Should return original unchanged
	assert.Equal(t, input, result)
}
