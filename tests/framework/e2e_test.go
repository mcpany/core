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
	"testing"

	"github.com/mcpany/core/tests/integration"
	apiv1 "github.com/mcpany/core/proto/api/v1"
)

func TestE2E(t *testing.T) {
	t.Run("TestE2ERegistration", func(t *testing.T) {
		RunE2ETest(t, &E2ETestCase{
			Name:                "test-e2e-registration",
			UpstreamServiceType: "http",
			BuildUpstream: func(t *testing.T) *integration.ManagedProcess {
				return BuildHTTPEchoServer(t)
			},
			RegisterUpstream: func(t *testing.T, registrationClient apiv1.RegistrationServiceClient, upstreamEndpoint string) {
				RegisterHTTPEchoService(t, registrationClient, upstreamEndpoint)
			},
			ValidateTool: func(t *testing.T, mcpanyEndpoint string) {
				// No validation needed for this test
			},
			ValidateMiddlewares: func(t *testing.T, mcpanyEndpoint string, upstreamEndpoint string) {
				// No validation needed for this test
			},
			InvokeAIClient: func(t *testing.T, mcpanyEndpoint string) {
				// No AI client invocation needed for this test
			},
			RegistrationMethods: []RegistrationMethod{GRPCRegistration},
			GenerateUpstreamConfig: func(upstreamEndpoint string) string {
				return fmt.Sprintf(`
services:
- name: "e2e-http-echo"
  http:
    address: "%s"
    tools:
    - name: "echo"
      call_id: "echo-call"
    calls:
      "echo-call":
        endpoint_path: "/echo"
        method: "POST"
`, upstreamEndpoint)
			},
			RegisterUpstreamWithJSONRPC: func(t *testing.T, mcpanyEndpoint, upstreamEndpoint string) {
				integration.RegisterHTTPServiceWithJSONRPC(t, mcpanyEndpoint, "e2e-http-echo", upstreamEndpoint, "echo", "/echo", "POST", nil)
			},
		})
	})
}
