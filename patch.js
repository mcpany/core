const fs = require('fs');
const file = 'ui/tests/inspector.spec.ts';
let code = fs.readFileSync(file, 'utf8');

code = code.replace(/ws\.send\(data\)/g, "ws.send(data)");
code = code.replace(/ws: any/g, "ws: any");

// The issue is `routeWebSocket` is not present in all playwright versions or is being called incorrectly based on the TS types.
// We should check the playwright version or how it expects routeWebSocket to be called.
// Actually, `page.routeWebSocket` is available in recent playwright. The problem is `wsSend = (data: string) => ws.send(data);` might not be right because `ws` might be of a specific type. Wait, the error is:
// tests/inspector.spec.ts(71,7): error TS2349: This expression is not callable. Type 'never' has no call signatures.
// Line 71: wsSend(JSON.stringify(MOCK_TRACE));
// This is because `let wsSend: ((data: string) => void) | null = null;` and TS thinks it's still `null`.
// We need to use `wsSend!(...)` or `if (wsSend) { (wsSend as any)(...) }`. But the code has `if (wsSend) { wsSend(JSON.stringify(MOCK_TRACE)); }`.
// Let's replace `let wsSend: ((data: string) => void) | null = null;` with `let wsSend: any = null;`.
