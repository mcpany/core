// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strings"
	"text/template"
)

// Generator handles the interactive generation of configuration files.
// It prompts the user for input and uses templates to generate YAML configuration
// for different types of services (HTTP, gRPC, OpenAPI, GraphQL).
type Generator struct {
	Reader *bufio.Reader
}

// NewGenerator creates a new Generator instance that reads from standard input.
//
// Returns:
//   - A pointer to a new Generator initialized with os.Stdin.
func NewGenerator() *Generator {
	return &Generator{
		Reader: bufio.NewReader(os.Stdin),
	}
}

// Generate prompts the user for service details and returns the generated
// configuration as a byte slice. It supports multiple service types including
// HTTP, gRPC, OpenAPI, and GraphQL.
//
// Returns:
//   - A byte slice containing the generated YAML configuration.
//   - An error if the generation fails or the user provides invalid input.
func (g *Generator) Generate() ([]byte, error) {
	serviceType, err := g.prompt("🤖 Enter service type (http, grpc, openapi, graphql): ")
	if err != nil {
		return nil, err
	}

	switch strings.ToLower(serviceType) {
	case "http":
		return g.generateHTTPService()
	case "grpc":
		return g.generateGRPCService()
	case "openapi":
		return g.generateOpenAPIService()
	case "graphql":
		return g.generateGraphQLService()
	default:
		return nil, fmt.Errorf("unsupported service type: %s", serviceType)
	}
}

// FullConfigData holds the data for the full configuration.
type FullConfigData struct {
	ListenAddress  string
	LogLevel       string
	ServiceType    string
	ServiceHTTP    *HTTPServiceData
	ServiceGRPC    *GRPCServiceData
	ServiceOpenAPI *OpenAPIServiceData
	ServiceGraphQL *GraphQLServiceData
}

const fullConfigTemplate = `# Copyright 2025 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

global_settings:
  mcp_listen_address: "{{ .ListenAddress }}"
  log_level: "{{ .LogLevel }}"

upstream_services:
{{- if eq .ServiceType "http" }}
  - name: "{{ .ServiceHTTP.Name }}"
    http_service:
      address: "{{ .ServiceHTTP.Address }}"
      calls:
        - operationId: "{{ .ServiceHTTP.OperationID }}"
          description: "{{ .ServiceHTTP.Description }}"
          method: "{{ .ServiceHTTP.Method }}"
          endpointPath: "{{ .ServiceHTTP.EndpointPath }}"
{{- end }}
{{- if eq .ServiceType "grpc" }}
  - name: "{{ .ServiceGRPC.Name }}"
    grpc_service:
      address: "{{ .ServiceGRPC.Address }}"
      reflection:
        enabled: {{ .ServiceGRPC.ReflectionEnabled }}
{{- end }}
{{- if eq .ServiceType "openapi" }}
  - name: "{{ .ServiceOpenAPI.Name }}"
    openapi_service:
      spec:
        path: "{{ .ServiceOpenAPI.SpecPath }}"
{{- end }}
{{- if eq .ServiceType "graphql" }}
  - name: "{{ .ServiceGraphQL.Name }}"
    graphql_service:
      address: "{{ .ServiceGraphQL.Address }}"
      calls:
        - name: "{{ .ServiceGraphQL.CallName }}"
          selectionSet: "{{ .ServiceGraphQL.SelectionSet }}"
{{- end }}
`

// GenerateFull prompts the user for global settings and service details
// and returns the generated full configuration as a byte slice.
func (g *Generator) GenerateFull() ([]byte, error) {
	data := FullConfigData{}
	var err error

	fmt.Println("👋 Welcome to MCP Any! Let's set up your configuration.")

	// Global Settings
	data.ListenAddress, err = g.prompt("🌐 Enter listen address [0.0.0.0:50050]: ")
	if err != nil {
		return nil, err
	}
	if data.ListenAddress == "" {
		data.ListenAddress = "0.0.0.0:50050"
	}

	data.LogLevel, err = g.prompt("📊 Enter log level (DEBUG, INFO, WARN, ERROR) [INFO]: ")
	if err != nil {
		return nil, err
	}
	if data.LogLevel == "" {
		data.LogLevel = "INFO"
	}
	data.LogLevel = strings.ToUpper(data.LogLevel)

	addService, err := g.promptBool("➕ Do you want to add a service now?", true)
	if err != nil {
		return nil, err
	}

	if addService {
		serviceType, err := g.prompt("🤖 Enter service type (http, grpc, openapi, graphql): ")
		if err != nil {
			return nil, err
		}
		data.ServiceType = strings.ToLower(serviceType)

		switch data.ServiceType {
		case "http":
			data.ServiceHTTP, err = g.promptHTTPServiceData()
		case "grpc":
			data.ServiceGRPC, err = g.promptGRPCServiceData()
		case "openapi":
			data.ServiceOpenAPI, err = g.promptOpenAPIServiceData()
		case "graphql":
			data.ServiceGraphQL, err = g.promptGraphQLServiceData()
		default:
			fmt.Printf("⚠️  Unsupported service type: %s. Skipping service addition.\n", serviceType)
			data.ServiceType = ""
		}
		if err != nil {
			return nil, err
		}
	}

	tmpl, err := template.New("fullConfig").Parse(fullConfigTemplate)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (g *Generator) prompt(prompt string) (string, error) {
	fmt.Print(prompt)
	input, err := g.Reader.ReadString('\n')
	if err != nil && len(input) == 0 {
		return "", err
	}
	return strings.TrimSpace(input), nil
}

// promptBool prompts the user for a boolean value.
// It accepts "true", "t", "yes", "y" as true, and "false", "f", "no", "n" as false.
// It is case-insensitive.
func (g *Generator) promptBool(prompt string, defaultValue bool) (bool, error) {
	defStr := "y"
	if !defaultValue {
		defStr = "n"
	}
	// append default to prompt
	fullPrompt := fmt.Sprintf("%s [%s]: ", prompt, defStr)

	for {
		input, err := g.prompt(fullPrompt)
		if err != nil {
			return false, err
		}

		if input == "" {
			return defaultValue, nil
		}

		lower := strings.ToLower(input)
		switch lower {
		case "true", "t", "yes", "y":
			return true, nil
		case "false", "f", "no", "n":
			return false, nil
		default:
			fmt.Println("❌ Invalid input. Please enter 'y' or 'n'.")
		}
	}
}

const httpServiceTemplate = `upstreamServices:
  - name: "{{ .Name }}"
    httpService:
      address: "{{ .Address }}"
      calls:
        - operationId: "{{ .OperationID }}"
          description: "{{ .Description }}"
          method: "{{ .Method }}"
          endpointPath: "{{ .EndpointPath }}"
`

// HTTPServiceData holds the data required to generate an HTTP service configuration.
// It is used as the data context for the httpServiceTemplate.
type HTTPServiceData struct {
	// Name is the name of the service.
	Name string
	// Address is the base URL/address of the service.
	Address string
	// OperationID is the unique identifier for the operation.
	OperationID string
	// Description is a human-readable description of the service operation.
	Description string
	// Method is the HTTP method to use (e.g., "GET", "POST").
	Method string
	// EndpointPath is the path of the endpoint (e.g., "/api/v1/users").
	EndpointPath string
}

func (g *Generator) promptHTTPServiceData() (*HTTPServiceData, error) {
	data := &HTTPServiceData{}
	var err error

	data.Name, err = g.prompt("🏷️  Enter service name: ")
	if err != nil {
		return nil, err
	}

	data.Address, err = g.prompt("🔗 Enter service address: ")
	if err != nil {
		return nil, err
	}

	data.OperationID, err = g.prompt("🆔 Enter operation ID: ")
	if err != nil {
		return nil, err
	}

	data.Description, err = g.prompt("📝 Enter description: ")
	if err != nil {
		return nil, err
	}

	methodInput, err := g.prompt("📡 Enter HTTP method (e.g., HTTP_METHOD_GET): ")
	if err != nil {
		return nil, err
	}
	data.Method = normalizeHTTPMethod(methodInput)

	data.EndpointPath, err = g.prompt("🛣️  Enter endpoint path: ")
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (g *Generator) generateHTTPService() ([]byte, error) {
	data, err := g.promptHTTPServiceData()
	if err != nil {
		return nil, err
	}

	tmpl, err := template.New("httpService").Parse(httpServiceTemplate)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

const grpcServiceTemplate = `upstreamServices:
  - name: "{{ .Name }}"
    grpcService:
      address: "{{ .Address }}"
      reflection:
        enabled: {{ .ReflectionEnabled }}
`

// GRPCServiceData holds the data required to generate a gRPC service configuration.
// It is used as the data context for the grpcServiceTemplate.
type GRPCServiceData struct {
	// Name is the name of the service.
	Name string
	// Address is the address of the gRPC service (host:port).
	Address string
	// ReflectionEnabled indicates whether gRPC reflection should be enabled.
	ReflectionEnabled bool
}

func (g *Generator) promptGRPCServiceData() (*GRPCServiceData, error) {
	data := &GRPCServiceData{}
	var err error

	data.Name, err = g.prompt("🏷️  Enter service name: ")
	if err != nil {
		return nil, err
	}

	data.Address, err = g.prompt("🔗 Enter service address: ")
	if err != nil {
		return nil, err
	}

	data.ReflectionEnabled, err = g.promptBool("🪞 Enable reflection", true)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (g *Generator) generateGRPCService() ([]byte, error) {
	data, err := g.promptGRPCServiceData()
	if err != nil {
		return nil, err
	}

	tmpl, err := template.New("grpcService").Parse(grpcServiceTemplate)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

const openapiServiceTemplate = `upstreamServices:
  - name: "{{ .Name }}"
    openapiService:
      spec:
        path: "{{ .SpecPath }}"
`

// OpenAPIServiceData holds the data required to generate an OpenAPI service configuration.
// It is used as the data context for the openapiServiceTemplate.
type OpenAPIServiceData struct {
	// Name is the name of the service.
	Name string
	// SpecPath is the path or URL to the OpenAPI specification file.
	SpecPath string
}

func (g *Generator) promptOpenAPIServiceData() (*OpenAPIServiceData, error) {
	data := &OpenAPIServiceData{}
	var err error

	data.Name, err = g.prompt("🏷️  Enter service name: ")
	if err != nil {
		return nil, err
	}

	data.SpecPath, err = g.prompt("📂 Enter OpenAPI spec path: ")
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (g *Generator) generateOpenAPIService() ([]byte, error) {
	data, err := g.promptOpenAPIServiceData()
	if err != nil {
		return nil, err
	}

	tmpl, err := template.New("openapiService").Parse(openapiServiceTemplate)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

const graphqlServiceTemplate = `upstreamServices:
  - name: "{{ .Name }}"
    graphqlService:
      address: "{{ .Address }}"
      calls:
        - name: "{{ .CallName }}"
          selectionSet: "{{ .SelectionSet }}"
`

// GraphQLServiceData holds the data required to generate a GraphQL service configuration.
// It is used as the data context for the graphqlServiceTemplate.
type GraphQLServiceData struct {
	// Name is the name of the service.
	Name string
	// Address is the URL of the GraphQL endpoint.
	Address string
	// CallName is the name of the GraphQL query or mutation to expose.
	CallName string
	// SelectionSet is the GraphQL selection set for the operation.
	SelectionSet string
}

func (g *Generator) promptGraphQLServiceData() (*GraphQLServiceData, error) {
	data := &GraphQLServiceData{}
	var err error

	data.Name, err = g.prompt("🏷️  Enter service name: ")
	if err != nil {
		return nil, err
	}

	data.Address, err = g.prompt("🔗 Enter service address: ")
	if err != nil {
		return nil, err
	}

	data.CallName, err = g.prompt("📞 Enter call name: ")
	if err != nil {
		return nil, err
	}

	data.SelectionSet, err = g.prompt("📋 Enter selection set: ")
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (g *Generator) generateGraphQLService() ([]byte, error) {
	data, err := g.promptGraphQLServiceData()
	if err != nil {
		return nil, err
	}

	tmpl, err := template.New("graphqlService").Parse(graphqlServiceTemplate)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
