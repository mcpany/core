// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package check provides functionality for validating configuration files
// against the official JSON schema.
package check

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/mcpany/core/server/pkg/schema"
	"github.com/santhosh-tekuri/jsonschema/v5"
	"sigs.k8s.io/yaml"
	yamlv3 "gopkg.in/yaml.v3"
)

// Result represents a validation error with location.
type Result struct {
	Message string
	Path    string
	JSONPointer string
	Line    int
	Column  int
}

func (r Result) String() string {
	return fmt.Sprintf("%s:%d:%d: %s", r.Path, r.Line, r.Column, r.Message)
}

// ValidateFile validates the configuration file at the given path.
func ValidateFile(_ context.Context, path string) ([]Result, error) {
	data, err := os.ReadFile(path) //nolint:gosec // intended file reading
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// 1. Convert YAML to JSON for validation
	jsonData, err := yaml.YAMLToJSON(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	// 2. Load Schema
	compiler := jsonschema.NewCompiler()
	if err := compiler.AddResource("config.json", strings.NewReader(string(schema.ConfigSchema))); err != nil {
		return nil, fmt.Errorf("failed to load schema: %w", err)
	}
	sch, err := compiler.Compile("config.json")
	if err != nil {
		return nil, fmt.Errorf("failed to compile schema: %w", err)
	}

	// 3. Unmarshal JSON to interface{} for validation
	var v interface{}
	if err := json.Unmarshal(jsonData, &v); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	// 4. Validate
	if err := sch.Validate(v); err != nil {
		var results []Result
		// Parse YAML with yaml.v3 to find line numbers
		var node yamlv3.Node
		if yamlErr := yamlv3.Unmarshal(data, &node); yamlErr != nil {
			// Should not happen if YAMLToJSON succeeded, but just in case
			return nil, fmt.Errorf("failed to parse YAML for line numbers: %w", yamlErr)
		}

		if valErr, ok := err.(*jsonschema.ValidationError); ok {
			// jsonschema returns a tree of errors. We need to flatten it.
			flattened := flattenErrors(valErr)
			for _, e := range flattened {
				loc := e.InstanceLocation
				// If error is about additional properties, point to the property itself
				if strings.Contains(e.Message, "additionalProperties") {
					if start := strings.Index(e.Message, "'"); start != -1 {
						if end := strings.Index(e.Message[start+1:], "'"); end != -1 {
							propName := e.Message[start+1 : start+1+end]
							loc = fmt.Sprintf("%s/%s", loc, propName)
						}
					}
				}

				line, col := findLocation(&node, loc)
				results = append(results, Result{
					Message:     e.Message,
					Path:        path,
					JSONPointer: e.InstanceLocation,
					Line:        line,
					Column:      col,
				})
			}
			return results, nil
		}
		return nil, err
	}

	return nil, nil
}

func flattenErrors(err *jsonschema.ValidationError) []*jsonschema.ValidationError {
	var results []*jsonschema.ValidationError
	if len(err.Causes) == 0 {
		return []*jsonschema.ValidationError{err}
	}
	for _, cause := range err.Causes {
		results = append(results, flattenErrors(cause)...)
	}
	return results
}

func findLocation(node *yamlv3.Node, path string) (int, int) {
	// path is like /upstream_services/0/name
	parts := strings.Split(strings.TrimPrefix(path, "/"), "/")
	if len(parts) == 1 && parts[0] == "" {
		return node.Line, node.Column
	}

	if len(node.Content) == 0 {
		return node.Line, node.Column
	}

	current := node.Content[0] // Root document node usually has one child

	for _, part := range parts {
		if current.Kind == yamlv3.MappingNode {
			found := false
			for i := 0; i < len(current.Content); i += 2 {
				key := current.Content[i]
				if key.Value == part {
					current = current.Content[i+1]
					found = true
					break
				}
			}
			if !found {
				return current.Line, current.Column // Best effort
			}
		} else if current.Kind == yamlv3.SequenceNode {
			idx, err := strconv.Atoi(part)
			if err != nil || idx >= len(current.Content) {
				return current.Line, current.Column
			}
			current = current.Content[idx]
		}
	}

	return current.Line, current.Column
}
