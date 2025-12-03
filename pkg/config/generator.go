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

package config

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strings"
	"text/template"
)

type Generator struct {
	Reader *bufio.Reader
}

func NewGenerator() *Generator {
	return &Generator{
		Reader: bufio.NewReader(os.Stdin),
	}
}

type Validator func(input string) error

func (g *Generator) prompt(prompt string, validator Validator) (string, error) {
	for {
		fmt.Print(prompt)
		input, err := g.Reader.ReadString('\n')
		if err != nil {
			return "", err
		}
		input = strings.TrimSpace(input)

		if err := validator(input); err != nil {
			fmt.Printf("Invalid input: %v\n", err)
			continue
		}
		return input, nil
	}
}

func notEmpty(input string) error {
	if input == "" {
		return fmt.Errorf("input cannot be empty")
	}
	return nil
}

func (g *Generator) Generate() ([]byte, error) {
	serviceType, err := g.prompt("Enter service type (http, grpc, openapi, graphql): ", notEmpty)
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

const httpServiceTemplate = `upstreamServices:
  - name: "{{ .Name }}"
    httpService:
      address: "{{ .Address }}"
      calls:
        - operationId: "{{ .OperationID }}"
          description: "{{ .Description }}"
          method: "{{ .Method }}"
          endpointPath: "{{ .EndpointPath }}"
{{- if .Auth.Type }}
    upstreamAuthentication:
      {{- if eq .Auth.Type "apiKey" }}
      apiKey:
        headerName: "{{ .Auth.APIKey.HeaderName }}"
        apiKey:
          plainText: "{{ .Auth.APIKey.APIKey }}"
      {{- else if eq .Auth.Type "bearerToken" }}
      bearerToken:
        token:
          plainText: "{{ .Auth.BearerToken.Token }}"
      {{- else if eq .Auth.Type "basicAuth" }}
      basicAuth:
        username: "{{ .Auth.BasicAuth.Username }}"
        password:
          plainText: "{{ .Auth.BasicAuth.Password }}"
      {{- end }}
{{- end }}
`

type HTTPServiceData struct {
	Name         string
	Address      string
	OperationID  string
	Description  string
	Method       string
	EndpointPath string
	Auth         AuthData
}

func (g *Generator) generateHTTPService() ([]byte, error) {
	data := HTTPServiceData{}
	var err error

	data.Name, err = g.prompt("Enter service name: ", notEmpty)
	if err != nil {
		return nil, err
	}

	data.Address, err = g.prompt("Enter service address: ", notEmpty)
	if err != nil {
		return nil, err
	}

	data.OperationID, err = g.prompt("Enter operation ID: ", notEmpty)
	if err != nil {
		return nil, err
	}

	data.Description, err = g.prompt("Enter description: ", notEmpty)
	if err != nil {
		return nil, err
	}

	data.Method, err = g.prompt("Enter HTTP method (e.g., HTTP_METHOD_GET): ", notEmpty)
	if err != nil {
		return nil, err
	}

	data.EndpointPath, err = g.prompt("Enter endpoint path: ", notEmpty)
	if err != nil {
		return nil, err
	}

	data.Auth, err = g.generateAuthData()
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
{{- if .Auth.Type }}
    upstreamAuthentication:
      {{- if eq .Auth.Type "apiKey" }}
      apiKey:
        headerName: "{{ .Auth.APIKey.HeaderName }}"
        apiKey:
          plainText: "{{ .Auth.APIKey.APIKey }}"
      {{- else if eq .Auth.Type "bearerToken" }}
      bearerToken:
        token:
          plainText: "{{ .Auth.BearerToken.Token }}"
      {{- else if eq .Auth.Type "basicAuth" }}
      basicAuth:
        username: "{{ .Auth.BasicAuth.Username }}"
        password:
          plainText: "{{ .Auth.BasicAuth.Password }}"
      {{- end }}
{{- end }}
`

type GRPCServiceData struct {
	Name              string
	Address           string
	ReflectionEnabled bool
	Auth              AuthData
}

func (g *Generator) generateGRPCService() ([]byte, error) {
	data := GRPCServiceData{}
	var err error

	data.Name, err = g.prompt("Enter service name: ", notEmpty)
	if err != nil {
		return nil, err
	}

	data.Address, err = g.prompt("Enter service address: ", notEmpty)
	if err != nil {
		return nil, err
	}

	reflectionEnabled, err := g.prompt("Enable reflection (true/false): ", notEmpty)
	if err != nil {
		return nil, err
	}
	data.ReflectionEnabled = reflectionEnabled == "true"

	data.Auth, err = g.generateAuthData()
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
{{- if .Auth.Type }}
    upstreamAuthentication:
      {{- if eq .Auth.Type "apiKey" }}
      apiKey:
        headerName: "{{ .Auth.APIKey.HeaderName }}"
        apiKey:
          plainText: "{{ .Auth.APIKey.APIKey }}"
      {{- else if eq .Auth.Type "bearerToken" }}
      bearerToken:
        token:
          plainText: "{{ .Auth.BearerToken.Token }}"
      {{- else if eq .Auth.Type "basicAuth" }}
      basicAuth:
        username: "{{ .Auth.BasicAuth.Username }}"
        password:
          plainText: "{{ .Auth.BasicAuth.Password }}"
      {{- end }}
{{- end }}
`

type OpenAPIServiceData struct {
	Name     string
	SpecPath string
	Auth     AuthData
}

func (g *Generator) generateOpenAPIService() ([]byte, error) {
	data := OpenAPIServiceData{}
	var err error

	data.Name, err = g.prompt("Enter service name: ", notEmpty)
	if err != nil {
		return nil, err
	}

	data.SpecPath, err = g.prompt("Enter OpenAPI spec path: ", notEmpty)
	if err != nil {
		return nil, err
	}

	data.Auth, err = g.generateAuthData()
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
{{- if .Auth.Type }}
    upstreamAuthentication:
      {{- if eq .Auth.Type "apiKey" }}
      apiKey:
        headerName: "{{ .Auth.APIKey.HeaderName }}"
        apiKey:
          plainText: "{{ .Auth.APIKey.APIKey }}"
      {{- else if eq .Auth.Type "bearerToken" }}
      bearerToken:
        token:
          plainText: "{{ .Auth.BearerToken.Token }}"
      {{- else if eq .Auth.Type "basicAuth" }}
      basicAuth:
        username: "{{ .Auth.BasicAuth.Username }}"
        password:
          plainText: "{{ .Auth.BasicAuth.Password }}"
      {{- end }}
{{- end }}
`

type GraphQLServiceData struct {
	Name         string
	Address      string
	CallName     string
	SelectionSet string
	Auth         AuthData
}

func (g *Generator) generateGraphQLService() ([]byte, error) {
	data := GraphQLServiceData{}
	var err error

	data.Name, err = g.prompt("Enter service name: ", notEmpty)
	if err != nil {
		return nil, err
	}

	data.Address, err = g.prompt("Enter service address: ", notEmpty)
	if err != nil {
		return nil, err
	}

	data.CallName, err = g.prompt("Enter call name: ", notEmpty)
	if err != nil {
		return nil, err
	}

	data.SelectionSet, err = g.prompt("Enter selection set: ", notEmpty)
	if err != nil {
		return nil, err
	}

	data.Auth, err = g.generateAuthData()
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

type AuthData struct {
	Type        string
	APIKey      APIKeyData
	BearerToken BearerTokenData
	BasicAuth   BasicAuthData
}

type APIKeyData struct {
	HeaderName string
	APIKey     string
}

type BearerTokenData struct {
	Token string
}

type BasicAuthData struct {
	Username string
	Password string
}

func (g *Generator) generateAuthData() (AuthData, error) {
	authData := AuthData{}
	addAuth, err := g.prompt("Add upstream authentication? (yes/no): ", func(input string) error {
		if input != "yes" && input != "no" {
			return fmt.Errorf("please enter 'yes' or 'no'")
		}
		return nil
	})
	if err != nil {
		return authData, err
	}

	if addAuth == "yes" {
		authType, err := g.prompt("Enter auth type (apiKey, bearerToken, basicAuth): ", func(input string) error {
			if input != "apiKey" && input != "bearerToken" && input != "basicAuth" {
				return fmt.Errorf("unsupported auth type")
			}
			return nil
		})
		if err != nil {
			return authData, err
		}
		authData.Type = authType

		switch authType {
		case "apiKey":
			headerName, err := g.prompt("Enter API key header name: ", notEmpty)
			if err != nil {
				return authData, err
			}
			apiKey, err := g.prompt("Enter API key: ", notEmpty)
			if err != nil {
				return authData, err
			}
			authData.APIKey = APIKeyData{HeaderName: headerName, APIKey: apiKey}
		case "bearerToken":
			token, err := g.prompt("Enter bearer token: ", notEmpty)
			if err != nil {
				return authData, err
			}
			authData.BearerToken = BearerTokenData{Token: token}
		case "basicAuth":
			username, err := g.prompt("Enter username: ", notEmpty)
			if err != nil {
				return authData, err
			}
			password, err := g.prompt("Enter password: ", notEmpty)
			if err != nil {
				return authData, err
			}
			authData.BasicAuth = BasicAuthData{Username: username, Password: password}
		}
	}
	return authData, nil
}
