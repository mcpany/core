/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect } from "react";
import { ToolDefinition } from "@proto/config/v1/tool";
import { HttpCallDefinition, HttpCallDefinition_HttpMethod, HttpParameterMapping, ParameterType, InputTransformer, OutputTransformer } from "@proto/config/v1/call";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Button } from "@/components/ui/button";
import { Trash2, Plus, Play, Loader2, AlertTriangle } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Switch } from "@/components/ui/switch";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Textarea } from "@/components/ui/textarea";
import { InputTransformerEditor } from "./input-transformer-editor";
import { OutputTransformerEditor } from "./output-transformer-editor";
import { RequestPreview } from "./request-preview";
import { apiClient } from "@/lib/client";
import { cn } from "@/lib/utils";

interface HttpToolEditorProps {
    tool: ToolDefinition;
    call: HttpCallDefinition;
    serviceName?: string;
    serviceAddress?: string;
    onChange: (tool: ToolDefinition, call: HttpCallDefinition) => void;
}

/**
 * Editor for configuring a single HTTP tool.
 * Allows defining tool metadata, mapped HTTP request details, and live testing.
 * @param props - The component props.
 * @returns The rendered tool editor.
 */
export function HttpToolEditor({ tool, call, serviceName, serviceAddress, onChange }: HttpToolEditorProps) {
    const [localTool, setLocalTool] = useState<ToolDefinition>(tool);
    const [localCall, setLocalCall] = useState<HttpCallDefinition>(call);
    const [activeTab, setActiveTab] = useState("request");

    // Test State
    const [testArgs, setTestArgs] = useState("{}");
    const [parsedArgs, setParsedArgs] = useState<Record<string, unknown>>({});
    const [argError, setArgError] = useState<string | null>(null);
    const [isExecuting, setIsExecuting] = useState(false);
    const [executionResult, setExecutionResult] = useState<any>(null);
    const [executionError, setExecutionError] = useState<string | null>(null);

    useEffect(() => {
        setLocalTool(tool);
        setLocalCall(call);
    }, [tool, call]);

    useEffect(() => {
        try {
            const parsed = JSON.parse(testArgs);
            setParsedArgs(parsed);
            setArgError(null);
        } catch (e: any) {
            setArgError(e.message);
        }
    }, [testArgs]);

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

    const handleExecute = async () => {
        if (!serviceName) return;
        setIsExecuting(true);
        setExecutionError(null);
        setExecutionResult(null);

        try {
            // Construct fully qualified name
            const fullName = `${serviceName}.${localTool.name}`;
            const result = await apiClient.executeTool({
                name: fullName,
                arguments: parsedArgs
            });
            setExecutionResult(result);
        } catch (e: any) {
            setExecutionError(e.message || "Execution failed");
        } finally {
            setIsExecuting(false);
        }
    };

    return (
        <div className="space-y-6">
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
                    <TabsTrigger value="request">Parameters</TabsTrigger>
                    <TabsTrigger value="input-transform">Input Transform</TabsTrigger>
                    <TabsTrigger value="output-transform">Output Transform</TabsTrigger>
                    <TabsTrigger value="test" className="ml-auto bg-primary/5 text-primary data-[state=active]:bg-primary/10">Live Test</TabsTrigger>
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

                <TabsContent value="test" className="mt-4 space-y-6">
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-6 h-[500px]">
                        <div className="flex flex-col gap-4 h-full">
                            <div className="space-y-2">
                                <Label>Test Arguments (JSON)</Label>
                                <div className="relative h-[200px]">
                                    <Textarea
                                        className={cn("font-mono h-full resize-none", argError && "border-destructive")}
                                        value={testArgs}
                                        onChange={(e) => setTestArgs(e.target.value)}
                                        placeholder='{"city": "San Francisco"}'
                                    />
                                    {argError && (
                                        <span className="text-xs text-destructive absolute bottom-2 left-2 bg-background/90 px-1 rounded">
                                            {argError}
                                        </span>
                                    )}
                                </div>
                            </div>

                            <RequestPreview
                                tool={localTool}
                                call={localCall}
                                args={parsedArgs}
                                baseUrl={serviceAddress}
                            />
                        </div>

                        <div className="flex flex-col gap-4 h-full">
                            <div className="flex items-center justify-between">
                                <Label>Execution Result</Label>
                                <Button
                                    onClick={handleExecute}
                                    disabled={isExecuting || !!argError || !serviceName}
                                    size="sm"
                                >
                                    {isExecuting ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <Play className="mr-2 h-4 w-4" />}
                                    Execute
                                </Button>
                            </div>

                            {!serviceName && (
                                <div className="p-3 bg-yellow-50 dark:bg-yellow-900/20 text-yellow-800 dark:text-yellow-200 text-xs rounded border border-yellow-200 dark:border-yellow-900 flex items-center gap-2">
                                    <AlertTriangle className="h-4 w-4" />
                                    Save the service to enable execution.
                                </div>
                            )}

                            <div className="flex-1 bg-muted/30 rounded border border-muted p-4 font-mono text-xs overflow-auto">
                                {executionResult ? (
                                    <pre>{JSON.stringify(executionResult, null, 2)}</pre>
                                ) : executionError ? (
                                    <span className="text-destructive whitespace-pre-wrap">{executionError}</span>
                                ) : (
                                    <span className="text-muted-foreground italic">
                                        Run execution to see results...
                                    </span>
                                )}
                            </div>
                        </div>
                    </div>
                </TabsContent>
            </Tabs>
        </div>
    );
}
