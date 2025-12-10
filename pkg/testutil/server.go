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

package testutil

import (
	"context"
	"testing"

	"github.com/mcpany/core/pkg/app"
	"github.com/mcpany/core/pkg/mcpserver"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

func StartTestServer(ctx context.Context, t *testing.T, configPaths []string) (*app.Application, *mcpserver.Server, error) {
	t.Helper()

	fs := afero.NewOsFs()
	app := app.NewApplication()

	go func() {
		err := app.Run(ctx, fs, false, "0", "", configPaths, 0)
		require.NoError(t, err)
	}()

	return app, app.Server, nil
}

func CreateTestClient(t *testing.T, ctx context.Context, mcpServer *mcpserver.Server) (*mcp.Client, *mcp.ClientSession) {
	t.Helper()

	client := mcp.NewClient(&mcp.Implementation{Name: "test-client"}, nil)
	clientTransport, serverTransport := mcp.NewInMemoryTransports()

	serverSession, err := mcpServer.Server().Connect(ctx, serverTransport, nil)
	require.NoError(t, err)

	clientSession, err := client.Connect(ctx, clientTransport, nil)
	require.NoError(t, err)

	return client, clientSession
}
