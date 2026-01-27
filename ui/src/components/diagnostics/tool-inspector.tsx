/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect } from "react";
import { apiClient, ToolDefinition } from "@/lib/client";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Textarea } from "@/components/ui/textarea";
import { Play, Loader2, Wrench, AlertTriangle, CheckCircle2 } from "lucide-react";
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
    DialogTrigger,
} from "@/components/ui/dialog";
import { ScrollArea } from "@/components/ui/scroll-area";
import dynamic from "next/dynamic";

const JsonViewer = dynamic(() => import("@/components/logs/json-viewer"), { ssr: false });

interface ToolInspectorProps {
    serviceName: string;
}

export function ToolInspector({ serviceName }: ToolInspectorProps) {
    const [tools, setTools] = useState<ToolDefinition[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);

    const fetchTools = async () => {
        setLoading(true);
        setError(null);
        try {
            const res = await apiClient.listTools();
            // Filter tools for this service
            // Note: serviceId usually matches serviceName in simple configs, or we check both
            // Using type assertion for compatibility if apiClient returns slightly different shapes internally
            const serviceTools = (res.tools || []).filter((t: ToolDefinition & { service_id?: string }) =>
                t.serviceId === serviceName || t.service_id === serviceName
            );
            setTools(serviceTools);
        } catch (err) {
            console.error("Failed to fetch tools", err);
            setError("Failed to load tools.");
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchTools();
    }, [serviceName]);

    if (loading) {
        return <div className="flex justify-center p-8"><Loader2 className="animate-spin text-muted-foreground" /></div>;
    }

    if (error) {
        return <div className="text-red-500 p-4 border border-red-500/20 rounded bg-red-500/10">{error}</div>;
    }

    if (tools.length === 0) {
        return (
            <div className="flex flex-col items-center justify-center p-8 text-muted-foreground space-y-2">
                <Wrench className="h-8 w-8 opacity-20" />
                <p>No tools found for this service.</p>
                <p className="text-xs">Check if the service exposes tools and if 'tool_export_policy' allows them.</p>
            </div>
        );
    }

    return (
        <div className="space-y-4 p-4">
             <div className="rounded-md border bg-background/50">
                <Table>
                    <TableHeader>
                        <TableRow>
                            <TableHead>Tool Name</TableHead>
                            <TableHead>Description</TableHead>
                            <TableHead className="w-[100px]">Actions</TableHead>
                        </TableRow>
                    </TableHeader>
                    <TableBody>
                        {tools.map((tool) => (
                            <ToolRow key={tool.name} tool={tool} />
                        ))}
                    </TableBody>
                </Table>
             </div>
        </div>
    );
}

function ToolRow({ tool }: { tool: ToolDefinition }) {
    const [isOpen, setIsOpen] = useState(false);
    const [args, setArgs] = useState("{}");
    const [executing, setExecuting] = useState(false);
    const [result, setResult] = useState<{ success: boolean; data?: unknown; error?: string } | null>(null);

    const handleRun = async () => {
        setExecuting(true);
        setResult(null);
        try {
            let parsedArgs = {};
            try {
                parsedArgs = JSON.parse(args);
            } catch {
                throw new Error("Invalid JSON arguments");
            }

            const res = await apiClient.executeTool({
                name: tool.name,
                arguments: parsedArgs
            });
            setResult({ success: true, data: res });
        } catch (e: unknown) {
            const msg = e instanceof Error ? e.message : String(e);
            setResult({ success: false, error: msg });
        } finally {
            setExecuting(false);
        }
    };

    return (
        <TableRow>
            <TableCell className="font-medium font-mono text-xs">{tool.name}</TableCell>
            <TableCell className="text-sm text-muted-foreground line-clamp-2 max-w-[300px]" title={tool.description}>
                {tool.description || "-"}
            </TableCell>
            <TableCell>
                <Dialog open={isOpen} onOpenChange={setIsOpen}>
                    <DialogTrigger asChild>
                        <Button variant="outline" size="sm" className="h-7 text-xs">
                            <Play className="mr-2 h-3 w-3" /> Test
                        </Button>
                    </DialogTrigger>
                    <DialogContent className="sm:max-w-[600px] h-[80vh] flex flex-col">
                        <DialogHeader>
                            <DialogTitle>Test Tool: {tool.name}</DialogTitle>
                            <DialogDescription>
                                {tool.description}
                            </DialogDescription>
                        </DialogHeader>

                        <div className="flex-1 flex flex-col gap-4 overflow-hidden py-2">
                             <div className="space-y-2 flex-1 flex flex-col">
                                <label className="text-sm font-medium">Input Arguments (JSON)</label>
                                <div className="flex-1 border rounded-md overflow-hidden relative">
                                    <Textarea
                                        value={args}
                                        onChange={(e) => setArgs(e.target.value)}
                                        className="font-mono text-xs h-full w-full border-0 resize-none p-4 focus-visible:ring-0"
                                        placeholder="{}"
                                    />
                                </div>
                             </div>

                             {result && (
                                 <div className="space-y-2 flex-1 flex flex-col min-h-[30%]">
                                     <div className="flex items-center gap-2">
                                         <label className="text-sm font-medium">Result</label>
                                         {result.success ? (
                                             <Badge variant="outline" className="text-green-500 border-green-500/20 bg-green-500/10 text-[10px] gap-1">
                                                 <CheckCircle2 className="h-3 w-3" /> Success
                                             </Badge>
                                         ) : (
                                              <Badge variant="outline" className="text-red-500 border-red-500/20 bg-red-500/10 text-[10px] gap-1">
                                                 <AlertTriangle className="h-3 w-3" /> Error
                                             </Badge>
                                         )}
                                     </div>
                                     <div className="flex-1 rounded-md border bg-muted/30 overflow-hidden relative">
                                         <ScrollArea className="h-full w-full">
                                            <div className="p-4 text-xs">
                                                {result.success ? (
                                                     <JsonViewer data={result.data} />
                                                ) : (
                                                    <span className="text-red-500 font-mono whitespace-pre-wrap">{result.error}</span>
                                                )}
                                            </div>
                                         </ScrollArea>
                                     </div>
                                 </div>
                             )}
                        </div>

                        <DialogFooter>
                             <Button onClick={handleRun} disabled={executing}>
                                 {executing && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                                 Execute
                             </Button>
                        </DialogFooter>
                    </DialogContent>
                </Dialog>
            </TableCell>
        </TableRow>
    )
}
