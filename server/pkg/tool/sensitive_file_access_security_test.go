package tool

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestSensitiveFileAccessBlocked(t *testing.T) {
	// Create a dummy .env file
	tmpDir := t.TempDir()
	envPath := filepath.Join(tmpDir, ".env")
	secretContent := "MCPANY_API_KEY=super-secret-key"
	err := os.WriteFile(envPath, []byte(secretContent), 0600)
	require.NoError(t, err)

	// Save original WD
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		_ = os.Chdir(originalWd)
	}()

	// Switch to tmpDir
	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	// Define a tool that uses 'cat' to read a file
	tool := v1.Tool_builder{
		Name: proto.String("cat-tool"),
	}.Build()

	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("cat"),
		Local:   proto.Bool(true),
	}.Build()

	param := configv1.CommandLineParameterMapping_builder{
		Schema: configv1.ParameterSchema_builder{
			Name: proto.String("file"),
		}.Build(),
	}.Build()

	callDef := configv1.CommandLineCallDefinition_builder{
		Args:       []string{"{{file}}"},
		Parameters: []*configv1.CommandLineParameterMapping{param},
	}.Build()

    // Create policies (empty)
    policies := []*configv1.CallPolicy{}

	// Create the tool
	localTool := NewLocalCommandTool(tool, service, callDef, policies, "call-id")

	// Execute the tool requesting access to .env
	req := &ExecutionRequest{
		ToolName: "cat-tool",
		// JSON payload for input
		ToolInputs: []byte(`{"file": ".env"}`),
	}

	result, err := localTool.Execute(context.Background(), req)

	// We expect an error now because .env is blocked
	assert.Error(t, err)
	if err != nil {
		assert.Contains(t, err.Error(), "access to sensitive file")
	}
	assert.Nil(t, result)
}
