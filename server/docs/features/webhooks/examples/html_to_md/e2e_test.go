// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/mcpany/core/server/pkg/bus"
	"github.com/mcpany/core/server/pkg/pool"
	"github.com/mcpany/core/server/pkg/prompt"
	"github.com/mcpany/core/server/pkg/resource"
	"github.com/mcpany/core/server/pkg/tool"
	http_upstream "github.com/mcpany/core/server/pkg/upstream/http"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestHtmlToMdE2E_Binary(t *testing.T) {
	// 1. Build Webhook Server
	tempDir := t.TempDir()
	binaryPath := filepath.Join(tempDir, "html_to_md_server")

	buildCmd := exec.Command("go", "build", "-o", binaryPath, "main.go")
	buildCmd.Stdout = os.Stdout
	buildCmd.Stderr = os.Stderr
	require.NoError(t, buildCmd.Run(), "Failed to build webhook server")

	// 2. Start Webhook Server
	cmd := exec.Command(binaryPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	require.NoError(t, cmd.Start(), "Failed to start webhook server")
	defer func() {
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
	}()

	webhookUrl := "http://127.0.0.1:8082/convert"
	waitForServer(t, webhookUrl)

	// 3. Setup Mock Upstream Service
	mockUpstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("<h1>Hello World</h1><p>This is a test.</p>"))
	}))
	defer mockUpstream.Close()

	// 4. Configure MCP Any
	busProvider, _ := bus.NewProvider(nil)
	toolManager := tool.NewManager(busProvider)
	promptManager := prompt.NewManager()
	resourceManager := resource.NewManager()
	poolManager := pool.NewManager()

	config := &configv1.UpstreamServiceConfig{
		Name: proto.String("chrome-devtools"),
		ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
			HttpService: &configv1.HttpUpstreamService{
				Address: proto.String(mockUpstream.URL),
				Tools: []*configv1.ToolDefinition{
					{
						Name:   proto.String("get_page_content"),
						CallId: proto.String("get_page_content"),
						InputSchema: &structpb.Struct{
							Fields: map[string]*structpb.Value{
								"properties": structpb.NewStructValue(&structpb.Struct{
									Fields: map[string]*structpb.Value{
										"url": structpb.NewStructValue(&structpb.Struct{
											Fields: map[string]*structpb.Value{
												"type": structpb.NewStringValue("string"),
											},
										}),
									},
								}),
								"type": structpb.NewStringValue("object"),
							},
						},
					},
				},
				Calls: map[string]*configv1.HttpCallDefinition{
					"get_page_content": {
						Id:           proto.String("get_page_content"),
						EndpointPath: proto.String("/json/version"),
						Method:       configv1.HttpCallDefinition_HTTP_METHOD_GET.Enum(),
					},
				},
			},
		},
		PostCallHooks: []*configv1.CallHook{
			{
				Name: proto.String("convert_to_md"),
				HookConfig: &configv1.CallHook_Webhook{
					Webhook: &configv1.WebhookConfig{
						Url:     webhookUrl,
						Timeout: durationpb.New(2 * time.Second),
					},
				},
			},
		},
	}

	// 5. Register and Execute
	upstream := http_upstream.NewUpstream(poolManager)
	serviceID, _, _, err := upstream.Register(context.Background(), config, toolManager, promptManager, resourceManager, false)
	require.NoError(t, err)
	require.NotEmpty(t, serviceID)

	t.Run("Convert HTML to Markdown", func(t *testing.T) {
		inputs := map[string]interface{}{
			"url": "http://example.com",
		}
		inputsBytes, _ := json.Marshal(inputs)
		req := &tool.ExecutionRequest{
			ToolName:   "chrome-devtools.get_page_content",
			ToolInputs: json.RawMessage(inputsBytes),
		}

		result, err := toolManager.ExecuteTool(context.Background(), req)
		require.NoError(t, err)
		require.NotNil(t, result)

		// Verification
		resMap, ok := result.(map[string]interface{})
		require.True(t, ok, "Result should be a map")

		content, ok := resMap["content"].(string)
		if !ok {
			t.Logf("Result keys: %v", resMap)
			t.FailNow()
		}

		require.Contains(t, content, "# Hello World")
		require.Contains(t, content, "This is a test.")
		require.Equal(t, "markdown", resMap["format"])
	})
}

func waitForServer(t *testing.T, url string) {
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		resp, err := http.Get(url)
		if err == nil || (resp != nil && resp.StatusCode == 405) {
			if resp != nil {
				resp.Body.Close()
			}
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Fatalf("Server failed to start at %s within timeout", url)
}
