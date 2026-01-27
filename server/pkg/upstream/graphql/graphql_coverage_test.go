// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package graphql

import (
	"context"
	"errors"
	"net/http"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestGraphQLUpstream_Shutdown(t *testing.T) {
	u := NewGraphQLUpstream()
	err := u.Shutdown(context.Background())
	assert.NoError(t, err)
}

func TestMapGraphQLTypeToJSONSchemaType(t *testing.T) {
	tests := []struct {
		in  string
		out string
	}{
		{"String", "string"},
		{"ID", "string"},
		{"Int", "number"},
		{"Float", "number"},
		{"Boolean", "boolean"},
		{"CustomType", "object"},
	}

	for _, tc := range tests {
		assert.Equal(t, tc.out, mapGraphQLTypeToJSONSchemaType(tc.in), "Input: %s", tc.in)
	}
}

func TestGraphQLUpstream_Register_MissingConfig(t *testing.T) {
	u := NewGraphQLUpstream()
	tm := tool.NewManager(nil)

	serviceConfig := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("missing-config"),
	}.Build()

	_, _, _, err := u.Register(context.Background(), serviceConfig, tm, nil, nil, false)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing graphql service config")
}

func TestGraphQLUpstream_Call_AuthFailure(t *testing.T) {
	mockAuth := &failingAuthenticator{err: errors.New("injected auth error")}
	callable := &Callable{
		authenticator: mockAuth,
		address:       "http://example.com",
		query:         "query { foo }",
	}

	_, err := callable.Call(context.Background(), &tool.ExecutionRequest{Arguments: map[string]interface{}{}})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to authenticate graphql query")
	assert.Contains(t, err.Error(), "injected auth error")
}

type failingAuthenticator struct {
	err error
}

func (f *failingAuthenticator) Authenticate(_ *http.Request) error {
	return f.err
}
