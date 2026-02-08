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

// NewService creates and returns a new Service instance.
//
// Summary: Initializes a new Prompt Service.
//
// Parameters:
//   - promptManager: ManagerInterface. The manager for prompt lifecycles.
//
// Returns:
//   - *Service: The initialized Service.
func NewService(promptManager ManagerInterface) *Service {
	s := &Service{
		promptManager: promptManager,
	}
	// s.promptManager.OnListChanged(s.onPromptListChanged)
	return s
}

// SetMCPServer sets the MCP server instance for the service.
//
// Summary: Configures the MCP server instance.
//
// Parameters:
//   - mcpServer: *mcp.Server. The MCP server instance.
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

// ListPrompts handles the "prompts/list" MCP request.
//
// Summary: Lists all available prompts.
//
// Parameters:
//   - ctx: context.Context. The request context.
//   - req: *mcp.ListPromptsRequest. The list prompts request.
//
// Returns:
//   - *mcp.ListPromptsResult: The list of prompts.
//   - error: An error if the operation fails.
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

// GetPrompt handles the "prompts/get" MCP request.
//
// Summary: Retrieves and executes a specific prompt.
//
// Parameters:
//   - ctx: context.Context. The request context.
//   - req: *mcp.GetPromptRequest. The get prompt request containing arguments.
//
// Returns:
//   - *mcp.GetPromptResult: The prompt execution result.
//   - error: An error if the prompt is not found or execution fails.
//
// Throws/Errors:
//   - ErrPromptNotFound: If the prompt does not exist.
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
