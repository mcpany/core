/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { memo } from 'react';
import { BaseEdge, EdgeLabelRenderer, EdgeProps, getSmoothStepPath } from '@xyflow/react';
import { cn } from '@/lib/utils';

/**
 * Intent: Document TrafficEdge
 *
 * Params:
 *   - None
 *
 * Returns:
 *   - None
 *
 * Errors:
 *   - None
 *
 * Side Effects:
 *   - None
 *
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
  const glowColor = isError ? 'rgba(239, 68, 68, 0.5)' : 'rgba(34, 197, 94, 0.5)';
  const strokeWidth = qps > 0 ? Math.min(3, 1 + Math.log10(qps + 1)) : 1.5;
  const opacity = qps > 0 ? 1 : 0.4;

  return (
    <>
      {/* Glow Effect for active edges */}
      {qps > 0.1 && (
        <BaseEdge
          path={edgePath}
          style={{
            stroke: glowColor,
            strokeWidth: strokeWidth * 3,
            strokeLinecap: 'round',
            opacity: 0.3,
            transition: 'stroke 0.5s, stroke-width 0.5s, opacity 0.5s',
            filter: 'blur(4px)',
          }}
        />
      )}

      <BaseEdge
        path={edgePath}
        markerEnd={markerEnd}
        style={{
            ...style,
            stroke: edgeColor,
            strokeWidth,
            opacity,
            strokeLinecap: 'round',
            transition: 'stroke 0.5s, stroke-width 0.5s, opacity 0.5s'
        }}
      />

      {/* Moving Particles */}
      {qps > 0.1 && (
        <>
          <circle r={Math.min(4, 2 + qps/10)} fill={edgeColor} filter={`drop-shadow(0 0 4px ${edgeColor})`}>
            <animateMotion dur={`${duration}s`} repeatCount="indefinite" path={edgePath} calcMode="linear" />
          </circle>

          {/* Trailing Particle for smooth flow effect */}
          <circle r={Math.min(2.5, 1 + qps/15)} fill={edgeColor} opacity="0.6" filter={`drop-shadow(0 0 2px ${edgeColor})`}>
            <animateMotion dur={`${duration}s`} begin={`${duration * 0.15}s`} repeatCount="indefinite" path={edgePath} calcMode="linear" />
          </circle>
        </>
      )}

      {/* Multiple Particles for high traffic */}
      {qps > 10 && (
        <>
          <circle r={Math.min(3, 1 + qps/20)} fill={edgeColor} opacity="0.8" filter={`drop-shadow(0 0 3px ${edgeColor})`}>
            <animateMotion dur={`${duration}s`} begin={`${duration * 0.4}s`} repeatCount="indefinite" path={edgePath} calcMode="linear" />
          </circle>
          <circle r={Math.min(2, 0.5 + qps/25)} fill={edgeColor} opacity="0.5" filter={`drop-shadow(0 0 2px ${edgeColor})`}>
            <animateMotion dur={`${duration}s`} begin={`${duration * 0.55}s`} repeatCount="indefinite" path={edgePath} calcMode="linear" />
          </circle>
        </>
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
                    "px-2 py-1 rounded-full text-[10px] font-medium border bg-background/60 backdrop-blur-md shadow-lg transition-all hover:scale-105 hover:bg-background/80",
                    isError ? "border-red-500/50 text-red-500 shadow-red-500/20" : "border-green-500/30 text-green-600 shadow-green-500/10"
                )}
            >
                {qps.toFixed(1)} req/s
            </div>
        </EdgeLabelRenderer>
      )}
    </>
  );
});

TrafficEdge.displayName = 'TrafficEdge';
