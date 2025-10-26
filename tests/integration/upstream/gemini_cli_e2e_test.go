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

func TestGeminiCLIE2E_Calculator(t *testing.T) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		t.Skip("GEMINI_API_KEY not set, skipping test")
	}

	gemini := framework.NewGeminiCLI(t)
	gemini.Install()

	testCase := &framework.E2ETestCase{
		Name:                "Gemini CLI with HTTP Calculator Service",
		UpstreamServiceType: "http",
		BuildUpstream:       framework.BuildCalculatorServer,
		RegisterUpstream:    framework.RegisterCalculatorService,
		InvokeAIClient: func(t *testing.T, mcpxyEndpoint string) {
			gemini.AddMCP("mcpxy-server", mcpxyEndpoint)
			defer gemini.RemoveMCP("mcpxy-server")
			output, err := gemini.Run(apiKey, "gemini-2.5-flash", "what is the result of 10 + 5")
			require.NoError(t, err)
			require.Contains(t, output, "15")
		},
	}
	framework.RunE2ETest(t, testCase)
}
