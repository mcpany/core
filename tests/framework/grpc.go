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

package framework

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/mcpany/core/tests/integration"
	"github.com/stretchr/testify/require"
)

// BuildGRPCHealthServer builds and starts a new instance of the gRPC health server.
func BuildGRPCHealthServer(t *testing.T, tls bool) *integration.ManagedProcess {
	port := integration.FindFreePort(t)
	root, err := integration.GetProjectRoot()
	require.NoError(t, err)
	args := []string{fmt.Sprintf("--port=%d", port)}
	if tls {
		args = append(args, "--tls")
	}
	proc := integration.NewManagedProcess(t, "grpc_health_server", filepath.Join(root, "build/test/bin/grpc_health_server"), args, nil)
	proc.Port = port
	return proc
}
