// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package prompt

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Service handles the business logic for the prompts feature. It provides
// methods for listing available prompts and retrieving a specific prompt by
// name.
type Service struct {
	promptManager ManagerInterface
	mcpServer     *mcp.Server
}

// NewService creates a new prompt service instance.
//
// Parameters:
//   - promptManager: ManagerInterface. The manager responsible for prompt lifecycle and storage.
//
// Returns:
//   - *Service: A new Service instance.
func NewService(promptManager ManagerInterface) *Service {
	s := &Service{
		promptManager: promptManager,
	}
	// s.promptManager.OnListChanged(s.onPromptListChanged)
	return s
}

// SetMCPServer sets the MCP server instance for the service.
//
// mcpServer is the mcpServer.
func (s *Service) SetMCPServer(mcpServer *mcp.Server) {
	s.mcpServer = mcpServer
	s.promptManager.SetMCPServer(NewMCPServerProvider(mcpServer))
}

// onPromptListChanged notifies clients that the prompt list has changed.
// Currently this is a no-op as the go-sdk does not expose a public Notify method
// for PromptListChanged.
// func (s *Service) onPromptListChanged() {
//    // Waiting for SDK support for public notification triggering
//	  // log.Warn("Prompt list changed notification not sent (SDK limitation)")
// }

// ListPrompts handles the "prompts/list" MCP request. It retrieves the list of
// available prompts from the Manager, converts them to the MCP format, and
// returns them to the client.
func (s *Service) ListPrompts(
	_ context.Context,
	_ *mcp.ListPromptsRequest,
) (*mcp.ListPromptsResult, error) {
	prompts := s.promptManager.ListPrompts()
	mcpPrompts := make([]*mcp.Prompt, len(prompts))
	for i, p := range prompts {
		mcpPrompts[i] = p.Prompt()
	}
	return &mcp.ListPromptsResult{
		Prompts: mcpPrompts,
	}, nil
}

// GetPrompt handles the "prompts/get" MCP request. It retrieves a specific
// prompt by name from the Manager and executes it with the provided
// arguments, returning the result. If the prompt is not found, it returns a
// ErrPromptNotFound error.
func (s *Service) GetPrompt(
	ctx context.Context,
	req *mcp.GetPromptRequest,
) (*mcp.GetPromptResult, error) {
	p, ok := s.promptManager.GetPrompt(req.Params.Name)
	if !ok {
		return nil, ErrPromptNotFound
	}

	argsBytes, err := json.Marshal(req.Params.Arguments)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal prompt arguments: %w", err)
	}

	return p.Get(ctx, argsBytes)
}
