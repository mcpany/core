// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package command

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/mcpany/core/server/pkg/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLocalExecutor_Security(t *testing.T) {
	// Note: We do not call t.Parallel() because validation.SetAllowedPaths modifies global state.
	// This test might still race with other tests if they run in parallel, but we can't fix other tests easily.

	t.Run("CommandValidation", func(t *testing.T) {
		executor := NewLocalExecutor()

		// 1. Command without separators should pass (e.g. "ls" or "echo")
		// We use "echo" assuming it is in PATH.
		_, _, _, err := executor.Execute(context.Background(), "echo", []string{"hello"}, "", nil)
		// We only care that it didn't fail with "command path not allowed"
		if err != nil {
			assert.NotContains(t, err.Error(), "command path not allowed")
		}

		// 2. Relative path in CWD should pass
		// Create a dummy script in CWD
		cwd, err := os.Getwd()
		require.NoError(t, err)
		tmpScript, err := os.CreateTemp(cwd, "testscript*.sh")
		require.NoError(t, err)
		scriptName := tmpScript.Name()
		_ = tmpScript.Close()
		defer os.Remove(scriptName)
		_ = os.Chmod(scriptName, 0700)

		// Execute using relative path
		relPath := "." + string(os.PathSeparator) + filepath.Base(scriptName)
		_, _, _, err = executor.Execute(context.Background(), relPath, nil, "", nil)
		if err != nil {
			assert.NotContains(t, err.Error(), "command path not allowed")
		}

		// 3. Absolute path outside CWD/AllowedPaths should fail
		tmpDir := t.TempDir() // Created in /tmp usually, outside CWD
		absScript := filepath.Join(tmpDir, "bad.sh")
		err = os.WriteFile(absScript, []byte("#!/bin/sh\necho hi"), 0700)
		require.NoError(t, err)

		// Make sure tmpDir is not accidentally CWD or allowed
		validation.SetAllowedPaths(nil)

		_, _, _, err = executor.Execute(context.Background(), absScript, nil, "", nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "command path not allowed")

		// 4. Absolute path in AllowedPaths should pass
		validation.SetAllowedPaths([]string{tmpDir})
		defer validation.SetAllowedPaths(nil)

		_, _, _, err = executor.Execute(context.Background(), absScript, nil, "", nil)
		// It should pass validation now.
		if err != nil {
			assert.NotContains(t, err.Error(), "command path not allowed")
		}
	})

	t.Run("ExecuteWithStdIO_Security", func(t *testing.T) {
		executor := NewLocalExecutor()
		tmpDir := t.TempDir()
		absScript := filepath.Join(tmpDir, "bad_stdio.sh")
		err := os.WriteFile(absScript, []byte("#!/bin/sh\necho hi"), 0700)
		require.NoError(t, err)

		validation.SetAllowedPaths(nil)

		_, _, _, _, err = executor.ExecuteWithStdIO(context.Background(), absScript, nil, "", nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "command path not allowed")
	})
}
