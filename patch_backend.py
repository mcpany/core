import re

with open('server/pkg/app/api.go', 'r') as f:
    content = f.read()

# I need to update the handleTools function to actually save the tool via the service configuration
replacement = """
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

			toolInfo, ok := a.ToolManager.GetTool(req.Name)
			if !ok {
				http.Error(w, "tool not found", http.StatusNotFound)
				return
			}

			serviceID := toolInfo.Tool().GetServiceId()
			if serviceID == "" {
				http.Error(w, "service ID not found for tool", http.StatusInternalServerError)
				return
			}

			// Save to Storage
			service, err := a.Storage.GetService(r.Context(), serviceID)
			if err != nil {
				http.Error(w, "failed to get service", http.StatusInternalServerError)
				return
			}
			if service == nil {
				http.Error(w, "service not found", http.StatusNotFound)
				return
			}

			// Look through tools and update the disable state, if it doesn't exist, we should add it as an override
			found := false
			for _, t := range service.GetTools() {
				if t.GetName() == req.Name {
					t.SetDisable(req.Disable)
					found = true
					break
				}
			}

			if !found {
				// If not found, add it
				newTool := toolInfo.Tool()
				newTool.SetDisable(req.Disable)
				service.SetTools(append(service.GetTools(), newTool))
			}

			if err := a.Storage.SaveService(r.Context(), service); err != nil {
				http.Error(w, "failed to save service", http.StatusInternalServerError)
				return
			}

			// Reload config
			if err := a.ReloadConfig(r.Context(), a.fs, a.configPaths); err != nil {
				logging.GetLogger().Error("failed to reload config after tool update", "error", err)
			}

			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{"status": "ok", "name": req.Name, "disable": req.Disable})
"""

# replace it in content using regex or string replace
pattern = r"""\s*case http\.MethodPut:
\s*var req struct {
\s*Name    string `json:"name"`
\s*Disable bool   `json:"disable"`
\s*}
\s*body, err := io\.ReadAll\(io\.LimitReader\(r\.Body, 1024\*1024\)\)
\s*if err != nil {
\s*http\.Error\(w, "failed to read body", http\.StatusBadRequest\)
\s*return
\s*}
\s*if err := json\.Unmarshal\(body, &req\); err != nil {
\s*http\.Error\(w, "invalid json", http\.StatusBadRequest\)
\s*return
\s*}
\s*w\.Header\(\)\.Set\("Content-Type", "application/json"\)
\s*_ = json\.NewEncoder\(w\)\.Encode\(map\[string\]any{"status": "ok", "name": req\.Name, "disable": req\.Disable}\)"""

content = re.sub(pattern, replacement, content, flags=re.MULTILINE | re.DOTALL)

with open('server/pkg/app/api.go', 'w') as f:
    f.write(content)
