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

		cols, err := a.Storage.ListRequestCollections(r.Context())
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

		if err := a.Storage.SaveRequestCollection(r.Context(), &col); err != nil {
			http.Error(w, "Failed to save collection", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		// Return the saved object
		resp, _ := protojson.Marshal(&col)
		if _, err := w.Write(resp); err != nil {
			// Failed to write response, cannot do much but log if we had a logger
		}
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

		if err := a.Storage.DeleteRequestCollection(r.Context(), id); err != nil {
			http.Error(w, "Failed to delete collection", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
