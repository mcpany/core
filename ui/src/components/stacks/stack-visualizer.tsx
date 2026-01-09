/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import React, { useMemo } from "react";
import jsyaml from "js-yaml";
import {
    AlertTriangle,
    Box,
    Layers,
    Terminal,
    Globe,
    Database,
    Cpu,
    CheckCircle2,
    Settings2
} from "lucide-react";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";

interface StackVisualizerProps {
    yamlContent: string;
}

interface ParsedService {
    name: string;
    image?: string;
    command?: string;
    envCount: number;
    ports: string[];
    type: "image" | "command" | "unknown";
}

export function StackVisualizer({ yamlContent }: StackVisualizerProps) {
    const { services, error } = useMemo(() => {
        try {
            const parsed = jsyaml.load(yamlContent) as any;
            if (!parsed || typeof parsed !== 'object') {
                return { services: [], error: null };
            }

            const rawServices = parsed.services || {};
            const serviceList: ParsedService[] = Object.entries(rawServices).map(([key, val]: [string, any]) => {
                const env = val.environment || val.env || {};
                const envCount = Array.isArray(env) ? env.length : Object.keys(env).length;
                const ports = val.ports || [];

                let type: ParsedService['type'] = "unknown";
                if (val.image) type = "image";
                else if (val.command) type = "command";

                return {
                    name: key,
                    image: val.image,
                    command: val.command,
                    envCount,
                    ports,
                    type
                };
            });

            return { services: serviceList, error: null };
        } catch (e: any) {
            return { services: [], error: e.message };
        }
    }, [yamlContent]);

    if (error) {
        return (
            <div className="flex flex-col items-center justify-center h-full text-muted-foreground p-4 gap-2 bg-red-50/10">
                <AlertTriangle className="h-8 w-8 text-destructive opacity-50" />
                <p className="text-xs text-destructive font-medium text-center">YAML Syntax Error</p>
                <p className="text-[10px] font-mono opacity-75 max-w-[200px] break-all text-center">{error}</p>
            </div>
        );
    }

    if (services.length === 0) {
        return (
            <div className="flex flex-col items-center justify-center h-full text-muted-foreground p-4 gap-2">
                <Box className="h-8 w-8 opacity-20" />
                <p className="text-xs">No services defined</p>
            </div>
        );
    }

    return (
        <div className="flex flex-col h-full bg-muted/5 w-[280px] border-l stack-visualizer-container">
            <div className="p-4 border-b flex items-center justify-between">
                <div className="flex items-center gap-2 font-semibold">
                    <Layers className="h-5 w-5 text-primary" />
                    <span>Live Preview</span>
                </div>
                <Badge variant="outline" className="text-[10px]">{services.length} services</Badge>
            </div>
            <ScrollArea className="flex-1 p-4">
                <div className="space-y-3">
                    {services.map(svc => (
                        <Card key={svc.name} className="overflow-hidden border-l-4 border-l-primary/50 shadow-sm" role="article" aria-label={`Service: ${svc.name}`}>
                            <CardHeader className="p-3 pb-2 bg-muted/20">
                                <CardTitle className="text-sm font-medium flex items-center gap-2">
                                    <ServiceIcon type={svc.type} />
                                    <span className="truncate" title={svc.name}>{svc.name}</span>
                                </CardTitle>
                            </CardHeader>
                            <CardContent className="p-3 pt-2 space-y-2">
                                {svc.image && (
                                    <div className="flex items-start gap-2 text-[10px] text-muted-foreground">
                                        <Database className="h-3 w-3 mt-0.5" />
                                        <span className="font-mono break-all">{svc.image}</span>
                                    </div>
                                )}
                                {svc.command && (
                                    <div className="flex items-start gap-2 text-[10px] text-muted-foreground">
                                        <Terminal className="h-3 w-3 mt-0.5" />
                                        <span className="font-mono break-all line-clamp-2">{svc.command}</span>
                                    </div>
                                )}
                                <div className="flex items-center gap-2 mt-2">
                                    {svc.envCount > 0 && (
                                        <Badge variant="secondary" className="text-[9px] px-1 h-4 flex gap-1">
                                            <Settings2 className="h-2 w-2" /> {svc.envCount} Env
                                        </Badge>
                                    )}
                                    {svc.ports.length > 0 && (
                                        <Badge variant="secondary" className="text-[9px] px-1 h-4 flex gap-1">
                                            <Globe className="h-2 w-2" /> {svc.ports.length} Port
                                        </Badge>
                                    )}
                                </div>
                                <div className="mt-2 flex items-center gap-1 text-[10px] text-green-600 dark:text-green-400">
                                    <CheckCircle2 className="h-3 w-3" />
                                    <span>Valid Configuration</span>
                                </div>
                            </CardContent>
                        </Card>
                    ))}
                </div>
            </ScrollArea>
        </div>
    );
}

function ServiceIcon({ type }: { type: ParsedService['type'] }) {
    switch (type) {
        case "image": return <Database className="h-4 w-4 text-indigo-500" />;
        case "command": return <Terminal className="h-4 w-4 text-amber-500" />;
        default: return <Cpu className="h-4 w-4 text-gray-500" />;
    }
}
