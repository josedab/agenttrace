package model

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/99designs/gqlgen/graphql"
	"github.com/google/uuid"
)

// UUID is a custom scalar type for UUID values
// This wrapper allows gqlgen to properly marshal/unmarshal UUID values
type UUID uuid.UUID

// MarshalGQL implements the graphql.Marshaler interface
func (u UUID) MarshalGQL(w io.Writer) {
	io.WriteString(w, `"`+uuid.UUID(u).String()+`"`)
}

// UnmarshalGQL implements the graphql.Unmarshaler interface
func (u *UUID) UnmarshalGQL(v interface{}) error {
	switch v := v.(type) {
	case string:
		parsed, err := uuid.Parse(v)
		if err != nil {
			return err
		}
		*u = UUID(parsed)
		return nil
	case []byte:
		parsed, err := uuid.Parse(string(v))
		if err != nil {
			return err
		}
		*u = UUID(parsed)
		return nil
	default:
		return fmt.Errorf("cannot unmarshal %T into UUID", v)
	}
}

// ToUUID converts the model UUID to google/uuid.UUID
func (u UUID) ToUUID() uuid.UUID {
	return uuid.UUID(u)
}

// FromUUID creates a model UUID from google/uuid.UUID
func FromUUID(id uuid.UUID) UUID {
	return UUID(id)
}

// MarshalUUID marshals a google/uuid.UUID to a GraphQL string
func MarshalUUID(u uuid.UUID) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		io.WriteString(w, `"`+u.String()+`"`)
	})
}

// UnmarshalUUID unmarshals a GraphQL string to a google/uuid.UUID
func UnmarshalUUID(v interface{}) (uuid.UUID, error) {
	switch v := v.(type) {
	case string:
		return uuid.Parse(v)
	case []byte:
		return uuid.Parse(string(v))
	default:
		return uuid.UUID{}, fmt.Errorf("cannot unmarshal %T into UUID", v)
	}
}

// JSON represents a JSON value
type JSON map[string]interface{}

// MarshalJSON implements the graphql.Marshaler interface
func MarshalJSON(val map[string]interface{}) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		if val == nil {
			w.Write([]byte("null"))
			return
		}
		data, err := json.Marshal(val)
		if err != nil {
			// Return null instead of panicking to avoid crashing the server
			w.Write([]byte("null"))
			return
		}
		w.Write(data)
	})
}

// UnmarshalJSON implements the graphql.Unmarshaler interface
func UnmarshalJSON(v interface{}) (map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	switch val := v.(type) {
	case map[string]interface{}:
		return val, nil
	case string:
		var result map[string]interface{}
		if err := json.Unmarshal([]byte(val), &result); err != nil {
			return nil, err
		}
		return result, nil
	case []byte:
		var result map[string]interface{}
		if err := json.Unmarshal(val, &result); err != nil {
			return nil, err
		}
		return result, nil
	default:
		return nil, fmt.Errorf("cannot unmarshal %T into JSON", v)
	}
}

// MarshalJSONAny marshals any value as JSON
func MarshalJSONAny(val interface{}) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		if val == nil {
			w.Write([]byte("null"))
			return
		}
		data, err := json.Marshal(val)
		if err != nil {
			// Return null instead of panicking to avoid crashing the server
			w.Write([]byte("null"))
			return
		}
		w.Write(data)
	})
}

// UnmarshalJSONAny unmarshals any JSON value
func UnmarshalJSONAny(v interface{}) (interface{}, error) {
	return v, nil
}
