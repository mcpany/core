// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// ClaudeDesktopConfig represents the structure of claude_desktop_config.json.
type ClaudeDesktopConfig struct {
	// MCPServers is a map of server names to their configurations.
	MCPServers map[string]MCPServerConfig `json:"mcpServers"`
}

// MCPServerConfig represents a single server configuration in Claude Desktop.
type MCPServerConfig struct {
	// Command is the executable command to run the server.
	Command string `json:"command"`
	// Args is the list of arguments to pass to the command.
	Args []string `json:"args"`
	// Env is a map of environment variables to set for the server process.
	Env map[string]string `json:"env,omitempty"`
}

// McpAnyConfig represents the target configuration structure for MCP Any.
type McpAnyConfig struct {
	// UpstreamServices is a list of upstream services to configure in MCP Any.
	UpstreamServices []UpstreamService `yaml:"upstream_services"`
}

// UpstreamService represents a single upstream service configuration.
type UpstreamService struct {
	// Name is the unique identifier for the service.
	Name string `yaml:"name"`
	// McpService holds the configuration for an MCP-compatible service.
	McpService *McpService `yaml:"mcp_service,omitempty"`
}

// McpService represents the configuration for an MCP service.
type McpService struct {
	// StdioConnection configures the connection to the service via standard I/O.
	StdioConnection *StdioConnection `yaml:"stdio_connection,omitempty"`
}

// StdioConnection represents the configuration for a stdio-based connection.
type StdioConnection struct {
	// Command is the executable command to run.
	Command string `yaml:"command"`
	// Args is the list of arguments to pass to the command.
	Args []string `yaml:"args"`
	// Env is a map of environment variables to set for the process.
	Env map[string]string `yaml:"env,omitempty"`
}

func newImportCmd() *cobra.Command {
	var outputPath string

	cmd := &cobra.Command{
		Use:   "import [path to claude_desktop_config.json]",
		Short: "Import configuration from Claude Desktop",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			inputPath := filepath.Clean(args[0])

			// Read input file
			data, err := os.ReadFile(inputPath)
			if err != nil {
				return fmt.Errorf("failed to read input file: %w", err)
			}

			// Parse JSON
			var claudeConfig ClaudeDesktopConfig
			if err := json.Unmarshal(data, &claudeConfig); err != nil {
				return fmt.Errorf("failed to parse Claude Desktop config: %w", err)
			}

			// Convert to McpAnyConfig
			mcpAnyConfig := McpAnyConfig{
				UpstreamServices: make([]UpstreamService, 0, len(claudeConfig.MCPServers)),
			}

			for name, server := range claudeConfig.MCPServers {
				service := UpstreamService{
					Name: name,
					McpService: &McpService{
						StdioConnection: &StdioConnection{
							Command: server.Command,
							Args:    server.Args,
							Env:     server.Env,
						},
					},
				}
				mcpAnyConfig.UpstreamServices = append(mcpAnyConfig.UpstreamServices, service)
			}

			// Marshal to YAML
			yamlData, err := yaml.Marshal(&mcpAnyConfig)
			if err != nil {
				return fmt.Errorf("failed to marshal to YAML: %w", err)
			}

			// Output
			if outputPath != "" {
				if err := os.WriteFile(outputPath, yamlData, 0600); err != nil {
					return fmt.Errorf("failed to write output file: %w", err)
				}
				fmt.Printf("Successfully imported configuration to %s\n", outputPath)
			} else {
				fmt.Println(string(yamlData))
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&outputPath, "output", "o", "", "Path to write the output YAML file (default: stdout)")

	return cmd
}
