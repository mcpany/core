// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"encoding/json"
	"fmt"

	configpb "github.com/mcpany/core/proto/config/v1"
	"google.golang.org/protobuf/encoding/protojson"
	"gopkg.in/yaml.v3"
)

// ClaudeConfig represents the structure of claude_desktop_config.json.
type ClaudeConfig struct {
	MCPServers map[string]ClaudeServerConfig `json:"mcpServers"`
}

// ClaudeServerConfig represents a single server entry in Claude's config.
type ClaudeServerConfig struct {
	Command string            `json:"command"`
	Args    []string          `json:"args"`
	Env     map[string]string `json:"env"`
}

// MigrateClaudeConfig converts a Claude Desktop configuration JSON byte slice
// into an MCP Any configuration YAML string.
func MigrateClaudeConfig(input []byte) (string, error) {
	var claudeCfg ClaudeConfig
	if err := json.Unmarshal(input, &claudeCfg); err != nil {
		return "", fmt.Errorf("failed to parse Claude config: %w", err)
	}

	mcpAnyConfig := &configpb.McpAnyServerConfig{
		UpstreamServices: []*configpb.UpstreamServiceConfig{},
	}

	for name, server := range claudeCfg.MCPServers {
		envMap := make(map[string]*configpb.SecretValue)
		for k, v := range server.Env {
			envMap[k] = &configpb.SecretValue{
				Value: &configpb.SecretValue_PlainText{
					PlainText: v,
				},
			}
		}

		mcpService := &configpb.McpUpstreamService{
			ConnectionType: &configpb.McpUpstreamService_StdioConnection{
				StdioConnection: &configpb.McpStdioConnection{
					Command: &server.Command,
					Args:    server.Args,
					Env:     envMap,
				},
			},
		}

		service := &configpb.UpstreamServiceConfig{
			Name: &name,
			ServiceConfig: &configpb.UpstreamServiceConfig_McpService{
				McpService: mcpService,
			},
		}

		mcpAnyConfig.UpstreamServices = append(mcpAnyConfig.UpstreamServices, service)
	}

	// 1. Marshal to JSON first to handle Protobuf-specifics nicely
	marshaler := protojson.MarshalOptions{
		Multiline:     false, // compact for intermediate
		UseProtoNames: true,  // Use snake_case as defined in .proto
	}
	jsonBytes, err := marshaler.Marshal(mcpAnyConfig)
	if err != nil {
		return "", fmt.Errorf("failed to marshal MCP Any config to intermediate JSON: %w", err)
	}

	// 2. Unmarshal into a generic map to prepare for YAML
	var genericMap map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &genericMap); err != nil {
		return "", fmt.Errorf("failed to unmarshal intermediate JSON: %w", err)
	}

	// 3. Marshal to YAML
	yamlBytes, err := yaml.Marshal(genericMap)
	if err != nil {
		return "", fmt.Errorf("failed to marshal to YAML: %w", err)
	}

	return string(yamlBytes), nil
}
