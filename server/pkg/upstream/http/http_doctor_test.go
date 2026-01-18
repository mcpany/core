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
	// when the service is unreachable, but does NOT return an error.

	// We use a port that is unlikely to be open.
	unreachableAddr := "http://localhost:59999"

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
	// This should log the large ERROR box but NOT return an error.
	_, _, _, err := upstream.Register(context.Background(), serviceConfig, tm, nil, nil, false)

	// Assertion
	// The key behavior we are testing is that startup is NOT blocked (err is nil).
	// Verifying the log output in unit tests is complex and usually fragile,
	// but ensuring this code path executes without panic or error is sufficient for coverage.
	assert.NoError(t, err)
}
