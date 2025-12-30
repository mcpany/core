/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import * as React from "react";
import { ToolDefinition } from "@/lib/client";
import {
    Sheet,
    SheetContent,
    SheetDescription,
    SheetHeader,
    SheetTitle,
} from "@/components/ui/sheet";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { Switch } from "@/components/ui/switch";
import { Badge } from "@/components/ui/badge";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Play, Loader2, Braces, Terminal } from "lucide-react";
import { cn } from "@/lib/utils";

interface ToolInspectorProps {
    tool: ToolDefinition | null;
    open: boolean;
    onOpenChange: (open: boolean) => void;
}

export function ToolInspector({ tool, open, onOpenChange }: ToolInspectorProps) {
    const [activeTab, setActiveTab] = React.useState("test");
    const [args, setArgs] = React.useState<Record<string, any>>({});
    const [loading, setLoading] = React.useState(false);
    const [result, setResult] = React.useState<any>(null);

    React.useEffect(() => {
        if (open && tool) {
            setArgs({});
            setResult(null);
            // Initialize default values based on schema if needed
        }
    }, [open, tool]);

    const handleArgChange = (key: string, value: any) => {
        setArgs(prev => ({ ...prev, [key]: value }));
    };

    const handleExecute = async () => {
        if (!tool) return;
        setLoading(true);
        setResult(null);

        try {
            const res = await fetch('/api/tools', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    action: 'execute',
                    name: tool.name,
                    arguments: args
                })
            });
            const data = await res.json();
            setResult(data);
        } catch (e) {
            setResult({ error: "Failed to execute tool" });
        } finally {
            setLoading(false);
        }
    };

    if (!tool) return null;

    return (
        <Sheet open={open} onOpenChange={onOpenChange}>
            <SheetContent className="sm:max-w-[600px] flex flex-col h-full p-0">
                <SheetHeader className="p-6 border-b">
                    <SheetTitle className="flex items-center gap-2">
                        <Terminal className="h-5 w-5 text-primary" />
                        {tool.name}
                    </SheetTitle>
                    <SheetDescription>
                        {tool.description}
                    </SheetDescription>
                    <div className="flex gap-2 mt-2">
                         <Badge variant="secondary">{tool.serviceName}</Badge>
                         {tool.enabled ? <Badge variant="default" className="bg-green-600">Enabled</Badge> : <Badge variant="destructive">Disabled</Badge>}
                    </div>
                </SheetHeader>

                <Tabs value={activeTab} onValueChange={setActiveTab} className="flex-1 flex flex-col overflow-hidden">
                    <div className="px-6 pt-4">
                        <TabsList className="grid w-full grid-cols-2">
                            <TabsTrigger value="test">Test Tool</TabsTrigger>
                            <TabsTrigger value="schema">Schema Definition</TabsTrigger>
                        </TabsList>
                    </div>

                    <TabsContent value="test" className="flex-1 overflow-hidden flex flex-col p-6 space-y-4 pt-4">
                        <ScrollArea className="flex-1 pr-4">
                            <div className="space-y-6">
                                <div className="space-y-4">
                                    <h3 className="text-sm font-medium text-muted-foreground uppercase tracking-wider">Arguments</h3>
                                    {tool.schema?.properties ? (
                                        Object.entries(tool.schema.properties).map(([key, prop]: [string, any]) => (
                                            <div key={key} className="space-y-2">
                                                <Label htmlFor={key} className="flex items-center gap-2">
                                                    {key}
                                                    {tool.schema?.required?.includes(key) && <span className="text-red-500">*</span>}
                                                    <span className="text-xs font-normal text-muted-foreground">({prop.type})</span>
                                                </Label>
                                                <p className="text-xs text-muted-foreground">{prop.description}</p>

                                                {prop.enum ? (
                                                     <Select onValueChange={(v) => handleArgChange(key, v)}>
                                                        <SelectTrigger>
                                                            <SelectValue placeholder={`Select ${key}`} />
                                                        </SelectTrigger>
                                                        <SelectContent>
                                                            {prop.enum.map((opt: string) => (
                                                                <SelectItem key={opt} value={opt}>{opt}</SelectItem>
                                                            ))}
                                                        </SelectContent>
                                                    </Select>
                                                ) : prop.type === 'boolean' ? (
                                                    <div className="flex items-center space-x-2">
                                                        <Switch id={key} onCheckedChange={(checked) => handleArgChange(key, checked)} />
                                                        <span className="text-sm text-muted-foreground">{args[key] ? "True" : "False"}</span>
                                                    </div>
                                                ) : prop.type === 'integer' || prop.type === 'number' ? (
                                                     <Input
                                                        id={key}
                                                        type="number"
                                                        placeholder={`Enter ${key}`}
                                                        onChange={(e) => handleArgChange(key, Number(e.target.value))}
                                                    />
                                                ) : (
                                                    <Input
                                                        id={key}
                                                        placeholder={`Enter ${key}`}
                                                        onChange={(e) => handleArgChange(key, e.target.value)}
                                                    />
                                                )}
                                            </div>
                                        ))
                                    ) : (
                                        <div className="text-sm text-muted-foreground italic">No arguments required.</div>
                                    )}
                                </div>
                            </div>
                        </ScrollArea>

                        <div className="space-y-4 pt-4 border-t">
                            <Button className="w-full" onClick={handleExecute} disabled={loading || !tool.enabled}>
                                {loading ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <Play className="mr-2 h-4 w-4" />}
                                Execute Tool
                            </Button>

                            {result && (
                                <div className="space-y-2 animate-in fade-in slide-in-from-bottom-2">
                                    <div className="flex items-center justify-between">
                                         <h3 className="text-sm font-medium text-muted-foreground uppercase tracking-wider">Result</h3>
                                         <Badge variant={result.error ? "destructive" : "outline"} className={cn(result.error ? "" : "text-green-600 border-green-600/30 bg-green-500/10")}>
                                            {result.error ? "Error" : "Success"}
                                         </Badge>
                                    </div>
                                    <div className="relative rounded-md border bg-muted/50 p-4">
                                        <pre className="text-xs font-mono overflow-auto max-h-[200px] whitespace-pre-wrap">
                                            {JSON.stringify(result, null, 2)}
                                        </pre>
                                    </div>
                                </div>
                            )}
                        </div>
                    </TabsContent>

                    <TabsContent value="schema" className="flex-1 overflow-hidden p-6 pt-4">
                        <div className="h-full rounded-md border bg-muted/30 p-4 font-mono text-xs overflow-auto">
                            <pre>{JSON.stringify(tool.schema, null, 2)}</pre>
                        </div>
                    </TabsContent>
                </Tabs>
            </SheetContent>
        </Sheet>
    );
}
