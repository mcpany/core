package app

import (
	"encoding/json"
	"net/http"
	"strings"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/auth"
	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/storage"
	"google.golang.org/protobuf/encoding/protojson"
)

func (a *Application) handleSecrets(store storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			secrets, err := store.ListSecrets(r.Context())
			if err != nil {
				logging.GetLogger().Error("failed to list secrets", "error", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			// Redact sensitive values
			for _, s := range secrets {
				s.SetValue("[REDACTED]")
			}
			w.Header().Set("Content-Type", "application/json")
			opts := protojson.MarshalOptions{UseProtoNames: true}
			var buf []byte
			buf = append(buf, '[')
			for i, s := range secrets {
				if i > 0 {
					buf = append(buf, ',')
				}
				b, _ := opts.Marshal(s)
				buf = append(buf, b...)
			}
			buf = append(buf, ']')
			_, _ = w.Write(buf)

		case http.MethodPost:
			var secret configv1.Secret
			body, err := readBodyWithLimit(w, r, 1048576)
			if err != nil {
				return
			}
			if err := protojson.Unmarshal(body, &secret); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			if secret.GetId() == "" {
				http.Error(w, "id is required", http.StatusBadRequest)
				return
			}

			// Validate skipped as config.ValidateOrError expects UpstreamServiceConfig

			if err := store.SaveSecret(r.Context(), &secret); err != nil {
				logging.GetLogger().Error("failed to save secret", "error", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			// Reload
			if err := a.ReloadConfig(r.Context(), a.fs, a.configPaths); err != nil {
				logging.GetLogger().Error("failed to reload config after secret save", "error", err)
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("{}"))
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func (a *Application) handleSecretDetail(store storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/secrets/")
		if path == "" {
			http.Error(w, "id required", http.StatusBadRequest)
			return
		}

		if strings.HasSuffix(path, "/reveal") {
			id := strings.TrimSuffix(path, "/reveal")
			if id == "" {
				http.Error(w, "id required", http.StatusBadRequest)
				return
			}
			a.handleSecretReveal(w, r, id, store)
			return
		}

		switch r.Method {
		case http.MethodGet:
			secret, err := store.GetSecret(r.Context(), path)
			if err != nil {
				logging.GetLogger().Error("failed to get secret", "id", path, "error", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			if secret == nil {
				http.NotFound(w, r)
				return
			}
			// Redact
			secret.SetValue("[REDACTED]")
			w.Header().Set("Content-Type", "application/json")
			opts := protojson.MarshalOptions{UseProtoNames: true}
			b, _ := opts.Marshal(secret)
			_, _ = w.Write(b)

		case http.MethodPut:
			var secret configv1.Secret
			body, err := readBodyWithLimit(w, r, 1048576)
			if err != nil {
				return
			}
			if err := protojson.Unmarshal(body, &secret); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if secret.GetName() == "" && secret.GetId() != "" {
				secret.SetName(secret.GetId())
			}

			// Force ID
			secret.SetId(path)

			if err := store.SaveSecret(r.Context(), &secret); err != nil {
				logging.GetLogger().Error("failed to save secret", "error", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			if err := a.ReloadConfig(r.Context(), a.fs, a.configPaths); err != nil {
				logging.GetLogger().Error("failed to reload config after secret update", "error", err)
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("{}"))

		case http.MethodDelete:
			if err := store.DeleteSecret(r.Context(), path); err != nil {
				logging.GetLogger().Error("failed to delete secret", "id", path, "error", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			if err := a.ReloadConfig(r.Context(), a.fs, a.configPaths); err != nil {
				logging.GetLogger().Error("failed to reload config after secret delete", "error", err)
			}
			w.WriteHeader(http.StatusNoContent)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func (a *Application) handleSecretReveal(w http.ResponseWriter, r *http.Request, id string, store storage.Storage) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	secret, err := store.GetSecret(r.Context(), id)
	if err != nil {
		logging.GetLogger().Error("failed to get secret for reveal", "id", id, "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if secret == nil {
		http.NotFound(w, r)
		return
	}

	// Log the access (Audit)
	user, _ := auth.UserFromContext(r.Context())
	logging.GetLogger().Info("Secret revealed", "id", id, "user", user)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"value": secret.GetValue(),
	})
}
