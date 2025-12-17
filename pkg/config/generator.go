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
// It prompts the user for input and uses templates to generate YAML configuration.
type Generator struct {
	Reader *bufio.Reader
}

// NewGenerator creates a new Generator instance that reads from standard input.
func NewGenerator() *Generator {
	return &Generator{
		Reader: bufio.NewReader(os.Stdin),
	}
}

// Generate prompts the user for service details and returns the generated
// configuration as a byte slice. It supports multiple service types including
// HTTP, gRPC, OpenAPI, and GraphQL.
func (g *Generator) Generate() ([]byte, error) {
	serviceType, err := g.prompt("Enter service type (http, grpc, openapi, graphql): ")
	if err != nil {
		return nil, err
	}

	switch serviceType {
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

func (g *Generator) prompt(prompt string) (string, error) {
	fmt.Print(prompt)
	input, err := g.Reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(input), nil
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
type HTTPServiceData struct {
	Name         string
	Address      string
	OperationID  string
	Description  string
	Method       string
	EndpointPath string
}

func (g *Generator) generateHTTPService() ([]byte, error) {
	data := HTTPServiceData{}
	var err error

	data.Name, err = g.prompt("Enter service name: ")
	if err != nil {
		return nil, err
	}

	data.Address, err = g.prompt("Enter service address: ")
	if err != nil {
		return nil, err
	}

	data.OperationID, err = g.prompt("Enter operation ID: ")
	if err != nil {
		return nil, err
	}

	data.Description, err = g.prompt("Enter description: ")
	if err != nil {
		return nil, err
	}

	data.Method, err = g.prompt("Enter HTTP method (e.g., HTTP_METHOD_GET): ")
	if err != nil {
		return nil, err
	}

	data.EndpointPath, err = g.prompt("Enter endpoint path: ")
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
type GRPCServiceData struct {
	Name              string
	Address           string
	ReflectionEnabled bool
}

func (g *Generator) generateGRPCService() ([]byte, error) {
	data := GRPCServiceData{}
	var err error

	data.Name, err = g.prompt("Enter service name: ")
	if err != nil {
		return nil, err
	}

	data.Address, err = g.prompt("Enter service address: ")
	if err != nil {
		return nil, err
	}

	for {
		reflectionEnabled, err := g.prompt("Enable reflection (true/false): ")
		if err != nil {
			return nil, err
		}
		if reflectionEnabled == "true" {
			data.ReflectionEnabled = true
			break
		}
		if reflectionEnabled == "false" {
			data.ReflectionEnabled = false
			break
		}
		fmt.Println("Invalid input. Please enter 'true' or 'false'.")
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
type OpenAPIServiceData struct {
	Name     string
	SpecPath string
}

func (g *Generator) generateOpenAPIService() ([]byte, error) {
	data := OpenAPIServiceData{}
	var err error

	data.Name, err = g.prompt("Enter service name: ")
	if err != nil {
		return nil, err
	}

	data.SpecPath, err = g.prompt("Enter OpenAPI spec path: ")
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
type GraphQLServiceData struct {
	Name         string
	Address      string
	CallName     string
	SelectionSet string
}

func (g *Generator) generateGraphQLService() ([]byte, error) {
	data := GraphQLServiceData{}
	var err error

	data.Name, err = g.prompt("Enter service name: ")
	if err != nil {
		return nil, err
	}

	data.Address, err = g.prompt("Enter service address: ")
	if err != nil {
		return nil, err
	}

	data.CallName, err = g.prompt("Enter call name: ")
	if err != nil {
		return nil, err
	}

	data.SelectionSet, err = g.prompt("Enter selection set: ")
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
