// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/client"
	"github.com/mcpany/core/server/pkg/pool"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

type mockTransport struct {
	roundTripFunc func(req *http.Request) (*http.Response, error)
}

func (m *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if m.roundTripFunc != nil {
		return m.roundTripFunc(req)
	}
	return nil, nil
}

// Mock pool
type mockHTTPPool struct {
	transport *mockTransport
}

func (m *mockHTTPPool) Get(ctx context.Context) (*client.HTTPClientWrapper, error) {
	c := &http.Client{Transport: m.transport}
	return &client.HTTPClientWrapper{Client: c}, nil
}
func (m *mockHTTPPool) Put(c *client.HTTPClientWrapper) {}
func (m *mockHTTPPool) Close() error                    { return nil }
func (m *mockHTTPPool) Len() int                        { return 0 }

func TestHTTPTool_PrepareInputs_PathTraversal(t *testing.T) {
	pm := pool.NewManager()
	transport := &mockTransport{}
	pm.Register("s1", &mockHTTPPool{transport: transport})

	callDef := &configv1.HttpCallDefinition{
		Parameters: []*configv1.HttpParameterMapping{
			{
				Schema: &configv1.ParameterSchema{Name: proto.String("pathParam")},
			},
		},
	}

	toolProto := &v1.Tool{
		Name:                proto.String("test"),
		ServiceId:           proto.String("s1"),
		UnderlyingMethodFqn: proto.String("GET http://example.com/{{pathParam}}"),
	}

	httpTool := NewHTTPTool(toolProto, pm, "s1", nil, callDef, nil, nil, "call1")

	req := &ExecutionRequest{
		ToolInputs: json.RawMessage(`{"pathParam": "../etc/passwd"}`),
	}

	_, err := httpTool.Execute(context.Background(), req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "path traversal attempt detected")
}

func TestHTTPTool_PrepareInputs_MissingRequired(t *testing.T) {
	pm := pool.NewManager()
	transport := &mockTransport{}
	pm.Register("s1", &mockHTTPPool{transport: transport})

	callDef := &configv1.HttpCallDefinition{
		Parameters: []*configv1.HttpParameterMapping{
			{
				Schema: &configv1.ParameterSchema{
					Name:       proto.String("req"),
					IsRequired: proto.Bool(true),
				},
			},
		},
	}

	toolProto := &v1.Tool{
		Name:                proto.String("test"),
		ServiceId:           proto.String("s1"),
		UnderlyingMethodFqn: proto.String("GET http://example.com"),
	}

	httpTool := NewHTTPTool(toolProto, pm, "s1", nil, callDef, nil, nil, "call1")

	req := &ExecutionRequest{
		ToolInputs: json.RawMessage(`{}`),
	}

	_, err := httpTool.Execute(context.Background(), req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing required parameter")
}
