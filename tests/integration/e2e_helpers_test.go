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
	"os"
	"os/exec"
	"path/filepath"
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
	// Skipping this test due to Docker pull rate limits in the CI environment.
	t.Skip("Skipping Docker tests due to CI environment limitations.")
	t.Parallel()
	if !IsDockerSocketAccessible() {
		t.Skip("Docker is not available")
	}

	// Test StartDockerContainer
	imageName := "hello-world"
	cleanup := StartDockerContainer(t, imageName, "mcpany-test-container")
	defer cleanup()

	// Verify the container is running
	cmd := exec.Command("docker", "ps", "-f", "name=mcpany-test-container")
	out, err := cmd.Output()
	require.NoError(t, err)
	assert.Contains(t, string(out), "mcpany-test-container")
}
