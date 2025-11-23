// Copyright (C) 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package upstream

import (
	"context"

	"github.com/mcpany/core/pkg/prompt"
	"github.com/mcpany/core/pkg/resource"
	"github.com/mcpany/core/pkg/tool"
	configv1 "github.com/mcpany/core/proto/config/v1"
)

// Upstream defines the standard interface for all backend service integrations.
// Each implementation of this interface is responsible for discovering and
// registering its capabilities, such as tools, prompts, and resources, with the
// appropriate managers.
type Upstream interface {
	// Register inspects the upstream service defined by the serviceConfig,
	// discovers its capabilities, and registers them.
	//
	// ctx is the context for the registration process.
	// serviceConfig contains the configuration for the upstream service.
	// toolManager is the manager where discovered tools will be registered.
	// promptManager is the manager where discovered prompts will be registered.
	// resourceManager is the manager where discovered resources will be registered.
	// isReload indicates whether this is an initial registration or a reload of an
	// existing service.
	// It returns a unique service key, a list of discovered tool definitions, and
	// an error if registration fails.
	Register(
		ctx context.Context,
		serviceConfig *configv1.UpstreamServiceConfig,
		toolManager tool.ToolManagerInterface,
		promptManager prompt.PromptManagerInterface,
		resourceManager resource.ResourceManagerInterface,
		isReload bool,
	) (string, []*configv1.ToolDefinition, []*configv1.ResourceDefinition, error)
}
