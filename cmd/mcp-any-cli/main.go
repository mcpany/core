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
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/mcpany/core/pkg/config"
	"github.com/spf13/afero"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "generate":
		if err := generate(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "validate":
		validateCmd := flag.NewFlagSet("validate", flag.ExitOnError)
		configPaths := validateCmd.String("config-paths", "", "Paths to configuration files or directories")

		if err := validateCmd.Parse(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing flags: %v\n", err)
			os.Exit(1)
		}

		if *configPaths == "" {
			fmt.Fprintln(os.Stderr, "Error: --config-paths is required for validate command")
			printUsage()
			os.Exit(1)
		}

		paths := strings.Split(*configPaths, ",")
		store := config.NewFileStore(afero.NewOsFs(), paths)
		if _, err := config.LoadServices(store, "server"); err != nil {
			fmt.Fprintf(os.Stderr, "Configuration validation failed: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Configuration is valid.")
	default:
		printUsage()
		os.Exit(1)
	}
}

func generate() error {
	fmt.Println("MCP Any CLI: Configuration Generator")

	generator := config.NewGenerator()
	configData, err := generator.Generate()
	if err != nil {
		return err
	}

	fmt.Println("\nGenerated configuration:")
	fmt.Print(string(configData))

	return nil
}

func printUsage() {
	fmt.Println("Usage: mcp-any-cli <command> [flags]")
	fmt.Println("\nCommands:")
	fmt.Println("  generate   Generate a new configuration file interactively")
	fmt.Println("  validate   Validate one or more configuration files")
	fmt.Println("\nFlags for validate:")
	fmt.Println("  --config-paths string   Paths to configuration files or directories")
}
