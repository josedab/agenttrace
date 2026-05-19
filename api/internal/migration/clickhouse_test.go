package migration

import (
	"net/url"
	"testing"
)

func TestNormalizeClickHouseURL(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		username string
		password string
		database string
	}{
		{
			name:     "query parameters",
			input:    "clickhouse://localhost:9000?username=agenttrace&password=secret&database=agenttrace",
			username: "agenttrace",
			password: "secret",
			database: "agenttrace",
		},
		{
			name:     "userinfo and path",
			input:    "clickhouse://agenttrace:secret@127.0.0.1:9000/agenttrace",
			username: "agenttrace",
			password: "secret",
			database: "agenttrace",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			normalized, err := NormalizeClickHouseURL(tt.input)
			if err != nil {
				t.Fatalf("NormalizeClickHouseURL() error = %v", err)
			}

			parsed, err := url.Parse(normalized)
			if err != nil {
				t.Fatalf("parse normalized URL: %v", err)
			}

			query := parsed.Query()
			if got := query.Get("username"); got != tt.username {
				t.Errorf("username = %q, want %q", got, tt.username)
			}
			if got := query.Get("password"); got != tt.password {
				t.Errorf("password = %q, want %q", got, tt.password)
			}
			if got := query.Get("database"); got != tt.database {
				t.Errorf("database = %q, want %q", got, tt.database)
			}
			if got := query.Get("x-multi-statement"); got != "true" {
				t.Errorf("x-multi-statement = %q, want true", got)
			}
			if parsed.User != nil || parsed.Path != "" {
				t.Errorf("normalized URL retained unsupported userinfo or path: %s", normalized)
			}
		})
	}
}

func TestNormalizeClickHouseURLRejectsUnsupportedScheme(t *testing.T) {
	if _, err := NormalizeClickHouseURL("clickhouses://localhost:9440/agenttrace"); err == nil {
		t.Fatal("NormalizeClickHouseURL() expected an error for clickhouses scheme")
	}
}
