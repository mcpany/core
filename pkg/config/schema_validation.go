/*
 * Copyright 2025 Author(s) of MCP Any
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package config

import (
	"fmt"
	"sync"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/santhosh-tekuri/jsonschema/v5"
)

var (
	schemaOnce     sync.Once
	compiledSchema *jsonschema.Schema
	schemaGenErr   error
)

// ensureSchema generates and compiles the JSON schema for McpAnyServerConfig.
// It does this only once.
func ensureSchema() (*jsonschema.Schema, error) {
	schemaOnce.Do(func() {
		// 1. Generate JSON Schema from the Proto definition
		cfg := &configv1.McpAnyServerConfig{}
		var err error
		compiledSchema, err = GenerateSchemaFromProto(cfg.ProtoReflect())
		if err != nil {
			schemaGenErr = fmt.Errorf("failed to generate schema from proto: %w", err)
			return
		}
	})
	return compiledSchema, schemaGenErr
}

// ValidateConfigAgainstSchema validates the raw configuration map against the generated JSON schema.
func ValidateConfigAgainstSchema(rawConfig map[string]interface{}) error {
	schema, err := ensureSchema()
	if err != nil {
		return fmt.Errorf("schema generation failed: %w", err)
	}

	if err := schema.Validate(rawConfig); err != nil {
		return fmt.Errorf("schema validation failed: %w", err)
	}
	return nil
}
