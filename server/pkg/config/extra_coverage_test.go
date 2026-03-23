// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

// TestWrapActionableError_Nested covers the path where WrapActionableError
// wraps an existing ActionableError.
func TestWrapActionableError_Nested(t *testing.T) {
	innerErr := &ActionableError{
		Err:        errors.New("inner error"),
		Suggestion: "fix inner",
	}

	wrapped := WrapActionableError("outer context", innerErr)
	assert.Error(t, wrapped)

	// Verify it preserved the type
	var ae *ActionableError
	require.True(t, errors.As(wrapped, &ae))

	assert.Contains(t, ae.Err.Error(), "outer context: inner error")
	assert.Equal(t, "fix inner", ae.Suggestion)

	// Verify .Error() output
	assert.Contains(t, wrapped.Error(), "outer context: inner error")
	assert.Contains(t, wrapped.Error(), "-> Fix: fix inner")
}

func TestActionableError_Unwrap(t *testing.T) {
	baseErr := errors.New("base error")
	ae := &ActionableError{Err: baseErr}
	assert.Equal(t, baseErr, ae.Unwrap())
	assert.True(t, errors.Is(ae, baseErr))
}

// TestLoadServices_InvalidBinary covers the "unknown binary type" error path.
func TestLoadServices_InvalidBinary(t *testing.T) {
	store := NewFileStore(afero.NewMemMapFs(), []string{})
	cfg, err := LoadServices(context.Background(), store, "unknown")
	assert.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "unknown binary type: unknown")
}

// TestStore_SkipValidation_Logic verifies that SetSkipValidation works.
func TestStore_SkipValidation_Logic(t *testing.T) {
    // This tests the plumbing of SetSkipValidation from FileStore to yamlEngine.
	fs := afero.NewMemMapFs()
	configContent := `
global_settings:
  mcp_listen_address: "127.0.0.1:9090"
`
	err := afero.WriteFile(fs, "/config.yaml", []byte(configContent), 0o644)
	require.NoError(t, err)

	store := NewFileStore(fs, []string{"/config.yaml"})
	store.SetSkipValidation(true)

    cfg, err := store.Load(context.Background())
    require.NoError(t, err)
    assert.Equal(t, "127.0.0.1:9090", cfg.GetGlobalSettings().GetMcpListenAddress())
}

// TestLoadResolvedConfig_Empty verifies LoadResolvedConfig when store returns empty.
func TestLoadResolvedConfig_Empty(t *testing.T) {
    fs := afero.NewMemMapFs()
    // Empty file
    err := afero.WriteFile(fs, "/config.yaml", []byte(""), 0o644)
    require.NoError(t, err)

    store := NewFileStore(fs, []string{"/config.yaml"})
    cfg, err := LoadResolvedConfig(context.Background(), store)
    // Should fail with "configuration sources provided but loaded configuration is empty"
    assert.Error(t, err)
    assert.Nil(t, cfg)
    assert.Contains(t, err.Error(), "configuration sources provided but loaded configuration is empty")
}

// TestLoadResolvedConfig_NoSources verifies default config when no sources.
func TestLoadResolvedConfig_NoSources(t *testing.T) {
    store := NewFileStore(afero.NewMemMapFs(), []string{})
    cfg, err := LoadResolvedConfig(context.Background(), store)
    require.NoError(t, err)
    assert.NotNil(t, cfg)
    // Should have default user
    assert.NotEmpty(t, cfg.GetUsers())
    assert.Equal(t, "default", cfg.GetUsers()[0].GetId())
}

// TestLoadServices_ActionableError covers handling of ActionableError in LoadServices.
func TestLoadServices_ActionableError(t *testing.T) {
	mockErr := &ActionableError{
		Err:        errors.New("mock error"),
		Suggestion: "do something",
	}
	store := &MockStoreForError{err: mockErr}

	cfg, err := LoadServices(context.Background(), store, "server")
	assert.Error(t, err)
	assert.Nil(t, cfg)

	// Should be formatted as ActionableError
	assert.Contains(t, err.Error(), "❌ Configuration Loading Failed")
	assert.Contains(t, err.Error(), "mock error")
	assert.Contains(t, err.Error(), "💡 Fix: do something")
}

type MockStoreForError struct {
	err error
}

func (s *MockStoreForError) Load(ctx context.Context) (*configv1.McpAnyServerConfig, error) {
	return nil, s.err
}

func (s *MockStoreForError) HasConfigSources() bool {
	return true
}

// TestValidate_DuplicateService covers duplicate service check in Validate.
func TestValidate_DuplicateService(t *testing.T) {
	cfg := func() *configv1.McpAnyServerConfig {
		svc1 := configv1.UpstreamServiceConfig_builder{
			Name: proto.String("svc1"),
			HttpService: configv1.HttpUpstreamService_builder{
				Address: proto.String("http://a.com"),
			}.Build(),
		}.Build()

		svc2 := configv1.UpstreamServiceConfig_builder{
			Name: proto.String("svc1"),
			HttpService: configv1.HttpUpstreamService_builder{
				Address: proto.String("http://b.com"),
			}.Build(),
		}.Build()

		return configv1.McpAnyServerConfig_builder{
			UpstreamServices: []*configv1.UpstreamServiceConfig{svc1, svc2},
		}.Build()
	}()
	errs := Validate(context.Background(), cfg, Server)
	assert.NotEmpty(t, errs)
	found := false
	for _, e := range errs {
		if e.ServiceName == "svc1" && e.Err.Error() == "duplicate service name found" {
			found = true
			break
		}
	}
	assert.True(t, found, "Expected duplicate service name error")
}

// TestValidate_DuplicateUser covers duplicate user check.
func TestValidate_DuplicateUser(t *testing.T) {
	cfg := func() *configv1.McpAnyServerConfig {
		u1 := configv1.User_builder{
			Id: proto.String("u1"),
		}.Build()

		u2 := configv1.User_builder{
			Id: proto.String("u1"),
		}.Build()

		return configv1.McpAnyServerConfig_builder{
			Users: []*configv1.User{u1, u2},
		}.Build()
	}()
	errs := Validate(context.Background(), cfg, Server)
	assert.NotEmpty(t, errs)
	found := false
	for _, e := range errs {
		if e.ServiceName == "user:u1" && e.Err.Error() == "duplicate user id" {
			found = true
			break
		}
	}
	assert.True(t, found, "Expected duplicate user id error")
}


func TestUpstreamServiceManager_LoadFromURL_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{"services": [{"name": "remote-service", "http_service": {"address": "http://remote.com"}}]}`)
	}))
	defer ts.Close()

	manager := NewUpstreamServiceManager([]string{"default"})
	manager.httpClient = ts.Client()

	config := func() *configv1.McpAnyServerConfig {
		col := configv1.Collection_builder{
			Name:    proto.String("remote-collection"),
			HttpUrl: proto.String(ts.URL),
		}.Build()

		return configv1.McpAnyServerConfig_builder{
			Collections: []*configv1.Collection{col},
		}.Build()
	}()

	services, err := manager.LoadAndMergeServices(context.Background(), config)
	require.NoError(t, err)
	assert.Len(t, services, 1)
	assert.Equal(t, "remote-service", services[0].GetName())
}

func TestUpstreamServiceManager_LoadFromURL_Failure(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	manager := NewUpstreamServiceManager([]string{"default"})
	manager.httpClient = ts.Client()

	config := func() *configv1.McpAnyServerConfig {
		col := configv1.Collection_builder{
			Name:    proto.String("remote-collection"),
			HttpUrl: proto.String(ts.URL),
		}.Build()

		return configv1.McpAnyServerConfig_builder{
			Collections: []*configv1.Collection{col},
		}.Build()
	}()

	// LoadAndMergeServices swallows collection loading errors but logs them
	services, err := manager.LoadAndMergeServices(context.Background(), config)
	require.NoError(t, err)
	assert.Len(t, services, 0)
}
