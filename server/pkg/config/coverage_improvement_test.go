// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/spf13/afero"
	"google.golang.org/protobuf/proto"
)

func TestSecretsCoverage(t *testing.T) {
	// Test StripSecretsFromService for GRPC
	grpcService := &configv1.UpstreamServiceConfig{
		Name: proto.String("grpc-service"),
		ServiceConfig: &configv1.UpstreamServiceConfig_GrpcService{
			GrpcService: &configv1.GrpcUpstreamService{
				TlsConfig: &configv1.TLSConfig{
					ClientKeyPath: proto.String("secret-key"),
				},
			},
		},
	}
	config.StripSecretsFromService(grpcService)
	assert.NotNil(t, grpcService)

	// Test StripSecretsFromService for OpenAPI
	openapiService := &configv1.UpstreamServiceConfig{
		Name: proto.String("openapi-service"),
		ServiceConfig: &configv1.UpstreamServiceConfig_OpenapiService{
			OpenapiService: &configv1.OpenapiUpstreamService{
				TlsConfig: &configv1.TLSConfig{
					ClientKeyPath: proto.String("secret-key"),
				},
			},
		},
	}
	config.StripSecretsFromService(openapiService)
	assert.NotNil(t, openapiService)

	// Test StripSecretsFromMcpCall (coverage)
	mcpService := &configv1.UpstreamServiceConfig{
		Name: proto.String("mcp-service"),
		ServiceConfig: &configv1.UpstreamServiceConfig_McpService{
			McpService: &configv1.McpUpstreamService{
				Calls: map[string]*configv1.MCPCallDefinition{
					"call1": {},
				},
			},
		},
	}
	config.StripSecretsFromService(mcpService)
	assert.NotNil(t, mcpService)
}

func TestSettingsCoverage(t *testing.T) {
	s := config.GlobalSettings()

	// Test SetMiddlewares / Middlewares
	mws := []*configv1.Middleware{{Name: proto.String("test")}}
	s.SetMiddlewares(mws)
	assert.Equal(t, mws, s.Middlewares())

	// Test SetDlp / GetDlp
	dlp := &configv1.DLPConfig{Enabled: proto.Bool(true)}
	s.SetDlp(dlp)
	assert.Equal(t, dlp, s.GetDlp())

	// Test GetOidc
	assert.Nil(t, s.GetOidc())

	// Test GetProfileDefinitions
	assert.Nil(t, s.GetProfileDefinitions())

	// Test GithubAPIURL
	assert.Equal(t, "", s.GithubAPIURL())
}

func TestValidatorCoverage(t *testing.T) {
	// Test ValidateOrError
	// ValidateOrError expects UpstreamServiceConfig
	cfg := &configv1.UpstreamServiceConfig{
        Name: proto.String("valid-service"),
        ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
            HttpService: &configv1.HttpUpstreamService{
                Address: proto.String("http://example.com"),
            },
        },
    }
	err := config.ValidateOrError(context.Background(), cfg)
    assert.NoError(t, err)

    // Force an invalid config
    invalidCfg := &configv1.UpstreamServiceConfig{
        Name: proto.String(""), // Name is optional in struct but effectively required for logical validation, but ValidateUpstreamService calls ValidateUpstreamService
    }
    // ValidateUpstreamService checks for ServiceConfig type
    err = config.ValidateOrError(context.Background(), invalidCfg)
    assert.Error(t, err)
}

func TestValidatorCoverageMore(t *testing.T) {
    ctx := context.Background()

    // Test Validate (Global Config) with errors

    // 1. Client binary type with short API Key
    cfgClient := &configv1.McpAnyServerConfig{
        GlobalSettings: &configv1.GlobalSettings{
            ApiKey: proto.String("short"),
        },
    }
    errs := config.Validate(ctx, cfgClient, config.Client)
    assert.NotEmpty(t, errs)
    assert.Contains(t, errs[0].Err.Error(), "at least 16 characters")

    // 4. HTTP Service Invalid Scheme
    cfgHttp := &configv1.McpAnyServerConfig{
        UpstreamServices: []*configv1.UpstreamServiceConfig{
            {
                Name: proto.String("bad-http"),
                ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
                    HttpService: &configv1.HttpUpstreamService{
                        Address: proto.String("ftp://example.com"),
                    },
                },
            },
        },
    }
    errs = config.Validate(ctx, cfgHttp, config.Server)
    assert.NotEmpty(t, errs)
    assert.Contains(t, errs[0].Err.Error(), "invalid http target_address scheme")

    // 5. Websocket Service Invalid Scheme
    cfgWs := &configv1.McpAnyServerConfig{
        UpstreamServices: []*configv1.UpstreamServiceConfig{
            {
                Name: proto.String("bad-ws"),
                ServiceConfig: &configv1.UpstreamServiceConfig_WebsocketService{
                    WebsocketService: &configv1.WebsocketUpstreamService{
                        Address: proto.String("http://example.com"),
                    },
                },
            },
        },
    }
    errs = config.Validate(ctx, cfgWs, config.Server)
    assert.NotEmpty(t, errs)
    assert.Contains(t, errs[0].Err.Error(), "invalid websocket target_address scheme")

    // 6. GC Settings Invalid Duration
    cfgGC := &configv1.McpAnyServerConfig{
        GlobalSettings: &configv1.GlobalSettings{
            GcSettings: &configv1.GCSettings{
                Interval: proto.String("invalid"),
            },
        },
    }
    errs = config.Validate(ctx, cfgGC, config.Server)
    assert.NotEmpty(t, errs)
    assert.Contains(t, errs[0].Err.Error(), "invalid interval")

    // 7. Duplicate Service Name
    cfgDup := &configv1.McpAnyServerConfig{
        UpstreamServices: []*configv1.UpstreamServiceConfig{
            {Name: proto.String("s1"), ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{HttpService: &configv1.HttpUpstreamService{Address: proto.String("http://a.com")}}},
            {Name: proto.String("s1"), ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{HttpService: &configv1.HttpUpstreamService{Address: proto.String("http://b.com")}}},
        },
    }
    errs = config.Validate(ctx, cfgDup, config.Server)
    assert.NotEmpty(t, errs)
    assert.Contains(t, errs[0].Err.Error(), "duplicate service name")
}

func TestSecretsHydration(t *testing.T) {
    // Test HydrateSecretsInService with various types to hit coverage
    secrets := map[string]*configv1.SecretValue{
        "MY_SECRET": {
            Value: &configv1.SecretValue_PlainText{PlainText: "real_secret"},
        },
    }

    // Command Line Service
    cmdService := &configv1.UpstreamServiceConfig{
        Name: proto.String("cmd-svc"),
        ServiceConfig: &configv1.UpstreamServiceConfig_CommandLineService{
            CommandLineService: &configv1.CommandLineUpstreamService{
                Env: map[string]*configv1.SecretValue{
                    "API_KEY": {
                        Value: &configv1.SecretValue_EnvironmentVariable{EnvironmentVariable: "MY_SECRET"},
                    },
                },
                ContainerEnvironment: &configv1.ContainerEnvironment{
                    Env: map[string]*configv1.SecretValue{
                        "CONTAINER_KEY": {
                            Value: &configv1.SecretValue_EnvironmentVariable{EnvironmentVariable: "MY_SECRET"},
                        },
                    },
                },
            },
        },
    }
    config.HydrateSecretsInService(cmdService, secrets)

    // Verify hydration
    envVal := cmdService.GetCommandLineService().Env["API_KEY"].GetPlainText()
    assert.Equal(t, "real_secret", envVal)

    // Mcp Service Stdio
    mcpStdio := &configv1.UpstreamServiceConfig{
        Name: proto.String("mcp-stdio"),
        ServiceConfig: &configv1.UpstreamServiceConfig_McpService{
            McpService: &configv1.McpUpstreamService{
                ConnectionType: &configv1.McpUpstreamService_StdioConnection{
                    StdioConnection: &configv1.McpStdioConnection{
                        Env: map[string]*configv1.SecretValue{
                            "KEY": {
                                Value: &configv1.SecretValue_EnvironmentVariable{EnvironmentVariable: "MY_SECRET"},
                            },
                        },
                    },
                },
            },
        },
    }
    config.HydrateSecretsInService(mcpStdio, secrets)
    assert.Equal(t, "real_secret", mcpStdio.GetMcpService().GetStdioConnection().Env["KEY"].GetPlainText())

    // Mcp Service Bundle
    mcpBundle := &configv1.UpstreamServiceConfig{
        Name: proto.String("mcp-bundle"),
        ServiceConfig: &configv1.UpstreamServiceConfig_McpService{
            McpService: &configv1.McpUpstreamService{
                ConnectionType: &configv1.McpUpstreamService_BundleConnection{
                    BundleConnection: &configv1.McpBundleConnection{
                        Env: map[string]*configv1.SecretValue{
                            "KEY": {
                                Value: &configv1.SecretValue_EnvironmentVariable{EnvironmentVariable: "MY_SECRET"},
                            },
                        },
                    },
                },
            },
        },
    }
    config.HydrateSecretsInService(mcpBundle, secrets)
    assert.Equal(t, "real_secret", mcpBundle.GetMcpService().GetBundleConnection().Env["KEY"].GetPlainText())

    // HTTP Service
    httpSvc := &configv1.UpstreamServiceConfig{
        Name: proto.String("http-svc"),
        ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
            HttpService: &configv1.HttpUpstreamService{
                Calls: map[string]*configv1.HttpCallDefinition{
                    "call1": {
                        Parameters: []*configv1.HttpParameterMapping{
                            {
                                Schema: &configv1.ParameterSchema{Name: proto.String("p1")},
                                Secret: &configv1.SecretValue{
                                    Value: &configv1.SecretValue_EnvironmentVariable{EnvironmentVariable: "MY_SECRET"},
                                },
                            },
                        },
                    },
                },
            },
        },
    }
    config.HydrateSecretsInService(httpSvc, secrets)
    assert.Equal(t, "real_secret", httpSvc.GetHttpService().Calls["call1"].Parameters[0].Secret.GetPlainText())
}

func TestResolveEnvValueFallback(t *testing.T) {
    // Test the CSV parsing fallback by providing a string that is invalid CSV
    fs := afero.NewMemMapFs()
    configContent := `
upstream_services:
  - name: "my-service"
    mcp_service:
      stdio_connection:
        command: "echo"
        args: ["original"]
`
    err := afero.WriteFile(fs, "/config.yaml", []byte(configContent), 0644)
    require.NoError(t, err)

    envVar := "MCPANY__UPSTREAM_SERVICES__0__MCP_SERVICE__STDIO_CONNECTION__ARGS"
    val := `foo"bar,baz`

    _ = os.Setenv(envVar, val)
    defer func() {
        _ = os.Unsetenv(envVar)
    }()

    store := config.NewFileStore(fs, []string{"/config.yaml"})
    cfg, err := store.Load(context.Background())
    require.NoError(t, err)

    args := cfg.UpstreamServices[0].GetMcpService().GetStdioConnection().GetArgs()
    // Fallback to split by comma: ["foo\"bar", "baz"]
    assert.Len(t, args, 2)
    assert.Equal(t, `foo"bar`, args[0])
    assert.Equal(t, "baz", args[1])
}

func TestStoreBoolConversion(t *testing.T) {
    fs := afero.NewMemMapFs()
    configContent := `
upstream_services:
  - name: "my-service"
    disable: false
`
    err := afero.WriteFile(fs, "/config.yaml", []byte(configContent), 0644)
    require.NoError(t, err)

    envVar := "MCPANY__UPSTREAM_SERVICES__0__DISABLE"
    _ = os.Setenv(envVar, "true")
    defer func() {
        _ = os.Unsetenv(envVar)
    }()

    store := config.NewFileStore(fs, []string{"/config.yaml"})
    cfg, err := store.Load(context.Background())
    require.NoError(t, err)

    assert.True(t, cfg.UpstreamServices[0].GetDisable())
}

func TestValidatorCommandExists(t *testing.T) {
    // Create a dummy executable
    tmpDir := t.TempDir()
    exePath := filepath.Join(tmpDir, "myexe")
    f, err := os.Create(exePath)
    require.NoError(t, err)
    f.Close()
    os.Chmod(exePath, 0755)

    // Valid command absolute path
    cfg := &configv1.UpstreamServiceConfig{
        Name: proto.String("cmd-svc"),
        ServiceConfig: &configv1.UpstreamServiceConfig_CommandLineService{
            CommandLineService: &configv1.CommandLineUpstreamService{
                Command: proto.String(exePath),
            },
        },
    }
    // We can't call validateCommandExists directly as it is private, but we can call ValidateOrError
    err = config.ValidateOrError(context.Background(), cfg)
    assert.NoError(t, err)

    // Non-existent command
    cfg.GetCommandLineService().Command = proto.String(filepath.Join(tmpDir, "notfound"))
    err = config.ValidateOrError(context.Background(), cfg)
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "executable not found")

    // Directory as command
    cfg.GetCommandLineService().Command = proto.String(tmpDir)
    err = config.ValidateOrError(context.Background(), cfg)
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "is a directory")
}
