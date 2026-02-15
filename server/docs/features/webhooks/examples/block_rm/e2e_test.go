package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/mcpany/core/server/pkg/bus"
	"github.com/mcpany/core/server/pkg/prompt"
	"github.com/mcpany/core/server/pkg/resource"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/mcpany/core/server/pkg/upstream/command"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestBlockRmE2E_Binary(t *testing.T) {
	// 1. Build Webhook Server Binary
	tempDir := t.TempDir()
	binaryPath := filepath.Join(tempDir, "block_rm_server")

	buildCmd := exec.Command("go", "build", "-o", binaryPath, "main.go")
	buildCmd.Stdout = os.Stdout
	buildCmd.Stderr = os.Stderr
	require.NoError(t, buildCmd.Run(), "Failed to build webhook server")

	// 2. Start Webhook Server
	// Find a free port
	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	port := l.Addr().(*net.TCPAddr).Port
	_ = l.Close()

	cmd := exec.Command(binaryPath)
	cmd.Env = append(os.Environ(), fmt.Sprintf("PORT=%d", port))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	require.NoError(t, cmd.Start(), "Failed to start webhook server")
	defer func() {
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
	}()

	url := fmt.Sprintf("http://127.0.0.1:%d/validate", port)
	waitForServer(t, url)

	// 3. Setup Tool Manager and Upstream
	busProvider, _ := bus.NewProvider(nil)
	toolManager := tool.NewManager(busProvider)
	promptManager := prompt.NewManager()
	resourceManager := resource.NewManager()

	config := &configv1.UpstreamServiceConfig{
		Name: proto.String("busybox"),
		ServiceConfig: &configv1.UpstreamServiceConfig_CommandLineService{
			CommandLineService: &configv1.CommandLineUpstreamService{
				Command: proto.String("echo"),
				Local:   proto.Bool(true),
				Tools: []*configv1.ToolDefinition{
					{
						Name:   proto.String("execute_command"),
						CallId: proto.String("execute_command"),
						InputSchema: &structpb.Struct{
							Fields: map[string]*structpb.Value{
								"properties": structpb.NewStructValue(&structpb.Struct{
									Fields: map[string]*structpb.Value{
										"command": structpb.NewStructValue(&structpb.Struct{
											Fields: map[string]*structpb.Value{
												"type": structpb.NewStringValue("string"),
											},
										}),
									},
								}),
								"required": structpb.NewListValue(&structpb.ListValue{
									Values: []*structpb.Value{structpb.NewStringValue("command")},
								}),
								"type": structpb.NewStringValue("object"),
							},
						},
					},
				},
				Calls: map[string]*configv1.CommandLineCallDefinition{
					"execute_command": {
						Id: proto.String("execute_command"),
						Parameters: []*configv1.CommandLineParameterMapping{
							{
								Schema: &configv1.ParameterSchema{
									Name:       proto.String("command"),
									Type:       configv1.ParameterType_STRING.Enum(),
									IsRequired: proto.Bool(true),
								},
							},
						},
						Args: []string{"running", "{{command}}"},
					},
				},
			},
		},
		PreCallHooks: []*configv1.CallHook{
			{
				Name: proto.String("block_rm"),
				HookConfig: &configv1.CallHook_Webhook{
					Webhook: &configv1.WebhookConfig{
						Url:     url, // Not a pointer
						Timeout: durationpb.New(2 * time.Second),
					},
				},
			},
		},
	}

	upstream := command.NewUpstream()
	serviceID, _, _, err := upstream.Register(context.Background(), config, toolManager, promptManager, resourceManager, false)
	require.NoError(t, err)
	require.NotEmpty(t, serviceID)

	// 4. Test Allowed Logic
	t.Run("Allowed Command", func(t *testing.T) {
		inputs := map[string]interface{}{
			"command": "ls -la",
		}
		inputsBytes, _ := json.Marshal(inputs)
		req := &tool.ExecutionRequest{
			ToolName:   "busybox.execute_command",
			ToolInputs: json.RawMessage(inputsBytes),
		}

		result, err := toolManager.ExecuteTool(context.Background(), req)
		require.NoError(t, err)
		resMap := result.(map[string]interface{})
		stdout := resMap["stdout"].(string)
		require.Contains(t, stdout, "ls -la")
	})

	// 5. Test Blocked Logic
	t.Run("Blocked Command", func(t *testing.T) {
		inputs := map[string]interface{}{
			"command": "rm -rf /",
		}
		inputsBytes, _ := json.Marshal(inputs)
		req := &tool.ExecutionRequest{
			ToolName:   "busybox.execute_command",
			ToolInputs: json.RawMessage(inputsBytes),
		}

		result, err := toolManager.ExecuteTool(context.Background(), req)
		require.Error(t, err)
		require.Nil(t, result)
		require.Contains(t, err.Error(), "denied by webhook")
	})
}

func waitForServer(t *testing.T, url string) {
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		resp, err := http.Get(url)
		if err == nil || (resp != nil && resp.StatusCode == http.StatusMethodNotAllowed) {
			if resp != nil {
				resp.Body.Close()
			}
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Fatalf("Server failed to start at %s within timeout", url)
}
