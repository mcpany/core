/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

// Global mock database for E2E tests

// Global mock database for E2E tests

const globalForMockDB = global as unknown as { mockDB: any };

export const MockDB = globalForMockDB.mockDB || {
    services: [
        { name: "weather-service", version: "1.2.0", disable: false, httpService: { address: "http://weather:8080" }, id: "srv-1" },
        { name: "memory-store", version: "0.9.5", disable: true, grpcService: { address: "memory:9090" }, id: "srv-2" },
        { name: "local-files", version: "1.0.0", disable: false, commandLineService: { command: "npx", args: ["-y", "@modelcontextprotocol/server-filesystem", "/users/me/docs"] }, id: "srv-3" },
    ],
    tools: [
        { name: "get_weather", description: "Get current weather for a location", enabled: true, serviceId: "weather-service", input_schema: { type: "object", properties: { location: { type: "string" } } } },
        { name: "read_file", description: "Read file from filesystem", enabled: true, serviceId: "local-files", input_schema: { type: "object", properties: { path: { type: "string" } } } },
        { name: "list_directory", description: "List directory contents", enabled: false, serviceId: "local-files", input_schema: { type: "object", properties: { path: { type: "string" } } } },
        { name: "search_memory", description: "Search vector memory", enabled: true, serviceId: "memory-store", input_schema: { type: "object", properties: { query: { type: "string" } } } },
    ],
    settings: {
        mcp_listen_address: ":8080",
        log_level: 1, // INFO
        log_format: 1, // Text
        audit: { enabled: true },
        dlp: { enabled: false },
        gc_settings: { interval: "1h" },
        api_key: "****************",
        profiles: ["default", "dev"],
        allowed_ips: ["127.0.0.1", "10.0.0.0/8"]
    }
};

if (process.env.NODE_ENV !== 'production') globalForMockDB.mockDB = MockDB;
