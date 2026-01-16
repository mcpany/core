// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
)

func TestDeepValidator_ValidateHTTP(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	tsError := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer tsError.Close()

	validator := NewDeepValidator(2 * time.Second)

	t.Run("Valid HTTP Service", func(t *testing.T) {
		config := &configv1.McpAnyServerConfig{
			UpstreamServices: []*configv1.UpstreamServiceConfig{
				{
					Name: proto.String("valid-service"),
					ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
						HttpService: &configv1.HttpUpstreamService{
							Address: proto.String(ts.URL),
						},
					},
				},
			},
		}

		errors := validator.Validate(context.Background(), config)
		assert.Empty(t, errors)
	})

	t.Run("Invalid HTTP Service (Connection Refused)", func(t *testing.T) {
		config := &configv1.McpAnyServerConfig{
			UpstreamServices: []*configv1.UpstreamServiceConfig{
				{
					Name: proto.String("invalid-service"),
					ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
						HttpService: &configv1.HttpUpstreamService{
							Address: proto.String("http://localhost:54321"),
						},
					},
				},
			},
		}

		errors := validator.Validate(context.Background(), config)
		require.Len(t, errors, 1)
		assert.Equal(t, "invalid-service", errors[0].ServiceName)
		assert.Contains(t, errors[0].Err.Error(), "connection refused")
	})

	t.Run("Invalid HTTP Service (Server Error)", func(t *testing.T) {
		config := &configv1.McpAnyServerConfig{
			UpstreamServices: []*configv1.UpstreamServiceConfig{
				{
					Name: proto.String("error-service"),
					ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
						HttpService: &configv1.HttpUpstreamService{
							Address: proto.String(tsError.URL),
						},
					},
				},
			},
		}

		errors := validator.Validate(context.Background(), config)
		require.Len(t, errors, 1)
		assert.Equal(t, "error-service", errors[0].ServiceName)
		assert.Contains(t, errors[0].Err.Error(), "server returned server error: 500")
	})
}

func TestDeepValidator_ValidateGRPC(t *testing.T) {
	// Start a dummy gRPC server
	listener, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)
	s := grpc.NewServer()
	go func() {
		if err := s.Serve(listener); err != nil {
			panic(err)
		}
	}()
	defer s.Stop()

	validator := NewDeepValidator(2 * time.Second)

	t.Run("Valid gRPC Service", func(t *testing.T) {
		config := &configv1.McpAnyServerConfig{
			UpstreamServices: []*configv1.UpstreamServiceConfig{
				{
					Name: proto.String("valid-grpc"),
					ServiceConfig: &configv1.UpstreamServiceConfig_GrpcService{
						GrpcService: &configv1.GrpcUpstreamService{
							Address: proto.String(listener.Addr().String()),
							// Use insecure transport for the test server
							// Note: Real implementation uses TlsConfig to decide credentials.
							// Since we don't pass TlsConfig, it defaults to insecure in our implementation,
							// which matches the test server.
						},
					},
				},
			},
		}

		errors := validator.Validate(context.Background(), config)
		assert.Empty(t, errors)
	})

	t.Run("Invalid gRPC Service (Unreachable)", func(t *testing.T) {
		config := &configv1.McpAnyServerConfig{
			UpstreamServices: []*configv1.UpstreamServiceConfig{
				{
					Name: proto.String("invalid-grpc"),
					ServiceConfig: &configv1.UpstreamServiceConfig_GrpcService{
						GrpcService: &configv1.GrpcUpstreamService{
							Address: proto.String("localhost:54322"),
						},
					},
				},
			},
		}

		errors := validator.Validate(context.Background(), config)
		require.Len(t, errors, 1)
		assert.Equal(t, "invalid-grpc", errors[0].ServiceName)
		assert.Contains(t, errors[0].Err.Error(), "failed to dial gRPC server")
	})
}

func TestDeepValidator_ValidateSQL(t *testing.T) {
	// Use sqlmock to simulate a database
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	// We can't easily inject the mock DB into the validator because it calls sql.Open directly.
	// Ideally, we'd refactor to allow injecting a driver or connector.
	// For now, we'll test with the sqlite driver which is in-memory.

	validator := NewDeepValidator(2 * time.Second)

	t.Run("Valid SQLite Service", func(t *testing.T) {
		config := &configv1.McpAnyServerConfig{
			UpstreamServices: []*configv1.UpstreamServiceConfig{
				{
					Name: proto.String("valid-sql"),
					ServiceConfig: &configv1.UpstreamServiceConfig_SqlService{
						SqlService: &configv1.SqlUpstreamService{
							Driver: proto.String("sqlite"),
							Dsn:    proto.String(":memory:"),
						},
					},
				},
			},
		}

		errors := validator.Validate(context.Background(), config)
		assert.Empty(t, errors)
	})

	t.Run("Invalid SQL Service (Bad DSN)", func(t *testing.T) {
		// Note: sqlite often accepts anything as a file path, so it's hard to make it fail "connection".
		// But for other drivers like postgres/mysql it would try to connect.
		// Let's try a postgres one (requires running postgres or mock).
		// Since we can't easily mock `sql.Open` global state without side effects,
		// we will skip testing the *failure* case for SQL with a real driver unless we have an integration environment.
		// However, we can test that it ATTEMPTS to connect and fails if the driver is known but target is unreachable.

		config := &configv1.McpAnyServerConfig{
			UpstreamServices: []*configv1.UpstreamServiceConfig{
				{
					Name: proto.String("invalid-sql"),
					ServiceConfig: &configv1.UpstreamServiceConfig_SqlService{
						SqlService: &configv1.SqlUpstreamService{
							Driver: proto.String("postgres"),
							Dsn:    proto.String("postgres://user:pass@localhost:54323/db?sslmode=disable"),
						},
					},
				},
			},
		}

		errors := validator.Validate(context.Background(), config)
		require.Len(t, errors, 1)
		assert.Equal(t, "invalid-sql", errors[0].ServiceName)
		// Error message depends on the driver, but usually involves connection refused
		assert.Contains(t, errors[0].Err.Error(), "connection refused")
	})
}

// TestDeepValidator_Concurrency verifies that validation happens in parallel.
func TestDeepValidator_Concurrency(t *testing.T) {
	// Create a slow server
	slowHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond) // Sleep 100ms
		w.WriteHeader(http.StatusOK)
	})
	ts := httptest.NewServer(slowHandler)
	defer ts.Close()

	validator := NewDeepValidator(2 * time.Second)

	// Configure 10 services pointing to the slow server
	numServices := 10
	services := make([]*configv1.UpstreamServiceConfig, numServices)
	for i := 0; i < numServices; i++ {
		services[i] = &configv1.UpstreamServiceConfig{
			Name: proto.String(string(rune('a' + i))),
			ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
				HttpService: &configv1.HttpUpstreamService{
					Address: proto.String(ts.URL),
				},
			},
		}
	}

	config := &configv1.McpAnyServerConfig{
		UpstreamServices: services,
	}

	start := time.Now()
	errors := validator.Validate(context.Background(), config)
	duration := time.Since(start)

	assert.Empty(t, errors)
	// If sequential, it would take > 1s (10 * 100ms).
	// If parallel, it should take ~100ms + overhead.
	// We'll assert it takes less than 500ms.
	assert.Less(t, duration, 500*time.Millisecond, "Validation took too long, likely sequential")
}
