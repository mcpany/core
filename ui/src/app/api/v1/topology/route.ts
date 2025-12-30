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
        }
    ]
  };

  const clients: Node[] = [
      {
          id: "client-web",
          label: "Web Dashboard",
          type: "NODE_TYPE_CLIENT",
          status: "NODE_STATUS_ACTIVE",
      },
      {
          id: "client-cli",
          label: "MCP CLI",
          type: "NODE_TYPE_CLIENT",
          status: "NODE_STATUS_ACTIVE",
      }
  ];

  const graph: Graph = {
      core: coreNode,
      clients: clients
  };

  return NextResponse.json(graph);
}
