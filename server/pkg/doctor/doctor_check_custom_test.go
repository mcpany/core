// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package doctor_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/doctor"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestCheckService_HTTP_CustomHealthCheck(t *testing.T) {
	t.Setenv("MCPANY_ALLOW_LOOPBACK_RESOURCES", "true")
	// Start a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" && r.Method == "GET" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("OK"))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	ctx := context.Background()

	// Configure service with explicit health check pointing to /health
	svc := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("test-custom-health"),
		HttpService: configv1.HttpUpstreamService_builder{
			Address: proto.String(server.URL + "/api"), // API path (returns 404)
			HealthCheck: configv1.HttpHealthCheck_builder{
				Url:          proto.String(server.URL + "/health"),
				ExpectedCode: proto.Int32(200),
				Method:       proto.String("GET"),
			}.Build(),
		}.Build(),
	}.Build()

	// Run check
	result := doctor.CheckService(ctx, svc)

	// Assertions
	assert.Equal(t, doctor.StatusOk, result.Status)
	assert.Contains(t, result.Message, "Service reachable")
}

func TestCheckService_HTTP_CustomHealthCheck_Failure(t *testing.T) {
	t.Setenv("MCPANY_ALLOW_LOOPBACK_RESOURCES", "true")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	ctx := context.Background()

	svc := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("test-custom-health-fail"),
		HttpService: configv1.HttpUpstreamService_builder{
			Address: proto.String(server.URL),
			HealthCheck: configv1.HttpHealthCheck_builder{
				Url:          proto.String(server.URL),
				ExpectedCode: proto.Int32(200),
			}.Build(),
		}.Build(),
	}.Build()

	result := doctor.CheckService(ctx, svc)

	assert.Equal(t, doctor.StatusError, result.Status)
	assert.Contains(t, result.Message, "Unexpected status code")
}
