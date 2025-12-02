// Copyright 2024 Author(s) of MCP Any
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

// Package scaffold provides functionality for scaffolding MCP Any configurations.
package scaffold

import (
	"encoding/json"
	"fmt"
	"os"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/getkin/kin-openapi/openapi3"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protoreflect"
	"gopkg.in/yaml.v3"
)

// OpenAPIScaffolder is a scaffolder for OpenAPI specifications.
type OpenAPIScaffolder struct {
	spec *openapi3.T
}

// NewOpenAPIScaffolder creates a new OpenAPIScaffolder.
func NewOpenAPIScaffolder(spec *openapi3.T) *OpenAPIScaffolder {
	return &OpenAPIScaffolder{
		spec: spec,
	}
}

// Scaffold scaffolds an MCP Any configuration from an OpenAPI specification.
func (s *OpenAPIScaffolder) Scaffold() (*configv1.McpAnyServerConfig, error) {
	if s.spec == nil {
		return nil, fmt.Errorf("OpenAPI spec is nil")
	}

	serviceConfigMsg := (&configv1.UpstreamServiceConfig{}).ProtoReflect()

	if s.spec.Info != nil {
		nameField := serviceConfigMsg.Descriptor().Fields().ByName("name")
		serviceConfigMsg.Set(nameField, protoreflect.ValueOfString(s.spec.Info.Title))
	}

	openapiServiceMsg := (&configv1.OpenapiUpstreamService{}).ProtoReflect()
	if len(s.spec.Servers) > 0 {
		addressField := openapiServiceMsg.Descriptor().Fields().ByName("address")
		openapiServiceMsg.Set(addressField, protoreflect.ValueOfString(s.spec.Servers[0].URL))
	}

	oneofField := serviceConfigMsg.Descriptor().Fields().ByName("openapi_service")
	serviceConfigMsg.Set(oneofField, protoreflect.ValueOfMessage(openapiServiceMsg))

	configMsg := (&configv1.McpAnyServerConfig{}).ProtoReflect()
	servicesField := configMsg.Descriptor().Fields().ByName("upstream_services")
	list := configMsg.Mutable(servicesField).List()
	list.Append(protoreflect.ValueOfMessage(serviceConfigMsg))

	return configMsg.Interface().(*configv1.McpAnyServerConfig), nil
}

// ScaffoldFile scaffolds an MCP Any configuration from an OpenAPI specification file.
func ScaffoldFile(openapiFile string, outputFile string) error {
	data, err := os.ReadFile(openapiFile)
	if err != nil {
		return fmt.Errorf("failed to read OpenAPI file: %w", err)
	}

	spec, err := openapi3.NewLoader().LoadFromData(data)
	if err != nil {
		return fmt.Errorf("failed to load OpenAPI spec: %w", err)
	}

	scaffolder := NewOpenAPIScaffolder(spec)
	config, err := scaffolder.Scaffold()
	if err != nil {
		return fmt.Errorf("failed to scaffold config: %w", err)
	}

	// Marshal to JSON first, as the yaml marshaller doesn't handle protobuf structs well.
	jsonBytes, err := protojson.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config to JSON: %w", err)
	}

	// Unmarshal JSON into a generic interface
	var genericConfig interface{}
	if err := json.Unmarshal(jsonBytes, &genericConfig); err != nil {
		return fmt.Errorf("failed to unmarshal JSON to generic interface: %w", err)
	}

	// Marshal the generic interface to YAML
	yamlData, err := yaml.Marshal(genericConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal config to YAML: %w", err)
	}

	if err := os.WriteFile(outputFile, yamlData, 0644); err != nil {
		return fmt.Errorf("failed to write config to file: %w", err)
	}

	return nil
}
