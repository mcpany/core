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
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type Generator struct {
	reader *bufio.Reader
}

func NewGenerator() *Generator {
	return &Generator{
		reader: bufio.NewReader(os.Stdin),
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
	input, err := g.reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(input), nil
}

func (g *Generator) generateHTTPService() ([]byte, error) {
	serviceName, err := g.prompt("Enter service name: ")
	if err != nil {
		return nil, err
	}

	address, err := g.prompt("Enter service address: ")
	if err != nil {
		return nil, err
	}

	upstreamService := &UpstreamService{
		Name: serviceName,
		HTTPService: &HTTPService{
			Address: address,
		},
	}

	authMethod, err := g.prompt("Enter auth method (none, apiKey, bearerToken, basicAuth): ")
	if err != nil {
		return nil, err
	}

	switch authMethod {
	case "apiKey":
		headerName, err := g.prompt("Enter API key header name: ")
		if err != nil {
			return nil, err
		}
		apiKey, err := g.prompt("Enter API key: ")
		if err != nil {
			return nil, err
		}
		upstreamService.UpstreamAuthentication = &UpstreamAuthentication{
			APIKey: &APIKeyAuth{
				HeaderName: headerName,
				APIKey:     &ValueSource{PlainText: apiKey},
			},
		}
	case "bearerToken":
		token, err := g.prompt("Enter bearer token: ")
		if err != nil {
			return nil, err
		}
		upstreamService.UpstreamAuthentication = &UpstreamAuthentication{
			BearerToken: &BearerTokenAuth{
				Token: &ValueSource{PlainText: token},
			},
		}
	case "basicAuth":
		username, err := g.prompt("Enter basic auth username: ")
		if err != nil {
			return nil, err
		}
		password, err := g.prompt("Enter basic auth password: ")
		if err != nil {
			return nil, err
		}
		upstreamService.UpstreamAuthentication = &UpstreamAuthentication{
			BasicAuth: &BasicAuth{
				Username: &ValueSource{PlainText: username},
				Password: &ValueSource{PlainText: password},
			},
		}
	}

	for {
		addTool, err := g.prompt("Add a tool? (y/n): ")
		if err != nil {
			return nil, err
		}
		if addTool != "y" {
			break
		}

		operationID, err := g.prompt("Enter operation ID: ")
		if err != nil {
			return nil, err
		}
		description, err := g.prompt("Enter description: ")
		if err != nil {
			return nil, err
		}
		method, err := g.prompt("Enter HTTP method: ")
		if err != nil {
			return nil, err
		}
		endpointPath, err := g.prompt("Enter endpoint path: ")
		if err != nil {
			return nil, err
		}

		call := &Call{
			OperationID: operationID,
			Description: description,
			Method:      method,
			EndpointPath:    endpointPath,
		}

		for {
			addParam, err := g.prompt("Add a parameter mapping? (y/n): ")
			if err != nil {
				return nil, err
			}
			if addParam != "y" {
				break
			}

			inputName, err := g.prompt("Enter input parameter name: ")
			if err != nil {
				return nil, err
			}
			targetName, err := g.prompt("Enter target parameter name: ")
			if err != nil {
				return nil, err
			}

			call.ParameterMappings = append(call.ParameterMappings, &ParameterMapping{
				InputParameterName:  inputName,
				TargetParameterName: targetName,
			})
		}
		upstreamService.HTTPService.Calls = append(upstreamService.HTTPService.Calls, call)
	}

	config := &Config{
		UpstreamServices: []*UpstreamService{upstreamService},
	}

	return marshalConfig(config)
}

func (g *Generator) generateGRPCService() ([]byte, error) {
	serviceName, err := g.prompt("Enter service name: ")
	if err != nil {
		return nil, err
	}

	address, err := g.prompt("Enter service address: ")
	if err != nil {
		return nil, err
	}

	reflection, err := g.prompt("Enable reflection? (y/n): ")
	if err != nil {
		return nil, err
	}

	upstreamService := &UpstreamService{
		Name: serviceName,
		GRPCService: &GRPCService{
			Address: address,
			Reflection: &GRPCReflection{
				Enabled: reflection == "y",
			},
		},
	}

	config := &Config{
		UpstreamServices: []*UpstreamService{upstreamService},
	}

	return marshalConfig(config)
}

func (g *Generator) generateOpenAPIService() ([]byte, error) {
	serviceName, err := g.prompt("Enter service name: ")
	if err != nil {
		return nil, err
	}

	specPath, err := g.prompt("Enter OpenAPI spec path: ")
	if err != nil {
		return nil, err
	}

	upstreamService := &UpstreamService{
		Name: serviceName,
		OpenAPIService: &OpenAPIService{
			Spec: &Spec{
				Path: specPath,
			},
		},
	}

	config := &Config{
		UpstreamServices: []*UpstreamService{upstreamService},
	}

	return marshalConfig(config)
}

func (g *Generator) generateGraphQLService() ([]byte, error) {
	serviceName, err := g.prompt("Enter service name: ")
	if err != nil {
		return nil, err
	}

	address, err := g.prompt("Enter service address: ")
	if err != nil {
		return nil, err
	}

	upstreamService := &UpstreamService{
		Name: serviceName,
		GraphQLService: &GraphQLService{
			Address: address,
		},
	}

	for {
		addCall, err := g.prompt("Add a call? (y/n): ")
		if err != nil {
			return nil, err
		}
		if addCall != "y" {
			break
		}

		name, err := g.prompt("Enter call name: ")
		if err != nil {
			return nil, err
		}
		selectionSet, err := g.prompt("Enter selection set: ")
		if err != nil {
			return nil, err
		}

		call := &GraphQLCall{
			Name:         name,
			SelectionSet: selectionSet,
		}
		upstreamService.GraphQLService.Calls = append(upstreamService.GraphQLService.Calls, call)
	}

	config := &Config{
		UpstreamServices: []*UpstreamService{upstreamService},
	}

	return marshalConfig(config)
}

func marshalConfig(config *Config) ([]byte, error) {
	return yaml.Marshal(config)
}
