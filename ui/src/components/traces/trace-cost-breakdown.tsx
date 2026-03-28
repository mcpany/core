/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useMemo, useState } from "react";
import { Trace, Span } from "@/types/trace";
import { estimateTokens, calculateCost, formatCost, formatTokenCount } from "@/lib/tokens";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";
import { Coins, Code, Terminal, Activity, ChevronRight, ChevronDown, Cpu, Database, Globe, MessageSquare } from "lucide-react";
import { cn } from "@/lib/utils";

interface SpanMetrics {
    span: Span;
    inputTokens: number;
    outputTokens: number;
    totalTokens: number;
    cost: number;
    depth: number;
    id: string;
}

function SpanIcon({ type, className }: { type: Span['type'], className?: string }) {
    switch (type) {
        case 'tool': return <Terminal className={cn("h-4 w-4 text-amber-500", className)} />;
        case 'service': return <Globe className={cn("h-4 w-4 text-indigo-500", className)} />;
        case 'resource': return <Database className={cn("h-4 w-4 text-cyan-500", className)} />;
        case 'core': return <Cpu className={cn("h-4 w-4 text-blue-500", className)} />;
        case 'prompt': return <MessageSquare className={cn("h-4 w-4 text-purple-500", className)} />;
        default: return <Activity className={cn("h-4 w-4 text-muted-foreground", className)} />;
    }
}

interface TraceCostBreakdownProps {
    trace: Trace;
}

/**
 * Intent: Document TraceCostBreakdown
 *
 * Params:
 *   - trace (Trace): The trace data containing span details.
 *
 * Returns:
 *   - React component displaying token usage and estimated cost for a trace.
 *
 * Side Effects:
 *   - None
 */
export function TraceCostBreakdown({ trace }: TraceCostBreakdownProps) {
    const [expandedRows, setExpandedRows] = useState<Set<string>>(new Set());

    const toggleRow = (id: string) => {
        setExpandedRows(prev => {
            const newSet = new Set(prev);
            if (newSet.has(id)) {
                newSet.delete(id);
            } else {
                newSet.add(id);
            }
            return newSet;
        });
    };

    const metricsData = useMemo(() => {
        const calculateSpanMetrics = (span: Span, depth: number): SpanMetrics[] => {
            const inputTokens = estimateTokens(span.input);
            const outputTokens = estimateTokens(span.output);
            const totalTokens = inputTokens + outputTokens;
            const cost = calculateCost(totalTokens);

            const metrics: SpanMetrics = {
                span,
                inputTokens,
                outputTokens,
                totalTokens,
                cost,
                depth,
                id: span.id,
            };

            let allMetrics = [metrics];

            if (span.children) {
                for (const child of span.children) {
                    allMetrics = allMetrics.concat(calculateSpanMetrics(child, depth + 1));
                }
            }

            return allMetrics;
        };

        return calculateSpanMetrics(trace.rootSpan, 0);
    }, [trace]);

    const totalInputTokens = useMemo(() => metricsData.reduce((acc, curr) => acc + curr.inputTokens, 0), [metricsData]);
    const totalOutputTokens = useMemo(() => metricsData.reduce((acc, curr) => acc + curr.outputTokens, 0), [metricsData]);
    const totalTokens = totalInputTokens + totalOutputTokens;
    const totalCost = calculateCost(totalTokens);

    const visibleRows = useMemo(() => {
        const rows: SpanMetrics[] = [];

        let i = 0;
        while (i < metricsData.length) {
            const item = metricsData[i];

            // If it's a root, it's always visible
            let isVisible = item.depth === 0;
            if (!isVisible) {
                // To be visible, all ancestors must be expanded
                let allAncestorsExpanded = true;
                let currentDepth = item.depth;
                let currentIndex = i;

                while (currentDepth > 0) {
                     let foundParent = false;
                     // Look backwards for the parent (the first item with depth = currentDepth - 1)
                     for (let j = currentIndex - 1; j >= 0; j--) {
                         if (metricsData[j].depth === currentDepth - 1) {
                             if (!expandedRows.has(metricsData[j].id)) {
                                 allAncestorsExpanded = false;
                             }
                             currentDepth = metricsData[j].depth;
                             currentIndex = j;
                             foundParent = true;
                             break;
                         }
                     }
                     if (!foundParent || !allAncestorsExpanded) break;
                }
                isVisible = allAncestorsExpanded;
            }

            if (isVisible) {
                rows.push(item);
            }
            i++;
        }
        return rows;
    }, [metricsData, expandedRows]);


    return (
        <div className="p-6 space-y-6 h-full overflow-y-auto">
            <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                <Card>
                    <CardHeader className="pb-2">
                        <CardTitle className="text-sm font-medium text-muted-foreground flex items-center gap-2">
                            <Code className="h-4 w-4" /> Input Tokens
                        </CardTitle>
                    </CardHeader>
                    <CardContent>
                        <div className="text-2xl font-bold font-mono">{formatTokenCount(totalInputTokens)}</div>
                    </CardContent>
                </Card>
                <Card>
                    <CardHeader className="pb-2">
                        <CardTitle className="text-sm font-medium text-muted-foreground flex items-center gap-2">
                            <Terminal className="h-4 w-4" /> Output Tokens
                        </CardTitle>
                    </CardHeader>
                    <CardContent>
                        <div className="text-2xl font-bold font-mono">{formatTokenCount(totalOutputTokens)}</div>
                    </CardContent>
                </Card>
                <Card>
                    <CardHeader className="pb-2">
                        <CardTitle className="text-sm font-medium text-muted-foreground flex items-center gap-2">
                            <Coins className="h-4 w-4 text-amber-500" /> Estimated Cost
                        </CardTitle>
                    </CardHeader>
                    <CardContent>
                        <div className="text-2xl font-bold font-mono text-amber-500">{formatCost(totalCost)}</div>
                    </CardContent>
                </Card>
            </div>

            <Card>
                <CardHeader>
                    <CardTitle className="text-lg">Cost Breakdown by Span</CardTitle>
                    <CardDescription>Detailed view of token usage and estimated cost for each step in the trace.</CardDescription>
                </CardHeader>
                <CardContent className="p-0">
                    <Table>
                        <TableHeader className="bg-muted/50">
                            <TableRow>
                                <TableHead className="w-[400px] pl-6">Span</TableHead>
                                <TableHead className="text-right">Input Tokens</TableHead>
                                <TableHead className="text-right">Output Tokens</TableHead>
                                <TableHead className="text-right">Total Tokens</TableHead>
                                <TableHead className="text-right pr-6">Estimated Cost</TableHead>
                            </TableRow>
                        </TableHeader>
                        <TableBody>
                            {visibleRows.map((row) => {
                                const hasChildren = row.span.children && row.span.children.length > 0;
                                const isExpanded = expandedRows.has(row.id);

                                return (
                                    <TableRow key={row.id} className={cn("hover:bg-muted/30", row.depth === 0 ? "font-medium" : "")}>
                                        <TableCell className="pl-6">
                                            <div className="flex items-center gap-2" style={{ paddingLeft: `${row.depth * 1.5}rem` }}>
                                                {hasChildren ? (
                                                    <button
                                                        onClick={(e) => { e.stopPropagation(); toggleRow(row.id); }}
                                                        className="p-0.5 rounded-sm hover:bg-muted"
                                                    >
                                                        {isExpanded ? <ChevronDown className="h-4 w-4 text-muted-foreground" /> : <ChevronRight className="h-4 w-4 text-muted-foreground" />}
                                                    </button>
                                                ) : (
                                                    <div className="w-5" />
                                                )}
                                                <SpanIcon type={row.span.type} />
                                                <span className="truncate max-w-[250px]" title={row.span.name}>{row.span.name}</span>
                                                {row.span.status === 'error' && <Badge variant="destructive" className="h-4 px-1 text-[10px] ml-2">ERR</Badge>}
                                            </div>
                                        </TableCell>
                                        <TableCell className="text-right font-mono text-xs">{row.inputTokens.toLocaleString()}</TableCell>
                                        <TableCell className="text-right font-mono text-xs">{row.outputTokens.toLocaleString()}</TableCell>
                                        <TableCell className="text-right font-mono text-xs font-semibold">{row.totalTokens.toLocaleString()}</TableCell>
                                        <TableCell className="text-right font-mono text-xs text-amber-500 pr-6">{formatCost(row.cost)}</TableCell>
                                    </TableRow>
                                );
                            })}
                        </TableBody>
                    </Table>
                </CardContent>
            </Card>
        </div>
    );
}
