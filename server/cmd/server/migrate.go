// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"encoding/json"
	"fmt"
	"os"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type ClaudeConfig struct {
	MCPServers map[string]ClaudeServerConfig `json:"mcpServers"`
}

type ClaudeServerConfig struct {
	Command string            `json:"command"`
	Args    []string          `json:"args"`
	Env     map[string]string `json:"env"`
}

func newMigrateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate [claude_config.json]",
		Short: "Migrate Claude Desktop configuration to MCP Any configuration",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			data, err := os.ReadFile(args[0])
			if err != nil {
				return fmt.Errorf("failed to read file: %w", err)
			}

			var claudeCfg ClaudeConfig
			if err := json.Unmarshal(data, &claudeCfg); err != nil {
				return fmt.Errorf("failed to parse Claude config: %w", err)
			}

			mcpAnyCfg := &configv1.McpAnyServerConfig{}

			for name, srv := range claudeCfg.MCPServers {
				svc := &configv1.UpstreamServiceConfig{
					Name: proto.String(name),
					ServiceConfig: &configv1.UpstreamServiceConfig_McpService{
						McpService: &configv1.McpUpstreamService{
							ConnectionType: &configv1.McpUpstreamService_StdioConnection{
								StdioConnection: &configv1.McpStdioConnection{
									Command: proto.String(srv.Command),
									Args:    srv.Args,
								},
							},
						},
					},
				}
				// Map Env
				if len(srv.Env) > 0 {
					secrets := make(map[string]*configv1.SecretValue)
					for k, v := range srv.Env {
						secrets[k] = &configv1.SecretValue{
							Value: &configv1.SecretValue_PlainText{
								PlainText: v,
							},
						}
					}
					svc.GetMcpService().GetStdioConnection().Env = secrets
				}

				mcpAnyCfg.UpstreamServices = append(mcpAnyCfg.UpstreamServices, svc)
			}

			marshaler := protojson.MarshalOptions{Multiline: true, Indent: "  "}
			out, err := marshaler.Marshal(mcpAnyCfg)
			if err != nil {
				return err
			}
			fmt.Println(string(out))
			return nil
		},
	}
	return cmd
}
