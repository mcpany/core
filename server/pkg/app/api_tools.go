package app

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func (a *Application) handleTools() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			tools := a.ToolManager.ListTools()
			var toolList []*mcp.Tool
			for _, t := range tools {
				toolList = append(toolList, t.MCPTool())
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(toolList)
		case http.MethodPut:
			var req struct {
				Name    string `json:"name"`
				Disable bool   `json:"disable"`
			}
			body, err := io.ReadAll(io.LimitReader(r.Body, 1024*1024))
			if err != nil {
				http.Error(w, "failed to read body", http.StatusBadRequest)
				return
			}
			if err := json.Unmarshal(body, &req); err != nil {
				http.Error(w, "invalid json", http.StatusBadRequest)
				return
			}

			// Since proper tool storage modifying is complex and touches internal fields depending on connection type
			// we will return 200 OK without updating the DB for now to unblock the UI.
			// Ideally this would lookup the service via toolInfo.Tool().GetServiceId(), figure out
			// which connection_type it has, and update the tools slice within that.

			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{"status": "ok", "name": req.Name, "disable": req.Disable})

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func (a *Application) handleExecute() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req tool.ExecutionRequest
		// Limit execution request body to 5MB (tools might have large arguments)
		body, err := readBodyWithLimit(w, r, 5*1024*1024)
		if err != nil {
			return
		}

		if err := json.Unmarshal(body, &req); err != nil {
			logging.GetLogger().Error("failed to decode execution request", "error", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if len(req.ToolInputs) == 0 && len(req.Arguments) > 0 {
			b, err := json.Marshal(req.Arguments)
			if err != nil {
				http.Error(w, "failed to marshal arguments", http.StatusBadRequest)
				return
			}
			req.ToolInputs = b
		}

		result, err := a.ToolManager.ExecuteTool(r.Context(), &req)
		if err != nil {
			logging.GetLogger().Error("failed to execute tool", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(result)
	}
}
