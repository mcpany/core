// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"
	"strings"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/config"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage configuration",
	}

	cmd.AddCommand(newConfigDocCmd())
	return cmd
}

func newConfigDocCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "doc",
		Short: "Generate Markdown documentation from configuration",
		RunE: func(cmd *cobra.Command, _ []string) error {
			osFs := afero.NewOsFs()
			cfg := config.GlobalSettings()
			if err := cfg.Load(cmd, osFs); err != nil {
				return fmt.Errorf("configuration load failed: %w", err)
			}

			store := config.NewFileStore(osFs, cfg.ConfigPaths())
			configs, err := config.LoadServices(context.Background(), store, "server")
			if err != nil {
				return fmt.Errorf("failed to load configurations: %w", err)
			}

			var sb strings.Builder
			sb.WriteString("# MCP Server Configuration\n\n")
			sb.WriteString("## Global Settings\n\n")
			sb.WriteString(fmt.Sprintf("- **Listen Address**: `%s`\n", cfg.MCPListenAddress()))
			sb.WriteString(fmt.Sprintf("- **Log Level**: `%s`\n", cfg.LogLevel().String()))
			sb.WriteString("\n## Upstream Services\n\n")

			for _, s := range configs.GetUpstreamServices() {
				sb.WriteString(fmt.Sprintf("### %s\n\n", s.GetName()))

				switch s.WhichServiceConfig() {
				case configv1.UpstreamServiceConfig_HttpService_case:
					sb.WriteString("- **Type**: HTTP\n")
					sb.WriteString(fmt.Sprintf("- **Address**: `%s`\n", s.GetHttpService().GetAddress()))
				case configv1.UpstreamServiceConfig_GrpcService_case:
					sb.WriteString("- **Type**: gRPC\n")
					sb.WriteString(fmt.Sprintf("- **Address**: `%s`\n", s.GetGrpcService().GetAddress()))
				case configv1.UpstreamServiceConfig_FilesystemService_case:
					sb.WriteString("- **Type**: Filesystem\n")
					fs := s.GetFilesystemService()
					sb.WriteString("- **Roots**:\n")
					for v, r := range fs.GetRootPaths() {
						sb.WriteString(fmt.Sprintf("  - `%s` -> `%s`\n", v, r))
					}
					sb.WriteString("\n**Exposed Tools:**\n")
					sb.WriteString("- `list_directory`: List files and directories.\n")
					sb.WriteString("- `read_file`: Read file content.\n")
					if !fs.GetReadOnly() {
						sb.WriteString("- `write_file`: Write file content.\n")
						sb.WriteString("- `delete_file`: Delete file or directory.\n")
					}
					sb.WriteString("- `search_files`: Search for files.\n")
					sb.WriteString("- `get_file_info`: Get file metadata.\n")
				case configv1.UpstreamServiceConfig_CommandLineService_case:
					sb.WriteString("- **Type**: Command Line\n")
					sb.WriteString(fmt.Sprintf("- **Command**: `%s`\n", s.GetCommandLineService().GetCommand()))
				default:
					sb.WriteString("- **Type**: Unknown/Other\n")
				}
				sb.WriteString("\n---\n\n")
			}

			fmt.Fprintln(cmd.OutOrStdout(), sb.String())
			return nil
		},
	}
}
