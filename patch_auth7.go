package main

import (
    "io/ioutil"
    "log"
    "strings"
)

func main() {
    content, err := ioutil.ReadFile("server/pkg/middleware/auth.go")
    if err != nil {
        log.Fatal(err)
    }

    modified := strings.Replace(string(content), `			// Special handling for tool calls
			if method == consts.MethodToolsCall {
				if r, ok := req.(*mcp.CallToolRequest); ok && r != nil && r.Params != nil {
					// We expect tool names to be prefixed with the service ID (e.g. "service.tool")
					// Optimization: Use strings.Cut to avoid allocating a slice.
					if before, _, found := strings.Cut(r.Params.Name, "."); found {
						serviceID = before
					}
				}
			}`, `			// Special handling for tool calls
			if method == consts.MethodToolsCall {
				if r, ok := req.(*mcp.CallToolRequest); ok && r != nil && r.Params != nil {
					// We expect tool names to be prefixed with the service ID (e.g. "service.tool")
					// Optimization: Use strings.Cut to avoid allocating a slice.
					if before, _, found := strings.Cut(r.Params.Name, "."); found {
						serviceID = before
					}
				} else if m, ok := req.(map[string]interface{}); ok {
					if params, ok := m["params"].(map[string]interface{}); ok {
						if name, ok := params["name"].(string); ok {
							if before, _, found := strings.Cut(name, "."); found {
								serviceID = before
							}
						}
					}
				}
			}`, 1)

    modified = strings.Replace(modified, `			// Special handling for prompts/get
			if method == consts.MethodPromptsGet {
				if r, ok := req.(*mcp.GetPromptRequest); ok && r != nil && r.Params != nil {
					// We expect prompt names to be prefixed with the service ID (e.g. "service.prompt")
					// Optimization: Use strings.Cut to avoid allocating a slice.
					if before, _, found := strings.Cut(r.Params.Name, "."); found {
						serviceID = before
					}
				}
			}`, `			// Special handling for prompts/get
			if method == consts.MethodPromptsGet {
				if r, ok := req.(*mcp.GetPromptRequest); ok && r != nil && r.Params != nil {
					// We expect prompt names to be prefixed with the service ID (e.g. "service.prompt")
					// Optimization: Use strings.Cut to avoid allocating a slice.
					if before, _, found := strings.Cut(r.Params.Name, "."); found {
						serviceID = before
					}
				} else if m, ok := req.(map[string]interface{}); ok {
					if params, ok := m["params"].(map[string]interface{}); ok {
						if name, ok := params["name"].(string); ok {
							if before, _, found := strings.Cut(name, "."); found {
								serviceID = before
							}
						}
					}
				}
			}`, 1)

    modified = strings.Replace(modified, `			// Fallback to method-based extraction if serviceID not yet found
			if serviceID == "" {
				// Extract serviceID from the method. Assuming the format is "service.method".
				// Optimization: Use strings.Cut to avoid allocating a slice.
				if before, _, found := strings.Cut(method, "."); found {
					serviceID = before
				}
			}`, `			// Fallback to method-based extraction if serviceID not yet found
			if serviceID == "" {
				// We don't want to fallback if it's a known service-prefixed parameter method
				// because an attacker could provide an invalid type and bypass auth.
				if method != consts.MethodToolsCall && method != consts.MethodPromptsGet {
					// Extract serviceID from the method. Assuming the format is "service.method".
					// Optimization: Use strings.Cut to avoid allocating a slice.
					if before, _, found := strings.Cut(method, "."); found {
						serviceID = before
					}
				} else {
					// Sentinel Security Update: Secure By Design, Fail Closed.
					// If we reach here for tools/call or prompts/get, we couldn't extract a valid
					// serviceID from the payload. This either means the payload is malformed,
					// or it's an exploit attempt. We MUST fail closed.
					// By setting a dummy invalid service ID, we force authentication to fail.
					serviceID = "__invalid_missing_service_id__"
				}
			}`, 1)

    err = ioutil.WriteFile("server/pkg/middleware/auth.go", []byte(modified), 0644)
    if err != nil {
        log.Fatal(err)
    }
}
