/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useEffect, useRef, useState } from "react";
import { Trace } from "@/types/trace";
import { cn } from "@/lib/utils";

interface LiveSignalProps {
  traces: Trace[];
  className?: string;
  isConnected?: boolean;
}

interface SignalPoint {
  x: number;
  y: number;
  duration: number;
  status: "success" | "error";
  timestamp: number;
  alpha: number;
}

/**
 * LiveSignal component renders a real-time visualization of trace traffic.
 * It uses a canvas to draw a moving timeline of requests.
 */
export function LiveSignal({ traces, className, isConnected }: LiveSignalProps) {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const containerRef = useRef<HTMLDivElement>(null);
  const pointsRef = useRef<SignalPoint[]>([]);
  const lastProcessedTraceIdRef = useRef<string | null>(null);
  const animationFrameRef = useRef<number>();

  // Determine which traces are new since the last render cycle
  useEffect(() => {
    if (!traces || traces.length === 0) return;

    // We only want to add "new" traces to the visualization
    // The traces array is sorted newest-first by useTraces hook.
    // We iterate from the start until we hit the last processed ID.
    const newPoints: SignalPoint[] = [];
    const now = Date.now();

    for (const trace of traces) {
      if (trace.id === lastProcessedTraceIdRef.current) break;

      // Calculate initial position (off-screen right)
      // We'll normalize duration to a height factor later
      newPoints.push({
        x: 1.0, // 1.0 = 100% width (right edge)
        y: 0, // Calculated during draw based on duration
        duration: trace.totalDuration,
        status: trace.status as "success" | "error",
        timestamp: now,
        alpha: 1.0,
      });
    }

    if (newPoints.length > 0) {
      // Add new points to our ref
      pointsRef.current = [...pointsRef.current, ...newPoints];
      lastProcessedTraceIdRef.current = traces[0].id;
    }
  }, [traces]);

  useEffect(() => {
    const canvas = canvasRef.current;
    const container = containerRef.current;
    if (!canvas || !container) return;

    const ctx = canvas.getContext("2d");
    if (!ctx) return;

    // Handle resizing
    const resizeObserver = new ResizeObserver((entries) => {
      for (const entry of entries) {
        const { width, height } = entry.contentRect;
        // Handle high DPI displays
        const dpr = window.devicePixelRatio || 1;
        canvas.width = width * dpr;
        canvas.height = height * dpr;
        canvas.style.width = `${width}px`;
        canvas.style.height = `${height}px`;
        ctx.scale(dpr, dpr);
      }
    });

    resizeObserver.observe(container);

    // Animation Loop
    const render = () => {
      const width = canvas.width / (window.devicePixelRatio || 1);
      const height = canvas.height / (window.devicePixelRatio || 1);

      // Clear canvas with a trail effect
      // Instead of clearRect, we draw a semi-transparent rect to create trails if desired
      // But for a clean signal look, clearRect is better
      ctx.clearRect(0, 0, width, height);

      // Draw grid lines
      ctx.strokeStyle = "rgba(255, 255, 255, 0.05)";
      ctx.lineWidth = 1;
      ctx.beginPath();
      // Horizontal lines (log scale-ish markers)
      for (let i = 1; i < 4; i++) {
        const y = height - (height * i) / 4;
        ctx.moveTo(0, y);
        ctx.lineTo(width, y);
      }
      ctx.stroke();

      // Update and Draw Points
      const speed = 0.002; // Movement speed per frame (fraction of width)
      const survivingPoints: SignalPoint[] = [];

      pointsRef.current.forEach((point) => {
        // Move point left
        point.x -= speed;

        // If point is still visible
        if (point.x > -0.1) {
          survivingPoints.push(point);

          // Calculate height based on duration (Log scale to handle spikes)
          // Min height 4px, Max height 80% of canvas
          const logDuration = Math.log(point.duration + 1);
          const maxLog = Math.log(10000); // Baseline max 10s
          const normalizedHeight = Math.min(Math.max(logDuration / maxLog, 0.1), 1.0);
          const barHeight = normalizedHeight * (height * 0.8);

          const xPos = point.x * width;
          const yPos = height - barHeight;

          // Color
          if (point.status === "error") {
            ctx.fillStyle = `rgba(239, 68, 68, ${point.alpha})`; // Red-500
            ctx.shadowColor = "rgba(239, 68, 68, 0.5)";
          } else {
            ctx.fillStyle = `rgba(34, 197, 94, ${point.alpha})`; // Green-500
            ctx.shadowColor = "rgba(34, 197, 94, 0.5)";
          }
          ctx.shadowBlur = 10;

          // Draw Bar
          // Width of bar is fixed or dynamic? Fixed looks like a matrix rain
          const barWidth = 4;

          // Draw a pill shape
          ctx.beginPath();
          ctx.roundRect(xPos, yPos, barWidth, barHeight, 2);
          ctx.fill();

          // Reset shadow
          ctx.shadowBlur = 0;
        }
      });

      pointsRef.current = survivingPoints;
      animationFrameRef.current = requestAnimationFrame(render);
    };

    animationFrameRef.current = requestAnimationFrame(render);

    return () => {
      if (animationFrameRef.current) cancelAnimationFrame(animationFrameRef.current);
      resizeObserver.disconnect();
    };
  }, []);

  return (
    <div
        ref={containerRef}
        className={cn(
            "relative w-full h-24 bg-black/90 rounded-md border border-border overflow-hidden shadow-inner",
            className
        )}
    >
       {/* Background Grid / Texture (Optional) */}
       <div className="absolute inset-0 pointer-events-none opacity-20"
            style={{
                backgroundImage: "linear-gradient(0deg, transparent 24%, rgba(255, 255, 255, .05) 25%, rgba(255, 255, 255, .05) 26%, transparent 27%, transparent 74%, rgba(255, 255, 255, .05) 75%, rgba(255, 255, 255, .05) 76%, transparent 77%, transparent), linear-gradient(90deg, transparent 24%, rgba(255, 255, 255, .05) 25%, rgba(255, 255, 255, .05) 26%, transparent 27%, transparent 74%, rgba(255, 255, 255, .05) 75%, rgba(255, 255, 255, .05) 76%, transparent 77%, transparent)",
                backgroundSize: "30px 30px"
            }}
       ></div>

      <canvas ref={canvasRef} className="block w-full h-full relative z-10" />

      {/* Legend / Info Overlay */}
      <div className="absolute top-2 left-2 flex items-center gap-4 text-[10px] font-mono text-muted-foreground z-20 pointer-events-none select-none">
          <div className="flex items-center gap-1.5">
             <div className="w-2 h-2 rounded-full bg-green-500 shadow-[0_0_8px_rgba(34,197,94,0.6)] animate-pulse" />
             <span>LIVE SIGNAL</span>
          </div>
          <div className="opacity-50">Height = Duration (Log)</div>
      </div>

      {!isConnected && (
           <div className="absolute inset-0 flex items-center justify-center bg-background/50 backdrop-blur-[1px] z-30">
               <span className="text-xs font-mono text-muted-foreground">SIGNAL LOST</span>
           </div>
      )}
    </div>
  );
}
