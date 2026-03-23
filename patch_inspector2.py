import re
with open("ui/tests/inspector.spec.ts", "r") as f:
    content = f.read()

replacement = """
    let wsSend: ((data: string) => void) | null = null;
    await page.routeWebSocket('**/api/v1/ws/traces', (ws: any) => {
      wsSend = (data: string) => ws.send(data);
    });
"""

# replace it back so we don't break types
new_content = re.sub(
    r"    let wsSend: \(\(data: string\) => void\) \| null = null;\n    await page\.routeWebSocket\('\*\*/api/v1/ws/traces', \(ws: any\) => \{\n      ws\.onMessage\(\(message: any\) => \{\n        // Handle incoming messages if needed\n      \}\);\n      wsSend = \(data: string\) => ws\.send\(data\);\n    \}\);",
    replacement.strip(),
    content
)

with open("ui/tests/inspector.spec.ts", "w") as f:
    f.write(new_content)
