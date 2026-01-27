// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"
	"strings"

	"github.com/mcpany/core/server/pkg/logging"
)

func (a *Application) handleResourceDownload() http.HandlerFunc {
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

		if len(result.Contents) == 0 {
			http.Error(w, "resource is empty", http.StatusNoContent)
			return
		}

		content := result.Contents[0]

		// Determine filename
		filename := "resource"
		if res.Resource().Name != "" {
			filename = res.Resource().Name
		} else {
			// Try to extract from URI
			base := path.Base(uri)
			if base != "" && base != "." && base != "/" {
				filename = base
			}
		}

		// Sanitize filename (basic)
		filename = strings.ReplaceAll(filename, "/", "_")
		filename = strings.ReplaceAll(filename, "\\", "_")

		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
		w.Header().Set("Content-Type", content.MIMEType)

		if len(content.Blob) > 0 {
			_, _ = w.Write(content.Blob)
		} else {
			_, _ = w.Write([]byte(content.Text))
		}
	}
}

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
