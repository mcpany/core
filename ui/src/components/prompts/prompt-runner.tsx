/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState } from "react";
import { PromptDefinition, apiClient } from "@/lib/client";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Card, CardContent } from "@/components/ui/card";
import { Play, Loader2, Sparkles, Copy, ExternalLink, RefreshCw } from "lucide-react";
import { useToast } from "@/hooks/use-toast";
import { cn } from "@/lib/utils";
import { useRouter } from "next/navigation";

interface PromptRunnerProps {
    prompt: PromptDefinition;
}

export function PromptRunner({ prompt }: PromptRunnerProps) {
    const [argumentValues, setArgumentValues] = useState<Record<string, string>>({});
    const [executionResult, setExecutionResult] = useState<any | null>(null);
    const [isLoading, setIsLoading] = useState(false);
    const { toast } = useToast();
    const router = useRouter();

    const getArguments = () => {
        if (!prompt.inputSchema || !prompt.inputSchema.properties) return [];
        // Handle google.protobuf.Struct or plain JSON object
        const props = (prompt.inputSchema as any).fields?.properties?.structValue?.fields ||
                      (prompt.inputSchema as any).properties ||
                      {};
        const required = (prompt.inputSchema as any).required || []; // Simplify required check

        return Object.entries(props).map(([key, val]: [string, any]) => {
             // Handle Proto Struct Value format if necessary, or assume mapped JSON
             const description = val.description || val.structValue?.fields?.description?.stringValue;
             return {
                name: key,
                description: description,
                required: required.includes(key),
             };
        });
    };

    const handleExecute = async () => {
        setIsLoading(true);
        try {
            const result = await apiClient.executePrompt(prompt.name, argumentValues);
            setExecutionResult(result);
            toast({
                title: "Executed",
                description: "Prompt generated successfully.",
            });
        } catch (e: any) {
            console.error("Execution failed", e);
            toast({
                variant: "destructive",
                title: "Error",
                description: e.message || "Failed to execute prompt."
            });
        } finally {
            setIsLoading(false);
        }
    };

    const copyToClipboard = () => {
        if (executionResult) {
            navigator.clipboard.writeText(JSON.stringify(executionResult, null, 2));
            toast({ title: "Copied", description: "Result copied to clipboard." });
        }
    };

    return (
        <div className="flex flex-col h-full bg-background">
            <div className="p-6 border-b pb-4 shrink-0">
                <h2 className="text-2xl font-bold tracking-tight flex items-center gap-2">
                    <Play className="h-5 w-5 text-primary" /> Run: {prompt.name}
                </h2>
                <p className="text-muted-foreground mt-1">
                    Execute the prompt to preview generated messages.
                </p>
            </div>

            <div className="flex-1 overflow-y-auto p-6">
                <div className="grid grid-cols-1 lg:grid-cols-2 gap-8 h-full">
                    {/* Input Column */}
                    <div className="flex flex-col gap-6">
                        <Card>
                            <CardContent className="p-4 space-y-4">
                                <h3 className="font-semibold text-sm">Arguments</h3>
                                {getArguments().length > 0 ? (
                                    getArguments().map((arg) => (
                                        <div key={arg.name} className="space-y-1.5">
                                            <Label htmlFor={arg.name} className="flex items-center gap-1 text-xs font-mono uppercase text-muted-foreground">
                                                {arg.name}
                                                {arg.required && <span className="text-red-500">*</span>}
                                            </Label>
                                            <Input
                                                id={arg.name}
                                                placeholder={arg.description || ""}
                                                value={argumentValues[arg.name] || ""}
                                                onChange={(e) => setArgumentValues({...argumentValues, [arg.name]: e.target.value})}
                                            />
                                            {arg.description && <p className="text-[10px] text-muted-foreground">{arg.description}</p>}
                                        </div>
                                    ))
                                ) : (
                                    <div className="text-sm text-muted-foreground italic text-center py-4">
                                        No arguments required.
                                    </div>
                                )}
                                <Button className="w-full mt-4" onClick={handleExecute} disabled={isLoading}>
                                    {isLoading ? <><Loader2 className="mr-2 h-4 w-4 animate-spin" /> Generating...</> : <><Play className="mr-2 h-4 w-4" /> Generate</>}
                                </Button>
                            </CardContent>
                        </Card>
                    </div>

                    {/* Output Column */}
                    <div className="flex flex-col h-full min-h-[400px]">
                        <div className="flex items-center justify-between mb-4">
                            <h3 className="text-sm font-medium flex items-center gap-2 text-primary">
                                <Sparkles className="h-4 w-4" /> Output Preview
                            </h3>
                            <Button variant="ghost" size="sm" onClick={copyToClipboard} disabled={!executionResult}>
                                <Copy className="h-3 w-3" />
                            </Button>
                        </div>
                        <Card className="flex-1 flex flex-col overflow-hidden bg-muted/30 border-dashed">
                            <CardContent className="flex-1 p-0 overflow-auto">
                                {executionResult ? (
                                    <div className="p-4 space-y-4">
                                        {(executionResult?.messages || []).map((msg: any, idx: number) => (
                                            <div key={idx} className="space-y-1">
                                                <div className="text-[10px] font-mono uppercase text-muted-foreground flex items-center gap-2">
                                                    <span className={cn(
                                                        "w-2 h-2 rounded-full",
                                                        msg.role === "user" ? "bg-blue-500" : "bg-green-500"
                                                    )} />
                                                    {msg.role}
                                                </div>
                                                <div className="bg-background border rounded-md p-3 text-sm whitespace-pre-wrap font-mono">
                                                    {msg.content?.type === 'text' ? msg.content.text : typeof msg.content === 'string' ? msg.content : JSON.stringify(msg.content)}
                                                </div>
                                            </div>
                                        ))}
                                    </div>
                                ) : (
                                    <div className="flex flex-col items-center justify-center h-full text-muted-foreground text-sm p-8 text-center">
                                        <Sparkles className="h-10 w-10 opacity-20 mb-3" />
                                        <p>Configure arguments and click Generate to see the prompt result.</p>
                                    </div>
                                )}
                            </CardContent>
                        </Card>
                    </div>
                </div>
            </div>
        </div>
    );
}
