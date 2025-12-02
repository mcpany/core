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
