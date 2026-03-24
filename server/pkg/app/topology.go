// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"

	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/storage"
	"google.golang.org/protobuf/encoding/protojson"
)

// handleTopology returns an HTTP handler for the topology API.
//
// Summary: Returns the current service topology graph.
//
// Returns:
//   - http.HandlerFunc: The HTTP handler function.
func (a *Application) handleTopology() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			graph := a.TopologyManager.GetGraph(r.Context())
			w.Header().Set("Content-Type", "application/json")
			opts := protojson.MarshalOptions{UseProtoNames: true, EmitUnpopulated: true}
			b, err := opts.Marshal(graph)
			if err != nil {
				logging.GetLogger().Error("failed to marshal topology graph", "error", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			// ⚡ Bolt Optimization: Compute ETag to allow client-side caching and reduce bandwidth.
			// Randomized Selection from Top 5 High-Impact Targets
			hash := sha256.Sum256(b)
			etag := hex.EncodeToString(hash[:])

			if r.Header.Get("If-None-Match") == etag {
				w.WriteHeader(http.StatusNotModified)
				return
			}

			w.Header().Set("ETag", etag)
			_, _ = w.Write(b)

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

// handleMockSwarmTopology returns an HTTP handler for the mock swarm topology API.
//
// Summary: Returns mock data for the swarm topology widget.
//
// Parameters:
//   - store (storage.Storage): The backend storage.
//
// Returns:
//   - http.HandlerFunc: The HTTP handler function.
func (a *Application) handleMockSwarmTopology(store storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			data, err := store.GetMockData(r.Context(), "swarm-topology")
			if err != nil {
				logging.GetLogger().Error("failed to get mock swarm topology", "error", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			if data == "" {
				// Return an empty valid JSON structure if no mock data is seeded
				data = `{"nodes": [], "edges": [], "anomalies": []}`
			}

			w.Header().Set("Content-Type", "application/json")

			hash := sha256.Sum256([]byte(data))
			etag := hex.EncodeToString(hash[:])

			if r.Header.Get("If-None-Match") == etag {
				w.WriteHeader(http.StatusNotModified)
				return
			}

			w.Header().Set("ETag", etag)
			_, _ = w.Write([]byte(data))

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}
