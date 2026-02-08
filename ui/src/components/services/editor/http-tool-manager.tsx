/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState } from "react";
import { UpstreamServiceConfig } from "@/lib/client";
import { Button } from "@/components/ui/button";
import { Plus, Trash2, Edit, Wrench } from "lucide-react";
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
    DialogHeader,
    DialogTitle,
    DialogDescription,
} from "@/components/ui/dialog";
import { ToolDefinition } from "@proto/config/v1/tool";
import { HttpCallDefinition } from "@proto/config/v1/call";
import { HttpToolEditor } from "./http-tool-editor";
import { v4 as uuidv4 } from "uuid";

interface HttpToolManagerProps {
    service: UpstreamServiceConfig;
    onChange: (service: UpstreamServiceConfig) => void;
}

/**
 * Manages the list of HTTP tools for a service.
 * Provides functionality to add, edit, and delete tools mapped to HTTP endpoints.
 * @param props The component props.
 * @param props.service The service configuration.
 * @param props.onChange Callback when the service configuration is updated.
 * @returns The rendered component.
 */
export function HttpToolManager({ service, onChange }: HttpToolManagerProps) {
    const [editingTool, setEditingTool] = useState<ToolDefinition | null>(null);
    const [isDialogOpen, setIsDialogOpen] = useState(false);

    const tools = service.httpService?.tools || [];
    const calls = service.httpService?.calls || {};

    const handleAddTool = () => {
        const newToolId = uuidv4();
        const newCallId = uuidv4();

        // Initialize new tool
        const newTool: ToolDefinition = {
            name: "new_tool",
            description: "New Tool Description",
            callId: newCallId,
            serviceId: service.id,
            inputSchema: undefined,
            isStream: false,
            title: "",
            readOnlyHint: false,
            destructiveHint: false,
            idempotentHint: false,
            openWorldHint: false,
            disable: false,
            profiles: [],
            mergeStrategy: 0,
            tags: [],
            integrity: undefined
        };

        // Initialize new call
        const newCall: HttpCallDefinition = {
            id: newCallId,
            endpointPath: "/",
            method: 1, // GET
            parameters: [],
            inputTransformer: undefined,
            outputTransformer: undefined,
            cache: undefined,
            inputSchema: undefined,
            outputSchema: undefined
        };

        setEditingTool(newTool);

        // Update service state temporarily to include the new call so the editor can find it
        // But we don't commit to parent yet until save?
        // Actually, for simplicity, let's just pass the new objects to the editor
        // and only update the parent when the editor saves.
        // However, the Editor might need to see the 'service' context.
        // Let's rely on the editor to return the updated tool and call.

        setIsDialogOpen(true);
    };

    const handleEditTool = (tool: ToolDefinition) => {
        setEditingTool(tool);
        setIsDialogOpen(true);
    };

    const handleDeleteTool = (toolName: string) => {
        if (!confirm(`Are you sure you want to delete tool "${toolName}"?`)) return;

        const toolToDelete = tools.find(t => t.name === toolName);
        if (!toolToDelete) return;

        const newTools = tools.filter(t => t.name !== toolName);
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

    const handleSaveTool = (updatedTool: ToolDefinition, updatedCall: HttpCallDefinition) => {
        const newTools = [...tools];
        const existingIndex = newTools.findIndex(t => t.name === editingTool?.name); // Use original name to find?
        // Wait, if we rename, we might lose it. Better use object identity or ID if stable.
        // Since we are editing 'editingTool', we can assume we are updating or adding.

        // If it's a new tool (not in list), add it.
        // But we initialized 'editingTool' with a dummy.
        // Let's use the callId to match, as it should be unique and stable during the edit session.
        const indexByCallId = newTools.findIndex(t => t.callId === updatedTool.callId);

        if (indexByCallId >= 0) {
            newTools[indexByCallId] = updatedTool;
        } else {
            newTools.push(updatedTool);
        }

        const newCalls = { ...calls, [updatedCall.id]: updatedCall };

        onChange({
            ...service,
            httpService: {
                ...service.httpService!,
                tools: newTools,
                calls: newCalls
            }
        });
        setIsDialogOpen(false);
        setEditingTool(null);
    };

    return (
        <div className="space-y-4">
            <div className="flex justify-between items-center">
                <div className="space-y-1">
                    <h3 className="text-lg font-medium">HTTP Tools</h3>
                    <p className="text-sm text-muted-foreground">
                        Define tools that map to HTTP endpoints on this service.
                    </p>
                </div>
                <Button onClick={handleAddTool} size="sm">
                    <Plus className="mr-2 h-4 w-4" /> Add Tool
                </Button>
            </div>

            <div className="rounded-md border">
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
                        {tools.length === 0 ? (
                            <TableRow>
                                <TableCell colSpan={5} className="h-24 text-center text-muted-foreground">
                                    No tools defined. Click "Add Tool" to create one.
                                </TableCell>
                            </TableRow>
                        ) : (
                            tools.map((tool) => {
                                const call = calls[tool.callId];
                                return (
                                    <TableRow key={tool.name}>
                                        <TableCell className="font-medium flex items-center gap-2">
                                            <Wrench className="h-4 w-4 text-muted-foreground" />
                                            {tool.name}
                                        </TableCell>
                                        <TableCell>{tool.description}</TableCell>
                                        <TableCell>
                                            {call ? (
                                                <span className={`font-mono text-xs font-bold ${
                                                    call.method === 1 ? "text-blue-500" :
                                                    call.method === 2 ? "text-green-500" :
                                                    call.method === 3 ? "text-orange-500" :
                                                    call.method === 4 ? "text-red-500" : ""
                                                }`}>
                                                    {call.method === 1 ? "GET" :
                                                     call.method === 2 ? "POST" :
                                                     call.method === 3 ? "PUT" :
                                                     call.method === 4 ? "DELETE" :
                                                     call.method === 5 ? "PATCH" : "UNK"}
                                                </span>
                                            ) : "-"}
                                        </TableCell>
                                        <TableCell className="font-mono text-xs">
                                            {call?.endpointPath || "-"}
                                        </TableCell>
                                        <TableCell className="text-right">
                                            <div className="flex justify-end gap-2">
                                                <Button variant="ghost" size="icon" onClick={() => handleEditTool(tool)}>
                                                    <Edit className="h-4 w-4" />
                                                </Button>
                                                <Button variant="ghost" size="icon" onClick={() => handleDeleteTool(tool.name)}>
                                                    <Trash2 className="h-4 w-4 text-destructive" />
                                                </Button>
                                            </div>
                                        </TableCell>
                                    </TableRow>
                                );
                            })
                        )}
                    </TableBody>
                </Table>
            </div>

            <Dialog open={isDialogOpen} onOpenChange={setIsDialogOpen}>
                <DialogContent className="max-w-4xl max-h-[90vh] overflow-y-auto">
                    <DialogHeader>
                        <DialogTitle>{editingTool && tools.find(t => t.callId === editingTool.callId) ? "Edit Tool" : "New Tool"}</DialogTitle>
                        <DialogDescription>
                            Configure the tool definition and its mapping to the HTTP endpoint.
                        </DialogDescription>
                    </DialogHeader>
                    {editingTool && (
                        <HttpToolEditor
                            initialTool={editingTool}
                            initialCall={calls[editingTool.callId] || {
                                id: editingTool.callId,
                                endpointPath: "",
                                method: 1,
                                parameters: [],
                                inputTransformer: undefined,
                                outputTransformer: undefined,
                                cache: undefined,
                                inputSchema: undefined,
                                outputSchema: undefined
                            }}
                            onSave={handleSaveTool}
                            onCancel={() => setIsDialogOpen(false)}
                        />
                    )}
                </DialogContent>
            </Dialog>
        </div>
    );
}
