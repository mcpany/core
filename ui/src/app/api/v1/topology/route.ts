/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { NextResponse } from 'next/server';
import { Graph, Node } from '@/types/topology';

export async function GET() {
  const coreNode: Node = {
    id: "core-1",
    label: "MCP Any Core",
    type: "NODE_TYPE_CORE",
    status: "NODE_STATUS_ACTIVE",
    metrics: { qps: 120, latencyMs: 5, errorRate: 0.001 },
    children: [
        {
            id: "middleware-pipeline",
            label: "Middleware Pipeline",
            type: "NODE_TYPE_MIDDLEWARE",
            status: "NODE_STATUS_ACTIVE",
            children: [
                {
                    id: "mw-auth",
                    label: "Authentication",
                    type: "NODE_TYPE_MIDDLEWARE",
                    status: "NODE_STATUS_ACTIVE",
                },
                {
                    id: "mw-log",
                    label: "Logging",
                    type: "NODE_TYPE_MIDDLEWARE",
                    status: "NODE_STATUS_ACTIVE",
                }
            ]
        },
        {
            id: "srv-1",
            label: "weather-service",
            type: "NODE_TYPE_SERVICE",
            status: "NODE_STATUS_ACTIVE",
            metrics: { qps: 45, latencyMs: 120, errorRate: 0.01 },
            children: [
                {
                    id: "tool-get-weather",
                    label: "get_weather",
                    type: "NODE_TYPE_TOOL",
                    status: "NODE_STATUS_ACTIVE",
                    children: [
                        {
                            id: "api-get-weather",
                            label: "GET /api.weather.gov",
                            type: "NODE_TYPE_API_CALL",
                            status: "NODE_STATUS_ACTIVE",
                        }
                    ]
                },
                {
                    id: "res-weather-sf",
                    label: "SF Weather",
                    type: "NODE_TYPE_RESOURCE",
                    status: "NODE_STATUS_ACTIVE",
                },
                {
                    id: "prompt-weather-report",
                    label: "weather_report",
                    type: "NODE_TYPE_PROMPT",
                    status: "NODE_STATUS_INACTIVE",
                }
            ]
        },
        {
            id: "srv-2",
            label: "memory-store",
            type: "NODE_TYPE_SERVICE",
            status: "NODE_STATUS_INACTIVE",
            metrics: { qps: 0, latencyMs: 0, errorRate: 0 },
            children: [
                 {
                    id: "tool-save-memory",
                    label: "save_memory",
                    type: "NODE_TYPE_TOOL",
                    status: "NODE_STATUS_INACTIVE",
                }
            ]
        },
        {
            id: "srv-3",
            label: "local-files",
            type: "NODE_TYPE_SERVICE",
            status: "NODE_STATUS_ACTIVE",
            metrics: { qps: 12, latencyMs: 25, errorRate: 0.0 },
             children: [
                 {
                    id: "tool-list-files",
                    label: "list_files",
                    type: "NODE_TYPE_TOOL",
                    status: "NODE_STATUS_ACTIVE",
                    children: [
                        {
                            id: "api-list-files",
                            label: "OS: ls -la",
                            type: "NODE_TYPE_API_CALL",
                            status: "NODE_STATUS_ACTIVE",
                        }
                    ]
                },
                {
                    id: "tool-read-file",
                    label: "read_file",
                    type: "NODE_TYPE_TOOL",
                    status: "NODE_STATUS_INACTIVE",
                },
                 {
                    id: "res-notes",
                    label: "notes.txt",
                    type: "NODE_TYPE_RESOURCE",
                    status: "NODE_STATUS_ACTIVE",
                }
            ]
        },
        {
            id: "srv-4",
            label: "slack-integration",
            type: "NODE_TYPE_SERVICE",
            status: "NODE_STATUS_ERROR",
            metrics: { qps: 2, latencyMs: 5000, errorRate: 0.8 },
            children: [
                {
                    id: "tool-send-msg",
                    label: "send_message",
                    type: "NODE_TYPE_TOOL",
                    status: "NODE_STATUS_ERROR",
                    children: [
                        {
                            id: "api-slack-post",
                            label: "POST /chat.postMessage",
                            type: "NODE_TYPE_API_CALL",
                            status: "NODE_STATUS_ERROR",
                        }
                    ]
                }
            ]
        },
        {
            id: "webhooks",
            label: "Webhooks",
            type: "NODE_TYPE_WEBHOOK",
            status: "NODE_STATUS_ACTIVE",
            children: [
                {
                    id: "wh-github",
                    label: "github-push",
                    type: "NODE_TYPE_WEBHOOK",
                    status: "NODE_STATUS_ACTIVE",
                },
                {
                    id: "wh-stripe",
                    label: "stripe-payment",
                    type: "NODE_TYPE_WEBHOOK",
                    status: "NODE_STATUS_INACTIVE",
                }
            ]
        }
    ]
  };

  const clients: Node[] = [
      {
          id: "client-web",
          label: "Web Dashboard (Admin)",
          type: "NODE_TYPE_CLIENT",
          status: "NODE_STATUS_ACTIVE",
          metadata: { ip: "192.168.1.1", userAgent: "Chrome 122" }
      },
      {
          id: "client-cli",
          label: "MCP CLI",
          type: "NODE_TYPE_CLIENT",
          status: "NODE_STATUS_ACTIVE",
           metadata: { ip: "10.0.0.5", userAgent: "mcp-cli/1.0" }
      },
       {
          id: "client-claude",
          label: "Claude Desktop",
          type: "NODE_TYPE_CLIENT",
          status: "NODE_STATUS_ACTIVE",
           metadata: { ip: "127.0.0.1", userAgent: "Claude/3.5" }
      }
  ];

  const graph: Graph = {
      core: coreNode,
      clients: clients
  };

  return NextResponse.json(graph);
}
