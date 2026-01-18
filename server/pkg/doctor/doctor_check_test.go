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

func TestCheckService_HTTP_Unreachable(t *testing.T) {
	// Create a context
	ctx := context.Background()

	// Configure a service pointing to an invalid port (assuming nothing runs on 54321)
	svc := &configv1.UpstreamServiceConfig{
		Name: proto.String("test-service"),
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
	assert.Contains(t, result.Message, "Failed to connect")
}

func TestCheckService_HTTP_InvalidURL(t *testing.T) {
	ctx := context.Background()
	svc := &configv1.UpstreamServiceConfig{
		Name: proto.String("test-service"),
		ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
			HttpService: &configv1.HttpUpstreamService{
				Address: proto.String("::invalid-url::"),
			},
		},
	}

	result := doctor.CheckService(ctx, svc)

	assert.Equal(t, doctor.StatusError, result.Status)
	assert.Contains(t, result.Message, "Invalid URL")
}
