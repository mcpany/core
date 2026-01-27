// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"context"
	"os"
	"testing"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/pool"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestHTTPUpstream_Coverage_InputSchemaMerge(t *testing.T) {
	pm := pool.NewManager()
	tm := tool.NewManager(nil)
	upstream := NewUpstream(pm)

	// Case: input_schema AND parameters provided.
	// We want to ensure that properties and required fields are merged.
	configJSON := `{
		"name": "schema-merge-service",
		"http_service": {
			"address": "http://127.0.0.1",
			"tools": [{"name": "tool1", "call_id": "call1"}],
			"calls": {
				"call1": {
					"id": "call1",
					"method": "HTTP_METHOD_GET",
					"input_schema": {
						"type": "object",
						"properties": {
							"existing_prop": {"type": "string"}
						},
						"required": ["existing_prop"]
					},
					"parameters": [
						{
							"schema": {
								"name": "new_param",
								"type": "STRING",
								"is_required": true
							}
						}
					]
				}
			}
		}
	}`
	serviceConfig := &configv1.UpstreamServiceConfig{}
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

	_, tools, _, err := upstream.Register(context.Background(), serviceConfig, tm, nil, nil, false)
	require.NoError(t, err)
	require.Len(t, tools, 1)

	toolDef := tm.ListTools()[0]
	schema := toolDef.Tool().GetAnnotations().GetInputSchema()
	fields := schema.GetFields()

	// Check properties
	props := fields["properties"].GetStructValue().GetFields()
	assert.Contains(t, props, "existing_prop")
	assert.Contains(t, props, "new_param")

	// Check required
	req := fields["required"].GetListValue().GetValues()
	reqStr := make([]string, len(req))
	for i, v := range req {
		reqStr[i] = v.GetStringValue()
	}
	assert.Contains(t, reqStr, "existing_prop")
	assert.Contains(t, reqStr, "new_param")
}

func TestHTTPUpstream_Coverage_InputSchemaMerge_NoProperties(t *testing.T) {
	pm := pool.NewManager()
	tm := tool.NewManager(nil)
	upstream := NewUpstream(pm)

	// Case: input_schema has no properties, but parameters exist.
	// Logic should create "properties" field.
	configJSON := `{
		"name": "schema-merge-noprops-service",
		"http_service": {
			"address": "http://127.0.0.1",
			"tools": [{"name": "tool1", "call_id": "call1"}],
			"calls": {
				"call1": {
					"id": "call1",
					"method": "HTTP_METHOD_GET",
					"input_schema": {
						"type": "object"
					},
					"parameters": [
						{
							"schema": {
								"name": "new_param",
								"type": "STRING",
								"is_required": true
							}
						}
					]
				}
			}
		}
	}`
	serviceConfig := &configv1.UpstreamServiceConfig{}
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

	_, tools, _, err := upstream.Register(context.Background(), serviceConfig, tm, nil, nil, false)
	require.NoError(t, err)
	require.Len(t, tools, 1)

	toolDef := tm.ListTools()[0]
	schema := toolDef.Tool().GetAnnotations().GetInputSchema()
	fields := schema.GetFields()

	require.Contains(t, fields, "properties")
	props := fields["properties"].GetStructValue().GetFields()
	assert.Contains(t, props, "new_param")
}

func TestHTTPUpstream_Coverage_DoubleSlashEmpty(t *testing.T) {
	pm := pool.NewManager()
	tm := tool.NewManager(nil)
	upstream := NewUpstream(pm)

	// Case: endpoint_path is "//"
	configJSON := `{
		"name": "double-slash-empty",
		"http_service": {
			"address": "http://127.0.0.1",
			"tools": [{"name": "tool1", "call_id": "call1"}],
			"calls": {
				"call1": {
					"id": "call1",
					"method": "HTTP_METHOD_GET",
					"endpoint_path": "//"
				}
			}
		}
	}`
	serviceConfig := &configv1.UpstreamServiceConfig{}
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

	_, tools, _, err := upstream.Register(context.Background(), serviceConfig, tm, nil, nil, false)
	require.NoError(t, err)
	require.Len(t, tools, 1)

	fqn := tm.ListTools()[0].Tool().GetUnderlyingMethodFqn()
	assert.Contains(t, fqn, "//")
}

func TestHTTPUpstream_Coverage_DoubleSlashHost(t *testing.T) {
	pm := pool.NewManager()
	tm := tool.NewManager(nil)
	upstream := NewUpstream(pm)

	// Case: endpoint_path starts with // and has host-like part
	configJSON := `{
		"name": "double-slash-host",
		"http_service": {
			"address": "http://127.0.0.1",
			"tools": [{"name": "tool1", "call_id": "call1"}],
			"calls": {
				"call1": {
					"id": "call1",
					"method": "HTTP_METHOD_GET",
					"endpoint_path": "//foo/bar"
				}
			}
		}
	}`
	serviceConfig := &configv1.UpstreamServiceConfig{}
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

	_, tools, _, err := upstream.Register(context.Background(), serviceConfig, tm, nil, nil, false)
	require.NoError(t, err)
	require.Len(t, tools, 1)

	fqn := tm.ListTools()[0].Tool().GetUnderlyingMethodFqn()
	assert.Contains(t, fqn, "//foo/bar")
}

func TestHTTPUpstream_Coverage_QueryMergeInvalidKey(t *testing.T) {
	pm := pool.NewManager()
	tm := tool.NewManager(nil)
	upstream := NewUpstream(pm)

	// Case: Base URL has invalid query key (%ZZ)
	// Endpoint has query
	configJSON := `{
		"name": "query-merge-invalid",
		"http_service": {
			"address": "http://127.0.0.1?%ZZ=1",
			"tools": [{"name": "tool1", "call_id": "call1"}],
			"calls": {
				"call1": {
					"id": "call1",
					"method": "HTTP_METHOD_GET",
					"endpoint_path": "/path?a=2"
				}
			}
		}
	}`
	serviceConfig := &configv1.UpstreamServiceConfig{}
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

	_, tools, _, err := upstream.Register(context.Background(), serviceConfig, tm, nil, nil, false)
	require.NoError(t, err)
	require.Len(t, tools, 1)

	fqn := tm.ListTools()[0].Tool().GetUnderlyingMethodFqn()
	// %ZZ=1 should be preserved
	assert.Contains(t, fqn, "%ZZ=1")
	assert.Contains(t, fqn, "a=2")
}

func TestHTTPUpstream_Coverage_QueryMergeInvalidKeyEndpoint(t *testing.T) {
	pm := pool.NewManager()
	tm := tool.NewManager(nil)
	upstream := NewUpstream(pm)

	// Case: Endpoint URL has invalid query key (%ZZ)
	configJSON := `{
		"name": "query-merge-invalid-ep",
		"http_service": {
			"address": "http://127.0.0.1",
			"tools": [{"name": "tool1", "call_id": "call1"}],
			"calls": {
				"call1": {
					"id": "call1",
					"method": "HTTP_METHOD_GET",
					"endpoint_path": "/path?%ZZ=2"
				}
			}
		}
	}`
	serviceConfig := &configv1.UpstreamServiceConfig{}
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

	_, tools, _, err := upstream.Register(context.Background(), serviceConfig, tm, nil, nil, false)
	require.NoError(t, err)
	require.Len(t, tools, 1)

	fqn := tm.ListTools()[0].Tool().GetUnderlyingMethodFqn()
	assert.Contains(t, fqn, "%ZZ=2")
}

func TestHTTPUpstream_Coverage_InputSchemaRequiredMerge_NonString(t *testing.T) {
	pm := pool.NewManager()
	tm := tool.NewManager(nil)
	upstream := NewUpstream(pm)

	// Case: input_schema.required contains non-string values.
	// Code should ignore them.
	configJSON := `{
		"name": "schema-req-nonstring",
		"http_service": {
			"address": "http://127.0.0.1",
			"tools": [{"name": "tool1", "call_id": "call1"}],
			"calls": {
				"call1": {
					"id": "call1",
					"method": "HTTP_METHOD_GET",
					"input_schema": {
						"type": "object",
						"required": ["valid_req", 123]
					},
					"parameters": [
						{
							"schema": {
								"name": "new_param",
								"type": "STRING",
								"is_required": true
							}
						}
					]
				}
			}
		}
	}`
	serviceConfig := &configv1.UpstreamServiceConfig{}
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

	_, tools, _, err := upstream.Register(context.Background(), serviceConfig, tm, nil, nil, false)
	require.NoError(t, err)
	require.Len(t, tools, 1)

	toolDef := tm.ListTools()[0]
	schema := toolDef.Tool().GetAnnotations().GetInputSchema()
	fields := schema.GetFields()

	req := fields["required"].GetListValue().GetValues()
	reqStr := make([]string, 0)
	for _, v := range req {
		if s, ok := v.GetKind().(*structpb.Value_StringValue); ok {
			reqStr = append(reqStr, s.StringValue)
		}
	}

	assert.Contains(t, reqStr, "valid_req")
	assert.Contains(t, reqStr, "new_param")
	assert.NotContains(t, reqStr, "123") // Should not be string "123"
	assert.Len(t, reqStr, 2)
}

func TestHTTPPool_New_MTLS_Errors(t *testing.T) {
	// Create temp files for certs
	certFile, err := os.CreateTemp("", "cert")
	require.NoError(t, err)
	defer os.Remove(certFile.Name())

	keyFile, err := os.CreateTemp("", "key")
	require.NoError(t, err)
	defer os.Remove(keyFile.Name())

	caFile, err := os.CreateTemp("", "ca")
	require.NoError(t, err)
	defer os.Remove(caFile.Name())

	// Case 1: Invalid Cert Path
	configJSON := `{
		"name": "mtls-fail-1",
		"http_service": {
			"address": "http://127.0.0.1"
		},
		"upstream_auth": {
			"mtls": {
				"client_cert_path": "/non/existent/cert",
				"client_key_path": "/non/existent/key",
				"ca_cert_path": "/non/existent/ca"
			}
		}
	}`
	serviceConfig := &configv1.UpstreamServiceConfig{}
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

	_, err = NewHTTPPool(1, 1, time.Second, serviceConfig)
	assert.Error(t, err)

	// Case 2: Valid paths but invalid content (LoadX509KeyPair fails)
	// Write garbage to files
	os.WriteFile(certFile.Name(), []byte("garbage"), 0600)
	os.WriteFile(keyFile.Name(), []byte("garbage"), 0600)

	configJSON2 := `{
		"name": "mtls-fail-2",
		"http_service": {
			"address": "http://127.0.0.1"
		},
		"upstream_auth": {
			"mtls": {
				"client_cert_path": "` + certFile.Name() + `",
				"client_key_path": "` + keyFile.Name() + `",
				"ca_cert_path": "` + caFile.Name() + `"
			}
		}
	}`
	serviceConfig2 := &configv1.UpstreamServiceConfig{}
	require.NoError(t, protojson.Unmarshal([]byte(configJSON2), serviceConfig2))

	_, err = NewHTTPPool(1, 1, time.Second, serviceConfig2)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "tls: failed to find any PEM data")
}

func TestHTTPPool_EnvVars(t *testing.T) {
	t.Setenv("MCPANY_ALLOW_LOOPBACK_RESOURCES", "true")
	t.Setenv("MCPANY_ALLOW_PRIVATE_NETWORK_RESOURCES", "true")

	configJSON := `{
		"name": "env-vars-pool",
		"http_service": {
			"address": "http://127.0.0.1"
		}
	}`
	serviceConfig := &configv1.UpstreamServiceConfig{}
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

	p, err := NewHTTPPool(1, 1, time.Second, serviceConfig)
	require.NoError(t, err)
	defer func() { _ = p.Close() }()

	// Verify ForceAttemptHTTP2 in transport
	hp, ok := p.(*httpPool)
	require.True(t, ok)
	assert.True(t, hp.transport.ForceAttemptHTTP2)
}
