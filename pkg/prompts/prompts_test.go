// Copyright (C) 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package prompts

import (
	"context"
	"testing"

	"github.com/mcpany/core/proto/api/v1"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestListPrompts(t *testing.T) {
	promptConfigs := []*configv1.PromptDefinition{
		(&configv1.PromptDefinition_builder{
			Name: stringp("code_review"),
		}).Build(),
	}
	service := NewService(promptConfigs)
	req := &v1.ListPromptsRequest{}
	resp, err := service.ListPrompts(context.Background(), req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp.GetPrompts(), 1)
	assert.Equal(t, "code_review", resp.GetPrompts()[0].GetName())
}

func TestGetPrompt(t *testing.T) {
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
	}
	service := NewService(promptConfigs)
	args, err := structpb.NewStruct(map[string]interface{}{
		"code": "def hello():\n  print('world')",
	})
	require.NoError(t, err)

	req := &v1.GetPromptRequest{}
	req.SetName("code_review")
	req.SetArguments(args)

	resp, err := service.GetPrompt(context.Background(), req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "Code review prompt", resp.GetDescription())
	assert.Len(t, resp.GetMessages(), 1)
	assert.Equal(t, "Please review this Python code:\ndef hello():\n  print('world')", resp.GetMessages()[0].GetContent().GetText().GetText())
}

func TestGetPromptNotFound(t *testing.T) {
	service := NewService(nil)
	req := &v1.GetPromptRequest{}
	req.SetName("non_existent_prompt")

	resp, err := service.GetPrompt(context.Background(), req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	st, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.NotFound, st.Code())
}

func TestGetPromptResourceContent(t *testing.T) {
	promptConfigs := []*configv1.PromptDefinition{
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
	service := NewService(promptConfigs)

	req := &v1.GetPromptRequest{}
	req.SetName("resource_prompt")

	resp, err := service.GetPrompt(context.Background(), req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp.GetMessages(), 1)
	resource := resp.GetMessages()[0].GetContent().GetResource().GetResource()
	assert.Equal(t, "resource://example", resource.GetUri())
	assert.Equal(t, "text/plain", resource.GetMimeType())
	assert.Equal(t, "Resource content", resource.GetText())
}
