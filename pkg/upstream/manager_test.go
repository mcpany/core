package upstream

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/prototext"
)

// TestUpstreamServiceManager_Load_LocalOnly tests loading only local services.
func TestUpstreamServiceManager_Load_LocalOnly(t *testing.T) {
	manager := NewUpstreamServiceManager()
	config := &configv1.McpxServerConfig{}
	config.SetUpstreamServices([]*configv1.UpstreamServiceConfig{
		{},
		{},
	})
	config.GetUpstreamServices()[0].SetName("service-a")
	config.GetUpstreamServices()[1].SetName("service-b")

	services, err := manager.Load(config)
	require.NoError(t, err)

	assert.Len(t, services, 2)
	assert.Equal(t, "service-a", services[0].GetName())
	assert.Equal(t, "service-b", services[1].GetName())
}

// TestUpstreamServiceManager_Load_RemoteAndMerge tests merging local and remote services.
func TestUpstreamServiceManager_Load_RemoteAndMerge(t *testing.T) {
	// Mock remote server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/yaml")
		fmt.Fprint(w, `
- name: "service-a"
  version: "v2"
- name: "service-c"
`)
	}))
	defer server.Close()

	manager := NewUpstreamServiceManager()
	config := &configv1.McpxServerConfig{}
	config.SetUpstreamServices([]*configv1.UpstreamServiceConfig{
		{}, // service-a
		{}, // service-b
	})
	config.GetUpstreamServices()[0].SetName("service-a")
	config.GetUpstreamServices()[0].SetVersion("v1")
	config.GetUpstreamServices()[1].SetName("service-b")

	config.SetUpstreamServiceCollections([]*configv1.UpstreamServiceCollection{
		{},
	})
	config.GetUpstreamServiceCollections()[0].SetName("remote-collection")
	config.GetUpstreamServiceCollections()[0].SetHttpUrl(server.URL)

	services, err := manager.Load(config)
	require.NoError(t, err)

	assert.Len(t, services, 3)
	// Service A should be from the remote collection (v2)
	assert.Equal(t, "service-a", services[0].GetName())
	assert.Equal(t, "v1", services[0].GetVersion()) // local service has higher priority
	// Service B is local only
	assert.Equal(t, "service-b", services[1].GetName())
	// Service C is remote only
	assert.Equal(t, "service-c", services[2].GetName())
}

// TestUpstreamServiceManager_Load_PriorityRules tests the priority and tie-breaking logic.
func TestUpstreamServiceManager_Load_PriorityRules(t *testing.T) {
	// Mock remote server
	server1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `- name: "service-a"
  version: "v2"`) // Priority -1
	}))
	defer server1.Close()

	server2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `- name: "service-a"
  version: "v3"`) // Priority -1, loaded later
	}))
	defer server2.Close()

	manager := NewUpstreamServiceManager()
	config := &configv1.McpxServerConfig{}
	config.SetUpstreamServices([]*configv1.UpstreamServiceConfig{
		{}, // service-a
	})
	config.GetUpstreamServices()[0].SetName("service-a")
	config.GetUpstreamServices()[0].SetVersion("local")

	config.SetUpstreamServiceCollections([]*configv1.UpstreamServiceCollection{
		{}, // higher-priority
		{}, // same-priority
	})
	config.GetUpstreamServiceCollections()[0].SetName("higher-priority")
	config.GetUpstreamServiceCollections()[0].SetHttpUrl(server1.URL)
	config.GetUpstreamServiceCollections()[0].SetPriority(-1)
	config.GetUpstreamServiceCollections()[1].SetName("same-priority")
	config.GetUpstreamServiceCollections()[1].SetHttpUrl(server2.URL)
	config.GetUpstreamServiceCollections()[1].SetPriority(-1)

	_, err := manager.Load(config)
	require.NoError(t, err)

	finalServices := manager.GetServices()
	assert.Len(t, finalServices, 1)
	assert.Equal(t, "service-a", finalServices[0].GetName())
	assert.Equal(t, "v2", finalServices[0].GetVersion())
}

// TestUpstreamServiceManager_Load_FailedCollection tests that the manager continues on a failed fetch.
func TestUpstreamServiceManager_Load_FailedCollection(t *testing.T) {
	// Mock a failing server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	server.Close() // Make it fail immediately

	manager := NewUpstreamServiceManager()
	config := &configv1.McpxServerConfig{}
	config.SetUpstreamServices([]*configv1.UpstreamServiceConfig{
		{}, // local-service
	})
	config.GetUpstreamServices()[0].SetName("local-service")
	config.SetUpstreamServiceCollections([]*configv1.UpstreamServiceCollection{
		{}, // failing-collection
	})
	config.GetUpstreamServiceCollections()[0].SetName("failing-collection")
	config.GetUpstreamServiceCollections()[0].SetHttpUrl(server.URL)

	services, err := manager.Load(config)
	require.NoError(t, err)

	// Should still have the local service
	assert.Len(t, services, 1)
	assert.Equal(t, "local-service", services[0].GetName())
}

// TestUpstreamServiceManager_Load_Authentication tests authentication for remote collections.
func TestUpstreamServiceManager_Load_Authentication(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer my-token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/yaml")
		fmt.Fprint(w, `- name: "authed-service"`)
	}))
	defer server.Close()

	manager := NewUpstreamServiceManager()
	config := &configv1.McpxServerConfig{}
	config.SetUpstreamServiceCollections([]*configv1.UpstreamServiceCollection{
		{},
	})
	collection := config.GetUpstreamServiceCollections()[0]
	collection.SetName("authed-collection")
	collection.SetHttpUrl(server.URL)
	auth := &configv1.UpstreamAuthentication{}
	bearerToken := &configv1.UpstreamBearerTokenAuth{}
	bearerToken.SetToken("my-token")
	auth.SetBearerToken(bearerToken)
	collection.SetUpstreamAuthentication(auth)

	services, err := manager.Load(config)
	require.NoError(t, err)
	assert.Len(t, services, 1)
	assert.Equal(t, "authed-service", services[0].GetName())
}

// TestUpstreamServiceManager_Load_ContentTypes tests parsing of different content types.
func TestUpstreamServiceManager_Load_ContentTypes(t *testing.T) {
	// YAML server
	yamlServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/yaml")
		fmt.Fprint(w, `- name: "yaml-service"`)
	}))
	defer yamlServer.Close()

	// JSON server
	jsonServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `[{"name": "json-service"}]`)
	}))
	defer jsonServer.Close()

	// Prototext server
	prototextServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/x-prototext")
		serviceConfig := &configv1.UpstreamServiceConfig{}
		serviceConfig.SetName("prototext-service")
		prototextMsg, _ := prototext.Marshal(serviceConfig)
		fmt.Fprint(w, string(prototextMsg))
	}))
	defer prototextServer.Close()

	manager := NewUpstreamServiceManager()
	config := &configv1.McpxServerConfig{}
	config.SetUpstreamServiceCollections([]*configv1.UpstreamServiceCollection{
		{}, // yaml-collection
		{}, // json-collection
		{}, // prototext-collection
	})
	config.GetUpstreamServiceCollections()[0].SetName("yaml-collection")
	config.GetUpstreamServiceCollections()[0].SetHttpUrl(yamlServer.URL)
	config.GetUpstreamServiceCollections()[1].SetName("json-collection")
	config.GetUpstreamServiceCollections()[1].SetHttpUrl(jsonServer.URL)
	config.GetUpstreamServiceCollections()[2].SetName("prototext-collection")
	config.GetUpstreamServiceCollections()[2].SetHttpUrl(prototextServer.URL)

	services, err := manager.Load(config)
	require.NoError(t, err)

	assert.Len(t, services, 3)
	assert.Equal(t, "json-service", services[0].GetName())
	assert.Equal(t, "prototext-service", services[1].GetName())
	assert.Equal(t, "yaml-service", services[2].GetName())
}
