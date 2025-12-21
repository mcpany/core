/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


import { UpstreamServiceConfig } from "@/proto/config/v1/upstream_service";

// Mock data generator
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
        idleTimeout: { seconds: 30 }
    },
    serviceConfig: {
        case: "httpService",
        value: {
            address: "https://api.stripe.com"
        }
    }
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
            address: "localhost:50051"
        }
    }
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
                    args: ["indexer.py"]
                }
            }
        }
    }
  },
] as any[]; // Type assertion to bypass strict proto typing for mock if needed, or adjust types
