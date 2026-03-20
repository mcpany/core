console.log(`
Ah! Wait.
When I ran git diff server/pkg/app/api_traces.go:
+			Result: []any{
+				map[string]any{
+					"type": "text",
+					"text": "# Q3 Financial Report\n\nRevenue is **up 15%** quarter-over-quarter.\n\n## Key Metrics\n- **Revenue:** $1.2M\n- **Profit Margin:** 24%\n- **Customer Acquisition:** +850 new users",
+				},
+				map[string]any{
+					"month":   "July",
+					"revenue": 350000,
+					"target":  320000,
+				},
+				map[string]any{
+					"month":   "August",
+					"revenue": 400000,
+					"target":  350000,
+				},
+				map[string]any{
+					"month":   "September",
+					"revenue": 450000,
+					"target":  380000,
+				},
+			},

Did I write a malformed Go file?
Let's build it.
`);
