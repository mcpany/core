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

// ClaudeDesktopConfig claudeDesktopConfig represents a claude desktop config.
//
// Summary: ClaudeDesktopConfig represents a claude desktop config.
type ClaudeDesktopConfig struct {
	MCPServers map[string]MCPServerConfig `json:"mcpServers"`
}

// MCPServerConfig mCPServerConfig represents a mcp server config.
//
// Summary: MCPServerConfig represents a mcp server config.
type MCPServerConfig struct {
	Command string            `json:"command"`
	Args    []string          `json:"args"`
	Env     map[string]string `json:"env,omitempty"`
}

// McpAnyConfig mcpAnyConfig represents a mcp any config.
//
// Summary: McpAnyConfig represents a mcp any config.
type McpAnyConfig struct {
	UpstreamServices []UpstreamService `yaml:"upstream_services"`
}

// UpstreamService upstreamService represents a upstream service.
//
// Summary: UpstreamService represents a upstream service.
type UpstreamService struct {
	Name       string      `yaml:"name"`
	McpService *McpService `yaml:"mcp_service,omitempty"`
}

// McpService mcpService represents a mcp service.
//
// Summary: McpService represents a mcp service.
type McpService struct {
	StdioConnection *StdioConnection `yaml:"stdio_connection,omitempty"`
}

// StdioConnection stdioConnection represents a stdio connection.
//
// Summary: StdioConnection represents a stdio connection.
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
		RunE: func(cmd *cobra.Command, args []string) error {
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
				fmt.Fprintf(cmd.OutOrStdout(), "Successfully imported configuration to %s\n", outputPath)
			} else {
				fmt.Fprintln(cmd.OutOrStdout(), string(yamlData))
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&outputPath, "output", "o", "", "Path to write the output YAML file (default: stdout)")

	return cmd
}
