package app

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/storage/memory"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func TestAPI_Services(t *testing.T) {
	memStore := memory.NewStore()
	app := NewApplication()
	handler := app.createAPIHandler(memStore)
	server := httptest.NewServer(handler)
	defer server.Close()

	t.Run("CreateService", func(t *testing.T) {
		svc := &configv1.UpstreamServiceConfig{
			Name: proto.String("test-service"),
			ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
				HttpService: &configv1.HttpUpstreamService{
					Address: proto.String("http://localhost:8080"),
				},
			},
		}
		// Use protojson to marshal
		body, _ := protojson.Marshal(svc)
		resp, err := http.Post(server.URL+"/services", "application/json", bytes.NewReader(body))
		require.NoError(t, err)
		if resp.StatusCode != http.StatusCreated {
			t.Logf("Response status: %d", resp.StatusCode)
			var buf bytes.Buffer
			_, _ = buf.ReadFrom(resp.Body)
			t.Logf("Response body: %s", buf.String())
		}
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		// Verify it was saved
		saved, err := memStore.GetService(context.Background(), "test-service")
		require.NoError(t, err)
		if saved != nil {
			assert.Equal(t, "test-service", saved.GetName())
		}
	})

	t.Run("ListServices", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/services")
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var services []map[string]any
		err = json.NewDecoder(resp.Body).Decode(&services)
		require.NoError(t, err)
		if len(services) > 0 {
			assert.Equal(t, "test-service", services[0]["name"])
		}
	})

	t.Run("GetService", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/services/test-service")
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("UpdateService", func(t *testing.T) {
		// Ensure service exists first
		svc := &configv1.UpstreamServiceConfig{
			Name: proto.String("test-service"),
			ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
				HttpService: &configv1.HttpUpstreamService{
					Address: proto.String("http://localhost:8080"),
				},
			},
		}
		_ = memStore.SaveService(context.Background(), svc)

		svcUpdate := &configv1.UpstreamServiceConfig{
			Name: proto.String("test-service"),
			ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
				HttpService: &configv1.HttpUpstreamService{
					Address: proto.String("http://localhost:8081"),
				},
			},
		}
		body, _ := protojson.Marshal(svcUpdate)
		req, _ := http.NewRequest(http.MethodPut, server.URL+"/services/test-service", bytes.NewReader(body))
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		saved, _ := memStore.GetService(context.Background(), "test-service")
		assert.Equal(t, "http://localhost:8081", saved.GetHttpService().GetAddress())
	})

	t.Run("ServiceStatus", func(t *testing.T) {
		_ = memStore.SaveService(context.Background(), &configv1.UpstreamServiceConfig{
			Name: proto.String("test-service"),
			ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
				HttpService: &configv1.HttpUpstreamService{
					Address: proto.String("http://localhost:8080"),
				},
			},
		})

		resp, err := http.Get(server.URL + "/services/test-service/status")
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var status map[string]any
		err = json.NewDecoder(resp.Body).Decode(&status)
		require.NoError(t, err)
		assert.Equal(t, "test-service", status["name"])
		assert.Equal(t, "Inactive", status["status"])
	})

	t.Run("DeleteService", func(t *testing.T) {
		_ = memStore.SaveService(context.Background(), &configv1.UpstreamServiceConfig{
			Name: proto.String("test-service"),
			ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
				HttpService: &configv1.HttpUpstreamService{
					Address: proto.String("http://localhost:8080"),
				},
			},
		})

		req, _ := http.NewRequest(http.MethodDelete, server.URL+"/services/test-service", nil)
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNoContent, resp.StatusCode)

		saved, _ := memStore.GetService(context.Background(), "test-service")
		assert.Nil(t, saved)
	})
}

func TestAPI_Settings(t *testing.T) {
	memStore := memory.NewStore()
	app := NewApplication()
	handler := app.createAPIHandler(memStore)
	server := httptest.NewServer(handler)
	defer server.Close()

	t.Run("GetSettings_Empty", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/settings")
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("SaveSettings", func(t *testing.T) {
		settings := &configv1.GlobalSettings{
			AllowedIps: []string{"127.0.0.1"},
		}
		body, _ := protojson.Marshal(settings)
		resp, err := http.Post(server.URL+"/settings", "application/json", bytes.NewReader(body))
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		saved, err := memStore.GetGlobalSettings(context.Background())
		require.NoError(t, err)
		assert.Contains(t, saved.AllowedIps, "127.0.0.1")
	})
}

func TestAPI_Tools(t *testing.T) {
	app := NewApplication()
	// Mock a tool
	app.ToolManager.AddTool(tool.NewSimpleTool(&v1.Tool{
		Name:        proto.String("mock_tool"),
		Description: proto.String("A mock tool"),
		ServiceId:   proto.String("mock-service"),
	}, nil))

	handler := app.createAPIHandler(nil) // Store not needed for tools
	server := httptest.NewServer(handler)
	defer server.Close()

	resp, err := http.Get(server.URL + "/tools")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var tools []map[string]any
	err = json.NewDecoder(resp.Body).Decode(&tools)
	require.NoError(t, err)
	assert.NotEmpty(t, tools)
	found := false
	for _, t := range tools {
		if t["name"] == "mock_tool" {
			found = true
			break
		}
	}
	assert.True(t, found)
}

func TestAPI_Execute(t *testing.T) {
	app := NewApplication()
	// Register a mock tool that echoes
	app.ToolManager.AddTool(tool.NewSimpleTool(&v1.Tool{
		Name:      proto.String("echo"),
		ServiceId: proto.String("mock-service"),
	}, func(ctx context.Context, args json.RawMessage) (*tool.Result, error) {
		return &tool.Result{Content: []any{map[string]any{"text": string(args)}}}, nil
	}))

	handler := app.createAPIHandler(nil)
	server := httptest.NewServer(handler)
	defer server.Close()

	reqBody := map[string]any{
		"name": "echo",
		"arguments": map[string]string{"msg": "hello"},
	}
	body, _ := json.Marshal(reqBody)
	resp, err := http.Post(server.URL+"/execute", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestAPI_Profiles(t *testing.T) {
	memStore := memory.NewStore()
	app := NewApplication()
	handler := app.createAPIHandler(memStore)
	server := httptest.NewServer(handler)
	defer server.Close()

	t.Run("CreateProfile", func(t *testing.T) {
		profile := &configv1.ProfileDefinition{
			Name: proto.String("dev"),
			RequiredRoles: []string{"developer"},
		}
		body, _ := protojson.Marshal(profile)
		resp, err := http.Post(server.URL+"/profiles", "application/json", bytes.NewReader(body))
		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	})

	t.Run("ListProfiles", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/profiles")
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("GetProfileDetail", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/profiles/dev")
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("ExportProfile", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/profiles/dev/export")
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Contains(t, resp.Header.Get("Content-Disposition"), "dev.json")
	})

	t.Run("UpdateProfile", func(t *testing.T) {
		profile := &configv1.ProfileDefinition{
			Name: proto.String("dev"),
			RequiredRoles: []string{"admin"},
		}
		body, _ := protojson.Marshal(profile)
		req, _ := http.NewRequest(http.MethodPut, server.URL+"/profiles/dev", bytes.NewReader(body))
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("DeleteProfile", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodDelete, server.URL+"/profiles/dev", nil)
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	})
}

func TestAPI_Secrets(t *testing.T) {
	memStore := memory.NewStore()
	app := NewApplication()
	handler := app.createAPIHandler(memStore)
	server := httptest.NewServer(handler)
	defer server.Close()

	t.Run("CreateSecret", func(t *testing.T) {
		secret := &configv1.Secret{
			Id: proto.String("secret-1"),
			Name: proto.String("api-key"),
			Value: proto.String("hidden"),
		}
		body, _ := protojson.Marshal(secret)
		resp, err := http.Post(server.URL+"/secrets", "application/json", bytes.NewReader(body))
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("ListSecrets", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/secrets")
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var secrets []map[string]any
		err = json.NewDecoder(resp.Body).Decode(&secrets)
		require.NoError(t, err)
		assert.Equal(t, "[REDACTED]", secrets[0]["value"])
	})

	t.Run("GetSecretDetail", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/secrets/secret-1")
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var secret map[string]any
		err = json.NewDecoder(resp.Body).Decode(&secret)
		require.NoError(t, err)
		assert.Equal(t, "[REDACTED]", secret["value"])
	})

	t.Run("DeleteSecret", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodDelete, server.URL+"/secrets/secret-1", nil)
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	})
}

func TestAPI_Collections(t *testing.T) {
	memStore := memory.NewStore()
	app := NewApplication()
	handler := app.createAPIHandler(memStore)
	server := httptest.NewServer(handler)
	defer server.Close()

	t.Run("CreateCollection", func(t *testing.T) {
		col := &configv1.UpstreamServiceCollectionShare{
			Name: proto.String("test-col"),
			Services: []*configv1.UpstreamServiceConfig{
				{Name: proto.String("svc1")},
			},
		}
		body, _ := protojson.Marshal(col)
		resp, err := http.Post(server.URL+"/collections", "application/json", bytes.NewReader(body))
		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	})

	t.Run("ListCollections", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/collections")
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("GetCollectionDetail", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/collections/test-col")
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("ExportCollection", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/collections/test-col/export")
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("UpdateCollection", func(t *testing.T) {
		col := &configv1.UpstreamServiceCollectionShare{
			Name: proto.String("test-col"),
			Description: proto.String("Updated desc"),
		}
		body, _ := protojson.Marshal(col)
		req, _ := http.NewRequest(http.MethodPut, server.URL+"/collections/test-col", bytes.NewReader(body))
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("ApplyCollection", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodPost, server.URL+"/collections/test-col/apply", nil)
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Verify svc1 was created
		svc, err := memStore.GetService(context.Background(), "svc1")
		require.NoError(t, err)
		assert.NotNil(t, svc)
	})

	t.Run("DeleteCollection", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodDelete, server.URL+"/collections/test-col", nil)
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	})
}
