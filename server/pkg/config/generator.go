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
// for different types of services (HTTP, gRPC, OpenAPI, GraphQL, Command).
type Generator struct {
	Reader *bufio.Reader
}

// ConfigData holds the list of generated services to be rendered in the final config.
type ConfigData struct {
	UpstreamServices []string
}

// NewGenerator creates a new Generator instance.
//
// Parameters:
//   - reader: A *bufio.Reader to read input from. If nil, defaults to os.Stdin.
//
// Returns:
//   - A pointer to a new Generator.
func NewGenerator(reader *bufio.Reader) *Generator {
	if reader == nil {
		reader = bufio.NewReader(os.Stdin)
	}
	return &Generator{
		Reader: reader,
	}
}

// Generate prompts the user for service details and returns the generated
// configuration as a byte slice. It supports multiple service types including
// HTTP, gRPC, OpenAPI, GraphQL, and Command.
//
// Returns:
//   - A byte slice containing the generated YAML configuration.
//   - An error if the generation fails or the user provides invalid input.
func (g *Generator) Generate() ([]byte, error) {
	configData := ConfigData{
		UpstreamServices: []string{},
	}

	fmt.Println("Welcome to the MCP Any Configuration Wizard! 🧙‍♂️")
	fmt.Println("Let's set up your upstream services.")

	for {
		fmt.Println("\n--- New Service ---")
		serviceType, err := g.prompt("🤖 Enter service type (http, grpc, openapi, graphql, command) [or 'done' to finish]: ")
		if err != nil {
			return nil, err
		}

		serviceType = strings.ToLower(serviceType)
		if serviceType == "done" || serviceType == "" {
			if len(configData.UpstreamServices) > 0 {
				break
			}
			if serviceType == "done" {
				break // Allow exiting with empty config if explicitly requested
			}
			// If empty input and no services, just continue (loop)
			continue
		}

		var serviceConfig []byte
		switch serviceType {
		case "http":
			serviceConfig, err = g.generateHTTPService()
		case "grpc":
			serviceConfig, err = g.generateGRPCService()
		case "openapi":
			serviceConfig, err = g.generateOpenAPIService()
		case "graphql":
			serviceConfig, err = g.generateGraphQLService()
		case "command", "cmd", "stdio":
			serviceConfig, err = g.generateCommandService()
		default:
			fmt.Printf("❌ Unsupported service type: %s\n", serviceType)
			continue
		}

		if err != nil {
			return nil, err
		}

		configData.UpstreamServices = append(configData.UpstreamServices, string(serviceConfig))
		fmt.Println("✅ Service added!")

		// Ask if they want to add another one
		addMore, err := g.promptBool("➕ Add another service?", false)
		if err != nil {
			return nil, err
		}
		if !addMore {
			break
		}
	}

	tmpl, err := template.New("config").Parse(configTemplate)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, configData); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (g *Generator) prompt(prompt string) (string, error) {
	fmt.Print(prompt)
	input, err := g.Reader.ReadString('\n')
	if err != nil {
		if len(input) > 0 {
			// If we got some input but then error (e.g. EOF without newline), return the input
			return strings.TrimSpace(input), nil
		}
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

const configTemplate = `upstreamServices:
{{- range .UpstreamServices }}
{{ . }}
{{- end }}
`

const httpServiceTemplate = `  - name: "{{ .Name }}"
    httpService:
      address: "{{ .Address }}"
      calls:
        - operationId: "{{ .OperationID }}"
          description: "{{ .Description }}"
          method: "{{ .Method }}"
          endpointPath: "{{ .EndpointPath }}"`

// HTTPServiceData holds the data required to generate an HTTP service configuration.
// It is used as the data context for the httpServiceTemplate.
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

const grpcServiceTemplate = `  - name: "{{ .Name }}"
    grpcService:
      address: "{{ .Address }}"
      reflection:
        enabled: {{ .ReflectionEnabled }}`

// GRPCServiceData holds the data required to generate a gRPC service configuration.
// It is used as the data context for the grpcServiceTemplate.
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

const openapiServiceTemplate = `  - name: "{{ .Name }}"
    openapiService:
      spec:
        path: "{{ .SpecPath }}"`

// OpenAPIServiceData holds the data required to generate an OpenAPI service configuration.
// It is used as the data context for the openapiServiceTemplate.
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

const graphqlServiceTemplate = `  - name: "{{ .Name }}"
    graphqlService:
      address: "{{ .Address }}"
      calls:
        - name: "{{ .CallName }}"
          selectionSet: "{{ .SelectionSet }}"`

// GraphQLServiceData holds the data required to generate a GraphQL service configuration.
// It is used as the data context for the graphqlServiceTemplate.
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

const commandServiceTemplate = `  - name: "{{ .Name }}"
    command:
      command: "{{ .Command }}"
      args:
{{- range .Args }}
        - "{{ . }}"
{{- end }}
      env:
{{- range $key, $value := .Env }}
        {{ $key }}: "{{ $value }}"
{{- end }}`

// CommandServiceData holds the data required to generate a Command (stdio) service configuration.
type CommandServiceData struct {
	Name    string
	Command string
	Args    []string
	Env     map[string]string
}

func (g *Generator) generateCommandService() ([]byte, error) {
	data := CommandServiceData{
		Env: make(map[string]string),
	}
	var err error

	data.Name, err = g.prompt("🏷️  Enter service name: ")
	if err != nil {
		return nil, err
	}

	data.Command, err = g.prompt("💻 Enter command (e.g. npx, python, /path/to/binary): ")
	if err != nil {
		return nil, err
	}

	argsInput, err := g.prompt("📥 Enter arguments (space separated, or leave empty): ")
	if err != nil {
		return nil, err
	}
	if argsInput != "" {
		// Simple space splitting, handling quotes would be better but this is a simple wizard
		// For now let's just split by space
		data.Args = strings.Fields(argsInput)
	}

	// Env vars loop
	for {
		addEnv, err := g.promptBool("🌱 Add environment variable?", false)
		if err != nil {
			return nil, err
		}
		if !addEnv {
			break
		}
		key, err := g.prompt("  Key: ")
		if err != nil {
			return nil, err
		}
		val, err := g.prompt("  Value: ")
		if err != nil {
			return nil, err
		}
		if key != "" {
			data.Env[key] = val
		}
	}

	tmpl, err := template.New("commandService").Parse(commandServiceTemplate)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
