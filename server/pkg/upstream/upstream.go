// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package upstream provides the upstream service implementation.
package upstream

import (
	"context"

	"github.com/mcpany/core/server/pkg/prompt"
	"github.com/mcpany/core/server/pkg/resource"
	"github.com/mcpany/core/server/pkg/tool"
	configv1 "github.com/mcpany/core/proto/config/v1"
)

// Upstream defines the standard interface for all backend service integrations.
// Each implementation of this interface is responsible for discovering and
// registering its capabilities, such as tools, prompts, and resources, with the
// appropriate managers.
type Upstream interface {
	// Shutdown gracefully terminates the upstream service.
	//
	// Parameters:
	//   - ctx: The context for the request.
	//
	// Returns:
	//   - error: An error if the operation fails.
	Shutdown(ctx context.Context) error

	// Register inspects the upstream service defined by the serviceConfig,
	// discovers its capabilities, and registers them.
	//
	// Parameters:
	//   - ctx: The context for the registration process.
	//   - serviceConfig: The configuration for the upstream service.
	//   - toolManager: The manager where discovered tools will be registered.
	//   - promptManager: The manager where discovered prompts will be registered.
	//   - resourceManager: The manager where discovered resources will be registered.
	//   - isReload: Indicates whether this is an initial registration or a reload.
	//
	// Returns:
	//   - string: A unique service key.
	//   - []*configv1.ToolDefinition: A list of discovered tool definitions.
	//   - []*configv1.ResourceDefinition: A list of discovered resource definitions.
	//   - error: An error if registration fails.
	Register(
		ctx context.Context,
		serviceConfig *configv1.UpstreamServiceConfig,
		toolManager tool.ManagerInterface,
		promptManager prompt.ManagerInterface,
		resourceManager resource.ManagerInterface,
		isReload bool,
	) (string, []*configv1.ToolDefinition, []*configv1.ResourceDefinition, error)
}

// HealthChecker is an optional interface that Upstreams can implement to provide
// runtime health status.
type HealthChecker interface {
	// CheckHealth performs a health check on the upstream service.
	//
	// Parameters:
	//   - ctx: The context for the health check.
	//
	// Returns:
	//   - error: nil if healthy, or an error if unhealthy.
	CheckHealth(ctx context.Context) error
}
