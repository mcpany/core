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
	"encoding/json"
	"testing"

	"github.com/mcpxy/core/tests/framework"
	"github.com/mcpxy/core/pkg/util"
	"github.com/google/jsonschema-go/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestUpstreamService_HTTP_Calculator(t *testing.T) {
	testCase := &framework.E2ETestCase{
		Name:                "HTTP Calculator",
		UpstreamServiceType: "http",
		BuildUpstream:       framework.BuildCalculatorServer,
		RegisterUpstream:    framework.RegisterCalculatorService,
		ValidateTool: func(t *testing.T, mcpxyEndpoint string) {
			serviceID := "e2e_http_calculator"
			toolName := "add"
			serviceKey, err := util.GenerateID(serviceID)
			if err != nil {
				t.Fatalf("Failed to generate service key: %v", err)
			}
			expectedToolName, err := util.GenerateToolID(serviceKey, toolName)
			if err != nil {
				t.Fatalf("Failed to generate tool ID: %v", err)
			}

			expectedSchema := &jsonschema.Schema{
				Type: "object",
				Properties: map[string]*jsonschema.Schema{
					"a": {
						Type:        "string",
						Description: "first number",
					},
					"b": {
						Type:        "string",
						Description: "second number",
					},
				},
			}

			var expectedSchemaMap map[string]interface{}
			schemaBytes, err := json.Marshal(expectedSchema)
			if err != nil {
				t.Fatalf("Failed to marshal expected schema: %v", err)
			}
			if err := json.Unmarshal(schemaBytes, &expectedSchemaMap); err != nil {
				t.Fatalf("Failed to unmarshal expected schema: %v", err)
			}

			expectedTool := &mcp.Tool{
				Name:        expectedToolName,
				Description: "add two numbers",
				InputSchema: expectedSchemaMap,
			}
			framework.ValidateRegisteredTool(t, mcpxyEndpoint, expectedTool)
		},
		InvokeAIClient: func(t *testing.T, mcpxyEndpoint string) {
			// TODO: Implement AI client invocation
		},
	}

	framework.RunE2ETest(t, testCase)
}
