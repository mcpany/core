/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

// Global mock database for E2E tests

export const MockDB = {
    services: [
        { name: "weather-service", version: "1.2.0", disable: false, httpService: { address: "http://weather:8080" }, id: "srv-1" },
        { name: "memory-store", version: "0.9.5", disable: true, grpcService: { address: "memory:9090" }, id: "srv-2" },
        { name: "local-files", version: "1.0.0", disable: false, commandLineService: { command: "npx", args: ["-y", "@modelcontextprotocol/server-filesystem", "/users/me/docs"] }, id: "srv-3" },
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
