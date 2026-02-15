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
	_ = w.Close() // Close the write end of the pipe immediately to cause a write error
	rootCmd.SetArgs([]string{"version"})

	// Run the command and check the error
	err = rootCmd.Execute()
	_ = r.Close()

	// The command should return an error
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to print version")
}

// newRootCmd is a helper function to create a new root command for testing.
// This function is not available in the test package, so it's copied here.
func newRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "mcpany",
		Short: "MCP Any is a versatile proxy for backend services.",
		RunE: func(_ *cobra.Command, _ []string) error {
			return nil
		},
	}

	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version number of mcpany",
		RunE: func(cmd *cobra.Command, _ []string) error {
			_, err := fmt.Fprintf(cmd.OutOrStdout(), "%s version %s\n", "mcpany", "dev")
			if err != nil {
				return fmt.Errorf("failed to print version: %w", err)
			}
			return nil
		},
	}
	rootCmd.AddCommand(versionCmd)

	return rootCmd
}
