/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

// Manual TypeScript definitions matching proto/topology/v1/topology.proto

export enum NodeType {
  NODE_TYPE_UNSPECIFIED = 0,
  NODE_TYPE_CLIENT = 1,
  NODE_TYPE_CORE = 2,
  NODE_TYPE_SERVICE = 3,
  NODE_TYPE_TOOL = 4,
  NODE_TYPE_RESOURCE = 5,
  NODE_TYPE_PROMPT = 6,
  NODE_TYPE_API_CALL = 7,
  NODE_TYPE_MIDDLEWARE = 8,
  NODE_TYPE_WEBHOOK = 9,
}

export enum NodeStatus {
  NODE_STATUS_UNSPECIFIED = 0,
  NODE_STATUS_ACTIVE = 1,
  NODE_STATUS_INACTIVE = 2,
  NODE_STATUS_ERROR = 3,
}

export interface NodeMetrics {
  qps: number;
  latencyMs: number;
  errorRate: number;
}

export interface TopologyNode {
  id: string;
  label: string;
  type: NodeType;
  status: NodeStatus;
  metadata?: Record<string, string>;
  children?: TopologyNode[];
  metrics?: NodeMetrics;
}

export interface TopologyGraph {
  clients: TopologyNode[];
  core?: TopologyNode;
}
