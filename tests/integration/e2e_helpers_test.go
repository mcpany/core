// Copyright 2024 Author
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package integration

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateTempConfigFile(t *testing.T) {
	t.Parallel()
	// Create a dummy config
	serviceName := "test-service"
	config := &configv1.UpstreamServiceConfig{}
	config.SetName(serviceName)

	// Create a dummy config file
	configFilePath := CreateTempConfigFile(t, config)
	defer os.Remove(configFilePath)

	// Verify that the file was created
	_, err := os.Stat(configFilePath)
	require.NoError(t, err, "Temp config file should exist")
}

func TestProjectRoot(t *testing.T) {
	t.Parallel()
	// Get the project root
	root := ProjectRoot(t)

	// Verify that the project root contains a known file
	_, err := os.Stat(filepath.Join(root, "go.mod"))
	require.NoError(t, err, "go.mod should exist in the project root")
}

func TestManagedProcess(t *testing.T) {
	t.Parallel()
	mp := NewManagedProcess(t, "echo", "echo", []string{"hello"}, nil)
	require.NoError(t, mp.Start())
	// Wait for the process to finish on its own
	<-mp.waitDone
	assert.Contains(t, mp.StdoutString(), "hello")
	// Calling stop on an already finished process should be a no-op
	mp.Stop()
}

func TestWaitForText(t *testing.T) {
	t.Parallel()
	mp := NewManagedProcess(t, "echo", "sh", []string{"-c", "sleep 0.1; echo hello"}, nil)
	require.NoError(t, mp.Start())
	mp.WaitForText(t, "hello", 1*time.Second)
	mp.Stop()
}

func TestDockerHelpers(t *testing.T) {
	t.Skip("Skipping test due to DockerHub rate limiting issues")
	t.Parallel()
	if !IsDockerSocketAccessible() {
		t.Skip("Docker is not available")
	}

	// Test StartDockerContainer
	imageName := "mcpany-e2e-time-server"
	containerName := fmt.Sprintf("mcpany-test-container-%d", time.Now().UnixNano())
	cleanup := StartDockerContainer(t, imageName, containerName, []string{"-d"})
	defer cleanup()

	// Verify the container is running
	dockerExe, dockerArgs := getDockerCommand()
	psCmd := exec.Command(dockerExe, append(dockerArgs, "ps", "-f", fmt.Sprintf("name=%s", containerName))...)
	out, err := psCmd.Output()
	require.NoError(t, err, "docker ps command failed. Output: %s", string(out))
	assert.Contains(t, string(out), containerName)

	// Test StartRedisContainer
	_, redisCleanup := StartRedisContainer(t)
	defer redisCleanup()

	// The StartRedisContainer function has internal checks to ensure the container
	// starts and is responsive. A successful return is a pass.
}

// countProcesses checks how many processes with a given name are running.
func countProcesses(t *testing.T, name string) int {
	t.Helper()
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("tasklist")
	default:
		cmd = exec.Command("pgrep", "-c", name)
	}

	output, err := cmd.Output()
	if err != nil {
		// On Unix, pgrep returns exit code 1 if no process is found
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return 0
		}
		// On Windows, tasklist might fail if no process is found, but it's better to check output.
		if runtime.GOOS == "windows" && strings.Contains(string(output), "No tasks are running") {
			return 0
		}
		t.Logf("Error checking for running processes '%s': %v, output: %s", name, err, string(output))
		return 0 // Return 0 on error to avoid false positives
	}

	if runtime.GOOS == "windows" {
		return strings.Count(strings.ToLower(string(output)), strings.ToLower(name))
	}
	count := 0
	fmt.Sscanf(string(output), "%d", &count)
	return count
}

// TestStartNatsServerCleanup ensures that the NATS server is properly cleaned up
// after the test, leaving no orphaned processes. This is critical for preventing
// resource leaks in the test suite.
func TestStartNatsServerCleanup(t *testing.T) {
	// Let's give a bit of time for any lingering processes from previous tests to terminate
	time.Sleep(2 * time.Second)

	initialCount := countProcesses(t, "nats-server")
	t.Logf("Initial nats-server processes: %d", initialCount)

	// We'll run this multiple times to ensure that cleanup is consistent.
	for i := 0; i < 3; i++ {
		t.Run(fmt.Sprintf("iteration_%d", i), func(t *testing.T) {
			_, cleanup := StartNatsServer(t)
			// Give the server a moment to fully start up and be discoverable by pgrep.
			time.Sleep(500 * time.Millisecond)
			// At this point, we expect at least one nats-server process to be running.
			require.Greater(t, countProcesses(t, "nats-server"), initialCount, "nats-server process should have been started")
			cleanup()
			// After cleanup, the number of nats-server processes should return to the initial count.
			// We use `Eventually` to account for the small delay it might take for the OS to reap the terminated process.
			require.Eventually(t, func() bool {
				return countProcesses(t, "nats-server") == initialCount
			}, 5*time.Second, 250*time.Millisecond, "nats-server process was not cleaned up properly")
		})
	}

	// Final check to ensure no processes were leaked after all iterations.
	finalCount := countProcesses(t, "nats-server")
	require.Equal(t, initialCount, finalCount, "The number of nats-server processes should be the same as before the test suite started.")
}
