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

package integration

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/mcpany/core/pkg/prompts"
	"github.com/mcpany/core/proto/api/v1"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/types/known/structpb"
)

func stringp(s string) *string {
	return &s
}

func TestPromptServiceConfigIntegration(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	promptConfigs := []*configv1.PromptDefinition{
		(&configv1.PromptDefinition_builder{
			Name:        stringp("code_review"),
			Description: stringp("Code review prompt"),
			Messages: []*configv1.PromptMessage{
				(&configv1.PromptMessage_builder{
					Content: (&configv1.Content_builder{
						Text: (&configv1.TextContent_builder{
							Text: stringp("Please review this Python code:\n{{code}}"),
						}).Build(),
					}).Build(),
				}).Build(),
			},
		}).Build(),
		(&configv1.PromptDefinition_builder{
			Name:        stringp("resource_prompt"),
			Description: stringp("Resource prompt"),
			Messages: []*configv1.PromptMessage{
				(&configv1.PromptMessage_builder{
					Content: (&configv1.Content_builder{
						Resource: (&configv1.ResourceContent_builder{
							Resource: (&configv1.ResourceDefinition_builder{
								Uri:      stringp("resource://example"),
								MimeType: stringp("text/plain"),
								Static: (&configv1.StaticResource_builder{
									TextContent: stringp("Resource content"),
								}).Build(),
							}).Build(),
						}).Build(),
					}).Build(),
				}).Build(),
			},
		}).Build(),
	}

	lis := bufconn.Listen(1024 * 1024)
	s := grpc.NewServer()
	prompts.NewService(promptConfigs).RegisterService(s)

	go func() {
		if err := s.Serve(lis); err != nil {
			t.Fatalf("Server exited with error: %v", err)
		}
	}()

	dialer := func(context.Context, string) (net.Conn, error) {
		return lis.Dial()
	}

	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(dialer), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer conn.Close()

	client := v1.NewPromptServiceClient(conn)

	// Test ListPrompts
	listResp, err := client.ListPrompts(ctx, &v1.ListPromptsRequest{})
	require.NoError(t, err)
	assert.Len(t, listResp.GetPrompts(), 2)
	assert.Equal(t, "code_review", listResp.GetPrompts()[0].GetName())

	// Test GetPrompt
	args, err := structpb.NewStruct(map[string]interface{}{
		"code": "def hello():\n  print('world')",
	})
	require.NoError(t, err)

	getReq := &v1.GetPromptRequest{}
	getReq.SetName("code_review")
	getReq.SetArguments(args)

	getResp, err := client.GetPrompt(ctx, getReq)
	require.NoError(t, err)
	assert.Equal(t, "Code review prompt", getResp.GetDescription())
	assert.Len(t, getResp.GetMessages(), 1)
	assert.Equal(t, "Please review this Python code:\ndef hello():\n  print('world')", getResp.GetMessages()[0].GetContent().GetText().GetText())

	// Test GetPrompt with resource
	getReq = &v1.GetPromptRequest{}
	getReq.SetName("resource_prompt")

	getResp, err = client.GetPrompt(ctx, getReq)
	require.NoError(t, err)
	assert.Equal(t, "Resource prompt", getResp.GetDescription())
	assert.Len(t, getResp.GetMessages(), 1)
	resource := getResp.GetMessages()[0].GetContent().GetResource().GetResource()
	assert.Equal(t, "resource://example", resource.GetUri())
	assert.Equal(t, "text/plain", resource.GetMimeType())
	assert.Equal(t, "Resource content", resource.GetText())
}
