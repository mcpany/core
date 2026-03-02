const { request } = require('@playwright/test');
(async () => {
    const ctx = await request.newContext({ baseURL: 'http://localhost:50050' });
    const res = await ctx.post('/api/v1/execute', {
        data: {
            tool_name: "weather-service.get_weather",
            arguments: {location: "San Francisco"}
        },
        headers: {
            'X-API-Key': 'dev-key'
        }
    });
    console.log(await res.json());
})();
