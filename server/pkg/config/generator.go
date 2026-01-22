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

// AuthData holds the data for authentication configuration.
type AuthData struct {
	Type              string
	ParamName         string // For API Key
	In                string // For API Key
	Value             string // For API Key, Bearer Token
	Username          string // For Basic Auth
	Password          string // For Basic Auth
	TokenURL          string // For OAuth2
	ClientID          string // For OAuth2
	ClientSecret      string // For OAuth2
	Scopes            string // For OAuth2
	HeaderName        string // For Trusted Header
	HeaderValue       string // For Trusted Header
}

func (g *Generator) generateAuth() (*AuthData, error) {
	authType, err := g.prompt("🔒 Enter authentication type (none, apikey, bearer, basic, oauth2, header) [none]: ")
	if err != nil {
		return nil, err
	}

	if authType == "" || strings.ToLower(authType) == "none" {
		return nil, nil
	}

	data := &AuthData{Type: strings.ToLower(authType)}

	switch data.Type {
	case "apikey":
		data.ParamName, err = g.prompt("🔑 Enter API key param name (e.g. X-API-Key): ")
		if err != nil { return nil, err }

		data.In, err = g.prompt("📍 Enter API key location (HEADER, QUERY, COOKIE) [HEADER]: ")
		if err != nil { return nil, err }
		if data.In == "" { data.In = "HEADER" }

		data.Value, err = g.prompt("🤫 Enter API key value (or env var like ${API_KEY}): ")
		if err != nil { return nil, err }

	case "bearer":
		data.Value, err = g.prompt("🤫 Enter bearer token (or env var like ${TOKEN}): ")
		if err != nil { return nil, err }

	case "basic":
		data.Username, err = g.prompt("👤 Enter username: ")
		if err != nil { return nil, err }
		data.Password, err = g.prompt("🤫 Enter password (or env var like ${PASSWORD}): ")
		if err != nil { return nil, err }

	case "oauth2":
		data.TokenURL, err = g.prompt("🌐 Enter token URL: ")
		if err != nil { return nil, err }
		data.ClientID, err = g.prompt("🆔 Enter client ID: ")
		if err != nil { return nil, err }
		data.ClientSecret, err = g.prompt("🤫 Enter client secret: ")
		if err != nil { return nil, err }
		data.Scopes, err = g.prompt("🔭 Enter scopes (space separated): ")
		if err != nil { return nil, err }

	case "header":
		data.HeaderName, err = g.prompt("🏷️ Enter header name: ")
		if err != nil { return nil, err }
		data.HeaderValue, err = g.prompt("📝 Enter header value: ")
		if err != nil { return nil, err }

	default:
		fmt.Printf("⚠️ Unknown auth type '%s', skipping auth config.\n", data.Type)
		return nil, nil
	}

	return data, nil
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
    {{- if .Auth }}
    upstreamAuth:
      {{- if eq .Auth.Type "apikey" }}
      apiKey:
        paramName: "{{ .Auth.ParamName }}"
        in: {{ .Auth.In }}
        value:
          {{- if contains "$" .Auth.Value }}
          environmentVariable: "{{ trimPrefix "${" (trimSuffix "}" .Auth.Value) }}"
          {{- else }}
          plainText: "{{ .Auth.Value }}"
          {{- end }}
      {{- else if eq .Auth.Type "bearer" }}
      bearerToken:
        token:
          {{- if contains "$" .Auth.Value }}
          environmentVariable: "{{ trimPrefix "${" (trimSuffix "}" .Auth.Value) }}"
          {{- else }}
          plainText: "{{ .Auth.Value }}"
          {{- end }}
      {{- else if eq .Auth.Type "basic" }}
      basicAuth:
        username: "{{ .Auth.Username }}"
        password:
          {{- if contains "$" .Auth.Password }}
          environmentVariable: "{{ trimPrefix "${" (trimSuffix "}" .Auth.Password) }}"
          {{- else }}
          plainText: "{{ .Auth.Password }}"
          {{- end }}
      {{- else if eq .Auth.Type "oauth2" }}
      oauth2:
        tokenUrl: "{{ .Auth.TokenURL }}"
        clientId:
          plainText: "{{ .Auth.ClientID }}"
        clientSecret:
          plainText: "{{ .Auth.ClientSecret }}"
        scopes: "{{ .Auth.Scopes }}"
      {{- else if eq .Auth.Type "header" }}
      trustedHeader:
        headerName: "{{ .Auth.HeaderName }}"
        headerValue: "{{ .Auth.HeaderValue }}"
      {{- end }}
    {{- end }}
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
	// Auth holds authentication configuration
	Auth *AuthData
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

	data.Auth, err = g.generateAuth()
	if err != nil {
		return nil, err
	}

	tmpl, err := template.New("httpService").Funcs(template.FuncMap{
		"contains": strings.Contains,
		"trimPrefix": strings.TrimPrefix,
		"trimSuffix": strings.TrimSuffix,
	}).Parse(httpServiceTemplate)
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
    {{- if .Auth }}
    upstreamAuth:
      {{- if eq .Auth.Type "apikey" }}
      apiKey:
        paramName: "{{ .Auth.ParamName }}"
        in: {{ .Auth.In }}
        value:
          {{- if contains "$" .Auth.Value }}
          environmentVariable: "{{ trimPrefix "${" (trimSuffix "}" .Auth.Value) }}"
          {{- else }}
          plainText: "{{ .Auth.Value }}"
          {{- end }}
      {{- else if eq .Auth.Type "bearer" }}
      bearerToken:
        token:
          {{- if contains "$" .Auth.Value }}
          environmentVariable: "{{ trimPrefix "${" (trimSuffix "}" .Auth.Value) }}"
          {{- else }}
          plainText: "{{ .Auth.Value }}"
          {{- end }}
      {{- else if eq .Auth.Type "basic" }}
      basicAuth:
        username: "{{ .Auth.Username }}"
        password:
          {{- if contains "$" .Auth.Password }}
          environmentVariable: "{{ trimPrefix "${" (trimSuffix "}" .Auth.Password) }}"
          {{- else }}
          plainText: "{{ .Auth.Password }}"
          {{- end }}
      {{- else if eq .Auth.Type "oauth2" }}
      oauth2:
        tokenUrl: "{{ .Auth.TokenURL }}"
        clientId:
          plainText: "{{ .Auth.ClientID }}"
        clientSecret:
          plainText: "{{ .Auth.ClientSecret }}"
        scopes: "{{ .Auth.Scopes }}"
      {{- else if eq .Auth.Type "header" }}
      trustedHeader:
        headerName: "{{ .Auth.HeaderName }}"
        headerValue: "{{ .Auth.HeaderValue }}"
      {{- end }}
    {{- end }}
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
	// Auth holds authentication configuration
	Auth *AuthData
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

	data.Auth, err = g.generateAuth()
	if err != nil {
		return nil, err
	}

	tmpl, err := template.New("grpcService").Funcs(template.FuncMap{
		"contains": strings.Contains,
		"trimPrefix": strings.TrimPrefix,
		"trimSuffix": strings.TrimSuffix,
	}).Parse(grpcServiceTemplate)
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
    {{- if .Auth }}
    upstreamAuth:
      {{- if eq .Auth.Type "apikey" }}
      apiKey:
        paramName: "{{ .Auth.ParamName }}"
        in: {{ .Auth.In }}
        value:
          {{- if contains "$" .Auth.Value }}
          environmentVariable: "{{ trimPrefix "${" (trimSuffix "}" .Auth.Value) }}"
          {{- else }}
          plainText: "{{ .Auth.Value }}"
          {{- end }}
      {{- else if eq .Auth.Type "bearer" }}
      bearerToken:
        token:
          {{- if contains "$" .Auth.Value }}
          environmentVariable: "{{ trimPrefix "${" (trimSuffix "}" .Auth.Value) }}"
          {{- else }}
          plainText: "{{ .Auth.Value }}"
          {{- end }}
      {{- else if eq .Auth.Type "basic" }}
      basicAuth:
        username: "{{ .Auth.Username }}"
        password:
          {{- if contains "$" .Auth.Password }}
          environmentVariable: "{{ trimPrefix "${" (trimSuffix "}" .Auth.Password) }}"
          {{- else }}
          plainText: "{{ .Auth.Password }}"
          {{- end }}
      {{- else if eq .Auth.Type "oauth2" }}
      oauth2:
        tokenUrl: "{{ .Auth.TokenURL }}"
        clientId:
          plainText: "{{ .Auth.ClientID }}"
        clientSecret:
          plainText: "{{ .Auth.ClientSecret }}"
        scopes: "{{ .Auth.Scopes }}"
      {{- else if eq .Auth.Type "header" }}
      trustedHeader:
        headerName: "{{ .Auth.HeaderName }}"
        headerValue: "{{ .Auth.HeaderValue }}"
      {{- end }}
    {{- end }}
`

// OpenAPIServiceData holds the data required to generate an OpenAPI service configuration.
// It is used as the data context for the openapiServiceTemplate.
type OpenAPIServiceData struct {
	// Name is the name of the service.
	Name string
	// SpecPath is the path or URL to the OpenAPI specification file.
	SpecPath string
	// Auth holds authentication configuration
	Auth *AuthData
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

	data.Auth, err = g.generateAuth()
	if err != nil {
		return nil, err
	}

	tmpl, err := template.New("openapiService").Funcs(template.FuncMap{
		"contains": strings.Contains,
		"trimPrefix": strings.TrimPrefix,
		"trimSuffix": strings.TrimSuffix,
	}).Parse(openapiServiceTemplate)
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
    {{- if .Auth }}
    upstreamAuth:
      {{- if eq .Auth.Type "apikey" }}
      apiKey:
        paramName: "{{ .Auth.ParamName }}"
        in: {{ .Auth.In }}
        value:
          {{- if contains "$" .Auth.Value }}
          environmentVariable: "{{ trimPrefix "${" (trimSuffix "}" .Auth.Value) }}"
          {{- else }}
          plainText: "{{ .Auth.Value }}"
          {{- end }}
      {{- else if eq .Auth.Type "bearer" }}
      bearerToken:
        token:
          {{- if contains "$" .Auth.Value }}
          environmentVariable: "{{ trimPrefix "${" (trimSuffix "}" .Auth.Value) }}"
          {{- else }}
          plainText: "{{ .Auth.Value }}"
          {{- end }}
      {{- else if eq .Auth.Type "basic" }}
      basicAuth:
        username: "{{ .Auth.Username }}"
        password:
          {{- if contains "$" .Auth.Password }}
          environmentVariable: "{{ trimPrefix "${" (trimSuffix "}" .Auth.Password) }}"
          {{- else }}
          plainText: "{{ .Auth.Password }}"
          {{- end }}
      {{- else if eq .Auth.Type "oauth2" }}
      oauth2:
        tokenUrl: "{{ .Auth.TokenURL }}"
        clientId:
          plainText: "{{ .Auth.ClientID }}"
        clientSecret:
          plainText: "{{ .Auth.ClientSecret }}"
        scopes: "{{ .Auth.Scopes }}"
      {{- else if eq .Auth.Type "header" }}
      trustedHeader:
        headerName: "{{ .Auth.HeaderName }}"
        headerValue: "{{ .Auth.HeaderValue }}"
      {{- end }}
    {{- end }}
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
	// Auth holds authentication configuration
	Auth *AuthData
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

	data.Auth, err = g.generateAuth()
	if err != nil {
		return nil, err
	}

	tmpl, err := template.New("graphqlService").Funcs(template.FuncMap{
		"contains": strings.Contains,
		"trimPrefix": strings.TrimPrefix,
		"trimSuffix": strings.TrimSuffix,
	}).Parse(graphqlServiceTemplate)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
