// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/auth"
	"github.com/mcpany/core/server/pkg/bus"
	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/mcpserver"
	"github.com/mcpany/core/server/pkg/middleware"
	"github.com/mcpany/core/server/pkg/pool"
	"github.com/mcpany/core/server/pkg/profile"
	"github.com/mcpany/core/server/pkg/prompt"
	"github.com/mcpany/core/server/pkg/resource"
	"github.com/mcpany/core/server/pkg/serviceregistry"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/mcpany/core/server/pkg/upstream/factory"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func httpMethodPtr(m configv1.HttpCallDefinition_HttpMethod) *configv1.HttpCallDefinition_HttpMethod {
	return &m
}

func TestServeUI(t *testing.T) {
	fs := afero.NewMemMapFs()
	// Create UI directory structure in "out" (Next.js export)
	require.NoError(t, fs.MkdirAll("./ui/out", 0755))
	require.NoError(t, afero.WriteFile(fs, "./ui/out/index.html", []byte("<html>Index</html>"), 0644))
	require.NoError(t, fs.MkdirAll("./ui/out/_next/static", 0755))
	require.NoError(t, afero.WriteFile(fs, "./ui/out/_next/static/script.js", []byte("console.log('hi')"), 0644))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	addr := l.Addr().String()
	_ = l.Close()

	app := NewApplication()
	app.fs = fs
	app.SettingsManager = NewGlobalSettingsManager("", nil, nil)

	busProvider, _ := bus.NewProvider(nil)
	poolManager := pool.NewManager()
	upstreamFactory := factory.NewUpstreamServiceFactory(poolManager, nil)
	toolManager := tool.NewManager(busProvider)
	authManager := auth.NewManager()
	serviceRegistry := serviceregistry.New(upstreamFactory, toolManager, prompt.NewManager(), resource.NewManager(), authManager)
	mcpSrv, err := mcpserver.NewServer(ctx, toolManager, prompt.NewManager(), resource.NewManager(), authManager, serviceRegistry, busProvider, false)
	require.NoError(t, err)

	errChan := make(chan error, 1)
	go func() {
		errChan <- app.runServerMode(ctx, mcpSrv, busProvider, addr, "", 5*time.Second, nil, middleware.NewCachingMiddleware(toolManager), nil, nil, serviceRegistry, nil, "", "", "")
	}()

	waitForServerReady(t, addr)
	baseURL := "http://" + addr

	t.Run("Serve Index", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/ui/")
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		body, _ := io.ReadAll(resp.Body)
		assert.Equal(t, "<html>Index</html>", string(body))
		// Should have no-cache for HTML
		assert.Contains(t, resp.Header.Get("Cache-Control"), "no-cache")
	})

	t.Run("Serve Static Asset", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/ui/_next/static/script.js")
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		// Should have immutable cache
		assert.Contains(t, resp.Header.Get("Cache-Control"), "immutable")
	})

	cancel()
	<-errChan
}

func TestStatelessJSONRPC_ToolsList(t *testing.T) {
	fs := afero.NewMemMapFs()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	addr := l.Addr().String()
	_ = l.Close()

	// 1. Setup Managers and Dependencies manually
	busProvider, _ := bus.NewProvider(nil)
	poolManager := pool.NewManager()
	upstreamFactory := factory.NewUpstreamServiceFactory(poolManager, nil)
	toolManager := tool.NewManager(busProvider)
	promptManager := prompt.NewManager()
	resourceManager := resource.NewManager()
	authManager := auth.NewManager()
	serviceRegistry := serviceregistry.New(upstreamFactory, toolManager, promptManager, resourceManager, authManager)

	// 2. Setup Config (Users & Profiles)
	// We need to inject these into Managers.
	app := NewApplication()
	app.fs = fs
	app.SettingsManager = NewGlobalSettingsManager("", nil, nil)
	app.AuthManager = authManager
	app.ToolManager = toolManager

	profileDefs := []*configv1.ProfileDefinition{
		configv1.ProfileDefinition_builder{
			Name: proto.String("dev-profile"),
			ServiceConfig: map[string]*configv1.ProfileServiceConfig{
				"service-a": configv1.ProfileServiceConfig_builder{Enabled: proto.Bool(true)}.Build(),
				"service-b": configv1.ProfileServiceConfig_builder{Enabled: proto.Bool(false)}.Build(),
			},
		}.Build(),
	}
	app.ProfileManager = profile.NewManager(profileDefs)
	toolManager.SetProfiles([]string{}, profileDefs)

	// Users
	authManager.SetUsers([]*configv1.User{
		configv1.User_builder{
			Id:         proto.String("test-user"),
			ProfileIds: []string{"dev-profile"},
		}.Build(),
	})

	// 3. Register Services
	svc1 := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("service-a"),
		Id:   proto.String("service-a"),
		HttpService: configv1.HttpUpstreamService_builder{
			Address: proto.String("http://127.0.0.1:9090"),
			Tools: []*configv1.ToolDefinition{
				configv1.ToolDefinition_builder{Name: proto.String("tool-a"), CallId: proto.String("call-a")}.Build(),
			},
			Calls: map[string]*configv1.HttpCallDefinition{
				"call-a": configv1.HttpCallDefinition_builder{
					Id:           proto.String("call-a"),
					Method:       httpMethodPtr(configv1.HttpCallDefinition_HTTP_METHOD_GET),
					EndpointPath: proto.String("/"),
				}.Build(),
			},
		}.Build(),
	}.Build()

	svc2 := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("service-b"),
		Id:   proto.String("service-b"),
		HttpService: configv1.HttpUpstreamService_builder{
			Address: proto.String("http://127.0.0.1:9091"),
			Tools: []*configv1.ToolDefinition{
				configv1.ToolDefinition_builder{Name: proto.String("tool-b"), CallId: proto.String("call-b")}.Build(),
			},
			Calls: map[string]*configv1.HttpCallDefinition{
				"call-b": configv1.HttpCallDefinition_builder{
					Id:           proto.String("call-b"),
					Method:       httpMethodPtr(configv1.HttpCallDefinition_HTTP_METHOD_GET),
					EndpointPath: proto.String("/"),
				}.Build(),
			},
		}.Build(),
	}.Build()

	// Manually register to populate ToolManager
	serviceRegistry.RegisterService(ctx, svc1)
	serviceRegistry.RegisterService(ctx, svc2)

	mcpSrv, err := mcpserver.NewServer(ctx, toolManager, promptManager, resourceManager, authManager, serviceRegistry, busProvider, false)
	require.NoError(t, err)

	errChan := make(chan error, 1)
	go func() {
		// Use manual runServerMode to skip DB init
		errChan <- app.runServerMode(ctx, mcpSrv, busProvider, addr, "", 5*time.Second, nil, middleware.NewCachingMiddleware(toolManager), nil, nil, serviceRegistry, nil, "", "", "")
	}()

	waitForServerReady(t, addr)
	baseURL := "http://" + addr

	t.Run("Tools List Filtered", func(t *testing.T) {
		reqBody := `{"jsonrpc":"2.0","method":"tools/list","id":1}`
		resp, err := http.Post(baseURL+"/mcp/u/test-user/profile/dev-profile", "application/json", strings.NewReader(reqBody))
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]any
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))

		// Check response structure
		resMap, ok := result["result"].(map[string]any)
		require.True(t, ok)
		tools, ok := resMap["tools"].([]any)
		require.True(t, ok)

		// Expected: tool-a should be present, tool-b should be absent
		foundA := false
		foundB := false
		for _, tRaw := range tools {
			t := tRaw.(map[string]any)
			if t["name"] == "tool-a" {
				foundA = true
			}
			if t["name"] == "tool-b" {
				foundB = true
			}
		}

		assert.True(t, foundA, "tool-a should be present")
		assert.False(t, foundB, "tool-b should be filtered out by profile")
	})

	cancel()
	<-errChan
}

func TestMultiUserHandler_RBAC_Enforcement(t *testing.T) {
	fs := afero.NewMemMapFs()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	addr := l.Addr().String()
	_ = l.Close()

	// 1. Setup Dependencies
	busProvider, _ := bus.NewProvider(nil)
	poolManager := pool.NewManager()
	upstreamFactory := factory.NewUpstreamServiceFactory(poolManager, nil)
	toolManager := tool.NewManager(busProvider)
	authManager := auth.NewManager()
	serviceRegistry := serviceregistry.New(upstreamFactory, toolManager, prompt.NewManager(), resource.NewManager(), authManager)

	// 2. Setup App with RBAC Config
	app := NewApplication()
	app.fs = fs
	app.SettingsManager = NewGlobalSettingsManager("", nil, nil)
	app.AuthManager = authManager
	app.ProfileManager = profile.NewManager([]*configv1.ProfileDefinition{
		configv1.ProfileDefinition_builder{
			Name:          proto.String("admin-profile"),
			RequiredRoles: []string{"admin"},
		}.Build(),
	})

	// Users
	authManager.SetUsers([]*configv1.User{
		configv1.User_builder{
			Id:         proto.String("dev-user"),
			ProfileIds: []string{"admin-profile"},
			Roles:      []string{"dev"},
		}.Build(),
		configv1.User_builder{
			Id:         proto.String("admin-user"),
			ProfileIds: []string{"admin-profile"},
			Roles:      []string{"admin"},
		}.Build(),
	})

	mcpSrv, err := mcpserver.NewServer(ctx, toolManager, prompt.NewManager(), resource.NewManager(), authManager, serviceRegistry, busProvider, false)
	require.NoError(t, err)

	errChan := make(chan error, 1)
	go func() {
		errChan <- app.runServerMode(ctx, mcpSrv, busProvider, addr, "", 5*time.Second, nil, middleware.NewCachingMiddleware(toolManager), nil, nil, serviceRegistry, nil, "", "", "")
	}()

	waitForServerReady(t, addr)
	baseURL := "http://" + addr

	t.Run("Access Denied - Role Mismatch", func(t *testing.T) {
		reqBody := `{"jsonrpc":"2.0","method":"tools/list","id":1}`
		resp, err := http.Post(baseURL+"/mcp/u/dev-user/profile/admin-profile", "application/json", strings.NewReader(reqBody))
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})

	t.Run("Access Allowed - Role Match", func(t *testing.T) {
		reqBody := `{"jsonrpc":"2.0","method":"tools/list","id":1}`
		resp, err := http.Post(baseURL+"/mcp/u/admin-user/profile/admin-profile", "application/json", strings.NewReader(reqBody))
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	cancel()
	<-errChan
}

func TestServeUI_BlockSourceCode(t *testing.T) {
	fs := afero.NewMemMapFs()
	// Create "ui" directory WITH package.json (simulating source code root)
	require.NoError(t, fs.MkdirAll("./ui", 0755))
	require.NoError(t, afero.WriteFile(fs, "./ui/package.json", []byte("{}"), 0644))
	require.NoError(t, afero.WriteFile(fs, "./ui/index.html", []byte("<html>Source</html>"), 0644))

	// Ensure logging captures output
	logging.ForTestsOnlyResetLogger()
	var buf ThreadSafeBuffer
	logging.Init(slog.LevelInfo, &buf)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	addr := l.Addr().String()
	_ = l.Close()

	app := NewApplication()
	app.fs = fs
	app.SettingsManager = NewGlobalSettingsManager("", nil, nil)

	// Minimal dependencies
	busProvider, _ := bus.NewProvider(nil)
	poolManager := pool.NewManager()
	upstreamFactory := factory.NewUpstreamServiceFactory(poolManager, nil)
	toolManager := tool.NewManager(busProvider)
	authManager := auth.NewManager()
	serviceRegistry := serviceregistry.New(upstreamFactory, toolManager, prompt.NewManager(), resource.NewManager(), authManager)
	mcpSrv, _ := mcpserver.NewServer(ctx, toolManager, prompt.NewManager(), resource.NewManager(), authManager, serviceRegistry, busProvider, false)

	errChan := make(chan error, 1)
	go func() {
		errChan <- app.runServerMode(ctx, mcpSrv, busProvider, addr, "", 5*time.Second, nil, middleware.NewCachingMiddleware(toolManager), nil, nil, serviceRegistry, nil, "", "", "")
	}()

	waitForServerReady(t, addr)
	baseURL := "http://" + addr

	// Requesting UI should fail (404 or 400 or not 200) because serving was disabled due to package.json
	resp, err := http.Get(baseURL + "/ui/")
	require.NoError(t, err)
	defer resp.Body.Close()
	// Should NOT be 200
	assert.NotEqual(t, http.StatusOK, resp.StatusCode)

	// Check Logs
	assert.Contains(t, buf.String(), "UI directory ./ui contains package.json. Refusing to serve source code for security.")

	cancel()
	<-errChan
}

func TestServeUI_DistFallback(t *testing.T) {
	fs := afero.NewMemMapFs()
	// No ./ui/out, but ./ui/dist exists
	require.NoError(t, fs.MkdirAll("./ui/dist", 0755))
	require.NoError(t, afero.WriteFile(fs, "./ui/dist/index.html", []byte("<html>Dist</html>"), 0644))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	addr := l.Addr().String()
	_ = l.Close()

	app := NewApplication()
	app.fs = fs
	app.SettingsManager = NewGlobalSettingsManager("", nil, nil)

	busProvider, _ := bus.NewProvider(nil)
	poolManager := pool.NewManager()
	upstreamFactory := factory.NewUpstreamServiceFactory(poolManager, nil)
	toolManager := tool.NewManager(busProvider)
	authManager := auth.NewManager()
	serviceRegistry := serviceregistry.New(upstreamFactory, toolManager, prompt.NewManager(), resource.NewManager(), authManager)
	mcpSrv, _ := mcpserver.NewServer(ctx, toolManager, prompt.NewManager(), resource.NewManager(), authManager, serviceRegistry, busProvider, false)

	errChan := make(chan error, 1)
	go func() {
		errChan <- app.runServerMode(ctx, mcpSrv, busProvider, addr, "", 5*time.Second, nil, middleware.NewCachingMiddleware(toolManager), nil, nil, serviceRegistry, nil, "", "", "")
	}()

	waitForServerReady(t, addr)
	baseURL := "http://" + addr

	resp, err := http.Get(baseURL + "/ui/")
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	body, _ := io.ReadAll(resp.Body)
	assert.Equal(t, "<html>Dist</html>", string(body))

	cancel()
	<-errChan
}

func TestTLSConfiguration(t *testing.T) {
	t.Skip("Skipping TLS test to avoid complexity of cert generation in test")
}
