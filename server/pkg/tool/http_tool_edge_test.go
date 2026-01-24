// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/mcpany/core/server/pkg/pool"
	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestHTTPTool_Execute_InitError(t *testing.T) {
	// Create a tool with initError
	tool := &v1.Tool{Name: proto.String("error-tool")}
	ht := NewHTTPTool(tool, nil, "svc", nil, &configv1.HttpCallDefinition{}, nil, nil, "call")
	ht.initError = errors.New("initialization failed")

	req := &ExecutionRequest{ToolName: "error-tool", ToolInputs: []byte("{}")}
	_, err := ht.Execute(context.Background(), req)
	assert.Error(t, err)
	assert.Equal(t, "initialization failed", err.Error())
}

func TestHTTPTool_Execute_NoPool(t *testing.T) {
	// Create a tool with valid init but no pool registered
	tool := &v1.Tool{Name: proto.String("tool"), UnderlyingMethodFqn: proto.String("GET http://example.com")}
	ht := NewHTTPTool(tool, pool.NewManager(), "missing-svc", nil, &configv1.HttpCallDefinition{}, nil, nil, "call")

	req := &ExecutionRequest{ToolName: "tool", ToolInputs: []byte("{}")}
	_, err := ht.Execute(context.Background(), req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no http pool found")
}

func TestHTTPTool_PrepareBody_WebhookError(t *testing.T) {
	// Test webhook error path
	tool := &v1.Tool{Name: proto.String("tool")}
	callDef := &configv1.HttpCallDefinition{
		InputTransformer: &configv1.InputTransformer{
			Webhook: &configv1.WebhookConfig{Url: "http://webhook"},
		},
	}
	ht := NewHTTPTool(tool, nil, "svc", nil, callDef, nil, nil, "call")

	inputs := map[string]any{"a": 1}
	_, _, err := ht.prepareBody(context.Background(), inputs, http.MethodGet, "tool", nil, false)
	assert.NoError(t, err)

	// If POST and webhook configured.
	// It will try to call webhook and fail (connection refused).
	_, _, err = ht.prepareBody(context.Background(), inputs, http.MethodPost, "tool", nil, false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "transformation webhook failed")
}

func TestHTTPTool_ProcessResponse_ReadError(t *testing.T) {
	tool := &v1.Tool{Name: proto.String("tool")}
	ht := NewHTTPTool(tool, nil, "svc", nil, &configv1.HttpCallDefinition{}, nil, nil, "call")

	// Mock a body that fails on Read
	resp := &http.Response{
		Body: &errReader{err: errors.New("read failed")},
	}

	_, err := ht.processResponse(context.Background(), resp)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read http response body")
}

type errReader struct {
	err error
}

func (e *errReader) Read(p []byte) (n int, err error) {
	return 0, e.err
}

func (e *errReader) Close() error {
	return nil
}

func TestHTTPTool_ProcessResponse_InvalidJSON(t *testing.T) {
	tool := &v1.Tool{Name: proto.String("tool")}
	ht := NewHTTPTool(tool, nil, "svc", nil, &configv1.HttpCallDefinition{}, nil, nil, "call")

	// Response is not JSON
	respBody := "not json"
	resp := &http.Response{
		Body: &stringReader{s: respBody},
	}

	result, err := ht.processResponse(context.Background(), resp)
	assert.NoError(t, err)
	// Expect raw string if JSON unmarshal fails
	assert.Equal(t, respBody, result)
}

type stringReader struct {
	s string
}

func (r *stringReader) Read(p []byte) (n int, err error) {
	if len(r.s) == 0 {
		return 0, io.EOF
	}
	n = copy(p, r.s)
	r.s = r.s[n:]
	return n, nil
}

func (r *stringReader) Close() error {
	return nil
}
