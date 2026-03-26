/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import React, { useMemo } from "react";
import { ResponsiveContainer, Treemap, Tooltip } from "recharts";
import { useRecursiveContext } from "./context-provider";
import { formatTokenCount } from "@/lib/tokens";
import { ToolDefinition } from "@/lib/client";

// Custom colors for different services (generated based on name or index)
const COLORS = [
  "#3b82f6", // Blue
  "#10b981", // Emerald
  "#f59e0b", // Amber
  "#8b5cf6", // Violet
  "#ef4444", // Red
  "#ec4899", // Pink
  "#6366f1", // Indigo
  "#14b8a6", // Teal
];

const getColor = (index: number) => COLORS[index % COLORS.length];

interface TreemapNode {
    name: string;
    size: number; // Token cost
    serviceId?: string;
    children?: TreemapNode[];
    color?: string;
}

const CustomizedContent = (props: any) => {
  const { root, depth, x, y, width, height, index, name, size } = props;

  // We only render labels for leaf nodes (tools) or top-level services if there's space
  // depth 1 = Service
  // depth 2 = Tool

  if (width < 50 || height < 30) return null;

  const fontSize = Math.min(width / 5, height / 2, 14);
  const color = props.colors ? props.colors[index % props.colors.length] : '#fff';

  return (
    <g>
      <rect
        x={x}
        y={y}
        width={width}
        height={height}
        style={{
          fill: depth === 1 ? 'rgba(255,255,255,0.1)' : (props.color || getColor(index)),
          stroke: '#fff',
          strokeWidth: 2 / (depth + 1e-10),
          strokeOpacity: 1 / (depth + 1e-10),
        }}
      />
      {depth > 1 && (
        <text
          x={x + width / 2}
          y={y + height / 2}
          textAnchor="middle"
          fill="#fff"
          fontSize={fontSize}
          fontWeight="bold"
          style={{ pointerEvents: 'none', textShadow: '0px 1px 2px rgba(0,0,0,0.5)' }}
        >
          {name}
        </text>
      )}
       {depth > 1 && (
        <text
          x={x + width / 2}
          y={y + height / 2 + fontSize + 4}
          textAnchor="middle"
          fill="rgba(255,255,255,0.8)"
          fontSize={fontSize * 0.8}
          style={{ pointerEvents: 'none' }}
        >
          {formatTokenCount(size)}
        </text>
      )}
    </g>
  );
};

const CustomTooltip = ({ active, payload }: any) => {
    if (active && payload && payload.length) {
        const data = payload[0].payload;
        return (
            <div className="bg-popover text-popover-foreground border rounded-md shadow-md p-2 text-sm">
                <p className="font-semibold">{data.name}</p>
                {data.serviceId && <p className="text-xs text-muted-foreground">{data.serviceId}</p>}
                <p className="font-mono mt-1">
                    {formatTokenCount(data.size)} tokens
                </p>
            </div>
        );
    }
    return null;
};

/**
 * Visualization component that renders a treemap of tool token costs.
 * It groups tools by service and color-codes them for easy analysis.
 */
export function ContextTreemap() {
    const { tools, getToolCost, loading, disabledToolIds } = useRecursiveContext();

    const data = useMemo(() => {
        // Group by service
        const servicesMap: Record<string, TreemapNode> = {};

        tools.forEach((tool) => {
            const toolId = `${tool.serviceId}.${tool.name}`;
            if (disabledToolIds.has(toolId)) return; // Skip disabled tools in simulation

            if (!servicesMap[tool.serviceId]) {
                servicesMap[tool.serviceId] = {
                    name: tool.serviceId,
                    size: 0,
                    children: []
                };
            }

            const cost = getToolCost(tool);
            servicesMap[tool.serviceId].children?.push({
                name: tool.name,
                size: cost,
                serviceId: tool.serviceId
            });
            servicesMap[tool.serviceId].size += cost;
        });

        // Convert map to array
        return Object.values(servicesMap).map((node, index) => ({
            ...node,
            color: getColor(index) // Assign color to service
        })).sort((a, b) => b.size - a.size); // Sort largest first

    }, [tools, getToolCost, disabledToolIds]);

    if (loading) {
        return (
            <div className="flex h-full items-center justify-center text-muted-foreground">
                Loading context data...
            </div>
        );
    }

    if (data.length === 0) {
        return (
            <div className="flex h-full items-center justify-center flex-col gap-2 text-muted-foreground">
                <p>No active tools in context.</p>
                <p className="text-sm">Seed data or enable tools to visualize usage.</p>
            </div>
        );
    }

    return (
        <div className="w-full h-full min-h-[400px]">
            <ResponsiveContainer width="100%" height="100%">
                <Treemap
                    data={data}
                    dataKey="size"
                    aspectRatio={4 / 3}
                    stroke="#fff"
                    fill="#8884d8"
                    content={<CustomizedContent colors={COLORS} />}
                >
                    <Tooltip content={<CustomTooltip />} />
                </Treemap>
            </ResponsiveContainer>
        </div>
    );
}
