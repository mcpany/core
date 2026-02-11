/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { apiClient, UpstreamServiceConfig } from "@/lib/client";

/**
 * Registers a dummy service with tools for testing purposes.
 * @param serviceName The name of the service to register (default: "math-service")
 */
export async function seedMathService(serviceName = "math-service") {
  const config: UpstreamServiceConfig = {
    id: serviceName,
    name: serviceName,
    version: "1.0.0",
    disable: false,
    priority: 0,
    loadBalancingStrategy: 0,
    sanitizedName: serviceName,
    readOnly: false,
    callPolicies: [],
    preCallHooks: [],
    postCallHooks: [],
    prompts: [],
    autoDiscoverTool: true,
    configError: "",
    tags: ["test", "math"],
    // Use a mock command or existing one.
    // Since we can't easily spin up a real python server in E2E without heavy setup,
    // we might register a service that points to a known internal endpoint or just mocks the response.
    // However, for the purpose of "listTools", we just need the backend to THINK there are tools.
    // The current backend likely discovers tools by querying the service.
    // If we use "mcpService" (proxy), we need a running server.
    // If we use "httpService", we need a running HTTP server.
    //
    // HACK: Use `commandLineService` with a simple echo command that outputs the listTools JSON?
    // Or, if the backend supports "Static Tool Definitions" in config (some implementations do)?
    // Looking at `UpstreamServiceConfig` structure in `client.ts`:
    // It has `httpService`, `grpcService`, `commandLineService`, `mcpService`, `openapiService`.
    // And `openapiService` has `tools` field!
    //
    // Using `openapiService` allows defining tools statically via the spec or explicit config.
    // Let's try to register an OpenAPI service which is easier to mock or just define tools.
    // Actually, `UpstreamServiceConfig` definition in `client.ts`:
    /*
        openapiService: {
            // ...
            tools: config.openapiService.tools,
            // ...
        }
    */
    // If we pass `tools` array here, maybe the backend registers them?
    // Let's assume yes.
    openapiService: {
        address: "http://localhost:9999", // Dummy
        specUrl: "",
        tools: [
            {
                name: "add",
                description: "Adds two numbers",
                inputSchema: {
                    type: "object",
                    properties: {
                        a: { type: "number" },
                        b: { type: "number" }
                    },
                    required: ["a", "b"]
                }
            },
            {
                name: "subtract",
                description: "Subtracts two numbers",
                inputSchema: {
                    type: "object",
                    properties: {
                        a: { type: "number" },
                        b: { type: "number" }
                    },
                    required: ["a", "b"]
                }
            }
        ],
        resources: [],
        prompts: [],
        calls: {}
    }
  };

  try {
    await apiClient.registerService(config);
    console.log(`Seeded service: ${serviceName}`);
  } catch (e) {
    console.error(`Failed to seed service ${serviceName}:`, e);
    // If it exists, try update?
    try {
        await apiClient.updateService(config);
        console.log(`Updated seeded service: ${serviceName}`);
    } catch (e2) {
        console.error("Failed to update seed service", e2);
    }
  }
}
