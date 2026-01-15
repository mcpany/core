// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"encoding/json"
	"fmt"
	"sort"

	"sigs.k8s.io/yaml"
)

// ClaudeDesktopConfig represents the structure of the claude_desktop_config.json file.
type ClaudeDesktopConfig struct {
	MCPServers map[string]ClaudeMCPServer `json:"mcpServers"`
}

// ClaudeMCPServer represents a single server configuration in Claude Desktop.
type ClaudeMCPServer struct {
	Command string            `json:"command"`
	Args    []string          `json:"args"`
	Env     map[string]string `json:"env,omitempty"`
}

// MigrateClaudeConfig converts a Claude Desktop configuration JSON to an MCP Any configuration YAML.
func MigrateClaudeConfig(input []byte) ([]byte, error) {
	var claudeConfig ClaudeDesktopConfig
	if err := json.Unmarshal(input, &claudeConfig); err != nil {
		return nil, fmt.Errorf("failed to parse Claude Desktop config: %w", err)
	}

	// MCP Any configuration structure (simplified for YAML generation)
	type StdioConnection struct {
		Command          string            `json:"command"`
		Args             []string          `json:"args"`
		Env              map[string]string `json:"env,omitempty"`
		WorkingDirectory string            `json:"working_directory,omitempty"`
	}

	type MCPService struct {
		StdioConnection StdioConnection `json:"stdioConnection"`
	}

	type UpstreamService struct {
		Name       string     `json:"name"`
		MCPService MCPService `json:"mcpService"`
	}

	type MCPAnyConfig struct {
		UpstreamServices []UpstreamService `json:"upstreamServices"`
	}

	mcpAnyConfig := MCPAnyConfig{
		UpstreamServices: []UpstreamService{},
	}

	// Sort keys for deterministic output
	keys := make([]string, 0, len(claudeConfig.MCPServers))
	for k := range claudeConfig.MCPServers {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, name := range keys {
		server := claudeConfig.MCPServers[name]
		upstream := UpstreamService{
			Name: name,
			MCPService: MCPService{
				StdioConnection: StdioConnection{
					Command: server.Command,
					Args:    server.Args,
					Env:     server.Env,
				},
			},
		}
		mcpAnyConfig.UpstreamServices = append(mcpAnyConfig.UpstreamServices, upstream)
	}

	// Marshal to YAML
	yamlBytes, err := yaml.Marshal(mcpAnyConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to generate MCP Any config: %w", err)
	}

	return yamlBytes, nil
}
