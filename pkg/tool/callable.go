/*
 * Copyright 2024 Author(s) of MCP Any
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

package tool

import (
	"context"

	configv1 "github.com/mcpany/core/proto/config/v1"
)

// CallableTool implements the Tool interface for a tool that is executed by a
// Callable.
type CallableTool struct {
	*baseTool
}

// NewCallableTool creates a new CallableTool.
func NewCallableTool(toolDef *configv1.ToolDefinition, serviceConfig *configv1.UpstreamServiceConfig, callable Callable) (*CallableTool, error) {
	base, err := newBaseTool(toolDef, serviceConfig, callable)
	if err != nil {
		return nil, err
	}
	return &CallableTool{base}, nil
}

// Execute handles the execution of the tool.
func (t *CallableTool) Execute(ctx context.Context, req *ExecutionRequest) (any, error) {
	return t.callable.Call(ctx, req)
}
