import re

with open('ui/tests/inspector.spec.ts', 'r') as f:
    content = f.read()

content = content.replace("wsSend = (data: string) => ws.send(data);", "wsSend = ws.send.bind(ws);")
content = content.replace("let wsSend: ((data: string) => void) | null = null;", "let wsSend: any = null;")
content = content.replace("if (wsSend) {\\n      wsSend(JSON.stringify(MOCK_TRACE));\\n    }", "if (wsSend) (wsSend as any)(JSON.stringify(MOCK_TRACE));")

with open('ui/tests/inspector.spec.ts', 'w') as f:
    f.write(content)
