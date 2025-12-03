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
    {{- if .AuthEnabled }}
    upstreamAuthentication:
      {{- if eq .AuthType "apiKey" }}
      apiKey:
        headerName: "{{ .APIKeyHeaderName }}"
        apiKey:
          {{- if .IsAPIKeyFromEnv }}
          environmentVariable: "{{ .APIKeyEnvVar }}"
          {{- else }}
          plainText: "{{ .APIKeyValue }}"
          {{- end }}
      {{- end }}
    {{- end }}
`

type AuthData struct {
	AuthEnabled     bool
	AuthType        string
	APIKeyHeaderName string
	APIKeyValue     string
	APIKeyEnvVar    string
	IsAPIKeyFromEnv bool
}

type HTTPServiceData struct {
	Name         string
	Address      string
	OperationID  string
	Description  string
	Method       string
	EndpointPath string
	AuthData
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

	if err := g.promptForAuth(&data.AuthData); err != nil {
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

func (g *Generator) promptForAuth(authData *AuthData) error {
	addAuth, err := g.prompt("Add authentication (yes/no): ")
	if err != nil {
		return err
	}

	if strings.ToLower(addAuth) == "yes" {
		authData.AuthEnabled = true
		authType, err := g.prompt("Enter authentication type (apiKey): ")
		if err != nil {
			return err
		}
		authData.AuthType = authType

		if authData.AuthType == "apiKey" {
			authData.APIKeyHeaderName, err = g.prompt("Enter API key header name: ")
			if err != nil {
				return err
			}
			apiKeySource, err := g.prompt("Source API key from (plainText, environmentVariable): ")
			if err != nil {
				return err
			}
			if apiKeySource == "plainText" {
				authData.APIKeyValue, err = g.prompt("Enter API key value: ")
				if err != nil {
					return err
				}
			} else if apiKeySource == "environmentVariable" {
				authData.IsAPIKeyFromEnv = true
				authData.APIKeyEnvVar, err = g.prompt("Enter environment variable name: ")
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

const grpcServiceTemplate = `upstreamServices:
  - name: "{{ .Name }}"
    grpcService:
      address: "{{ .Address }}"
      reflection:
        enabled: {{ .ReflectionEnabled }}
    {{- if .AuthEnabled }}
    upstreamAuthentication:
      {{- if eq .AuthType "apiKey" }}
      apiKey:
        headerName: "{{ .APIKeyHeaderName }}"
        apiKey:
          {{- if .IsAPIKeyFromEnv }}
          environmentVariable: "{{ .APIKeyEnvVar }}"
          {{- else }}
          plainText: "{{ .APIKeyValue }}"
          {{- end }}
      {{- end }}
    {{- end }}
`

type GRPCServiceData struct {
	Name              string
	Address           string
	ReflectionEnabled bool
	AuthData
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

	reflectionEnabled, err := g.prompt("Enable reflection (true/false): ")
	if err != nil {
		return nil, err
	}
	data.ReflectionEnabled = reflectionEnabled == "true"

	if err := g.promptForAuth(&data.AuthData); err != nil {
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
    {{- if .AuthEnabled }}
    upstreamAuthentication:
      {{- if eq .AuthType "apiKey" }}
      apiKey:
        headerName: "{{ .APIKeyHeaderName }}"
        apiKey:
          {{- if .IsAPIKeyFromEnv }}
          environmentVariable: "{{ .APIKeyEnvVar }}"
          {{- else }}
          plainText: "{{ .APIKeyValue }}"
          {{- end }}
      {{- end }}
    {{- end }}
`

type OpenAPIServiceData struct {
	Name     string
	SpecPath string
	AuthData
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

	if err := g.promptForAuth(&data.AuthData); err != nil {
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
    {{- if .AuthEnabled }}
    upstreamAuthentication:
      {{- if eq .AuthType "apiKey" }}
      apiKey:
        headerName: "{{ .APIKeyHeaderName }}"
        apiKey:
          {{- if .IsAPIKeyFromEnv }}
          environmentVariable: "{{ .APIKeyEnvVar }}"
          {{- else }}
          plainText: "{{ .APIKeyValue }}"
          {{- end }}
      {{- end }}
    {{- end }}
`

type GraphQLServiceData struct {
	Name         string
	Address      string
	CallName     string
	SelectionSet string
	AuthData
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

	if err := g.promptForAuth(&data.AuthData); err != nil {
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
