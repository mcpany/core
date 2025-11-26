/*
 * Copyright 2024 Author(s) of MCP Any
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"fmt"

	"github.com/mcpany/core/pkg/config"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

func newValidateCommand(fs afero.Fs) *cobra.Command {
	validateCmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate the MCP Any configuration",
		Long: `Validate the MCP Any configuration files.

This command checks the syntax and semantics of the configuration files,
ensuring that they are well-formed and valid.`,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			configPaths, _ := cmd.Flags().GetStringSlice("config-path")
			if len(configPaths) == 0 {
				return fmt.Errorf("at least one configuration path must be specified")
			}

			store := config.NewFileStore(fs, configPaths)
			if _, err := config.LoadServices(store, "server"); err != nil {
				return fmt.Errorf("configuration is invalid: %w", err)
			}

			fmt.Fprintln(cmd.OutOrStdout(), "Configuration is valid")
			return nil
		},
	}

	return validateCmd
}
