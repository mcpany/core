with open('ui/tests/inspector.spec.ts', 'r') as f:
    content = f.read()

content = content.replace('if (wsSend) (wsSend as any)(JSON.stringify(MOCK_TRACE));', '(wsSend as any)(JSON.stringify(MOCK_TRACE));')

with open('ui/tests/inspector.spec.ts', 'w') as f:
    f.write(content)
