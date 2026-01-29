package mcpserver

import "github.com/modelcontextprotocol/go-sdk/mcp"

// GetRouter returns the server's router. This is for testing purposes only.
func (s *Server) GetRouter() *Router {
	return s.router
}

func (s *Server) RouterMiddleware(next mcp.MethodHandler) mcp.MethodHandler {
	return s.routerMiddleware(next)
}

func (s *Server) ToolListFilteringMiddleware(next mcp.MethodHandler) mcp.MethodHandler {
	return s.toolListFilteringMiddleware(next)
}

func (s *Server) ResourceListFilteringMiddleware(next mcp.MethodHandler) mcp.MethodHandler {
	return s.resourceListFilteringMiddleware(next)
}

func (s *Server) PromptListFilteringMiddleware(next mcp.MethodHandler) mcp.MethodHandler {
	return s.promptListFilteringMiddleware(next)
}
