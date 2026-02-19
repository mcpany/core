/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect, useMemo } from "react";
import { ToolDefinition } from "@proto/config/v1/tool";
import { HttpCallDefinition, HttpCallDefinition_HttpMethod, HttpParameterMapping, ParameterType, InputTransformer, OutputTransformer } from "@proto/config/v1/call";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Button } from "@/components/ui/button";
import { Trash2, Plus, Play, Loader2 } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Switch } from "@/components/ui/switch";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { InputTransformerEditor } from "./input-transformer-editor";
import { OutputTransformerEditor } from "./output-transformer-editor";
import { Textarea } from "@/components/ui/textarea";
import { ResizableHandle, ResizablePanel, ResizablePanelGroup } from "@/components/ui/resizable";
import { ScrollArea } from "@/components/ui/scroll-area";
import { apiClient } from "@/lib/client";
import { useToast } from "@/hooks/use-toast";

interface HttpToolEditorProps {
    tool: ToolDefinition;
    call: HttpCallDefinition;
    onChange: (tool: ToolDefinition, call: HttpCallDefinition) => void;
}

/**
 * Editor for configuring a single HTTP tool.
 * Allows defining tool metadata and the mapped HTTP request details.
 * @param props - The component props.
 * @returns The rendered tool editor.
 */
export function HttpToolEditor({ tool, call, onChange }: HttpToolEditorProps) {
    const [localTool, setLocalTool] = useState<ToolDefinition>(tool);
    const [localCall, setLocalCall] = useState<HttpCallDefinition>(call);
    const [activeTab, setActiveTab] = useState("request");

    // Test & Preview State
    const [testArguments, setTestArguments] = useState("{\n  \n}");
    const [previewResult, setPreviewResult] = useState<any>(null);
    const [isExecuting, setIsExecuting] = useState(false);
    const { toast } = useToast();

    useEffect(() => {
        setLocalTool(tool);
        setLocalCall(call);
    }, [tool, call]);

    const updateTool = (updates: Partial<ToolDefinition>) => {
        const newTool = { ...localTool, ...updates };
        setLocalTool(newTool);
        onChange(newTool, localCall);
    };

    const updateCall = (updates: Partial<HttpCallDefinition>) => {
        const newCall = { ...localCall, ...updates };
        setLocalCall(newCall);
        onChange(localTool, newCall);
    };

    const addParameter = () => {
        const newParams = [
            ...(localCall.parameters || []),
            {
                schema: {
                    name: "",
                    description: "",
                    type: ParameterType.STRING,
                    isRequired: true,
                },
                disableEscape: false,
            } as HttpParameterMapping
        ];
        updateCall({ parameters: newParams });
    };

    const updateParameterSchema = (index: number, updates: any) => {
        const newParams = [...(localCall.parameters || [])];
        newParams[index] = {
            ...newParams[index],
            schema: { ...newParams[index].schema!, ...updates }
        };
        updateCall({ parameters: newParams });
    };

    const removeParameter = (index: number) => {
        const newParams = [...(localCall.parameters || [])];
        newParams.splice(index, 1);
        updateCall({ parameters: newParams });
    };

    const handleInputTransformerChange = (transformer: InputTransformer) => {
        updateCall({ inputTransformer: transformer });
    };

    const handleOutputTransformerChange = (transformer: OutputTransformer) => {
        updateCall({ outputTransformer: transformer });
    };

    // Calculate Preview
    const previewData = useMemo(() => {
        let method = "GET";
        switch (localCall.method) {
            case HttpCallDefinition_HttpMethod.HTTP_METHOD_POST: method = "POST"; break;
            case HttpCallDefinition_HttpMethod.HTTP_METHOD_PUT: method = "PUT"; break;
            case HttpCallDefinition_HttpMethod.HTTP_METHOD_DELETE: method = "DELETE"; break;
            case HttpCallDefinition_HttpMethod.HTTP_METHOD_PATCH: method = "PATCH"; break;
        }

        let url = localCall.endpointPath || "/";
        let args: Record<string, any> = {};
        try {
            args = JSON.parse(testArguments);
        } catch {
            // Ignore parse errors for preview
        }

        // 1. Path Replacement
        for (const key in args) {
            // Support both {param} and {{param}} styles in preview
            url = url.replace(`{${key}}`, String(args[key]));
            url = url.replace(`{{${key}}}`, String(args[key]));
        }

        // 2. Query Params (Simulation)
        // In reality, logic is driven by parameter definitions, but for preview we can infer or just show args
        const queryParams = new URLSearchParams();
        // If we want accurate preview, we should look at 'in' or just assume remaining args might be query?
        // The current editor doesn't explicitly set 'in' (Path/Query) for parameters,
        // the backend logic infers it from path templates.
        // So any arg NOT used in path is candidate for query/body.

        // Let's iterate defined parameters
        const bodyArgs: Record<string, any> = {};

        (localCall.parameters || []).forEach(p => {
            const name = p.schema?.name;
            if (!name || args[name] === undefined) return;

            // Check if used in path
            const inPath = localCall.endpointPath.includes(`{${name}}`) || localCall.endpointPath.includes(`{{${name}}}`);

            if (!inPath) {
                if (method === "GET" || method === "DELETE") {
                    queryParams.append(name, String(args[name]));
                } else {
                    bodyArgs[name] = args[name];
                }
            }
        });

        const queryString = queryParams.toString();
        const fullUrl = queryString ? `${url}?${queryString}` : url;

        return {
            method,
            url: fullUrl,
            body: Object.keys(bodyArgs).length > 0 ? JSON.stringify(bodyArgs, null, 2) : undefined
        };
    }, [localCall, testArguments]);

    const handleExecute = async () => {
        // We need serviceId to execute.
        // If it's a new tool in an existing service, serviceId is populated.
        // If it's a completely new service (unsaved), serviceId might be empty.
        if (!localTool.serviceId) {
            toast({
                title: "Cannot Execute",
                description: "Please save the service first to register the tool.",
                variant: "destructive"
            });
            return;
        }

        setIsExecuting(true);
        setPreviewResult(null);
        try {
            let args = {};
            try { args = JSON.parse(testArguments); } catch { throw new Error("Invalid JSON arguments"); }

            // Construct fully qualified name used by backend
            const toolName = `${localTool.serviceId}.${localTool.name}`;

            const res = await apiClient.executeTool({
                name: toolName,
                arguments: args
            });
            setPreviewResult(res);
        } catch (e: any) {
            const msg = e instanceof Error ? e.message : (typeof e === 'string' ? e : JSON.stringify(e));
            setPreviewResult({ error: msg });
        } finally {
            setIsExecuting(false);
        }
    };

    return (
        <ResizablePanelGroup direction="horizontal" className="h-[600px] border rounded-lg bg-background">
            <ResizablePanel defaultSize={60} minSize={30}>
                <ScrollArea className="h-full">
                    <div className="p-6 space-y-6">
                        <div className="space-y-4">
                            <div className="grid grid-cols-2 gap-4">
                                <div className="space-y-2">
                                    <Label htmlFor="tool-name">Tool Name</Label>
                                    <Input
                                        id="tool-name"
                                        value={localTool.name}
                                        onChange={(e) => updateTool({ name: e.target.value })}
                                        placeholder="get_weather"
                                    />
                                </div>
                                <div className="space-y-2">
                                    <Label htmlFor="tool-description">Description</Label>
                                    <Input
                                        id="tool-description"
                                        value={localTool.description}
                                        onChange={(e) => updateTool({ description: e.target.value })}
                                        placeholder="Get the weather for a location"
                                    />
                                </div>
                            </div>

                            <div className="grid grid-cols-3 gap-4">
                                <div className="space-y-2">
                                    <Label htmlFor="http-method">Method</Label>
                                    <Select
                                        value={localCall.method.toString()}
                                        onValueChange={(val) => updateCall({ method: parseInt(val) })}
                                    >
                                        <SelectTrigger id="http-method">
                                            <SelectValue placeholder="Method" />
                                        </SelectTrigger>
                                        <SelectContent>
                                            <SelectItem value={HttpCallDefinition_HttpMethod.HTTP_METHOD_GET.toString()}>GET</SelectItem>
                                            <SelectItem value={HttpCallDefinition_HttpMethod.HTTP_METHOD_POST.toString()}>POST</SelectItem>
                                            <SelectItem value={HttpCallDefinition_HttpMethod.HTTP_METHOD_PUT.toString()}>PUT</SelectItem>
                                            <SelectItem value={HttpCallDefinition_HttpMethod.HTTP_METHOD_DELETE.toString()}>DELETE</SelectItem>
                                            <SelectItem value={HttpCallDefinition_HttpMethod.HTTP_METHOD_PATCH.toString()}>PATCH</SelectItem>
                                        </SelectContent>
                                    </Select>
                                </div>
                                <div className="col-span-2 space-y-2">
                                    <Label htmlFor="endpoint-path">Endpoint Path</Label>
                                    <Input
                                        id="endpoint-path"
                                        value={localCall.endpointPath}
                                        onChange={(e) => updateCall({ endpointPath: e.target.value })}
                                        placeholder="/users/{userId}"
                                    />
                                    <p className="text-[10px] text-muted-foreground">Use <code>{"{paramName}"}</code> for path parameters.</p>
                                </div>
                            </div>
                        </div>

                        <Tabs value={activeTab} onValueChange={setActiveTab} className="w-full">
                            <TabsList className="w-full justify-start">
                                <TabsTrigger value="request">Request Parameters</TabsTrigger>
                                <TabsTrigger value="input-transform">Input Transform</TabsTrigger>
                                <TabsTrigger value="output-transform">Output Transform</TabsTrigger>
                            </TabsList>

                            <TabsContent value="request" className="space-y-4 mt-4">
                                <div className="flex items-center justify-between">
                                    <Label className="text-base">Parameters</Label>
                                    <Button variant="outline" size="sm" onClick={addParameter}>
                                        <Plus className="mr-2 h-4 w-4" /> Add Parameter
                                    </Button>
                                </div>

                                {localCall.parameters?.map((param, index) => (
                                    <Card key={index} className="relative">
                                        <Button
                                            variant="ghost"
                                            size="icon"
                                            className="absolute right-2 top-2 h-6 w-6 text-muted-foreground hover:text-destructive"
                                            onClick={() => removeParameter(index)}
                                        >
                                            <Trash2 className="h-4 w-4" />
                                        </Button>
                                        <CardContent className="p-4 grid grid-cols-12 gap-4">
                                            <div className="col-span-3 space-y-2">
                                                <Label htmlFor={`param-name-${index}`}>Name</Label>
                                                <Input
                                                    id={`param-name-${index}`}
                                                    value={param.schema?.name}
                                                    onChange={(e) => updateParameterSchema(index, { name: e.target.value })}
                                                    placeholder="userId"
                                                />
                                            </div>
                                            <div className="col-span-3 space-y-2">
                                                <Label htmlFor={`param-type-${index}`}>Type</Label>
                                                <Select
                                                    value={param.schema?.type.toString()}
                                                    onValueChange={(val) => updateParameterSchema(index, { type: parseInt(val) })}
                                                >
                                                    <SelectTrigger id={`param-type-${index}`}>
                                                        <SelectValue />
                                                    </SelectTrigger>
                                                    <SelectContent>
                                                        <SelectItem value={ParameterType.STRING.toString()}>String</SelectItem>
                                                        <SelectItem value={ParameterType.NUMBER.toString()}>Number</SelectItem>
                                                        <SelectItem value={ParameterType.INTEGER.toString()}>Integer</SelectItem>
                                                        <SelectItem value={ParameterType.BOOLEAN.toString()}>Boolean</SelectItem>
                                                    </SelectContent>
                                                </Select>
                                            </div>
                                            <div className="col-span-4 space-y-2">
                                                <Label htmlFor={`param-desc-${index}`}>Description</Label>
                                                <Input
                                                    id={`param-desc-${index}`}
                                                    value={param.schema?.description}
                                                    onChange={(e) => updateParameterSchema(index, { description: e.target.value })}
                                                    placeholder="The ID of the user"
                                                />
                                            </div>
                                            <div className="col-span-2 flex items-center justify-center pt-8 space-x-2">
                                                <Switch
                                                    checked={param.schema?.isRequired}
                                                    onCheckedChange={(checked) => updateParameterSchema(index, { isRequired: checked })}
                                                />
                                                <Label className="text-xs">Required</Label>
                                            </div>
                                        </CardContent>
                                    </Card>
                                ))}
                                {(!localCall.parameters || localCall.parameters.length === 0) && (
                                    <div className="text-center py-8 border border-dashed rounded-md text-muted-foreground text-sm">
                                        No parameters defined.
                                    </div>
                                )}
                            </TabsContent>

                            <TabsContent value="input-transform" className="mt-4">
                                <InputTransformerEditor
                                    transformer={localCall.inputTransformer}
                                    onChange={handleInputTransformerChange}
                                />
                            </TabsContent>

                            <TabsContent value="output-transform" className="mt-4">
                                <OutputTransformerEditor
                                    transformer={localCall.outputTransformer}
                                    onChange={handleOutputTransformerChange}
                                />
                            </TabsContent>
                        </Tabs>
                    </div>
                </ScrollArea>
            </ResizablePanel>

            <ResizableHandle />

            <ResizablePanel defaultSize={40} minSize={30} className="bg-muted/10">
                <div className="flex flex-col h-full p-4 gap-4">
                    <h3 className="font-medium flex items-center gap-2">
                        <Play className="h-4 w-4" /> Live Preview & Test
                    </h3>

                    <Card className="flex-shrink-0">
                        <CardHeader className="py-3 px-4 bg-muted/20 border-b">
                            <CardTitle className="text-xs font-mono uppercase tracking-wider text-muted-foreground">Request Preview</CardTitle>
                        </CardHeader>
                        <CardContent className="p-4 font-mono text-xs space-y-2" data-testid="request-preview">
                            <div className="flex gap-2 items-center">
                                <span className="font-bold text-primary">{previewData.method}</span>
                                <span className="break-all">{previewData.url}</span>
                            </div>
                            {previewData.body && (
                                <div className="mt-2 pt-2 border-t border-dashed">
                                    <div className="text-muted-foreground mb-1">Body:</div>
                                    <pre className="text-muted-foreground whitespace-pre-wrap">{previewData.body}</pre>
                                </div>
                            )}
                        </CardContent>
                    </Card>

                    <div className="flex-1 flex flex-col min-h-0 gap-2">
                        <Label htmlFor="test-arguments" className="text-xs">Test Arguments (JSON)</Label>
                        <Textarea
                            id="test-arguments"
                            className="font-mono text-xs flex-1 resize-none bg-background"
                            value={testArguments}
                            onChange={e => setTestArguments(e.target.value)}
                            placeholder='{ "param": "value" }'
                        />
                    </div>

                    <div className="flex-shrink-0 space-y-2">
                        <Button
                            className="w-full"
                            onClick={handleExecute}
                            disabled={isExecuting || !localTool.serviceId}
                            title={!localTool.serviceId ? "Save service to execute" : "Execute tool"}
                        >
                            {isExecuting ? <Loader2 className="h-4 w-4 animate-spin mr-2"/> : <Play className="h-4 w-4 mr-2"/>}
                            Execute
                        </Button>

                        {previewResult && (
                            <div className="rounded-md border bg-zinc-950 text-zinc-50 p-3 font-mono text-xs h-[150px] overflow-auto whitespace-pre-wrap shadow-inner">
                                {previewResult.error ? (
                                    <span className="text-red-400">Error: {previewResult.error}</span>
                                ) : (
                                    JSON.stringify(previewResult, null, 2)
                                )}
                            </div>
                        )}
                    </div>
                </div>
            </ResizablePanel>
        </ResizablePanelGroup>
    );
}
