package app

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/config"
	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/storage"
	"github.com/mcpany/core/server/pkg/util"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func (a *Application) handleCollections(store storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			collections, err := store.ListServiceCollections(r.Context())
			if err != nil {
				logging.GetLogger().Error("failed to list collections", "error", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			opts := protojson.MarshalOptions{UseProtoNames: true}
			var buf []byte
			buf = append(buf, '[')
			for i, c := range collections {
				if i > 0 {
					buf = append(buf, ',')
				}
				b, _ := opts.Marshal(c)
				buf = append(buf, b...)
			}
			buf = append(buf, ']')
			_, _ = w.Write(buf)

		case http.MethodPost:
			var collection configv1.Collection
			body, err := readBodyWithLimit(w, r, 1048576)
			if err != nil {
				return
			}
			if err := protojson.Unmarshal(body, &collection); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if collection.GetName() == "" {
				http.Error(w, "name is required", http.StatusBadRequest)
				return
			}
			if err := store.SaveServiceCollection(r.Context(), &collection); err != nil {
				logging.GetLogger().Error("failed to save collection", "error", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte("{}"))
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func (a *Application) handleCollectionDetail(store storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(r.URL.Path, "/collections/")
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
			collection, err := store.GetServiceCollection(r.Context(), name)
			if err != nil {
				logging.GetLogger().Error("failed to get collection for export", "name", name, "error", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			if collection == nil {
				http.NotFound(w, r)
				return
			}
			exportCollection := proto.Clone(collection).(*configv1.Collection)
			config.StripSecretsFromCollection(exportCollection)
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.json\"", name))
			opts := protojson.MarshalOptions{UseProtoNames: true, Multiline: true, Indent: "  "}
			b, _ := opts.Marshal(exportCollection)
			_, _ = w.Write(b)
			return
		}

		if strings.HasSuffix(name, "/apply") {
			name = strings.TrimSuffix(name, "/apply")
			a.handleCollectionApply(w, r, name, store)
			return
		}

		switch r.Method {
		case http.MethodGet:
			collection, err := store.GetServiceCollection(r.Context(), name)
			if err != nil {
				logging.GetLogger().Error("failed to get collection", "name", name, "error", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			if collection == nil {
				http.NotFound(w, r)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			opts := protojson.MarshalOptions{UseProtoNames: true}
			b, _ := opts.Marshal(collection)
			_, _ = w.Write(b)

		case http.MethodPut:
			var collection configv1.Collection
			body, err := readBodyWithLimit(w, r, 1048576)
			if err != nil {
				return
			}
			if err := protojson.Unmarshal(body, &collection); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			collection.SetName(name) // Force name match

			if err := store.SaveServiceCollection(r.Context(), &collection); err != nil {
				logging.GetLogger().Error("failed to save collection", "name", name, "error", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("{}"))

		case http.MethodDelete:
			if err := store.DeleteServiceCollection(r.Context(), name); err != nil {
				logging.GetLogger().Error("failed to delete collection", "name", name, "error", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusNoContent)

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func (a *Application) handleCollectionApply(w http.ResponseWriter, r *http.Request, name string, store storage.Storage) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	collection, err := store.GetServiceCollection(r.Context(), name)
	if err != nil {
		logging.GetLogger().Error("failed to get collection for apply", "name", name, "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if collection == nil {
		http.NotFound(w, r)
		return
	}

	// Apply services
	for _, rawSvc := range collection.GetServices() {
		svc := proto.Clone(rawSvc).(*configv1.UpstreamServiceConfig)
		// We should probably check if service already exists?
		// "Upsert" logic ideally.
		// And we need to validate it.
		if err := config.ValidateOrError(r.Context(), svc); err != nil {
			logging.GetLogger().Error("invalid service in collection", "service", svc.GetName(), "error", err)
			continue // Skip invalid? Or error out?
		}

		if isUnsafeConfig(svc) && os.Getenv("MCPANY_ALLOW_UNSAFE_CONFIG") != util.TrueStr {
			logging.GetLogger().Warn("Skipping unsafe service in collection apply", "service", svc.GetName())
			continue
		}

		if err := store.SaveService(r.Context(), svc); err != nil {
			logging.GetLogger().Error("failed to save service from collection", "service", svc.GetName(), "error", err)
			// Continue or abort?
			// Maybe best effort?
		}
	}

	// Trigger reload
	if err := a.ReloadConfig(r.Context(), a.fs, a.configPaths); err != nil {
		logging.GetLogger().Error("failed to reload config after collection apply", "error", err)
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("{}"))
}
