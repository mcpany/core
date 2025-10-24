/*
 * Copyright 2025 Author(s) of MCP-XY
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package framework

import (
	"fmt"
	"net/http"
	"path/filepath"
	"testing"

	apiv1 "github.com/mcpxy/core/proto/api/v1"
	configv1 "github.com/mcpxy/core/proto/config/v1"
	"github.com/mcpxy/core/tests/integration"
	"github.com/stretchr/testify/require"
)

func BuildCalculatorServer(t *testing.T) *integration.ManagedProcess {
	port := integration.FindFreePort(t)
	root, err := integration.GetProjectRoot()
	require.NoError(t, err)

	calculatorPath := filepath.Join(root, "build/test/bin/calculator")
	proc := integration.NewManagedProcess(t, "http_calculator_server", calculatorPath, []string{fmt.Sprintf("--port=%d", port)}, nil)
	proc.Port = port
	return proc
}

func RegisterCalculatorService(t *testing.T, registrationClient apiv1.RegistrationServiceClient, upstreamEndpoint string) {
	const serviceID = "e2e_http_calculator"
	paramA := "a"
	paramB := "b"
	paramADesc := "first number"
	paramBDesc := "second number"
	params := []*configv1.HttpParameterMapping{
		configv1.HttpParameterMapping_builder{Schema: configv1.ParameterSchema_builder{Name: &paramA, Description: &paramADesc}.Build()}.Build(),
		configv1.HttpParameterMapping_builder{Schema: configv1.ParameterSchema_builder{Name: &paramB, Description: &paramBDesc}.Build()}.Build(),
	}

	toolName := "add"
	toolDesc := "add two numbers"
	toolSchema := configv1.ToolSchema_builder{
		Name:        &toolName,
		Description: &toolDesc,
	}.Build()

	integration.RegisterHTTPServiceWithParams(t, registrationClient, serviceID, upstreamEndpoint, toolSchema, "/add", http.MethodPost, params, nil)
}
