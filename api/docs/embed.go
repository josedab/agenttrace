// Package docs provides embedded API documentation files.
package docs

import (
	_ "embed"
)

// OpenAPISpec contains the embedded OpenAPI specification in YAML format.
//
//go:embed openapi.yaml
var OpenAPISpec []byte
