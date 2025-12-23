/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


import {
  UpstreamServiceConfig
} from "@/proto/config/v1/upstream_service";
import { Tool } from "@/proto/config/v1/tool";

// Mock Services
export const mockServices: UpstreamServiceConfig[] = [
  {
    id: "svc_1",
    name: "Payment Gateway",
    version: "v1.2.0",
    disable: false,
    priority: 1,
    connectionPool: {
        maxConnections: 100,
        maxIdleConnections: 10,
        idleTimeout: { seconds: 30, nanos: 0 }
    },
    serviceConfig: {
        case: "httpService",
        value: {
            address: "https://api.stripe.com",
            tools: [],
            calls: {},
            healthCheck: undefined,
            tlsConfig: undefined,
            resources: [],
            prompts: []
        }
    },
    sanitizedName: "payment-gateway",
    upstreamAuthentication: undefined,
    cache: undefined,
    rateLimit: undefined,
    loadBalancingStrategy: 0,
    resilience: undefined,
    authentication: undefined,
    callPolicies: [],
    preCallHooks: [],
    postCallHooks: [],
    profiles: [],
    prompts: [],
    toolExportPolicy: undefined,
    promptExportPolicy: undefined,
    resourceExportPolicy: undefined,
    autoDiscoverTool: false
  },
  {
    id: "svc_2",
    name: "User Service",
    version: "v2.1.0",
    disable: false,
    priority: 2,
    serviceConfig: {
        case: "grpcService",
        value: {
            address: "localhost:50051",
             useReflection: true,
             tlsConfig: undefined,
             tools: [],
             healthCheck: undefined,
             protoDefinitions: [],
             protoCollection: [],
             resources: [],
             calls: {},
             prompts: []
        }
    },
    sanitizedName: "user-service",
    upstreamAuthentication: undefined,
    cache: undefined,
    rateLimit: undefined,
    loadBalancingStrategy: 0,
    resilience: undefined,
    authentication: undefined,
    callPolicies: [],
    preCallHooks: [],
    postCallHooks: [],
    profiles: [],
    prompts: [],
    toolExportPolicy: undefined,
    promptExportPolicy: undefined,
    resourceExportPolicy: undefined,
    autoDiscoverTool: false
  },
  {
    id: "svc_3",
    name: "Search Indexer",
    version: "v0.9.0",
    disable: true,
    priority: 5,
    serviceConfig: {
        case: "mcpService",
        value: {
            connectionType: {
                case: "stdioConnection",
                value: {
                    command: "python",
                    args: ["indexer.py"],
                    workingDirectory: "",
                    containerImage: "",
                    setupCommands: [],
                    env: {}
                }
            },
            toolAutoDiscovery: true,
            tools: [],
            resources: [],
            calls: {},
            prompts: []
        }
    },
    sanitizedName: "search-indexer",
    upstreamAuthentication: undefined,
    cache: undefined,
    rateLimit: undefined,
    loadBalancingStrategy: 0,
    resilience: undefined,
    authentication: undefined,
    callPolicies: [],
    preCallHooks: [],
    postCallHooks: [],
    profiles: [],
    prompts: [],
    toolExportPolicy: undefined,
    promptExportPolicy: undefined,
    resourceExportPolicy: undefined,
    autoDiscoverTool: false
  },
] as any[];


// Mock Metrics for Dashboard
export const mockMetrics = {
    totalRequests: 12450,
    activeServices: 12,
    errorRate: 0.05,
    avgLatency: 45, // ms
    history: [
        { time: "10:00", reqs: 400, errors: 2, latency: 40 },
        { time: "10:05", reqs: 450, errors: 5, latency: 42 },
        { time: "10:10", reqs: 800, errors: 12, latency: 55 },
        { time: "10:15", reqs: 600, errors: 4, latency: 48 },
        { time: "10:20", reqs: 500, errors: 3, latency: 44 },
        { time: "10:25", reqs: 700, errors: 8, latency: 50 },
    ]
};

export const mockTools: Tool[] = [
    {
        name: "get_weather",
        description: "Get current weather for a location",
        parameters: JSON.stringify({
            type: "object",
            properties: {
                location: { type: "string" }
            }
        })
    },
    {
        name: "send_email",
        description: "Send an email to a user",
        parameters: JSON.stringify({
            type: "object",
            properties: {
                to: { type: "string" },
                subject: { type: "string" },
                body: { type: "string" }
            }
        })
    }
] as any[];
