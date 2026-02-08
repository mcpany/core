/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState } from "react";
import { UpstreamServiceConfig } from "@/lib/client";
import { ToolDefinition } from "@proto/config/v1/tool";
import { HttpCallDefinition, HttpCallDefinition_HttpMethod } from "@proto/config/v1/call";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Plus, Settings, Trash2, Edit } from "lucide-react";
import {
    Sheet,
    SheetContent,
    SheetDescription,
    SheetHeader,
    SheetTitle,
    SheetTrigger,
} from "@/components/ui/sheet";
import { HttpToolEditor } from "./http-tool-editor";
import { Badge } from "@/components/ui/badge";

interface HttpToolManagerProps {
    service: UpstreamServiceConfig;
    onChange: (service: UpstreamServiceConfig) => void;
}

/**
 * Manager component for HTTP tools within a service.
 * Displays a list of tools and allows adding, editing, and deleting them.
 * @param props - The component props.
 * @returns The rendered tool manager.
 */
export function HttpToolManager({ service, onChange }: HttpToolManagerProps) {
    const [editingToolIndex, setEditingToolIndex] = useState<number | null>(null);
    const [isSheetOpen, setIsSheetOpen] = useState(false);

    // Helper to safely get tools/calls
    const tools = service.httpService?.tools || [];
    const calls = service.httpService?.calls || {};

    const handleAddTool = () => {
        const callId = crypto.randomUUID();
        const newTool: ToolDefinition = {
            name: "new_tool",
            description: "New Tool Description",
            callId: callId,
            disable: false,
            // Initialize other required fields
            mergeStrategy: 0,
            tags: [],
            profiles: [],
            inputSchema: undefined,
            isStream: false,
            title: "",
            readOnlyHint: false,
            destructiveHint: false,
            idempotentHint: false,
            openWorldHint: false,
            integrity: undefined,
            serviceId: "", // Will be set by backend or context
        };

        const newCall: HttpCallDefinition = {
            id: callId,
            method: HttpCallDefinition_HttpMethod.HTTP_METHOD_GET,
            endpointPath: "/",
            parameters: [],
            cache: undefined,
            inputSchema: undefined,
            inputTransformer: undefined,
            outputSchema: undefined,
            outputTransformer: undefined,
        };

        const newTools = [...tools, newTool];
        const newCalls = { ...calls, [callId]: newCall };

        onChange({
            ...service,
            httpService: {
                ...service.httpService!,
                tools: newTools,
                calls: newCalls
            }
        });

        setEditingToolIndex(newTools.length - 1);
        setIsSheetOpen(true);
    };

    const handleEditTool = (index: number) => {
        setEditingToolIndex(index);
        setIsSheetOpen(true);
    };

    const handleDeleteTool = (index: number) => {
        const toolToDelete = tools[index];
        const newTools = [...tools];
        newTools.splice(index, 1);

        const newCalls = { ...calls };
        if (toolToDelete.callId) {
            delete newCalls[toolToDelete.callId];
        }

        onChange({
            ...service,
            httpService: {
                ...service.httpService!,
                tools: newTools,
                calls: newCalls
            }
        });
    };

    const handleToolChange = (updatedTool: ToolDefinition, updatedCall: HttpCallDefinition) => {
        if (editingToolIndex === null) return;

        const newTools = [...tools];
        newTools[editingToolIndex] = updatedTool;

        const newCalls = { ...calls };
        // If call ID changed (unlikely but possible), remove old one
        if (updatedCall.id !== updatedTool.callId) {
             // This shouldn't happen in normal flow, but good to be safe.
             // We generally keep callId constant.
        }
        newCalls[updatedCall.id] = updatedCall;

        onChange({
            ...service,
            httpService: {
                ...service.httpService!,
                tools: newTools,
                calls: newCalls
            }
        });
    };

    const getCallForTool = (tool: ToolDefinition): HttpCallDefinition => {
        return calls[tool.callId] || {
            id: tool.callId,
            method: HttpCallDefinition_HttpMethod.HTTP_METHOD_GET,
            endpointPath: "/",
            parameters: [],
        } as HttpCallDefinition;
    };

    const getMethodName = (method: HttpCallDefinition_HttpMethod) => {
        switch (method) {
            case HttpCallDefinition_HttpMethod.HTTP_METHOD_GET: return "GET";
            case HttpCallDefinition_HttpMethod.HTTP_METHOD_POST: return "POST";
            case HttpCallDefinition_HttpMethod.HTTP_METHOD_PUT: return "PUT";
            case HttpCallDefinition_HttpMethod.HTTP_METHOD_DELETE: return "DELETE";
            case HttpCallDefinition_HttpMethod.HTTP_METHOD_PATCH: return "PATCH";
            default: return "UNK";
        }
    };

    const getMethodColor = (method: HttpCallDefinition_HttpMethod) => {
        switch (method) {
            case HttpCallDefinition_HttpMethod.HTTP_METHOD_GET: return "bg-blue-500/10 text-blue-500 border-blue-500/20";
            case HttpCallDefinition_HttpMethod.HTTP_METHOD_POST: return "bg-green-500/10 text-green-500 border-green-500/20";
            case HttpCallDefinition_HttpMethod.HTTP_METHOD_PUT: return "bg-orange-500/10 text-orange-500 border-orange-500/20";
            case HttpCallDefinition_HttpMethod.HTTP_METHOD_DELETE: return "bg-red-500/10 text-red-500 border-red-500/20";
            default: return "bg-gray-500/10 text-gray-500";
        }
    };

    return (
        <div className="space-y-4">
            <div className="flex items-center justify-between">
                <div>
                    <h3 className="text-lg font-medium">Defined Tools</h3>
                    <p className="text-sm text-muted-foreground">
                        Define the tools exposed by this HTTP service.
                    </p>
                </div>
                <Button onClick={handleAddTool} size="sm">
                    <Plus className="mr-2 h-4 w-4" /> Add Tool
                </Button>
            </div>

            <div className="grid gap-4">
                {tools.length === 0 && (
                    <div className="text-center py-10 border border-dashed rounded-lg">
                        <p className="text-muted-foreground mb-2">No tools defined.</p>
                        <Button variant="outline" onClick={handleAddTool}>
                            Create your first tool
                        </Button>
                    </div>
                )}
                {tools.map((tool, index) => {
                    const call = getCallForTool(tool);
                    return (
                        <Card key={index} className="flex items-center justify-between p-4">
                            <div className="flex flex-col gap-1">
                                <div className="flex items-center gap-2">
                                    <span className="font-semibold">{tool.name}</span>
                                    <Badge variant="outline" className={getMethodColor(call.method)}>
                                        {getMethodName(call.method)}
                                    </Badge>
                                </div>
                                <div className="text-sm text-muted-foreground font-mono">
                                    {call.endpointPath}
                                </div>
                                <div className="text-xs text-muted-foreground">
                                    {tool.description || "No description"}
                                </div>
                            </div>
                            <div className="flex items-center gap-2">
                                <Button variant="ghost" size="icon" onClick={() => handleEditTool(index)}>
                                    <Edit className="h-4 w-4" />
                                </Button>
                                <Button variant="ghost" size="icon" onClick={() => handleDeleteTool(index)}>
                                    <Trash2 className="h-4 w-4 text-destructive" />
                                </Button>
                            </div>
                        </Card>
                    );
                })}
            </div>

            <Sheet open={isSheetOpen} onOpenChange={setIsSheetOpen}>
                <SheetContent className="sm:max-w-xl w-full overflow-y-auto">
                    <SheetHeader>
                        <SheetTitle>
                            {editingToolIndex !== null && tools[editingToolIndex] ? `Edit ${tools[editingToolIndex].name}` : "Edit Tool"}
                        </SheetTitle>
                        <SheetDescription>
                            Configure the tool definition and its mapping to the HTTP endpoint.
                        </SheetDescription>
                    </SheetHeader>
                    <div className="mt-6">
                        {editingToolIndex !== null && tools[editingToolIndex] && (
                            <HttpToolEditor
                                tool={tools[editingToolIndex]}
                                call={getCallForTool(tools[editingToolIndex])}
                                onChange={handleToolChange}
                            />
                        )}
                    </div>
                    <div className="mt-6 flex justify-end">
                        <Button onClick={() => setIsSheetOpen(false)}>Done</Button>
                    </div>
                </SheetContent>
            </Sheet>
        </div>
    );
}
