// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package doctor_test

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/doctor"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestCheckService_HTTP_SSRF_Loopback(t *testing.T) {
	// Create a context
	ctx := context.Background()

	// Configure a service pointing to localhost
	svc := &configv1.UpstreamServiceConfig{
		Name: proto.String("test-service-local"),
		ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
			HttpService: &configv1.HttpUpstreamService{
				Address: proto.String("http://localhost:54321"),
			},
		},
	}

	// Run check
	result := doctor.CheckService(ctx, svc)

	// Assertions
	assert.Equal(t, doctor.StatusError, result.Status)
	// It should contain the original error
	assert.Contains(t, result.Message, "ssrf attempt blocked")

	// Check if the hint is present
	assert.Contains(t, result.Message, "Hint: To allow connections to loopback addresses (localhost), set the environment variable MCPANY_ALLOW_LOOPBACK_RESOURCES=true")
}

func TestCheckService_HTTP_SSRF_Private(t *testing.T) {
	// Create a context
	ctx := context.Background()

	// Configure a service pointing to a private IP
	svc := &configv1.UpstreamServiceConfig{
		Name: proto.String("test-service-private"),
		ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
			HttpService: &configv1.HttpUpstreamService{
				Address: proto.String("http://192.168.1.1:80"),
			},
		},
	}

	// Run check
	result := doctor.CheckService(ctx, svc)

	// Assertions
	assert.Equal(t, doctor.StatusError, result.Status)
	assert.Contains(t, result.Message, "ssrf attempt blocked")

	// Check if the hint is present
	assert.Contains(t, result.Message, "Hint: To allow connections to private network addresses, set the environment variable MCPANY_ALLOW_PRIVATE_NETWORK_RESOURCES=true")
}
