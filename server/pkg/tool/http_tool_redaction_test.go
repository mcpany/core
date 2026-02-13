package tool

import (
	"context"
	"encoding/json"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestHTTPTool_PrepareInputsAndURL_Redaction(t *testing.T) {
	// Setup
	secretValue := "super-secret-key"
	paramName := "apiKey"

	// Create a parameter mapping that uses a secret
	paramMapping := configv1.HttpParameterMapping_builder{
		Schema: configv1.ParameterSchema_builder{
			Name: proto.String(paramName),
		}.Build(),
		Secret: configv1.SecretValue_builder{
			PlainText: proto.String(secretValue),
		}.Build(),
	}.Build()

	callDef := configv1.HttpCallDefinition_builder{
		Parameters: []*configv1.HttpParameterMapping{paramMapping},
	}.Build()

	toolDef := v1.Tool_builder{
		Name:                proto.String("test-tool"),
		UnderlyingMethodFqn: proto.String("GET http://example.com/api/{{apiKey}}/data"),
	}.Build()

	// Initialize HTTPTool using constructor to ensure all fields (secretParams etc) are initialized
	httpTool := NewHTTPTool(toolDef, nil, "service-id", nil, callDef, nil, nil, "")

	// Prepare inputs
	// prepareInputsAndURL expects inputs to be resolved? No, it takes raw ToolInputs (JSON)
	// But it resolves secrets using util.ResolveSecret.
	// We need to provide the secret value.
	// In the test setup above, we used PlainText secret, so ResolveSecret should work without extra context setup.

	// We don't need to pass the secret in inputs, because it's configured as a secret param.
	// But prepareInputsAndURL iterates t.parameters.

	req := &ExecutionRequest{
		ToolName:   "test-tool",
		ToolInputs: json.RawMessage(`{}`), // Secrets are not passed in input JSON typically
	}

	// Action: Call prepareInputsAndURL
	_, rawURL, redactedURL, _, err := httpTool.prepareInputsAndURL(context.Background(), req)
	require.NoError(t, err)

	// Assertion
	assert.Contains(t, rawURL, secretValue, "Raw URL must contain the secret value for the request to work")
	assert.NotContains(t, redactedURL, secretValue, "Redacted URL must NOT contain the secret value")
	assert.Contains(t, redactedURL, redactedPlaceholder, "Redacted URL must contain the placeholder")

	t.Logf("Raw URL: %s", rawURL)
	t.Logf("Redacted URL: %s", redactedURL)
}
