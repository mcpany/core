
const BACKEND_URL = process.env.BACKEND_URL || 'http://localhost:50050';
const API_KEY = process.env.MCPANY_API_KEY || 'test-token';
const ECHO_SERVER_URL = process.env.ECHO_SERVER_URL || 'http://localhost:5678';

async function seedData() {
    console.log(`Seeding data to ${BACKEND_URL} using Key: ${API_KEY}...`);

    const headers = {
        'X-API-Key': API_KEY,
        'Content-Type': 'application/json'
    };

    const post = async (path: string, body: any) => {
        const res = await fetch(`${BACKEND_URL}${path}`, {
            method: 'POST',
            headers,
            body: JSON.stringify(body)
        });
        if (!res.ok) {
            const txt = await res.text();
            throw new Error(`Failed to POST ${path}: ${res.status} ${txt}`);
        }
        return res;
    };

    // 1. Seed Secrets
    console.log('Seeding secrets...');
    try {
        await post('/api/v1/secrets', {
            id: 'secret-api-key',
            name: 'API_KEY',
            key: 'API_KEY',
            value: 'sk-test-1234567890',
            provider: 'openai'
        });
    } catch (e) {
        console.warn('Error seeding secrets (might already exist):', e);
    }

    // 2. Seed Services
    console.log('Seeding services...');

    // Postgres Primary (Healthy)
    try {
        await post('/api/v1/services', {
            id: 'postgres-primary',
            name: 'Primary DB',
            http_service: {
                base_url: ECHO_SERVER_URL
            },
            version: '1.0.0',
            tags: ['db', 'production']
        });
    } catch (e) { console.warn('Error seeding postgres-primary:', e); }

    // OpenAI Gateway (Healthy)
    try {
        await post('/api/v1/services', {
            id: 'openai-gateway',
            name: 'OpenAI Gateway',
            http_service: {
                base_url: ECHO_SERVER_URL
            },
            version: '2.1.0',
            tags: ['ai', 'gateway']
        });
    } catch (e) { console.warn('Error seeding openai-gateway:', e); }

    // Broken Service (Unhealthy)
    try {
        await post('/api/v1/services', {
            id: 'broken-service',
            name: 'Legacy API',
            http_service: {
                base_url: 'http://non-existent-service:1234'
            },
            version: '1.0.0',
            tags: ['legacy']
        });
    } catch (e) { console.warn('Error seeding broken-service:', e); }

    // 3. Seed Traffic
    console.log('Seeding traffic stats...');
    const trafficPoints = Array.from({length: 24}, (_, i) => ({
        timestamp: new Date(Date.now() - i * 3600000).toISOString(),
        requests: Math.floor(Math.random() * 500) + 100,
        errors: Math.floor(Math.random() * 10)
    })).reverse();

    try {
        await post('/api/v1/debug/seed_traffic', trafficPoints);
    } catch (e) { console.warn('Error seeding traffic:', e); }

    // 4. Generate Traces / Logs via Tool Execution
    console.log('Generating traces...');
    try {
        // execute get_weather tool from weather-service (defined in config.minimal.yaml)
        // We assume weather-service is present (config.minimal.yaml is loaded)
        for (let i = 0; i < 5; i++) {
            await post('/api/v1/execute', {
                toolName: 'get_weather',
                arguments: { weather: `sunny-${i}` }
            });
        }
    } catch (e) { console.warn('Error generating traces:', e); }

    console.log('Seeding complete.');
}

export default seedData;
