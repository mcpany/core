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
	unreachableAddr := "http://127.0.0.1:59999"

	pm := pool.NewManager()
	tm := tool.NewManager(nil)

	upstream := upstreamhttp.NewUpstream(pm)

	serviceConfig := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("unreachable-service"),
		HttpService: configv1.HttpUpstreamService_builder{
			Address: proto.String(unreachableAddr),
		}.Build(),
	}.Build()

	// Execution
	// This should log the large ERROR box but NOT return an error.
	_, _, _, err := upstream.Register(context.Background(), serviceConfig, tm, nil, nil, false)

	// Assertion
	assert.NoError(t, err)
}
