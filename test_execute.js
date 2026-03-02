/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

const { request } = require('playwright');
(async () => {
    const ctx = await request.newContext({ baseURL: 'http://localhost:50050' });
    const res = await ctx.post('/api/v1/execute', {
        data: {
            server_name: "weather",
            tool_name: "get_weather",
            arguments: "{\"location\":\"San Francisco\"}"
        }
    });
    console.log(await res.json());
})();
