package config

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestUpstreamServiceManager_LoadAndMergeServices_Errors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/error" {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if r.URL.Path == "/bad-yaml" {
			_, _ = w.Write([]byte(`bad yaml content:`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	t.Run("HTTP Non-200 Error", func(t *testing.T) {
		config := func() *configv1.McpAnyServerConfig {
			col := configv1.Collection_builder{
				Name:    proto.String("error-collection"),
				HttpUrl: proto.String(server.URL + "/error"),
			}.Build()

			return configv1.McpAnyServerConfig_builder{
				Collections: []*configv1.Collection{col},
			}.Build()
		}()
		manager := NewUpstreamServiceManager(nil)
		manager.httpClient = &http.Client{}

		_, err := manager.LoadAndMergeServices(context.Background(), config)
		assert.NoError(t, err)
	})
}
