/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useState, useEffect } from "react";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Activity, ShieldAlert, Cpu } from "lucide-react";
import { cn } from "@/lib/utils";
import { Progress } from "@/components/ui/progress";
import { apiClient } from "@/lib/client";

interface SubagentStatus {
    id: string;
    name: string;
    status: 'aligned' | 'drifting' | 'hijacked';
    entropyScore: number;
    lastHeartbeat: number;
}



/**
 * Intent: Document ActiveIntentAlignmentWidget
 *
 * Params:
 *   - None
 *
 * Returns:
 *   - Documented below.
 *
 * Errors:
 *   - None
 *
 * Side Effects:
 *   - None
 *
 * ActiveIntentAlignmentWidget displays the intent alignment of active subagents.
 *
 * @returns The rendered component.
 */
export function ActiveIntentAlignmentWidget() {
    const [agents, setAgents] = useState<SubagentStatus[]>([]);

    useEffect(() => {
        const fetchStatus = async () => {
            try {
                const data = await apiClient.getActiveIntentAlignment();
                setAgents(data || []);
            } catch (err) {
                console.error("Failed to fetch intent alignment status:", err);
            }
        };
        fetchStatus();
        const interval = setInterval(fetchStatus, 1500);
        return () => clearInterval(interval);
    }, []);

    return (
        <Card className="col-span-12 lg:col-span-4 rounded-lg border bg-card text-card-foreground shadow-sm h-full backdrop-blur-sm bg-background/50 overflow-hidden relative">
            <div className="absolute inset-0 bg-gradient-to-br from-background/80 to-muted/20 pointer-events-none" />
            <CardHeader className="relative pb-2">
                <div className="flex items-center justify-between">
                    <CardTitle className="text-sm font-medium tracking-tight flex items-center gap-2">
                        <Cpu className="h-4 w-4 text-primary" />
                        Active Intent Alignment
                    </CardTitle>
                    <div className="flex items-center gap-1.5 px-2 py-0.5 rounded-full bg-primary/10 border border-primary/20">
                        <div className="h-1.5 w-1.5 rounded-full bg-primary animate-pulse" />
                        <span className="text-[10px] font-medium text-primary">LIVE</span>
                    </div>
                </div>
                <CardDescription className="text-xs">
                    Monitoring semantic drift and hardware-attested heartbeats.
                </CardDescription>
            </CardHeader>
            <CardContent className="relative space-y-4">
                {agents.map(agent => (
                    <div
                        key={agent.id}
                        className={cn(
                            "flex flex-col gap-2 p-3 rounded-md border transition-all duration-500",
                            agent.status === 'aligned' ? "bg-emerald-500/5 border-emerald-500/20" :
                            agent.status === 'drifting' ? "bg-amber-500/10 border-amber-500/30" :
                            "bg-rose-500/10 border-rose-500/40 shadow-[0_0_15px_rgba(225,29,72,0.15)] animate-in slide-in-from-right-1"
                        )}
                    >
                        <div className="flex justify-between items-center">
                            <div className="flex items-center gap-2">
                                <div className="font-mono text-xs font-semibold">{agent.id}</div>
                                <div className="text-xs text-muted-foreground truncate max-w-[120px]">{agent.name}</div>
                            </div>
                            {agent.status === 'aligned' ? (
                                <Activity className="h-3 w-3 text-emerald-500" />
                            ) : (
                                <ShieldAlert className={cn(
                                    "h-3 w-3",
                                    agent.status === 'drifting' ? "text-amber-500" : "text-rose-500 animate-pulse"
                                )} />
                            )}
                        </div>

                        <div className="space-y-1">
                            <div className="flex justify-between text-[10px] font-mono">
                                <span className="text-muted-foreground">Entropy</span>
                                <span className={cn(
                                    agent.status === 'aligned' ? "text-emerald-500" :
                                    agent.status === 'drifting' ? "text-amber-500" : "text-rose-500 font-bold"
                                )}>
                                    {agent.entropyScore.toFixed(1)}%
                                </span>
                            </div>
                            <Progress
                                value={agent.entropyScore}
                                className={cn(
                                    "h-1 bg-muted/50",
                                    agent.status === 'aligned' ? "[&>div]:bg-emerald-500" :
                                    agent.status === 'drifting' ? "[&>div]:bg-amber-500" :
                                    "[&>div]:bg-rose-500"
                                )}
                            />
                        </div>
                    </div>
                ))}
            </CardContent>
        </Card>
    );
}
