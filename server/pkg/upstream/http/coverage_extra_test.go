package http

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alexliesenfeld/health"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/pool"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestHTTPUpstream_CheckHealth_Coverage(t *testing.T) {
	t.Run("no address and no checker", func(t *testing.T) {
		pm := pool.NewManager()
		u := NewUpstream(pm)
		// Cast to *Upstream to access fields since they are private but we are in same package
		uu, ok := u.(*Upstream)
		assert.True(t, ok)

		err := uu.CheckHealth(context.Background())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no address configured")
	})

    t.Run("checker returns error", func(t *testing.T) {
		pm := pool.NewManager()
		u := NewUpstream(pm)
		uu, ok := u.(*Upstream)
        assert.True(t, ok)

        // Create a real checker that always fails
        checker := health.NewChecker(
            health.WithCheck(health.Check{
                Name: "fail",
                Check: func(ctx context.Context) error {
                    return errors.New("forced failure")
                },
            }),
             health.WithTimeout(10*time.Millisecond),
        )
        uu.checker = checker

		err := uu.CheckHealth(context.Background())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "health check failed")
	})

    t.Run("checker returns success", func(t *testing.T) {
		pm := pool.NewManager()
		u := NewUpstream(pm)
		uu, ok := u.(*Upstream)
        assert.True(t, ok)

        checker := health.NewChecker(
            health.WithCheck(health.Check{
                Name: "success",
                Check: func(ctx context.Context) error {
                    return nil
                },
            }),
        )
        uu.checker = checker

		err := uu.CheckHealth(context.Background())
		assert.NoError(t, err)
	})
}

func TestHTTPMethodToString_Coverage(t *testing.T) {
    // Test invalid method
    _, err := httpMethodToString(configv1.HttpCallDefinition_HttpMethod(999))
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "unsupported HTTP method")
}

func TestHTTPUpstream_Register_Coverage(t *testing.T) {
    pm := pool.NewManager()
    u := NewUpstream(pm)
    tm := tool.NewManager(nil)

    t.Run("nil http service", func(t *testing.T) {
        config := configv1.UpstreamServiceConfig_builder{
            Name: proto.String("test"),
            // ServiceConfig is nil, GetHttpService() will return nil
        }.Build()
        _, _, _, err := u.Register(context.Background(), config, tm, nil, nil, false)
        assert.Error(t, err)
        assert.Contains(t, err.Error(), "http service config is nil")
    })

    t.Run("empty address", func(t *testing.T) {
        config := configv1.UpstreamServiceConfig_builder{
            Name: proto.String("test"),
            HttpService: configv1.HttpUpstreamService_builder{
                Address: proto.String(""),
            }.Build(),
        }.Build()
        _, _, _, err := u.Register(context.Background(), config, tm, nil, nil, false)
        assert.Error(t, err)
        assert.Contains(t, err.Error(), "http service address is required")
    })

    t.Run("invalid address scheme", func(t *testing.T) {
        config := configv1.UpstreamServiceConfig_builder{
            Name: proto.String("test"),
            HttpService: configv1.HttpUpstreamService_builder{
                Address: proto.String("ftp://example.com"),
            }.Build(),
        }.Build()
        _, _, _, err := u.Register(context.Background(), config, tm, nil, nil, false)
        assert.Error(t, err)
        assert.Contains(t, err.Error(), "invalid http service address scheme")
    })

    t.Run("invalid service name", func(t *testing.T) {
        config := configv1.UpstreamServiceConfig_builder{
            Name: proto.String(""),
            HttpService: configv1.HttpUpstreamService_builder{
                Address: proto.String("http://example.com"),
            }.Build(),
        }.Build()
        _, _, _, err := u.Register(context.Background(), config, tm, nil, nil, false)
        assert.Error(t, err)
    })
}

func TestHTTPUpstream_CreateTools_Coverage(t *testing.T) {
    pm := pool.NewManager()
    u := NewUpstream(pm)
    tm := tool.NewManager(nil)

    t.Run("invalid method", func(t *testing.T) {
        invalidMethod := configv1.HttpCallDefinition_HttpMethod(999)
        config := configv1.UpstreamServiceConfig_builder{
            Name: proto.String("test-invalid-method"),
            HttpService: configv1.HttpUpstreamService_builder{
                Address: proto.String("http://example.com"),
                Calls: map[string]*configv1.HttpCallDefinition{
                    "call1": configv1.HttpCallDefinition_builder{
                        Method:       &invalidMethod,
                        EndpointPath: proto.String("/foo"),
                    }.Build(),
                },
                Tools: []*configv1.ToolDefinition{
                    configv1.ToolDefinition_builder{Name: proto.String("tool1"), CallId: proto.String("call1")}.Build(),
                },
            }.Build(),
        }.Build()
        _, tools, _, err := u.Register(context.Background(), config, tm, nil, nil, false)
        assert.NoError(t, err)
        assert.Len(t, tools, 0) // Should skip the tool
    })

    t.Run("invalid url", func(t *testing.T) {
        validMethod := configv1.HttpCallDefinition_HTTP_METHOD_GET
        config := configv1.UpstreamServiceConfig_builder{
            Name: proto.String("test-invalid-url"),
            HttpService: configv1.HttpUpstreamService_builder{
                Address: proto.String("http://example.com"),
                Calls: map[string]*configv1.HttpCallDefinition{
                    "call1": configv1.HttpCallDefinition_builder{
                        Method:       validMethod.Enum(),
                        EndpointPath: proto.String(":/foo\nbar"), // Invalid URL char
                    }.Build(),
                },
                Tools: []*configv1.ToolDefinition{
                    configv1.ToolDefinition_builder{Name: proto.String("tool1"), CallId: proto.String("call1")}.Build(),
                },
            }.Build(),
        }.Build()
        _, tools, _, err := u.Register(context.Background(), config, tm, nil, nil, false)
        assert.NoError(t, err)
        assert.Len(t, tools, 0)
    })

    t.Run("input schema merging", func(t *testing.T) {
        validMethod := configv1.HttpCallDefinition_HTTP_METHOD_GET
        inputSchema, _ := structpb.NewStruct(map[string]interface{}{
            "type": "object",
            "properties": map[string]interface{}{
                "existing": map[string]interface{}{"type": "string"},
            },
            "required": []interface{}{"existing"},
        })

        config := configv1.UpstreamServiceConfig_builder{
            Name: proto.String("test-schema-merge"),
            HttpService: configv1.HttpUpstreamService_builder{
                Address: proto.String("http://example.com"),
                Calls: map[string]*configv1.HttpCallDefinition{
                    "call1": configv1.HttpCallDefinition_builder{
                        Method:       validMethod.Enum(),
                        EndpointPath: proto.String("/foo"),
                        InputSchema:  inputSchema,
                        Parameters: []*configv1.HttpParameterMapping{
                            configv1.HttpParameterMapping_builder{
                                Schema: configv1.ParameterSchema_builder{
                                    Name:       proto.String("new_param"),
                                    IsRequired: proto.Bool(true),
                                }.Build(),
                            }.Build(),
                            configv1.HttpParameterMapping_builder{
                                Schema: configv1.ParameterSchema_builder{
                                    Name:       proto.String("existing"), // Should merge/overwrite
                                    IsRequired: proto.Bool(true),
                                }.Build(),
                            }.Build(),
                        },
                    }.Build(),
                },
                Tools: []*configv1.ToolDefinition{
                    configv1.ToolDefinition_builder{Name: proto.String("tool1"), CallId: proto.String("call1")}.Build(),
                },
            }.Build(),
        }.Build()
        _, tools, _, err := u.Register(context.Background(), config, tm, nil, nil, false)
        assert.NoError(t, err)
        assert.Len(t, tools, 1)

        // Verify tool schema
        registeredTool, ok := tm.GetTool("test-schema-merge.tool1")
        assert.True(t, ok)

        props := registeredTool.Tool().GetAnnotations().GetInputSchema().GetFields()["properties"].GetStructValue().GetFields()
        assert.Contains(t, props, "existing")
        assert.Contains(t, props, "new_param")

        req := registeredTool.Tool().GetAnnotations().GetInputSchema().GetFields()["required"].GetListValue().GetValues()
        // Check required fields logic
        // "existing" was in schema, and in params.
        // "new_param" was in params.
        assert.GreaterOrEqual(t, len(req), 2)
    })

    t.Run("disabled tool", func(t *testing.T) {
        validMethod := configv1.HttpCallDefinition_HTTP_METHOD_GET
        config := configv1.UpstreamServiceConfig_builder{
            Name: proto.String("test-disabled-tool"),
            HttpService: configv1.HttpUpstreamService_builder{
                Address: proto.String("http://example.com"),
                Calls: map[string]*configv1.HttpCallDefinition{
                    "call1": configv1.HttpCallDefinition_builder{
                        Method:       validMethod.Enum(),
                        EndpointPath: proto.String("/foo"),
                    }.Build(),
                },
                Tools: []*configv1.ToolDefinition{
                    configv1.ToolDefinition_builder{Name: proto.String("tool1"), CallId: proto.String("call1"), Disable: proto.Bool(true)}.Build(),
                },
            }.Build(),
        }.Build()
        _, tools, _, err := u.Register(context.Background(), config, tm, nil, nil, false)
        assert.NoError(t, err)
        assert.Len(t, tools, 0)
    })

    t.Run("export policy skip", func(t *testing.T) {
        validMethod := configv1.HttpCallDefinition_HTTP_METHOD_GET
        config := configv1.UpstreamServiceConfig_builder{
            Name: proto.String("test-export-policy"),
            ToolExportPolicy: configv1.ExportPolicy_builder{
                DefaultAction: configv1.ExportPolicy_UNEXPORT.Enum(),
            }.Build(),
            HttpService: configv1.HttpUpstreamService_builder{
                Address: proto.String("http://example.com"),
                Calls: map[string]*configv1.HttpCallDefinition{
                    "call1": configv1.HttpCallDefinition_builder{
                        Method:       validMethod.Enum(),
                        EndpointPath: proto.String("/foo"),
                    }.Build(),
                },
                Tools: []*configv1.ToolDefinition{
                    configv1.ToolDefinition_builder{Name: proto.String("tool1"), CallId: proto.String("call1")}.Build(),
                },
            }.Build(),
        }.Build()
        _, tools, _, err := u.Register(context.Background(), config, tm, nil, nil, false)
        assert.NoError(t, err)
        assert.Len(t, tools, 0)
    })
}
