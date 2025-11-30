/*
 * Copyright 2024 Author(s) of MCP Any
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

package integration

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	apiv1 "github.com/mcpany/core/proto/api/v1"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/durationpb"
)

func TestHTTPTool_Execute_WithRetry_E2E(t *testing.T) {
	attempt := 0
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if attempt < 2 {
			attempt++
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		attempt++
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok"}`))
	})
	mockUpstream := httptest.NewServer(handler)
	defer mockUpstream.Close()

	serverInfo := StartInProcessMCPANYServer(t, "RetryE2ETest")
	defer serverInfo.CleanupFunc()

	serviceID := "retry-e2e-test-service"
	callID := "test-call"
	fullToolName := serviceID + "/-/" + callID

	retryPolicy := configv1.RetryConfig_builder{
		NumberOfRetries: lo.ToPtr[int32](3),
		BaseBackoff:     durationpb.New(10 * time.Millisecond),
	}.Build()

	resilience := configv1.ResilienceConfig_builder{
		RetryPolicy: retryPolicy,
	}.Build()

	httpService := configv1.HttpUpstreamService_builder{
		Address: &mockUpstream.URL,
		Calls: map[string]*configv1.HttpCallDefinition{
			callID: configv1.HttpCallDefinition_builder{
				Method:       configv1.HttpCallDefinition_HTTP_METHOD_GET.Enum(),
				EndpointPath: lo.ToPtr("/"),
			}.Build(),
		},
	}.Build()

	serviceConfig := configv1.UpstreamServiceConfig_builder{
		Name:        &serviceID,
		HttpService: httpService,
		Resilience:  resilience,
	}.Build()

	req := apiv1.RegisterServiceRequest_builder{
		Config: serviceConfig,
	}.Build()

	RegisterServiceViaAPI(t, serverInfo.RegistrationClient, req)

	params := &mcp.CallToolParams{
		Name:      fullToolName,
		Arguments: json.RawMessage(`{}`),
	}
	result, err := serverInfo.CallTool(context.Background(), params)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Len(t, result.Content, 1)

	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok)

	var resultMap map[string]any
	err = json.Unmarshal([]byte(textContent.Text), &resultMap)
	require.NoError(t, err)

	assert.Equal(t, "ok", resultMap["status"])
	assert.Equal(t, 3, attempt)
}
