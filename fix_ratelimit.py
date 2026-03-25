import os
import re

for root, _, files in os.walk("server/pkg/middleware"):
    for file in files:
        if file.endswith(".go"):
            filename = os.path.join(root, file)
            with open(filename, 'r') as f:
                content = f.read()

            types_to_fix = [
                "MockToolForCost", "MockTool"
            ]

            modified = False
            for t in types_to_fix:
                if t in content and f"func (t *{t}) IsStreaming() bool" not in content and f"func (m *{t}) IsStreaming() bool" not in content and f"func (c *{t}) IsStreaming() bool" not in content:
                    match = re.search(r"func \(([a-zA-Z0-9_]+) \*" + t + r"\) Execute\(", content)
                    if match:
                        receiver = match.group(1)
                        replacement = f"""
func ({receiver} *{t}) IsStreaming() bool {{
	return false
}}

func ({receiver} *{t}) StreamExecute(ctx context.Context, req *tool.ExecutionRequest) (<-chan any, error) {{
	return nil, nil
}}
"""
                        content = content.replace(match.group(0), replacement + "\n" + match.group(0))
                        modified = True

            if modified:
                with open(filename, 'w') as f:
                    f.write(content)
