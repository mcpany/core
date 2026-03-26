// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package openapi

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/mcpany/core/server/pkg/util"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

// MockToolManager is a mock of ToolManagerInterface.
// MockToolManager is a mock of ToolManagerInterface.
// Summary: MockToolManager
	mock.Mock
}

// AddTool ...
// Summary: AddTool
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	args := m.Called(tool)
	return args.Error(0)
}

// GetTool ...
// Summary: GetTool
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	args := m.Called(toolID)
	if args.Get(0) == nil {
		return nil, args.Bool(1)
	}
	return args.Get(0).(tool.Tool), args.Bool(1)
}

// ListServices ...
// Summary: ListServices
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return nil
}

// ListTools ...
// Summary: ListTools
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	args := m.Called()
	return args.Get(0).([]tool.Tool)
}

// ListMCPTools ...
// Summary: ListMCPTools
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return nil
}

// ClearToolsForService ...
// Summary: ClearToolsForService
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	m.Called(serviceID)
}

// AddServiceInfo ...
// Summary: AddServiceInfo
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	m.Called(serviceID, info)
}

// GetServiceInfo ...
// Summary: GetServiceInfo
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	args := m.Called(serviceID)
	if args.Get(0) == nil {
		return nil, args.Bool(1)
	}
	return args.Get(0).(*tool.ServiceInfo), args.Bool(1)
}

// SetMCPServer ...
// Summary: SetMCPServer
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	m.Called(provider)
}

// ExecuteTool ...
// Summary: ExecuteTool
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0), args.Error(1)
}

// AddMiddleware ...
// Summary: AddMiddleware
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
}

// ToolMatchesProfile ...
// Summary: ToolMatchesProfile
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return true
}

// SetProfiles ...
// Summary: SetProfiles
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	m.Called(enabled, defs)
}

// IsServiceAllowed ...
// Summary: IsServiceAllowed
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return true
}

// TestNewOpenAPIUpstream ...
// Summary: TestNewOpenAPIUpstream
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	u := NewOpenAPIUpstream()
	assert.NotNil(t, u)
	ou, ok := u.(*OpenAPIUpstream)
	assert.True(t, ok)
	assert.NotNil(t, ou.openapiCache)
	assert.NotNil(t, ou.httpClients)
}

// TestOpenAPIUpstream_getHTTPClient ...
// Summary: TestOpenAPIUpstream_getHTTPClient
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
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

// TestOpenAPIUpstream_Register_Errors ...
// Summary: TestOpenAPIUpstream_Register_Errors
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	ctx := context.Background()
	mockToolManager := new(MockToolManager)
	upstream := NewOpenAPIUpstream()

	t.Run("invalid service name", func(t *testing.T) {
		config := configv1.UpstreamServiceConfig_builder{
			Name: proto.String(""),
		}.Build()
		_, _, _, err := upstream.Register(ctx, config, mockToolManager, nil, nil, false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "id cannot be empty")
	})

	t.Run("nil openapi service config", func(t *testing.T) {
		config := configv1.UpstreamServiceConfig_builder{
			Name:           proto.String("test-service"),
			OpenapiService: nil,
		}.Build()
		expectedKey, _ := util.SanitizeServiceName("test-service")
		mockToolManager.On("AddServiceInfo", expectedKey, mock.Anything).Return().Once()
		_, _, _, err := upstream.Register(ctx, config, mockToolManager, nil, nil, false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "openapi service config is nil")
	})

	t.Run("missing openapi spec", func(t *testing.T) {
		config := configv1.UpstreamServiceConfig_builder{
			Name: proto.String("test-service"),
			OpenapiService: configv1.OpenapiUpstreamService_builder{
				SpecContent: proto.String(""),
			}.Build(),
		}.Build()
		expectedKey, _ := util.SanitizeServiceName("test-service")
		mockToolManager.On("AddServiceInfo", expectedKey, mock.Anything).Return().Once()
		_, _, _, err := upstream.Register(ctx, config, mockToolManager, nil, nil, false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "OpenAPI spec content is missing")
	})

	t.Run("invalid openapi spec", func(t *testing.T) {
		config := configv1.UpstreamServiceConfig_builder{
			Name: proto.String("test-service"),
			OpenapiService: configv1.OpenapiUpstreamService_builder{
				SpecContent: proto.String("invalid spec"),
			}.Build(),
		}.Build()
		expectedKey, _ := util.SanitizeServiceName("test-service")
		mockToolManager.On("AddServiceInfo", expectedKey, mock.Anything).Return().Once()
		_, _, _, err := upstream.Register(ctx, config, mockToolManager, nil, nil, false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse OpenAPI spec")
	})

	t.Run("invalid address scheme", func(t *testing.T) {
		config := configv1.UpstreamServiceConfig_builder{
			Name: proto.String("test-service-scheme"),
			OpenapiService: configv1.OpenapiUpstreamService_builder{
				Address: proto.String("file:///etc/passwd"),
			}.Build(),
		}.Build()
		_, _, _, err := upstream.Register(ctx, config, mockToolManager, nil, nil, false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid openapi service address scheme")
	})
}

// TestOpenAPIUpstream_Register_SpecUrl ...
// Summary: TestOpenAPIUpstream_Register_SpecUrl
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	ctx := context.Background()
	mockToolManager := new(MockToolManager)
	upstream := NewOpenAPIUpstream()

	// Start a test server to serve the spec
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, sampleOpenAPISpecJSONForCacheTest)
	}))
	defer ts.Close()

	// Need to allow loopback for the httptest server to work
	t.Setenv("MCPANY_ALLOW_LOOPBACK_RESOURCES", "true")

	config := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("test-service-url"),
		OpenapiService: configv1.OpenapiUpstreamService_builder{
			SpecUrl: proto.String(ts.URL),
		}.Build(),
	}.Build()

	expectedKey, _ := util.SanitizeServiceName("test-service-url")
	mockToolManager.On("AddServiceInfo", expectedKey, mock.Anything).Return().Once()
	mockToolManager.On("GetTool", mock.Anything).Return(nil, false)
	mockToolManager.On("AddTool", mock.Anything).Return(nil)

	// Register should fetch spec from URL
	_, _, _, err := upstream.Register(ctx, config, mockToolManager, nil, nil, false)
	assert.NoError(t, err)
	mockToolManager.AssertExpectations(t)
}

// TestOpenAPIUpstream_Register_SpecUrl_SSRF ...
// Summary: TestOpenAPIUpstream_Register_SpecUrl_SSRF
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	ctx := context.Background()
	mockToolManager := new(MockToolManager)
	upstream := NewOpenAPIUpstream()

	// Start a test server to serve the spec
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, sampleOpenAPISpecJSONForCacheTest)
	}))
	defer ts.Close()

	// Ensure loopback is NOT allowed
	t.Setenv("MCPANY_ALLOW_LOOPBACK_RESOURCES", "false")

	config := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("test-service-url-ssrf"),
		OpenapiService: configv1.OpenapiUpstreamService_builder{
			SpecUrl: proto.String(ts.URL),
		}.Build(),
	}.Build()

	expectedKey, _ := util.SanitizeServiceName("test-service-url-ssrf")
	mockToolManager.On("AddServiceInfo", expectedKey, mock.Anything).Return().Once()

	// Register should fail to fetch spec due to SSRF protection
	_, _, _, err := upstream.Register(ctx, config, mockToolManager, nil, nil, false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "OpenAPI spec content is missing")
}

// TestOpenAPIUpstream_Register_InvalidSpecUrl ...
// Summary: TestOpenAPIUpstream_Register_InvalidSpecUrl
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	ctx := context.Background()
	mockToolManager := new(MockToolManager)
	upstream := NewOpenAPIUpstream()

	config := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("test-service-invalid-spec-url"),
		OpenapiService: configv1.OpenapiUpstreamService_builder{
			SpecUrl: proto.String("file:///etc/passwd"),
		}.Build(),
	}.Build()

	expectedKey, _ := util.SanitizeServiceName("test-service-invalid-spec-url")
	mockToolManager.On("AddServiceInfo", expectedKey, mock.Anything).Return().Once()

	// Register should fail because spec_url is invalid scheme (file://) and thus content is missing
	_, _, _, err := upstream.Register(ctx, config, mockToolManager, nil, nil, false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "OpenAPI spec content is missing")
}

// TestAddOpenAPIToolsToIndex_Errors ...
// Summary: TestAddOpenAPIToolsToIndex_Errors
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	ctx := context.Background()
	u := NewOpenAPIUpstream().(*OpenAPIUpstream)
	serviceID := "test-service"
	doc := &openapi3.T{}
	serviceConfig := configv1.UpstreamServiceConfig_builder{
		OpenapiService: configv1.OpenapiUpstreamService_builder{
			Tools: []*configv1.ToolDefinition{
				configv1.ToolDefinition_builder{
					Name:   proto.String("test-tool"),
					CallId: proto.String("test-call"),
				}.Build(),
			},
			Calls: map[string]*configv1.OpenAPICallDefinition{
				"test-call": configv1.OpenAPICallDefinition_builder{
					Id: proto.String("test-call"),
				}.Build(),
			},
		}.Build(),
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
		count := u.addOpenAPIToolsToIndex(ctx, pbTools, serviceID, mockToolManager, nil, false, doc, serviceConfig)
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
		count := u.addOpenAPIToolsToIndex(ctx, pbTools, serviceID, mockToolManager, nil, false, doc, serviceConfig)
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
		count := u.addOpenAPIToolsToIndex(ctx, pbTools, serviceID, mockToolManager, nil, false, doc, serviceConfig)
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
		count := u.addOpenAPIToolsToIndex(ctx, pbTools, serviceID, mockToolManager, nil, false, doc, serviceConfig)
		assert.Equal(t, 0, count)
		mockToolManager.AssertExpectations(t)
	})
}

// TestOpenAPIUpstream_Register_Cache ...
// Summary: TestOpenAPIUpstream_Register_Cache
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	ctx := context.Background()
	mockToolManager := new(MockToolManager)
	u := NewOpenAPIUpstream()
	ou := u.(*OpenAPIUpstream)

	config := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("test-service"),
		OpenapiService: configv1.OpenapiUpstreamService_builder{
			Address:     proto.String("http://127.0.0.1"),
			SpecContent: proto.String(sampleOpenAPISpecJSONForCacheTest),
			Tools: []*configv1.ToolDefinition{
				configv1.ToolDefinition_builder{
					Name:   proto.String("getTest"),
					CallId: proto.String("getTest-call"),
				}.Build(),
			},
			Calls: map[string]*configv1.OpenAPICallDefinition{
				"getTest-call": configv1.OpenAPICallDefinition_builder{
					Id: proto.String("getTest-call"),
				}.Build(),
			},
		}.Build(),
	}.Build()

	expectedKey, _ := util.SanitizeServiceName("test-service")
	mockToolManager.On("AddServiceInfo", expectedKey, mock.Anything).Twice()
	mockToolManager.On("GetTool", mock.Anything).Return(nil, false)
	mockToolManager.On("AddTool", mock.Anything).Return(nil)

	// First registration, should parse and cache
	_, _, _, err := u.Register(ctx, config, mockToolManager, nil, nil, false)
	assert.NoError(t, err)
	assert.Equal(t, uint64(1), uint64(ou.openapiCache.Len())) //nolint:gosec // safe cast

	// Second registration with same spec, should use cache
	_, _, _, err = u.Register(ctx, config, mockToolManager, nil, nil, false)
	assert.NoError(t, err)
	assert.Equal(t, uint64(1), uint64(ou.openapiCache.Len()), "Cache length should not increase") //nolint:gosec // safe cast

	mockToolManager.AssertExpectations(t)
}

// TestInputSchemaGeneration ...
// Summary: TestInputSchemaGeneration
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	ctx := context.Background()
	mockToolManager := new(MockToolManager)
	upstream := NewOpenAPIUpstream()

	spec := `
openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /users/{userId}:
    get:
      operationId: getUser
      parameters:
        - name: userId
          in: path
          required: true
          schema:
            type: string
        - name: format
          in: query
          schema:
            type: string
            enum: [json, xml]
      responses:
        '200':
          description: OK
`
	config := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("test-service"),
		OpenapiService: configv1.OpenapiUpstreamService_builder{
			SpecContent: proto.String(spec),
			Tools: []*configv1.ToolDefinition{
				configv1.ToolDefinition_builder{
					Name:   proto.String("getUser"),
					CallId: proto.String("getUser-call"),
				}.Build(),
			},
			Calls: map[string]*configv1.OpenAPICallDefinition{
				"getUser-call": configv1.OpenAPICallDefinition_builder{
					Id: proto.String("getUser-call"),
				}.Build(),
			},
		}.Build(),
	}.Build()

	expectedKey, _ := util.SanitizeServiceName("test-service")
	mockToolManager.On("AddServiceInfo", expectedKey, mock.Anything).Return()
	mockToolManager.On("GetTool", mock.Anything).Return(nil, false)

	// Expect AddTool to be called and capture the tool
	var addedTool tool.Tool
	mockToolManager.On("AddTool", mock.Anything).Run(func(args mock.Arguments) {
		addedTool = args.Get(0).(tool.Tool)
	}).Return(nil)

	_, _, _, err := upstream.Register(ctx, config, mockToolManager, nil, nil, false)
	assert.NoError(t, err)

	mockToolManager.AssertCalled(t, "AddTool", mock.Anything)
	require.NotNil(t, addedTool)

	// Verify the input schema
	inputSchema := addedTool.Tool().GetAnnotations().GetInputSchema()
	assert.NotNil(t, inputSchema)
	assert.Equal(t, "object", inputSchema.GetFields()["type"].GetStringValue())

	properties := inputSchema.GetFields()["properties"].GetStructValue().GetFields()
	assert.Contains(t, properties, "userId")
	assert.Contains(t, properties, "format")

	userIDProp := properties["userId"].GetStructValue().GetFields()
	assert.Equal(t, "string", userIDProp["type"].GetStringValue())

	formatProp := properties["format"].GetStructValue().GetFields()
	assert.Equal(t, "string", formatProp["type"].GetStringValue())

	enumVals := formatProp["enum"].GetListValue().GetValues()
	assert.Len(t, enumVals, 2)
	assert.Equal(t, "json", enumVals[0].GetStringValue())
	assert.Equal(t, "xml", enumVals[1].GetStringValue())
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

// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// TestOpenAPIUpstream_Shutdown ...
// Summary: TestOpenAPIUpstream_Shutdown
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	u := NewOpenAPIUpstream()
	ou := u.(*OpenAPIUpstream)

	// Pre-populate a client with empty service ID, which matches default u.serviceID
	client := ou.getHTTPClient("")
	assert.NotNil(t, client)

	// Shutdown
	err := u.Shutdown(context.Background())
	assert.NoError(t, err)

	ou.mu.Lock()
	_, exists := ou.httpClients[""]
	ou.mu.Unlock()

	assert.False(t, exists, "Client should be removed after shutdown")
}

// TestHttpClientImpl_Do ...
// Summary: TestHttpClientImpl_Do
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	// Create a real client but mock the transport to avoid network calls
	client := &http.Client{
		Transport: &mockTransport{},
	}
	wrapper := &httpClientImpl{client: client}

	req, _ := http.NewRequest("GET", "http://example.com", nil)
	resp, err := wrapper.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

type mockTransport struct{}

// RoundTrip ...
// Summary: RoundTrip
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return &http.Response{
		StatusCode: 200,
	}, nil
}

// TestConvertMcpOperationsToTools_NonObjectResponseBody ...
// Summary: TestConvertMcpOperationsToTools_NonObjectResponseBody
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	spec := `
{
  "openapi": "3.0.0",
  "info": { "title": "Test", "version": "1.0" },
  "paths": {
    "/array-response": {
      "get": {
        "operationId": "getArray",
        "responses": {
          "200": {
            "description": "OK",
            "content": {
              "application/json": {
                "schema": {
                  "type": "array",
                  "items": { "type": "string" }
                }
              }
            }
          }
        }
      }
    }
  }
}`
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData([]byte(spec))
	assert.NoError(t, err)

	ops := extractMcpOperationsFromOpenAPI(doc)
	tools := convertMcpOperationsToTools(ops, doc, "test-service")

	assert.Len(t, tools, 1)
	tool := tools[0]
	outputSchema := tool.GetAnnotations().GetOutputSchema()
	props := outputSchema.GetFields()["properties"].GetStructValue().GetFields()

	// Should be wrapped in "response_body"
	assert.Contains(t, props, "response_body")
	respBody := props["response_body"].GetStructValue().GetFields()
	assert.Equal(t, "array", respBody["type"].GetStringValue())
}

// TestConvertSchemaToStructPB_ArrayTypes ...
// Summary: TestConvertSchemaToStructPB_ArrayTypes
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	// Test array without items
	schema := &openapi3.Schema{
		Type: &openapi3.Types{"array"},
	}
	doc := &openapi3.T{}

	val, err := convertSchemaToStructPB("testArray", &openapi3.SchemaRef{Value: schema}, "", doc)
	assert.NoError(t, err)
	fields := val.GetStructValue().GetFields()
	assert.Equal(t, "array", fields["type"].GetStringValue())
	assert.Nil(t, fields["items"], "Items should be nil if not specified")

	// Test array with items
	schemaWithItems := &openapi3.Schema{
		Type: &openapi3.Types{"array"},
		Items: &openapi3.SchemaRef{
			Value: &openapi3.Schema{
				Type: &openapi3.Types{"string"},
			},
		},
	}
	val, err = convertSchemaToStructPB("testArrayItems", &openapi3.SchemaRef{Value: schemaWithItems}, "", doc)
	assert.NoError(t, err)
	fields = val.GetStructValue().GetFields()
	items := fields["items"].GetStructValue().GetFields()
	assert.Equal(t, "string", items["type"].GetStringValue())
}

// TestResolveSchemaRef_EdgeCases ...
// Summary: TestResolveSchemaRef_EdgeCases
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	// Nil schema ref
	val, err := resolveSchemaRef(nil, nil)
	assert.NoError(t, err)
	assert.Nil(t, val)

	// Ref without doc
	ref := &openapi3.SchemaRef{Ref: "#/components/schemas/Test"}
	_, err = resolveSchemaRef(ref, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "nil doc")
}

// TestMergeSchemaProperties_Recursion ...
// Summary: TestMergeSchemaProperties_Recursion
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	// Construct a schema with nested AllOf
	doc := &openapi3.T{
		Components: &openapi3.Components{
			Schemas: make(openapi3.Schemas),
		},
	}

	baseSchema := &openapi3.Schema{
		Properties: map[string]*openapi3.SchemaRef{
			"baseProp": {
				Value: &openapi3.Schema{Type: &openapi3.Types{"string"}},
			},
		},
	}
	doc.Components.Schemas["Base"] = &openapi3.SchemaRef{Value: baseSchema}

	middleSchema := &openapi3.Schema{
		AllOf: openapi3.SchemaRefs{
			{Ref: "#/components/schemas/Base"},
		},
		Properties: map[string]*openapi3.SchemaRef{
			"middleProp": {
				Value: &openapi3.Schema{Type: &openapi3.Types{"integer"}},
			},
		},
	}
	doc.Components.Schemas["Middle"] = &openapi3.SchemaRef{Value: middleSchema}

	finalSchema := &openapi3.Schema{
		AllOf: openapi3.SchemaRefs{
			{Ref: "#/components/schemas/Middle"},
		},
		Properties: map[string]*openapi3.SchemaRef{
			"finalProp": {
				Value: &openapi3.Schema{Type: &openapi3.Types{"boolean"}},
			},
		},
	}

	merged, err := mergeSchemaProperties(finalSchema, doc)
	assert.NoError(t, err)
	assert.Contains(t, merged, "baseProp")
	assert.Contains(t, merged, "middleProp")
	assert.Contains(t, merged, "finalProp")
}

// TestConvertSchemaToStructPB_NoTypeButAllOf ...
// Summary: TestConvertSchemaToStructPB_NoTypeButAllOf
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	// Case where Type is nil but AllOf is present -> should be treated as object
	schema := &openapi3.Schema{
		AllOf: openapi3.SchemaRefs{
			{
				Value: &openapi3.Schema{
					Properties: map[string]*openapi3.SchemaRef{
						"prop": {Value: &openapi3.Schema{Type: &openapi3.Types{"string"}}},
					},
				},
			},
		},
	}
	doc := &openapi3.T{}
	val, err := convertSchemaToStructPB("testAllOf", &openapi3.SchemaRef{Value: schema}, "", doc)
	assert.NoError(t, err)
	fields := val.GetStructValue().GetFields()
	assert.Equal(t, "object", fields["type"].GetStringValue())
	assert.Contains(t, fields["properties"].GetStructValue().GetFields(), "prop")
}

// TestConvertSchemaToStructPB_UnhandledType ...
// Summary: TestConvertSchemaToStructPB_UnhandledType
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	schema := &openapi3.Schema{
		Type: &openapi3.Types{"unknown"},
	}
	doc := &openapi3.T{}
	val, err := convertSchemaToStructPB("testUnknown", &openapi3.SchemaRef{Value: schema}, "", doc)
	assert.NoError(t, err)
	fields := val.GetStructValue().GetFields()
	// Should default to string
	assert.Equal(t, "string", fields["type"].GetStringValue())
}

// TestConvertSchemaToStructPB_WithEnumAndDefault ...
// Summary: TestConvertSchemaToStructPB_WithEnumAndDefault
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	schema := &openapi3.Schema{
		Type:    &openapi3.Types{"string"},
		Enum:    []interface{}{"A", "B"},
		Default: "A",
	}
	doc := &openapi3.T{}
	val, err := convertSchemaToStructPB("testEnum", &openapi3.SchemaRef{Value: schema}, "", doc)
	assert.NoError(t, err)
	fields := val.GetStructValue().GetFields()

	assert.NotNil(t, fields["enum"])
	enums := fields["enum"].GetListValue().GetValues()
	assert.Len(t, enums, 2)
	assert.Equal(t, "A", enums[0].GetStringValue())

	assert.NotNil(t, fields["default"])
	assert.Equal(t, "A", fields["default"].GetStringValue())
}

// GetAllowedServiceIDs ...
// Summary: GetAllowedServiceIDs
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	args := m.Called(profileID)
	return args.Get(0).(map[string]bool), args.Bool(1)
}

// GetToolCountForService ...
// Summary: GetToolCountForService
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return 0
}
