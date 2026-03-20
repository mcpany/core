import { request, APIRequestContext } from '@playwright/test';

const BASE_URL = process.env.BACKEND_URL || 'http://localhost:50050';
const API_KEY = process.env.MCPANY_API_KEY || 'test-token';
const HEADERS = { 'X-API-Key': API_KEY, 'Content-Type': 'application/json' };

export const seedAuditLogs = async (requestContext?: APIRequestContext) => {
    const context = requestContext || await request.newContext({ baseURL: BASE_URL });

    const auditData = {
        entries: [
            {
                timestamp: new Date().toISOString(),
                tool_name: "test_tool_1",
                user_id: "test_user",
                profile_id: "default",
                arguments: JSON.stringify({ test_argument: "test_value", count: 42 }),
                result: JSON.stringify({ success: true, nested: { key: "value" } }),
                error: "",
                duration: "15ms",
                duration_ms: 15
            }
        ]
    };

    try {
        const res = await context.post('/api/v1/debug/seed_audit', { data: auditData, headers: HEADERS });
        if (!res.ok() && res.status() !== 404) {
            console.log(`Failed to seed audit logs: ${res.status()}`);
        }
    } catch (e) {
        console.log(`Failed to seed audit logs: ${e}`);
    }
};
