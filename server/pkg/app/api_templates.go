// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/google/uuid"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/logging"
	"google.golang.org/protobuf/encoding/protojson"
)

func (a *Application) handleTemplates() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			templates := a.TemplateManager.ListTemplates()
			w.Header().Set("Content-Type", "application/json")

			// Wrap in object? Or list? protojson marshals slice if passed? No, protojson marshals Message.
			// We need to marshal list manually.
			opts := protojson.MarshalOptions{UseProtoNames: true}
			var buf []byte
			buf = append(buf, '[')
			for i, t := range templates {
				if i > 0 {
					buf = append(buf, ',')
				}
				b, _ := opts.Marshal(t)
				buf = append(buf, b...)
			}
			buf = append(buf, ']')
			_, _ = w.Write(buf)

		case http.MethodPost:
			// Limit 1MB
			r.Body = http.MaxBytesReader(w, r.Body, 1048576)
			var tmpl configv1.UpstreamServiceConfig
			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "failed to read body", http.StatusBadRequest)
				return
			}

			if err := protojson.Unmarshal(body, &tmpl); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			if tmpl.GetId() == "" {
				tmpl.SetId(uuid.New().String())
			}

			if err := a.TemplateManager.SaveTemplate(&tmpl); err != nil {
				logging.GetLogger().Error("failed to save template", "error", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusOK)

			resp := map[string]string{"id": tmpl.GetId()}
			if err := json.NewEncoder(w).Encode(resp); err != nil {
				logging.GetLogger().Error("failed to write response", "error", err)
			}

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func (a *Application) handleTemplateDetail() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/templates/")
		if id == "" {
			http.Error(w, "id required", http.StatusBadRequest)
			return
		}

		switch r.Method {
		case http.MethodDelete:
			if err := a.TemplateManager.DeleteTemplate(id); err != nil {
				logging.GetLogger().Error("failed to delete template", "id", id, "error", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}
