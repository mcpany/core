/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

export type NodeType =
  | 'NODE_TYPE_UNSPECIFIED'
  | 'NODE_TYPE_CLIENT'
  | 'NODE_TYPE_CORE'
  | 'NODE_TYPE_SERVICE'
  | 'NODE_TYPE_TOOL'
  | 'NODE_TYPE_RESOURCE'
  | 'NODE_TYPE_PROMPT'
  | 'NODE_TYPE_API_CALL'
  | 'NODE_TYPE_MIDDLEWARE'
  | 'NODE_TYPE_WEBHOOK';

export type NodeStatus =
  | 'NODE_STATUS_UNSPECIFIED'
  | 'NODE_STATUS_ACTIVE'
  | 'NODE_STATUS_INACTIVE'
  | 'NODE_STATUS_ERROR';

export interface NodeMetrics {
  qps?: number;
  latencyMs?: number;
  errorRate?: number;
}

export interface Node {
  id: string;
  label: string;
  type: NodeType;
  status: NodeStatus;
  metadata?: Record<string, string>;
  children?: Node[];
  metrics?: NodeMetrics;
}

export interface Graph {
  clients?: Node[];
  core?: Node;
}
