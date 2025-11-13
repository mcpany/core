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

package openapi

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/jellydator/ttlcache/v3"
	"github.com/mcpany/core/pkg/auth"
	"github.com/mcpany/core/pkg/logging"
	"github.com/mcpany/core/pkg/prompt"
	"github.com/mcpany/core/pkg/resource"
	"github.com/mcpany/core/pkg/tool"
	"github.com/mcpany/core/pkg/upstream"
	"github.com/mcpany/core/pkg/util"
	configv1 "github.com/mcpany/core/proto/config/v1"
	pb "github.com/mcpany/core/proto/mcp_router/v1"
	"google.golang.org/protobuf/proto"
)

// OpenAPIUpstream implements the upstream.Upstream interface for services that
// are defined by an OpenAPI specification. It parses the spec, discovers the
// available operations, and registers them as tools.
type OpenAPIUpstream struct {
	openapiCache *ttlcache.Cache[string, *openapi3.T]
	httpClients  map[string]*http.Client
	mu           sync.Mutex
}

// NewOpenAPIUpstream creates a new instance of OpenAPIUpstream. It initializes a
// cache for storing parsed OpenAPI documents to avoid redundant parsing.
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

// Register processes an OpenAPI service configuration. It parses the OpenAPI
// specification, extracts the operations, converts them into tools, and
// registers them with the tool manager.
func (u *OpenAPIUpstream) Register(
	ctx context.Context,
	serviceConfig *configv1.UpstreamServiceConfig,
	toolManager tool.ToolManagerInterface,
	promptManager prompt.PromptManagerInterface,
	resourceManager resource.ResourceManagerInterface,
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

	serviceID := sanitizedName // for internal use

	openapiService := serviceConfig.GetOpenapiService()
	if openapiService == nil {
		return "", nil, nil, fmt.Errorf("openapi service config is nil")
	}

	info := &tool.ServiceInfo{
		Name:   serviceConfig.GetName(),
		Config: serviceConfig,
	}
	toolManager.AddServiceInfo(serviceID, info)

	specContent := openapiService.GetOpenapiSpec()
	if specContent == "" {
		return "", nil, nil, fmt.Errorf("OpenAPI spec content is missing for service %s", serviceID)
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

	transport := &http.Transport{
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
// client.HttpClient interface. This is used to adapt the standard library's
// HTTP client for use in components that expect this interface.
type httpClientImpl struct {
	client *http.Client
}

// Do sends an HTTP request and returns an HTTP response, fulfilling the
// client.HttpClient interface.
func (c *httpClientImpl) Do(req *http.Request) (*http.Response, error) {
	return c.client.Do(req)
}

// addOpenAPIToolsToIndex iterates through a list of protobuf tool definitions,
// creates an OpenAPITool for each, and registers it with the tool manager.
func (u *OpenAPIUpstream) addOpenAPIToolsToIndex(ctx context.Context, pbTools []*pb.Tool, serviceID string, toolManager tool.ToolManagerInterface, resourceManager resource.ResourceManagerInterface, isReload bool, doc *openapi3.T, serviceConfig *configv1.UpstreamServiceConfig) int {
	log := logging.GetLogger()
	numToolsForThisService := 0

	httpClient := u.getHTTPClient(serviceID)
	httpC := &httpClientImpl{client: httpClient}

	authenticator, err := auth.NewUpstreamAuthenticator(serviceConfig.GetUpstreamAuthentication())
	if err != nil {
		log.Error("Failed to create authenticator for OpenAPI upstream", "serviceID", serviceID, "error", err)
	}

	openapiService := serviceConfig.GetOpenapiService()
	definitions := openapiService.GetTools()
	calls := openapiService.GetCalls()

	for _, toolDefinition := range definitions {
		schema := toolDefinition.GetSchema()
		callID := toolDefinition.GetCallId()
		callDef, ok := calls[callID]
		if !ok {
			log.Error("Call definition not found for tool", "call_id", callID, "tool_name", schema.GetName())
			continue
		}

		toolName := schema.GetName()

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

		methodAndPath := strings.Fields(pbTool.GetUnderlyingMethodFqn())
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

		serverURL := ""
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

	for _, resourceDef := range openapiService.GetResources() {
		if resourceDef.GetDynamic() != nil {
			toolName := resourceDef.GetDynamic().GetHttpCall().GetSchema().GetName()
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

	return numToolsForThisService
}
