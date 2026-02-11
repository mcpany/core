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
type ClaudeDesktopConfig struct {
	// MCPServers is a map of server names to their configurations.
	MCPServers map[string]MCPServerConfig `json:"mcpServers"`
}

// MCPServerConfig represents a single server configuration in Claude Desktop.
//
// Summary: Configuration for a single MCP server in Claude Desktop.
type MCPServerConfig struct {
	// Command is the command to execute to start the server.
	Command string `json:"command"`
	// Args is a list of arguments to pass to the command.
	Args []string `json:"args"`
	// Env is a map of environment variables to set for the server process.
	Env map[string]string `json:"env,omitempty"`
}

// McpAnyConfig represents the target configuration structure for MCP Any.
//
// Summary:
//   Configuration for the MCP Any server.
type McpAnyConfig struct {
	// UpstreamServices is a list of upstream services to configure.
	UpstreamServices []UpstreamService `yaml:"upstream_services"`
}

// UpstreamService represents a single upstream service configuration.
//
// Summary:
//   Configuration for a single upstream service.
type UpstreamService struct {
	// Name is the name of the service.
	Name string `yaml:"name"`
	// McpService contains the MCP service configuration (optional).
	McpService *McpService `yaml:"mcp_service,omitempty"`
}

// McpService defines the configuration for an MCP-based service.
//
// Summary:
//   Configuration for a service using the Model Context Protocol (MCP).
type McpService struct {
	// StdioConnection contains parameters for connecting via standard I/O (optional).
	StdioConnection *StdioConnection `yaml:"stdio_connection,omitempty"`
}

// StdioConnection defines the parameters for connecting to an MCP server via standard I/O.
//
// Summary:
//   Parameters for connecting to an MCP server using standard input/output streams.
type StdioConnection struct {
	// Command is the command to execute.
	Command string `yaml:"command"`
	// Args is a list of arguments to pass to the command.
	Args []string `yaml:"args"`
	// Env is a map of environment variables to set for the command.
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
