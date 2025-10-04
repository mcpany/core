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

package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/mcpxy/core/pkg/auth"
	"github.com/mcpxy/core/pkg/client"
	"github.com/mcpxy/core/pkg/logging"
	"github.com/mcpxy/core/pkg/prompt"
	"github.com/mcpxy/core/pkg/resource"
	"github.com/mcpxy/core/pkg/tool"
	"github.com/mcpxy/core/pkg/upstream"
	"github.com/mcpxy/core/pkg/util"
	configv1 "github.com/mcpxy/core/proto/config/v1"
	pb "github.com/mcpxy/core/proto/mcp_router/v1"
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

func SetConnectForTesting(f func(client *mcp.Client, ctx context.Context, transport mcp.Transport, roots []mcp.Root) (ClientSession, error)) {
	connectForTesting = f
}

// MCPUpstream implements the upstream.Upstream interface for services that are
// themselves MCP-compliant. It connects to the downstream MCP service, discovers
// its tools, prompts, and resources, and registers them with the current server.
type MCPUpstream struct{}

// NewMCPUpstream creates a new instance of MCPUpstream.
func NewMCPUpstream() upstream.Upstream {
	return &MCPUpstream{}
}

type mcpPrompt struct {
	mcpPrompt *mcp.Prompt
	service   string
	*mcpConnection
}

func (p *mcpPrompt) Prompt() *mcp.Prompt {
	return p.mcpPrompt
}

func (p *mcpPrompt) Service() string {
	return p.service
}

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

type mcpResource struct {
	mcpResource *mcp.Resource
	service     string
	*mcpConnection
}

func (r *mcpResource) Resource() *mcp.Resource {
	return r.mcpResource
}

func (r *mcpResource) Service() string {
	return r.service
}

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
	toolManager tool.ToolManagerInterface,
	promptManager prompt.PromptManagerInterface,
	resourceManager resource.ResourceManagerInterface,
	isReload bool,
) (string, []*configv1.ToolDefinition, error) {
	log := logging.GetLogger()
	serviceKey, err := util.GenerateServiceKey(serviceConfig.GetName())
	if err != nil {
		return "", nil, err
	}

	mcpService := serviceConfig.GetMcpService()
	if mcpService == nil {
		return "", nil, fmt.Errorf("mcp service config is nil")
	}

	info := &tool.ServiceInfo{
		Name:   serviceConfig.GetName(),
		Config: serviceConfig,
	}
	toolManager.AddServiceInfo(serviceKey, info)

	var discoveredTools []*configv1.ToolDefinition
	switch mcpService.WhichConnectionType() {
	case configv1.McpUpstreamService_StdioConnection_case:
		discoveredTools, err = u.createAndRegisterMCPItemsFromStdio(ctx, serviceKey, mcpService.GetStdioConnection(), toolManager, promptManager, resourceManager, isReload, serviceConfig)
		if err != nil {
			return "", nil, err
		}
	case configv1.McpUpstreamService_HttpConnection_case:
		discoveredTools, err = u.createAndRegisterMCPItemsFromStreamableHTTP(ctx, serviceKey, mcpService.GetHttpConnection(), toolManager, promptManager, resourceManager, isReload, serviceConfig)
		if err != nil {
			return "", nil, err
		}
	default:
		return "", nil, fmt.Errorf("MCPService definition requires either stdio_connection or http_connection")
	}

	log.Info("Registered MCP service", "serviceKey", serviceKey, "toolsAdded", len(discoveredTools))
	return serviceKey, discoveredTools, nil
}

type mcpConnection struct {
	client      *mcp.Client
	stdioConfig *configv1.McpStdioConnection
	httpAddress string
	httpClient  *http.Client
}

func (c *mcpConnection) withMCPClientSession(ctx context.Context, f func(cs ClientSession) error) error {
	var transport mcp.Transport
	switch {
	case c.stdioConfig != nil:
		cmd := buildCommandFromStdioConfig(c.stdioConfig)
		transport = &mcp.CommandTransport{
			Command: cmd,
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

func (c *mcpConnection) CallTool(ctx context.Context, params *mcp.CallToolParams) (*mcp.CallToolResult, error) {
	var result *mcp.CallToolResult
	err := c.withMCPClientSession(ctx, func(cs ClientSession) error {
		var err error
		result, err = cs.CallTool(ctx, params)
		return err
	})
	return result, err
}

func buildCommandFromStdioConfig(stdio *configv1.McpStdioConnection) *exec.Cmd {
	command := stdio.GetCommand()
	args := stdio.GetArgs()

	// If the command is 'docker', handle it directly, including sudo if needed.
	if command == "docker" {
		useSudo := os.Getenv("USE_SUDO_FOR_DOCKER") == "1"
		if useSudo {
			fullArgs := append([]string{command}, args...)
			return exec.Command("sudo", fullArgs...)
		}
		return exec.Command(command, args...)
	}

	// Combine all commands into a single script.
	var scriptCommands []string
	scriptCommands = append(scriptCommands, stdio.GetSetupCommands()...)

	// Add the main command. `exec` is used to replace the shell process with the main command.
	mainCommandParts := []string{"exec", command}
	mainCommandParts = append(mainCommandParts, args...)
	scriptCommands = append(scriptCommands, strings.Join(mainCommandParts, " "))

	script := strings.Join(scriptCommands, " && ")

	// Determine the container image.
	image := stdio.GetContainerImage()
	if image == "" {
		// Try to guess based on the main command if no image is provided.
		image = GetContainerImageForCommand(command)
	}

	// If an image is determined, try to run it in a container.
	if image != "" {
		if util.IsDockerSocketAccessible() {
			dockerArgs := []string{"run", "--rm", "-i"}
			if wd := stdio.GetWorkingDirectory(); wd != "" {
				dockerArgs = append(dockerArgs, "-w", wd)
			}
			dockerArgs = append(dockerArgs, image, "/bin/sh", "-c", script)
			return exec.Command("docker", dockerArgs...)
		}
		logging.GetLogger().Warn("Docker socket not accessible, falling back to running command on host. This may fail if the command is not in the system's PATH.", "command", command)
	}

	// Otherwise, run the script directly on the host.
	cmd := exec.Command("/bin/sh", "-c", script)
	cmd.Dir = stdio.GetWorkingDirectory()
	return cmd
}

func (u *MCPUpstream) createAndRegisterMCPItemsFromStdio(
	ctx context.Context,
	serviceKey string,
	stdio *configv1.McpStdioConnection,
	toolManager tool.ToolManagerInterface,
	promptManager prompt.PromptManagerInterface,
	resourceManager resource.ResourceManagerInterface,
	isReload bool,
	serviceConfig *configv1.UpstreamServiceConfig,
) ([]*configv1.ToolDefinition, error) {
	if stdio == nil {
		return nil, fmt.Errorf("stdio connection config is nil")
	}

	cmd := buildCommandFromStdioConfig(stdio)
	transport := &mcp.CommandTransport{
		Command: cmd,
	}
	var mcpSdkClient *mcp.Client
	if newClientForTesting != nil {
		mcpSdkClient = newClientForTesting(&mcp.Implementation{
			Name:    "mcpxy",
			Version: "0.1.0",
		})
	} else {
		mcpSdkClient = mcp.NewClient(&mcp.Implementation{
			Name:    "mcpxy",
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
		return nil, fmt.Errorf("failed to connect to MCP service: %w", err)
	}

	// Register tools
	listToolsResult, err := cs.ListTools(ctx, &mcp.ListToolsParams{})
	if err != nil {
		return nil, fmt.Errorf("failed to list tools from MCP service: %w", err)
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
	callDefs := mcpService.GetCalls()
	callDefMap := make(map[string]*configv1.MCPCallDefinition)
	for _, def := range callDefs {
		callDefMap[def.GetToolName()] = def
	}

	discoveredTools := make([]*configv1.ToolDefinition, 0, len(listToolsResult.Tools))
	for _, mcpSDKTool := range listToolsResult.Tools {
		callDef, ok := callDefMap[mcpSDKTool.Name]
		if !ok {
			callDef = &configv1.MCPCallDefinition{}
		}

		newTool := tool.NewMCPTool(
			pb.Tool_builder{
				Name:        proto.String(mcpSDKTool.Name),
				Description: proto.String(mcpSDKTool.Description),
				ServiceId:   proto.String(serviceKey),
			}.Build(),
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
	if err != nil {
		// Do not fail if prompts are not supported
		logging.GetLogger().Warn("Failed to list prompts from MCP service", "error", err)
	} else {
		for _, mcpSDKPrompt := range listPromptsResult.Prompts {
			promptManager.AddPrompt(&mcpPrompt{
				mcpPrompt: mcpSDKPrompt,
				service:   serviceKey,
				mcpConnection: &mcpConnection{
					client:      mcpSdkClient,
					stdioConfig: stdio,
				},
			})
		}
	}

	// Register resources
	listResourcesResult, err := cs.ListResources(ctx, &mcp.ListResourcesParams{})
	if err != nil {
		// Do not fail if resources are not supported
		logging.GetLogger().Warn("Failed to list resources from MCP service", "error", err)
	} else {
		for _, mcpSDKResource := range listResourcesResult.Resources {
			resourceManager.AddResource(&mcpResource{
				mcpResource: mcpSDKResource,
				service:     serviceKey,
				mcpConnection: &mcpConnection{
					client:      mcpSdkClient,
					stdioConfig: stdio,
				},
			})
		}
	}

	return discoveredTools, nil
}

func (u *MCPUpstream) createAndRegisterMCPItemsFromStreamableHTTP(
	ctx context.Context,
	serviceKey string,
	httpConnection *configv1.McpStreamableHttpConnection,
	toolManager tool.ToolManagerInterface,
	promptManager prompt.PromptManagerInterface,
	resourceManager resource.ResourceManagerInterface,
	isReload bool,
	serviceConfig *configv1.UpstreamServiceConfig,
) ([]*configv1.ToolDefinition, error) {
	authenticator, err := auth.NewUpstreamAuthenticator(serviceConfig.GetUpstreamAuthentication())
	if err != nil {
		return nil, fmt.Errorf("failed to create authenticator for MCP upstream: %w", err)
	}

	httpClient, err := util.NewHttpClientWithTLS(httpConnection.GetTlsConfig())
	if err != nil {
		return nil, fmt.Errorf("failed to create http client with tls config: %w", err)
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
			Name:    "mcpxy",
			Version: "0.1.0",
		})
	} else {
		mcpSdkClient = mcp.NewClient(&mcp.Implementation{
			Name:    "mcpxy",
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
		return nil, fmt.Errorf("failed to connect to MCP service: %w", err)
	}

	// Register tools
	listToolsResult, err := cs.ListTools(ctx, &mcp.ListToolsParams{})
	if err != nil {
		return nil, fmt.Errorf("failed to list tools from MCP service: %w", err)
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
	callDefs := mcpService.GetCalls()
	callDefMap := make(map[string]*configv1.MCPCallDefinition)
	for _, def := range callDefs {
		callDefMap[def.GetToolName()] = def
	}

	discoveredTools := make([]*configv1.ToolDefinition, 0, len(listToolsResult.Tools))
	for _, mcpSDKTool := range listToolsResult.Tools {
		callDef, ok := callDefMap[mcpSDKTool.Name]
		if !ok {
			callDef = &configv1.MCPCallDefinition{}
		}

		newTool := tool.NewMCPTool(
			pb.Tool_builder{
				Name:        proto.String(mcpSDKTool.Name),
				Description: proto.String(mcpSDKTool.Description),
				ServiceId:   proto.String(serviceKey),
			}.Build(),
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
	if err != nil {
		logging.GetLogger().Warn("Failed to list prompts from MCP service", "error", err)
	} else {
		for _, mcpSDKPrompt := range listPromptsResult.Prompts {
			promptManager.AddPrompt(&mcpPrompt{
				mcpPrompt: mcpSDKPrompt,
				service:   serviceKey,
				mcpConnection: &mcpConnection{
					client:      mcpSdkClient,
					httpAddress: httpAddress,
					httpClient:  httpClient,
				},
			})
		}
	}

	// Register resources
	listResourcesResult, err := cs.ListResources(ctx, &mcp.ListResourcesParams{})
	if err != nil {
		logging.GetLogger().Warn("Failed to list resources from MCP service", "error", err)
	} else {
		for _, mcpSDKResource := range listResourcesResult.Resources {
			resourceManager.AddResource(&mcpResource{
				mcpResource: mcpSDKResource,
				service:     serviceKey,
				mcpConnection: &mcpConnection{
					client:      mcpSdkClient,
					httpAddress: httpAddress,
					httpClient:  httpClient,
				},
			})
		}
	}

	return discoveredTools, nil
}

type authenticatedRoundTripper struct {
	authenticator auth.UpstreamAuthenticator
	base          http.RoundTripper
}

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
