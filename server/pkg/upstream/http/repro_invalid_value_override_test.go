package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/pool"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/mcpany/core/server/pkg/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
)

func TestHTTPUpstream_URLConstruction_InvalidValueOverrideBug(t *testing.T) {
	// Part 1: Static FQN check
	testCases := []struct {
		name          string
		address       string
		endpointPath  string
		expectedFqn   string
	}{
		{
			name:         "base url with invalid value but valid key should be overridden",
			address:      "http://example.com/api?q=invalid%",
			endpointPath: "/v1/test?q=valid",
			// Desired behavior: override
			expectedFqn:  "GET http://example.com/api/v1/test?q=valid",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pm := pool.NewManager()
			tm := tool.NewManager(nil)
			upstream := NewUpstream(pm)

			configJSON := `{
				"name": "invalid-override-service",
				"http_service": {
					"address": "` + tc.address + `",
					"tools": [{"name": "test-op", "call_id": "test-op-call"}],
					"calls": {
						"test-op-call": {
							"id": "test-op-call",
							"method": "HTTP_METHOD_GET",
							"endpoint_path": "` + tc.endpointPath + `"
						}
					}
				}
			}`
			serviceConfig := &configv1.UpstreamServiceConfig{}
			require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

			serviceID, _, _, err := upstream.Register(context.Background(), serviceConfig, tm, nil, nil, false)
			assert.NoError(t, err)

			sanitizedToolName, _ := util.SanitizeToolName("test-op")
			toolID := serviceID + "." + sanitizedToolName
			registeredTool, ok := tm.GetTool(toolID)
			assert.True(t, ok)
			assert.NotNil(t, registeredTool)

			actualFqn := registeredTool.Tool().GetUnderlyingMethodFqn()
			assert.Equal(t, tc.expectedFqn, actualFqn)
		})
	}

	// Part 2: Runtime E2E verification
	t.Run("runtime_e2e_verification", func(t *testing.T) {
		t.Setenv("MCPANY_DANGEROUS_ALLOW_LOCAL_IPS", "true")

		var capturedRawQuery string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			capturedRawQuery = r.URL.RawQuery
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status": "ok"}`))
		}))
		defer server.Close()

		pm := pool.NewManager()
		tm := tool.NewManager(nil)
		upstream := NewUpstream(pm)

		// Base has invalid value for 'q', endpoint overrides 'q'.
		configJSON := `{
			"name": "invalid-override-runtime",
			"http_service": {
				"address": "` + server.URL + `?q=invalid%",
				"tools": [{"name": "test-op", "call_id": "test-op-call"}],
				"calls": {
					"test-op-call": {
						"id": "test-op-call",
						"method": "HTTP_METHOD_GET",
						"endpoint_path": "/test?q=valid"
					}
				}
			}
		}`
		serviceConfig := &configv1.UpstreamServiceConfig{}
		require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

		serviceID, _, _, err := upstream.Register(context.Background(), serviceConfig, tm, nil, nil, false)
		assert.NoError(t, err)

		sanitizedToolName, _ := util.SanitizeToolName("test-op")
		toolID := serviceID + "." + sanitizedToolName
		registeredTool, ok := tm.GetTool(toolID)
		assert.True(t, ok)

		t.Logf("Registered Tool FQN: %s", registeredTool.Tool().GetUnderlyingMethodFqn())

		// Verify FQN is correct (should exclude the invalid base param)
		// This assertion will FAIL before the fix
		fqn := registeredTool.Tool().GetUnderlyingMethodFqn()
		assert.NotContains(t, fqn, "invalid%", "FQN should not contain overridden invalid base param")

		// Execute the tool
		req := &tool.ExecutionRequest{
			ToolName:   "test-op",
			ToolInputs: []byte("{}"),
		}
		_, err = registeredTool.Execute(context.Background(), req)
		assert.NoError(t, err)

		// Expect only q=valid.
		assert.Equal(t, "q=valid", capturedRawQuery)
	})
}
