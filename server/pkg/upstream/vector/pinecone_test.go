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
	"google.golang.org/protobuf/proto"
)

func TestNewPineconeClient(t *testing.T) {
	// Missing API Key
	_, err := NewPineconeClient(configv1.PineconeVectorDB_builder{}.Build())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "api_key is required")

	// Missing Host and Environment
	_, err = NewPineconeClient(configv1.PineconeVectorDB_builder{
		ApiKey: proto.String("key"),
	}.Build())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "host OR (index_name, project_id, environment) must be provided")

	// Success with Host
	c, err := NewPineconeClient(configv1.PineconeVectorDB_builder{
		ApiKey: proto.String("key"),
		Host:   proto.String("https://custom-host"),
	}.Build())
	assert.NoError(t, err)
	assert.Equal(t, "https://custom-host", c.baseURL)

	// Success with components
	c, err = NewPineconeClient(configv1.PineconeVectorDB_builder{
		ApiKey:      proto.String("key"),
		IndexName:   proto.String("idx"),
		ProjectId:   proto.String("proj"),
		Environment: proto.String("env"),
	}.Build())
	assert.NoError(t, err)
	assert.Equal(t, "https://idx-proj.svc.env.pinecone.io", c.baseURL)
}

func TestPineconeClient_Operations(t *testing.T) {
	// Mock Server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "key", r.Header.Get("Api-Key"))

		switch r.URL.Path {
		case "/query":
			assert.Equal(t, "POST", r.Method)
			var req map[string]interface{}
			json.NewDecoder(r.Body).Decode(&req)
			assert.NotNil(t, req["vector"])
			assert.Equal(t, float64(5), req["topK"])

			json.NewEncoder(w).Encode(map[string]interface{}{
				"matches": []interface{}{
					map[string]interface{}{"id": "1", "score": 0.9},
				},
			})
		case "/vectors/upsert":
			assert.Equal(t, "POST", r.Method)
			var req map[string]interface{}
			json.NewDecoder(r.Body).Decode(&req)
			assert.NotNil(t, req["vectors"])
			json.NewEncoder(w).Encode(map[string]interface{}{"upsertedCount": 1})
		case "/vectors/delete":
			assert.Equal(t, "POST", r.Method)
			var req map[string]interface{}
			json.NewDecoder(r.Body).Decode(&req)
			// Check logic for deleteAll vs ids
			if _, ok := req["deleteAll"]; ok {
				assert.Equal(t, true, req["deleteAll"])
			} else {
				assert.NotNil(t, req["ids"])
			}
			json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
		case "/describe_index_stats":
			assert.Equal(t, "POST", r.Method)
			json.NewEncoder(w).Encode(map[string]interface{}{"totalVectorCount": 100})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	c, err := NewPineconeClient(configv1.PineconeVectorDB_builder{
		ApiKey: proto.String("key"),
		Host:   proto.String(server.URL),
	}.Build())
	assert.NoError(t, err)
	// Override client to trust test server cert if needed, but httptest usually works with default client if using URL
	// Actually default client might fail if server uses https without valid cert, but httptest.NewServer uses http by default unless NewTLSServer
	// PineconeClient creates its own http.Client.
	// Since httptest.NewServer returns http:// URL, standard client works.

	ctx := context.Background()

	// Test Query
	res, err := c.Query(ctx, []float32{0.1, 0.2}, 5, nil, "")
	assert.NoError(t, err)
	assert.NotNil(t, res["matches"])

	// Test Upsert
	res, err = c.Upsert(ctx, []map[string]interface{}{{"id": "1", "values": []float32{0.1}}}, "")
	assert.NoError(t, err)
	assert.Equal(t, float64(1), res["upsertedCount"])

	// Test Delete IDs
	res, err = c.Delete(ctx, []string{"1"}, "", nil)
	assert.NoError(t, err)

	// Test Delete All
	res, err = c.Delete(ctx, nil, "", nil)
	assert.NoError(t, err)

	// Test DescribeIndexStats
	res, err = c.DescribeIndexStats(ctx, nil)
	assert.NoError(t, err)
	assert.Equal(t, float64(100), res["totalVectorCount"])
}

func TestPineconeClient_Errors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "bad request"}`))
	}))
	defer server.Close()

	c, _ := NewPineconeClient(configv1.PineconeVectorDB_builder{
		ApiKey: proto.String("key"),
		Host:   proto.String(server.URL),
	}.Build())

	_, err := c.Query(context.Background(), []float32{0.1}, 1, nil, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "pinecone request failed: status=400")

	// Invalid URL
	c.baseURL = "://invalid-url"
	_, err = c.Query(context.Background(), []float32{0.1}, 1, nil, "")
	assert.Error(t, err)

	// Connection Error
	c.baseURL = "http://127.0.0.1:12345" // Port likely closed
	_, err = c.Query(context.Background(), []float32{0.1}, 1, nil, "")
	assert.Error(t, err)

	// Malformed JSON Response
	badServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{invalid-json`))
	}))
	defer badServer.Close()
	c.baseURL = badServer.URL
	_, err = c.Query(context.Background(), []float32{0.1}, 1, nil, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to unmarshal response")
}
