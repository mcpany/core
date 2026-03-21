/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useState } from "react";
import { useTraces } from "@/hooks/use-traces";
import { Card, CardContent } from "@/components/ui/card";
import { ScrollArea } from "@/components/ui/scroll-area";
import { formatDistanceToNow } from "date-fns";
import { ChevronDown, ChevronRight, Activity, Wrench, AlertCircle, CheckCircle2 } from "lucide-react";
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from "@/components/ui/collapsible";
import { cn } from "@/lib/utils";

interface ToolActivityFeedProps {
    serviceIdFilter?: string;
}

export function ToolActivityFeed({ serviceIdFilter }: ToolActivityFeedProps) {
    const { traces, isConnected } = useTraces();

    // Filter for tool traces
    const toolTraces = traces.filter(t => t.rootSpan?.type === 'tool' || t.rootSpan?.name === 'calculate_sum');



    if (toolTraces.length === 0) {
        return (
            <div className="flex flex-col items-center justify-center p-12 text-muted-foreground border rounded-lg border-dashed">
                <Activity className="h-8 w-8 mb-4 opacity-50" />
                <p>No tool activity yet</p>
                <p className="text-xs mt-2">Execute a tool to see traces appear here</p>
            </div>
        );
    }

    return (
        <div className="space-y-4">
            <div className="flex justify-between items-center mb-4">
                <h3 className="text-lg font-medium">Live Activity Feed</h3>
                <div className="flex items-center space-x-2">
                    <span className="relative flex h-3 w-3">
                        {isConnected && (
                            <>
                                <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-green-400 opacity-75"></span>
                                <span className="relative inline-flex rounded-full h-3 w-3 bg-green-500"></span>
                            </>
                        )}
                        {!isConnected && <span className="relative inline-flex rounded-full h-3 w-3 bg-red-500"></span>}
                    </span>
                    <span className="text-xs text-muted-foreground">
                        {isConnected ? "Live" : "Disconnected"}
                    </span>
                </div>
            </div>

            <ScrollArea className="h-[600px] pr-4">
                <div className="space-y-4">
                    {toolTraces.map((trace) => (
                        <TraceItem key={trace.id} trace={trace} />
                    ))}
                </div>
            </ScrollArea>
        </div>
    );
}

function TraceItem({ trace }: { trace: any }) {
    const [isOpen, setIsOpen] = useState(false);
    const isSuccess = trace.status === 'success';
    const toolName = trace.rootSpan?.name || 'Unknown Tool';
    const timestamp = trace.timestamp ? new Date(trace.timestamp) : new Date();

    return (
        <Card className={cn(
            "backdrop-blur-sm bg-background/50 border-l-4 transition-all duration-200",
            isSuccess ? "border-l-green-500" : "border-l-red-500",
            isOpen ? "shadow-md" : "shadow-sm hover:shadow-md"
        )}>
            <CardContent className="p-0">
                <Collapsible open={isOpen} onOpenChange={setIsOpen}>
                    <CollapsibleTrigger className="w-full flex items-center justify-between p-4 hover:bg-muted/50 rounded-t-lg transition-colors focus:outline-none">
                        <div className="flex items-center space-x-4">
                            <div className="bg-muted p-2 rounded-full">
                                <Wrench className="h-4 w-4 text-muted-foreground" />
                            </div>
                            <div className="text-left">
                                <div className="flex items-center space-x-2">
                                    <span className="font-semibold">{toolName}</span>
                                    {isSuccess ? (
                                        <CheckCircle2 className="h-4 w-4 text-green-500" />
                                    ) : (
                                        <AlertCircle className="h-4 w-4 text-red-500" />
                                    )}
                                </div>
                                <div className="text-xs text-muted-foreground flex items-center space-x-2 mt-1">
                                    <span>{formatDistanceToNow(timestamp, { addSuffix: true })}</span>
                                    <span>•</span>
                                    <span>{trace.totalDuration}ms</span>
                                </div>
                            </div>
                        </div>
                        <div className="flex items-center space-x-4 text-muted-foreground">
                            {isOpen ? <ChevronDown className="h-5 w-5" /> : <ChevronRight className="h-5 w-5" />}
                        </div>
                    </CollapsibleTrigger>

                    <CollapsibleContent>
                        <div className="p-4 border-t bg-muted/10 space-y-4">
                            <div>
                                <h4 className="text-xs font-semibold uppercase text-muted-foreground mb-2">Trace Data</h4>
                                <div className="bg-background border rounded-md p-3 overflow-x-auto max-h-[300px] overflow-y-auto">
                                    <pre className="text-xs font-mono">
                                        {JSON.stringify(trace, null, 2)}
                                    </pre>
                                </div>
                            </div>
                        </div>
                    </CollapsibleContent>
                </Collapsible>
            </CardContent>
        </Card>
    );
}
