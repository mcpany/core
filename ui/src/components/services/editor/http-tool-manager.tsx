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
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";
import { Plus, Edit2, Trash2, AlertCircle } from "lucide-react";
import { HttpToolEditor } from "./http-tool-editor";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";

interface HttpToolManagerProps {
    service: UpstreamServiceConfig;
    onChange: (service: UpstreamServiceConfig) => void;
}

/**
 * Manages the list of HTTP tools for a service.
 * Allows adding, editing, and deleting tools.
 *
 * @param props - Component props
 * @param props.service - The full service configuration.
 * @param props.onChange - Callback fired when the service configuration changes.
 */
export function HttpToolManager({ service, onChange }: HttpToolManagerProps) {
    const [editingTool, setEditingTool] = useState<ToolDefinition | null>(null);
    const [editingCall, setEditingCall] = useState<HttpCallDefinition | null>(null);
    const [isEditing, setIsEditing] = useState(false);

    const httpService = service.httpService;

    if (!httpService) {
        return (
            <Alert variant="destructive">
                <AlertCircle className="h-4 w-4" />
                <AlertTitle>Error</AlertTitle>
                <AlertDescription>
                    HTTP Service configuration is missing. Please ensure the service type is set to HTTP.
                </AlertDescription>
            </Alert>
        );
    }

    const tools = httpService.tools || [];
    const calls = httpService.calls || {};

    const handleAddTool = () => {
        setEditingTool(null);
        setEditingCall(null);
        setIsEditing(true);
    };

    const handleEditTool = (tool: ToolDefinition) => {
        const callId = tool.callId;
        const call = calls[callId];
        setEditingTool(tool);
        setEditingCall(call || null);
        setIsEditing(true);
    };

    const handleDeleteTool = (tool: ToolDefinition) => {
        if (!confirm(`Are you sure you want to delete tool "${tool.name}"?`)) return;

        const newTools = tools.filter(t => t.name !== tool.name);
        const newCalls = { ...calls };
        if (tool.callId) {
            delete newCalls[tool.callId];
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

    const handleSaveTool = (newTool: ToolDefinition, newCall: HttpCallDefinition) => {
        let newTools = [...tools];
        const newCalls = { ...calls };

        // Update or Add Call
        newCalls[newCall.id] = newCall;

        // Update or Add Tool
        const existingToolIndex = newTools.findIndex(t => t.name === (editingTool?.name || newTool.name)); // Use original name if editing to find index, or new name
        // Ideally we use a stable ID, but tools are identified by name in the list usually.
        // If we are editing, we replace the one at the index we started with?
        // But the user might have renamed the tool.

        if (editingTool) {
             // We are editing. Find the original tool in the list.
             const index = newTools.findIndex(t => t.name === editingTool.name);
             if (index !== -1) {
                 newTools[index] = newTool;
             } else {
                 // Should not happen, but if it does, treat as new
                 newTools.push(newTool);
             }
        } else {
            // New tool
            newTools.push(newTool);
        }

        onChange({
            ...service,
            httpService: {
                ...httpService,
                tools: newTools,
                calls: newCalls
            }
        });
        setIsEditing(false);
    };

    const getMethodBadge = (method: number) => {
        switch (method) {
            case HttpCallDefinition_HttpMethod.HTTP_METHOD_GET: return <Badge variant="outline" className="bg-blue-500/10 text-blue-500 border-blue-500/20">GET</Badge>;
            case HttpCallDefinition_HttpMethod.HTTP_METHOD_POST: return <Badge variant="outline" className="bg-green-500/10 text-green-500 border-green-500/20">POST</Badge>;
            case HttpCallDefinition_HttpMethod.HTTP_METHOD_PUT: return <Badge variant="outline" className="bg-orange-500/10 text-orange-500 border-orange-500/20">PUT</Badge>;
            case HttpCallDefinition_HttpMethod.HTTP_METHOD_DELETE: return <Badge variant="outline" className="bg-red-500/10 text-red-500 border-red-500/20">DELETE</Badge>;
            case HttpCallDefinition_HttpMethod.HTTP_METHOD_PATCH: return <Badge variant="outline" className="bg-yellow-500/10 text-yellow-500 border-yellow-500/20">PATCH</Badge>;
            default: return <Badge variant="outline">UNK</Badge>;
        }
    };

    if (isEditing) {
        return (
            <HttpToolEditor
                tool={editingTool}
                call={editingCall}
                onSave={handleSaveTool}
                onCancel={() => setIsEditing(false)}
            />
        );
    }

    return (
        <div className="space-y-4">
            <div className="flex justify-between items-center">
                <div>
                    <h3 className="text-lg font-medium">Tools</h3>
                    <p className="text-sm text-muted-foreground">
                        Define the tools exposed by this service and map them to HTTP endpoints.
                    </p>
                </div>
                <Button onClick={handleAddTool}>
                    <Plus className="mr-2 h-4 w-4" /> Add Tool
                </Button>
            </div>

            <div className="border rounded-md">
                <Table>
                    <TableHeader>
                        <TableRow>
                            <TableHead>Name</TableHead>
                            <TableHead>Method</TableHead>
                            <TableHead>Endpoint</TableHead>
                            <TableHead>Description</TableHead>
                            <TableHead className="w-[100px] text-right">Actions</TableHead>
                        </TableRow>
                    </TableHeader>
                    <TableBody>
                        {tools.map((tool) => {
                            const call = calls[tool.callId];
                            return (
                                <TableRow key={tool.name}>
                                    <TableCell className="font-medium">{tool.name}</TableCell>
                                    <TableCell>{call ? getMethodBadge(call.method) : <Badge variant="destructive">Missing Call</Badge>}</TableCell>
                                    <TableCell className="font-mono text-xs">{call?.endpointPath || "N/A"}</TableCell>
                                    <TableCell className="max-w-[200px] truncate">{tool.description}</TableCell>
                                    <TableCell className="text-right">
                                        <div className="flex justify-end gap-2">
                                            <Button variant="ghost" size="icon" onClick={() => handleEditTool(tool)}>
                                                <Edit2 className="h-4 w-4" />
                                            </Button>
                                            <Button variant="ghost" size="icon" onClick={() => handleDeleteTool(tool)} className="text-destructive hover:text-destructive">
                                                <Trash2 className="h-4 w-4" />
                                            </Button>
                                        </div>
                                    </TableCell>
                                </TableRow>
                            );
                        })}
                        {tools.length === 0 && (
                            <TableRow>
                                <TableCell colSpan={5} className="text-center py-8 text-muted-foreground">
                                    No tools defined. Click "Add Tool" to create one.
                                </TableCell>
                            </TableRow>
                        )}
                    </TableBody>
                </Table>
            </div>
        </div>
    );
}
