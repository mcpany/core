// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"

	"github.com/mcpany/core/server/pkg/check"
	"github.com/spf13/cobra"
)

func newCheckCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "check [file]",
		Short: "Check configuration file against strict schema",
		Long:  "Check validates the configuration file against the JSON schema and provides line-number precise error messages.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := args[0]
			results, err := check.ValidateFile(cmd.Context(), path)
			if err != nil {
				// If ValidateFile returns an error (like file not found), return it.
				// If it returns validation errors, they are in results.
				return err
			}

			if len(results) == 0 {
				fmt.Printf("✅  Configuration %q is valid.\n", path)
				return nil
			}

			fmt.Printf("❌  Configuration %q has %d errors:\n", path, len(results))
			for _, r := range results {
				fmt.Println(r.String())
			}

			return fmt.Errorf("configuration check failed")
		},
	}
	return cmd
}
