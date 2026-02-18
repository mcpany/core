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
import { Trash2, Plus, Play, Loader2, Save } from "lucide-react";
import { Card, CardContent } from "@/components/ui/card";
import { Switch } from "@/components/ui/switch";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Textarea } from "@/components/ui/textarea";
import { InputTransformerEditor } from "./input-transformer-editor";
import { OutputTransformerEditor } from "./output-transformer-editor";
import { RequestPreview } from "./request-preview";
import { apiClient } from "@/lib/client";
import { useToast } from "@/hooks/use-toast";
import { Separator } from "@/components/ui/separator";

interface HttpToolEditorProps {
    serviceName: string;
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
export function HttpToolEditor({ serviceName, tool, call, onChange }: HttpToolEditorProps) {
    const [localTool, setLocalTool] = useState<ToolDefinition>(tool);
    const [localCall, setLocalCall] = useState<HttpCallDefinition>(call);
    const [activeTab, setActiveTab] = useState("request");
    const [testArguments, setTestArguments] = useState("{}");
    const [testResult, setTestResult] = useState<string | null>(null);
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

    const handleExecute = async () => {
        setIsExecuting(true);
        setTestResult(null);
        try {
            let args = {};
            try {
                args = JSON.parse(testArguments);
            } catch {
                throw new Error("Invalid JSON arguments");
            }

            // Construct fully qualified name
            // If serviceName is empty (e.g. creating new service), we can't really execute unless backend supports it.
            // But usually service has a name if we are in this editor.
            // If the tool is NEW and not saved, execution will fail on backend because it's not registered.
            // We should warn about this.

            const toolName = `${serviceName}.${localTool.name}`;

            // We use executeTool from client which calls /api/v1/execute
            const result = await apiClient.executeTool({
                name: toolName, // Map to ToolName in backend request
                arguments: args
            });

            setTestResult(JSON.stringify(result, null, 2));
            toast({ title: "Execution Successful", description: "Tool executed successfully." });
        } catch (e: any) {
            console.error(e);
            setTestResult(JSON.stringify({ error: e.message }, null, 2));
            toast({ variant: "destructive", title: "Execution Failed", description: e.message });
        } finally {
            setIsExecuting(false);
        }
    };

    const parsedTestArgs = (() => {
        try {
            return JSON.parse(testArguments);
        } catch {
            return {};
        }
    })();

    return (
        <div className="flex flex-col lg:flex-row gap-6 h-full">
            {/* Left: Configuration */}
            <div className="flex-1 space-y-6 overflow-y-auto pr-2">
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

            {/* Right: Live Preview & Test */}
            <div className="lg:w-[400px] flex flex-col gap-4 border-l pl-6">
                <div>
                    <h3 className="text-lg font-medium mb-2">Live Preview</h3>
                    <p className="text-sm text-muted-foreground mb-4">
                        See how arguments map to the HTTP request.
                    </p>
                </div>

                <div className="space-y-2">
                    <Label htmlFor="test-args">Test Arguments (JSON)</Label>
                    <Textarea
                        id="test-args"
                        value={testArguments}
                        onChange={(e) => setTestArguments(e.target.value)}
                        className="font-mono text-xs h-[100px]"
                        placeholder='{"userId": "123"}'
                    />
                </div>

                <div className="flex-1 min-h-[200px]">
                    <RequestPreview
                        call={localCall}
                        tool={localTool}
                        args={parsedTestArgs}
                    />
                </div>

                <Separator />

                <div className="space-y-2">
                    <div className="flex items-center justify-between">
                        <h3 className="text-sm font-medium">Test Execution</h3>
                        <Button
                            size="sm"
                            onClick={handleExecute}
                            disabled={isExecuting}
                            className="h-8"
                        >
                            {isExecuting ? <Loader2 className="mr-2 h-3 w-3 animate-spin" /> : <Play className="mr-2 h-3 w-3" />}
                            Execute
                        </Button>
                    </div>
                    <p className="text-[10px] text-muted-foreground">
                        Note: You must <strong>Save Changes</strong> to the service before executing.
                    </p>

                    {testResult && (
                        <div className="mt-2 bg-muted/50 rounded-md p-2 border overflow-auto max-h-[200px]">
                            <pre className="text-xs font-mono whitespace-pre-wrap">{testResult}</pre>
                        </div>
                    )}
                </div>
            </div>
        </div>
    );
}
