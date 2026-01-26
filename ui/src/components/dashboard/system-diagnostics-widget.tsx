/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect } from "react";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { AlertTriangle, ArrowRight, Bug, RefreshCcw } from "lucide-react";
import { apiClient } from "@/lib/client";
import { analyzeTrace } from "@/lib/diagnostics";
import { Button } from "@/components/ui/button";
import Link from "next/link";
import { ScrollArea } from "@/components/ui/scroll-area";

interface DiagnosticSummary {
    title: string;
    count: number;
    type: 'error' | 'warning' | 'info';
    description: string;
}

/**
 * SystemDiagnosticsWidget component.
 * aggregates and displays system diagnostics from traces.
 * @returns The rendered component.
 */
export function SystemDiagnosticsWidget() {
    const [summaries, setSummaries] = useState<DiagnosticSummary[]>([]);
    const [loading, setLoading] = useState(true);
    const [totalAnalyzed, setTotalAnalyzed] = useState(0);

    const loadDiagnostics = async () => {
        setLoading(true);
        try {
            const traces = await apiClient.listTraces();
            setTotalAnalyzed(traces.length);

            const counts = new Map<string, { count: number, type: 'error' | 'warning' | 'info', description: string }>();

            traces.forEach(trace => {
                // Only look at errors
                if (trace.status !== 'error') return;

                const diags = analyzeTrace(trace);
                diags.forEach(diag => {
                    const existing = counts.get(diag.title);
                    if (existing) {
                        existing.count++;
                    } else {
                        counts.set(diag.title, {
                            count: 1,
                            type: diag.type,
                            description: diag.message
                        });
                    }
                });
            });

            const sorted = Array.from(counts.entries())
                .map(([title, data]) => ({ title, ...data }))
                .sort((a, b) => b.count - a.count);

            setSummaries(sorted);
        } catch (error) {
            console.error("Failed to fetch diagnostics", error);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        loadDiagnostics();
    }, []);

    return (
        <Card className="col-span-3 backdrop-blur-sm bg-background/50 flex flex-col h-full">
            <CardHeader className="flex flex-row items-center justify-between pb-2 space-y-0">
                <div className="space-y-1">
                    <CardTitle className="text-sm font-medium flex items-center gap-2">
                        <Bug className="h-4 w-4 text-primary" />
                        System Diagnostics
                    </CardTitle>
                    <CardDescription className="text-xs">
                        Analysis of {totalAnalyzed} recent traces
                    </CardDescription>
                </div>
                <Button variant="ghost" size="icon" onClick={loadDiagnostics} disabled={loading} className="h-8 w-8">
                    <RefreshCcw className={`h-4 w-4 ${loading ? 'animate-spin' : ''}`} />
                </Button>
            </CardHeader>
            <CardContent className="flex-1 overflow-hidden">
                <ScrollArea className="h-full pr-4">
                    <div className="space-y-4">
                        {summaries.map((summary) => (
                            <div key={summary.title} className="flex flex-col gap-1 border-l-2 pl-3 py-1" style={{
                                borderColor: summary.type === 'error' ? 'var(--destructive)' : 'var(--warning)'
                            }}>
                                <div className="flex items-center justify-between text-xs">
                                    <span className="font-semibold">{summary.title}</span>
                                    <Badge variant={summary.type === 'error' ? 'destructive' : 'secondary'} className="text-[10px] px-1.5 h-5">
                                        {summary.count} issues
                                    </Badge>
                                </div>
                                <p className="text-[10px] text-muted-foreground line-clamp-2">
                                    {summary.description}
                                </p>
                            </div>
                        ))}
                        {summaries.length === 0 && !loading && (
                            <div className="flex flex-col items-center justify-center py-8 text-center gap-2">
                                <div className="h-8 w-8 rounded-full bg-green-100 dark:bg-green-900/30 flex items-center justify-center">
                                    <AlertTriangle className="h-4 w-4 text-green-600 dark:text-green-400" />
                                </div>
                                <div className="space-y-1">
                                    <p className="text-sm font-medium">All Systems Operational</p>
                                    <p className="text-xs text-muted-foreground">No common errors detected in recent traffic.</p>
                                </div>
                            </div>
                        )}
                        {loading && summaries.length === 0 && (
                            <div className="text-center py-4 text-xs text-muted-foreground">
                                Analyzing system traces...
                            </div>
                        )}
                    </div>
                </ScrollArea>
            </CardContent>
            <div className="p-4 pt-0 mt-auto">
                <Link href="/inspector" className="w-full">
                    <Button variant="outline" size="sm" className="w-full text-xs h-8">
                        View All Traces <ArrowRight className="ml-2 h-3 w-3" />
                    </Button>
                </Link>
            </div>
        </Card>
    );
}
