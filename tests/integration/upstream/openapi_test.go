//go:build e2e

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

package upstream

import (
	"os"
	"testing"

	"github.com/mcpany/core/tests/framework"
	"github.com/stretchr/testify/require"
)

func TestUpstreamService_OpenAPI(t *testing.T) {
	gemini := framework.NewGeminiCLI(t)
	gemini.Install()

	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		t.Skip("GEMINI_API_KEY is not set")
	}

	testCase := &framework.E2ETestCase{
		Name:                "OpenAPI Weather Server",
		UpstreamServiceType: "openapi",
		BuildUpstream:       framework.BuildOpenAPIWeatherServer,
		RegisterUpstream:    framework.RegisterOpenAPIWeatherService,
		InvokeAIClient: func(t *testing.T, mcpanyEndpoint string) {
			framework.VerifyMCPClient(t, mcpanyEndpoint)
			gemini.AddMCP("mcpany-server", mcpanyEndpoint)
			defer gemini.RemoveMCP("mcpany-server")
			output, err := gemini.Run(apiKey, "what is the weather in london")
			require.NoError(t, err)
			require.Contains(t, output, "Cloudy, 15°C")
		},
	}

	framework.RunE2ETest(t, testCase)
}
