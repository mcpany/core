// Copyright 2024 Author(s) of MCP Any
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateCommand(t *testing.T) {
	// Build the CLI tool
	cmd := exec.Command("go", "build", "-o", "mcp-any-cli-test")
	err := cmd.Run()
	assert.NoError(t, err)
	defer os.Remove("mcp-any-cli-test")

	t.Run("valid config", func(t *testing.T) {
		// Create a temporary valid config file
		dir := t.TempDir()
		validConfig := `
upstreamServices:
  - name: "my-http-service"
    httpService:
      address: "https://api.example.com"
`
		configPath := filepath.Join(dir, "valid.yaml")
		err := os.WriteFile(configPath, []byte(validConfig), 0644)
		assert.NoError(t, err)

		// Run the validate command
		cmd := exec.Command("./mcp-any-cli-test", "validate", "--config-paths", configPath)
		output, err := cmd.CombinedOutput()
		assert.NoError(t, err, "validate command should succeed for valid config: %s", string(output))
		assert.Contains(t, string(output), "Configuration is valid.")
	})

	t.Run("invalid config", func(t *testing.T) {
		// Create a temporary invalid config file
		dir := t.TempDir()
		invalidConfig := `
upstreamServices:
  - name: "my-http-service"
    httpService:
      address: "" # Invalid address
`
		configPath := filepath.Join(dir, "invalid.yaml")
		err := os.WriteFile(configPath, []byte(invalidConfig), 0644)
		assert.NoError(t, err)

		// Run the validate command
		cmd := exec.Command("./mcp-any-cli-test", "validate", "--config-paths", configPath)
		output, err := cmd.CombinedOutput()
		assert.Error(t, err, "validate command should fail for invalid config")
		assert.Contains(t, string(output), "Configuration validation failed")
	})

	t.Run("non-existent config", func(t *testing.T) {
		// Run the validate command with a non-existent file
		cmd := exec.Command("./mcp-any-cli-test", "validate", "--config-paths", "non-existent-file.yaml")
		output, err := cmd.CombinedOutput()
		assert.Error(t, err, "validate command should fail for non-existent config")
		assert.Contains(t, string(output), "Configuration validation failed")
	})
}
