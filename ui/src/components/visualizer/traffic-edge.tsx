/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { memo } from 'react';
import { BaseEdge, EdgeLabelRenderer, EdgeProps, getSmoothStepPath } from '@xyflow/react';
import { cn } from '@/lib/utils';

/**
 * TrafficEdge is a custom edge component that visualizes traffic flow.
 * It renders particles moving along the path if QPS > 0.
 */
export const TrafficEdge = memo(({
  id,
  sourceX,
  sourceY,
  targetX,
  targetY,
  sourcePosition,
  targetPosition,
  style = {},
  markerEnd,
  data,
}: EdgeProps) => {
  const [edgePath, labelX, labelY] = getSmoothStepPath({
    sourceX,
    sourceY,
    sourcePosition,
    targetX,
    targetY,
    targetPosition,
  });

  const qps = (data?.qps as number) || 0;
  const errorRate = (data?.errorRate as number) || 0;
  const isError = errorRate > 0.05; // 5% error rate threshold

  // Calculate duration: higher QPS = faster (lower duration)
  // 1 QPS -> 2s, 10 QPS -> 1s, 100 QPS -> 0.5s
  const duration = qps > 0 ? Math.max(0.5, 3 / Math.log2(qps + 2)) : 0;

  // Determine colors
  // Slate-400 for idle, Green-500 for active, Red-500 for error
  const edgeColor = isError ? '#ef4444' : (qps > 0.1 ? '#22c55e' : '#94a3b8');
  const strokeWidth = qps > 0 ? Math.min(3, 1 + Math.log10(qps + 1)) : 1.5;
  const opacity = qps > 0 ? 1 : 0.4;

  return (
    <>
      <BaseEdge
        path={edgePath}
        markerEnd={markerEnd}
        style={{
            ...style,
            stroke: edgeColor,
            strokeWidth,
            opacity,
            transition: 'stroke 0.5s, stroke-width 0.5s, opacity 0.5s'
        }}
      />

      {/* Moving Particle */}
      {qps > 0.1 && (
          <circle r={Math.min(4, 2 + qps/10)} fill={edgeColor}>
            <animateMotion dur={`${duration}s`} repeatCount="indefinite" path={edgePath} calcMode="linear" />
          </circle>
      )}

      {/* Multiple Particles for high traffic */}
      {qps > 10 && (
          <circle r={Math.min(3, 1 + qps/20)} fill={edgeColor} opacity="0.7">
            <animateMotion dur={`${duration}s`} begin={`${duration/2}s`} repeatCount="indefinite" path={edgePath} calcMode="linear" />
          </circle>
      )}

      {qps > 0.1 && (
        <EdgeLabelRenderer>
            <div
                style={{
                    position: 'absolute',
                    transform: `translate(-50%, -50%) translate(${labelX}px,${labelY}px)`,
                    pointerEvents: 'all',
                }}
                className={cn(
                    "px-1.5 py-0.5 rounded text-[9px] font-mono border bg-background/80 backdrop-blur-sm shadow-sm transition-colors",
                    isError ? "border-red-500 text-red-500" : "border-green-500 text-green-600"
                )}
            >
                {qps.toFixed(1)}/s
            </div>
        </EdgeLabelRenderer>
      )}
    </>
  );
});

TrafficEdge.displayName = 'TrafficEdge';
