package config

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestLoadFromURL_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"services": [{"name": "s1", "http_service": {"address": "http://127.0.0.1"}}]}`))
	}))
	defer ts.Close()

	m := NewUpstreamServiceManager(nil)
	m.httpClient = http.DefaultClient // Allow 127.0.0.1
	collection := configv1.Collection_builder{
		Name: proto.String("test"),
	}.Build()

	err := m.loadFromURL(context.Background(), ts.URL, collection)
	require.NoError(t, err)
	assert.Len(t, m.services, 1)
}

func TestLoadFromURL_Auth(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"services": []}`))
	}))
	defer ts.Close()

	m := NewUpstreamServiceManager(nil)
	m.httpClient = http.DefaultClient // Allow 127.0.0.1
	collection := configv1.Collection_builder{
		Name: proto.String("test"),
		Authentication: configv1.Authentication_builder{
			BearerToken: configv1.BearerTokenAuth_builder{
				Token: configv1.SecretValue_builder{
					PlainText: proto.String("token"),
				}.Build(),
			}.Build(),
		}.Build(),
	}.Build()

	err := m.loadFromURL(context.Background(), ts.URL, collection)
	require.NoError(t, err)
}

func TestLoadFromURL_Error(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	m := NewUpstreamServiceManager(nil)
	m.httpClient = http.DefaultClient // Allow 127.0.0.1
	collection := configv1.Collection_builder{Name: proto.String("test")}.Build()

	err := m.loadFromURL(context.Background(), ts.URL, collection)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "status code 500")
}

func TestLoadFromURL_RequestFail(t *testing.T) {
	m := NewUpstreamServiceManager(nil)
	collection := configv1.Collection_builder{Name: proto.String("test")}.Build()

	// Invalid URL
	err := m.loadFromURL(context.Background(), "http://invalid-url", collection)
	assert.Error(t, err)
}

func TestLoadFromURL_BadAuth(t *testing.T) {
	m := NewUpstreamServiceManager(nil)
	// Env var that doesn't exist
	collection := configv1.Collection_builder{
		Name: proto.String("test"),
		Authentication: configv1.Authentication_builder{
			BearerToken: configv1.BearerTokenAuth_builder{
				Token: configv1.SecretValue_builder{
					EnvironmentVariable: proto.String("MISSING_VAR"),
				}.Build(),
			}.Build(),
		}.Build(),
	}.Build()
	// We need context with config that fails?
	// ResolveSecret checks os.LookupEnv.

	err := m.loadFromURL(context.Background(), "http://example.com", collection)
	// It should fail at applyAuthentication
	assert.Error(t, err)
}
