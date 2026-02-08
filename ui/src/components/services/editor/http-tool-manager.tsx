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
import { Plus, Trash2, Edit, AlertCircle } from "lucide-react";
import {
    Table,
    TableBody,
    TableCell,
    TableHead,
    TableHeader,
    TableRow,
} from "@/components/ui/table";
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogHeader,
    DialogTitle,
} from "@/components/ui/dialog";
import { HttpToolEditor } from "./http-tool-editor";
import { Badge } from "@/components/ui/badge";

interface HttpToolManagerProps {
    service: UpstreamServiceConfig;
    onChange: (service: UpstreamServiceConfig) => void;
}

/**
 * Manages the list of tools for an HTTP service.
 * Allows adding, editing, and deleting tools and their associated HTTP call definitions.
 *
 * @param props - The component props.
 * @param props.service - The current service configuration.
 * @param props.onChange - Callback to update the service configuration.
 */
export function HttpToolManager({ service, onChange }: HttpToolManagerProps) {
    const [isEditorOpen, setIsEditorOpen] = useState(false);
    const [editingToolIndex, setEditingToolIndex] = useState<number | null>(null);

    // Helper to get tools list safely
    const tools = service.httpService?.tools || [];
    const calls = service.httpService?.calls || {};

    const handleAddTool = () => {
        setEditingToolIndex(null);
        setIsEditorOpen(true);
    };

    const handleEditTool = (index: number) => {
        setEditingToolIndex(index);
        setIsEditorOpen(true);
    };

    const handleDeleteTool = (index: number) => {
        const httpService = service.httpService || { address: "", tools: [], calls: {} };
        const currentTools = httpService.tools || [];
        const currentCalls = httpService.calls || {};

        const toolToDelete = currentTools[index];
        if (!toolToDelete) return;

        const newTools = [...currentTools];
        newTools.splice(index, 1);

        // Also remove the associated call if it exists and isn't used by other tools (though 1:1 is typical)
        const newCalls = { ...currentCalls };
        if (toolToDelete.callId && newCalls[toolToDelete.callId]) {
            delete newCalls[toolToDelete.callId];
        }

        onChange({
            ...service,
            httpService: {
                ...httpService,
                tools: newTools,
                calls: newCalls
            }
        });
    };

    const handleSaveTool = (tool: ToolDefinition, call: HttpCallDefinition) => {
        const httpService = service.httpService || { address: "", tools: [], calls: {} };
        const currentTools = httpService.tools || [];
        const currentCalls = httpService.calls || {};

        const newTools = [...currentTools];
        const newCalls = { ...currentCalls };

        // Ensure call ID match
        tool.callId = call.id;

        if (editingToolIndex !== null) {
            // Update existing
            const oldTool = newTools[editingToolIndex];
            // If call ID changed (unlikely but possible if regenerated), remove old call
            if (oldTool && oldTool.callId && oldTool.callId !== call.id) {
                delete newCalls[oldTool.callId];
            }
            newTools[editingToolIndex] = tool;
        } else {
            // Add new
            newTools.push(tool);
        }

        // Save call definition
        newCalls[call.id] = call;

        onChange({
            ...service,
            httpService: {
                ...httpService,
                tools: newTools,
                calls: newCalls
            }
        });
        setIsEditorOpen(false);
    };

    const getMethodBadge = (method: HttpCallDefinition_HttpMethod) => {
        switch (method) {
            case HttpCallDefinition_HttpMethod.HTTP_METHOD_GET: return <Badge variant="outline" className="text-blue-500 border-blue-500">GET</Badge>;
            case HttpCallDefinition_HttpMethod.HTTP_METHOD_POST: return <Badge variant="outline" className="text-green-500 border-green-500">POST</Badge>;
            case HttpCallDefinition_HttpMethod.HTTP_METHOD_PUT: return <Badge variant="outline" className="text-orange-500 border-orange-500">PUT</Badge>;
            case HttpCallDefinition_HttpMethod.HTTP_METHOD_DELETE: return <Badge variant="outline" className="text-red-500 border-red-500">DELETE</Badge>;
            case HttpCallDefinition_HttpMethod.HTTP_METHOD_PATCH: return <Badge variant="outline" className="text-yellow-500 border-yellow-500">PATCH</Badge>;
            default: return <Badge variant="secondary">UNK</Badge>;
        }
    };

    return (
        <div className="space-y-4">
            <div className="flex justify-between items-center">
                <div>
                    <h3 className="text-lg font-medium">HTTP Tools</h3>
                    <p className="text-sm text-muted-foreground">
                        Define tools that map to HTTP endpoints on this service.
                    </p>
                </div>
                <Button onClick={handleAddTool}>
                    <Plus className="mr-2 h-4 w-4" /> Add Tool
                </Button>
            </div>

            {tools.length === 0 ? (
                <div className="border border-dashed rounded-lg p-8 text-center text-muted-foreground">
                    No tools defined. Click "Add Tool" to create one.
                </div>
            ) : (
                <div className="border rounded-md">
                    <Table>
                        <TableHeader>
                            <TableRow>
                                <TableHead>Name</TableHead>
                                <TableHead>Description</TableHead>
                                <TableHead>Method</TableHead>
                                <TableHead>Path</TableHead>
                                <TableHead className="text-right">Actions</TableHead>
                            </TableRow>
                        </TableHeader>
                        <TableBody>
                            {tools.map((tool, index) => {
                                const call = tool.callId ? calls[tool.callId] : undefined;
                                return (
                                    <TableRow key={index}>
                                        <TableCell className="font-medium">{tool.name}</TableCell>
                                        <TableCell className="max-w-xs truncate" title={tool.description}>{tool.description}</TableCell>
                                        <TableCell>
                                            {call ? getMethodBadge(call.method) : <span className="text-muted-foreground">-</span>}
                                        </TableCell>
                                        <TableCell className="font-mono text-xs">
                                            {call?.endpointPath || <span className="text-destructive flex items-center gap-1"><AlertCircle className="h-3 w-3"/> Missing Call</span>}
                                        </TableCell>
                                        <TableCell className="text-right space-x-2">
                                            <Button variant="ghost" size="icon" onClick={() => handleEditTool(index)}>
                                                <Edit className="h-4 w-4" />
                                            </Button>
                                            <Button variant="ghost" size="icon" className="text-destructive hover:text-destructive" onClick={() => handleDeleteTool(index)}>
                                                <Trash2 className="h-4 w-4" />
                                            </Button>
                                        </TableCell>
                                    </TableRow>
                                );
                            })}
                        </TableBody>
                    </Table>
                </div>
            )}

            <Dialog open={isEditorOpen} onOpenChange={setIsEditorOpen}>
                <DialogContent className="max-w-3xl max-h-[90vh] overflow-y-auto">
                    <DialogHeader>
                        <DialogTitle>{editingToolIndex !== null ? "Edit Tool" : "Add New Tool"}</DialogTitle>
                        <DialogDescription>
                            Configure the tool definition and its mapping to an HTTP endpoint.
                        </DialogDescription>
                    </DialogHeader>
                    {isEditorOpen && (
                        <HttpToolEditor
                            initialTool={editingToolIndex !== null ? tools[editingToolIndex] : undefined}
                            initialCall={editingToolIndex !== null && tools[editingToolIndex].callId ? calls[tools[editingToolIndex].callId] : undefined}
                            onSave={handleSaveTool}
                            onCancel={() => setIsEditorOpen(false)}
                        />
                    )}
                </DialogContent>
            </Dialog>
        </div>
    );
}
