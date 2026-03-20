import re

with open("ui/tests/e2e/test-data.ts", "r") as f:
    content = f.read()

new_content = content.replace(
    "console.log(`Failed to seed traffic: ${e}`);\n    }",
    """console.log(`Failed to seed traffic: ${e}`);
    }

    // Generate an explicit audit log entry by executing a tool
    try {
        await context.post('/api/v1/execute', {
            headers: HEADERS,
            data: {
                tool_name: "echo_tool",
                service_name: "Echo Service",
                arguments: { input: "test-audit-log-generation" }
            }
        });
    } catch (e) {
        console.log(`Failed to generate audit log: ${e}`);
    }"""
)

with open("ui/tests/e2e/test-data.ts", "w") as f:
    f.write(new_content)
