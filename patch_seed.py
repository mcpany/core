import sys

with open("server/pkg/app/api_traces.go", "r") as f:
    content = f.read()

search = """	rootArgs, _ := json.Marshal(map[string]any{
		"query":   "Analyze Q3 financial report",
		"context": "user-session-123",
	})
	child1Args, _ := json.Marshal(map[string]any{
		"query": "Q3 2024 financials",
	})
	child2Args, _ := json.Marshal(map[string]any{
		"files": []string{"data_q3.xlsx"},
	})"""

replace = """	rootArgs, _ := json.Marshal(map[string]any{
		"query":   "Analyze Q3 financial report",
		"context": map[string]any{
            "session_id": "user-session-123",
            "flags": []string{"fast", "experimental"},
            "settings": map[string]any{
                "timeout_ms": 5000,
                "retry": true,
                "max_retries": 3,
                "null_val": nil,
            },
        },
	})
	child1Args, _ := json.Marshal(map[string]any{
		"query": "Q3 2024 financials",
	})
	child2Args, _ := json.Marshal(map[string]any{
		"files": []string{"data_q3.xlsx"},
	})"""

search2 = """			Result: map[string]any{
				"summary":    "Revenue up 15%",
				"confidence": 0.98,
			},"""

replace2 = """			Result: map[string]any{
				"summary":    "Revenue up 15%",
				"confidence": 0.98,
                "metadata": map[string]any{
                    "processed_at": now.Format(time.RFC3339),
                    "sources": []map[string]any{
                        {"id": "src-1", "type": "pdf", "pages": 15},
                        {"id": "src-2", "type": "database", "rows_scanned": 10500},
                    },
                    "tags": []string{"finance", "q3", "internal"},
                },
			},"""

content = content.replace(search, replace)
content = content.replace(search2, replace2)

with open("server/pkg/app/api_traces.go", "w") as f:
    f.write(content)
