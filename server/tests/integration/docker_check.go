// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package integration

import (
	"context"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// CanRunDockerWithMount checks if Docker can run a container with a volume mount.
// This is used to detect environments where Docker-in-Docker overlayfs mounts fail.
func CanRunDockerWithMount(t *testing.T) bool {
	t.Helper()
	dockerExe, dockerBaseArgs := getDockerCommand()

	tempDir := t.TempDir()
	absPath, err := filepath.Abs(tempDir)
	if err != nil {
		return false
	}

	containerName := "check-mount-capability"
	// Try to mimic the actual usage that fails: detached mode, alpine, simple command.
	// We use sleep 1 to ensure it runs briefly but finishes.
	// We check if it starts successfully.
	args := []string{"run", "--rm", "--name", containerName, "-v", absPath + ":/test", "alpine:latest", "echo", "check"}
	cmd := exec.CommandContext(context.Background(), dockerExe, append(dockerBaseArgs, args...)...) //nolint:gosec // Test helper
	output, err := cmd.CombinedOutput()

	if err != nil {
		outStr := string(output)
		if strings.Contains(outStr, "failed to mount") && strings.Contains(outStr, "overlay") && strings.Contains(outStr, "invalid argument") {
			t.Logf("Docker mount check failed with known overlayfs issue: %v", err)
			return false
		}
		// Log other errors but assume false to be safe, or true if it's just a transient thing?
		// Better to skip if we can't verify capability.
		t.Logf("Docker mount check failed: %v. Output: %s", err, outStr)
		return false
	}
	return true
}
