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

package openapi

import (
	"context"
	"fmt"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/mcpxy/core/pkg/tool"
	"github.com/mcpxy/core/pkg/util"
	configv1 "github.com/mcpxy/core/proto/config/v1"
	v1 "github.com/mcpxy/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/proto"
)

// MockToolManager is a mock of ToolManagerInterface.
type MockToolManager struct {
	mock.Mock
}

func (m *MockToolManager) AddTool(tool tool.Tool) error {
	args := m.Called(tool)
	return args.Error(0)
}

func (m *MockToolManager) GetTool(toolID string) (tool.Tool, bool) {
	args := m.Called(toolID)
	if args.Get(0) == nil {
		return nil, args.Bool(1)
	}
	return args.Get(0).(tool.Tool), args.Bool(1)
}

func (m *MockToolManager) ListTools() []tool.Tool {
	args := m.Called()
	return args.Get(0).([]tool.Tool)
}

func (m *MockToolManager) ClearToolsForService(serviceKey string) {
	m.Called(serviceKey)
}

func (m *MockToolManager) AddServiceInfo(serviceKey string, info *tool.ServiceInfo) {
	m.Called(serviceKey, info)
}

func (m *MockToolManager) GetServiceInfo(serviceKey string) (*tool.ServiceInfo, bool) {
	args := m.Called(serviceKey)
	if args.Get(0) == nil {
		return nil, args.Bool(1)
	}
	return args.Get(0).(*tool.ServiceInfo), args.Bool(1)
}

func (m *MockToolManager) SetMCPServer(provider tool.MCPServerProvider) {
	m.Called(provider)
}

func (m *MockToolManager) ExecuteTool(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0), args.Error(1)
}

func TestNewOpenAPIUpstream(t *testing.T) {
	u := NewOpenAPIUpstream()
	assert.NotNil(t, u)
	ou, ok := u.(*OpenAPIUpstream)
	assert.True(t, ok)
	assert.NotNil(t, ou.openapiCache)
	assert.NotNil(t, ou.httpClients)
}

func TestOpenAPIUpstream_getHTTPClient(t *testing.T) {
	u := NewOpenAPIUpstream()
	ou := u.(*OpenAPIUpstream)

	client1 := ou.getHTTPClient("service1")
	assert.NotNil(t, client1)

	client2 := ou.getHTTPClient("service1")
	assert.Same(t, client1, client2, "Should return the same client for the same service key")

	client3 := ou.getHTTPClient("service2")
	assert.NotNil(t, client3)
	assert.NotSame(t, client1, client3, "Should return different clients for different service keys")
}

func TestOpenAPIUpstream_Register_Errors(t *testing.T) {
	ctx := context.Background()
	mockToolManager := new(MockToolManager)
	upstream := NewOpenAPIUpstream()

	t.Run("invalid service name", func(t *testing.T) {
		config := configv1.UpstreamServiceConfig_builder{
			Name: proto.String(""),
		}.Build()
		_, _, err := upstream.Register(ctx, config, mockToolManager, nil, nil, false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "name cannot be empty")
	})

	t.Run("nil openapi service config", func(t *testing.T) {
		config := configv1.UpstreamServiceConfig_builder{
			Name:           proto.String("test-service"),
			OpenapiService: nil,
		}.Build()
		expectedKey, _ := util.GenerateID("test-service")
		mockToolManager.On("AddServiceInfo", expectedKey, mock.Anything).Return().Once()
		_, _, err := upstream.Register(ctx, config, mockToolManager, nil, nil, false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "openapi service config is nil")
	})

	t.Run("missing openapi spec", func(t *testing.T) {
		config := configv1.UpstreamServiceConfig_builder{
			Name: proto.String("test-service"),
			OpenapiService: configv1.OpenapiUpstreamService_builder{
				OpenapiSpec: proto.String(""),
			}.Build(),
		}.Build()
		expectedKey, _ := util.GenerateID("test-service")
		mockToolManager.On("AddServiceInfo", expectedKey, mock.Anything).Return().Once()
		_, _, err := upstream.Register(ctx, config, mockToolManager, nil, nil, false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "OpenAPI spec content is missing")
	})

	t.Run("invalid openapi spec", func(t *testing.T) {
		config := configv1.UpstreamServiceConfig_builder{
			Name: proto.String("test-service"),
			OpenapiService: configv1.OpenapiUpstreamService_builder{
				OpenapiSpec: proto.String("invalid spec"),
			}.Build(),
		}.Build()
		expectedKey, _ := util.GenerateID("test-service")
		mockToolManager.On("AddServiceInfo", expectedKey, mock.Anything).Return().Once()
		_, _, err := upstream.Register(ctx, config, mockToolManager, nil, nil, false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse OpenAPI spec")
	})
}

func TestAddOpenAPIToolsToIndex_Errors(t *testing.T) {
	ctx := context.Background()
	u := NewOpenAPIUpstream().(*OpenAPIUpstream)
	serviceKey := "test-service"
	doc := &openapi3.T{}
	serviceConfig := configv1.UpstreamServiceConfig_builder{
		OpenapiService: &configv1.OpenapiUpstreamService{},
	}.Build()

	t.Run("invalid underlying method FQN", func(t *testing.T) {
		mockToolManager := new(MockToolManager)
		mockToolManager.On("GetTool", mock.Anything).Return(nil, false)
		pbTools := []*v1.Tool{
			v1.Tool_builder{
				Name:                proto.String("test-tool"),
				UnderlyingMethodFqn: proto.String("invalid"),
			}.Build(),
		}
		count := u.addOpenAPIToolsToIndex(ctx, pbTools, serviceKey, mockToolManager, false, doc, serviceConfig)
		assert.Equal(t, 0, count)
	})

	t.Run("path not found in spec", func(t *testing.T) {
		mockToolManager := new(MockToolManager)
		mockToolManager.On("GetTool", mock.Anything).Return(nil, false)
		pbTools := []*v1.Tool{
			v1.Tool_builder{
				Name:                proto.String("test-tool"),
				UnderlyingMethodFqn: proto.String("GET /nonexistent"),
			}.Build(),
		}
		doc.Paths = openapi3.NewPaths()
		count := u.addOpenAPIToolsToIndex(ctx, pbTools, serviceKey, mockToolManager, false, doc, serviceConfig)
		assert.Equal(t, 0, count)
	})

	t.Run("operation not found for method", func(t *testing.T) {
		mockToolManager := new(MockToolManager)
		mockToolManager.On("GetTool", mock.Anything).Return(nil, false)
		pbTools := []*v1.Tool{
			v1.Tool_builder{
				Name:                proto.String("test-tool"),
				UnderlyingMethodFqn: proto.String("POST /test"),
			}.Build(),
		}
		doc.Paths = openapi3.NewPaths()
		doc.Paths.Set("/test", &openapi3.PathItem{
			Get: &openapi3.Operation{},
		})
		count := u.addOpenAPIToolsToIndex(ctx, pbTools, serviceKey, mockToolManager, false, doc, serviceConfig)
		assert.Equal(t, 0, count)
	})

	t.Run("add tool fails", func(t *testing.T) {
		mockToolManager := new(MockToolManager)
		mockToolManager.On("GetTool", mock.Anything).Return(nil, false)
		pbTools := []*v1.Tool{
			v1.Tool_builder{
				Name:                proto.String("test-tool"),
				UnderlyingMethodFqn: proto.String("GET /test"),
			}.Build(),
		}
		doc.Paths = openapi3.NewPaths()
		doc.Paths.Set("/test", &openapi3.PathItem{
			Get: &openapi3.Operation{},
		})

		mockToolManager.On("AddTool", mock.Anything).Return(fmt.Errorf("failed to add tool")).Once()
		count := u.addOpenAPIToolsToIndex(ctx, pbTools, serviceKey, mockToolManager, false, doc, serviceConfig)
		assert.Equal(t, 0, count)
		mockToolManager.AssertExpectations(t)
	})
}

func TestOpenAPIUpstream_Register_Cache(t *testing.T) {
	ctx := context.Background()
	mockToolManager := new(MockToolManager)
	u := NewOpenAPIUpstream()
	ou := u.(*OpenAPIUpstream)

	config := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("test-service"),
		OpenapiService: configv1.OpenapiUpstreamService_builder{
			Address:     proto.String("http://localhost"),
			OpenapiSpec: proto.String(sampleOpenAPISpecJSONForCacheTest),
		}.Build(),
	}.Build()

	expectedKey, _ := util.GenerateID("test-service")
	mockToolManager.On("AddServiceInfo", expectedKey, mock.Anything).Twice()
	mockToolManager.On("GetTool", mock.Anything).Return(nil, false)
	mockToolManager.On("AddTool", mock.Anything).Return(nil)

	// First registration, should parse and cache
	_, _, err := u.Register(ctx, config, mockToolManager, nil, nil, false)
	assert.NoError(t, err)
	assert.Equal(t, uint64(1), uint64(ou.openapiCache.Len()))

	// Second registration with same spec, should use cache
	_, _, err = u.Register(ctx, config, mockToolManager, nil, nil, false)
	assert.NoError(t, err)
	assert.Equal(t, uint64(1), uint64(ou.openapiCache.Len()), "Cache length should not increase")

	mockToolManager.AssertExpectations(t)
}

const sampleOpenAPISpecJSONForCacheTest = `{
  "openapi": "3.0.0",
  "info": {
    "title": "Sample API",
    "version": "1.0.0"
  },
  "paths": {
    "/test": {
      "get": {
        "summary": "A test endpoint",
        "operationId": "getTest",
        "responses": {
          "200": {
            "description": "A successful response"
          }
        }
      }
    }
  }
}`
