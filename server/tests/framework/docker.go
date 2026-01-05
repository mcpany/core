// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package framework

import (
	"os"
	"os/exec"
	"testing"
)

// GetDockerCommand returns the command to run docker, prepending sudo if the
// USE_SUDO_FOR_DOCKER environment variable is set to "true".
func GetDockerCommand(t *testing.T) []string {
	t.Helper()
	if os.Getenv("USE_SUDO_FOR_DOCKER") == "true" {
		return []string{"sudo", "docker"}
	}
	return []string{"docker"}
}

// CommandExists checks if a command exists in the system's PATH.
func CommandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}
