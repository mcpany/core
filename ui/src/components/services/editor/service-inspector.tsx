/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useTraces } from "@/hooks/use-traces";
import { InspectorTable } from "@/components/inspector/inspector-table";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { RefreshCcw, Unplug, Pause, Play, Trash2 } from "lucide-react";
import { UpstreamServiceConfig } from "@/lib/client";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";

interface ServiceInspectorProps {
    service: UpstreamServiceConfig;
}

/**
 * ServiceInspector component.
 * @param props - The component props.
 * @param props.service - The service property.
 * @returns The rendered component.
 */
export function ServiceInspector({ service }: ServiceInspectorProps) {
    const {
        traces,
        loading,
        isConnected,
        isPaused,
        setIsPaused,
        clearTraces,
        refresh
    } = useTraces();

    // Filter traces by service name
    // The serviceName in rootSpan usually matches the service ID or Name
    const filteredTraces = traces.filter(
        t => t.rootSpan.serviceName === service.name || t.rootSpan.serviceName === service.id
    );

    return (
        <Card className="h-[600px] flex flex-col">
            <CardHeader className="pb-2">
                <div className="flex items-center justify-between">
                    <div>
                        <CardTitle className="text-lg flex items-center gap-2">
                            Live Traffic
                            <Badge variant={isConnected ? "outline" : "destructive"} className="font-mono text-xs gap-1 ml-2">
                                {isConnected ? (
                                    <>
                                        <span className="relative flex h-2 w-2">
                                            <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-green-400 opacity-75"></span>
                                            <span className="relative inline-flex rounded-full h-2 w-2 bg-green-500"></span>
                                        </span>
                                        Connected
                                    </>
                                ) : (
                                    <>
                                        <Unplug className="h-3 w-3" /> Disconnected
                                    </>
                                )}
                            </Badge>
                        </CardTitle>
                        <CardDescription>
                            Real-time JSON-RPC traces for {service.name}.
                        </CardDescription>
                    </div>
                    <div className="flex items-center gap-2">
                        <Button
                            variant="outline"
                            size="sm"
                            onClick={() => setIsPaused(!isPaused)}
                        >
                            {isPaused ? <Play className="h-4 w-4" /> : <Pause className="h-4 w-4" />}
                        </Button>
                        <Button variant="outline" size="sm" onClick={clearTraces}>
                            <Trash2 className="h-4 w-4" />
                        </Button>
                        <Button variant="outline" size="sm" onClick={refresh} disabled={loading && !isConnected}>
                            <RefreshCcw className={`h-4 w-4 ${loading && !isConnected ? 'animate-spin' : ''}`} />
                        </Button>
                    </div>
                </div>
            </CardHeader>
            <CardContent className="flex-1 min-h-0 p-0 flex flex-col">
                <div className="border-t flex-1 min-h-0">
                    <InspectorTable traces={filteredTraces} loading={loading && filteredTraces.length === 0} />
                </div>
            </CardContent>
        </Card>
    );
}
