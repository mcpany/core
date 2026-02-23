package tool_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestHTTPTool_ContentType_Set_For_JSON_Template(t *testing.T) {
	expectedBody := `{"q": "test"}`
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		assert.JSONEq(t, expectedBody, string(body))

		// VERIFY: Content-Type MUST be application/json
		contentType := r.Header.Get("Content-Type")
		assert.Equal(t, "application/json", contentType, "Content-Type header should be set to application/json for JSON template")

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	})

	callDef := configv1.HttpCallDefinition_builder{
		Method: configv1.HttpCallDefinition_HTTP_METHOD_POST.Enum(),
		InputTransformer: configv1.InputTransformer_builder{
			Template: lo.ToPtr(`{"q": "{{query}}"}`),
		}.Build(),
		Parameters: []*configv1.HttpParameterMapping{
			configv1.HttpParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{Name: proto.String("query")}.Build(),
			}.Build(),
		},
	}.Build()

	httpTool, server := setupHTTPToolTest(t, handler, callDef)
	defer server.Close()

	inputs := json.RawMessage(`{"query": "test"}`)
	req := &tool.ExecutionRequest{ToolInputs: inputs}
	_, err := httpTool.Execute(context.Background(), req)
	require.NoError(t, err)
}
