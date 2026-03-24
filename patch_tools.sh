cat << 'INNER_EOF' > /tmp/api_patch.diff
<<<<<<< SEARCH
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
=======
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

			// We must find the parent service to save the tool since tools are nested inside services
			store := a.storage
			services, err := store.ListServices(r.Context())
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			var parentService *configv1.ServiceDefinition
			for _, srv := range services {
				for _, t := range srv.GetTools() {
					if t.GetName() == req.Name {
						parentService = srv
						break
					}
				}
				if parentService != nil {
					break
				}
			}

			if parentService == nil {
				http.NotFound(w, r)
				return
			}

			// Create/update tool override
			if parentService.ToolOverrides == nil {
				parentService.ToolOverrides = make(map[string]*configv1.ToolOverride)
			}
			if parentService.ToolOverrides[req.Name] == nil {
				parentService.ToolOverrides[req.Name] = &configv1.ToolOverride{}
			}
			parentService.ToolOverrides[req.Name].Disabled = req.Disable

			// Save the updated parent service
			if err := store.SaveService(r.Context(), parentService); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			// Signal a configuration reload
			a.configCh <- &configv1.ConfigChangedEvent{Type: configv1.ConfigChangedEvent_UPDATED, Component: configv1.ConfigChangedEvent_SERVICE, Name: parentService.GetName()}

			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{"status": "ok", "name": req.Name, "disable": req.Disable})

		default:
>>>>>>> REPLACE
INNER_EOF
python3 -c '
import sys
with open("/tmp/api_patch.diff", "r") as f:
    diff_text = f.read()
import re
search = re.search(r"<<<<<<< SEARCH\n(.*?)\n=======\n(.*?)\n>>>>>>> REPLACE", diff_text, re.DOTALL)
if search:
    with open("server/pkg/app/api.go", "r") as f:
        content = f.read()
    content = content.replace(search.group(1), search.group(2))
    with open("server/pkg/app/api.go", "w") as f:
        f.write(content)
'
