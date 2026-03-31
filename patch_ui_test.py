import re

with open("ui/src/app/universal-agent-bus/page.tsx", "r") as f:
    content = f.read()

# Revert previous hack
content = content.replace(
    '<div className="hidden"><span className="text-sm font-medium">Agent Chain Tracer (A2A)</span></div>\n        <AgentChainTracer />',
    '<AgentChainTracer />'
)

with open("ui/src/app/universal-agent-bus/page.tsx", "w") as f:
    f.write(content)
