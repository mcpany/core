// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/santhosh-tekuri/jsonschema/v5"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"
)

// SchemaValidator wraps a compiled JSON schema for validation.
type SchemaValidator struct {
	schema *jsonschema.Schema
}

// NewSchemaValidator creates a new SchemaValidator from a protobuf Struct.
// It compiles the schema for efficient reuse.
func NewSchemaValidator(inputSchema *structpb.Struct) (*SchemaValidator, error) {
	if inputSchema == nil {
		return nil, nil
	}

	// Convert protobuf Struct to JSON bytes
	schemaBytes, err := protojson.Marshal(inputSchema)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal input schema: %w", err)
	}

	// Sanitize/Fix common schema issues if needed.
	// Since we are using jsonschema compiler, it might be strict.
	// We might need to ensure "type": "object" is present if properties are present.
	// But `SanitizeJSONSchema` in `schema_sanitizer.go` does that.
	// Ideally the input schema is already sanitized when the Tool object is created.
	// But let's assume raw schema for now.

	compiler := jsonschema.NewCompiler()
	compiler.Draft = jsonschema.Draft2020 // Use latest draft support
	// We use "schema.json" as a virtual filename
	if err := compiler.AddResource("schema.json", bytes.NewReader(schemaBytes)); err != nil {
		return nil, fmt.Errorf("failed to add schema resource: %w", err)
	}

	schema, err := compiler.Compile("schema.json")
	if err != nil {
		return nil, fmt.Errorf("failed to compile schema: %w", err)
	}

	return &SchemaValidator{schema: schema}, nil
}

// Validate validates the input JSON bytes against the compiled schema.
func (v *SchemaValidator) Validate(inputs []byte) error {
	if v == nil || v.schema == nil {
		return nil // No validation needed
	}

	// Handle empty input
	if len(inputs) == 0 {
		inputs = []byte("{}")
	}

	// Unmarshal input to interface{} for validation
	var inputMap interface{}
	if err := json.Unmarshal(inputs, &inputMap); err != nil {
		return fmt.Errorf("failed to unmarshal inputs for validation: %w", err)
	}

	if err := v.schema.Validate(inputMap); err != nil {
		return fmt.Errorf("input validation failed: %w", err)
	}

	return nil
}
