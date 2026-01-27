// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package openapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/proto"
)

type MockToolManager struct {
	mock.Mock
	tool.ManagerInterface
}

func (m *MockToolManager) AddServiceInfo(serviceID string, info *tool.ServiceInfo) {
	m.Called(serviceID, info)
}

func TestRegister_SSRFProtection(t *testing.T) {
	// 1. Start a local server (target for SSRF)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("openapi: 3.0.0\ninfo:\n  title: Test\n  version: 1.0.0\npaths: {}"))
	}))
	defer server.Close()

	// 2. Configure service to point to it
	config := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("test-service"),
		OpenapiService: configv1.OpenapiUpstreamService_builder{
			SpecUrl: proto.String(server.URL),
		}.Build(),
	}.Build()

	// 3. Ensure env vars are cleared so we test default secure behavior
	t.Setenv("MCPANY_DANGEROUS_ALLOW_LOCAL_IPS", "")
	t.Setenv("MCPANY_ALLOW_LOOPBACK_RESOURCES", "")
	t.Setenv("MCPANY_ALLOW_PRIVATE_NETWORK_RESOURCES", "")

	u := NewOpenAPIUpstream()
	tm := &MockToolManager{}
	tm.On("AddServiceInfo", mock.Anything, mock.Anything).Return()

	// 4. Try to register
	// We pass nil for managers as we expect failure before they are heavily used
	// But AddServiceInfo is called early, so we mocked it.
	_, _, _, err := u.(*OpenAPIUpstream).Register(context.Background(), config, tm, nil, nil, false)

	// 5. Assert failure
	assert.Error(t, err)
	// The error might come from client.Do or from body read failure logging, but Register checks for failure.
	// Actually, the code logs warnings on failure and returns "OpenAPI spec content is missing or failed to load"
	// if content is empty.
	if err != nil {
		// It should fail because client.Do fails, so specContent remains empty.
		// The error message from Register is "OpenAPI spec content is missing or failed to load from ..."
		assert.Contains(t, err.Error(), "OpenAPI spec content is missing or failed to load")
	}
}
