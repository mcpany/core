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

package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/mcpany/core/pkg/auth"
	"github.com/mcpany/core/pkg/client"
	"github.com/mcpany/core/pkg/logging"
	"github.com/mcpany/core/pkg/prompt"
	"github.com/mcpany/core/pkg/resource"
	"github.com/mcpany/core/pkg/tool"
	"github.com/mcpany/core/pkg/upstream"
	"github.com/mcpany/core/pkg/util"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"google.golang.org/protobuf/proto"
)

var (
	newClientImplForTesting func(client *mcp.Client, stdioConfig *configv1.McpStdioConnection, httpAddress string, httpClient *http.Client) client.MCPClient
	newClientForTesting     func(impl *mcp.Implementation) *mcp.Client
	connectForTesting       func(client *mcp.Client, ctx context.Context, transport mcp.Transport, roots []mcp.Root) (ClientSession, error)
)

// ClientSession defines an interface that abstracts the capabilities of an
// mcp.ClientSession. This is used primarily for testing, allowing mock sessions
// to be injected.
type ClientSession interface {
	ListTools(ctx context.Context, params *mcp.ListToolsParams) (*mcp.ListToolsResult, error)
	ListPrompts(ctx context.Context, params *mcp.ListPromptsParams) (*mcp.ListPromptsResult, error)
	ListResources(ctx context.Context, params *mcp.ListResourcesParams) (*mcp.ListResourcesResult, error)
	GetPrompt(ctx context.Context, params *mcp.GetPromptParams) (*mcp.GetPromptResult, error)
	ReadResource(ctx context.Context, params *mcp.ReadResourceParams) (*mcp.ReadResourceResult, error)
	CallTool(ctx context.Context, params *mcp.CallToolParams) (*mcp.CallToolResult, error)
	Close() error
}

// SetNewClientImplForTesting provides a hook for injecting a mock MCP client
// implementation during tests. This should only be used for testing purposes.
func SetNewClientImplForTesting(f func(client *mcp.Client, stdioConfig *configv1.McpStdioConnection, httpAddress string, httpClient *http.Client) client.MCPClient) {
	newClientImplForTesting = f
}

// SetNewClientForTesting provides a hook for injecting a mock mcp.Client
// during tests. This should only be used for testing purposes.
func SetNewClientForTesting(f func(impl *mcp.Implementation) *mcp.Client) {
	newClientForTesting = f
}

// SetConnectForTesting provides a hook for injecting a mock mcp.Client.Connect
// function during tests. This should only be used for testing purposes.
func SetConnectForTesting(f func(client *mcp.Client, ctx context.Context, transport mcp.Transport, roots []mcp.Root) (ClientSession, error)) {
	connectForTesting = f
}

// MCPUpstream implements the upstream.Upstream interface for services that are
// themselves MCP-compliant. It connects to the downstream MCP service, discovers
// its tools, prompts, and resources, and registers them with the current server,
// effectively acting as a proxy or aggregator.
type MCPUpstream struct{}

// Shutdown is a no-op for the MCP upstream, as the connections it manages are
// transient and established on a per-call basis. There are no persistent
// connections to tear down.
func (u *MCPUpstream) Shutdown(ctx context.Context) error {
	return nil
}

// NewMCPUpstream creates a new instance of MCPUpstream.
func NewMCPUpstream() upstream.Upstream {
	return &MCPUpstream{}
}

// mcpPrompt is a wrapper around the standard mcp.Prompt that associates it with
// a specific service and provides the necessary connection details for execution.
type mcpPrompt struct {
	mcpPrompt *mcp.Prompt
	service   string
	*mcpConnection
}

// Prompt returns the underlying *mcp.Prompt definition.
func (p *mcpPrompt) Prompt() *mcp.Prompt {
	return p.mcpPrompt
}

// Service returns the ID of the service that this prompt belongs to.
func (p *mcpPrompt) Service() string {
	return p.service
}

// Get executes the prompt by establishing a session with the downstream MCP
// service and calling its GetPrompt method.
func (p *mcpPrompt) Get(ctx context.Context, args json.RawMessage) (*mcp.GetPromptResult, error) {
	var arguments map[string]string
	if args != nil {
		var genericArgs map[string]any
		if err := json.Unmarshal(args, &genericArgs); err != nil {
			return nil, fmt.Errorf("failed to unmarshal prompt arguments to generic map: %w", err)
		}
		arguments = make(map[string]string)
		for k, v := range genericArgs {
			if s, ok := v.(string); ok {
				arguments[k] = s
			} else {
				arguments[k] = fmt.Sprintf("%v", v)
			}
		}
	}

	var result *mcp.GetPromptResult
	err := p.withMCPClientSession(ctx, func(cs ClientSession) error {
		var err error
		result, err = cs.GetPrompt(ctx, &mcp.GetPromptParams{
			Name:      p.mcpPrompt.Name,
			Arguments: arguments,
		})
		return err
	})
	return result, err
}

// mcpResource is a wrapper around the standard mcp.Resource that associates it
// with a specific service and provides the necessary connection details for
// interaction.
type mcpResource struct {
	mcpResource *mcp.Resource
	service     string
	*mcpConnection
}

// Resource returns the underlying *mcp.Resource definition.
func (r *mcpResource) Resource() *mcp.Resource {
	return r.mcpResource
}

// Service returns the ID of the service that this resource belongs to.
func (r *mcpResource) Service() string {
	return r.service
}

// Read retrieves the content of the resource by establishing a session with the
// downstream MCP service and calling its ReadResource method.
func (r *mcpResource) Read(ctx context.Context) (*mcp.ReadResourceResult, error) {
	var result *mcp.ReadResourceResult
	err := r.withMCPClientSession(ctx, func(cs ClientSession) error {
		var err error
		result, err = cs.ReadResource(ctx, &mcp.ReadResourceParams{
			URI: r.mcpResource.URI,
		})
		return err
	})
	return result, err
}

// Subscribe is not yet implemented for MCP resources. It returns an error
// indicating that this functionality is not available.
func (r *mcpResource) Subscribe(ctx context.Context) error {
	return fmt.Errorf("subscribing to resources on mcp upstreams is not yet implemented")
}

// Register handles the registration of another MCP service as an upstream. It
// determines the connection type (stdio or HTTP), connects to the downstream
// service, lists its available tools, prompts, and resources, and registers
// them with the appropriate managers.
func (u *MCPUpstream) Register(
	ctx context.Context,
	serviceConfig *configv1.UpstreamServiceConfig,
	toolManager tool.ManagerInterface,
	promptManager prompt.ManagerInterface,
	resourceManager resource.ManagerInterface,
	isReload bool,
) (string, []*configv1.ToolDefinition, []*configv1.ResourceDefinition, error) {
	log := logging.GetLogger()
	serviceID, err := util.SanitizeServiceName(serviceConfig.GetName())
	if err != nil {
		return "", nil, nil, err
	}

	mcpService := serviceConfig.GetMcpService()
	if mcpService == nil {
		return "", nil, nil, fmt.Errorf("mcp service config is nil")
	}

	info := &tool.ServiceInfo{
		Name:   serviceConfig.GetName(),
		Config: serviceConfig,
	}
	toolManager.AddServiceInfo(serviceID, info)

	var discoveredTools []*configv1.ToolDefinition
	var discoveredResources []*configv1.ResourceDefinition
	switch mcpService.WhichConnectionType() {
	case configv1.McpUpstreamService_StdioConnection_case:
		discoveredTools, discoveredResources, err = u.createAndRegisterMCPItemsFromStdio(ctx, serviceID, mcpService.GetStdioConnection(), toolManager, promptManager, resourceManager, isReload, serviceConfig)
		if err != nil {
			return "", nil, nil, err
		}
	case configv1.McpUpstreamService_HttpConnection_case:
		discoveredTools, discoveredResources, err = u.createAndRegisterMCPItemsFromStreamableHTTP(ctx, serviceID, mcpService.GetHttpConnection(), toolManager, promptManager, resourceManager, isReload, serviceConfig)
		if err != nil {
			return "", nil, nil, err
		}
	default:
		return "", nil, nil, fmt.Errorf("MCPService definition requires either stdio_connection or http_connection")
	}

	log.Info("Registered MCP service", "serviceID", serviceID, "toolsAdded", len(discoveredTools))
	return serviceID, discoveredTools, discoveredResources, nil
}

// mcpConnection holds the necessary information to connect to a downstream MCP
// service, whether it's via stdio or HTTP. It also implements the
// client.MCPClient interface, allowing it to be used as a proxy.
type mcpConnection struct {
	client      *mcp.Client
	stdioConfig *configv1.McpStdioConnection
	httpAddress string
	httpClient  *http.Client
}

// withMCPClientSession is a helper function that abstracts the process of
// establishing a connection to the downstream MCP service, executing a function
// with the active session, and ensuring the session is closed afterward.
func (c *mcpConnection) withMCPClientSession(ctx context.Context, f func(cs ClientSession) error) error {
	var transport mcp.Transport
	switch {
	case c.stdioConfig != nil:
		image := c.stdioConfig.GetContainerImage()
		if image != "" {
			if util.IsDockerSocketAccessible() {
				transport = &DockerTransport{
					StdioConfig: c.stdioConfig,
				}
			} else {
				return fmt.Errorf("docker socket not accessible, but container_image is specified")
			}
		} else {
			cmd := buildCommandFromStdioConfig(c.stdioConfig)
			transport = &mcp.CommandTransport{
				Command: cmd,
			}
		}
	case c.httpAddress != "":
		transport = &mcp.StreamableClientTransport{
			Endpoint:   c.httpAddress,
			HTTPClient: c.httpClient,
		}
	default:
		return fmt.Errorf("mcp transport is not configured")
	}

	var cs ClientSession
	var err error
	if connectForTesting != nil {
		cs, err = connectForTesting(c.client, ctx, transport, nil)
	} else {
		cs, err = c.client.Connect(ctx, transport, nil)
	}
	if err != nil {
		return fmt.Errorf("failed to connect to MCP server: %w", err)
	}
	defer cs.Close()

	return f(cs)
}

// CallTool executes a tool on the downstream MCP service by establishing a
// session and forwarding the tool call.
func (c *mcpConnection) CallTool(ctx context.Context, params *mcp.CallToolParams) (*mcp.CallToolResult, error) {
	var result *mcp.CallToolResult
	err := c.withMCPClientSession(ctx, func(cs ClientSession) error {
		var err error
		result, err = cs.CallTool(ctx, params)
		return err
	})
	return result, err
}

// buildCommandFromStdioConfig constructs an *exec.Cmd from an McpStdioConnection
// configuration. It combines the setup commands and the main command into a
// single shell script to be executed.
func buildCommandFromStdioConfig(stdio *configv1.McpStdioConnection) *exec.Cmd {
	command := stdio.GetCommand()
	args := stdio.GetArgs()

	// If the command is 'docker', handle it directly, including sudo if needed.
	if command == "docker" {
		useSudo := os.Getenv("USE_SUDO_FOR_DOCKER") == "1"
		if useSudo {
			fullArgs := append([]string{command}, args...)
			return exec.Command("sudo", fullArgs...) //nolint:gosec // controlled execution
		}
		return exec.Command(command, args...) //nolint:gosec // controlled execution
	}

	// Combine all commands into a single script.
	var scriptCommands []string
	scriptCommands = append(scriptCommands, stdio.GetSetupCommands()...)

	// Add the main command. `exec` is used to replace the shell process with the main command.
	mainCommandParts := []string{"exec", command}
	mainCommandParts = append(mainCommandParts, args...)
	scriptCommands = append(scriptCommands, strings.Join(mainCommandParts, " "))

	script := strings.Join(scriptCommands, " && ")

	// run the script directly on the host.
	cmd := exec.Command("/bin/sh", "-c", script)
	cmd.Dir = stdio.GetWorkingDirectory()
	return cmd
}

// createAndRegisterMCPItemsFromStdio handles the registration of an MCP service
// that is connected via standard I/O (e.g., a local command or a Docker
// container). It establishes the connection, discovers the service's
// capabilities, and registers them.
func (u *MCPUpstream) createAndRegisterMCPItemsFromStdio(
	ctx context.Context,
	serviceID string,
	stdio *configv1.McpStdioConnection,
	toolManager tool.ManagerInterface,
	promptManager prompt.ManagerInterface,
	resourceManager resource.ManagerInterface,
	isReload bool,
	serviceConfig *configv1.UpstreamServiceConfig,
) ([]*configv1.ToolDefinition, []*configv1.ResourceDefinition, error) {
	if stdio == nil {
		return nil, nil, fmt.Errorf("stdio connection config is nil")
	}

	var transport mcp.Transport
	image := stdio.GetContainerImage()
	if image != "" {
		if util.IsDockerSocketAccessible() {
			transport = &DockerTransport{
				StdioConfig: stdio,
			}
		} else {
			return nil, nil, fmt.Errorf("docker socket not accessible, but container_image is specified")
		}
	} else {
		cmd := buildCommandFromStdioConfig(stdio)
		transport = &mcp.CommandTransport{
			Command: cmd,
		}
	}
	var mcpSdkClient *mcp.Client
	if newClientForTesting != nil {
		mcpSdkClient = newClientForTesting(&mcp.Implementation{
			Name:    "mcpany",
			Version: "0.1.0",
		})
	} else {
		mcpSdkClient = mcp.NewClient(&mcp.Implementation{
			Name:    "mcpany",
			Version: "0.1.0",
		}, nil)
	}

	var cs ClientSession
	var err error
	if connectForTesting != nil {
		cs, err = connectForTesting(mcpSdkClient, ctx, transport, nil)
	} else {
		cs, err = mcpSdkClient.Connect(ctx, transport, nil)
	}
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to MCP service: %w", err)
	}
	defer cs.Close()

	// Register tools
	listToolsResult, err := cs.ListTools(ctx, &mcp.ListToolsParams{})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list tools from MCP service: %w", err)
	}

	var mcpClient client.MCPClient
	if newClientImplForTesting != nil {
		mcpClient = newClientImplForTesting(mcpSdkClient, stdio, "", nil)
	} else {
		mcpClient = &mcpConnection{
			client:      mcpSdkClient,
			stdioConfig: stdio,
		}
	}

	mcpService := serviceConfig.GetMcpService()
	configToolDefs := mcpService.GetTools()
	calls := mcpService.GetCalls()
	configToolMap := make(map[string]*configv1.ToolDefinition)
	for _, toolDef := range configToolDefs {
		configToolMap[toolDef.GetName()] = toolDef
	}

	discoveredTools := make([]*configv1.ToolDefinition, 0, len(listToolsResult.Tools))
	for _, mcpSDKTool := range listToolsResult.Tools {
		var callDef *configv1.MCPCallDefinition
		if configTool, ok := configToolMap[mcpSDKTool.Name]; ok {
			if configTool.GetDisable() {
				logging.GetLogger().Info("Skipping disabled tool", "toolName", mcpSDKTool.Name)
				continue
			}
			if call, callOk := calls[configTool.GetCallId()]; callOk {
				callDef = call
			} else {
				logging.GetLogger().Warn("Call definition not found for tool", "call_id", configTool.GetCallId(), "tool_name", mcpSDKTool.Name)
				callDef = &configv1.MCPCallDefinition{}
			}
		} else {
			callDef = &configv1.MCPCallDefinition{}
		}

		pbTool, err := tool.ConvertMCPToolToProto(mcpSDKTool)
		if err != nil {
			logging.GetLogger().Error("Failed to convert mcp tool to proto", "error", err)
			continue
		}
		pbTool.SetServiceId(serviceID)

		newTool := tool.NewMCPTool(
			pbTool,
			mcpClient,
			callDef,
		)
		if err := toolManager.AddTool(newTool); err != nil {
			logging.GetLogger().Error("Failed to add tool", "error", err)
			continue
		}
		discoveredTools = append(discoveredTools, configv1.ToolDefinition_builder{
			Name:        proto.String(mcpSDKTool.Name),
			Description: proto.String(mcpSDKTool.Description),
		}.Build())
	}

	// Register prompts
	listPromptsResult, err := cs.ListPrompts(ctx, &mcp.ListPromptsParams{})

	configPromptMap := make(map[string]*configv1.PromptDefinition)
	for _, p := range mcpService.GetPrompts() {
		configPromptMap[p.GetName()] = p
	}

	if err != nil {
		// Do not fail if prompts are not supported
		logging.GetLogger().Warn("Failed to list prompts from MCP service", "error", err)
	} else {
		for _, mcpSDKPrompt := range listPromptsResult.Prompts {
			if configPrompt, ok := configPromptMap[mcpSDKPrompt.Name]; ok {
				if configPrompt.GetDisable() {
					logging.GetLogger().Info("Skipping disabled prompt (auto-discovered)", "promptName", mcpSDKPrompt.Name)
					continue
				}
			}
			promptManager.AddPrompt(&mcpPrompt{
				mcpPrompt: mcpSDKPrompt,
				service:   serviceID,
				mcpConnection: &mcpConnection{
					client:      mcpSdkClient,
					stdioConfig: stdio,
				},
			})
		}
	}

	for _, promptDef := range mcpService.GetPrompts() {
		if promptDef.GetDisable() {
			logging.GetLogger().Info("Skipping disabled prompt (config)", "promptName", promptDef.GetName())
			continue
		}
		newPrompt := prompt.NewTemplatedPrompt(promptDef, serviceID)
		promptManager.AddPrompt(newPrompt)
	}

	// Register resources
	listResourcesResult, err := cs.ListResources(ctx, &mcp.ListResourcesParams{})
	if err != nil {
		// Do not fail if resources are not supported
		logging.GetLogger().Warn("Failed to list resources from MCP service", "error", err)
		return discoveredTools, nil, nil
	}

	configResourceMap := make(map[string]*configv1.ResourceDefinition)
	for _, r := range mcpService.GetResources() {
		configResourceMap[r.GetName()] = r
	}

	discoveredResources := make([]*configv1.ResourceDefinition, 0, len(listResourcesResult.Resources))
	for _, mcpSDKResource := range listResourcesResult.Resources {
		if configResource, ok := configResourceMap[mcpSDKResource.Name]; ok {
			if configResource.GetDisable() {
				logging.GetLogger().Info("Skipping disabled resource (auto-discovered)", "resourceName", mcpSDKResource.Name)
				continue
			}
		}
		resourceManager.AddResource(&mcpResource{
			mcpResource: mcpSDKResource,
			service:     serviceID,
			mcpConnection: &mcpConnection{
				client:      mcpSdkClient,
				stdioConfig: stdio,
			},
		})
		discoveredResources = append(discoveredResources, convertMCPResourceToProto(mcpSDKResource))
	}

	log := logging.GetLogger()
	callIDToName := make(map[string]string)
	for _, d := range configToolDefs {
		callIDToName[d.GetCallId()] = d.GetName()
	}
	for _, resourceDef := range mcpService.GetResources() {
		if resourceDef.GetDisable() {
			log.Info("Skipping disabled resource (config)", "resourceName", resourceDef.GetName())
			continue
		}
		if resourceDef.GetDynamic() != nil {
			call := resourceDef.GetDynamic().GetMcpCall()
			if call == nil {
				continue
			}
			toolName, ok := callIDToName[call.GetId()]
			if !ok {
				log.Error("tool not found for dynamic resource", "call_id", call.GetId())
				continue
			}
			sanitizedToolName, err := util.SanitizeToolName(toolName)
			if err != nil {
				log.Error("Failed to sanitize tool name", "error", err)
				continue
			}
			tool, ok := toolManager.GetTool(serviceID + "." + sanitizedToolName)
			if !ok {
				log.Error("Tool not found for dynamic resource", "toolName", toolName)
				continue
			}
			dynamicResource, err := resource.NewDynamicResource(resourceDef, tool)
			if err != nil {
				log.Error("Failed to create dynamic resource", "error", err)
				continue
			}
			resourceManager.AddResource(dynamicResource)
		}
	}

	return discoveredTools, discoveredResources, nil
}

// createAndRegisterMCPItemsFromStreamableHTTP handles the registration of an MCP
// service that is connected via HTTP. It establishes the connection, discovers
// the service's capabilities, and registers them.
func (u *MCPUpstream) createAndRegisterMCPItemsFromStreamableHTTP(
	ctx context.Context,
	serviceID string,
	httpConnection *configv1.McpStreamableHttpConnection,
	toolManager tool.ManagerInterface,
	promptManager prompt.ManagerInterface,
	resourceManager resource.ManagerInterface,
	isReload bool,
	serviceConfig *configv1.UpstreamServiceConfig,
) ([]*configv1.ToolDefinition, []*configv1.ResourceDefinition, error) {
	authenticator, err := auth.NewUpstreamAuthenticator(serviceConfig.GetUpstreamAuthentication())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create authenticator for MCP upstream: %w", err)
	}

	httpClient, err := util.NewHTTPClientWithTLS(httpConnection.GetTlsConfig())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create http client with tls config: %w", err)
	}
	httpClient.Transport = &authenticatedRoundTripper{
		authenticator: authenticator,
		base:          httpClient.Transport,
	}
	httpAddress := httpConnection.GetHttpAddress()
	transport := &mcp.StreamableClientTransport{
		Endpoint:   httpAddress,
		HTTPClient: httpClient,
	}
	var mcpSdkClient *mcp.Client
	if newClientForTesting != nil {
		mcpSdkClient = newClientForTesting(&mcp.Implementation{
			Name:    "mcpany",
			Version: "0.1.0",
		})
	} else {
		mcpSdkClient = mcp.NewClient(&mcp.Implementation{
			Name:    "mcpany",
			Version: "0.1.0",
		}, nil)
	}

	var cs ClientSession
	if connectForTesting != nil {
		cs, err = connectForTesting(mcpSdkClient, ctx, transport, nil)
	} else {
		cs, err = mcpSdkClient.Connect(ctx, transport, nil)
	}
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to MCP service: %w", err)
	}
	defer cs.Close()

	// Register tools
	listToolsResult, err := cs.ListTools(ctx, &mcp.ListToolsParams{})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list tools from MCP service: %w", err)
	}

	var mcpClient client.MCPClient
	if newClientImplForTesting != nil {
		mcpClient = newClientImplForTesting(mcpSdkClient, nil, httpAddress, httpClient)
	} else {
		mcpClient = &mcpConnection{
			client:      mcpSdkClient,
			httpAddress: httpAddress,
			httpClient:  httpClient,
		}
	}

	mcpService := serviceConfig.GetMcpService()
	configToolDefs := mcpService.GetTools()
	calls := mcpService.GetCalls()
	configToolMap := make(map[string]*configv1.ToolDefinition)
	for _, toolDef := range configToolDefs {
		configToolMap[toolDef.GetName()] = toolDef
	}

	discoveredTools := make([]*configv1.ToolDefinition, 0, len(listToolsResult.Tools))
	for _, mcpSDKTool := range listToolsResult.Tools {
		var callDef *configv1.MCPCallDefinition
		if configTool, ok := configToolMap[mcpSDKTool.Name]; ok {
			if configTool.GetDisable() {
				logging.GetLogger().Info("Skipping disabled tool", "toolName", mcpSDKTool.Name)
				continue
			}
			if call, callOk := calls[configTool.GetCallId()]; callOk {
				callDef = call
			} else {
				logging.GetLogger().Warn("Call definition not found for tool", "call_id", configTool.GetCallId(), "tool_name", mcpSDKTool.Name)
				callDef = &configv1.MCPCallDefinition{}
			}
		} else {
			callDef = &configv1.MCPCallDefinition{}
		}

		pbTool, err := tool.ConvertMCPToolToProto(mcpSDKTool)
		if err != nil {
			logging.GetLogger().Error("Failed to convert mcp tool to proto", "error", err)
			continue
		}
		pbTool.SetServiceId(serviceID)

		newTool := tool.NewMCPTool(
			pbTool,
			mcpClient,
			callDef,
		)
		if err := toolManager.AddTool(newTool); err != nil {
			logging.GetLogger().Error("Failed to add tool", "error", err)
			continue
		}
		discoveredTools = append(discoveredTools, configv1.ToolDefinition_builder{
			Name:        proto.String(mcpSDKTool.Name),
			Description: proto.String(mcpSDKTool.Description),
		}.Build())
	}

	// Register prompts
	listPromptsResult, err := cs.ListPrompts(ctx, &mcp.ListPromptsParams{})

	configPromptMap := make(map[string]*configv1.PromptDefinition)
	for _, p := range mcpService.GetPrompts() {
		configPromptMap[p.GetName()] = p
	}

	if err != nil {
		logging.GetLogger().Warn("Failed to list prompts from MCP service", "error", err)
	} else {
		for _, mcpSDKPrompt := range listPromptsResult.Prompts {
			if configPrompt, ok := configPromptMap[mcpSDKPrompt.Name]; ok {
				if configPrompt.GetDisable() {
					logging.GetLogger().Info("Skipping disabled prompt (auto-discovered)", "promptName", mcpSDKPrompt.Name)
					continue
				}
			}
			promptManager.AddPrompt(&mcpPrompt{
				mcpPrompt: mcpSDKPrompt,
				service:   serviceID,
				mcpConnection: &mcpConnection{
					client:      mcpSdkClient,
					httpAddress: httpAddress,
					httpClient:  httpClient,
				},
			})
		}
	}

	for _, promptDef := range mcpService.GetPrompts() {
		if promptDef.GetDisable() {
			logging.GetLogger().Info("Skipping disabled prompt (config)", "promptName", promptDef.GetName())
			continue
		}
		newPrompt := prompt.NewTemplatedPrompt(promptDef, serviceID)
		promptManager.AddPrompt(newPrompt)
	}

	// Register resources
	listResourcesResult, err := cs.ListResources(ctx, &mcp.ListResourcesParams{})
	if err != nil {
		logging.GetLogger().Warn("Failed to list resources from MCP service", "error", err)
		return discoveredTools, nil, nil
	}

	configResourceMap := make(map[string]*configv1.ResourceDefinition)
	for _, r := range mcpService.GetResources() {
		configResourceMap[r.GetName()] = r
	}

	discoveredResources := make([]*configv1.ResourceDefinition, 0, len(listResourcesResult.Resources))
	for _, mcpSDKResource := range listResourcesResult.Resources {
		if configResource, ok := configResourceMap[mcpSDKResource.Name]; ok {
			if configResource.GetDisable() {
				logging.GetLogger().Info("Skipping disabled resource (auto-discovered)", "resourceName", mcpSDKResource.Name)
				continue
			}
		}
		resourceManager.AddResource(&mcpResource{
			mcpResource: mcpSDKResource,
			service:     serviceID,
			mcpConnection: &mcpConnection{
				client:      mcpSdkClient,
				httpAddress: httpAddress,
				httpClient:  httpClient,
			},
		})
		discoveredResources = append(discoveredResources, convertMCPResourceToProto(mcpSDKResource))
	}

	log := logging.GetLogger()
	callIDToName := make(map[string]string)
	for _, d := range configToolDefs {
		callIDToName[d.GetCallId()] = d.GetName()
	}
	for _, resourceDef := range mcpService.GetResources() {
		if resourceDef.GetDisable() {
			log.Info("Skipping disabled resource (config)", "resourceName", resourceDef.GetName())
			continue
		}
		if resourceDef.GetDynamic() != nil {
			call := resourceDef.GetDynamic().GetMcpCall()
			if call == nil {
				continue
			}
			toolName, ok := callIDToName[call.GetId()]
			if !ok {
				log.Error("tool not found for dynamic resource", "call_id", call.GetId())
				continue
			}
			sanitizedToolName, err := util.SanitizeToolName(toolName)
			if err != nil {
				log.Error("Failed to sanitize tool name", "error", err)
				continue
			}
			tool, ok := toolManager.GetTool(serviceID + "." + sanitizedToolName)
			if !ok {
				log.Error("Tool not found for dynamic resource", "toolName", toolName)
				continue
			}
			dynamicResource, err := resource.NewDynamicResource(resourceDef, tool)
			if err != nil {
				log.Error("Failed to create dynamic resource", "error", err)
				continue
			}
			resourceManager.AddResource(dynamicResource)
		}
	}

	return discoveredTools, discoveredResources, nil
}

func convertMCPResourceToProto(resource *mcp.Resource) *configv1.ResourceDefinition {
	return configv1.ResourceDefinition_builder{
		Uri:         proto.String(resource.URI),
		Name:        proto.String(resource.Name),
		Title:       proto.String(resource.Title),
		Description: proto.String(resource.Description),
		MimeType:    proto.String(resource.MIMEType),
		Size:        proto.Int64(resource.Size),
	}.Build()
}

// authenticatedRoundTripper is an http.RoundTripper that wraps another
// RoundTripper and adds authentication to each request before it is sent.
type authenticatedRoundTripper struct {
	authenticator auth.UpstreamAuthenticator
	base          http.RoundTripper
}

// RoundTrip applies the configured authenticator to the request and then passes
// it to the base RoundTripper.
func (rt *authenticatedRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if rt.authenticator != nil {
		if err := rt.authenticator.Authenticate(req); err != nil {
			return nil, fmt.Errorf("failed to authenticate mcp upstream request: %w", err)
		}
	}
	base := rt.base
	if base == nil {
		base = http.DefaultTransport
	}
	return base.RoundTrip(req)
}

// StreamableHTTP implements the mcp.Transport interface for HTTP connections.
type StreamableHTTP struct {
	// Address is the HTTP address of the MCP service.
	Address string
	// Client is the HTTP client to use for the connection.
	Client *http.Client
}

// RoundTrip executes an HTTP request and returns the response.
func (t *StreamableHTTP) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.Client == nil {
		t.Client = http.DefaultClient
	}
	return t.Client.Do(req)
}
