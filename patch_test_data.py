import re

with open("ui/tests/e2e/test-data.ts", "r") as f:
    content = f.read()

# Make sure we import things we might need or just fix it if missing

export_seed = """
export const seedAuditLogs = async (requestContext?: APIRequestContext) => {
    const context = requestContext || await request.newContext({ baseURL: BASE_URL });
    try {
        // Create an audit log by calling a tool
        const res = await context.post('/mcp/jsonrpc', {
            data: {
                jsonrpc: "2.0",
                id: 1,
                method: "tools/call",
                params: {
                    name: "echo_tool",
                    arguments: { text: "hello world" }
                }
            },
            headers: HEADERS,
        });

        // Let's also do a second one
        await context.post('/mcp/jsonrpc', {
            data: {
                jsonrpc: "2.0",
                id: 2,
                method: "tools/call",
                params: {
                    name: "calculator_add",
                    arguments: { a: 5, b: 3 }
                }
            },
            headers: HEADERS,
        });
    } catch (e) {
        console.log(`seedAuditLogs error: ${e}`);
    }
};
"""

if "export const seedAuditLogs" not in content:
    content += "\n" + export_seed
    with open("ui/tests/e2e/test-data.ts", "w") as f:
        f.write(content)
