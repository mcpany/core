package app

import (
	"encoding/json"
	"net/http"

	"github.com/mcpany/core/server/pkg/logging"
	"google.golang.org/protobuf/encoding/protojson"
)

func (a *Application) handlePrompts() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			prompts := a.PromptManager.ListPrompts()
			w.Header().Set("Content-Type", "application/json")

			var jsonPrompts []map[string]any
			opts := protojson.MarshalOptions{UseProtoNames: false, EmitUnpopulated: false}

			for _, p := range prompts {
				if p.Definition() == nil {
					continue
				}
				b, err := opts.Marshal(p.Definition())
				if err != nil {
					logging.GetLogger().Error("failed to marshal prompt", "name", p.Definition().GetName(), "error", err)
					continue
				}
				var m map[string]any
				if err := json.Unmarshal(b, &m); err == nil {
					m["serviceId"] = p.Service()
					jsonPrompts = append(jsonPrompts, m)
				}
			}

			_ = json.NewEncoder(w).Encode(jsonPrompts)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}
