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

package openapi

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/jellydator/ttlcache/v3"
	"github.com/mcpxy/core/pkg/auth"
	"github.com/mcpxy/core/pkg/logging"
	"github.com/mcpxy/core/pkg/prompt"
	"github.com/mcpxy/core/pkg/resource"
	"github.com/mcpxy/core/pkg/tool"
	"github.com/mcpxy/core/pkg/upstream"
	"github.com/mcpxy/core/pkg/util"
	configv1 "github.com/mcpxy/core/proto/config/v1"
	pb "github.com/mcpxy/core/proto/mcp_router/v1"
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
) (string, []*configv1.ToolDefinition, error) {
	log := logging.GetLogger()
	serviceKey, err := util.GenerateServiceKey(serviceConfig.GetName())
	if err != nil {
		return "", nil, err
	}

	openapiService := serviceConfig.GetOpenapiService()
	if openapiService == nil {
		return "", nil, fmt.Errorf("openapi service config is nil")
	}

	info := &tool.ServiceInfo{
		Name:   serviceConfig.GetName(),
		Config: serviceConfig,
	}
	toolManager.AddServiceInfo(serviceKey, info)

	specContent, err := u.getSpecContent(ctx, openapiService)
	if err != nil {
		return "", nil, fmt.Errorf("failed to get OpenAPI spec content for service '%s': %w", serviceKey, err)
	}

	hash := sha256.Sum256(specContent)
	cacheKey := hex.EncodeToString(hash[:])

	item := u.openapiCache.Get(cacheKey)
	var doc *openapi3.T
	if item != nil {
		doc = item.Value()
	} else {
		var err error
		_, doc, err = parseOpenAPISpec(ctx, []byte(specContent))
		if err != nil {
			return "", nil, fmt.Errorf("failed to parse OpenAPI spec for service '%s' from content: %w", serviceKey, err)
		}
		u.openapiCache.Set(cacheKey, doc, ttlcache.DefaultTTL)
	}

	mcpOps := extractMcpOperationsFromOpenAPI(doc)
	pbTools := convertMcpOperationsToTools(mcpOps, doc, serviceKey)
	discoveredTools := make([]*configv1.ToolDefinition, 0, len(mcpOps))
	for _, op := range mcpOps {
		discoveredTools = append(discoveredTools, configv1.ToolDefinition_builder{
			Name:        proto.String(op.OperationID),
			Description: proto.String(op.Description),
		}.Build())
	}

	numToolsAdded := u.addOpenAPIToolsToIndex(ctx, pbTools, serviceKey, toolManager, isReload, doc, serviceConfig)
	log.Info("Registered OpenAPI service", "serviceKey", serviceKey, "toolsAdded", numToolsAdded)

	return serviceKey, discoveredTools, nil
}

// getHTTPClient retrieves or creates an HTTP client for a given service. It
// ensures that each service has its own dedicated client, which can be
// configured with specific transports or timeouts.
func (u *OpenAPIUpstream) getHTTPClient(serviceKey string) *http.Client {
	u.mu.Lock()
	defer u.mu.Unlock()

	if client, ok := u.httpClients[serviceKey]; ok {
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

	u.httpClients[serviceKey] = client
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

// getSpecContent determines the source of the OpenAPI spec and retrieves it.
// It can either be provided directly as a string or fetched from a URL.
func (u *OpenAPIUpstream) getSpecContent(ctx context.Context, openapiService *configv1.OpenapiUpstreamService) ([]byte, error) {
	if url := openapiService.GetOpenapiSpecUrl(); url != "" {
		return u.fetchSpecFromURL(ctx, url)
	}
	if spec := openapiService.GetOpenapiSpec(); spec != "" {
		return []byte(spec), nil
	}
	return nil, fmt.Errorf("no OpenAPI spec source provided")
}

// fetchSpecFromURL retrieves the content of an OpenAPI specification from a given URL.
func (u *OpenAPIUpstream) fetchSpecFromURL(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request for spec from URL %s: %w", url, err)
	}

	// Use a general-purpose HTTP client for fetching the spec.
	client := u.getHTTPClient("spec-fetcher")
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch spec from URL %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d when fetching spec from URL %s", resp.StatusCode, url)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body from URL %s: %w", url, err)
	}

	return body, nil
}

// addOpenAPIToolsToIndex iterates through a list of protobuf tool definitions,
// creates an OpenAPITool for each, and registers it with the tool manager.
func (u *OpenAPIUpstream) addOpenAPIToolsToIndex(ctx context.Context, pbTools []*pb.Tool, serviceKey string, toolManager tool.ToolManagerInterface, isReload bool, doc *openapi3.T, serviceConfig *configv1.UpstreamServiceConfig) int {
	log := logging.GetLogger()
	numToolsForThisService := 0

	httpClient := u.getHTTPClient(serviceKey)
	httpC := &httpClientImpl{client: httpClient}

	authenticator, err := auth.NewUpstreamAuthenticator(serviceConfig.GetUpstreamAuthentication())
	if err != nil {
		log.Error("Failed to create authenticator for OpenAPI upstream", "serviceKey", serviceKey, "error", err)
	}

	openapiService := serviceConfig.GetOpenapiService()
	callDefs := openapiService.GetCalls()
	callDefMap := make(map[string]*configv1.OpenAPICallDefinition)
	for _, def := range callDefs {
		callDefMap[def.GetSchema().GetName()] = def
	}

	for _, t := range pbTools {
		toolName := t.GetName()

		newToolProto := proto.Clone(t).(*pb.Tool)
		newToolProto.SetName(toolName)
		newToolProto.SetServiceId(serviceKey)

		if existingTool, exists := toolManager.GetTool(toolName); exists {
			if isReload {
				log.Warn("OpenAPI Tool already exists, overwriting during service reload.", "tool_id", toolName, "svc_id", serviceKey)
			} else {
				log.Warn("Tool ID already exists from a different service registration. Skipping.", "tool_id", toolName, "svc_key", serviceKey, "existing_svc_id", existingTool.Tool().GetServiceId())
				continue
			}
		}

		methodAndPath := strings.Fields(t.GetUnderlyingMethodFqn())
		if len(methodAndPath) != 2 {
			log.Error("Invalid underlying method FQN", "fqn", t.GetUnderlyingMethodFqn())
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

		callDef, ok := callDefMap[toolName]
		if !ok {
			callDef = &configv1.OpenAPICallDefinition{}
		}

		newTool := tool.NewOpenAPITool(newToolProto, httpC, parameterDefs, method, fullURL, authenticator, callDef)
		if err := toolManager.AddTool(newTool); err != nil {
			log.Error("Failed to add tool", "error", err)
			continue
		}
		numToolsForThisService++
		log.Info("Registered OpenAPI tool", "tool_id", toolName, "is_reload", isReload)
	}

	return numToolsForThisService
}
