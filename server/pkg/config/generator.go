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

// Generator represents the public Generator entity.
//
// Summary: Defines the structured data model representing a .
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
type Generator struct {
	Reader *bufio.Reader
}

// NewGenerator serves as a public interface for interacting with NewGenerator.
//
// Summary: Constructs and returns an initialized generator ready for consumption.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func NewGenerator() *Generator {
	return &Generator{
		Reader: bufio.NewReader(os.Stdin),
	}
}

// Generate serves as a public interface for interacting with Generate.
//
// Summary: Generate the  appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
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

// HTTPServiceData represents the public HTTPServiceData entity.
//
// Summary: Defines the structured data model representing a service data.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
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

// GRPCServiceData represents the public GRPCServiceData entity.
//
// Summary: Defines the structured data model representing a service data.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
type GRPCServiceData struct {
	Name              string
	Address           string
	ReflectionEnabled bool
}

func (g *Generator) generateGRPCService() ([]byte, error) {
	data := GRPCServiceData{}
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

// OpenAPIServiceData represents the public OpenAPIServiceData entity.
//
// Summary: Defines the structured data model representing a api service data.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
type OpenAPIServiceData struct {
	Name     string
	SpecPath string
}

func (g *Generator) generateOpenAPIService() ([]byte, error) {
	data := OpenAPIServiceData{}
	var err error

	data.Name, err = g.prompt("🏷️  Enter service name: ")
	if err != nil {
		return nil, err
	}

	data.SpecPath, err = g.prompt("📂 Enter OpenAPI spec path: ")
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

// GraphQLServiceData represents the public GraphQLServiceData entity.
//
// Summary: Defines the structured data model representing a ql service data.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
type GraphQLServiceData struct {
	Name         string
	Address      string
	CallName     string
	SelectionSet string
}

func (g *Generator) generateGraphQLService() ([]byte, error) {
	data := GraphQLServiceData{}
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
