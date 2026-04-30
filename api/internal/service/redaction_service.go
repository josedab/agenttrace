package service

import (
	"regexp"
)

// credentialURLPattern matches only the credential-bearing prefix of a URL
// (scheme://user:password@) so the surrounding text, including whitespace and
// newlines, is preserved exactly.
var credentialURLPattern = regexp.MustCompile(
	`(?i)\b([a-z][a-z0-9+.\-]*)://[^\s/?#@]+@`,
)

type redactionPattern struct {
	label   string
	pattern *regexp.Regexp
}

// SensitiveDataRedactor applies deterministic server-side redaction.
type SensitiveDataRedactor struct {
	patterns []redactionPattern
}

// NewSensitiveDataRedactor creates the shared redactor for links, digests, and migration errors.
func NewSensitiveDataRedactor() *SensitiveDataRedactor {
	return &SensitiveDataRedactor{
		patterns: []redactionPattern{
			// AgentTrace secret keys are generated as "sk-at-<id>.<secret>"
			// (auth_service.go generateAPIKeyPair). The optional dotted secret
			// suffix must be consumed in full so no secret material survives the
			// prefix; "." is not a word character, so an ordinary trailing
			// sentence period is left intact. This runs before the generic
			// prefix rule below so the whole token collapses to one label.
			{label: "api-key", pattern: regexp.MustCompile(`(?i)\bsk-at-[A-Za-z0-9]{8,}(?:\.[A-Za-z0-9_\-]{8,})?`)},
			// Embed access tokens are generated as "at_embed_<hex>"
			// (embed_service.go GenerateToken).
			{label: "embed-token", pattern: regexp.MustCompile(`(?i)\bat_embed_[A-Za-z0-9]{8,}\b`)},
			// Prior/legacy credential formats without the dotted suffix. Auth
			// still recognizes hyphenated, underscored, and bare at_ forms, so
			// the redactor must consume them as complete tokens too.
			{label: "api-key", pattern: regexp.MustCompile(`(?i)\b(?:sk-|sk_|pk-|pk_|at_|ghp_|github_pat_)[A-Za-z0-9_\-]{8,}\b`)},
			{label: "bearer-token", pattern: regexp.MustCompile(`(?i)\bBearer\s+[A-Za-z0-9._~+/=-]{8,}`)},
			{label: "aws-key", pattern: regexp.MustCompile(`\bAKIA[0-9A-Z]{16}\b`)},
			{label: "email", pattern: regexp.MustCompile(`\b[A-Za-z0-9._%+\-]+@[A-Za-z0-9.\-]+\.[A-Za-z]{2,}\b`)},
			{label: "credit-card", pattern: regexp.MustCompile(`\b(?:\d[ -]*?){13,19}\b`)},
			{label: "private-key", pattern: regexp.MustCompile(`(?s)-----BEGIN [A-Z ]*PRIVATE KEY-----.*?-----END [A-Z ]*PRIVATE KEY-----`)},
		},
	}
}

// RedactText replaces sensitive values with stable labels.
// Credential URLs are handled first so a URL password is never mistaken for an
// email address, which would leave the username visible.
func (r *SensitiveDataRedactor) RedactText(value string) string {
	redacted := redactCredentialURL(value)
	for _, item := range r.patterns {
		redacted = item.pattern.ReplaceAllString(redacted, "[REDACTED:"+item.label+"]")
	}
	return redacted
}

// redactCredentialURL replaces only the userinfo section of credential-bearing
// URLs and leaves every other byte, including whitespace and newlines, intact.
func redactCredentialURL(value string) string {
	return credentialURLPattern.ReplaceAllString(value, "${1}://[REDACTED:credentials]@")
}
