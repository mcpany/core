/*
 * Copyright 2025 Author(s) of MCP-XY
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

package upstream

import (
	"context"

	"github.com/mcpxy/core/pkg/prompt"
	"github.com/mcpxy/core/pkg/resource"
	"github.com/mcpxy/core/pkg/tool"
	configv1 "github.com/mcpxy/core/proto/config/v1"
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
	// It returns a unique service ID, a list of discovered tool definitions, and
	// an error if registration fails.
	Register(
		ctx context.Context,
		serviceConfig *configv1.UpstreamServiceConfig,
		toolManager tool.ToolManagerInterface,
		promptManager prompt.PromptManagerInterface,
		resourceManager resource.ResourceManagerInterface,
		isReload bool,
		strategy configv1.GlobalSettings_ServiceNameStrategy,
	) (string, []*configv1.ToolDefinition, []*configv1.ResourceDefinition, error)
}
