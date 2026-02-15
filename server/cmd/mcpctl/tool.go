package main

import (
	"context"
	"fmt"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/config"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

// newToolCmd creates the tool command group.
//
// This command provides subcommands for managing and inspecting tools,
// such as calculating integrity hashes for tool definitions.
//
// Returns:
//   - *cobra.Command: The configured tool command.
func newToolCmd() *cobra.Command {
	toolCmd := &cobra.Command{
		Use:   "tool",
		Short: "Manage tools",
	}

	hashCmd := &cobra.Command{
		Use:   "hash <tool-name>",
		Short: "Calculate the integrity hash for a tool definition",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			toolName := args[0]
			osFs := afero.NewOsFs()
			cfg := config.GlobalSettings()
			if err := cfg.Load(cmd, osFs); err != nil {
				return fmt.Errorf("configuration load failed: %w", err)
			}

			store := config.NewFileStore(osFs, cfg.ConfigPaths())
			serverConfig, err := store.Load(context.Background())
			if err != nil {
				return fmt.Errorf("failed to load configuration: %w", err)
			}

			var targetTool *configv1.ToolDefinition
			for _, service := range serverConfig.GetUpstreamServices() {
				// Check type of service and find tool
				if http := service.GetHttpService(); http != nil {
					for _, t := range http.GetTools() {
						if t.GetName() == toolName {
							targetTool = t
							break
						}
					}
				} else if grpc := service.GetGrpcService(); grpc != nil {
					for _, t := range grpc.GetTools() {
						if t.GetName() == toolName {
							targetTool = t
							break
						}
					}
				} else if cli := service.GetCommandLineService(); cli != nil {
					for _, t := range cli.GetTools() {
						if t.GetName() == toolName {
							targetTool = t
							break
						}
					}
				} else if openapi := service.GetOpenapiService(); openapi != nil {
					for _, t := range openapi.GetTools() {
						if t.GetName() == toolName {
							targetTool = t
							break
						}
					}
				} else if mcp := service.GetMcpService(); mcp != nil {
					for _, t := range mcp.GetTools() {
						if t.GetName() == toolName {
							targetTool = t
							break
						}
					}
				}

				if targetTool != nil {
					break
				}
			}

			if targetTool == nil {
				return fmt.Errorf("tool %q not found in any upstream service", toolName)
			}

			hash, err := tool.CalculateConfigHash(targetTool)
			if err != nil {
				return fmt.Errorf("failed to calculate hash: %w", err)
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Tool: %s\n", toolName)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Hash: %s\n", hash)
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "\nTo apply this integrity check, add the following to your tool definition in config:")
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "integrity:")
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "  algorithm: sha256")
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  hash: %s\n", hash)

			return nil
		},
	}

	toolCmd.AddCommand(hashCmd)
	return toolCmd
}
