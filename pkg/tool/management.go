/*
 * Copyright 2025 Author(s) of MCP Any
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
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	xsync "github.com/puzpuzpuz/xsync/v4"
	"github.com/mcpany/core/pkg/bus"
	"github.com/mcpany/core/pkg/logging"
	"github.com/mcpany/core/pkg/util"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// MCPServerProvider defines an interface for components that can provide an
// instance of an *mcp.Server. This is used to decouple the ToolManager from the
// concrete server implementation.
type MCPServerProvider interface {
	Server() *mcp.Server
}

// ToolManager is a thread-safe manager for registering, retrieving, and
// executing tools. It also handles the registration of tools with an MCP server,
// making them available for remote execution.
type ToolExecutionMiddleware interface {
	Execute(ctx context.Context, req *ExecutionRequest, next ToolExecutionFunc) (any, error)
}

type ToolManager struct {
	tools       *xsync.Map[string, Tool]
	serviceInfo *xsync.Map[string, *ServiceInfo]
	mcpServer   MCPServerProvider
	bus         *bus.BusProvider
	mu          sync.RWMutex
	middlewares []ToolExecutionMiddleware
}

// NewToolManager creates and returns a new, empty ToolManager.
func NewToolManager(bus *bus.BusProvider) *ToolManager {
	return &ToolManager{
		bus:         bus,
		tools:       xsync.NewMap[string, Tool](),
		serviceInfo: xsync.NewMap[string, *ServiceInfo](),
	}
}

// AddMiddleware adds a middleware to the tool manager.
func (tm *ToolManager) AddMiddleware(middleware ToolExecutionMiddleware) {
	tm.middlewares = append(tm.middlewares, middleware)
}

// SetMCPServer provides the ToolManager with a reference to the MCP server.
// This is necessary for registering tool handlers with the server.
func (tm *ToolManager) SetMCPServer(mcpServer MCPServerProvider) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.mcpServer = mcpServer
}

// ExecuteTool finds a tool by its name and executes it with the provided
// request context and inputs.
//
// ctx is the context for the tool execution.
// req contains the name of the tool and its inputs.
// It returns the result of the execution or an error if the tool is not found
// or if the execution fails.
func (tm *ToolManager) ExecuteTool(ctx context.Context, req *ExecutionRequest) (any, error) {
	log := logging.GetLogger().With("toolName", req.ToolName)
	log.Debug("Executing tool")

	execute := func(ctx context.Context, req *ExecutionRequest) (any, error) {
		t, ok := tm.GetTool(req.ToolName)
		if !ok {
			log.Error("Tool not found")
			return nil, ErrToolNotFound
		}
		ctx = NewContextWithTool(ctx, t)
		result, err := t.Execute(ctx, req)
		if err != nil {
			log.Error("Tool execution failed", "error", err)
		} else {
			log.Debug("Tool execution successful")
		}
		return result, err
	}

	chain := execute
	for i := len(tm.middlewares) - 1; i >= 0; i-- {
		m := tm.middlewares[i]
		chain = func(next ToolExecutionFunc) ToolExecutionFunc {
			return func(ctx context.Context, req *ExecutionRequest) (any, error) {
				log.Debug("Executing middleware", "middleware", i)
				return m.Execute(ctx, req, next)
			}
		}(chain)
	}

	result, err := chain(ctx, req)
	if err != nil {
		log.Error("Tool execution chain failed", "error", err)
	} else {
		log.Debug("Tool execution chain successful")
	}
	return result, err
}

// AddServiceInfo stores metadata about a service, indexed by its ID.
//
// serviceID is the unique identifier for the service.
// info is the ServiceInfo struct containing the service's metadata.
func (tm *ToolManager) AddServiceInfo(serviceID string, info *ServiceInfo) {
	tm.serviceInfo.Store(serviceID, info)
}

// GetServiceInfo retrieves the metadata for a service by its ID.
//
// serviceID is the unique identifier for the service.
// It returns the ServiceInfo and a boolean indicating whether the service was found.
func (tm *ToolManager) GetServiceInfo(serviceID string) (*ServiceInfo, bool) {
	info, ok := tm.serviceInfo.Load(serviceID)
	if !ok {
		return nil, false
	}
	return info, true
}

// AddTool registers a new tool with the manager. It generates a unique tool ID
// and, if an MCP server is configured, registers a handler for the tool with
// the server.
//
// tool is the tool to be added.
// It returns an error if the tool ID cannot be generated.
func (tm *ToolManager) AddTool(tool Tool) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	sanitizedToolName, err := util.SanitizeToolName(tool.Tool().GetName())
	if err != nil {
		logging.GetLogger().
			Error("Failed to sanitize tool name", "serviceID", tool.Tool().GetServiceId(), "toolName", tool.Tool().GetName(), "error", err)
		return fmt.Errorf("failed to sanitize tool name: %w", err)
	}
	toolID := tool.Tool().GetServiceId() + "." + sanitizedToolName
	log := logging.GetLogger().With("toolID", toolID)
	log.Debug("Adding tool to ToolManager")
	tm.tools.Store(toolID, tool)

	if tm.mcpServer != nil {
		mcpTool, err := ConvertProtoToMCPTool(tool.Tool())
		if err != nil {
			return fmt.Errorf("failed to convert proto tool to mcp tool: %w", err)
		}

		if tool.Tool().GetInputSchema() != nil {
			b, err := tool.Tool().GetInputSchema().MarshalJSON()
			if err != nil {
				return fmt.Errorf("failed to marshal input schema: %w", err)
			}
			if err := json.Unmarshal(b, &mcpTool.InputSchema); err != nil {
				return fmt.Errorf("failed to unmarshal input schema: %w", err)
			}
		}

		if tool.Tool().GetOutputSchema() != nil {
			b, err := tool.Tool().GetOutputSchema().MarshalJSON()
			if err != nil {
				return fmt.Errorf("failed to marshal output schema: %w", err)
			}
			if err := json.Unmarshal(b, &mcpTool.OutputSchema); err != nil {
				return fmt.Errorf("failed to unmarshal output schema: %w", err)
			}
		}

		log.Info(
			"Registering tool with MCP server",
			"toolName",
			mcpTool.Name,
			"tool",
			mcpTool,
		)

		handler := func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			logging.GetLogger().Info("Queueing tool execution", "toolName", req.Params.Name)

			correlationID := uuid.New().String()
			resultChan := make(chan *bus.ToolExecutionResult, 1)

			resultBus := bus.GetBus[*bus.ToolExecutionResult](tm.bus, "tool_execution_results")
			unsubscribe := resultBus.SubscribeOnce(
				ctx,
				correlationID,
				func(result *bus.ToolExecutionResult) {
					resultChan <- result
				},
			)
			defer unsubscribe()

			requestBus := bus.GetBus[*bus.ToolExecutionRequest](tm.bus, "tool_execution_requests")
			execReq := &bus.ToolExecutionRequest{
				Context:    ctx,
				ToolName:   req.Params.Name,
				ToolInputs: req.Params.Arguments,
			}
			execReq.SetCorrelationID(correlationID)
			requestBus.Publish(ctx, "request", execReq)

			select {
			case result := <-resultChan:
				if result.Error != nil {
					return nil, fmt.Errorf(
						"error executing tool %s: %w",
						req.Params.Name,
						result.Error,
					)
				}

				jsonResult, err := json.Marshal(result.Result)
				if err != nil {
					return nil, fmt.Errorf("failed to marshal tool result: %w", err)
				}

				return &mcp.CallToolResult{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: string(jsonResult),
						},
					},
				}, nil
			case <-ctx.Done():
				return nil, fmt.Errorf("context deadline exceeded while waiting for tool execution")
			case <-time.After(60 * time.Second): // Safety timeout
				return nil, fmt.Errorf(
					"timed out waiting for tool execution result for tool %s",
					req.Params.Name,
				)
			}
		}
		tm.mcpServer.Server().AddTool(mcpTool, handler)
	}
	return nil
}

// GetTool retrieves a tool from the manager by its fully qualified name.
//
// toolName is the name of the tool to retrieve.
// It returns the tool and a boolean indicating whether the tool was found.
func (tm *ToolManager) GetTool(toolName string) (Tool, bool) {
	tool, ok := tm.tools.Load(toolName)
	if !ok {
		return nil, false
	}
	return tool, true
}

// ListTools returns a slice containing all the tools currently registered with
// the manager.
func (tm *ToolManager) ListTools() []Tool {
	var tools []Tool
	tm.tools.Range(func(key string, value Tool) bool {
		tools = append(tools, value)
		return true
	})
	return tools
}

// ClearToolsForService removes all tools associated with a given service key from
// the manager. This is useful when a service is being re-registered or
// unregistered.
//
// serviceID is the unique identifier for the service whose tools should be
// cleared.
func (tm *ToolManager) ClearToolsForService(serviceID string) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	log := logging.GetLogger().With("serviceID", serviceID)
	log.Debug("Clearing existing tools for serviceID before reload/overwrite.")
	deletedCount := 0
	tm.tools.Range(func(key string, value Tool) bool {
		if value.Tool().GetServiceId() == serviceID {
			tm.tools.Delete(key)
			deletedCount++
		}
		return true
	})
	log.Debug("Cleared tools for serviceID", "count", deletedCount)
}
