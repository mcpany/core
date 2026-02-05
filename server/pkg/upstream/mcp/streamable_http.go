// Package mcp provides MCP upstream integration.

package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"al.essio.dev/pkg/shellescape"
	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/auth"
	"github.com/mcpany/core/server/pkg/client"
	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/prompt"
	"github.com/mcpany/core/server/pkg/resource"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/mcpany/core/server/pkg/upstream"
	"github.com/mcpany/core/server/pkg/util"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
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
	// ListTools lists the tools available in the session.
	//
	// ctx is the context for the request.
	// params is the params.
	//
	// Returns the result.
	// Returns an error if the operation fails.
	ListTools(ctx context.Context, params *mcp.ListToolsParams) (*mcp.ListToolsResult, error)
	// ListPrompts lists the prompts available in the session.
	//
	// ctx is the context for the request.
	// params is the params.
	//
	// Returns the result.
	// Returns an error if the operation fails.
	ListPrompts(ctx context.Context, params *mcp.ListPromptsParams) (*mcp.ListPromptsResult, error)
	// ListResources lists the resources available in the session.
	//
	// ctx is the context for the request.
	// params is the params.
	//
	// Returns the result.
	// Returns an error if the operation fails.
	ListResources(ctx context.Context, params *mcp.ListResourcesParams) (*mcp.ListResourcesResult, error)
	// GetPrompt retrieves a prompt from the session.
	//
	// ctx is the context for the request.
	// params is the params.
	//
	// Returns the result.
	// Returns an error if the operation fails.
	GetPrompt(ctx context.Context, params *mcp.GetPromptParams) (*mcp.GetPromptResult, error)
	// ReadResource reads a resource from the session.
	//
	// ctx is the context for the request.
	// params is the params.
	//
	// Returns the result.
	// Returns an error if the operation fails.
	ReadResource(ctx context.Context, params *mcp.ReadResourceParams) (*mcp.ReadResourceResult, error)
	// CallTool calls a tool in the session.
	//
	// ctx is the context for the request.
	// params is the params.
	//
	// Returns the result.
	// Returns an error if the operation fails.
	CallTool(ctx context.Context, params *mcp.CallToolParams) (*mcp.CallToolResult, error)
	// Close closes the session.
	//
	// Returns an error if the operation fails.
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

// Upstream implements the upstream.Upstream interface for services that are
// themselves MCP-compliant. It connects to the downstream MCP service, discovers
// its tools, prompts, and resources, and registers them with the current server,
// effectively acting as a proxy or aggregator.
type Upstream struct {
	sessionRegistry *SessionRegistry
	// BundleBaseDir is the directory where bundles are extracted.
	BundleBaseDir string
	globalSettings *configv1.GlobalSettings

	mu        sync.RWMutex
	serviceID string
}

// Shutdown cleans up any temporary resources associated with the upstream, such
// as extracted bundle directories.
func (u *Upstream) Shutdown(_ context.Context) error {
	u.mu.RLock()
	serviceID := u.serviceID
	u.mu.RUnlock()

	if serviceID != "" {
		untrackBundle(serviceID)
		tempDir := filepath.Join(u.BundleBaseDir, serviceID)
		if _, err := os.Stat(tempDir); err == nil {
			logging.GetLogger().Info("Cleaning up bundle temp directory", "dir", tempDir)
			if err := os.RemoveAll(tempDir); err != nil {
				return fmt.Errorf("failed to remove bundle temp directory: %w", err)
			}
		}
	}
	return nil
}

// NewUpstream creates a new instance of Upstream.
//
// globalSettings is the globalSettings.
//
// Returns the result.
func NewUpstream(globalSettings *configv1.GlobalSettings) upstream.Upstream {
	return &Upstream{
		sessionRegistry: NewSessionRegistry(),
		BundleBaseDir:   bundleBaseDir,
		globalSettings:  globalSettings,
	}
}

// mcpPrompt is a wrapper around the standard mcp.Prompt that associates it with
// a specific service and provides the necessary connection details for execution.
type mcpPrompt struct {
	mcpPrompt *mcp.Prompt
	service   string
	*mcpConnection
}

// Prompt returns the underlying *mcp.Prompt definition.
//
// Returns the result.
func (p *mcpPrompt) Prompt() *mcp.Prompt {
	return p.mcpPrompt
}

// Service returns the ID of the service that this prompt belongs to.
//
// Returns the result.
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
				arguments[k] = util.ToString(v)
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
//
// Returns the result.
func (r *mcpResource) Resource() *mcp.Resource {
	return r.mcpResource
}

// Service returns the ID of the service that this resource belongs to.
//
// Returns the result.
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
func (r *mcpResource) Subscribe(_ context.Context) error {
	return fmt.Errorf("subscribing to resources on mcp upstreams is not yet implemented")
}

// Register handles the registration of another MCP service as an upstream. It
// determines the connection type (stdio or HTTP), connects to the downstream
// service, lists its available tools, prompts, and resources, and registers
// them with the appropriate managers.
func (u *Upstream) Register(
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
	u.mu.Lock()
	u.serviceID = serviceID
	u.mu.Unlock()

	// Track bundle potential usage early to prevent GC race during setup
	trackBundle(serviceID)
	// Trigger GC lazily, but after we are safe
	triggerGC()
	defer func() {
		if err != nil {
			untrackBundle(serviceID)
		}
	}()

	mcpService := serviceConfig.GetMcpService()
	if mcpService == nil {
		err = fmt.Errorf("mcp service config is nil")
		return "", nil, nil, err
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
	case configv1.McpUpstreamService_BundleConnection_case:
		discoveredTools, discoveredResources, err = u.createAndRegisterMCPItemsFromBundle(ctx, serviceID, mcpService.GetBundleConnection(), toolManager, promptManager, resourceManager, isReload, serviceConfig)
		if err != nil {
			return "", nil, nil, err
		}
	default:
		err = fmt.Errorf("MCPService definition requires stdio_connection, http_connection, or bundle_connection")
		return "", nil, nil, err
	}

	log.Info("Registered MCP service", "serviceID", serviceID, "toolsAdded", len(discoveredTools))
	return serviceID, discoveredTools, discoveredResources, nil
}

// mcpConnection holds the necessary information to connect to a downstream MCP
// service, whether it's via stdio or HTTP. It also implements the
// client.MCPClient interface, allowing it to be used as a proxy.
type mcpConnection struct {
	client          *mcp.Client
	stdioConfig     *configv1.McpStdioConnection
	bundleTransport *BundleDockerTransport
	httpAddress     string
	httpClient      *http.Client
	sessionRegistry *SessionRegistry
	globalSettings  *configv1.GlobalSettings
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
			// We need global settings here if we want to support sudo.
			// However, mcpConnection doesn't have GlobalSettings stored?
			useSudo := false
			if c.globalSettings != nil {
				useSudo = c.globalSettings.GetUseSudoForDocker()
			}
			cmd, err := buildCommandFromStdioConfig(ctx, c.stdioConfig, useSudo)
			if err != nil {
				return fmt.Errorf("failed to build command from stdio config: %w", err)
			}
			transport = &StdioTransport{
				Command: cmd,
			}
		}
	case c.bundleTransport != nil:
		transport = c.bundleTransport
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
	defer func() {
		// Unregister session if registry is present
		if c.sessionRegistry != nil {
			if mcpSession, ok := cs.(mcp.Session); ok {
				c.sessionRegistry.Unregister(mcpSession)
			}
		}
		_ = cs.Close()
	}()

	// Register session if downstream session is available in context and registry is present
	if c.sessionRegistry != nil {
		if downstreamSession, ok := tool.GetSession(ctx); ok {
			if mcpSession, ok := cs.(mcp.Session); ok {
				c.sessionRegistry.Register(mcpSession, downstreamSession)
			}
		}
	}

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
func buildCommandFromStdioConfig(ctx context.Context, stdio *configv1.McpStdioConnection, useSudo bool) (*exec.Cmd, error) {
	command := stdio.GetCommand()
	args := stdio.GetArgs()
	resolvedEnv, err := util.ResolveSecretMap(ctx, stdio.GetEnv(), nil)
	if err != nil {
		return nil, err
	}

	// Pre-flight check: Ensure the command exists.
	// We only do this check if it's not a docker command, because "docker"
	// is generally assumed to be in PATH, and we might be prepending "sudo" later.
	// However, if the command is NOT docker, we should check if it exists.
	if command != "docker" {
		if _, err := exec.LookPath(command); err != nil {
			return nil, fmt.Errorf("command not found: %s. Please check that the executable exists and is in the system PATH", command)
		}
	}

	// If the command is docker, we might need to prepend sudo.
	if command == "docker" {
		if useSudo {
			// If we are using sudo, we need to prepend sudo to the command
			// and keep the arguments as is.
			newArgs := append([]string{command}, args...)
			command = "sudo"
			args = newArgs
		}
		cmd := exec.CommandContext(ctx, command, args...) //nolint:gosec // Command is validated/configured by user
		cmd.Dir = stdio.GetWorkingDirectory()
		cmd.Env = buildSafeEnv(resolvedEnv)
		env := cmd.Env // For validation below

		// Validate required environment variables
		if err := validateRequiredEnv(env, stdio.GetValidation()); err != nil {
			return nil, err
		}

		return cmd, nil
	}

	// Combine all commands into a single script.
	var scriptCommands []string
	setupCommands := stdio.GetSetupCommands()

	// If no setup commands are provided, execute the command directly.
	// This avoids shell injection risks and is safer.
	if len(setupCommands) == 0 {
		cmd := exec.CommandContext(ctx, command, args...) //nolint:gosec // Command is configured by user
		cmd.Dir = stdio.GetWorkingDirectory()
		cmd.Env = buildSafeEnv(resolvedEnv)
		env := cmd.Env // For validation below

		// Validate required environment variables
		if err := validateRequiredEnv(env, stdio.GetValidation()); err != nil {
			return nil, err
		}

		return cmd, nil
	}

	// Sentinel Security: Disable setup_commands by default as they allow arbitrary command execution.
	if os.Getenv("MCP_ALLOW_UNSAFE_SETUP_COMMANDS") != "true" {
		return nil, fmt.Errorf("setup_commands are disabled by default for security reasons. Set MCP_ALLOW_UNSAFE_SETUP_COMMANDS=true to enable them if you trust the configuration")
	}

	logging.GetLogger().Warn("Using setup_commands in StdioTransport is dangerous and allows Command Injection if config is untrusted.", "setup_commands", "HIDDEN")
	scriptCommands = append(scriptCommands, setupCommands...)

	// Add the main command. `exec` is used to replace the shell process with the main command.
	mainCommandParts := []string{"exec", shellescape.Quote(command)}
	for _, arg := range args {
		mainCommandParts = append(mainCommandParts, shellescape.Quote(arg))
	}
	scriptCommands = append(scriptCommands, strings.Join(mainCommandParts, " "))

	script := strings.Join(scriptCommands, " && ")

	cmd := exec.CommandContext(ctx, "/bin/sh", "-c", script) //nolint:gosec // Script is configured by user
	cmd.Dir = stdio.GetWorkingDirectory()
	cmd.Env = buildSafeEnv(resolvedEnv)
	env := cmd.Env // For validation below

	// Validate required environment variables
	if err := validateRequiredEnv(env, stdio.GetValidation()); err != nil {
		return nil, err
	}

	return cmd, nil
}

func validateRequiredEnv(env []string, validation *configv1.EnvValidation) error {
	if validation == nil {
		return nil
	}
	required := validation.GetRequiredEnv()
	if len(required) == 0 {
		return nil
	}

	envMap := make(map[string]struct{})
	for _, e := range env {
		parts := strings.SplitN(e, "=", 2)
		if len(parts) > 0 {
			envMap[parts[0]] = struct{}{}
		}
	}

	var missing []string
	for _, req := range required {
		if _, ok := envMap[req]; !ok {
			missing = append(missing, req)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required environment variables: %s", strings.Join(missing, ", "))
	}
	return nil
}

// createAndRegisterMCPItemsFromStdio handles the registration of an MCP service
// that is connected via standard I/O (e.g., a local command or a Docker
// container). It establishes the connection, discovers the service's
// capabilities, and registers them.
//
//nolint:funlen
func (u *Upstream) createAndRegisterMCPItemsFromStdio(
	ctx context.Context,
	serviceID string,
	stdio *configv1.McpStdioConnection,
	toolManager tool.ManagerInterface,
	promptManager prompt.ManagerInterface,
	resourceManager resource.ManagerInterface,
	_ bool,
	serviceConfig *configv1.UpstreamServiceConfig,
) ([]*configv1.ToolDefinition, []*configv1.ResourceDefinition, error) {
	if stdio == nil {
		return nil, nil, fmt.Errorf("stdio connection config is nil")
	}

	useSudo := false
	if u.globalSettings != nil {
		useSudo = u.globalSettings.GetUseSudoForDocker()
	}

	transport, err := createStdioTransport(ctx, stdio, useSudo)
	if err != nil {
		return nil, nil, err
	}

	mcpSdkClient, err := u.createMCPClient(ctx)
	if err != nil {
		return nil, nil, err
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
	defer func() { _ = cs.Close() }()

	// Register tools
	listToolsResult, err := cs.ListTools(ctx, &mcp.ListToolsParams{})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list tools from MCP service: %w", err)
	}

	var toolClient client.MCPClient
	var promptConnection *mcpConnection

	if newClientImplForTesting != nil {
		toolClient = newClientImplForTesting(mcpSdkClient, stdio, "", nil)
		promptConnection = &mcpConnection{
			client:          mcpSdkClient,
			stdioConfig:     stdio,
			sessionRegistry: u.sessionRegistry,
			globalSettings:  u.globalSettings,
		}
	} else {
		conn := &mcpConnection{
			client:          mcpSdkClient,
			stdioConfig:     stdio,
			sessionRegistry: u.sessionRegistry,
			globalSettings:  u.globalSettings,
		}
		toolClient = conn
		promptConnection = conn
	}

	// We pass the new client and connection to the processing logic
	return u.processMCPItems(ctx, serviceID, listToolsResult, toolClient, promptConnection, cs, toolManager, promptManager, resourceManager, serviceConfig)
}

func createStdioTransport(ctx context.Context, stdio *configv1.McpStdioConnection, useSudo bool) (mcp.Transport, error) {
	image := stdio.GetContainerImage()
	if image != "" {
		if util.IsDockerSocketAccessible() {
			return &DockerTransport{
				StdioConfig: stdio,
			}, nil
		}
		return nil, fmt.Errorf("docker socket not accessible, but container_image is specified")
	}
	cmd, err := buildCommandFromStdioConfig(ctx, stdio, useSudo)
	if err != nil {
		return nil, fmt.Errorf("failed to build command: %w", err)
	}
	// Use our robust StdioTransport instead of mcp.CommandTransport
	return &StdioTransport{
		Command: cmd,
	}, nil
}

// processMCPItems handles the common logic of registering tools, prompts, and resources.
// processMCPItems handles the common logic of registering tools, prompts, and resources.
func (u *Upstream) processMCPItems(
	ctx context.Context,
	serviceID string,
	listToolsResult *mcp.ListToolsResult,
	toolClient client.MCPClient,
	promptConnection *mcpConnection,
	cs ClientSession,
	toolManager tool.ManagerInterface,
	promptManager prompt.ManagerInterface,
	resourceManager resource.ManagerInterface,
	serviceConfig *configv1.UpstreamServiceConfig,
) ([]*configv1.ToolDefinition, []*configv1.ResourceDefinition, error) { //nolint:unparam
	mcpService := serviceConfig.GetMcpService()

	discoveredTools := u.registerTools(serviceID, mcpService, listToolsResult, toolClient, toolManager, serviceConfig)
	u.registerPrompts(ctx, serviceID, mcpService, cs, promptManager, promptConnection)
	discoveredResources := u.registerResources(ctx, serviceID, mcpService, cs, resourceManager, promptConnection)
	u.registerDynamicResources(serviceID, mcpService, toolManager, resourceManager, discoveredTools)

	return discoveredTools, discoveredResources, nil
}

func (u *Upstream) registerTools(
	serviceID string,
	mcpService *configv1.McpUpstreamService,
	listToolsResult *mcp.ListToolsResult,
	toolClient client.MCPClient,
	toolManager tool.ManagerInterface,
	serviceConfig *configv1.UpstreamServiceConfig,
) []*configv1.ToolDefinition {
	configToolDefs := mcpService.GetTools()
	calls := mcpService.GetCalls()
	configToolMap := make(map[string]*configv1.ToolDefinition)
	for _, toolDef := range configToolDefs {
		configToolMap[toolDef.GetName()] = toolDef
	}

	discoveredTools := make([]*configv1.ToolDefinition, 0, len(listToolsResult.Tools))
	for _, mcpSDKTool := range listToolsResult.Tools {
		configTool, hasConfig := configToolMap[mcpSDKTool.Name]

		// If validation fails or we shouldn't register this tool, continue
		if hasConfig && configTool.GetDisable() {
			logging.GetLogger().Info("Skipping disabled tool", "toolName", mcpSDKTool.Name)
			continue
		}

		// Check auto-discovery
		if !hasConfig && !serviceConfig.GetAutoDiscoverTool() && !mcpService.GetToolAutoDiscovery() {
			// Skip tool if not explicitly configured and auto-discovery is disabled
			continue
		}

		callDef := configv1.MCPCallDefinition_builder{}.Build()
		if hasConfig {
			if call, callOk := calls[configTool.GetCallId()]; callOk {
				callDef = call
			} else {
				logging.GetLogger().Warn("Call definition not found for tool", "call_id", configTool.GetCallId(), "tool_name", mcpSDKTool.Name)
			}
		}

		pbTool, err := tool.ConvertMCPToolToProto(mcpSDKTool)
		if err != nil {
			logging.GetLogger().Error("Failed to convert mcp tool to proto", "error", err)
			continue
		}
		pbTool.SetServiceId(serviceID)

		// Apply overrides from config
		if hasConfig {
			strategy := configTool.GetMergeStrategy()
			if strategy == configv1.ToolDefinition_MERGE_STRATEGY_OVERRIDE {
				// Override means we prefer the config definition, but we still keep the discovered
				// name and service ID to maintain connectivity.
				// We still fall back to discovered schema/description if not provided in config
				// to avoid breaking the tool unless the user explicitly wants to partial-override.
				// (Actually, a true override might want to replace schema too, which we do if provided).
				if configTool.GetDescription() != "" {
					pbTool.SetDescription(configTool.GetDescription())
				}
				if configTool.GetTitle() != "" {
					pbTool.SetDisplayName(configTool.GetTitle())
				}
				// If InputSchema is provided in config, use it.
				if configTool.GetInputSchema() != nil {
					// Override input schema
					pbTool.SetInputSchema(configTool.GetInputSchema())
				}
			} else {
				// Merge Strategy (Default)
				if configTool.GetDescription() != "" {
					pbTool.SetDescription(configTool.GetDescription())
				}
				if configTool.GetTitle() != "" {
					pbTool.SetDisplayName(configTool.GetTitle())
				}
				if configTool.GetInputSchema() != nil {
					if !pbTool.HasInputSchema() {
						pbTool.SetInputSchema(configTool.GetInputSchema())
					} else {
						mergeStructs(pbTool.GetInputSchema(), configTool.GetInputSchema())
					}
				}
			}

			// Always apply tags from config
			if len(configTool.GetTags()) > 0 {
				pbTool.SetTags(configTool.GetTags())
			}

			// Apply other annotations/hints
			if !pbTool.HasAnnotations() {
				pbTool.SetAnnotations(v1.ToolAnnotations_builder{}.Build())
			}
			if configTool.GetReadOnlyHint() { // Only override if true? Or if set? Proto bool is false by default.
				pbTool.GetAnnotations().SetReadOnlyHint(configTool.GetReadOnlyHint())
			}
			// Note: We might want headers/properties check for destructive/idempotent/open_world too

			// We can add more overrides here if needed (e.g. InputSchema)
		}

		mcpTool := tool.NewMCPTool(pbTool, toolClient, callDef)
		if err := toolManager.AddTool(mcpTool); err != nil {
			logging.GetLogger().Error("Failed to add tool", "error", err)
			continue
		}
		discoveredTools = append(discoveredTools, configv1.ToolDefinition_builder{
			Name:        proto.String(mcpSDKTool.Name),
			Description: proto.String(pbTool.GetDescription()),
		}.Build())
	}
	return discoveredTools
}

func (u *Upstream) registerPrompts(
	ctx context.Context,
	serviceID string,
	mcpService *configv1.McpUpstreamService,
	cs ClientSession,
	promptManager prompt.ManagerInterface,
	promptConnection *mcpConnection,
) {
	listPromptsResult, err := cs.ListPrompts(ctx, &mcp.ListPromptsParams{})
	if err != nil {
		logging.GetLogger().Warn("Failed to list prompts from MCP service", "error", err)
	} else {
		configPromptMap := make(map[string]*configv1.PromptDefinition)
		for _, p := range mcpService.GetPrompts() {
			configPromptMap[p.GetName()] = p
		}

		for _, mcpSDKPrompt := range listPromptsResult.Prompts {
			if configPrompt, ok := configPromptMap[mcpSDKPrompt.Name]; ok {
				if configPrompt.GetDisable() {
					logging.GetLogger().Info("Skipping disabled prompt (auto-discovered)", "promptName", mcpSDKPrompt.Name)
					continue
				}
			}
			promptManager.AddPrompt(&mcpPrompt{
				mcpPrompt:     mcpSDKPrompt,
				service:       serviceID,
				mcpConnection: promptConnection,
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
}

func (u *Upstream) registerResources(
	ctx context.Context,
	serviceID string,
	mcpService *configv1.McpUpstreamService,
	cs ClientSession,
	resourceManager resource.ManagerInterface,
	promptConnection *mcpConnection,
) []*configv1.ResourceDefinition {
	listResourcesResult, err := cs.ListResources(ctx, &mcp.ListResourcesParams{})
	if err != nil {
		logging.GetLogger().Warn("Failed to list resources from MCP service", "error", err)
		return nil
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
			mcpResource:   mcpSDKResource,
			service:       serviceID,
			mcpConnection: promptConnection,
		})
		discoveredResources = append(discoveredResources, convertMCPResourceToProto(mcpSDKResource))
	}
	return discoveredResources
}

func (u *Upstream) registerDynamicResources(
	serviceID string,
	mcpService *configv1.McpUpstreamService,
	toolManager tool.ManagerInterface,
	resourceManager resource.ManagerInterface,
	_ []*configv1.ToolDefinition,
) {
	log := logging.GetLogger()
	// Create a map of call ID to tool name only for the tools we just discovered/configured.
	// Actually we should look up in the toolManager for tools registered for this service to be safe,
	// or use the tool definitions we have.
	// The original code used configToolDefs to build the map, but we also discover tools.
	// Wait, the dynamic resource links a resource to a *Call ID*.
	// We need to find the tool that corresponds to that Call ID.
	// If the tool was auto-discovered, we assigned it a Call ID (empty or from config).
	// If it was from config, it has a Call ID.

	// Rebuilding callIDToName map might be tricky if we don't have all tools handy.
	// However, the dynamic resource config relies on the Call Defined in the Config usually.
	// Let's look at how original code did it:
	// It iterated `configToolDefs` (from mcpService.GetTools()) to build `callIDToName`.
	// So it only supports linking to tools *explicitly defined in config*?
	// The original code:
	// for _, d := range configToolDefs { callIDToName[d.GetCallId()] = d.GetName() }
	// Yes, `configToolDefs` comes from `mcpService.GetTools()`.
	// So we should do the same.

	configToolDefs := mcpService.GetTools()
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
			toolObj, ok := toolManager.GetTool(serviceID + "." + sanitizedToolName)
			if !ok {
				log.Error("Tool not found for dynamic resource", "toolName", toolName)
				continue
			}
			dynamicResource, err := resource.NewDynamicResource(resourceDef, toolObj)
			if err != nil {
				log.Error("Failed to create dynamic resource", "error", err)
				continue
			}
			resourceManager.AddResource(dynamicResource)
		}
	}
}

// createAndRegisterMCPItemsFromStreamableHTTP handles the registration of an MCP
// service that is connected via HTTP. It establishes the connection, discovers
// the service's capabilities, and registers them.
//

func (u *Upstream) createAndRegisterMCPItemsFromStreamableHTTP(
	ctx context.Context,
	serviceID string,
	httpConnection *configv1.McpStreamableHttpConnection,
	toolManager tool.ManagerInterface,
	promptManager prompt.ManagerInterface,
	resourceManager resource.ManagerInterface,
	_ bool,
	serviceConfig *configv1.UpstreamServiceConfig,
) ([]*configv1.ToolDefinition, []*configv1.ResourceDefinition, error) {
	authenticator, err := auth.NewUpstreamAuthenticator(serviceConfig.GetUpstreamAuth())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create authenticator for MCP upstream: %w", err)
	}

	httpClient, err := util.NewHTTPClientWithTLS(httpConnection.GetTlsConfig())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create http client with tls config: %w", err)
	}
	// Disable redirects to prevent credential leakage, unless explicitly allowed.
	if !httpConnection.GetAllowHttpRedirect() {
		httpClient.CheckRedirect = func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}
	httpClient.Transport = &authenticatedRoundTripper{
		authenticator: authenticator,
		base:          httpClient.Transport,
	}

	httpAddress := httpConnection.GetHttpAddress()
	if httpAddress == "" {
		return nil, nil, fmt.Errorf("mcp http service address is required")
	}
	uURL, err := url.ParseRequestURI(httpAddress)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid mcp http service address: %w", err)
	}
	if uURL.Scheme != "http" && uURL.Scheme != "https" {
		return nil, nil, fmt.Errorf("invalid mcp http service address scheme: %s (must be http or https)", uURL.Scheme)
	}

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
		}, &mcp.ClientOptions{
			CreateMessageHandler: u.handleCreateMessage,
		})
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
	defer func() { _ = cs.Close() }()

	// Register tools
	listToolsResult, err := cs.ListTools(ctx, &mcp.ListToolsParams{})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list tools from MCP service: %w", err)
	}

	var toolClient client.MCPClient
	var promptConnection *mcpConnection

	if newClientImplForTesting != nil {
		toolClient = newClientImplForTesting(mcpSdkClient, nil, httpAddress, httpClient)
		promptConnection = &mcpConnection{
			client:      mcpSdkClient,
			httpAddress: httpAddress,
			httpClient:  httpClient,
			sessionRegistry: u.sessionRegistry,
		}
	} else {
		conn := &mcpConnection{
			client:      mcpSdkClient,
			httpAddress: httpAddress,
			httpClient:  httpClient,
			sessionRegistry: u.sessionRegistry,
		}
		toolClient = conn
		promptConnection = conn
	}

	return u.processMCPItems(ctx, serviceID, listToolsResult, toolClient, promptConnection, cs, toolManager, promptManager, resourceManager, serviceConfig)
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
//
// req is the request object.
//
// Returns the response.
// Returns an error if the operation fails.
func (t *StreamableHTTP) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.Client == nil {
		t.Client = http.DefaultClient
	}
	return t.Client.Do(req)
}

// mergeStructs recursively merges src into dst.
// Keys in src override keys in dst.
// If both values are Structs, they are merged recursively.
func mergeStructs(dst, src *structpb.Struct) {
	for k, v := range src.GetFields() {
		if dstVal, ok := dst.Fields[k]; ok {
			// If both are Structs, recurse
			if dstStruct := dstVal.GetStructValue(); dstStruct != nil {
				if srcStruct := v.GetStructValue(); srcStruct != nil {
					mergeStructs(dstStruct, srcStruct)
					continue
				}
			}
		}
		// Otherwise replace
		dst.Fields[k] = v
	}
}

func (u *Upstream) createMCPClient(_ context.Context) (*mcp.Client, error) { //nolint:unparam
	if newClientForTesting != nil {
		return newClientForTesting(&mcp.Implementation{
			Name:    "mcpany",
			Version: "0.1.0",
		}), nil
	}

	return mcp.NewClient(&mcp.Implementation{
		Name:    "mcpany",
		Version: "0.1.0",
	}, &mcp.ClientOptions{
		CreateMessageHandler: u.handleCreateMessage,
	}), nil
}

func buildSafeEnv(resolvedEnv map[string]string) []string {
	// Sentinel Security: Only pass allowed environment variables to prevent secret leakage
	allowedEnvVars := []string{"PATH", "HOME", "USER", "TMPDIR", "TZ", "LANG", "LC_ALL"}
	env := make([]string, 0, len(allowedEnvVars)+len(resolvedEnv))
	for _, key := range allowedEnvVars {
		if val, ok := os.LookupEnv(key); ok {
			env = append(env, fmt.Sprintf("%s=%s", key, val))
		}
	}
	for k, v := range resolvedEnv {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}
	return env
}

func (u *Upstream) handleCreateMessage(ctx context.Context, req *mcp.ClientRequest[*mcp.CreateMessageParams]) (*mcp.CreateMessageResult, error) {
	session := req.Session
	// Check for nil interface value
	if util.IsNil(session) {
		return nil, fmt.Errorf("no session associated with request")
	}
	if toolSession, ok := u.sessionRegistry.Get(session); ok {
		return toolSession.CreateMessage(ctx, req.Params)
	}
	return nil, fmt.Errorf("no downstream session found for upstream session")
}
