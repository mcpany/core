package tool

import (
	"context"
	"io"
	"strings"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/command"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

// sentinelMockExecutor implements command.Executor
type sentinelMockExecutor struct{}

func (m *sentinelMockExecutor) Execute(ctx context.Context, cmd string, args []string, workingDir string, env []string) (io.ReadCloser, io.ReadCloser, <-chan int, error) {
	outR, outW := io.Pipe()
	outW.Close()
	errR, errW := io.Pipe()
	errW.Close()
	ch := make(chan int, 1)
	ch <- 0
	close(ch)
	return outR, errR, ch, nil
}

func (m *sentinelMockExecutor) ExecuteWithStdIO(ctx context.Context, cmd string, args []string, workingDir string, env []string) (io.WriteCloser, io.ReadCloser, io.ReadCloser, <-chan int, error) {
	return nil, nil, nil, nil, nil
}

func TestInterpreterInjectionBypass(t *testing.T) {
	// Vulnerability: RCE via Backslash Injection in Single Quoted Argument.
	// We demonstrate that we can inject a backslash to escape the closing quote of a Python string,
	// allowing us to execute arbitrary code.

	toolDef := v1.Tool_builder{
		Name: proto.String("python-tool"),
	}.Build()

	service := (&configv1.CommandLineUpstreamService_builder{
		Command: proto.String("python3"),
	}).Build()

	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-c", "print('{{input1}}', '{{input2}}')"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{
					Name: proto.String("input1"),
				}.Build(),
			}.Build(),
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{
					Name: proto.String("input2"),
				}.Build(),
			}.Build(),
		},
	}.Build()

	// Attack Scenario:
	// input1 = "\" (Backslash)
	// input2 = ", open('/etc/passwd')) #" (Code to execute)
	//
	// Resulting arg: print('\', ', open('/etc/passwd')) #')
	//
	// Python Parse:
	// String 1: '\', ' (The backslash escapes the first closing quote, so the string eats the comma, space and the opening quote of the second arg!)
	// Code: , open('/etc/passwd')) #
	//
	// Note: We use double quotes in input2 to avoid single quote restriction in Level 2 check.
	// input2 = `, open("/etc/passwd")) #`

	input1 := "\\"
	input2 := `, open("/etc/passwd")) #`

	cmdTool := NewCommandTool(toolDef, service, callDef, nil, "call-1")

	// Inject mock executor
	ct, ok := cmdTool.(*CommandTool)
	assert.True(t, ok)
	ct.executorFactory = func(env *configv1.ContainerEnvironment) command.Executor {
		return &sentinelMockExecutor{}
	}

	// Construct JSON inputs
	// input1 needs escaping for JSON: \ -> \\
	escapedInput1 := strings.ReplaceAll(input1, `\`, `\\`)
	// input2 needs escaping for JSON: " -> \"
	escapedInput2 := strings.ReplaceAll(input2, `"`, `\"`)

	jsonInputs := `{"input1": "` + escapedInput1 + `", "input2": "` + escapedInput2 + `"}`

	req := &ExecutionRequest{
		ToolName:   "python-tool",
		ToolInputs: []byte(jsonInputs),
	}

	_, err := cmdTool.Execute(context.Background(), req)

	if err == nil {
		t.Fatal("Vulnerability Confirmed: Backslash injection allowed! The fix is not working.")
	} else {
		t.Logf("Execution failed: %v", err)
		if strings.Contains(err.Error(), "injection detected") {
			// This is what we expect AFTER the fix
			t.Log("Secure: Injection detected.")
		} else {
			t.Fatalf("Execution failed with unexpected error: %v", err)
		}
	}
}
