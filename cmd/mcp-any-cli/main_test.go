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
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCLI(t *testing.T) {
	t.Run("no command", func(t *testing.T) {
		// Capture stdout
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		// Run the CLI with no arguments
		rootCmd := &cobra.Command{
			Use:   "mcp-any-cli",
			Short: "A CLI tool for MCP Any",
		}
		rootCmd.AddCommand(NewGenerateCommand())
		rootCmd.AddCommand(NewScaffoldCommand())
		err := rootCmd.Execute()
		require.NoError(t, err)

		// Restore stdout and read the output
		w.Close()
		os.Stdout = old
		var buf bytes.Buffer
		_, err = io.Copy(&buf, r)
		require.NoError(t, err)

		// Check the output
		assert.Contains(t, buf.String(), "A CLI tool for MCP Any", "unexpected output format")
	})

	t.Run("scaffold command", func(t *testing.T) {
		// Create a temporary OpenAPI spec file
		openapiFile, err := os.CreateTemp("", "openapi-*.yaml")
		require.NoError(t, err)
		defer os.Remove(openapiFile.Name())

		_, err = openapiFile.WriteString(`
openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
servers:
  - url: http://localhost:8080
paths: {}
`)
		require.NoError(t, err)
		openapiFile.Close()

		// Create a temporary output file
		outputFile, err := os.CreateTemp("", "config-*.yaml")
		require.NoError(t, err)
		defer os.Remove(outputFile.Name())
		outputFile.Close()

		// Run the scaffold command
		cmd := NewScaffoldCommand()
		cmd.SetArgs([]string{"-f", openapiFile.Name(), "-o", outputFile.Name()})
		err = cmd.Execute()
		require.NoError(t, err)

		// Check that the output file was created
		_, err = os.Stat(outputFile.Name())
		require.NoError(t, err)
	})
}
