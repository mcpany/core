// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package http_test

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/pool"
	"github.com/mcpany/core/server/pkg/tool"
	upstreamhttp "github.com/mcpany/core/server/pkg/upstream/http"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestHTTPUpstream_Register_Unreachable(t *testing.T) {
	// This test verifies that the Register method calls the doctor check (via log emission)
	// when the service is unreachable, and returns an error (blocking registration).

	// We use a port that is unlikely to be open.
	unreachableAddr := "http://127.0.0.1:59999"

	pm := pool.NewManager()
	// Use a nil tool manager to minimize dependencies, or a mock if needed.
	// Since the code calls AddServiceInfo, we need a functional or mock ToolManager.
	tm := tool.NewManager(nil)

	upstream := upstreamhttp.NewUpstream(pm)

	serviceConfig := &configv1.UpstreamServiceConfig{
		Name: proto.String("unreachable-service"),
		ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
			HttpService: &configv1.HttpUpstreamService{
				Address: proto.String(unreachableAddr),
			},
		},
	}

	// Execution
	// This should log the large ERROR box and return an error.
	// We do NOT use the skip flag here because we want to verify the failure logic.
	_, _, _, err := upstream.Register(context.Background(), serviceConfig, tm, nil, nil, false)

	// Assertion
	// The key behavior we are testing is that startup IS blocked (err is NOT nil) for unreachable services.
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to connect")
}
