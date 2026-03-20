const fs = require('fs');
const filepath = 'server/pkg/app/api_traces.go';
let content = fs.readFileSync(filepath, 'utf8');

const oldRootArgs = `	rootArgs, _ := json.Marshal(map[string]any{
		"query":   "Analyze Q3 financial report",
		"context": "user-session-123",
	})`;
const newRootArgs = `	rootArgs, _ := json.Marshal(map[string]any{
		"query":   "Analyze Q3 financial report",
		"context": "user-session-123",
		"options": map[string]any{
			"verbose": true,
			"format":  "markdown",
		},
	})`;

content = content.replace(oldRootArgs, newRootArgs);

const oldResult = `			Result: map[string]any{
				"summary":    "Revenue up 15%",
				"confidence": 0.98,
			},`;
const newResult = `			Result: []any{
				map[string]any{
					"type": "text",
					"text": "# Q3 Financial Report\\n\\nRevenue is **up 15%** quarter-over-quarter.\\n\\n## Key Metrics\\n- **Revenue:** $1.2M\\n- **Profit Margin:** 24%\\n- **Customer Acquisition:** +850 new users",
				},
				map[string]any{
					"month":   "July",
					"revenue": 350000,
					"target":  320000,
				},
				map[string]any{
					"month":   "August",
					"revenue": 400000,
					"target":  350000,
				},
				map[string]any{
					"month":   "September",
					"revenue": 450000,
					"target":  380000,
				},
			},`;
content = content.replace(oldResult, newResult);

fs.writeFileSync(filepath, content);
