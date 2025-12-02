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
	"fmt"
	"os"

	"github.com/mcpany/core/pkg/config"
	"github.com/mcpany/core/pkg/config/scaffold"
	"github.com/spf13/cobra"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	rootCmd := &cobra.Command{
		Use:   "mcp-any-cli",
		Short: "A CLI tool for MCP Any",
	}

	rootCmd.AddCommand(NewGenerateCommand())
	rootCmd.AddCommand(NewScaffoldCommand())

	return rootCmd.Execute()
}

// NewGenerateCommand creates a new cobra command for generating a configuration.
func NewGenerateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate a new MCP Any configuration interactively",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintf(os.Stderr, "MCP Any CLI: Configuration Generator\n")

			generator := config.NewGenerator()
			configData, err := generator.Generate()
			if err != nil {
				return err
			}

			fmt.Println("\nGenerated configuration:")
			fmt.Print(string(configData))
			return nil
		},
	}
	return cmd
}

// NewScaffoldCommand creates a new cobra command for scaffolding a configuration from an OpenAPI spec.
func NewScaffoldCommand() *cobra.Command {
	var (
		openapiFile string
		outputFile  string
	)

	cmd := &cobra.Command{
		Use:   "scaffold",
		Short: "Scaffold a new MCP Any configuration from an OpenAPI spec",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintf(os.Stderr, "Scaffolding a new configuration from %s to %s...\n", openapiFile, outputFile)
			return scaffold.ScaffoldFile(openapiFile, outputFile)
		},
	}

	cmd.Flags().StringVarP(&openapiFile, "openapi-file", "f", "", "Path to the OpenAPI specification file")
	cmd.Flags().StringVarP(&outputFile, "output-file", "o", "config.yaml", "Path to the output configuration file")
	_ = cmd.MarkFlagRequired("openapi-file")

	return cmd
}
