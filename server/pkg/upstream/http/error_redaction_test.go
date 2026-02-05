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

func TestHTTPUpstream_ErrorRedaction(t *testing.T) {
	t.Setenv("MCPANY_DANGEROUS_ALLOW_LOCAL_IPS", "true")

	// Create a test server that returns a 500 error with a stack trace
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{
			"error": "Internal Server Error",
			"stackTrace": "java.lang.NullPointerException at com.company.MyClass.method(MyClass.java:123)",
			"traceback": "Traceback (most recent call last): File \"app.py\", line 10, in <module>",
			"nested": {
				"cause": "something bad",
				"stack_trace": "nested stack trace"
			}
		}`))
	}))
	defer server.Close()

	pm := pool.NewManager()
	tm := tool.NewManager(nil)
	upstream := NewUpstream(pm)

	configJSON := `{
		"name": "redact-test",
		"http_service": {
			"address": "` + server.URL + `",
			"tools": [{"name": "boom", "call_id": "boom-call"}],
			"calls": {
				"boom-call": {
					"id": "boom-call",
					"method": "HTTP_METHOD_GET",
					"endpoint_path": "/"
				}
			}
		}
	}`
	serviceConfig := configv1.UpstreamServiceConfig_builder{}.Build()
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

	serviceID, _, _, err := upstream.Register(context.Background(), serviceConfig, tm, nil, nil, false)
	require.NoError(t, err)

	sanitizedToolName, _ := util.SanitizeToolName("boom")
	toolID := serviceID + "." + sanitizedToolName
	registeredTool, ok := tm.GetTool(toolID)
	require.True(t, ok)

	_, err = registeredTool.Execute(context.Background(), &tool.ExecutionRequest{
		ToolName: "boom",
	})
	require.Error(t, err)

	errMsg := err.Error()
	// Verify redaction
	assert.Contains(t, errMsg, "stackTrace")
	assert.NotContains(t, errMsg, "java.lang.NullPointerException")
	assert.Contains(t, errMsg, "[REDACTED]")

	assert.Contains(t, errMsg, "traceback")
	assert.NotContains(t, errMsg, "File \"app.py\"")
}
