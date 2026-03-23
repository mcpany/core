const fs = require('fs');
let file = fs.readFileSync('ui/tests/e2e/test-data.ts', 'utf-8');

const seedHealthFunc = `
export const seedHealth = async (requestContext?: APIRequestContext) => {
    const context = requestContext || await request.newContext({ baseURL: BASE_URL });
    const now = Date.now();

    // Create ~60 minutes of history, mostly healthy but with some degraded
    const history1 = [];
    const history2 = [];
    for (let i = 60; i >= 0; i--) {
        const ts = now - i * 60000;
        history1.push({ timestamp: ts, status: (i === 10 || i === 11) ? "error" : "healthy" });
        history2.push({ timestamp: ts, status: "healthy" });
    }

    const payload = {
        "svc_01": history1,
        "svc_02": history2
    };

    try {
        await context.post('/api/v1/debug/seed_health', { data: payload, headers: HEADERS });
    } catch (e) {
        console.log(\`Failed to seed health: \${e}\`);
    }
};
`;

file = file.replace('export const seedTraffic = async (requestContext?: APIRequestContext) => {', seedHealthFunc + '\nexport const seedTraffic = async (requestContext?: APIRequestContext) => {');

fs.writeFileSync('ui/tests/e2e/test-data.ts', file);
