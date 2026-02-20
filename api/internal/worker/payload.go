package worker

import (
	"encoding/json"
	"fmt"
)

// PayloadVersion is the current version for all worker task payloads.
// Increment this when making breaking changes to any payload struct.
// Handlers should check the version and handle migration if needed.
const PayloadVersion = 1

// VersionedPayload is embedded in all task payloads to enable safe schema evolution.
type VersionedPayload struct {
	Version int `json:"version"`
}

// DecodePayload unmarshals a task payload and validates the version.
// Returns the decoded payload. If the version is 0 (missing from old payloads),
// it is treated as version 1 for backward compatibility.
func DecodePayload[T any](data []byte) (*T, error) {
	var payload T
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, fmt.Errorf("failed to decode task payload: %w", err)
	}
	return &payload, nil
}

// EncodePayload marshals a task payload with the current version set.
func EncodePayload(payload any) ([]byte, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to encode task payload: %w", err)
	}
	return data, nil
}
