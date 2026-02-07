// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package main implements the mcpctl command line interface.
package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/mcpany/core/server/pkg/config"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

var (
	// Version is the version of the mcpctl CLI.
	// It is set at build time via -ldflags.
	Version = "dev"
)

// main is the entry point for the mcpctl CLI.
//
// Summary: Executes the CLI root command.
//
// Side Effects:
//   - Executes the CLI command.
//   - May terminate the process with exit code 1 on error.
func main() {
	if err := newRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}

// newRootCmd creates the root Cobra command for the CLI.
//
// Summary: Initializes the root command and registers subcommands.
//
// Returns:
//   - *cobra.Command: The configured root command with subcommands attached.
func newRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "mcpctl",
		Short: "mcpctl is a command line tool for MCP Any",
	}

	validateCmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate the configuration file",
		RunE: func(cmd *cobra.Command, _ []string) error {
			osFs := afero.NewOsFs()
			cfg := config.GlobalSettings()
			// Load checks flags and env vars to populate config struct
			if err := cfg.Load(cmd, osFs); err != nil {
				return fmt.Errorf("configuration load failed: %w", err)
			}

			store := config.NewFileStore(osFs, cfg.ConfigPaths())
			configs, err := config.LoadServices(context.Background(), store, "server")
			if err != nil {
				return fmt.Errorf("failed to load configurations from %v: %w", cfg.ConfigPaths(), err)
			}

			var allErrors []string
			// Validate using the same logic as the server
			if validationErrors := config.Validate(context.Background(), configs, config.Server); len(validationErrors) > 0 {
				for _, e := range validationErrors {
					allErrors = append(allErrors, e.Error())
				}
			}

			if len(allErrors) > 0 {
				return fmt.Errorf("configuration validation failed with errors:\n- %s", strings.Join(allErrors, "\n- "))
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), "Configuration is valid.")
			if err != nil {
				return fmt.Errorf("failed to print success message: %w", err)
			}
			return nil
		},
	}
	// Bind flags like --config, etc.
	config.BindRootFlags(rootCmd)
	rootCmd.AddCommand(validateCmd)
	rootCmd.AddCommand(newDoctorCmd())
	rootCmd.AddCommand(newToolCmd())
	rootCmd.AddCommand(newImportCmd())

	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version number of mcpctl",
		RunE: func(cmd *cobra.Command, _ []string) error {
			_, err := fmt.Fprintf(cmd.OutOrStdout(), "mcpctl version %s\n", Version)
			return err
		},
	}
	rootCmd.AddCommand(versionCmd)

	return rootCmd
}
