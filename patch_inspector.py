import re

with open('ui/tests/inspector.spec.ts', 'r') as f:
    content = f.read()

content = content.replace("wsSend = (data: string) => ws.send(data);", "wsSend = (ws.send.bind(ws) as any);")

with open('ui/tests/inspector.spec.ts', 'w') as f:
    f.write(content)
