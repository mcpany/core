/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useState, useEffect, useRef } from "react";
import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";
import { Loader2, Zap, Shield, ShieldAlert, Cpu } from "lucide-react";
import { cn } from "@/lib/utils";

// Types for the Swarm Topology Mock Data
interface SwarmNode {
    id: string;
    label: string;
    type: 'agent' | 'service' | 'validator';
    status: 'idle' | 'active' | 'stall' | 'locked';
    x?: number;
    y?: number;
}

interface SwarmEdge {
    source: string;
    target: string;
    status: 'healthy' | 'speculative' | 'blocked';
    hash: string;
}

interface SwarmTopologyData {
    nodes: SwarmNode[];
    edges: SwarmEdge[];
    anomalies: string[];
}

export function SwarmTopologyWidget() {
    const [data, setData] = useState<SwarmTopologyData | null>(null);
    const [loading, setLoading] = useState(true);
    const containerRef = useRef<HTMLDivElement>(null);

    useEffect(() => {
        // Mock data generator since we don't have the backend endpoint hooked up in this isolated UI component yet
        // In a real scenario, this would fetch from /api/v1/mock/swarm-topology
        const generateMockData = (): SwarmTopologyData => {
            const nodes: SwarmNode[] = [
                { id: 'n1', label: 'Primary Orchestrator', type: 'validator', status: 'locked', x: 50, y: 50 },
                { id: 'n2', label: 'Research Agent', type: 'agent', status: 'active', x: 20, y: 30 },
                { id: 'n3', label: 'Tool Exec', type: 'service', status: 'idle', x: 20, y: 70 },
                { id: 'n4', label: 'Synthesizer', type: 'agent', status: 'active', x: 80, y: 50 },
                { id: 'n5', label: 'Rogue Node', type: 'agent', status: 'stall', x: 80, y: 20 },
            ];

            const edges: SwarmEdge[] = [
                { source: 'n2', target: 'n1', status: 'healthy', hash: '0x1A4' },
                { source: 'n1', target: 'n3', status: 'healthy', hash: '0x2B9' },
                { source: 'n1', target: 'n4', status: 'healthy', hash: '0x3C1' },
                { source: 'n5', target: 'n1', status: 'blocked', hash: 'INVALID_GRAFT' },
            ];

            return {
                nodes,
                edges,
                anomalies: ['ARI Hub: Logic Graft Blocked from Rogue Node (n5)']
            };
        };

        const interval = setInterval(() => {
            // Simulate dynamic updates
            setData(generateMockData());
            setLoading(false);
        }, 3000);

        // Initial load
        setData(generateMockData());
        setLoading(false);

        return () => clearInterval(interval);
    }, []);

    if (loading) {
        return (
            <Card className="h-full flex flex-col relative overflow-hidden bg-slate-950 border-white/10">
                <div className="flex-1 flex items-center justify-center">
                    <Loader2 className="h-8 w-8 animate-spin text-cyan-500" />
                </div>
            </Card>
        );
    }

    if (!data) return null;

    return (
        <Card className="h-full flex flex-col relative overflow-hidden bg-slate-950 border-white/10 group">
            {/* Background Grid pattern */}
            <div className="absolute inset-0 z-0 pointer-events-none opacity-20"
                 style={{
                     backgroundImage: 'radial-gradient(circle at center, #334155 1px, transparent 1px)',
                     backgroundSize: '24px 24px',
                     maskImage: 'radial-gradient(ellipse at center, black 20%, transparent 80%)',
                     WebkitMaskImage: 'radial-gradient(ellipse at center, black 20%, transparent 80%)'
                 }}
            />

            <CardHeader className="relative z-10 pb-2 flex flex-row items-center justify-between border-b border-white/5 bg-slate-950/50 backdrop-blur-sm">
                <CardTitle className="text-sm font-medium flex items-center gap-2 text-slate-200">
                    <Shield className="w-4 h-4 text-cyan-400" />
                    Multi-Agent Swarm Topology
                </CardTitle>
                <div className="flex gap-2">
                     <span className="flex items-center gap-1 text-[10px] uppercase tracking-widest font-mono text-emerald-400">
                        <div className="w-1.5 h-1.5 rounded-full bg-emerald-400 animate-pulse" />
                        HAAL Active
                    </span>
                </div>
            </CardHeader>

            <CardContent className="flex-1 p-0 relative z-10" ref={containerRef}>
                {/* SVG Graph Layer */}
                <svg className="absolute inset-0 w-full h-full pointer-events-none">
                    <defs>
                        <linearGradient id="healthy-grad" x1="0%" y1="0%" x2="100%" y2="0%">
                            <stop offset="0%" stopColor="#0ea5e9" stopOpacity="0.8" />
                            <stop offset="100%" stopColor="#10b981" stopOpacity="0.8" />
                        </linearGradient>
                         <linearGradient id="blocked-grad" x1="0%" y1="0%" x2="100%" y2="0%">
                            <stop offset="0%" stopColor="#ef4444" stopOpacity="0.8" />
                            <stop offset="100%" stopColor="#ef4444" stopOpacity="0" />
                        </linearGradient>
                    </defs>

                    {data.edges.map((edge, i) => {
                        const source = data.nodes.find(n => n.id === edge.source);
                        const target = data.nodes.find(n => n.id === edge.target);
                        if (!source || !target) return null;

                        // Calculate SVG coordinates (percentage to pixels based on container, or just use % for simple rendering)
                        const sx = `${source.x}%`;
                        const sy = `${source.y}%`;
                        const tx = `${target.x}%`;
                        const ty = `${target.y}%`;

                        const isBlocked = edge.status === 'blocked';

                        // The path needs to use the same coordinate space, so we rely on the line element primarily,
                        // or provide actual pixel calculations. For simplicity and robustness in percentages,
                        // SVG can use viewbox, but since we're using absolute % coordinates overlaying HTML:
                        // line element works perfectly with percentages!

                        return (
                            <g key={`edge-${i}`}>
                                <line x1={sx} y1={sy} x2={tx} y2={ty}
                                      stroke={isBlocked ? "url(#blocked-grad)" : "url(#healthy-grad)"}
                                      strokeWidth="2"
                                      strokeDasharray={isBlocked ? "4 4" : "none"}
                                      className={cn("transition-all duration-1000", isBlocked ? "animate-pulse" : "")}
                                />
                            </g>
                        );
                    })}
                </svg>

                {/* Nodes Layer (HTML for easier interaction/styling) */}
                <div className="absolute inset-0">
                    {data.nodes.map((node) => {
                        const Icon = node.type === 'validator' ? Shield : node.type === 'agent' ? Cpu : Zap;
                        const isBlocked = node.status === 'stall';

                        return (
                            <div
                                key={node.id}
                                className={cn(
                                    "absolute w-12 h-12 -ml-6 -mt-6 rounded-2xl flex items-center justify-center backdrop-blur-md border cursor-pointer transition-all duration-300 hover:scale-110",
                                    node.status === 'locked' ? "bg-cyan-950/80 border-cyan-500 shadow-[0_0_15px_rgba(6,182,212,0.5)]" :
                                    isBlocked ? "bg-red-950/80 border-red-500 shadow-[0_0_15px_rgba(239,68,68,0.5)] animate-pulse" :
                                    "bg-slate-900/80 border-slate-700 hover:border-slate-500"
                                )}
                                style={{ left: `${node.x}%`, top: `${node.y}%` }}
                                title={`${node.label} (${node.status})`}
                            >
                                <Icon className={cn("w-5 h-5",
                                    node.status === 'locked' ? "text-cyan-400" :
                                    isBlocked ? "text-red-400" :
                                    "text-slate-400"
                                )} />

                                {/* Inner spin indicator for active nodes */}
                                {node.status === 'active' && (
                                    <div className="absolute inset-0 border border-emerald-500/50 rounded-2xl animate-[spin_4s_linear_infinite] [border-top-color:transparent]" />
                                )}
                            </div>
                        );
                    })}
                </div>

                {/* Anomalies Overlay */}
                {data.anomalies.length > 0 && (
                    <div className="absolute bottom-4 left-4 right-4 z-20">
                        {data.anomalies.map((msg, i) => (
                            <div key={i} className="bg-red-950/90 border border-red-500/50 rounded-md p-2 flex items-start gap-2 backdrop-blur-md shadow-lg transform transition-all translate-y-0 opacity-100">
                                <ShieldAlert className="w-4 h-4 text-red-400 mt-0.5 shrink-0" />
                                <div className="text-xs text-red-200 font-mono">
                                    {msg}
                                </div>
                            </div>
                        ))}
                    </div>
                )}
            </CardContent>
        </Card>
    );
}
