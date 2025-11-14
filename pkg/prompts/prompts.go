// Copyright (C) 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package prompts

import (
	"context"
	"fmt"
	"io"

	"github.com/mcpany/core/proto/api/v1"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/valyala/fasttemplate"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Service implements the PromptService gRPC service.
type Service struct {
	v1.UnimplementedPromptServiceServer
	prompts map[string]*configv1.PromptDefinition
}

// NewService creates a new instance of the prompts Service.
func NewService(promptConfigs []*configv1.PromptDefinition) *Service {
	s := &Service{
		prompts: make(map[string]*configv1.PromptDefinition),
	}
	for _, p := range promptConfigs {
		s.prompts[p.GetName()] = p
	}
	return s
}

// RegisterService registers the prompts service with a gRPC server.
func (s *Service) RegisterService(server *grpc.Server) {
	v1.RegisterPromptServiceServer(server, s)
}

func stringp(s string) *string {
	return &s
}

func boolp(b bool) *bool {
	return &b
}

// ListPrompts returns a list of available prompts.
func (s *Service) ListPrompts(ctx context.Context, req *v1.ListPromptsRequest) (*v1.ListPromptsResponse, error) {
	prompts := make([]*v1.Prompt, 0, len(s.prompts))
	for _, p := range s.prompts {
		args := make([]*v1.PromptArgument, 0, len(p.GetArguments()))
		for _, a := range p.GetArguments() {
			argBuilder := &v1.PromptArgument_builder{}
			argBuilder.Name = stringp(a.GetName())
			argBuilder.Description = stringp(a.GetDescription())
			argBuilder.Required = boolp(a.GetRequired())
			args = append(args, argBuilder.Build())
		}

		promptBuilder := &v1.Prompt_builder{}
		promptBuilder.Name = stringp(p.GetName())
		promptBuilder.Title = stringp(p.GetTitle())
		promptBuilder.Description = stringp(p.GetDescription())
		promptBuilder.Arguments = args
		prompts = append(prompts, promptBuilder.Build())
	}

	resp := &v1.ListPromptsResponse{}
	resp.SetPrompts(prompts)
	return resp, nil
}

// GetPrompt returns a specific prompt.
func (s *Service) GetPrompt(ctx context.Context, req *v1.GetPromptRequest) (*v1.GetPromptResponse, error) {
	prompt, ok := s.prompts[req.GetName()]
	if !ok {
		return nil, status.Errorf(codes.NotFound, "prompt %q not found", req.GetName())
	}

	messages := make([]*v1.PromptMessage, 0, len(prompt.GetMessages()))
	for _, m := range prompt.GetMessages() {
		content := &v1.Content{}
		if text := m.GetContent().GetText(); text != nil {
			template, err := fasttemplate.NewTemplate(text.GetText(), "{{", "}}")
			if err != nil {
				return nil, status.Errorf(codes.Internal, "failed to parse template: %v", err)
			}
			s := template.ExecuteFuncString(func(w io.Writer, tag string) (int, error) {
				if val, ok := req.GetArguments().GetFields()[tag]; ok {
					return w.Write([]byte(val.GetStringValue()))
				}
				return w.Write([]byte(fmt.Sprintf("{{%s}}", tag)))
			})

			textContent := &v1.TextContent{}
			textContent.SetText(s)
			content.SetText(textContent)
		} else if image := m.GetContent().GetImage(); image != nil {
			imageContent := &v1.ImageContent{}
			imageContent.SetData(image.GetData())
			imageContent.SetMimeType(image.GetMimeType())
			content.SetImage(imageContent)
		} else if audio := m.GetContent().GetAudio(); audio != nil {
			audioContent := &v1.AudioContent{}
			audioContent.SetData(audio.GetData())
			audioContent.SetMimeType(audio.GetMimeType())
			content.SetAudio(audioContent)
		} else if resource := m.GetContent().GetResource(); resource != nil {
			res := &v1.Resource{}
			res.SetUri(resource.GetResource().GetUri())
			res.SetMimeType(resource.GetResource().GetMimeType())
			if static := resource.GetResource().GetStatic(); static != nil {
				if text := static.GetTextContent(); text != "" {
					res.SetText(text)
				} else if binary := static.GetBinaryContent(); binary != nil {
					res.SetBlob(binary)
				}
			}
			resContent := &v1.ResourceContent{}
			resContent.SetResource(res)
			content.SetResource(resContent)
		}

		role := v1.PromptMessage_Role(m.GetRole())
		promptMessage := (&v1.PromptMessage_builder{
			Role:    &role,
			Content: content,
		}).Build()
		messages = append(messages, promptMessage)
	}

	respBuilder := &v1.GetPromptResponse_builder{}
	respBuilder.Description = stringp(prompt.GetDescription())
	respBuilder.Messages = messages
	resp := respBuilder.Build()

	return resp, nil
}
