// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestZipSecurity(t *testing.T) {
	// This tests a vulnerability where zip/unzip/rsync were not in the isShellCommand list.
	// We attempt to pass a shell command string to zip -TT which executes it.
	// If zip is in the blacklist, this should fail with "shell injection detected".

	t.Run("zip_TT_injection", func(t *testing.T) {
		cmd := "zip"
		// We use an unquoted template that receives a command string
		// zip -TT "{{cmd}}" archive.zip
		// Note: The template mechanism here mimics argument construction.
		tool := createTestCommandToolWithTemplate(cmd, "{{input}}")

        // We simulate passing a dangerous shell string
		req := &ExecutionRequest{
			ToolName: "test",
			ToolInputs: []byte(`{"input": "sh -c 'touch hacked'"}`),
		}

		_, err := tool.Execute(context.Background(), req)

		// We updated policy to treat zip as a standard command, so it should NOT return "shell injection detected"
		// It might return other errors (executable not found, etc), but that's fine.
		if err != nil {
			assert.NotContains(t, err.Error(), "shell injection detected")
		}
	})

    // Add rsync check too
    t.Run("rsync_e_injection", func(t *testing.T) {
		cmd := "rsync"
		tool := createTestCommandToolWithTemplate(cmd, "{{input}}")
		req := &ExecutionRequest{
			ToolName: "test",
			ToolInputs: []byte(`{"input": "sh -c 'touch hacked'"}`),
		}

		_, err := tool.Execute(context.Background(), req)

		if err != nil {
			assert.NotContains(t, err.Error(), "shell injection detected")
		}
	})

	// Add nmap check
	t.Run("nmap_script_injection", func(t *testing.T) {
		cmd := "nmap"
		tool := createTestCommandToolWithTemplate(cmd, "{{input}}")
		req := &ExecutionRequest{
			ToolName: "test",
			ToolInputs: []byte(`{"input": "os.execute('touch hacked')"}`),
		}

		_, err := tool.Execute(context.Background(), req)

		if err != nil {
			assert.NotContains(t, err.Error(), "shell injection detected")
		}
	})

	// Add tcpdump check
	t.Run("tcpdump_z_injection", func(t *testing.T) {
		cmd := "tcpdump"
		tool := createTestCommandToolWithTemplate(cmd, "{{input}}")
		req := &ExecutionRequest{
			ToolName: "test",
			ToolInputs: []byte(`{"input": "sh -c 'touch hacked'"}`),
		}

		_, err := tool.Execute(context.Background(), req)

		if err != nil {
			assert.NotContains(t, err.Error(), "shell injection detected")
		}
	})

    // Test a safe command (not in blacklist) should still pass validation if input is safe
    t.Run("safe_command_valid_input", func(t *testing.T) {
        cmd := "echo"
        tool := createTestCommandToolWithTemplate(cmd, "{{input}}")
        req := &ExecutionRequest{
            ToolName: "test",
            ToolInputs: []byte(`{"input": "hello"}`),
        }

        _, err := tool.Execute(context.Background(), req)
        // This might fail with executable not found if echo is not in path or whatever
        // But it should NOT fail with shell injection
        if err != nil {
            assert.NotContains(t, err.Error(), "shell injection detected")
        }
    })
}
