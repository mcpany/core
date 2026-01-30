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
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestSecretsCoverage(t *testing.T) {
	// Test StripSecretsFromService for GRPC
	grpcService := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("grpc-service"),
		GrpcService: configv1.GrpcUpstreamService_builder{
			TlsConfig: configv1.TLSConfig_builder{
				ClientKeyPath: proto.String("secret-key"),
			}.Build(),
		}.Build(),
	}.Build()
	config.StripSecretsFromService(grpcService)
	assert.NotNil(t, grpcService)

	// Test StripSecretsFromService for OpenAPI
	openapiService := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("openapi-service"),
		OpenapiService: configv1.OpenapiUpstreamService_builder{
			TlsConfig: configv1.TLSConfig_builder{
				ClientKeyPath: proto.String("secret-key"),
			}.Build(),
		}.Build(),
	}.Build()
	config.StripSecretsFromService(openapiService)
	assert.NotNil(t, openapiService)

	// Test StripSecretsFromMcpCall (coverage)
	mcpService := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("mcp-service"),
		McpService: configv1.McpUpstreamService_builder{
			Calls: map[string]*configv1.MCPCallDefinition{
				"call1": configv1.MCPCallDefinition_builder{}.Build(),
			},
		}.Build(),
	}.Build()
	config.StripSecretsFromService(mcpService)
	assert.NotNil(t, mcpService)
}

func TestSettingsCoverage(t *testing.T) {
	s := config.GlobalSettings()

	// Test SetMiddlewares / Middlewares
	mws := []*configv1.Middleware{
        configv1.Middleware_builder{Name: proto.String("test")}.Build(),
    }
	s.SetMiddlewares(mws)
	assert.Equal(t, mws, s.Middlewares())

	// Test SetDlp / GetDlp
	dlp := configv1.DLPConfig_builder{Enabled: proto.Bool(true)}.Build()
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
	cfg := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("valid-service"),
		HttpService: configv1.HttpUpstreamService_builder{
			Address: proto.String("http://example.com"),
		}.Build(),
	}.Build()
	err := config.ValidateOrError(context.Background(), cfg)
    assert.NoError(t, err)

    // Force an invalid config
    invalidCfg := configv1.UpstreamServiceConfig_builder{
        Name: proto.String(""),
    }.Build()
    // ValidateUpstreamService checks for ServiceConfig type
    err = config.ValidateOrError(context.Background(), invalidCfg)
    assert.Error(t, err)
}

func TestValidatorCoverageMore(t *testing.T) {
    ctx := context.Background()

    // Test Validate (Global Config) with errors

    // 1. Client binary type with short API Key
    cfgClient := configv1.McpAnyServerConfig_builder{
        GlobalSettings: configv1.GlobalSettings_builder{
            ApiKey: proto.String("short"),
        }.Build(),
    }.Build()
    errs := config.Validate(ctx, cfgClient, config.Client)
    assert.NotEmpty(t, errs)
    assert.Contains(t, errs[0].Err.Error(), "at least 16 characters")

    // 4. HTTP Service Invalid Scheme
    cfgHttp := configv1.McpAnyServerConfig_builder{
        UpstreamServices: []*configv1.UpstreamServiceConfig{
            configv1.UpstreamServiceConfig_builder{
                Name: proto.String("bad-http"),
                HttpService: configv1.HttpUpstreamService_builder{
                    Address: proto.String("ftp://example.com"),
                }.Build(),
            }.Build(),
        },
    }.Build()
    errs = config.Validate(ctx, cfgHttp, config.Server)
    assert.NotEmpty(t, errs)
    assert.Contains(t, errs[0].Err.Error(), "invalid http address scheme")

    // 5. Websocket Service Invalid Scheme
    cfgWs := configv1.McpAnyServerConfig_builder{
        UpstreamServices: []*configv1.UpstreamServiceConfig{
            configv1.UpstreamServiceConfig_builder{
                Name: proto.String("bad-ws"),
                WebsocketService: configv1.WebsocketUpstreamService_builder{
                    Address: proto.String("http://example.com"),
                }.Build(),
            }.Build(),
        },
    }.Build()
    errs = config.Validate(ctx, cfgWs, config.Server)
    assert.NotEmpty(t, errs)
    assert.Contains(t, errs[0].Err.Error(), "invalid websocket address scheme")

    // 6. GC Settings Invalid Duration
    cfgGC := configv1.McpAnyServerConfig_builder{
        GlobalSettings: configv1.GlobalSettings_builder{
            GcSettings: configv1.GCSettings_builder{
                Interval: proto.String("invalid"),
            }.Build(),
        }.Build(),
    }.Build()
    errs = config.Validate(ctx, cfgGC, config.Server)
    assert.NotEmpty(t, errs)
    assert.Contains(t, errs[0].Err.Error(), "invalid interval")

    // 7. Duplicate Service Name
    cfgDup := configv1.McpAnyServerConfig_builder{
        UpstreamServices: []*configv1.UpstreamServiceConfig{
            configv1.UpstreamServiceConfig_builder{
                Name: proto.String("s1"),
                HttpService: configv1.HttpUpstreamService_builder{
                    Address: proto.String("http://a.com"),
                }.Build(),
            }.Build(),
            configv1.UpstreamServiceConfig_builder{
                Name: proto.String("s1"),
                HttpService: configv1.HttpUpstreamService_builder{
                    Address: proto.String("http://b.com"),
                }.Build(),
            }.Build(),
        },
    }.Build()
    errs = config.Validate(ctx, cfgDup, config.Server)
    assert.NotEmpty(t, errs)
    assert.Contains(t, errs[0].Err.Error(), "duplicate service name")
}

func TestSecretsHydration(t *testing.T) {
    // Test HydrateSecretsInService with various types to hit coverage
    secrets := map[string]*configv1.SecretValue{
        "MY_SECRET": configv1.SecretValue_builder{
			PlainText: proto.String("real_secret"),
		}.Build(),
    }

	// Command Line Service
	cmdService := func() *configv1.UpstreamServiceConfig {
		cmd := configv1.CommandLineUpstreamService_builder{
			Env: map[string]*configv1.SecretValue{
				"API_KEY": configv1.SecretValue_builder{
					EnvironmentVariable: proto.String("MY_SECRET"),
				}.Build(),
			},
			ContainerEnvironment: configv1.ContainerEnvironment_builder{
				Env: map[string]*configv1.SecretValue{
					"CONTAINER_KEY": configv1.SecretValue_builder{
						EnvironmentVariable: proto.String("MY_SECRET"),
					}.Build(),
				},
			}.Build(),
		}.Build()

		return configv1.UpstreamServiceConfig_builder{
			Name:               proto.String("cmd-svc"),
			CommandLineService: cmd,
		}.Build()
	}()
	config.HydrateSecretsInService(cmdService, secrets)

	// Verify hydration
	envVal := cmdService.GetCommandLineService().GetEnv()["API_KEY"].GetPlainText()
	assert.Equal(t, "real_secret", envVal)

	// Mcp Service Stdio
	mcpStdio := func() *configv1.UpstreamServiceConfig {
		conn := configv1.McpStdioConnection_builder{
			Env: map[string]*configv1.SecretValue{
				"KEY": configv1.SecretValue_builder{
					EnvironmentVariable: proto.String("MY_SECRET"),
				}.Build(),
			},
		}.Build()

		mcp := configv1.McpUpstreamService_builder{
			StdioConnection: conn,
		}.Build()

		return configv1.UpstreamServiceConfig_builder{
			Name:       proto.String("mcp-stdio"),
			McpService: mcp,
		}.Build()
	}()
	config.HydrateSecretsInService(mcpStdio, secrets)
	assert.Equal(t, "real_secret", mcpStdio.GetMcpService().GetStdioConnection().GetEnv()["KEY"].GetPlainText())

	// Mcp Service Bundle
	mcpBundle := func() *configv1.UpstreamServiceConfig {
		conn := configv1.McpBundleConnection_builder{
			Env: map[string]*configv1.SecretValue{
				"KEY": configv1.SecretValue_builder{
					EnvironmentVariable: proto.String("MY_SECRET"),
				}.Build(),
			},
		}.Build()

		mcp := configv1.McpUpstreamService_builder{
			BundleConnection: conn,
		}.Build()

		return configv1.UpstreamServiceConfig_builder{
			Name:       proto.String("mcp-bundle"),
			McpService: mcp,
		}.Build()
	}()
	config.HydrateSecretsInService(mcpBundle, secrets)
	assert.Equal(t, "real_secret", mcpBundle.GetMcpService().GetBundleConnection().GetEnv()["KEY"].GetPlainText())

	// HTTP Service
	httpSvc := func() *configv1.UpstreamServiceConfig {
		param := configv1.HttpParameterMapping_builder{
			Schema: configv1.ParameterSchema_builder{
				Name: proto.String("p1"),
			}.Build(),
			Secret: configv1.SecretValue_builder{
				EnvironmentVariable: proto.String("MY_SECRET"),
			}.Build(),
		}.Build()

		callDef := configv1.HttpCallDefinition_builder{
			Parameters: []*configv1.HttpParameterMapping{param},
		}.Build()

		h := configv1.HttpUpstreamService_builder{
			Calls: map[string]*configv1.HttpCallDefinition{
				"call1": callDef,
			},
		}.Build()

		return configv1.UpstreamServiceConfig_builder{
			Name:        proto.String("http-svc"),
			HttpService: h,
		}.Build()
	}()
	config.HydrateSecretsInService(httpSvc, secrets)
    assert.Equal(t, "real_secret", httpSvc.GetHttpService().GetCalls()["call1"].GetParameters()[0].GetSecret().GetPlainText())
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

    args := cfg.GetUpstreamServices()[0].GetMcpService().GetStdioConnection().GetArgs()
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

    assert.True(t, cfg.GetUpstreamServices()[0].GetDisable())
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
	cfg := func() *configv1.UpstreamServiceConfig {
		cmd := configv1.CommandLineUpstreamService_builder{
			Command: proto.String(exePath),
		}.Build()

		return configv1.UpstreamServiceConfig_builder{
			Name:               proto.String("cmd-svc"),
			CommandLineService: cmd,
		}.Build()
	}()
    // We can't call validateCommandExists directly as it is private, but we can call ValidateOrError
    err = config.ValidateOrError(context.Background(), cfg)
    assert.NoError(t, err)

    // Non-existent command
    cfg.GetCommandLineService().SetCommand(filepath.Join(tmpDir, "notfound"))
    err = config.ValidateOrError(context.Background(), cfg)
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "executable not found")

    // Directory as command
    cfg.GetCommandLineService().SetCommand(tmpDir)
    err = config.ValidateOrError(context.Background(), cfg)
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "is a directory")
}
