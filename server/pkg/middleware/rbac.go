// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/mcpany/core/pkg/auth"
	"github.com/mcpany/core/pkg/consts"
	"github.com/mcpany/core/pkg/tool"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// RBACMiddleware creates a middleware that enforces Role-Based Access Control.
// It checks if the authenticated user has the necessary permissions to access
// the requested service or tool.
func RBACMiddleware(toolManager tool.ManagerInterface, profileDefinitions []*configv1.ProfileDefinition) mcp.Middleware {
	// Index profile definitions for fast lookup
	profileDefs := make(map[string]*configv1.ProfileDefinition)
	for _, pd := range profileDefinitions {
		profileDefs[pd.GetName()] = pd
	}

	return func(next mcp.MethodHandler) mcp.MethodHandler {
		return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			// 1. Identify Service
			var serviceID string

			// Special handling for tool calls
			if method == consts.MethodToolsCall {
				if r, ok := req.(*mcp.CallToolRequest); ok {
					// We expect tool names to be prefixed with the service ID (e.g. "service.tool")
					parts := strings.SplitN(r.Params.Name, consts.ToolNameServiceSeparator, 2)
					if len(parts) >= 2 {
						serviceID = parts[0]
					}
				}
			}

			// Fallback to method-based extraction if serviceID not yet found
			if serviceID == "" {
				// Extract serviceID from the method. Assuming the format is "service.method".
				parts := strings.SplitN(method, ".", 2)
				if len(parts) >= 2 {
					serviceID = parts[0]
				}
			}

			if serviceID == "" {
				// Not a service-specific call, or we couldn't determine the service.
				// For now, we allow global methods (like list_tools) to proceed,
				// or let downstream handle it.
				return next(ctx, method, req)
			}

			// 2. Check if User is present (Authenticated via User/RBAC)
			userID, hasUser := auth.UserFromContext(ctx)
			if !hasUser {
				// If no user context is present, it means either:
				// - Authentication was skipped (not configured)
				// - Authentication was done via legacy/simple methods (Global API Key, Service Auth)
				// In these cases, we assume "Simple Mode" and bypass strict RBAC checks.
				return next(ctx, method, req)
			}

			// 3. Get Service Config
			info, ok := toolManager.GetServiceInfo(serviceID)
			if !ok {
				// Service not found. Pass through, let the router handle 404/method not found.
				return next(ctx, method, req)
			}

			serviceConfig := info.Config
			serviceProfiles := serviceConfig.GetProfiles()

			// If the service is not assigned to any profile, we default to ALLOW for authenticated users.
			// This behavior ensures that adding RBAC doesn't break existing setups where profiles aren't used.
			if len(serviceProfiles) == 0 {
				return next(ctx, method, req)
			}

			// 4. Check User Access to Service Profiles
			userProfileIDs, _ := auth.ProfileIDsFromContext(ctx)
			userRoles, _ := auth.RolesFromContext(ctx)

			allowed := false
			for _, sp := range serviceProfiles {
				// Check if user has this profile assigned directly by ID
				if slices.Contains(userProfileIDs, sp.GetId()) {
					allowed = true
					break
				}

				// Check if user has required roles for this profile (via ProfileDefinition)
				if def, ok := profileDefs[sp.GetName()]; ok {
					reqRoles := def.GetRequiredRoles()
					if len(reqRoles) > 0 {
						// "If user has one of these roles, they get access."
						for _, r := range reqRoles {
							if slices.Contains(userRoles, r) {
								allowed = true
								break
							}
						}
					}
				}
				if allowed {
					break
				}
			}

			if !allowed {
				return nil, fmt.Errorf("forbidden: user %s does not have access to service %s", userID, serviceID)
			}

			return next(ctx, method, req)
		}
	}
}
