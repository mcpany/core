// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"context"
	"testing"

	"github.com/mcpany/core/server/pkg/pool"
	"github.com/mcpany/core/server/pkg/tool"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
)

func TestHTTPUpstream_Register_InvalidConfig(t *testing.T) {
	pm := pool.NewManager()
	tm := tool.NewManager(nil)
	upstream := NewUpstream(pm)

	testCases := []struct {
		name        string
		configJSON  string
		errorContains string
	}{
		{
			name: "missing http service",
			configJSON: `{
				"name": "missing-service"
			}`,
			errorContains: "http service config is nil",
		},
		{
			name: "missing address",
			configJSON: `{
				"name": "missing-address",
				"http_service": {}
			}`,
			errorContains: "http service address is required",
		},
		{
			name: "invalid address scheme",
			configJSON: `{
				"name": "invalid-scheme",
				"http_service": {
					"address": "ftp://example.com"
				}
			}`,
			errorContains: "invalid http service address scheme",
		},
		{
			name: "invalid address URL",
			configJSON: `{
				"name": "invalid-url",
				"http_service": {
					"address": ":/::"
				}
			}`,
			errorContains: "invalid http service address",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			serviceConfig := configv1.UpstreamServiceConfig_builder{}.Build()
			require.NoError(t, protojson.Unmarshal([]byte(tc.configJSON), serviceConfig))

			_, _, _, err := upstream.Register(context.Background(), serviceConfig, tm, nil, nil, false)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tc.errorContains)
		})
	}
}
