/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { useMemo } from "react";
import {
    Card,
    CardContent,
    CardDescription,
    CardHeader,
    CardTitle,
} from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import {
    Table,
    TableBody,
    TableCell,
    TableHead,
    TableHeader,
    TableRow,
} from "@/components/ui/table";
import { ToolDefinition, ToolAnalytics } from "@/lib/client";
import { estimateTokens, formatTokenCount } from "@/lib/tokens";
import { Ghost, Zap, CheckCircle, AlertCircle, Info } from "lucide-react";

interface OptimizationTabProps {
    tools: ToolDefinition[];
    toolUsage: Record<string, ToolAnalytics>;
    onToggleTool: (name: string, disable: boolean) => void;
}

export function OptimizationTab({ tools, toolUsage, onToggleTool }: OptimizationTabProps) {
    const analysis = useMemo(() => {
        const ghostTools: { tool: ToolDefinition; tokens: number }[] = [];
        let totalWastedTokens = 0;
        let potentialSavings = 0;

        tools.forEach((tool) => {
            if (tool.disable) return; // Skip already disabled tools

            const tokens = estimateTokens(JSON.stringify(tool));
            const usageKey = `${tool.name}@${tool.serviceId}`;
            const stats = toolUsage[usageKey];
            const calls = stats ? stats.totalCalls : 0;

            // Definition of a "Ghost Tool":
            // 1. Large definition (> 500 tokens) AND 0 calls
            // 2. OR Very large definition (> 2000 tokens) AND low calls (< 5)
            const isGhost = (tokens > 500 && calls === 0) || (tokens > 2000 && calls < 5);

            if (isGhost) {
                ghostTools.push({ tool, tokens });
                totalWastedTokens += tokens;
            }
        });

        // Sort by tokens descending
        ghostTools.sort((a, b) => b.tokens - a.tokens);

        return { ghostTools, totalWastedTokens };
    }, [tools, toolUsage]);

    return (
        <div className="space-y-4">
            <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
                <Card>
                    <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                        <CardTitle className="text-sm font-medium">Context Efficiency</CardTitle>
                        <Zap className="h-4 w-4 text-muted-foreground" />
                    </CardHeader>
                    <CardContent>
                        <div className="text-2xl font-bold">
                            {analysis.ghostTools.length === 0 ? "100%" : "Optimizable"}
                        </div>
                        <p className="text-xs text-muted-foreground">
                            {analysis.ghostTools.length === 0
                                ? "No unused heavy tools found."
                                : `${analysis.ghostTools.length} ghost tools detected.`}
                        </p>
                    </CardContent>
                </Card>
                <Card>
                    <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                        <CardTitle className="text-sm font-medium">Potential Savings</CardTitle>
                        <Ghost className="h-4 w-4 text-muted-foreground" />
                    </CardHeader>
                    <CardContent>
                        <div className="text-2xl font-bold text-amber-500">
                            {formatTokenCount(analysis.totalWastedTokens)}
                        </div>
                        <p className="text-xs text-muted-foreground">
                            Tokens wasted on unused tools.
                        </p>
                    </CardContent>
                </Card>
            </div>

            <Card>
                <CardHeader>
                    <CardTitle className="flex items-center">
                        <Ghost className="mr-2 h-5 w-5 text-amber-500" />
                        Ghost Tools
                    </CardTitle>
                    <CardDescription>
                        These tools are consuming significant context window but are rarely or never used.
                        Disabling them will save tokens and costs.
                    </CardDescription>
                </CardHeader>
                <CardContent>
                    {analysis.ghostTools.length === 0 ? (
                        <div className="flex flex-col items-center justify-center py-8 text-center">
                            <CheckCircle className="h-12 w-12 text-green-500 mb-4" />
                            <h3 className="text-lg font-medium">Your context is optimized!</h3>
                            <p className="text-muted-foreground">
                                No heavy unused tools were found in your configuration.
                            </p>
                        </div>
                    ) : (
                        <Table>
                            <TableHeader>
                                <TableRow>
                                    <TableHead>Tool Name</TableHead>
                                    <TableHead>Service</TableHead>
                                    <TableHead>Size (Tokens)</TableHead>
                                    <TableHead>Calls</TableHead>
                                    <TableHead className="text-right">Action</TableHead>
                                </TableRow>
                            </TableHeader>
                            <TableBody>
                                {analysis.ghostTools.map(({ tool, tokens }) => (
                                    <TableRow key={tool.name}>
                                        <TableCell className="font-medium">
                                            <div className="flex flex-col">
                                                <span>{tool.name}</span>
                                                <span className="text-xs text-muted-foreground truncate max-w-[200px]">
                                                    {tool.description}
                                                </span>
                                            </div>
                                        </TableCell>
                                        <TableCell>
                                            <Badge variant="outline">{tool.serviceId}</Badge>
                                        </TableCell>
                                        <TableCell>
                                            <div className="flex items-center text-amber-600 font-mono">
                                                <AlertCircle className="w-3 h-3 mr-1" />
                                                {formatTokenCount(tokens)}
                                            </div>
                                        </TableCell>
                                        <TableCell>
                                            {toolUsage[`${tool.name}@${tool.serviceId}`]?.totalCalls || 0}
                                        </TableCell>
                                        <TableCell className="text-right">
                                            <Button
                                                variant="secondary"
                                                size="sm"
                                                onClick={() => onToggleTool(tool.name, true)}
                                            >
                                                Disable
                                            </Button>
                                        </TableCell>
                                    </TableRow>
                                ))}
                            </TableBody>
                        </Table>
                    )}
                </CardContent>
            </Card>
        </div>
    );
}
