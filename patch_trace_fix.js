const fs = require('fs');
const filepath = 'server/pkg/app/api_traces.go';
let content = fs.readFileSync(filepath, 'utf8');

// I replaced:
// Result: []any{ ... }
// I will replace it back to a map that contains an array, maybe that's what's expected?
// Or maybe the audit log schema defines it.
// The previous code had:
// Result: map[string]any{ ... }
// Let's just make it map[string]any{"data": []any{...}}

const oldResult = `			Result: []any{
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
const newResult = `			Result: map[string]any{
				"content": []any{
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
				},
			},`;
content = content.replace(oldResult, newResult);

fs.writeFileSync(filepath, content);
