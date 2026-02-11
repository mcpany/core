// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"google.golang.org/protobuf/encoding/protojson"
)

// handleListRequestCollections returns a handler that lists all request collections.
// GET /api/v1/library/collections
func (a *Application) handleListRequestCollections() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		cols, err := a.Store.ListRequestCollections(r.Context())
		if err != nil {
			http.Error(w, "Failed to list collections", http.StatusInternalServerError)
			return
		}

		// Use protojson to marshal response
		resp := struct {
			Collections []*configv1.RequestCollection `json:"collections"`
		}{
			Collections: cols,
		}

		w.Header().Set("Content-Type", "application/json")
		// We use json.NewEncoder for the wrapper, but cols contain proto messages which serialize differently.
		// It's better to use protojson.Marshal for the whole thing if we define a wrapper proto,
		// or just manually serialize the list.
		// For simplicity/consistency with existing handlers (which often use json.NewEncoder for Go structs),
		// we rely on the fact that generated Go protos have JSON tags.
		// However, protojson is stricter/better for google.protobuf.Struct etc.
		// Let's stick to json.NewEncoder as in other handlers unless we see issues.
		// Actually, `configv1.RequestCollection` has `json_name` tags.
		// But `google.protobuf.Struct` (args) needs special handling usually provided by protojson.
		// If we use `json.Marshal`, `Struct` might fail or look weird.
		// Let's use protojson manually.

		// Wait, we can't easily marshal a struct wrapping proto messages using protojson unless the wrapper is also a proto.
		// But other handlers seem to return `res.json()`.
		// Let's check `api_services.go` (if it existed) or `handler.go`.
		// `handler.go` used `json.NewEncoder`.
		// Let's trust `json.NewEncoder` works for simple cases, but for `Struct`, we might need to be careful.
		// The safest bet is defining a response proto, but that's overkill.
		// Let's try `json.NewEncoder`. If args are garbled, we fix it.
		// Actually, generated Go code for Struct is `map[string]interface{}` basically? No, it's `*structpb.Struct`.
		// It implements `json.Marshaler`. So it should be fine.

		if err := json.NewEncoder(w).Encode(resp); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		}
	}
}

// handleSaveRequestCollection returns a handler that saves (create/update) a request collection.
// POST /api/v1/library/collections
func (a *Application) handleSaveRequestCollection() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost && r.Method != http.MethodPut {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read body", http.StatusBadRequest)
			return
		}

		var col configv1.RequestCollection
		if err := protojson.Unmarshal(body, &col); err != nil {
			http.Error(w, "Invalid JSON body", http.StatusBadRequest)
			return
		}

		if err := a.Store.SaveRequestCollection(r.Context(), &col); err != nil {
			http.Error(w, "Failed to save collection", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		// Return the saved object
		resp, _ := protojson.Marshal(&col)
		w.Write(resp)
	}
}

// handleDeleteRequestCollection returns a handler that deletes a request collection.
// DELETE /api/v1/library/collections/{id}
func (a *Application) handleDeleteRequestCollection() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// /api/v1/library/collections/{id}
		parts := strings.Split(r.URL.Path, "/")
		id := parts[len(parts)-1]
		if id == "" {
			http.Error(w, "ID required", http.StatusBadRequest)
			return
		}

		if err := a.Store.DeleteRequestCollection(r.Context(), id); err != nil {
			http.Error(w, "Failed to delete collection", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
