import json

with open('server/pkg/app/api_traces.go', 'r') as f:
    content = f.read()

# Replace the data analyzer mock
old_data_analyzer = """		{
			Timestamp: now.Add(500 * time.Millisecond),
			ToolName:  "data-analyzer",
			UserID:    "system",
			ProfileID: "default",
			TraceID:   traceID,
			SpanID:    traceID + "-2",
			ParentID:  traceID + "-0",
			Arguments: json.RawMessage(child2Args),
			Result: map[string]any{
				"analysis": "Growth detected",
				"metrics": map[string]any{
					"revenue": 1.15,
				},
			},
			Duration:   "700ms",
			DurationMs: 700,
		},"""

new_data_analyzer = """		{
			Timestamp: now.Add(500 * time.Millisecond),
			ToolName:  "data-analyzer",
			UserID:    "system",
			ProfileID: "default",
			TraceID:   traceID,
			SpanID:    traceID + "-2",
			ParentID:  traceID + "-0",
			Arguments: json.RawMessage(child2Args),
			Result: map[string]any{
				"analysis": "Growth detected across key sectors.",
				"metrics": map[string]any{
					"revenue_growth": 0.15,
					"user_acquisition": 12500,
					"churn_rate": 0.02,
				},
			},
			Duration:   "700ms",
			DurationMs: 700,
		},"""

# Replace the code refactor mock
old_code_refactor = """		{
			Timestamp: now.Add(1200 * time.Millisecond),
			ToolName:  "code-refactor",
			UserID:    "system",
			ProfileID: "default",
			TraceID:   traceID,
			SpanID:    traceID + "-3",
			ParentID:  traceID + "-0",
			Arguments: json.RawMessage(`{"file": "main.py", "action": "optimize"}`),
			Result: map[string]any{
				"diff":   "--- a/main.py\\n+++ b/main.py\\n@@ -1,5 +1,5 @@\\n-def slow_func():\\n-    pass\\n+def fast_func():\\n+    return True\\n",
				"status": "success",
			},
			Duration:   "150ms",
			DurationMs: 150,
		},"""

new_code_refactor = """		{
			Timestamp: now.Add(1200 * time.Millisecond),
			ToolName:  "code-refactor",
			UserID:    "system",
			ProfileID: "default",
			TraceID:   traceID,
			SpanID:    traceID + "-3",
			ParentID:  traceID + "-0",
			Arguments: json.RawMessage(`{"file": "src/utils/math.py", "action": "optimize"}`),
			Result: map[string]any{
				"diff":   "@@ -10,13 +10,6 @@\\n def calculate_growth(revenue, baseline):\\n-    if baseline == 0:\\n-        return 0\\n-    growth = (revenue - baseline) / baseline\\n-    return growth\\n+    return 0 if baseline == 0 else (revenue - baseline) / baseline\\n \\n def calculate_churn(lost_users, total_users):\\n",
				"status": "success",
			},
			Duration:   "150ms",
			DurationMs: 150,
		},"""

# Replace the db query mock
old_db_query = """		{
			Timestamp:  now.Add(1350 * time.Millisecond),
			ToolName:   "database-query",
			UserID:     "system",
			ProfileID:  "default",
			TraceID:    traceID,
			SpanID:     traceID + "-4",
			ParentID:   traceID + "-0",
			Arguments:  json.RawMessage(`{"query": "SELECT * FROM users WHERE active = 1"}`),
			Error:      "Timeout: Query exceeded 5000ms limit",
			Duration:   "5005ms",
			DurationMs: 5005,
		},"""

new_db_query = """		{
			Timestamp:  now.Add(1350 * time.Millisecond),
			ToolName:   "database-query",
			UserID:     "system",
			ProfileID:  "default",
			TraceID:    traceID,
			SpanID:     traceID + "-4",
			ParentID:   traceID + "-0",
			Arguments:  json.RawMessage(`{"query": "SELECT id, email, last_login FROM users WHERE active = 1 AND region = 'US-West' ORDER BY last_login DESC LIMIT 1000"}`),
			Error:      "TimeoutException: Query exceeded 5000ms execution limit. The database cluster may be under heavy load or the query requires missing indices on 'active' and 'region' columns.",
			Duration:   "5005ms",
			DurationMs: 5005,
		},"""

content = content.replace(old_data_analyzer, new_data_analyzer)
content = content.replace(old_code_refactor, new_code_refactor)
content = content.replace(old_db_query, new_db_query)

with open('server/pkg/app/api_traces.go', 'w') as f:
    f.write(content)
