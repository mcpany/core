/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

// Manual TypeScript definitions matching proto/topology/v1/topology.proto
// Note: JSON serialization of Protobuf enums uses strings by default.

/**
 * NodeType enum represents the type of a node in the topology.
 */
export type NodeType =
  | "NODE_TYPE_UNSPECIFIED"
  | "NODE_TYPE_CLIENT"
  | "NODE_TYPE_CORE"
  | "NODE_TYPE_SERVICE"
  | "NODE_TYPE_TOOL"
  | "NODE_TYPE_RESOURCE"
  | "NODE_TYPE_PROMPT"
  | "NODE_TYPE_API_CALL"
  | "NODE_TYPE_MIDDLEWARE"
  | "NODE_TYPE_WEBHOOK";

/**
 * NodeStatus enum represents the operational status of a node.
 */
export type NodeStatus =
  | "NODE_STATUS_UNSPECIFIED"
  | "NODE_STATUS_ACTIVE"
  | "NODE_STATUS_INACTIVE"
  | "NODE_STATUS_ERROR";

/**
 * NodeMetrics contains performance metrics for a node.
 */
export interface NodeMetrics {
  qps: number;
  latencyMs: number;
  errorRate: number;
}

/**
 * TopologyNode represents a single node in the graph.
 */
export interface TopologyNode {
  id: string;
  label: string;
  type: NodeType;
  status: NodeStatus;
  metadata?: Record<string, string>;
  children?: TopologyNode[];
  metrics?: NodeMetrics;
}

/**
 * TopologyGraph represents the full network topology.
 */
export interface TopologyGraph {
  clients: TopologyNode[];
  core?: TopologyNode;
}
