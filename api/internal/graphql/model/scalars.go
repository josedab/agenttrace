package model

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/99designs/gqlgen/graphql"
)

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
