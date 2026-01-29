package app

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

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

func (a *Application) handlePromptExecute() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Path: /prompts/{name}/execute
		path := strings.TrimPrefix(r.URL.Path, "/prompts/")
		parts := strings.Split(path, "/")
		if len(parts) < 2 {
			http.Error(w, "invalid path", http.StatusBadRequest)
			return
		}
		name := parts[0]
		action := parts[1]

		if action != "execute" {
			http.NotFound(w, r)
			return
		}

		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Read body as RawMessage
		r.Body = http.MaxBytesReader(w, r.Body, 1048576)
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "failed to read body", http.StatusBadRequest)
			return
		}

		prompt, ok := a.PromptManager.GetPrompt(name)
		if !ok {
			http.NotFound(w, r)
			return
		}

		result, err := prompt.Get(r.Context(), json.RawMessage(body))
		if err != nil {
			logging.GetLogger().Error("failed to execute prompt", "name", name, "error", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(result)
	}
}
