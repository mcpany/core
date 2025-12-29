/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { NextResponse } from 'next/server';
import { Graph, Node, NodeType, NodeStatus } from '@/types/topology';

// Helper to generate random metrics
function getMetrics() {
  const qps = Math.random() * 100;
  const latencyMs = Math.random() * 200 + 10;
  const errorRate = Math.random() > 0.9 ? Math.random() * 0.05 : 0; // 10% chance of error
  return { qps, latencyMs, errorRate };
}

// Helper to create a node
function createNode(
  id: string,
  label: string,
  type: NodeType,
  status: NodeStatus = 'NODE_STATUS_ACTIVE',
  children: Node[] = []
): Node {
  return {
    id,
    label,
    type,
    status,
    metrics: getMetrics(),
    metadata: {
      version: '1.0.0',
      region: 'us-east-1',
      uptime: `${Math.floor(Math.random() * 24)}h ${Math.floor(Math.random() * 60)}m`,
    },
    children,
  };
}

export async function GET() {
  // Construct the graph
  const clients: Node[] = [
    createNode('client-web', 'Web Dashboard', 'NODE_TYPE_CLIENT'),
    createNode('client-cli', 'Claude Desktop', 'NODE_TYPE_CLIENT'),
  ];

  const toolsPostgres: Node[] = [
    createNode('tool-pg-query', 'query_db', 'NODE_TYPE_TOOL'),
    createNode('tool-pg-migration', 'run_migration', 'NODE_TYPE_TOOL'),
  ];

  const toolsOpenAI: Node[] = [
    createNode('tool-openai-chat', 'chat_completion', 'NODE_TYPE_TOOL'),
    createNode('tool-openai-embed', 'create_embedding', 'NODE_TYPE_TOOL'),
  ];

  const toolsFS: Node[] = [
    createNode('tool-fs-list', 'list_files', 'NODE_TYPE_TOOL'),
    createNode('tool-fs-read', 'read_file', 'NODE_TYPE_TOOL'),
    createNode('tool-fs-write', 'write_file', 'NODE_TYPE_TOOL', 'NODE_STATUS_INACTIVE'),
  ];

  const resourcesFS: Node[] = [
      createNode('res-fs-logs', 'system.log', 'NODE_TYPE_RESOURCE'),
  ]

  const services: Node[] = [
    createNode('svc-postgres', 'Postgres DB', 'NODE_TYPE_SERVICE', 'NODE_STATUS_ACTIVE', toolsPostgres),
    createNode('svc-openai', 'OpenAI API', 'NODE_TYPE_SERVICE', 'NODE_STATUS_ACTIVE', toolsOpenAI),
    createNode('svc-fs', 'Local Filesystem', 'NODE_TYPE_SERVICE', 'NODE_STATUS_ACTIVE', [...toolsFS, ...resourcesFS]),
  ];

  const core: Node = createNode('core-mcp-any', 'MCP Any Server', 'NODE_TYPE_CORE', 'NODE_STATUS_ACTIVE', services);

  const graph: Graph = {
    clients,
    core,
  };

  return NextResponse.json(graph);
}
