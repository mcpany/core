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
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestYamlEngine_Unmarshal_Error(t *testing.T) {
	fs := afero.NewMemMapFs()
	// Create invalid YAML file (tab indentation is invalid in YAML)
	require.NoError(t, afero.WriteFile(fs, "/config/invalid.yaml", []byte("\tinvalid: yaml"), 0644))

	store := NewFileStore(fs, []string{"/config/invalid.yaml"})
	_, err := store.Load()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to unmarshal YAML")
}

func TestUnmarshal_NaN(t *testing.T) {
	engine := &yamlEngine{}
	// YAML supports NaN, but JSON does not.
	// This should fail at json.Marshal step.
	yamlData := []byte("global_settings:\n  log_level: .nan")

	// Wait, log_level is STRING in proto.
	// yaml.Unmarshal might decode .nan as string ".nan" if field is string?
	// But yaml.Unmarshal unmarshals into INTERFACE map first.
	// So it detects type. .nan is float in YAML.
	// So map["global_settings"]["log_level"] = NaN (float64).
	// json.Marshal will fail.

	cfg := &configv1.McpAnyServerConfig{}
	err := engine.Unmarshal(yamlData, cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to marshal map to JSON")
}
