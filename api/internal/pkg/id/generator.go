package id

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// TraceIDLength is the length of a W3C-compliant trace ID (32 hex chars = 16 bytes)
const TraceIDLength = 16

// SpanIDLength is the length of a W3C-compliant span ID (16 hex chars = 8 bytes)
const SpanIDLength = 8

var (
	randReader = rand.Reader

	// traceIDPool reuses buffers for trace ID generation (16 bytes)
	traceIDPool = sync.Pool{
		New: func() any {
			b := make([]byte, TraceIDLength)
			return &b
		},
	}

	// spanIDPool reuses buffers for span ID generation (8 bytes)
	spanIDPool = sync.Pool{
		New: func() any {
			b := make([]byte, SpanIDLength)
			return &b
		},
	}
)

// NewTraceID generates a new W3C-compliant trace ID (32 hex characters)
func NewTraceID() string {
	bufPtr := traceIDPool.Get().(*[]byte)
	defer traceIDPool.Put(bufPtr)
	buf := *bufPtr

	if _, err := randReader.Read(buf); err != nil {
		// Fallback to time-based ID if random fails
		return fmt.Sprintf("%016x%016x", time.Now().UnixNano(), time.Now().UnixNano())
	}

	return hex.EncodeToString(buf)
}

// NewSpanID generates a new W3C-compliant span ID (16 hex characters)
func NewSpanID() string {
	bufPtr := spanIDPool.Get().(*[]byte)
	defer spanIDPool.Put(bufPtr)
	buf := *bufPtr

	if _, err := randReader.Read(buf); err != nil {
		// Fallback to time-based ID if random fails
		return fmt.Sprintf("%016x", time.Now().UnixNano())
	}

	return hex.EncodeToString(buf)
}

// NewUUID generates a new UUID v4
func NewUUID() string {
	return uuid.New().String()
}

// NewUUIDFromString parses a UUID from string
func NewUUIDFromString(s string) (uuid.UUID, error) {
	return uuid.Parse(s)
}

// ValidateTraceID validates a trace ID format
func ValidateTraceID(id string) bool {
	if len(id) != 32 {
		return false
	}
	_, err := hex.DecodeString(id)
	return err == nil
}

// ValidateSpanID validates a span ID format
func ValidateSpanID(id string) bool {
	if len(id) != 16 {
		return false
	}
	_, err := hex.DecodeString(id)
	return err == nil
}

// ValidateUUID validates a UUID format
func ValidateUUID(id string) bool {
	_, err := uuid.Parse(id)
	return err == nil
}

// ParseUUID parses and validates a UUID string
func ParseUUID(id string) (uuid.UUID, error) {
	return uuid.Parse(id)
}

// ParseUUIDOrNil parses a UUID string, returning uuid.Nil on error.
// This is a safe alternative for user input that doesn't require error handling.
func ParseUUIDOrNil(id string) uuid.UUID {
	u, err := uuid.Parse(id)
	if err != nil {
		return uuid.Nil
	}
	return u
}

// ParseUUIDWithDefault parses a UUID string, returning the default on error.
// This is a safe alternative to MustParseUUID for user input.
func ParseUUIDWithDefault(id string, defaultUUID uuid.UUID) uuid.UUID {
	u, err := uuid.Parse(id)
	if err != nil {
		return defaultUUID
	}
	return u
}

// NewAPIKeyPublic generates a new public API key
func NewAPIKeyPublic() string {
	return "pk-at-" + generateRandomString(24)
}

// NewAPIKeySecret generates a new secret API key
func NewAPIKeySecret() string {
	return "sk-at-" + generateRandomString(32)
}

// generateRandomString generates a random alphanumeric string
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	buf := make([]byte, length)
	if _, err := randReader.Read(buf); err != nil {
		// Fallback using time
		for i := range buf {
			buf[i] = charset[time.Now().UnixNano()%int64(len(charset))]
		}
		return string(buf)
	}

	for i := range buf {
		buf[i] = charset[int(buf[i])%len(charset)]
	}
	return string(buf)
}

// NewInvitationToken generates a new invitation token
func NewInvitationToken() string {
	return generateRandomString(48)
}
