// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"encoding/json"
	"net/http"

	"github.com/mcpany/core/server/pkg/logging"
)

func (a *Application) handleResourceRead() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		uri := r.URL.Query().Get("uri")
		if uri == "" {
			http.Error(w, "uri required", http.StatusBadRequest)
			return
		}

		res, ok := a.ResourceManager.GetResource(uri)
		if !ok {
			http.NotFound(w, r)
			return
		}

		result, err := res.Read(r.Context())
		if err != nil {
			logging.GetLogger().Error("failed to read resource", "uri", uri, "error", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(result)
	}
}
