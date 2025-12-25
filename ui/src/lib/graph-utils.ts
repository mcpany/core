/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { Service } from "@/types/service";
import { Node, Edge, Position } from "@xyflow/react";
import dagre from "dagre";

const nodeWidth = 250;
const nodeHeight = 80;

export const getLayoutedElements = (nodes: Node[], edges: Edge[], direction = "TB") => {
  const dagreGraph = new dagre.graphlib.Graph();
  dagreGraph.setDefaultEdgeLabel(() => ({}));

  const isHorizontal = direction === "LR";
  dagreGraph.setGraph({ rankdir: direction });

  nodes.forEach((node) => {
    dagreGraph.setNode(node.id, { width: nodeWidth, height: nodeHeight });
  });

  edges.forEach((edge) => {
    dagreGraph.setEdge(edge.source, edge.target);
  });

  dagre.layout(dagreGraph);

  const layoutedNodes = nodes.map((node) => {
    const nodeWithPosition = dagreGraph.node(node.id);
    return {
      ...node,
      targetPosition: isHorizontal ? Position.Left : Position.Top,
      sourcePosition: isHorizontal ? Position.Right : Position.Bottom,
      // We are shifting the dagre node position (anchor=center center) to the top left
      // so it matches the React Flow node anchor point (top left).
      position: {
        x: nodeWithPosition.x - nodeWidth / 2,
        y: nodeWithPosition.y - nodeHeight / 2,
      },
    };
  });

  return { nodes: layoutedNodes, edges };
};

export const transformServicesToGraph = (services: Service[]) => {
    const nodes: Node[] = [];
    const edges: Edge[] = [];

    // Central Node
    nodes.push({
        id: "mcp-any",
        type: "central",
        data: { label: "MCP Any Gateway", status: "active" },
        position: { x: 0, y: 0 }, // Initial position, will be calculated by dagre
    });

    services.forEach((service) => {
        nodes.push({
            id: service.id,
            type: "service",
            data: {
                label: service.name,
                version: service.version,
                type: service.service_config?.http_service ? "HTTP" : service.service_config?.grpc_service ? "gRPC" : "Other",
                status: service.disable ? "disabled" : "active"
            },
            position: { x: 0, y: 0 },
        });

        edges.push({
            id: `e-mcp-${service.id}`,
            source: "mcp-any",
            target: service.id,
            animated: !service.disable,
            style: { stroke: service.disable ? "#9ca3af" : "#2563eb", strokeWidth: 2 },
        });
    });

    return { nodes, edges };
}
