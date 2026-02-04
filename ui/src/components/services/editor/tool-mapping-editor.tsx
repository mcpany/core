/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useMemo } from "react";
import { UpstreamServiceConfig } from "@/lib/client";
import { ToolDefinition } from "@proto/config/v1/tool";
import { HttpCallDefinition } from "@proto/config/v1/call";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
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
    DialogFooter,
    DialogHeader,
    DialogTitle,
} from "@/components/ui/dialog";
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from "@/components/ui/select";
import { Plus, Trash2, Edit2, Zap } from "lucide-react";
import { Textarea } from "@/components/ui/textarea";
import { Badge } from "@/components/ui/badge";
import { useToast } from "@/hooks/use-toast";

interface ToolMappingEditorProps {
    service: UpstreamServiceConfig;
    onChange: (service: UpstreamServiceConfig) => void;
}

/**
 * ToolMappingEditor component allows visual configuration of Tool Definitions
 * and their mapping to Upstream HTTP calls.
 *
 * @param props - Component props.
 * @param props.service - The current service configuration.
 * @param props.onChange - Callback when configuration changes.
 * @returns The rendered component.
 */
export function ToolMappingEditor({ service, onChange }: ToolMappingEditorProps) {
    const { toast } = useToast();
    const [isDialogOpen, setIsDialogOpen] = useState(false);
    const [editingTool, setEditingTool] = useState<Partial<ToolDefinition> | null>(null);
    const [editingCall, setEditingCall] = useState<Partial<HttpCallDefinition> | null>(null);
    const [schemaJson, setSchemaJson] = useState("{}");

    // Filter tools relevant to this editor (HTTP only for now)
    const httpTools = useMemo(() => {
        if (!service.httpService) return [];
        return service.httpService.tools || [];
    }, [service.httpService]);

    const handleAdd = () => {
        setEditingTool({
            name: "",
            description: "",
            inputSchema: { fields: {} }, // Default empty struct
        });
        setSchemaJson('{\n  "type": "object",\n  "properties": {\n    "arg1": { "type": "string" }\n  }\n}');
        setEditingCall({
            method: 1, // GET
            endpointPath: "/",
        });
        setIsDialogOpen(true);
    };

    const handleEdit = (tool: ToolDefinition) => {
        setEditingTool(tool);
        setSchemaJson(JSON.stringify(tool.inputSchema || {}, null, 2));
        // Find corresponding call
        const callId = tool.callId;
        if (callId && service.httpService?.calls && service.httpService.calls[callId]) {
            setEditingCall(service.httpService.calls[callId]);
        } else {
            // New call config if missing
            setEditingCall({
                method: 1,
                endpointPath: "/",
            });
        }
        setIsDialogOpen(true);
    };

    const handleDelete = (toolName: string) => {
        if (!service.httpService) return;

        const toolToDelete = service.httpService.tools?.find(t => t.name === toolName);
        const newTools = service.httpService.tools?.filter(t => t.name !== toolName) || [];

        // Cleanup call if it exists
        const newCalls = { ...(service.httpService.calls || {}) };
        if (toolToDelete?.callId) {
            delete newCalls[toolToDelete.callId];
        }

        onChange({
            ...service,
            httpService: {
                ...service.httpService,
                tools: newTools,
                calls: newCalls,
            }
        });
    };

    const handleSave = () => {
        if (!service.httpService || !editingTool || !editingTool.name) return;

        const callId = editingTool.callId || `call-${crypto.randomUUID()}`;

        let schema = {};
        try {
            schema = JSON.parse(schemaJson);
        } catch (e) {
            console.error("Invalid schema JSON", e);
            toast({
                variant: "destructive",
                title: "Invalid Input Schema",
                description: "Please ensure the input schema is valid JSON.",
            });
            return;
        }

        // Prepare Tool Definition
        const newTool: ToolDefinition = {
            ...editingTool,
            name: editingTool.name, // Ensure typed
            callId: callId,
            inputSchema: schema as any
        } as ToolDefinition;

        // Prepare Call Definition
        const newCall: HttpCallDefinition = {
            id: callId,
            method: editingCall?.method || 1,
            endpointPath: editingCall?.endpointPath || "/",
            // Note: Parameters mapping is implicitly handled by templating in endpointPath
            // e.g. /users/{{userId}}
        } as HttpCallDefinition;

        // Update Service Config
        const newTools = [...(service.httpService.tools || []).filter(t => t.name !== editingTool.name), newTool];
        const newCalls = { ...(service.httpService.calls || {}), [callId]: newCall };

        onChange({
            ...service,
            httpService: {
                ...service.httpService,
                tools: newTools,
                calls: newCalls
            }
        });

        setIsDialogOpen(false);
    };

    if (!service.httpService) {
        return (
            <div className="p-8 text-center text-muted-foreground border-2 border-dashed rounded-lg">
                Tool mapping is currently only supported for HTTP services.
            </div>
        );
    }

    return (
        <div className="space-y-4">
            <div className="flex justify-between items-center">
                <div>
                    <h3 className="text-lg font-medium">Tool Mappings</h3>
                    <p className="text-sm text-muted-foreground">
                        Define tools and map them to HTTP endpoints.
                    </p>
                </div>
                <Button onClick={handleAdd}>
                    <Plus className="mr-2 h-4 w-4" /> Add Tool
                </Button>
            </div>

            <div className="rounded-md border">
                <Table>
                    <TableHeader>
                        <TableRow>
                            <TableHead>Tool Name</TableHead>
                            <TableHead>Description</TableHead>
                            <TableHead>Mapped To</TableHead>
                            <TableHead className="text-right">Actions</TableHead>
                        </TableRow>
                    </TableHeader>
                    <TableBody>
                        {httpTools.length === 0 && (
                            <TableRow>
                                <TableCell colSpan={4} className="text-center h-24 text-muted-foreground">
                                    No tools defined.
                                </TableCell>
                            </TableRow>
                        )}
                        {httpTools.map((tool) => {
                            const call = service.httpService!.calls?.[tool.callId];
                            const method = call?.method === 1 ? "GET" : call?.method === 2 ? "POST" : "HTTP";
                            return (
                                <TableRow key={tool.name}>
                                    <TableCell className="font-medium flex items-center gap-2">
                                        <Zap className="h-4 w-4 text-amber-500" />
                                        {tool.name}
                                    </TableCell>
                                    <TableCell>{tool.description}</TableCell>
                                    <TableCell>
                                        {call ? (
                                            <Badge variant="outline" className="font-mono text-xs">
                                                {method} {call.endpointPath}
                                            </Badge>
                                        ) : (
                                            <span className="text-destructive text-xs">Unmapped</span>
                                        )}
                                    </TableCell>
                                    <TableCell className="text-right">
                                        <Button variant="ghost" size="icon" onClick={() => handleEdit(tool)}>
                                            <Edit2 className="h-4 w-4" />
                                        </Button>
                                        <Button variant="ghost" size="icon" className="text-destructive" onClick={() => handleDelete(tool.name)}>
                                            <Trash2 className="h-4 w-4" />
                                        </Button>
                                    </TableCell>
                                </TableRow>
                            );
                        })}
                    </TableBody>
                </Table>
            </div>

            <Dialog open={isDialogOpen} onOpenChange={setIsDialogOpen}>
                <DialogContent className="sm:max-w-[600px]">
                    <DialogHeader>
                        <DialogTitle>{editingTool?.callId ? "Edit Tool" : "Add Tool"}</DialogTitle>
                        <DialogDescription>
                            Configure the tool interface and its upstream HTTP call.
                        </DialogDescription>
                    </DialogHeader>

                    <div className="grid gap-6 py-4">
                        {/* Tool Definition */}
                        <div className="space-y-4">
                            <h4 className="text-sm font-medium border-b pb-2">Tool Interface</h4>
                            <div className="grid grid-cols-2 gap-4">
                                <div className="space-y-2">
                                    <Label htmlFor="tool-name">Name</Label>
                                    <Input
                                        id="tool-name"
                                        value={editingTool?.name || ""}
                                        onChange={(e) => setEditingTool(prev => ({ ...prev, name: e.target.value }))}
                                        placeholder="get_weather"
                                    />
                                </div>
                                <div className="space-y-2">
                                    <Label htmlFor="tool-desc">Description</Label>
                                    <Input
                                        id="tool-desc"
                                        value={editingTool?.description || ""}
                                        onChange={(e) => setEditingTool(prev => ({ ...prev, description: e.target.value }))}
                                        placeholder="Gets the weather for a location"
                                    />
                                </div>
                            </div>
                            <div className="space-y-2">
                                <Label htmlFor="input-schema">Input Schema (JSON)</Label>
                                <Textarea
                                    id="input-schema"
                                    className="font-mono text-xs h-24"
                                    value={schemaJson}
                                    onChange={(e) => setSchemaJson(e.target.value)}
                                />
                            </div>
                        </div>

                        {/* Call Definition */}
                        <div className="space-y-4">
                            <h4 className="text-sm font-medium border-b pb-2">Upstream Call</h4>
                            <div className="grid grid-cols-4 gap-4">
                                <div className="col-span-1 space-y-2">
                                    <Label>Method</Label>
                                    <Select
                                        value={editingCall?.method?.toString() || "1"}
                                        onValueChange={(val) => setEditingCall(prev => ({ ...prev, method: parseInt(val) }))}
                                    >
                                        <SelectTrigger>
                                            <SelectValue />
                                        </SelectTrigger>
                                        <SelectContent>
                                            <SelectItem value="1">GET</SelectItem>
                                            <SelectItem value="2">POST</SelectItem>
                                            <SelectItem value="3">PUT</SelectItem>
                                            <SelectItem value="4">DELETE</SelectItem>
                                        </SelectContent>
                                    </Select>
                                </div>
                                <div className="col-span-3 space-y-2">
                                    <div className="flex justify-between items-center">
                                        <Label htmlFor="endpoint-path">Endpoint Path</Label>
                                        <span className="text-[10px] text-muted-foreground">Use <code>{"{{prop}}"}</code> for variables</span>
                                    </div>
                                    <Input
                                        id="endpoint-path"
                                        value={editingCall?.endpointPath || ""}
                                        onChange={(e) => setEditingCall(prev => ({ ...prev, endpointPath: e.target.value }))}
                                        placeholder="/v1/weather?city={{city}}"
                                    />
                                </div>
                            </div>
                        </div>
                    </div>

                    <DialogFooter>
                        <Button variant="outline" onClick={() => setIsDialogOpen(false)}>Cancel</Button>
                        <Button onClick={handleSave}>Save Tool</Button>
                    </DialogFooter>
                </DialogContent>
            </Dialog>
        </div>
    );
}
