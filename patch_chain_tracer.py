import re

with open("ui/src/components/dashboard/agent-chain-tracer.tsx", "r") as f:
    content = f.read()

content = content.replace(
    '<CardTitle className="text-xl font-semibold flex items-center gap-2">',
    '<CardTitle className="text-sm font-medium flex items-center gap-2">'
)

with open("ui/src/components/dashboard/agent-chain-tracer.tsx", "w") as f:
    f.write(content)
