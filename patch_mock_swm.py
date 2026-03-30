import re

filename = "server/pkg/app/api.go"
with open(filename, 'r') as f:
    content = f.read()

func_code = """
// handleMockSwarmTopology returns mock data for the swarm topology widget.
func (a *Application) handleMockSwarmTopology() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		mockData := map[string]interface{}{
			"nodes": []map[string]interface{}{
				{"id": "n1", "label": "Primary Orchestrator", "type": "validator", "status": "locked", "x": 50, "y": 50},
				{"id": "n2", "label": "Research Agent", "type": "agent", "status": "active", "x": 20, "y": 30},
				{"id": "n3", "label": "Tool Exec", "type": "service", "status": "idle", "x": 20, "y": 70},
				{"id": "n4", "label": "Synthesizer", "type": "agent", "status": "active", "x": 80, "y": 50},
				{"id": "n5", "label": "Rogue Node", "type": "agent", "status": "stall", "x": 80, "y": 20},
			},
			"edges": []map[string]interface{}{
				{"source": "n2", "target": "n1", "status": "healthy", "hash": "0x1A4"},
				{"source": "n1", "target": "n3", "status": "healthy", "hash": "0x2B9"},
				{"source": "n1", "target": "n4", "status": "healthy", "hash": "0x3C1"},
				{"source": "n5", "target": "n1", "status": "blocked", "hash": "INVALID_GRAFT"},
			},
			"anomalies": []string{"ARI Hub: Logic Graft Blocked from Rogue Node (n5)"},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockData)
	}
}
"""

if "func (a *Application) handleMockSwarmTopology" not in content:
    content += "\n" + func_code

with open(filename, 'w') as f:
    f.write(content)
