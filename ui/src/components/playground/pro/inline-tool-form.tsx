/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useState, useMemo } from "react";
import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";
import { ToolDefinition } from "@/lib/client";
import { Zap, X, PlayCircle, Loader2 } from "lucide-react";
import { cn } from "@/lib/utils";
import { Badge } from "@/components/ui/badge";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { UniversalSchemaForm as SchemaForm, Schema } from "@/components/shared/universal-schema-form";

export interface InlineToolFormProps {
    tool: ToolDefinition;
    onSubmit: (toolName: string, args: Record<string, unknown>) => void;
    onCancel: () => void;
    isLoading?: boolean;
}

export function InlineToolForm({ tool, onSubmit, onCancel, isLoading }: InlineToolFormProps) {
    const [input, setInput] = useState("{}");
    const [activeTab, setActiveTab] = useState<"form" | "json">("form");

    const parsedInput = useMemo(() => {
        try {
            return JSON.parse(input);
        } catch {
            return undefined;
        }
    }, [input]);

    const handleSubmit = () => {
        let args = {};
        if (input.trim()) {
            try {
                args = JSON.parse(input);
            } catch (e) {
                // If they are on JSON tab and it's invalid, don't submit
                if (activeTab === "json") return;
            }
        }
        onSubmit(tool.name, args);
    };

    return (
        <div className="w-full max-w-4xl mx-auto my-4 overflow-hidden rounded-xl border border-primary/20 bg-background/80 backdrop-blur-xl shadow-lg animate-in slide-in-from-bottom-4 fade-in duration-300">
            <div className="flex items-center justify-between px-4 py-3 border-b bg-muted/30">
                <div className="flex items-center gap-3">
                    <div className="bg-primary/10 p-2 rounded-md">
                        <Zap className="h-4 w-4 text-primary" />
                    </div>
                    <div>
                        <h3 className="text-sm font-semibold flex items-center gap-2">
                            Configure {tool.name}
                            <Badge variant="outline" className="font-normal text-[10px] h-4 px-1.5 text-muted-foreground bg-background">
                                {tool.serviceId || 'core'}
                            </Badge>
                        </h3>
                        {tool.description && (
                            <p className="text-xs text-muted-foreground line-clamp-1 max-w-md mt-0.5" title={tool.description}>
                                {tool.description}
                            </p>
                        )}
                    </div>
                </div>
                <Button variant="ghost" size="icon" onClick={onCancel} className="h-8 w-8 text-muted-foreground hover:text-foreground">
                    <X className="h-4 w-4" />
                </Button>
            </div>

            <div className="p-4 max-h-[60vh] overflow-y-auto">
                <Tabs value={activeTab} onValueChange={(v) => setActiveTab(v as any)} className="w-full">
                    <TabsList className="grid w-[160px] grid-cols-2 h-8 mb-4">
                        <TabsTrigger value="form" className="text-xs">Form</TabsTrigger>
                        <TabsTrigger value="json" className="text-xs">JSON</TabsTrigger>
                    </TabsList>

                    <TabsContent value="form" className="mt-0">
                        {parsedInput === undefined ? (
                            <div className="text-destructive text-xs p-3 border border-destructive/50 rounded bg-destructive/10">
                                Invalid JSON arguments. Please fix errors in JSON view to use the visual builder.
                            </div>
                        ) : (
                            <SchemaForm
                                schema={(tool.inputSchema as Schema) || { type: "object" }}
                                value={parsedInput}
                                onChange={(v) => setInput(JSON.stringify(v, null, 2))}
                            />
                        )}
                    </TabsContent>

                    <TabsContent value="json" className="mt-0">
                        <Textarea
                            value={input}
                            onChange={(e) => setInput(e.target.value)}
                            className={cn(
                                "font-mono text-xs min-h-[150px] bg-muted/20",
                                parsedInput === undefined && "border-destructive focus-visible:ring-destructive"
                            )}
                            placeholder="{}"
                        />
                        {parsedInput === undefined && (
                            <p className="text-[10px] text-destructive mt-1 font-medium">Invalid JSON format</p>
                        )}
                    </TabsContent>
                </Tabs>
            </div>

            <div className="px-4 py-3 border-t bg-muted/10 flex items-center justify-between">
                <p className="text-[10px] text-muted-foreground flex items-center gap-1">
                    Press <kbd className="px-1.5 py-0.5 rounded border bg-muted font-sans font-medium text-[9px]">Esc</kbd> to cancel
                </p>
                <div className="flex items-center gap-2">
                    <Button variant="ghost" size="sm" onClick={onCancel} className="h-8 text-xs">
                        Cancel
                    </Button>
                    <Button
                        size="sm"
                        onClick={handleSubmit}
                        disabled={isLoading || (activeTab === "json" && parsedInput === undefined)}
                        className="h-8 text-xs gap-1.5"
                    >
                        {isLoading ? <Loader2 className="h-3.5 w-3.5 animate-spin" /> : <PlayCircle className="h-3.5 w-3.5" />}
                        Execute Tool
                    </Button>
                </div>
            </div>
        </div>
    );
}
