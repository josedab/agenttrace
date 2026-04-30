package service

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

func TestRedactTextPreservesWhitespaceAndNewlines(t *testing.T) {
	redactor := NewSensitiveDataRedactor()
	value := "line one\n\tindented line\n\nfinal   line with  double  spaces\n"

	redacted := redactor.RedactText(value)

	assert.Equal(t, value, redacted)
}

func TestRedactTextLeavesCleanTextUnchanged(t *testing.T) {
	redactor := NewSensitiveDataRedactor()
	value := "Deployment succeeded for project alpha (build 42)."

	assert.Equal(t, value, redactor.RedactText(value))
}

func TestRedactTextRedactsOnlyCredentialURLSubstring(t *testing.T) {
	redactor := NewSensitiveDataRedactor()
	value := "connect to postgres://admin:s3cr3t@db.internal:5432/agenttrace now"

	redacted := redactor.RedactText(value)

	assert.Equal(
		t,
		"connect to postgres://[REDACTED:credentials]@db.internal:5432/agenttrace now",
		redacted,
	)
	assert.NotContains(t, redacted, "s3cr3t")
	assert.NotContains(t, redacted, "admin")
}

func TestRedactTextKeepsMultilineLayoutAroundCredentialURL(t *testing.T) {
	redactor := NewSensitiveDataRedactor()
	value := "step 1\n  dsn: https://user:pass@example.com/path\n  retries: 3\n"

	redacted := redactor.RedactText(value)

	lines := strings.Split(redacted, "\n")
	require.Len(t, lines, 4)
	assert.Equal(t, "step 1", lines[0])
	assert.Equal(t, "  dsn: https://[REDACTED:credentials]@example.com/path", lines[1])
	assert.Equal(t, "  retries: 3", lines[2])
	assert.Equal(t, "", lines[3])
}

func TestRedactTextLeavesNonCredentialURLsIntact(t *testing.T) {
	redactor := NewSensitiveDataRedactor()
	value := "docs at https://example.com/guide?ref=a and https://example.com/#anchor"

	assert.Equal(t, value, redactor.RedactText(value))
}

func TestRedactTextRedactsAPIKeys(t *testing.T) {
	redactor := NewSensitiveDataRedactor()

	redacted := redactor.RedactText("key sk-at-abcdefghijklmnop used\nsecond line")

	assert.Equal(t, "key [REDACTED:api-key] used\nsecond line", redacted)
}

func TestRedactTextRedactsBearerTokens(t *testing.T) {
	redactor := NewSensitiveDataRedactor()

	redacted := redactor.RedactText("Authorization: Bearer abcdef.ghijkl.mnopqr\nnext")

	assert.Equal(t, "Authorization: [REDACTED:bearer-token]\nnext", redacted)
}

func TestRedactTextRedactsPrivateKeyBlocks(t *testing.T) {
	redactor := NewSensitiveDataRedactor()
	value := "before\n-----BEGIN RSA PRIVATE KEY-----\nMIIEow\nkeymaterial\n" +
		"-----END RSA PRIVATE KEY-----\nafter"

	redacted := redactor.RedactText(value)

	assert.Equal(t, "before\n[REDACTED:private-key]\nafter", redacted)
	assert.NotContains(t, redacted, "keymaterial")
}

func TestRedactTextIsIdempotent(t *testing.T) {
	redactor := NewSensitiveDataRedactor()
	value := "dsn https://user:pass@example.com\nkey sk-at-abcdefghijklmnop\n"

	once := redactor.RedactText(value)
	twice := redactor.RedactText(once)

	assert.Equal(t, once, twice)
}

// TestRedactTextConsumesFullSecretKey uses the exact secret-key format produced
// by AuthService.generateAPIKeyPair ("sk-at-<id>.<secret>") and asserts that
// neither the id nor the dotted secret suffix survives redaction.
func TestRedactTextConsumesFullSecretKey(t *testing.T) {
	redactor := NewSensitiveDataRedactor()
	authService := &AuthService{}
	_, fullKey, err := authService.generateAPIKeyPair()
	require.NoError(t, err)
	keyParts := strings.Split(strings.TrimPrefix(fullKey, "sk-at-"), ".")
	require.Len(t, keyParts, 2)
	keyID, secret := keyParts[0], keyParts[1]

	value := "before " + fullKey + " after\nsecond line"
	redacted := redactor.RedactText(value)

	assert.Equal(t, "before [REDACTED:api-key] after\nsecond line", redacted)
	assert.NotContains(t, redacted, keyID, "key id must not survive")
	assert.NotContains(t, redacted, secret, "secret suffix must not survive")
	assert.NotContains(t, redacted, "sk-at-")

	// Idempotent.
	assert.Equal(t, redacted, redactor.RedactText(redacted))
}

// TestRedactTextConsumesEmbedToken uses the exact embed-token format produced by
// EmbedService.GenerateToken ("at_embed_<hex>").
func TestRedactTextConsumesEmbedToken(t *testing.T) {
	redactor := NewSensitiveDataRedactor()
	embedService := NewEmbedService(zap.NewNop())
	projectID := uuid.New()
	_, err := embedService.CreateConfig(
		context.Background(),
		projectID,
		&domain.EmbedConfigInput{},
	)
	require.NoError(t, err)
	generated, err := embedService.GenerateToken(context.Background(), projectID)
	require.NoError(t, err)
	token := generated.Token
	hexPart := strings.TrimPrefix(token, "at_embed_")
	require.NotEqual(t, token, hexPart)

	value := "token=" + token + "\ttrailing"
	redacted := redactor.RedactText(value)

	assert.Equal(t, "token=[REDACTED:embed-token]\ttrailing", redacted)
	assert.NotContains(t, redacted, hexPart, "embed secret must not survive")
	assert.NotContains(t, redacted, "at_embed_")

	// Whitespace (tab) is preserved and redaction is idempotent.
	assert.Contains(t, redacted, "\t")
	assert.Equal(t, redacted, redactor.RedactText(redacted))
}

// TestRedactTextConsumesPriorKeyFormats guards the legacy prefixes that predate
// the dotted secret format.
func TestRedactTextConsumesPriorKeyFormats(t *testing.T) {
	redactor := NewSensitiveDataRedactor()
	cases := []string{
		"sk-abcdefghijklmnop",
		"sk_abcdefghijklmnop",
		"pk-abcdefghijklmnop",
		"pk_abcdefghijklmnop",
		"at_test_key_123456",
		"ghp_abcdefghijklmnop1234",
		"github_pat_abcdefghijklmnop1234",
	}
	for _, secret := range cases {
		value := "leading " + secret + " trailing"
		redacted := redactor.RedactText(value)
		assert.Equal(t, "leading [REDACTED:api-key] trailing", redacted, "format %q", secret)
		assert.NotContains(t, redacted, secret)
		assert.Equal(t, redacted, redactor.RedactText(redacted))
	}
}

// TestRedactTextSecretKeyPreservesSurroundingPunctuation ensures a trailing
// sentence period next to a full secret key is not swallowed by the pattern.
func TestRedactTextSecretKeyPreservesSurroundingPunctuation(t *testing.T) {
	redactor := NewSensitiveDataRedactor()
	fullKey := "sk-at-0123456789abcdef0123456789abcdef.fedcba9876543210fedcba9876543210"

	redacted := redactor.RedactText("The key is " + fullKey + ". Done.")

	assert.Equal(t, "The key is [REDACTED:api-key]. Done.", redacted)
}
