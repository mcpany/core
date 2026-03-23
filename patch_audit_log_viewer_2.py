import re
with open("ui/src/components/audit/audit-log-viewer.tsx", "r") as f:
    content = f.read()

# Make sure traceId is added to the interface
if "traceId?: string;" not in content and "traceId: string;" not in content:
    content = content.replace("durationMs: number;", "durationMs: number;\n    traceId?: string;")

with open("ui/src/components/audit/audit-log-viewer.tsx", "w") as f:
    f.write(content)
