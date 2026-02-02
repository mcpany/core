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

func newInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize a new configuration file",
		RunE: func(cmd *cobra.Command, _ []string) error {
			reader := bufio.NewReader(cmd.InOrStdin())

			_, err := fmt.Fprint(cmd.OutOrStdout(), "Enter service name: ")
			if err != nil {
				return err
			}
			name, err := reader.ReadString('\n')
			if err != nil {
				return err
			}
			name = strings.TrimSpace(name)
			if name == "" {
				name = "my-service"
			}

			_, err = fmt.Fprint(cmd.OutOrStdout(), "Enter service type (http/command) [http]: ")
			if err != nil {
				return err
			}
			svcType, err := reader.ReadString('\n')
			if err != nil {
				return err
			}
			svcType = strings.TrimSpace(svcType)
			if svcType == "" {
				svcType = "http"
			}

			configContent := fmt.Sprintf(`upstream_services:
  - name: %s
`, name)

			switch svcType {
			case "http":
				configContent += `    http_service:
      address: "http://example.com"
`
			case "command":
				configContent += `    command_line_service:
      command: "echo"
      arguments: ["hello"]
`
			default:
				// Default to simple
				configContent += `    # Unknown type selected, defaulting to disabled
    disable: true
`
			}

			// Add basic global settings
			configContent += `global_settings:
  mcp_listen_address: ":50050"
`

			// Check if config.yaml already exists
			if _, err := os.Stat("config.yaml"); err == nil {
				return fmt.Errorf("config.yaml already exists")
			}

			err = os.WriteFile("config.yaml", []byte(configContent), 0600)
			if err != nil {
				return fmt.Errorf("failed to write config.yaml: %w", err)
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), "Successfully created config.yaml")
			return err
		},
	}
	return cmd
}
