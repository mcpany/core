// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"

	"github.com/mcpany/core/server/pkg/logging"
	"google.golang.org/protobuf/encoding/protojson"
)

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
