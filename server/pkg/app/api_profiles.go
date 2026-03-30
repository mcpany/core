package app

import (
	"fmt"
	"net/http"
	"strings"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/config"
	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/storage"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func (a *Application) handleProfiles(store storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			// Mix of config defined and DB defined?
			// Profiles are in GlobalSettings.
			// Currently GlobalSettings are single object.
			// But DB can store user profiles separately?
			// The handler seems to treat them as separate entities, but config stores them in GlobalSettings.ProfileDefinitions.
			// Storage methods for Profiles might map to GlobalSettings mutation.

			// Assuming Store.ListProfiles exists (it usually extracts from GlobalSettings)
			profiles, err := store.ListProfiles(r.Context())
			if err != nil {
				logging.GetLogger().Error("failed to list profiles", "error", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			opts := protojson.MarshalOptions{UseProtoNames: true}
			var buf []byte
			buf = append(buf, '[')
			for i, p := range profiles {
				if i > 0 {
					buf = append(buf, ',')
				}
				b, _ := opts.Marshal(p)
				buf = append(buf, b...)
			}
			buf = append(buf, ']')
			_, _ = w.Write(buf)

		case http.MethodPost:
			var profile configv1.ProfileDefinition
			body, err := readBodyWithLimit(w, r, 1048576)
			if err != nil {
				return
			}
			if err := protojson.Unmarshal(body, &profile); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if profile.GetName() == "" {
				http.Error(w, "name is required", http.StatusBadRequest)
				return
			}
			// ProfileDefinition uses Name as identifier, no ID field.

			if err := store.SaveProfile(r.Context(), &profile); err != nil {
				logging.GetLogger().Error("failed to save profile", "error", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			// Trigger reload
			if err := a.ReloadConfig(r.Context(), a.fs, a.configPaths); err != nil {
				logging.GetLogger().Error("failed to reload config after profile save", "error", err)
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("{}"))
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func (a *Application) handleProfileDetail(store storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(r.URL.Path, "/profiles/")
		if name == "" {
			http.Error(w, "name required", http.StatusBadRequest)
			return
		}

		if strings.HasSuffix(name, "/export") {
			name = strings.TrimSuffix(name, "/export")
			if r.Method != http.MethodGet {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			profile, err := store.GetProfile(r.Context(), name)
			if err != nil {
				logging.GetLogger().Error("failed to get profile for export", "name", name, "error", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			if profile == nil {
				http.NotFound(w, r)
				return
			}
			exportProfile := proto.Clone(profile).(*configv1.ProfileDefinition)
			config.StripSecretsFromProfile(exportProfile)
			w.Header().Set("Content-Type", "application/json")
			// Force download? Maybe 'Content-Disposition: attachment; filename="profile.json"'
			w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.json\"", name))
			opts := protojson.MarshalOptions{UseProtoNames: true, Multiline: true, Indent: "  "}
			b, _ := opts.Marshal(exportProfile)
			_, _ = w.Write(b)
			return
		}

		switch r.Method {
		case http.MethodGet:
			profile, err := store.GetProfile(r.Context(), name)
			if err != nil {
				logging.GetLogger().Error("failed to get profile", "name", name, "error", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			if profile == nil {
				http.NotFound(w, r)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			opts := protojson.MarshalOptions{UseProtoNames: true}
			b, _ := opts.Marshal(profile)
			_, _ = w.Write(b)

		case http.MethodPut:
			var profile configv1.ProfileDefinition
			body, err := readBodyWithLimit(w, r, 1048576)
			if err != nil {
				return
			}
			if err := protojson.Unmarshal(body, &profile); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			profile.SetName(name) // Force name match

			if err := store.SaveProfile(r.Context(), &profile); err != nil {
				logging.GetLogger().Error("failed to save profile", "name", name, "error", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			if err := a.ReloadConfig(r.Context(), a.fs, a.configPaths); err != nil {
				logging.GetLogger().Error("failed to reload config after profile update", "error", err)
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("{}"))

		case http.MethodDelete:
			if err := store.DeleteProfile(r.Context(), name); err != nil {
				logging.GetLogger().Error("failed to delete profile", "name", name, "error", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			if err := a.ReloadConfig(r.Context(), a.fs, a.configPaths); err != nil {
				logging.GetLogger().Error("failed to reload config after profile delete", "error", err)
			}
			w.WriteHeader(http.StatusNoContent)

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}
