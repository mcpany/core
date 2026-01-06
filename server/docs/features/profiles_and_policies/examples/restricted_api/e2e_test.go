// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/mcpany/core/server/pkg/auth"
	"github.com/mcpany/core/server/pkg/bus"
	"github.com/mcpany/core/server/pkg/config"
	"github.com/mcpany/core/server/pkg/mcpserver"
	"github.com/mcpany/core/server/pkg/pool"
	"github.com/mcpany/core/server/pkg/prompt"
	"github.com/mcpany/core/server/pkg/resource"
	"github.com/mcpany/core/server/pkg/serviceregistry"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/mcpany/core/server/pkg/upstream/factory"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

func TestRestrictedApiE2E(t *testing.T) {
	// 1. Setup Mock Upstream Service (Petstore)
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Serve openapi.yaml
		if r.URL.Path == "/openapi.yaml" {
			w.Header().Set("Content-Type", "application/yaml")
			w.Write([]byte(`
openapi: 3.0.0
info:
  version: 1.0.0
  title: Swagger Petstore
paths:
  /pet/findByStatus:
    get:
      operationId: findPetsByStatus
      parameters:
        - name: status
          in: query
          required: true
          schema:
            type: array
            items:
              type: string
      responses:
        '200':
          description: successful operation
          content:
            application/json:
              schema:
                type: array
                items:
                  $ref: '#/components/schemas/Pet'
  /pet:
    post:
      operationId: addPet
      responses:
        '405':
          description: Invalid input
components:
  schemas:
    Pet:
      type: object
      properties:
        id:
          type: integer
          format: int64
        name:
          type: string
        photoUrls:
          type: array
          items:
            type: string
        status:
          type: string
`))
			return
		}

		// Mock API endpoints
		if r.URL.Path == "/pet/findByStatus" {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`[{"id": 1, "name": "doggie", "status": "available"}]`))
			return
		}

		if r.URL.Path == "/pet" && r.Method == "POST" {
			w.WriteHeader(http.StatusOK)
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}))
	defer mockServer.Close()

	// 2. Config Content using Mock Server
	configContent := fmt.Sprintf(`
upstream_services:
  - id: "petstore-service"
    name: "petstore-service"
    openapi_service:
      address: "%s"
      spec_url: "%s/openapi.yaml"
    call_policies:
      - default_action: DENY
        rules:
          - action: ALLOW
            name_regex: "^petstore-service\\.findPetsByStatus$"
          - action: ALLOW
            name_regex: "^petstore-service\\.getPetById$"
`, mockServer.URL, mockServer.URL)

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	// 3. Setup Server components
	ctx := context.Background()
	fs := afero.NewOsFs()
	store := config.NewFileStore(fs, []string{configPath})
	cfg, err := config.LoadServices(context.Background(), store, "server")
	require.NoError(t, err)

	busProvider, _ := bus.NewProvider(nil)
	toolManager := tool.NewManager(busProvider)
	promptManager := prompt.NewManager()
	resourceManager := resource.NewManager()
	authManager := auth.NewManager()
	poolManager := pool.NewManager()
	upstreamFactory := factory.NewUpstreamServiceFactory(poolManager)
	serviceRegistry := serviceregistry.New(upstreamFactory, toolManager, promptManager, resourceManager, authManager)

	// Note: We don't strictly *need* mcpServer for CallPolicy test as we use toolManager directly,
	// but we initialize it to ensure full setup.
	_, err = mcpserver.NewServer(ctx, toolManager, promptManager, resourceManager, authManager, serviceRegistry, busProvider, true)
	require.NoError(t, err)

	// 4. Register Services
	for _, serviceConfig := range cfg.GetUpstreamServices() {
		upstream, err := upstreamFactory.NewUpstream(serviceConfig)
		require.NoError(t, err)
		_, _, _, err = upstream.Register(ctx, serviceConfig, toolManager, promptManager, resourceManager, false)
		require.NoError(t, err)
	}

	// 5. Test logic
	tests := []struct {
		name          string
		toolName      string
		inputs        map[string]interface{}
		shouldSucceed bool
		expectedError string
	}{
		{
			name:     "Allowed_Tool_Should_Succeed",
			toolName: "petstore-service.findPetsByStatus",
			inputs: map[string]interface{}{
				"status": []string{"available"},
			},
			shouldSucceed: true,
		},
		{
			name:     "Denied_Tool_Should_Fail",
			toolName: "petstore-service.addPet",
			inputs: map[string]interface{}{
				"name":      "doggie",
				"photoUrls": []string{"string"},
			},
			shouldSucceed: false,
			expectedError: "denied by default policy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Check if tool exists
			_, ok := toolManager.GetTool(tt.toolName)
			require.True(t, ok, "Tool %s not found.", tt.toolName)

			// Execute using toolManager to test Call Policy
			inputsBytes, _ := json.Marshal(tt.inputs)
			req := &tool.ExecutionRequest{
				ToolName:   tt.toolName,
				ToolInputs: json.RawMessage(inputsBytes),
			}
			result, err := toolManager.ExecuteTool(ctx, req)

			if tt.shouldSucceed {
				if err != nil {
					t.Logf("Error: %v", err)
				}
				require.NoError(t, err)
				// Log success details
				t.Logf("Tool execution result: %+v", result)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.expectedError)
			}
		})
	}
}
