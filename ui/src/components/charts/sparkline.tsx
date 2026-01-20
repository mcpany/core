/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useMemo } from 'react';
import { cn } from "@/lib/utils";

interface SparklineProps {
    data: number[];
    width?: number;
    height?: number;
    className?: string;
    color?: string; // hex color usually
    max?: number;
}

export function Sparkline({ data, width = 60, height = 24, className, color = "#22c55e", max }: SparklineProps) {
    const { path, fillPath, gradientId } = useMemo(() => {
        if (!data || data.length === 0) return { path: "", fillPath: "", gradientId: "" };

        // Dynamic scaling
        const effectiveMax = max !== undefined ? max : Math.max(...data, 1);
        const step = width / (data.length - 1 || 1);

        const points = data.map((val, i) => {
            const x = i * step;
            // Invert Y (SVG 0 is top). Clamp to height-1 to avoid clipping stroke
            const y = Math.min(height - 1, Math.max(1, height - (val / effectiveMax) * height));
            return [x, y];
        });

        if (points.length === 1) {
             // If only one point, draw a flat line
             points.push([width, points[0][1]]);
        }

        const d = points.map((p, i) => `${i === 0 ? 'M' : 'L'} ${p[0].toFixed(1)} ${p[1].toFixed(1)}`).join(" ");
        const fillD = `${d} L ${points[points.length-1][0]} ${height} L ${points[0][0]} ${height} Z`;

        // Unique ID for gradient based on color
        const gradientId = `gradient-${color.replace(/[^a-zA-Z0-9]/g, '')}`;

        return { path: d, fillPath: fillD, gradientId };
    }, [data, width, height, max, color]);

    if (!data || data.length === 0) return <div className={cn("bg-muted/20 rounded", className)} style={{ width, height }} />;

    return (
        <svg width={width} height={height} className={cn("overflow-hidden", className)} viewBox={`0 0 ${width} ${height}`}>
             <defs>
                <linearGradient id={gradientId} x1="0" x2="0" y1="0" y2="1">
                    <stop offset="0%" stopColor={color} stopOpacity="0.3" />
                    <stop offset="100%" stopColor={color} stopOpacity="0.0" />
                </linearGradient>
            </defs>
            <path
                d={fillPath}
                fill={`url(#${gradientId})`}
                stroke="none"
            />
            <path
                d={path}
                fill="none"
                stroke={color}
                strokeWidth="1.5"
                strokeLinecap="round"
                strokeLinejoin="round"
            />
        </svg>
    );
}
