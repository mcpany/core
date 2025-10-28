//go:build e2e

/*
 * Copyright 2025 Author(s) of MCP-XY
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

package upstream

import (
	"os"
	"testing"

	"github.com/mcpxy/core/tests/framework"
	"github.com/stretchr/testify/require"
)

func TestUpstreamService_Websocket(t *testing.T) {
	gemini := framework.NewGeminiCLI(t)
	gemini.Install()

	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		t.Skip("GEMINI_API_KEY is not set")
	}

	testCase := &framework.E2ETestCase{
		Name:                "Websocket Echo Server",
		UpstreamServiceType: "websocket",
		BuildUpstream:       framework.BuildWebsocketServer,
		RegisterUpstream:    framework.RegisterWebsocketService,
		InvokeAIClient: func(t *testing.T, mcpxyEndpoint string) {
			framework.VerifyMCPClient(t, mcpxyEndpoint)
			gemini.AddMCP("mcpxy-server", mcpxyEndpoint)
			defer gemini.RemoveMCP("mcpxy-server")
			output, err := gemini.Run(apiKey, "echo hello world from websocket")
			require.NoError(t, err)
			require.Contains(t, output, "hello world from websocket")
		},
	}

	framework.RunE2ETest(t, testCase)
}
