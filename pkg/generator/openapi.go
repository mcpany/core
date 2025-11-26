// Copyright 2025 Author(s) of MCP Any
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package generator

import (
	"fmt"
	"log"
	"net/url"
	"strings"

	"github.com/getkin/kin-openapi/openapi2"
	"github.com/getkin/kin-openapi/openapi2conv"
	"github.com/getkin/kin-openapi/openapi3"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"google.golang.org/protobuf/encoding/protojson"
	"sigs.k8s.io/yaml"
)

// GenerateMCPConfigFromOpenAPI generates an MCP Any configuration from an OpenAPI specification.
func GenerateMCPConfigFromOpenAPI(specData []byte, source string) ([]byte, error) {
	var serverURL string
	var doc *openapi3.T

	// Attempt to parse as OpenAPI 2 first
	doc2 := &openapi2.T{}
	if err := yaml.Unmarshal(specData, doc2); err == nil {
		if doc2.Host != "" {
			scheme := "http"
			if len(doc2.Schemes) > 0 {
				scheme = doc2.Schemes[0]
			}
			serverURL = fmt.Sprintf("%s://%s", scheme, doc2.Host)
		}
	}

	// Try to load as OpenAPI 3
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData(specData)
	if err != nil {
		doc, err = openapi2conv.ToV3(doc2)
		if err != nil {
			return nil, fmt.Errorf("failed to convert OpenAPI 2 to 3: %w", err)
		}
	} else {
		if len(doc.Servers) > 0 {
			serverURL = doc.Servers[0].URL
		}
	}

	if serverURL == "" {
		return nil, fmt.Errorf("no servers found in the OpenAPI spec")
	}

	serviceName := doc.Info.Title
	if serviceName == "" {
		serviceName = "my-service"
	}

	u, err := url.Parse(serverURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse server URL: %w", err)
	}

	address := fmt.Sprintf("%s://%s", u.Scheme, u.Host)

	httpServiceBuilder := configv1.HttpUpstreamService_builder{
		Address: &address,
		Calls:   make(map[string]*configv1.HttpCallDefinition),
	}

	for path, pathItem := range doc.Paths.Map() {
		for method, operation := range pathItem.Operations() {
			if operation.OperationID == "" {
				log.Printf("warning: skipping operation with no operationId: %s %s", method, path)
				continue
			}

			httpMethod := configv1.HttpCallDefinition_HttpMethod(configv1.HttpCallDefinition_HttpMethod_value["HTTP_METHOD_"+strings.ToUpper(method)])
			callBuilder := configv1.HttpCallDefinition_builder{
				Method:       &httpMethod,
				EndpointPath: &path,
			}
			httpServiceBuilder.Calls[operation.OperationID] = callBuilder.Build()
		}
	}

	upstreamServiceBuilder := configv1.UpstreamServiceConfig_builder{
		Name:        &serviceName,
		HttpService: httpServiceBuilder.Build(),
	}

	cfg := &configv1.McpAnyServerConfig_builder{
		UpstreamServices: []*configv1.UpstreamServiceConfig{upstreamServiceBuilder.Build()},
	}

	jsonBytes, err := protojson.Marshal(cfg.Build())
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config to JSON: %w", err)
	}

	return yaml.JSONToYAML(jsonBytes)
}
