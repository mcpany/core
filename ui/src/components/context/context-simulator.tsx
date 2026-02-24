/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import React, { useMemo } from "react";
import { useRecursiveContext } from "./context-provider";
import { formatTokenCount, calculateCost, formatCost } from "@/lib/tokens";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Checkbox } from "@/components/ui/checkbox";
import { Label } from "@/components/ui/label";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Ghost, RefreshCw, Zap } from "lucide-react";
import { Separator } from "@/components/ui/separator";

/**
 * Component that simulates context reduction by allowing users to toggle tools.
 * It displays statistics about current tokens and potential savings.
 */
export function ContextSimulator() {
    const {
        tools,
        services,
        disabledToolIds,
        toggleTool,
        enableAll,
        totalTokens,
        projectedTokens,
        getToolCost
    } = useRecursiveContext();

    const groupedTools = useMemo(() => {
        const groups: Record<string, typeof tools> = {};
        tools.forEach(tool => {
            if (!groups[tool.serviceId]) groups[tool.serviceId] = [];
            groups[tool.serviceId].push(tool);
        });
        return groups;
    }, [tools]);

    const savings = totalTokens - projectedTokens;
    const savingsPercent = totalTokens > 0 ? (savings / totalTokens) * 100 : 0;
    const cost = calculateCost(projectedTokens);
    const costSavings = calculateCost(savings);

    return (
        <Card className="h-full flex flex-col border-l rounded-none lg:rounded-lg">
            <CardHeader className="pb-3">
                <CardTitle className="text-lg flex items-center justify-between">
                    <span>Simulator</span>
                    <Button variant="ghost" size="icon" onClick={enableAll} title="Reset">
                        <RefreshCw className="h-4 w-4" />
                    </Button>
                </CardTitle>
                <CardDescription>
                    Toggle tools to simulate context reduction.
                </CardDescription>
            </CardHeader>
            <CardContent className="flex-1 overflow-hidden flex flex-col gap-4">

                {/* Stats */}
                <div className="grid grid-cols-2 gap-4 bg-muted/30 p-3 rounded-lg">
                    <div>
                        <div className="text-xs text-muted-foreground">Current Context</div>
                        <div className="text-2xl font-bold font-mono">
                            {formatTokenCount(projectedTokens)}
                        </div>
                        <div className="text-xs text-muted-foreground mt-1">
                            ~{formatCost(cost)}/call
                        </div>
                    </div>
                    <div>
                        <div className="text-xs text-muted-foreground">Savings</div>
                        <div className={`text-2xl font-bold font-mono ${savings > 0 ? 'text-green-500' : ''}`}>
                            {formatTokenCount(savings)}
                        </div>
                        <div className="text-xs text-muted-foreground mt-1">
                            {savingsPercent.toFixed(1)}% reduction
                        </div>
                    </div>
                </div>

                <Separator />

                <div className="flex items-center justify-between">
                    <Label className="text-sm font-medium">Services & Tools</Label>
                    <Badge variant="outline" className="text-[10px]">
                        {tools.length - disabledToolIds.size} / {tools.length} Active
                    </Badge>
                </div>

                <ScrollArea className="flex-1 pr-4">
                    <div className="space-y-6">
                        {Object.entries(groupedTools).map(([serviceId, serviceTools]) => (
                            <div key={serviceId} className="space-y-2">
                                <div className="flex items-center justify-between sticky top-0 bg-background z-10 py-1">
                                    <Label className="font-semibold text-sm flex items-center gap-2">
                                        {serviceId}
                                        <Badge variant="secondary" className="text-[10px] h-4">
                                            {serviceTools.length}
                                        </Badge>
                                    </Label>
                                </div>
                                <div className="pl-2 space-y-2 border-l-2 border-muted ml-1">
                                    {serviceTools.map(tool => {
                                        const id = `${tool.serviceId}.${tool.name}`;
                                        const disabled = disabledToolIds.has(id);
                                        const cost = getToolCost(tool);
                                        const isGhost = cost > 1000; // Simplified ghost logic

                                        return (
                                            <div key={id} className="flex items-start space-x-2 group">
                                                <Checkbox
                                                    id={id}
                                                    checked={!disabled}
                                                    onCheckedChange={() => toggleTool(tool.serviceId, tool.name)}
                                                    className="mt-1"
                                                />
                                                <div className="grid gap-1.5 leading-none w-full">
                                                    <div className="flex items-center justify-between">
                                                        <Label
                                                            htmlFor={id}
                                                            className={`text-sm font-medium cursor-pointer ${disabled ? 'text-muted-foreground line-through' : ''}`}
                                                        >
                                                            {tool.name}
                                                        </Label>
                                                        <span className={`text-xs font-mono ${isGhost && !disabled ? 'text-amber-500 font-bold' : 'text-muted-foreground'}`}>
                                                            {formatTokenCount(cost)}
                                                        </span>
                                                    </div>
                                                    <p className="text-[10px] text-muted-foreground line-clamp-1" title={tool.description}>
                                                        {tool.description || "No description"}
                                                    </p>
                                                    {isGhost && !disabled && (
                                                        <div className="flex items-center text-[10px] text-amber-500">
                                                            <Ghost className="w-3 h-3 mr-1" /> Heavy Tool
                                                        </div>
                                                    )}
                                                </div>
                                            </div>
                                        );
                                    })}
                                </div>
                            </div>
                        ))}
                    </div>
                </ScrollArea>

                {savings > 0 && (
                    <div className="pt-2">
                        <Button className="w-full" variant="outline" disabled>
                            Apply Optimization (Coming Soon)
                        </Button>
                    </div>
                )}
            </CardContent>
        </Card>
    );
}
