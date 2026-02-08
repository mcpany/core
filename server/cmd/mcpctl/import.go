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
//
// Summary: Configuration format used by Claude Desktop.
//
// Fields:
//   - MCPServers: map[string]MCPServerConfig. A map of server names to their configurations.
type ClaudeDesktopConfig struct {
	MCPServers map[string]MCPServerConfig `json:"mcpServers"`
}

// MCPServerConfig represents a single server configuration in Claude Desktop.
//
// Summary: Configuration for a single MCP server in Claude Desktop.
//
// Fields:
//   - Command: string. The command to execute to start the server.
//   - Args: []string. The arguments to pass to the command.
//   - Env: map[string]string. Environment variables to set for the server process.
type MCPServerConfig struct {
	Command string            `json:"command"`
	Args    []string          `json:"args"`
	Env     map[string]string `json:"env,omitempty"`
}

// McpAnyConfig represents the target configuration structure for MCP Any.
//
// Summary:
//   Configuration for the MCP Any server.
//
// Fields:
//   - UpstreamServices: []UpstreamService. A list of upstream services to configure.
type McpAnyConfig struct {
	UpstreamServices []UpstreamService `yaml:"upstream_services"`
}

// UpstreamService represents a single upstream service configuration.
//
// Summary:
//   Configuration for a single upstream service.
//
// Fields:
//   - Name: string. The name of the service.
//   - McpService: *McpService. The MCP service configuration (optional).
type UpstreamService struct {
	Name       string      `yaml:"name"`
	McpService *McpService `yaml:"mcp_service,omitempty"`
}

// McpService defines the configuration for an MCP-based service.
//
// Summary:
//   Configuration for a service using the Model Context Protocol (MCP).
//
// Fields:
//   - StdioConnection: *StdioConnection. Parameters for connecting via standard I/O (optional).
type McpService struct {
	StdioConnection *StdioConnection `yaml:"stdio_connection,omitempty"`
}

// StdioConnection defines the parameters for connecting to an MCP server via standard I/O.
//
// Summary:
//   Parameters for connecting to an MCP server using standard input/output streams.
//
// Fields:
//   - Command: string. The command to execute.
//   - Args: []string. The arguments to pass to the command.
//   - Env: map[string]string. The environment variables to set for the command.
type StdioConnection struct {
	Command string            `yaml:"command"`
	Args    []string          `yaml:"args"`
	Env     map[string]string `yaml:"env,omitempty"`
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
