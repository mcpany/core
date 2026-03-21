import re

with open('ui/tests/inspector.spec.ts', 'r') as f:
    content = f.read()

content = content.replace("wsSend(JSON.stringify(MOCK_TRACE));", "if (wsSend) (wsSend as unknown as Function)(JSON.stringify(MOCK_TRACE));")

with open('ui/tests/inspector.spec.ts', 'w') as f:
    f.write(content)
