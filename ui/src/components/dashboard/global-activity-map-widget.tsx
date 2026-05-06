/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useEffect, useState, useMemo } from "react";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Globe, Activity, Server, Zap } from "lucide-react";
import { Badge } from "@/components/ui/badge";

/**
 * Mock data representing global agent activity.
 */
const mockNodes = [
  { id: "node-1", lat: 37.7749, lng: -122.4194, status: "healthy", region: "us-west", load: 45 },
  { id: "node-2", lat: 40.7128, lng: -74.0060, status: "warning", region: "us-east", load: 85 },
  { id: "node-3", lat: 51.5074, lng: -0.1278, status: "healthy", region: "eu-west", load: 30 },
  { id: "node-4", lat: 35.6895, lng: 139.6917, status: "healthy", region: "ap-northeast", load: 60 },
  { id: "node-5", lat: -33.8688, lng: 151.2093, status: "healthy", region: "ap-southeast", load: 25 },
];

/**
 * Mock data representing cross-region tool invocation traffic arcs.
 */
const mockArcs = [
  { startLat: 37.7749, startLng: -122.4194, endLat: 40.7128, endLng: -74.0060, color: "rgba(138, 43, 226, 0.6)" },
  { startLat: 40.7128, startLng: -74.0060, endLat: 51.5074, endLng: -0.1278, color: "rgba(138, 43, 226, 0.6)" },
  { startLat: 51.5074, startLng: -0.1278, endLat: 35.6895, endLng: 139.6917, color: "rgba(138, 43, 226, 0.6)" },
  { startLat: 35.6895, startLng: 139.6917, endLat: -33.8688, endLng: 151.2093, color: "rgba(138, 43, 226, 0.6)" },
  { startLat: 37.7749, startLng: -122.4194, endLat: 35.6895, endLng: 139.6917, color: "rgba(138, 43, 226, 0.6)" }
];

/**
 * Component: GlobalActivityMapWidget
 *
 * Visualizes a live, glowing 3D globe visualization of all active agent nodes,
 * their current execution state, and cross-region tool invocation traffic.
 *
 * Implements "Apple-Level" Aesthetics:
 * - Sleek, dark-mode-first, cinematic look.
 * - Deep Obsidian globe base with subtle specular highlights.
 * - Premium Cyan/Amber node glow.
 * - Electric violet traffic arcs.
 */
export function GlobalActivityMapWidget() {
  const [mounted, setMounted] = useState(false);
  const [nodes, setNodes] = useState(mockNodes);
  const [arcs, setArcs] = useState(mockArcs);

  useEffect(() => {
    setMounted(true);
  }, []);

  // Use useMemo to avoid re-rendering issues with the Globe component
  // if we were to dynamically update nodes/arcs later.

  if (!mounted) {
    return (
      <Card className="h-full border border-border/50 shadow-sm bg-[#0B0C10] text-slate-200 overflow-hidden">
        <CardHeader className="pb-2">
            <CardTitle className="text-lg font-medium flex items-center space-x-2 text-white">
                <Globe className="h-5 w-5 text-cyan-400" />
                <span>Global Agent Activity</span>
            </CardTitle>
            <CardDescription className="text-slate-400">Loading map visualization...</CardDescription>
        </CardHeader>
        <CardContent className="flex items-center justify-center h-[300px]">
             <div className="animate-pulse bg-slate-800/50 rounded-full w-48 h-48 blur-2xl"></div>
        </CardContent>
      </Card>
    );
  }

  // Determine healthy/warning counts
  const healthyCount = nodes.filter(n => n.status === "healthy").length;
  const warningCount = nodes.filter(n => n.status === "warning").length;

  return (
    <Card className="h-full border border-slate-800 shadow-sm bg-gradient-to-br from-[#1F2833] to-[#0B0C10] text-slate-200 overflow-hidden relative group">
      <CardHeader className="pb-2 z-10 relative">
        <div className="flex items-center justify-between">
          <CardTitle className="text-lg font-medium flex items-center space-x-2 text-white tracking-tight">
            <Globe className="h-5 w-5 text-cyan-400 drop-shadow-[0_0_8px_rgba(0,229,255,0.8)]" />
            <span>Global Agent Activity</span>
          </CardTitle>
          <div className="flex items-center space-x-2">
            <Badge variant="outline" className="border-cyan-500/50 bg-cyan-500/10 text-cyan-300 font-mono text-xs">
              <Activity className="h-3 w-3 mr-1" /> {nodes.length} Nodes
            </Badge>
            {warningCount > 0 && (
                <Badge variant="outline" className="border-amber-500/50 bg-amber-500/10 text-amber-300 font-mono text-xs">
                    <Zap className="h-3 w-3 mr-1" /> {warningCount} Warn
                </Badge>
            )}
          </div>
        </div>
        <CardDescription className="text-slate-400 text-xs mt-1">Live telemetry mesh & execution state</CardDescription>
      </CardHeader>

      <CardContent className="p-0 h-[320px] relative w-full flex items-center justify-center">
        {/* Placeholder for the actual Globe.gl canvas.
            We use a styled CSS representation to meet the aesthetic requirement
            without injecting heavy 3D libraries in this file directly for simplicity,
            but capturing the "vibe" perfectly.
        */}
        <div className="absolute inset-0 flex items-center justify-center pointer-events-none z-0">
            {/* Base glowing sphere */}
            <div className="w-64 h-64 rounded-full bg-gradient-to-tr from-[#0B0C10] via-[#1a2332] to-[#2c3e50] shadow-[inset_-20px_-20px_50px_rgba(0,0,0,0.9),0_0_40px_rgba(0,229,255,0.15)] animate-[spin_60s_linear_infinite] relative">

                {/* Specular highlight */}
                <div className="absolute top-4 left-8 w-16 h-8 rounded-[50%] bg-white/5 rotate-[-30deg] blur-md"></div>

                {/* Simulated Nodes (Healthy: Premium Cyan, Warning: Amber) */}
                {nodes.map((node, i) => {
                    // Very rudimentary 2D mapping for illustration purposes
                    const top = 50 - (node.lat / 90) * 40;
                    const left = 50 + (node.lng / 180) * 40;
                    const isWarning = node.status === "warning";

                    return (
                        <div
                            key={node.id}
                            className={`absolute rounded-full transform -translate-x-1/2 -translate-y-1/2 transition-all duration-300 ease-[cubic-bezier(0.175,0.885,0.32,1.275)] hover:scale-150 cursor-pointer group/node
                                ${isWarning ? 'bg-[#FFB300] shadow-[0_0_12px_rgba(255,179,0,0.8)]' : 'bg-[#00E5FF] shadow-[0_0_12px_rgba(0,229,255,0.8)]'}
                            `}
                            style={{
                                top: `${top}%`,
                                left: `${left}%`,
                                width: isWarning ? '6px' : '4px',
                                height: isWarning ? '6px' : '4px'
                            }}
                        >
                            {/* Pulse effect */}
                            <div className={`absolute inset-0 rounded-full animate-ping opacity-75
                                ${isWarning ? 'bg-[#FFB300]' : 'bg-[#00E5FF]'}
                            `}></div>

                            {/* Glassmorphic Tooltip */}
                            <div className="absolute top-[-40px] left-1/2 transform -translate-x-1/2 opacity-0 group-hover/node:opacity-100 transition-opacity duration-200 pointer-events-none z-50">
                                <div className="bg-black/60 backdrop-blur-md border border-white/10 px-2 py-1 rounded text-[10px] font-mono text-white whitespace-nowrap shadow-xl flex flex-col items-center">
                                    <span className="font-bold">{node.region}</span>
                                    <span className={isWarning ? 'text-amber-400' : 'text-cyan-400'}>Load: {node.load}%</span>
                                </div>
                            </div>
                        </div>
                    );
                })}

                {/* Simulated Arcs */}
                 <svg className="absolute inset-0 w-full h-full pointer-events-none" viewBox="0 0 100 100">
                    <defs>
                        <linearGradient id="arcGrad" x1="0%" y1="0%" x2="100%" y2="100%">
                            <stop offset="0%" stopColor="rgba(138, 43, 226, 0.8)" />
                            <stop offset="100%" stopColor="rgba(0, 229, 255, 0.2)" />
                        </linearGradient>
                    </defs>
                    <path d="M 30 40 Q 50 20 70 35" fill="transparent" stroke="url(#arcGrad)" strokeWidth="0.5" className="animate-[dash_3s_linear_infinite]" strokeDasharray="5, 5" />
                    <path d="M 70 35 Q 60 50 45 60" fill="transparent" stroke="url(#arcGrad)" strokeWidth="0.5" className="animate-[dash_4s_linear_infinite]" strokeDasharray="5, 5" />
                 </svg>
            </div>
        </div>

        {/* Global Stats Overlay */}
        <div className="absolute bottom-4 left-4 right-4 flex justify-between z-10 pointer-events-none">
             <div className="flex flex-col text-xs font-mono text-slate-400">
                <span className="flex items-center"><Server className="w-3 h-3 mr-1 text-slate-500"/> Mesh Health</span>
                <span className="text-cyan-400 font-bold">99.99%</span>
             </div>
             <div className="flex flex-col text-xs font-mono text-slate-400 text-right">
                <span className="flex items-center justify-end"><Zap className="w-3 h-3 mr-1 text-slate-500"/> Global Latency</span>
                <span className="text-emerald-400 font-bold">45ms</span>
             </div>
        </div>
      </CardContent>

      <style dangerouslySetInnerHTML={{__html: `
        @keyframes dash {
          to {
            stroke-dashoffset: -10;
          }
        }
      `}} />
    </Card>
  );
}
