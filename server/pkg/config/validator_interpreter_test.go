// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"context"
	"os"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestValidatorInterpreterDetection(t *testing.T) {
	// Create a temporary directory for dummy commands
	tmpDir, err := os.MkdirTemp("", "mcp-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Helper to create a dummy executable
	createDummyCmd := func(name string) string {
		path := tmpDir + "/" + name
		f, err := os.Create(path)
		require.NoError(t, err)
		f.Close()
		os.Chmod(path, 0755)
		return path
	}

	// Helper to run validation
	validate := func(cmd string, args []string) []ValidationError {
		config := func() *configv1.McpAnyServerConfig {
			conn := configv1.McpStdioConnection_builder{
				Command: proto.String(cmd),
				Args:    args,
			}.Build()

			mcp := configv1.McpUpstreamService_builder{
				StdioConnection: conn,
			}.Build()

			svc := configv1.UpstreamServiceConfig_builder{
				Name:       proto.String("test-service"),
				McpService: mcp,
			}.Build()

			return configv1.McpAnyServerConfig_builder{
				UpstreamServices: []*configv1.UpstreamServiceConfig{svc},
			}.Build()
		}()
		return Validate(context.Background(), config, Server)
	}

	t.Run("GoodTool_ShouldNotBeInterpreter", func(t *testing.T) {
		// "good-tool" starts with "go", but is NOT "go".
		// Previous bug: detected as "go" interpreter and validated "result.json".
		// Fix: should NOT be detected, so "result.json" (which doesn't exist) is ignored.
		cmdPath := createDummyCmd("good-tool")
		errs := validate(cmdPath, []string{"--output", "result.json"})
		assert.Empty(t, errs, "Validation should pass for non-interpreter command even if args look like files")
	})

	t.Run("Python3.9_ShouldBeInterpreter", func(t *testing.T) {
		// "python3.9" is a valid interpreter variant.
		// Should detect as interpreter and validate "script.py".
		cmdPath := createDummyCmd("python3.9")
		// "script.py" does NOT exist, so validation should FAIL.
		errs := validate(cmdPath, []string{"script.py"})
		assert.NotEmpty(t, errs)
		assert.Contains(t, errs[0].Err.Error(), "looks like a script file but does not exist")
	})

	t.Run("Python3.9Config_ShouldNotBeInterpreter", func(t *testing.T) {
		// "python3.9-config" starts with "python3", but suffix "-config" is not version.
		// Should NOT detect as interpreter.
		cmdPath := createDummyCmd("python3.9-config")
		// "script.py" does NOT exist, but since it's not an interpreter, it shouldn't be validated.
		errs := validate(cmdPath, []string{"script.py"})
		assert.Empty(t, errs)
	})

	t.Run("GoRun_ShouldBeInterpreter", func(t *testing.T) {
		// "go" is a valid interpreter.
		// Should detect as interpreter and validate "main.go".
		// We use "go" from path (assuming it exists), or create a dummy "go".
		// Creating dummy "go" in tmpDir to be safe.
		cmdPath := createDummyCmd("go")
		// "main.go" does NOT exist, so validation should FAIL.
		errs := validate(cmdPath, []string{"run", "main.go"})
		assert.NotEmpty(t, errs)
		assert.Contains(t, errs[0].Err.Error(), "looks like a script file but does not exist")
	})

	t.Run("NodeGyp_ShouldNotBeInterpreter", func(t *testing.T) {
		// "node-gyp" starts with "node", but is not "node".
		cmdPath := createDummyCmd("node-gyp")
		errs := validate(cmdPath, []string{"binding.gyp"})
		assert.Empty(t, errs)
	})

	t.Run("GoogleChrome_ShouldNotBeInterpreter", func(t *testing.T) {
		// "google-chrome" starts with "go".
		cmdPath := createDummyCmd("google-chrome")
		errs := validate(cmdPath, []string{"--headless", "page.html"})
		assert.Empty(t, errs)
	})
}
