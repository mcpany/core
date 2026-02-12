// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"

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
		Short: "Generate documentation from configuration",
		Long: `Generates Markdown documentation for the configured services and tools.
This includes service descriptions, tool names, descriptions, and input schemas.`,
		Example: `  mcpctl config doc --config-path ./config.yaml
  mcpctl config doc > docs/api.md`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			osFs := afero.NewOsFs()
			cfgSettings := config.GlobalSettings()
			// Load checks flags and env vars to populate config struct
			if err := cfgSettings.Load(cmd, osFs); err != nil {
				return fmt.Errorf("configuration load failed: %w", err)
			}

			store := config.NewFileStore(osFs, cfgSettings.ConfigPaths())
			// We load services as "server" type to ensure we validate it correctly as a server config
			configs, err := config.LoadServices(context.Background(), store, "server")
			if err != nil {
				return fmt.Errorf("failed to load configurations: %w", err)
			}

			// Generate documentation
			doc, err := config.GenerateDocumentation(context.Background(), configs)
			if err != nil {
				return fmt.Errorf("failed to generate documentation: %w", err)
			}

			// Print to stdout
			_, err = fmt.Fprintln(cmd.OutOrStdout(), doc)
			return err
		},
	}
}
