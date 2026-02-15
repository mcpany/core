package app

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/logging"
	"google.golang.org/protobuf/encoding/protojson"
)

// handleTemplates handles listing and creating service templates.
func (a *Application) handleTemplates() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			templates, err := a.Storage.ListServiceTemplates(r.Context())
			if err != nil {
				logging.GetLogger().Error("failed to list templates", "error", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			opts := protojson.MarshalOptions{UseProtoNames: true, EmitUnpopulated: true}
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
			var tmpl configv1.ServiceTemplate
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

			if err := a.Storage.SaveServiceTemplate(r.Context(), &tmpl); err != nil {
				logging.GetLogger().Error("failed to save template", "error", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusCreated)

			resp := map[string]string{"id": tmpl.GetId()}
			if err := json.NewEncoder(w).Encode(resp); err != nil {
				logging.GetLogger().Error("failed to write response", "error", err)
			}

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

// handleTemplateDetail handles retrieving and deleting a specific service template.
func (a *Application) handleTemplateDetail() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]
		// Fallback if mux vars not set (e.g. unit tests without router)
		if id == "" {
			pathParts := strings.Split(r.URL.Path, "/")
			if len(pathParts) > 0 {
				id = pathParts[len(pathParts)-1]
			}
		}

		if id == "" {
			http.Error(w, "id required", http.StatusBadRequest)
			return
		}

		switch r.Method {
		case http.MethodGet:
			tmpl, err := a.Storage.GetServiceTemplate(r.Context(), id)
			if err != nil {
				logging.GetLogger().Error("failed to get template", "id", id, "error", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			if tmpl == nil {
				http.NotFound(w, r)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			b, _ := protojson.MarshalOptions{UseProtoNames: true, EmitUnpopulated: true}.Marshal(tmpl)
			_, _ = w.Write(b)

		case http.MethodDelete:
			if err := a.Storage.DeleteServiceTemplate(r.Context(), id); err != nil {
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
