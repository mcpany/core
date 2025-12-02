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

type Config struct {
	UpstreamServices []*UpstreamService `yaml:"upstreamServices"`
}

type UpstreamService struct {
	Name                   string                  `yaml:"name"`
	HTTPService            *HTTPService            `yaml:"httpService,omitempty"`
	GRPCService            *GRPCService            `yaml:"grpcService,omitempty"`
	OpenAPIService         *OpenAPIService         `yaml:"openapiService,omitempty"`
	GraphQLService         *GraphQLService         `yaml:"graphqlService,omitempty"`
	UpstreamAuthentication *UpstreamAuthentication `yaml:"upstreamAuthentication,omitempty"`
}

type HTTPService struct {
	Address string  `yaml:"address"`
	Calls   []*Call `yaml:"calls,omitempty"`
}

type Call struct {
	OperationID       string              `yaml:"operationId,omitempty"`
	Description       string              `yaml:"description,omitempty"`
	Method            string              `yaml:"method,omitempty"`
	EndpointPath      string              `yaml:"endpointPath,omitempty"`
	ParameterMappings []*ParameterMapping `yaml:"parameterMappings,omitempty"`
}

type ParameterMapping struct {
	InputParameterName  string `yaml:"inputParameterName"`
	TargetParameterName string `yaml:"targetParameterName"`
}

type GRPCService struct {
	Address    string          `yaml:"address"`
	Reflection *GRPCReflection `yaml:"reflection,omitempty"`
}

type GRPCReflection struct {
	Enabled bool `yaml:"enabled"`
}

type OpenAPIService struct {
	Spec *Spec `yaml:"spec"`
}

type Spec struct {
	Path string `yaml:"path"`
}

type GraphQLService struct {
	Address string         `yaml:"address"`
	Calls   []*GraphQLCall `yaml:"calls,omitempty"`
}

type GraphQLCall struct {
	Name         string `yaml:"name"`
	SelectionSet string `yaml:"selectionSet"`
}

type UpstreamAuthentication struct {
	APIKey      *APIKeyAuth      `yaml:"apiKey,omitempty"`
	BearerToken *BearerTokenAuth `yaml:"bearerToken,omitempty"`
	BasicAuth   *BasicAuth       `yaml:"basicAuth,omitempty"`
}

type APIKeyAuth struct {
	HeaderName string       `yaml:"headerName"`
	APIKey     *ValueSource `yaml:"apiKey"`
}

type BearerTokenAuth struct {
	Token *ValueSource `yaml:"token"`
}

type BasicAuth struct {
	Username *ValueSource `yaml:"username"`
	Password *ValueSource `yaml:"password"`
}

type ValueSource struct {
	PlainText           string `yaml:"plainText,omitempty"`
	EnvironmentVariable string `yaml:"environmentVariable,omitempty"`
}
