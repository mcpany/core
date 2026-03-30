package app

import (
	"net/http"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/storage"
	"google.golang.org/protobuf/encoding/protojson"
)

func (a *Application) handleSettings(store storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			settings, err := store.GetGlobalSettings(r.Context())
			if err != nil {
				logging.GetLogger().Error("failed to get global settings", "error", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			if settings == nil {
				settings = configv1.GlobalSettings_builder{}.Build()
			}
			w.Header().Set("Content-Type", "application/json")
			opts := protojson.MarshalOptions{UseProtoNames: true, EmitUnpopulated: true}
			b, err := opts.Marshal(settings)
			if err != nil {
				logging.GetLogger().Error("failed to marshal settings", "error", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			_, _ = w.Write(b)

		case http.MethodPost:
			var settings configv1.GlobalSettings
			body, err := readBodyWithLimit(w, r, 1048576)
			if err != nil {
				return
			}
			if err := protojson.Unmarshal(body, &settings); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if err := store.SaveGlobalSettings(r.Context(), &settings); err != nil {
				logging.GetLogger().Error("failed to save global settings", "error", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			if err := a.ReloadConfig(r.Context(), a.fs, a.configPaths); err != nil {
				logging.GetLogger().Error("failed to reload config after settings save", "error", err)
			}

			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("{}"))
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}
