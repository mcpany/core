// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestImportCmd(t *testing.T) {
	tempDir := t.TempDir()

	// 1. Valid Input -> Stdout
	t.Run("Valid Input Stdout", func(t *testing.T) {
		inputFile := filepath.Join(tempDir, "valid_claude_config.json")
		err := os.WriteFile(inputFile, []byte(`{
			"mcpServers": {
				"test-server": {
					"command": "node",
					"args": ["server.js"],
					"env": {"KEY": "VALUE"}
				}
			}
		}`), 0600)
		require.NoError(t, err)

		cmd := newRootCmd()
		b := bytes.NewBufferString("")
		cmd.SetOut(b)
		cmd.SetArgs([]string{"import", inputFile})

		err = cmd.Execute()
		assert.NoError(t, err)

		output := b.String()
		assert.Contains(t, output, "upstream_services:")
		assert.Contains(t, output, "- name: test-server")
		assert.Contains(t, output, "command: node")
		assert.Contains(t, output, "args:")
		assert.Contains(t, output, "- server.js")
		assert.Contains(t, output, "KEY: VALUE")
	})

	// 2. Valid Input -> File Output
	t.Run("Valid Input File Output", func(t *testing.T) {
		inputFile := filepath.Join(tempDir, "valid_claude_config_2.json")
		outputFile := filepath.Join(tempDir, "output.yaml")
		err := os.WriteFile(inputFile, []byte(`{
			"mcpServers": {
				"server-2": {
					"command": "python",
					"args": ["script.py"]
				}
			}
		}`), 0600)
		require.NoError(t, err)

		cmd := newRootCmd()
		b := bytes.NewBufferString("")
		cmd.SetOut(b)
		cmd.SetArgs([]string{"import", inputFile, "--output", outputFile})

		err = cmd.Execute()
		assert.NoError(t, err)

		assert.Contains(t, b.String(), "Successfully imported configuration")

		// Verify output file content
		content, err := os.ReadFile(outputFile)
		assert.NoError(t, err)
		output := string(content)
		assert.Contains(t, output, "name: server-2")
		assert.Contains(t, output, "command: python")
	})

	// 3. File Not Found
	t.Run("File Not Found", func(t *testing.T) {
		cmd := newRootCmd()
		b := bytes.NewBufferString("")
		cmd.SetOut(b)
		cmd.SetArgs([]string{"import", "non_existent_file.json"})

		err := cmd.Execute()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to read input file")
	})

	// 4. Invalid JSON
	t.Run("Invalid JSON", func(t *testing.T) {
		inputFile := filepath.Join(tempDir, "invalid.json")
		err := os.WriteFile(inputFile, []byte(`{ "mcpServers": { ... invalid ... } }`), 0600)
		require.NoError(t, err)

		cmd := newRootCmd()
		b := bytes.NewBufferString("")
		cmd.SetOut(b)
		cmd.SetArgs([]string{"import", inputFile})

		err = cmd.Execute()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse Claude Desktop config")
	})

	// 5. Empty Config
	t.Run("Empty Config", func(t *testing.T) {
		inputFile := filepath.Join(tempDir, "empty.json")
		err := os.WriteFile(inputFile, []byte(`{ "mcpServers": {} }`), 0600)
		require.NoError(t, err)

		cmd := newRootCmd()
		b := bytes.NewBufferString("")
		cmd.SetOut(b)
		cmd.SetArgs([]string{"import", inputFile})

		err = cmd.Execute()
		assert.NoError(t, err)
		output := b.String()
		assert.Contains(t, output, "upstream_services: []")
	})

	// 6. Complex Config (Special chars, multiple args)
	t.Run("Complex Config", func(t *testing.T) {
		inputFile := filepath.Join(tempDir, "complex.json")
		err := os.WriteFile(inputFile, []byte(`{
			"mcpServers": {
				"complex-server": {
					"command": "/usr/bin/env",
					"args": ["bash", "-c", "echo 'hello world'"],
					"env": {"PATH": "/usr/bin:/bin"}
				}
			}
		}`), 0600)
		require.NoError(t, err)

		cmd := newRootCmd()
		b := bytes.NewBufferString("")
		cmd.SetOut(b)
		cmd.SetArgs([]string{"import", inputFile})

		err = cmd.Execute()
		assert.NoError(t, err)

		output := b.String()
		assert.Contains(t, output, "name: complex-server")
		assert.Contains(t, output, "command: /usr/bin/env")
		assert.Contains(t, output, "- bash")
		assert.Contains(t, output, "- -c")
		assert.Contains(t, output, "- echo 'hello world'")
		assert.Contains(t, output, "PATH: /usr/bin:/bin")
	})
}
