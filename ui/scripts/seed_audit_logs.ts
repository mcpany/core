
const BASE_URL = process.env.MCP_API_URL || 'http://localhost:50050';
// Default to 'test-token' if not set, matching CI config
const API_KEY = process.env.MCPANY_API_KEY || 'test-token';

async function seed() {
    console.log(`Seeding audit logs to ${BASE_URL}...`);

    const headers = {
        'Content-Type': 'application/json',
        'X-API-Key': API_KEY
    };

    // 1. Generate a generic error (Tool not found)
    try {
        console.log("Generating 'Tool Not Found' error log...");
        await fetch(`${BASE_URL}/api/v1/execute`, {
            method: 'POST',
            headers: headers,
            body: JSON.stringify({
                name: 'non_existent_tool',
                arguments: { foo: 'bar' }
            })
        });
    } catch (e) {
        console.log("Error generated (expected):", e);
    }

    // 2. Try to find a valid tool and execute it
    try {
        console.log("Fetching tools...");
        const toolsRes = await fetch(`${BASE_URL}/api/v1/tools`, { headers: headers });
        if (toolsRes.ok) {
            const data = await toolsRes.json();
            const tools = Array.isArray(data) ? data : (data.tools || []);

            // Try to use a tool found in list
            if (tools.length > 0) {
                const tool = tools[0];
                console.log(`Found tool: ${tool.name}. Executing...`);
                 await fetch(`${BASE_URL}/api/v1/execute`, {
                    method: 'POST',
                    headers: headers,
                    body: JSON.stringify({
                        name: tool.name,
                        arguments: {}
                    })
                });
            } else {
                console.log("No tools returned in list.");
            }
        } else {
            console.log("Failed to fetch tools:", toolsRes.status, toolsRes.statusText);
        }
    } catch (e) {
        console.log("Failed to list/execute tools:", e);
    }

    // 3. Hardcoded attempt for weather-service (known to exist from logs)
    try {
        console.log("Attempting hardcoded weather-service.get_weather...");
        await fetch(`${BASE_URL}/api/v1/execute`, {
            method: 'POST',
            headers: headers,
            body: JSON.stringify({
                name: 'weather-service.get_weather',
                arguments: {}
            })
        });
    } catch (e) {
        console.log("Hardcoded attempt failed:", e);
    }

    console.log("Seeding complete.");
}

seed().catch(console.error);
