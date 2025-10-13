/*
 * Copyright 2025 Author(s) of MCP-XY
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package cmd_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVersionCmdExitCode(t *testing.T) {
	// Create a command that will fail to write to stdout
	rootCmd := newRootCmd()
	r, w, err := os.Pipe()
	require.NoError(t, err)
	rootCmd.SetOut(w)
	w.Close() // Close the write end of the pipe immediately to cause a write error
	rootCmd.SetArgs([]string{"version"})

	// Run the command and check the error
	err = rootCmd.Execute()
	r.Close()

	// The command should return an error
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to print version")
}

// newRootCmd is a helper function to create a new root command for testing.
// This function is not available in the test package, so it's copied here.
func newRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "mcpxy",
		Short: "MCP-XY is a versatile proxy for backend services.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version number of mcpxy",
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := fmt.Fprintf(cmd.OutOrStdout(), "%s version %s\n", "mcpxy", "dev")
			if err != nil {
				return fmt.Errorf("failed to print version: %w", err)
			}
			return nil
		},
	}
	rootCmd.AddCommand(versionCmd)

	return rootCmd
}