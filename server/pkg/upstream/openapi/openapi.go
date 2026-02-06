// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package openapi provides OpenAPI integration for the upstream service.
package openapi

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/jellydator/ttlcache/v3"
	configv1 "github.com/mcpany/core/proto/config/v1"
	pb "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/auth"
	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/prompt"
	"github.com/mcpany/core/server/pkg/resource"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/mcpany/core/server/pkg/upstream"
	"github.com/mcpany/core/server/pkg/util"
	"google.golang.org/protobuf/proto"
)

// OpenAPIUpstream implements the upstream.Upstream interface for services that
// are defined by an OpenAPI specification.
//
// Summary: OpenAPI upstream implementation.
//
// It parses the spec, discovers the available operations, and registers them as tools.
type OpenAPIUpstream struct { //nolint:revive
	openapiCache *ttlcache.Cache[string, *openapi3.T]
	httpClients  map[string]*http.Client
	mu           sync.Mutex
	serviceID    string
}

// Shutdown gracefully terminates the OpenAPI upstream service.
//
// Summary: Shuts down the OpenAPI upstream service.
//
// For HTTP-based services, this typically means closing any persistent connections.
//
// Parameters:
//   - _ : context.Context. Unused.
//
// Returns:
//   - error: Always returns nil.
func (u *OpenAPIUpstream) Shutdown(_ context.Context) error {
	u.mu.Lock()
	defer u.mu.Unlock()

	if client, ok := u.httpClients[u.serviceID]; ok {
		client.CloseIdleConnections()
		delete(u.httpClients, u.serviceID)
	}
	return nil
}

// NewOpenAPIUpstream creates a new instance of OpenAPIUpstream.
//
// Summary: Creates a new instance of OpenAPIUpstream.
//
// It initializes a cache for storing parsed OpenAPI documents to avoid redundant parsing.
//
// Returns:
//   - upstream.Upstream: A new upstream instance.
func NewOpenAPIUpstream() upstream.Upstream {
	cache := ttlcache.New[string, *openapi3.T](
		ttlcache.WithTTL[string, *openapi3.T](5 * time.Minute),
	)
	go cache.Start()

	return &OpenAPIUpstream{
		openapiCache: cache,
		httpClients:  make(map[string]*http.Client),
	}
}

// Register processes an OpenAPI service configuration.
//
// Summary: Registers an OpenAPI service.
//
// It parses the OpenAPI specification, extracts the operations, converts them into tools, and
// registers them with the tool manager.
//
// Parameters:
//   - ctx: context.Context. The context for the registration.
//   - serviceConfig: *configv1.UpstreamServiceConfig. The service configuration.
//   - toolManager: tool.ManagerInterface. The tool manager.
//   - promptManager: prompt.ManagerInterface. The prompt manager.
//   - resourceManager: resource.ManagerInterface. The resource manager.
//   - isReload: bool. Whether this is a reload.
//
// Returns:
//   - string: The service ID.
//   - []*configv1.ToolDefinition: A list of discovered tool definitions.
//   - []*configv1.ResourceDefinition: A list of discovered resource definitions.
//   - error: An error if registration fails.
func (u *OpenAPIUpstream) Register(
	ctx context.Context,
	serviceConfig *configv1.UpstreamServiceConfig,
	toolManager tool.ManagerInterface,
	promptManager prompt.ManagerInterface,
	resourceManager resource.ManagerInterface,
	isReload bool,
) (string, []*configv1.ToolDefinition, []*configv1.ResourceDefinition, error) {
	log := logging.GetLogger()

	// Calculate SHA256 for the ID
	h := sha256.New()
	h.Write([]byte(serviceConfig.GetName()))
	serviceConfig.SetId(hex.EncodeToString(h.Sum(nil)))

	// Sanitize the service name
	sanitizedName, err := util.SanitizeServiceName(serviceConfig.GetName())
	if err != nil {
		return "", nil, nil, err
	}
	serviceConfig.SetSanitizedName(sanitizedName)

	u.mu.Lock()
	u.serviceID = sanitizedName
	u.mu.Unlock()
	serviceID := sanitizedName

	openapiService := serviceConfig.GetOpenapiService()
	if openapiService == nil {
		return "", nil, nil, fmt.Errorf("openapi service config is nil")
	}

	if address := openapiService.GetAddress(); address != "" {
		uURL, err := url.ParseRequestURI(address)
		if err != nil {
			return "", nil, nil, fmt.Errorf("invalid openapi service address: %w", err)
		}
		if uURL.Scheme != "http" && uURL.Scheme != "https" {
			return "", nil, nil, fmt.Errorf("invalid openapi service address scheme: %s (must be http or https)", uURL.Scheme)
		}
	}

	info := &tool.ServiceInfo{
		Name:   serviceConfig.GetName(),
		Config: serviceConfig,
	}
	toolManager.AddServiceInfo(serviceID, info)

	specContent := openapiService.GetSpecContent()
	if specContent == "" {
		specURL := openapiService.GetSpecUrl()
		if specURL != "" {
			if uURL, err := url.ParseRequestURI(specURL); err != nil {
				logging.GetLogger().Warn("Invalid OpenAPI spec URL", "url", specURL, "error", err)
				specURL = ""
			} else if uURL.Scheme != "http" && uURL.Scheme != "https" {
				logging.GetLogger().Warn("Invalid OpenAPI spec URL scheme (must be http or https)", "url", specURL)
				specURL = ""
			}
		}

		if specURL != "" {
			client := &http.Client{
				Timeout: 30 * time.Second,
			}
			req, err := http.NewRequestWithContext(ctx, "GET", specURL, nil)
			if err != nil {
				logging.GetLogger().Warn("Failed to create request for OpenAPI spec", "url", specURL, "error", err)
			}
			var resp *http.Response
			if err == nil {
				resp, err = client.Do(req)
			}
			if err != nil {
				logging.GetLogger().Warn("Failed to fetch OpenAPI spec from url (continuing without tools)", "url", specURL, "error", err)
			} else {
				defer func() { _ = resp.Body.Close() }()
				if resp.StatusCode != http.StatusOK {
					logging.GetLogger().Warn("Failed to fetch OpenAPI spec from url (continuing without tools)", "url", specURL, "status", resp.StatusCode)
				} else {
					bodyBytes, err := io.ReadAll(resp.Body)
					if err != nil {
						logging.GetLogger().Warn("Failed to read OpenAPI spec body (continuing without tools)", "url", specURL, "error", err)
					} else {
						specContent = string(bodyBytes)
					}
				}
			}
		}
	}

	if specContent == "" {
		// If spec_url was not provided either, this is a configuration error.
		if openapiService.GetSpecUrl() == "" {
			return "", nil, nil, fmt.Errorf("OpenAPI spec content is missing")
		}
		// If spec_url was provided but failed to load, we warned above.
		// We return an error so that the registration worker can retry later.
		// This is critical for startup reliability when upstream services might not be ready yet.
		return "", nil, nil, fmt.Errorf("OpenAPI spec content is missing or failed to load from %s", openapiService.GetSpecUrl())
	}

	hash := sha256.Sum256([]byte(specContent))
	cacheKey := hex.EncodeToString(hash[:])

	item := u.openapiCache.Get(cacheKey)
	var doc *openapi3.T
	if item != nil {
		doc = item.Value()
	} else {
		var err error
		_, doc, err = parseOpenAPISpec(ctx, []byte(specContent))
		if err != nil {
			return "", nil, nil, fmt.Errorf("failed to parse OpenAPI spec for service '%s' from content: %w", serviceID, err)
		}
		u.openapiCache.Set(cacheKey, doc, ttlcache.DefaultTTL)
	}

	mcpOps := extractMcpOperationsFromOpenAPI(doc)
	pbTools := convertMcpOperationsToTools(mcpOps, doc, serviceID)
	discoveredTools := make([]*configv1.ToolDefinition, 0, len(mcpOps))
	for _, op := range mcpOps {
		discoveredTools = append(discoveredTools, configv1.ToolDefinition_builder{
			Name:        proto.String(op.OperationID),
			Description: proto.String(op.Description),
		}.Build())
	}

	numToolsAdded := u.addOpenAPIToolsToIndex(ctx, pbTools, serviceID, toolManager, resourceManager, isReload, doc, serviceConfig)
	u.createAndRegisterPrompts(ctx, serviceID, serviceConfig, promptManager, isReload)
	log.Info("Registered OpenAPI service", "serviceID", serviceID, "toolsAdded", numToolsAdded)

	return serviceID, discoveredTools, nil, nil
}

// getHTTPClient retrieves or creates an HTTP client for a given service. It
// ensures that each service has its own dedicated client, which can be
// configured with specific transports or timeouts.
func (u *OpenAPIUpstream) getHTTPClient(serviceID string) *http.Client {
	u.mu.Lock()
	defer u.mu.Unlock()

	if client, ok := u.httpClients[serviceID]; ok {
		return client
	}

	dialer := util.NewSafeDialer()
	// Allow overriding safety checks via environment variables (consistent with validation package)
	if os.Getenv("MCPANY_DANGEROUS_ALLOW_LOCAL_IPS") == util.TrueStr || os.Getenv("MCPANY_ALLOW_LOOPBACK_RESOURCES") == util.TrueStr {
		dialer.AllowLoopback = true
		// Typically allowing local IPs implies allowing private IPs too, but let's be explicit
		dialer.AllowPrivate = true
	}
	if os.Getenv("MCPANY_ALLOW_PRIVATE_NETWORK_RESOURCES") == util.TrueStr {
		dialer.AllowPrivate = true
	}

	transport := &http.Transport{
		DialContext:         dialer.DialContext,
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
	}

	u.httpClients[serviceID] = client
	return client
}

// httpClientImpl is a simple wrapper around *http.Client that implements the
// client.HTTPClient interface. This is used to adapt the standard library's
// HTTP client for use in components that expect this interface.
type httpClientImpl struct {
	client *http.Client
}

// Do sends an HTTP request and returns an HTTP response, fulfilling the
// client.HTTPClient interface.
//
// Summary: Sends an HTTP request.
//
// Parameters:
//   - req: *http.Request. The HTTP request.
//
// Returns:
//   - *http.Response: The HTTP response.
//   - error: An error if the request fails.
func (c *httpClientImpl) Do(req *http.Request) (*http.Response, error) {
	return c.client.Do(req)
}

// addOpenAPIToolsToIndex iterates through a list of protobuf tool definitions,
// creates an OpenAPITool for each, and registers it with the tool manager.
func (u *OpenAPIUpstream) addOpenAPIToolsToIndex(_ context.Context, pbTools []*pb.Tool, serviceID string, toolManager tool.ManagerInterface, resourceManager resource.ManagerInterface, isReload bool, doc *openapi3.T, serviceConfig *configv1.UpstreamServiceConfig) int {
	log := logging.GetLogger()
	numToolsForThisService := 0

	httpClient := u.getHTTPClient(serviceID)
	httpC := &httpClientImpl{client: httpClient}

	authenticator, err := auth.NewUpstreamAuthenticator(serviceConfig.GetUpstreamAuth())
	if err != nil {
		log.Error("Failed to create authenticator for OpenAPI upstream", "serviceID", serviceID, "error", err)
	}

	openapiService := serviceConfig.GetOpenapiService()
	definitions := openapiService.GetTools()
	calls := openapiService.GetCalls()

	// Initialize map of existing tools to allow overrides
	existingToolNames := make(map[string]bool)
	for _, d := range definitions {
		existingToolNames[d.GetName()] = true
	}

	// Auto-discover if explicitly enabled OR if no tools are defined (legacy behavior)
	if serviceConfig.GetAutoDiscoverTool() || len(definitions) == 0 {
		if calls == nil {
			calls = make(map[string]*configv1.OpenAPICallDefinition)
		}
		for _, tool := range pbTools {
			toolName := tool.GetName()
			// If tool is already defined manually, skip auto-discovery (override)
			if existingToolNames[toolName] {
				continue
			}

			// Create a default tool definition
			definitions = append(definitions, configv1.ToolDefinition_builder{
				Name:   proto.String(toolName),
				CallId: proto.String(toolName),
			}.Build())

			// Create a default call definition if it doesn't exist
			if _, ok := calls[toolName]; !ok {
				calls[toolName] = configv1.OpenAPICallDefinition_builder{
					Id: proto.String(toolName),
				}.Build()
			}
		}
	}

	for _, definition := range definitions {
		if definition.GetDisable() {
			log.Info("Skipping disabled tool", "toolName", definition.GetName())
			continue
		}

		callID := definition.GetCallId()
		callDef, ok := calls[callID]
		if !ok {
			log.Error("Call definition not found for tool", "call_id", callID, "tool_name", definition.GetName())
			continue
		}

		toolName := definition.GetName()

		var pbTool *pb.Tool
		for _, t := range pbTools {
			if t.GetName() == toolName {
				pbTool = t
				break
			}
		}
		if pbTool == nil {
			log.Error("Tool not found in parsed OpenAPI spec", "tool_name", toolName)
			continue
		}

		newToolProto := proto.Clone(pbTool).(*pb.Tool)
		newToolProto.SetName(toolName)
		newToolProto.SetServiceId(serviceID)

		if existingTool, exists := toolManager.GetTool(toolName); exists {
			if isReload {
				log.Warn("OpenAPI Tool already exists, overwriting during service reload.", "tool_id", toolName, "svc_id", serviceID)
			} else {
				log.Warn("Tool ID already exists from a different service registration. Skipping.", "tool_id", toolName, "svc_key", serviceID, "existing_svc_id", existingTool.Tool().GetServiceId())
				continue
			}
		}

		methodAndPath := strings.SplitN(pbTool.GetUnderlyingMethodFqn(), " ", 2)
		if len(methodAndPath) != 2 {
			log.Error("Invalid underlying method FQN", "fqn", pbTool.GetUnderlyingMethodFqn())
			continue
		}
		method, path := methodAndPath[0], methodAndPath[1]

		pathItem := doc.Paths.Find(path)
		if pathItem == nil {
			log.Error("Path not found in spec", "path", path)
			continue
		}

		op := pathItem.GetOperation(method)
		if op == nil {
			log.Error("Operation not found for method", "method", method, "path", path)
			continue
		}

		parameterDefs := make(map[string]string)
		for _, paramRef := range op.Parameters {
			if paramRef.Value != nil {
				parameterDefs[paramRef.Value.Name] = paramRef.Value.In
			}
		}

		var serverURL string
		if len(doc.Servers) > 0 {
			serverURL = doc.Servers[0].URL
		} else {
			serverURL = serviceConfig.GetOpenapiService().GetAddress()
		}
		fullURL := serverURL + path

		newTool := tool.NewOpenAPITool(newToolProto, httpC, parameterDefs, method, fullURL, authenticator, callDef)
		if err := toolManager.AddTool(newTool); err != nil {
			log.Error("Failed to add tool", "error", err)
			continue
		}
		numToolsForThisService++
		log.Info("Registered OpenAPI tool", "tool_id", toolName, "is_reload", isReload)
	}

	u.registerDynamicResources(serviceID, definitions, openapiService.GetResources(), resourceManager, toolManager)

	return numToolsForThisService
}

func (u *OpenAPIUpstream) createAndRegisterPrompts(
	_ context.Context,
	serviceID string,
	serviceConfig *configv1.UpstreamServiceConfig,
	promptManager prompt.ManagerInterface,
	isReload bool,
) {
	log := logging.GetLogger()
	openapiService := serviceConfig.GetOpenapiService()
	for _, promptDef := range openapiService.GetPrompts() {
		if promptDef.GetDisable() {
			log.Info("Skipping disabled prompt", "promptName", promptDef.GetName())
			continue
		}
		newPrompt := prompt.NewTemplatedPrompt(promptDef, serviceID)
		promptManager.AddPrompt(newPrompt)
		log.Info("Registered prompt", "prompt_name", newPrompt.Prompt().Name, "is_reload", isReload)
	}
}
func (u *OpenAPIUpstream) registerDynamicResources(
	serviceID string,
	definitions []*configv1.ToolDefinition,
	resources []*configv1.ResourceDefinition,
	resourceManager resource.ManagerInterface,
	toolManager tool.ManagerInterface,
) {
	log := logging.GetLogger()
	callIDToName := make(map[string]string)
	for _, d := range definitions {
		callIDToName[d.GetCallId()] = d.GetName()
	}
	for _, resourceDef := range resources {
		if resourceDef.GetDisable() {
			log.Info("Skipping disabled resource", "resourceName", resourceDef.GetName())
			continue
		}
		if resourceDef.GetDynamic() != nil {
			call := resourceDef.GetDynamic().GetHttpCall()
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
}
