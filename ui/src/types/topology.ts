/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

/**
 * Represents the type of a node in the topology graph.
 * Used for visual classification in the graph.
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

/**
 * Represents the operational status of a node.
 */
export type NodeStatus =
  | 'NODE_STATUS_UNSPECIFIED'
  | 'NODE_STATUS_ACTIVE'
  | 'NODE_STATUS_INACTIVE'
  | 'NODE_STATUS_ERROR';

/**
 * Performance metrics associated with a node.
 */
export interface NodeMetrics {
  /** Queries per second. */
  qps?: number;
  /** Latency in milliseconds. */
  latencyMs?: number;
  /** Error rate as a percentage or fraction. */
  errorRate?: number;
}

/**
 * Represents a single node in the topology graph.
 */
export interface Node {
  /** Unique identifier for the node. */
  id: string;
  /** Display label for the node. */
  label: string;
  /** The type of the node. */
  type: NodeType;
  /** The current status of the node. */
  status: NodeStatus;
  /** Additional metadata associated with the node. */
  metadata?: Record<string, string>;
  /** Child nodes contained within this node (if any). */
  children?: Node[];
  /** Performance metrics for this node. */
  metrics?: NodeMetrics;
}

/**
 * Represents the entire topology graph structure.
 */
export interface Graph {
  /** List of client nodes connected to the system. */
  clients?: Node[];
  /** The core server node. */
  core?: Node;
}
