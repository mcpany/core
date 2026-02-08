package http

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/pool"
	"github.com/mcpany/core/server/pkg/resource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
)

func TestCoverageEnhancement_DynamicResourceErrors(t *testing.T) {
	pm := pool.NewManager()
	mockTm := newMockToolManager()
	upstream := NewUpstream(pm)
	rm := resource.NewManager()

	configJSON := `{
		"name": "resource-error-service",
		"http_service": {
			"address": "http://127.0.0.1",
			"tools": [
				{"name": "op1", "call_id": "op1-call"}
			],
			"calls": {
				"op1-call": {
					"id": "op1-call",
					"method": "HTTP_METHOD_GET",
                    "endpoint_path": "/op1"
				}
			},
			"resources": [
				{
					"name": "res-invalid-call",
					"uri": "http://res1",
					"dynamic": {
						"http_call": {"id": "non-existent-call"}
					}
				},
				{
					"name": "res-tool-not-found",
					"uri": "http://res2",
					"dynamic": {
						"http_call": {"id": "op1-call"}
					}
				}
			]
		}
	}`
	serviceConfig := configv1.UpstreamServiceConfig_builder{}.Build()
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

    // Make AddTool fail so op1 is not registered
    mockTm.addError = assert.AnError

	_, _, _, err := upstream.Register(context.Background(), serviceConfig, mockTm, nil, rm, false)
	require.NoError(t, err)

    // Verify resources were NOT registered
    _, ok := rm.GetResource("http://res1")
    assert.False(t, ok, "Resource with invalid call ID should not be registered")

    _, ok = rm.GetResource("http://res2")
    assert.False(t, ok, "Resource with missing tool should not be registered")
}

func TestCoverageEnhancement_InvalidEndpointPath(t *testing.T) {
	pm := pool.NewManager()
	mockTm := newMockToolManager()
	upstream := NewUpstream(pm)

    // \u007f is DEL, which url.Parse should reject as it's a control character.
	configJSON := `{
		"name": "invalid-path-service",
		"http_service": {
			"address": "http://127.0.0.1",
			"tools": [
				{"name": "op1", "call_id": "op1-call"}
			],
			"calls": {
				"op1-call": {
					"id": "op1-call",
					"method": "HTTP_METHOD_GET",
                    "endpoint_path": "/\u007f"
				}
			}
		}
	}`

	serviceConfig := configv1.UpstreamServiceConfig_builder{}.Build()
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

	_, discoveredTools, _, err := upstream.Register(context.Background(), serviceConfig, mockTm, nil, nil, false)
	require.NoError(t, err)

    // Expect tool to NOT be registered because url.Parse fails
    assert.Empty(t, discoveredTools)
}

func TestCoverageEnhancement_InvalidQueryEncoding(t *testing.T) {
	pm := pool.NewManager()
	mockTm := newMockToolManager()
	upstream := NewUpstream(pm)

    // %GG is invalid percent encoding
	configJSON := `{
		"name": "invalid-query-service",
		"http_service": {
			"address": "http://127.0.0.1",
			"tools": [
				{"name": "op1", "call_id": "op1-call"}
			],
			"calls": {
				"op1-call": {
					"id": "op1-call",
					"method": "HTTP_METHOD_GET",
                    "endpoint_path": "/test?q=%GG"
				}
			}
		}
	}`
	serviceConfig := configv1.UpstreamServiceConfig_builder{}.Build()
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

	_, discoveredTools, _, err := upstream.Register(context.Background(), serviceConfig, mockTm, nil, nil, false)
	require.NoError(t, err)

    // The tool should be registered, but the invalid part in query should be handled (preserved)
    assert.NotEmpty(t, discoveredTools)

    tools := mockTm.ListTools()
    require.Len(t, tools, 1)
    tTool := tools[0]

    fqn := tTool.Tool().GetUnderlyingMethodFqn()
    // It should preserve invalid query part
    assert.Contains(t, fqn, "%GG")
}

func TestCoverageEnhancement_MTLSError(t *testing.T) {
	pm := pool.NewManager()
	mockTm := newMockToolManager()
	upstream := NewUpstream(pm)

	configJSON := `{
		"name": "mtls-error-service",
		"http_service": {
			"address": "https://127.0.0.1",
			"tools": [
				{"name": "op1", "call_id": "op1-call"}
			],
			"calls": {
				"op1-call": {
					"id": "op1-call",
					"method": "HTTP_METHOD_GET",
                    "endpoint_path": "/op1"
				}
			}
		},
        "upstream_auth": {
            "mtls": {
                "client_cert_path": "/non/existent/cert.pem",
                "client_key_path": "/non/existent/key.pem",
                "ca_cert_path": "/non/existent/ca.pem"
            }
        }
	}`
	serviceConfig := configv1.UpstreamServiceConfig_builder{}.Build()
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

	_, _, _, err := upstream.Register(context.Background(), serviceConfig, mockTm, nil, nil, false)
    // NewHTTPPool should fail due to missing certs
	require.Error(t, err)
    assert.Contains(t, err.Error(), "failed to create HTTP pool")
}

func TestCoverageEnhancement_URLConstruction(t *testing.T) {
    // Tests for specific URL construction edge cases including double slash
	pm := pool.NewManager()
	mockTm := newMockToolManager()
	upstream := NewUpstream(pm)

    testCases := []struct {
        endpointPath string
        expectedPathSuffix string // suffix of the URL
    }{
        {"//foo", "/foo"}, // //foo -> host=foo path="" -> path=//foo (relative) -> appended to base
        {"//", "/"},       // // -> host="" path="" -> path=// -> / (relative)
        {"///foo", "///foo"},
        {"/foo%2Fbar", "/foo%2Fbar"}, // RawPath preservation
    }

    for _, tc := range testCases {
        configJSON := `{
            "name": "url-edge-service",
            "http_service": {
                "address": "http://127.0.0.1/base",
                "tools": [
                    {"name": "op1", "call_id": "op1-call"}
                ],
                "calls": {
                    "op1-call": {
                        "id": "op1-call",
                        "method": "HTTP_METHOD_GET",
                        "endpoint_path": "` + tc.endpointPath + `"
                    }
                }
            }
        }`
        serviceConfig := configv1.UpstreamServiceConfig_builder{}.Build()
        require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

        // Clear previous tools
        mockTm.addedTools = nil

        _, discoveredTools, _, err := upstream.Register(context.Background(), serviceConfig, mockTm, nil, nil, false)
        require.NoError(t, err)
        require.Len(t, discoveredTools, 1)

        tools := mockTm.ListTools()
        require.Len(t, tools, 1)
        fqn := tools[0].Tool().GetUnderlyingMethodFqn()

        assert.Contains(t, fqn, tc.expectedPathSuffix)
    }
}

func TestCoverageEnhancement_ExportPolicy(t *testing.T) {
	pm := pool.NewManager()
	mockTm := newMockToolManager()
	upstream := NewUpstream(pm)
    rm := resource.NewManager()

	configJSON := `{
		"name": "export-policy-service",
		"http_service": {
			"address": "http://127.0.0.1",
			"tools": [
				{"name": "public-tool", "call_id": "c1"},
                {"name": "private-tool", "call_id": "c2"}
			],
			"calls": {
				"c1": {"id": "c1", "method": "HTTP_METHOD_GET"},
                "c2": {"id": "c2", "method": "HTTP_METHOD_GET"}
			}
		},
        "tool_export_policy": {
            "default_action": "EXPORT",
            "rules": [
                {"name_regex": "^private-.*", "action": "UNEXPORT"}
            ]
        }
	}`
	serviceConfig := configv1.UpstreamServiceConfig_builder{}.Build()
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

	_, discoveredTools, _, err := upstream.Register(context.Background(), serviceConfig, mockTm, nil, rm, false)
	require.NoError(t, err)

    // Should only have public-tool
    require.Len(t, discoveredTools, 1)
    assert.Equal(t, "public-tool", discoveredTools[0].GetName())
}

func TestCoverageEnhancement_InvalidCallPolicy(t *testing.T) {
	pm := pool.NewManager()
	mockTm := newMockToolManager()
	upstream := NewUpstream(pm)

	configJSON := `{
		"name": "invalid-policy-service",
		"http_service": {
			"address": "http://127.0.0.1",
			"tools": [
				{"name": "op1", "call_id": "c1"}
			],
			"calls": {
				"c1": {"id": "c1", "method": "HTTP_METHOD_GET"}
			}
		},
        "call_policies": [
            {
                "rules": [
                    {"name_regex": "[", "action": "DENY"}
                ],
                "default_action": "ALLOW"
            }
        ]
	}`
    // "[" is invalid regex.

	serviceConfig := configv1.UpstreamServiceConfig_builder{}.Build()
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

	_, discoveredTools, _, err := upstream.Register(context.Background(), serviceConfig, mockTm, nil, nil, false)
    // Register should fail because CompileCallPolicies fails
	require.Nil(t, discoveredTools)
    assert.NoError(t, err) // It doesn't return error, just logs it and returns nil tools.
}

func TestCoverageEnhancement_DynamicResource_EmptyToolName(t *testing.T) {
	pm := pool.NewManager()
	mockTm := newMockToolManager()
	upstream := NewUpstream(pm)
    rm := resource.NewManager()

    // Tool without name (gets auto-generated), but we reference it in resource.
	configJSON := `{
		"name": "empty-tool-name-service",
		"http_service": {
			"address": "http://127.0.0.1",
			"tools": [
				{"call_id": "c1"}
			],
			"calls": {
				"c1": {"id": "c1", "method": "HTTP_METHOD_GET"}
			},
            "resources": [
                {
                    "name": "res1",
                    "uri": "http://res1",
                    "dynamic": {
                        "http_call": {"id": "c1"}
                    }
                }
            ]
		}
	}`
	serviceConfig := configv1.UpstreamServiceConfig_builder{}.Build()
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

	_, _, _, err := upstream.Register(context.Background(), serviceConfig, mockTm, nil, rm, false)
	require.NoError(t, err)

    // Resource should NOT be registered
    _, ok := rm.GetResource("http://res1")
    assert.False(t, ok, "Resource with unnamed tool should not be registered (due to sanitization failure)")
}

func TestCoverageEnhancement_DynamicResource_NoCall(t *testing.T) {
	pm := pool.NewManager()
	mockTm := newMockToolManager()
	upstream := NewUpstream(pm)
    rm := resource.NewManager()

	configJSON := `{
		"name": "no-call-resource-service",
		"http_service": {
			"address": "http://127.0.0.1",
			"tools": [
				{"name": "op1", "call_id": "c1"}
			],
			"calls": {
				"c1": {"id": "c1", "method": "HTTP_METHOD_GET"}
			},
            "resources": [
                {
                    "name": "res1",
                    "uri": "http://res1",
                    "dynamic": {
                    }
                }
            ]
		}
	}`
    // dynamic is present but empty. GetHttpCall() returns nil.

	serviceConfig := configv1.UpstreamServiceConfig_builder{}.Build()
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

	_, _, _, err := upstream.Register(context.Background(), serviceConfig, mockTm, nil, rm, false)
	require.NoError(t, err)

    // Resource should NOT be registered
    _, ok := rm.GetResource("http://res1")
    assert.False(t, ok)
}

func TestCoverageEnhancement_QueryFlags(t *testing.T) {
	pm := pool.NewManager()
	mockTm := newMockToolManager()
	upstream := NewUpstream(pm)

    // Case 1: Base ?f, Endpoint ?f -> ?f (flag preserved)
    // Case 2: Base ?f, Endpoint ?f= -> ?f (flag preserved due to restore logic)

	configJSON := `{
		"name": "query-flag-service",
		"http_service": {
			"address": "http://127.0.0.1?f",
			"tools": [
				{"name": "op1", "call_id": "c1"},
                {"name": "op2", "call_id": "c2"}
			],
			"calls": {
				"c1": {
                    "id": "c1",
                    "method": "HTTP_METHOD_GET",
                    "endpoint_path": "?f"
                },
                "c2": {
                    "id": "c2",
                    "method": "HTTP_METHOD_GET",
                    "endpoint_path": "?f="
                }
			}
		}
	}`
	serviceConfig := configv1.UpstreamServiceConfig_builder{}.Build()
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

	_, _, _, err := upstream.Register(context.Background(), serviceConfig, mockTm, nil, nil, false)
	require.NoError(t, err)

    tools := mockTm.ListTools()
    require.Len(t, tools, 2)

    // Check order/mapping
    for _, tool := range tools {
        fqn := tool.Tool().GetUnderlyingMethodFqn()
        name := tool.Tool().GetName() // "op1" or "op2"

        if name == "op1" {
            // op1 uses endpoint "?f", should be a flag
            assert.Contains(t, fqn, "?f", "Tool %s URL %s should contain flag ?f", name, fqn)
            assert.NotContains(t, fqn, "?f=", "Tool %s URL %s should NOT contain ?f=", name, fqn)
        } else {
            // op2 uses endpoint "?f=", should have equals
            assert.Contains(t, fqn, "?f=", "Tool %s URL %s should contain ?f=", name, fqn)
        }
    }
}

func TestCoverageEnhancement_InputSchemaOverlap(t *testing.T) {
	pm := pool.NewManager()
	mockTm := newMockToolManager()
	upstream := NewUpstream(pm)

    // Input Schema required: ["foo"]
    // Parameters required: ["foo"]
    // They overlap. Code path: check existing, find it, break.

	configJSON := `{
		"name": "overlap-service",
		"http_service": {
			"address": "http://127.0.0.1",
			"tools": [
				{"name": "op1", "call_id": "c1"}
			],
			"calls": {
				"c1": {
                    "id": "c1",
                    "method": "HTTP_METHOD_GET",
                    "input_schema": {
                        "type": "object",
                        "properties": {"foo": {"type": "string"}},
                        "required": ["foo"]
                    },
                    "parameters": [
                        {
                            "schema": {
                                "name": "foo",
                                "type": "STRING",
                                "is_required": true
                            }
                        }
                    ]
                }
			}
		}
	}`
	serviceConfig := configv1.UpstreamServiceConfig_builder{}.Build()
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

	_, _, _, err := upstream.Register(context.Background(), serviceConfig, mockTm, nil, nil, false)
	require.NoError(t, err)

    tools := mockTm.ListTools()
    require.Len(t, tools, 1)

    schema := tools[0].Tool().GetAnnotations().GetInputSchema()
    reqVal := schema.Fields["required"].GetListValue()
    require.Len(t, reqVal.Values, 1)
    assert.Equal(t, "foo", reqVal.Values[0].GetStringValue())
}

func TestCoverageEnhancement_EmptyAddress(t *testing.T) {
	pm := pool.NewManager()
	mockTm := newMockToolManager()
	upstream := NewUpstream(pm)

	configJSON := `{
		"name": "empty-address-service",
		"http_service": {
			"address": "",
			"tools": [
				{"name": "op1", "call_id": "c1"}
			],
			"calls": {
				"c1": {"id": "c1", "method": "HTTP_METHOD_GET"}
			}
		}
	}`
	serviceConfig := configv1.UpstreamServiceConfig_builder{}.Build()
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

	_, _, _, err := upstream.Register(context.Background(), serviceConfig, mockTm, nil, nil, false)
	require.Error(t, err)
    assert.Contains(t, err.Error(), "address is required")
}

func TestCoverageEnhancement_InputSchema_InvalidPropertiesType(t *testing.T) {
	pm := pool.NewManager()
	mockTm := newMockToolManager()
	upstream := NewUpstream(pm)

    // input_schema.properties is NOT a struct (object). e.g. a number.
    // "input_schema": { "properties": 123 }

	configJSON := `{
		"name": "invalid-props-service",
		"http_service": {
			"address": "http://127.0.0.1",
			"tools": [
				{"name": "op1", "call_id": "c1"}
			],
			"calls": {
				"c1": {
                    "id": "c1",
                    "method": "HTTP_METHOD_GET",
                    "input_schema": {
                        "type": "object",
                        "properties": 123
                    },
                    "parameters": [
                        {
                            "schema": {
                                "name": "foo",
                                "type": "STRING",
                                "is_required": true
                            }
                        }
                    ]
                }
			}
		}
	}`
	serviceConfig := configv1.UpstreamServiceConfig_builder{}.Build()
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

	_, _, _, err := upstream.Register(context.Background(), serviceConfig, mockTm, nil, nil, false)
	require.NoError(t, err)

    tools := mockTm.ListTools()
    require.Len(t, tools, 1)

    schema := tools[0].Tool().GetAnnotations().GetInputSchema()
    // It should have overwritten properties with new struct containing foo
    propsVal := schema.Fields["properties"].GetStructValue()
    require.NotNil(t, propsVal)
    _, ok := propsVal.Fields["foo"]
    assert.True(t, ok)
}

func generateCertFiles(t *testing.T, dir string) (string, string, string) {
    // Generate CA
    caKey, err := rsa.GenerateKey(rand.Reader, 2048)
    require.NoError(t, err)

    caTmpl := x509.Certificate{
        SerialNumber: big.NewInt(1),
        Subject:      pkix.Name{CommonName: "Test CA"},
        NotBefore:    time.Now(),
        NotAfter:     time.Now().Add(time.Hour),
        KeyUsage:     x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
        IsCA:         true,
        BasicConstraintsValid: true,
    }

    caDer, err := x509.CreateCertificate(rand.Reader, &caTmpl, &caTmpl, &caKey.PublicKey, caKey)
    require.NoError(t, err)

    caPath := filepath.Join(dir, "ca.pem")
    caFile, err := os.Create(caPath)
    require.NoError(t, err)
    pem.Encode(caFile, &pem.Block{Type: "CERTIFICATE", Bytes: caDer})
    caFile.Close()

    // Generate Client Cert
    clientKey, err := rsa.GenerateKey(rand.Reader, 2048)
    require.NoError(t, err)

    clientTmpl := x509.Certificate{
        SerialNumber: big.NewInt(2),
        Subject:      pkix.Name{CommonName: "Test Client"},
        NotBefore:    time.Now(),
        NotAfter:     time.Now().Add(time.Hour),
        KeyUsage:     x509.KeyUsageDigitalSignature,
        ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
    }

    clientDer, err := x509.CreateCertificate(rand.Reader, &clientTmpl, &caTmpl, &clientKey.PublicKey, caKey)
    require.NoError(t, err)

    certPath := filepath.Join(dir, "client.pem")
    certFile, err := os.Create(certPath)
    require.NoError(t, err)
    pem.Encode(certFile, &pem.Block{Type: "CERTIFICATE", Bytes: clientDer})
    certFile.Close()

    keyPath := filepath.Join(dir, "client.key")
    keyFile, err := os.Create(keyPath)
    require.NoError(t, err)
    pem.Encode(keyFile, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(clientKey)})
    keyFile.Close()

    return caPath, certPath, keyPath
}

func TestCoverageEnhancement_MTLSSuccess(t *testing.T) {
	pm := pool.NewManager()
	mockTm := newMockToolManager()
	upstream := NewUpstream(pm)

    tmpDir := t.TempDir()
    caPath, certPath, keyPath := generateCertFiles(t, tmpDir)

	configJSON := `{
		"name": "mtls-success-service",
		"http_service": {
			"address": "https://127.0.0.1",
			"tools": [
				{"name": "op1", "call_id": "op1-call"}
			],
			"calls": {
				"op1-call": {
					"id": "op1-call",
					"method": "HTTP_METHOD_GET",
                    "endpoint_path": "/op1"
				}
			}
		},
        "upstream_auth": {
            "mtls": {
                "client_cert_path": "` + certPath + `",
                "client_key_path": "` + keyPath + `",
                "ca_cert_path": "` + caPath + `"
            }
        }
	}`
	serviceConfig := configv1.UpstreamServiceConfig_builder{}.Build()
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

	_, _, _, err := upstream.Register(context.Background(), serviceConfig, mockTm, nil, nil, false)
	require.NoError(t, err)
}

func TestCoverageEnhancement_AutoDiscovery(t *testing.T) {
	pm := pool.NewManager()
	mockTm := newMockToolManager()
	upstream := NewUpstream(pm)

    // 1. Auto-discover: Tool not defined -> Generated
    // 2. Auto-discover: Tool defined -> Skipped (existing)

	configJSON := `{
		"name": "autodisc-service",
        "auto_discover_tool": true,
		"http_service": {
			"address": "http://127.0.0.1",
			"tools": [
				{"name": "explicit-tool", "call_id": "explicit-call"}
			],
			"calls": {
				"explicit-call": {
					"id": "explicit-call",
					"method": "HTTP_METHOD_GET"
				},
                "implicit-call": {
                    "id": "implicit-call",
                    "method": "HTTP_METHOD_GET"
                }
			}
		}
	}`
	serviceConfig := configv1.UpstreamServiceConfig_builder{}.Build()
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

	_, _, _, err := upstream.Register(context.Background(), serviceConfig, mockTm, nil, nil, false)
	require.NoError(t, err)

    tools := mockTm.ListTools()
    // Explicit tool + Implicit tool = 2
    require.Len(t, tools, 2)

    names := make(map[string]bool)
    for _, t := range tools {
        names[t.Tool().GetName()] = true
    }

    assert.True(t, names["explicit-tool"], "Explicit tool should be present")
    assert.True(t, names["implicit-call"], "Implicit tool should be present (named after call ID)")
}

func TestCoverageEnhancement_InputSchema_EmptyStruct(t *testing.T) {
	pm := pool.NewManager()
	mockTm := newMockToolManager()
	upstream := NewUpstream(pm)

    // Input schema present but empty fields. Should fall back to generating from parameters.

	configJSON := `{
		"name": "empty-schema-service",
		"http_service": {
			"address": "http://127.0.0.1",
			"tools": [
				{"name": "op1", "call_id": "c1"}
			],
			"calls": {
				"c1": {
                    "id": "c1",
                    "method": "HTTP_METHOD_GET",
                    "input_schema": {},
                    "parameters": [
                        {
                            "schema": {
                                "name": "foo",
                                "type": "STRING"
                            }
                        }
                    ]
                }
			}
		}
	}`
	serviceConfig := configv1.UpstreamServiceConfig_builder{}.Build()
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

	_, _, _, err := upstream.Register(context.Background(), serviceConfig, mockTm, nil, nil, false)
	require.NoError(t, err)

    tools := mockTm.ListTools()
    require.Len(t, tools, 1)

    schema := tools[0].Tool().GetAnnotations().GetInputSchema()
    // Check if properties were populated
    propsVal := schema.Fields["properties"].GetStructValue()
    require.NotNil(t, propsVal)
    _, ok := propsVal.Fields["foo"]
    assert.True(t, ok, "foo should be present in properties")
}
