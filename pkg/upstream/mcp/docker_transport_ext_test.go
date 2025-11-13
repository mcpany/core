/*
 * Copyright 2025 Author(s) of MCP Any
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

package mcp

import (
	"context"
	"testing"

	"github.com/mcpany/core/pkg/util"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
)

func TestDockerTransport_Connect_ContainerAttach_Error(t *testing.T) {
	if !util.IsDockerSocketAccessible() {
		t.Skip("Docker socket not accessible, skipping integration test.")
	}

	ctx := context.Background()
	stdioConfig := &configv1.McpStdioConnection{}
	stdioConfig.SetContainerImage("alpine:latest")
	stdioConfig.SetCommand("/bin/true") // This command exits immediately
	transport := &DockerTransport{StdioConfig: stdioConfig}

	_, err := transport.Connect(ctx)
	assert.Error(t, err)
	// The container might exit before we can attach, causing an attach error.
	// Or it might exit after we attach but before we start, causing a start error.
	// Or even during the copy from the hijacked response.
	// We'll just check for a non-nil error.
}

func TestDockerTransport_Connect_ContainerStart_Error(t *testing.T) {
	if !util.IsDockerSocketAccessible() {
		t.Skip("Docker socket not accessible, skipping integration test.")
	}

	ctx := context.Background()
	stdioConfig := &configv1.McpStdioConnection{}
	stdioConfig.SetContainerImage("alpine:latest")
	stdioConfig.SetCommand("/nonexistent/command")
	transport := &DockerTransport{StdioConfig: stdioConfig}

	_, err := transport.Connect(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to start container")
}
