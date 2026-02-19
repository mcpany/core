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
import { Trash2, Plus, Play, Loader2, Bug, Save } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Switch } from "@/components/ui/switch";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Textarea } from "@/components/ui/textarea";
import { useToast } from "@/hooks/use-toast";
import { InputTransformerEditor } from "./input-transformer-editor";
import { OutputTransformerEditor } from "./output-transformer-editor";
import { RequestPreview } from "./request-preview";
import { apiClient } from "@/lib/client";
import { cn } from "@/lib/utils";

interface HttpToolEditorProps {
    tool: ToolDefinition;
    call: HttpCallDefinition;
    serviceName: string;
    baseUrl?: string; // Added
    onChange: (tool: ToolDefinition, call: HttpCallDefinition) => void;
}

/**
 * Editor for configuring a single HTTP tool.
 * Allows defining tool metadata and the mapped HTTP request details.
 * @param props - The component props.
 * @returns The rendered tool editor.
 */
export function HttpToolEditor({ tool, call, serviceName, baseUrl, onChange }: HttpToolEditorProps) {
    const [localTool, setLocalTool] = useState<ToolDefinition>(tool);
    const [localCall, setLocalCall] = useState<HttpCallDefinition>(call);
    const [activeTab, setActiveTab] = useState("request");

    // Test & Preview State
    const [testArgs, setTestArgs] = useState("{\n  \n}");
    const [parsedArgs, setParsedArgs] = useState<Record<string, any>>({});
    const [argsError, setArgsError] = useState<string | null>(null);
    const [executing, setExecuting] = useState(false);
    const [executionResult, setExecutionResult] = useState<any>(null);
    const { toast } = useToast();

    useEffect(() => {
        setLocalTool(tool);
        setLocalCall(call);
    }, [tool, call]);

    // Parse test args on change
    useEffect(() => {
        try {
            const parsed = JSON.parse(testArgs);
            setParsedArgs(parsed);
            setArgsError(null);
        } catch (e: any) {
            setArgsError(e.message);
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
        if (argsError) {
            toast({ variant: "destructive", title: "Invalid Arguments", description: "Please fix JSON syntax errors." });
            return;
        }

        setExecuting(true);
        setExecutionResult(null);
        try {
            // Construct fully qualified name. Note: User must save service first for this to work reliably.
            // We assume the service name in context is the one registered.
            const toolName = `${serviceName}.${localTool.name}`;

            const result = await apiClient.executeTool({
                name: toolName,
                arguments: parsedArgs
            });
            setExecutionResult(result);
            toast({ title: "Execution Successful" });
        } catch (e: any) {
            setExecutionResult({ error: e.message });
            toast({ variant: "destructive", title: "Execution Failed", description: e.message });
        } finally {
            setExecuting(false);
        }
    };

    return (
        <div className="space-y-6 h-full flex flex-col">
            {/* Top Metadata */}
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

            {/* Split View */}
            <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 flex-1 min-h-0">
                {/* Left: Configuration */}
                <div className="flex flex-col min-h-0 border rounded-md p-4 bg-muted/10">
                    <Tabs value={activeTab} onValueChange={setActiveTab} className="w-full h-full flex flex-col">
                        <TabsList className="w-full justify-start flex-none">
                            <TabsTrigger value="request">Parameters</TabsTrigger>
                            <TabsTrigger value="input-transform">Input Transform</TabsTrigger>
                            <TabsTrigger value="output-transform">Output Transform</TabsTrigger>
                        </TabsList>

                        <div className="flex-1 overflow-y-auto mt-4 pr-2">
                            <TabsContent value="request" className="space-y-4 m-0">
                                <div className="flex items-center justify-between">
                                    <Label className="text-sm font-semibold">Defined Parameters</Label>
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
                                            <div className="col-span-12 sm:col-span-6 space-y-2">
                                                <Label htmlFor={`param-name-${index}`} className="text-xs">Name</Label>
                                                <Input
                                                    id={`param-name-${index}`}
                                                    value={param.schema?.name}
                                                    onChange={(e) => updateParameterSchema(index, { name: e.target.value })}
                                                    placeholder="userId"
                                                    className="h-8"
                                                />
                                            </div>
                                            <div className="col-span-12 sm:col-span-6 space-y-2">
                                                <Label htmlFor={`param-type-${index}`} className="text-xs">Type</Label>
                                                <Select
                                                    value={param.schema?.type.toString()}
                                                    onValueChange={(val) => updateParameterSchema(index, { type: parseInt(val) })}
                                                >
                                                    <SelectTrigger id={`param-type-${index}`} className="h-8">
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
                                            <div className="col-span-12 space-y-2">
                                                <Label htmlFor={`param-desc-${index}`} className="text-xs">Description</Label>
                                                <Input
                                                    id={`param-desc-${index}`}
                                                    value={param.schema?.description}
                                                    onChange={(e) => updateParameterSchema(index, { description: e.target.value })}
                                                    placeholder="The ID of the user"
                                                    className="h-8"
                                                />
                                            </div>
                                            <div className="col-span-12 flex items-center pt-2 space-x-2">
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

                            <TabsContent value="input-transform" className="m-0">
                                <InputTransformerEditor
                                    transformer={localCall.inputTransformer}
                                    onChange={handleInputTransformerChange}
                                />
                            </TabsContent>

                            <TabsContent value="output-transform" className="m-0">
                                <OutputTransformerEditor
                                    transformer={localCall.outputTransformer}
                                    onChange={handleOutputTransformerChange}
                                />
                            </TabsContent>
                        </div>
                    </Tabs>
                </div>

                {/* Right: Live Preview */}
                <div className="flex flex-col min-h-0 space-y-4">
                    <Card className="flex-1 flex flex-col min-h-0">
                        <CardHeader className="py-3 px-4 bg-muted/20 border-b">
                            <CardTitle className="text-sm font-medium flex items-center gap-2">
                                <Bug className="h-4 w-4" /> Live Preview & Test
                            </CardTitle>
                        </CardHeader>
                        <CardContent className="flex-1 overflow-y-auto p-4 space-y-4">
                            <div className="space-y-2">
                                <Label className="text-xs font-semibold">Test Arguments (JSON)</Label>
                                <div className={cn("rounded-md border", argsError ? "border-red-500" : "border-input")}>
                                    <Textarea
                                        value={testArgs}
                                        onChange={(e) => setTestArgs(e.target.value)}
                                        className="font-mono text-xs border-0 focus-visible:ring-0 min-h-[100px]"
                                        placeholder="{}"
                                    />
                                </div>
                                {argsError && <p className="text-[10px] text-red-500">{argsError}</p>}
                            </div>

                            <div className="space-y-2">
                                <Label className="text-xs font-semibold">Request Preview</Label>
                                <RequestPreview
                                    call={localCall}
                                    tool={localTool}
                                    args={parsedArgs}
                                    baseUrl={baseUrl} // Passed prop
                                />
                            </div>

                            {executionResult && (
                                <div className="space-y-2">
                                    <Label className="text-xs font-semibold">Last Execution Result</Label>
                                    <div className="bg-black/90 text-white p-2 rounded text-xs font-mono overflow-x-auto max-h-[200px]">
                                        <pre>{JSON.stringify(executionResult, null, 2)}</pre>
                                    </div>
                                </div>
                            )}
                        </CardContent>
                        <div className="p-4 border-t bg-muted/10 flex justify-between items-center">
                            <div className="text-[10px] text-muted-foreground flex items-center gap-1">
                                <Save className="h-3 w-3" />
                                <span>Save service before executing.</span>
                            </div>
                            <Button
                                size="sm"
                                onClick={handleExecute}
                                disabled={executing || !!argsError}
                                className="gap-2"
                            >
                                {executing ? <Loader2 className="h-3 w-3 animate-spin" /> : <Play className="h-3 w-3" />}
                                Execute
                            </Button>
                        </div>
                    </Card>
                </div>
            </div>
        </div>
    );
}
