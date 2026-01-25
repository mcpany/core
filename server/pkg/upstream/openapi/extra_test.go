// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package openapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/proto"
)

func TestOpenAPIUpstream_Register_SpecUrl_Failures(t *testing.T) {
	ctx := context.Background()
	mockToolManager := new(MockToolManager)
	upstream := NewOpenAPIUpstream()

	t.Run("server error", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer ts.Close()

		config := configv1.UpstreamServiceConfig_builder{
			Name: proto.String("test-service-500"),
			OpenapiService: configv1.OpenapiUpstreamService_builder{
				SpecUrl: proto.String(ts.URL),
			}.Build(),
		}.Build()

		expectedKey, _ := util.SanitizeServiceName("test-service-500")
		mockToolManager.On("AddServiceInfo", expectedKey, mock.Anything).Return().Once()

		_, _, _, err := upstream.Register(ctx, config, mockToolManager, nil, nil, false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "OpenAPI spec content is missing or failed to load")
	})

	t.Run("connection error", func(t *testing.T) {
		// Use a closed port (assuming 12345 is closed)
		config := configv1.UpstreamServiceConfig_builder{
			Name: proto.String("test-service-conn-fail"),
			OpenapiService: configv1.OpenapiUpstreamService_builder{
				SpecUrl: proto.String("http://127.0.0.1:12345"),
			}.Build(),
		}.Build()

		expectedKey, _ := util.SanitizeServiceName("test-service-conn-fail")
		mockToolManager.On("AddServiceInfo", expectedKey, mock.Anything).Return().Once()

		_, _, _, err := upstream.Register(ctx, config, mockToolManager, nil, nil, false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "OpenAPI spec content is missing or failed to load")
	})
}
