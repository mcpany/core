package app

import (
	"encoding/json"
	"net/http"
)

func (a *Application) handleResources() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			resources := a.ResourceManager.ListResources()
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resources)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}
