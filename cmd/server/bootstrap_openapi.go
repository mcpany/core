// Copyright 2025 Author(s) of MCP Any
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
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/mcpany/core/pkg/generator"
	"github.com/spf13/cobra"
)

func newBootstrapOpenAPICmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "openapi [flags] <spec_url_or_path>",
		Short: "Bootstrap a new MCP Any configuration from an OpenAPI specification.",
		Long:  `Bootstrap a new MCP Any configuration from an OpenAPI specification.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			specPath := args[0]
			specData, err := readSpec(specPath)
			if err != nil {
				return fmt.Errorf("failed to read spec: %w", err)
			}

			output, err := generator.GenerateMCPConfigFromOpenAPI(specData, specPath)
			if err != nil {
				return fmt.Errorf("failed to generate config: %w", err)
			}

			fmt.Fprint(os.Stdout, string(output))
			return nil
		},
	}
	return cmd
}

func readSpec(path string) ([]byte, error) {
	// First, try to read as a local file
	if _, err := os.Stat(path); err == nil {
		return os.ReadFile(path)
	} else if !os.IsNotExist(err) {
		// If there's an error other than "not found," return it
		return nil, err
	}

	// If it's not a local file, try to fetch it as a URL
	resp, err := http.Get(path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch spec from URL: %s", resp.Status)
	}

	return io.ReadAll(resp.Body)
}
