// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package vector

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func strPtr(s string) *string {
	return &s
}

func TestNewPineconeClient(t *testing.T) {
	tests := []struct {
		name    string
		config  *configv1.PineconeVectorDB
		wantErr bool
		errContains string
	}{
		{
			name: "missing api key",
			config: &configv1.PineconeVectorDB{
				ApiKey: strPtr(""),
			},
			wantErr: true,
			errContains: "api_key is required",
		},
		{
			name: "host provided",
			config: &configv1.PineconeVectorDB{
				ApiKey: strPtr("test-key"),
				Host:   strPtr("https://test.pinecone.io"),
			},
			wantErr: false,
		},
		{
			name: "construct host",
			config: &configv1.PineconeVectorDB{
				ApiKey:      strPtr("test-key"),
				IndexName:   strPtr("index"),
				ProjectId:   strPtr("project"),
				Environment: strPtr("env"),
			},
			wantErr: false,
		},
		{
			name: "missing host components",
			config: &configv1.PineconeVectorDB{
				ApiKey:    strPtr("test-key"),
				IndexName: strPtr("index"),
			},
			wantErr: true,
			errContains: "host OR (index_name, project_id, environment)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewPineconeClient(tt.config)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestPineconeClient_DoRequest(t *testing.T) {
	// Start a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "test-key", r.Header.Get("Api-Key"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		if r.URL.Path == "/query" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"matches": []}`))
			return
		}
		if r.URL.Path == "/error" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error": "bad request"}`))
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := &PineconeClient{
		config: &configv1.PineconeVectorDB{
			ApiKey: strPtr("test-key"),
		},
		client:  server.Client(),
		baseURL: server.URL,
	}

	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		resp, err := client.doRequest(ctx, "/query", map[string]interface{}{})
		require.NoError(t, err)
		assert.NotNil(t, resp)
	})

	t.Run("error response", func(t *testing.T) {
		_, err := client.doRequest(ctx, "/error", nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "pinecone request failed")
	})
}

func TestPineconeClient_Methods(t *testing.T) {
	// Start a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req map[string]interface{}
		json.NewDecoder(r.Body).Decode(&req)

		switch r.URL.Path {
		case "/query":
			assert.Equal(t, "ns1", req["namespace"])
			w.Write([]byte(`{"matches": []}`))
		case "/vectors/upsert":
			assert.Equal(t, "ns1", req["namespace"])
			w.Write([]byte(`{"upsertedCount": 1}`))
		case "/vectors/delete":
			assert.Equal(t, "ns1", req["namespace"])
			w.Write([]byte(`{"success": true}`))
		case "/describe_index_stats":
			w.Write([]byte(`{"totalVectorCount": 10}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client := &PineconeClient{
		config: &configv1.PineconeVectorDB{
			ApiKey: strPtr("test-key"),
		},
		client:  server.Client(),
		baseURL: server.URL,
	}

	ctx := context.Background()

	t.Run("Query", func(t *testing.T) {
		_, err := client.Query(ctx, []float32{0.1}, 10, nil, "ns1")
		require.NoError(t, err)
	})

	t.Run("Upsert", func(t *testing.T) {
		_, err := client.Upsert(ctx, []map[string]interface{}{{"id": "1"}}, "ns1")
		require.NoError(t, err)
	})

	t.Run("Delete", func(t *testing.T) {
		_, err := client.Delete(ctx, []string{"1"}, "ns1", nil)
		require.NoError(t, err)
	})

	t.Run("DescribeIndexStats", func(t *testing.T) {
		_, err := client.DescribeIndexStats(ctx, nil)
		require.NoError(t, err)
	})
}
