// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// newInitCmd creates the init command for generating a basic config file interactively.
//
// It returns a Cobra command that writes a default config.yaml to the current directory.
//
// Returns:
//   - *cobra.Command: The configured init command.
func newInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Interactively generate a basic config.yaml",
		RunE: func(cmd *cobra.Command, args []string) error {
			reader := bufio.NewReader(cmd.InOrStdin())

			_, err := fmt.Fprint(cmd.OutOrStdout(), "Enter the name of your first service (e.g., my-service): ")
			if err != nil {
				return err
			}

			serviceName, err := reader.ReadString('\n')
			if err != nil {
				return err
			}
			serviceName = strings.TrimSpace(serviceName)
			if serviceName == "" {
				serviceName = "my-service"
			}

			_, err = fmt.Fprint(cmd.OutOrStdout(), "Enter the upstream HTTP address (e.g., https://api.example.com): ")
			if err != nil {
				return err
			}

			address, err := reader.ReadString('\n')
			if err != nil {
				return err
			}
			address = strings.TrimSpace(address)
			if address == "" {
				address = "https://api.example.com"
			}

			configContent := fmt.Sprintf(`global_settings:
  log_level: info

upstream_services:
  - name: %s
    http_service:
      address: %s
`, serviceName, address)

			outputFile, err := cmd.Flags().GetString("output")
			if err != nil {
				outputFile = "config.yaml"
			}

			if err := os.WriteFile(outputFile, []byte(configContent), 0644); err != nil {
				return fmt.Errorf("failed to write %s: %w", outputFile, err)
			}
			_, err = fmt.Fprintln(cmd.OutOrStdout(), "Successfully generated config.yaml!")
			return err
		},
	}

	cmd.Flags().StringP("output", "o", "config.yaml", "Output file path")

	return cmd
}
