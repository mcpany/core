/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { Switch } from "@/components/ui/switch";
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from "@/components/ui/select";
import { ToolDefinition } from "@proto/config/v1/tool";
import { HttpCallDefinition, HttpParameterMapping, ParameterSchema, ParameterType } from "@proto/config/v1/call";
import { Plus, Trash2, HelpCircle } from "lucide-react";
import {
    Table,
    TableBody,
    TableCell,
    TableHead,
    TableHeader,
    TableRow,
} from "@/components/ui/table";
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";

interface HttpToolEditorProps {
    initialTool: ToolDefinition;
    initialCall: HttpCallDefinition;
    onSave: (tool: ToolDefinition, call: HttpCallDefinition) => void;
    onCancel: () => void;
}

/**
 * A form editor for creating and updating HTTP tools.
 * Allows configuration of tool details, HTTP request method/path, and parameter mappings.
 * @param props The component props.
 * @param props.initialTool The initial tool definition.
 * @param props.initialCall The initial HTTP call definition.
 * @param props.onSave Callback when the tool is saved.
 * @param props.onCancel Callback when the edit is cancelled.
 * @returns The rendered component.
 */
export function HttpToolEditor({ initialTool, initialCall, onSave, onCancel }: HttpToolEditorProps) {
    const [tool, setTool] = useState<ToolDefinition>(initialTool);
    const [call, setCall] = useState<HttpCallDefinition>(initialCall);

    // Helper to update tool
    const updateTool = (updates: Partial<ToolDefinition>) => {
        setTool({ ...tool, ...updates });
    };

    // Helper to update call
    const updateCall = (updates: Partial<HttpCallDefinition>) => {
        setCall({ ...call, ...updates });
    };

    // Helper to update parameter
    const updateParameter = (index: number, updates: Partial<ParameterSchema>) => {
        const newParams = [...call.parameters];
        const currentMapping = newParams[index];
        if (currentMapping.schema) {
            newParams[index] = {
                ...currentMapping,
                schema: { ...currentMapping.schema, ...updates }
            };
            setCall({ ...call, parameters: newParams });
        }
    };

    const addParameter = () => {
        const newParam: HttpParameterMapping = {
            schema: {
                name: "new_param",
                description: "",
                type: ParameterType.STRING,
                isRequired: true,
                defaultValue: undefined
            },
            secret: undefined,
            disableEscape: false
        };
        setCall({ ...call, parameters: [...call.parameters, newParam] });
    };

    const removeParameter = (index: number) => {
        const newParams = [...call.parameters];
        newParams.splice(index, 1);
        setCall({ ...call, parameters: newParams });
    };

    const handleSave = () => {
        // Build Input Schema from parameters for the ToolDefinition
        // MCP tools need an input schema to tell the LLM what arguments to provide.
        const properties: Record<string, any> = {};
        const required: string[] = [];

        call.parameters.forEach(p => {
            if (p.schema) {
                properties[p.schema.name] = {
                    type: p.schema.type === ParameterType.STRING ? "string" :
                          p.schema.type === ParameterType.NUMBER ? "number" :
                          p.schema.type === ParameterType.INTEGER ? "integer" :
                          p.schema.type === ParameterType.BOOLEAN ? "boolean" :
                          p.schema.type === ParameterType.ARRAY ? "array" : "object",
                    description: p.schema.description
                };
                if (p.schema.isRequired) {
                    required.push(p.schema.name);
                }
            }
        });

        const inputSchema = {
            type: "object",
            properties,
            required
        };

        const updatedTool = {
            ...tool,
            inputSchema: inputSchema as any // Cast because protobuf Struct is complex
        };

        onSave(updatedTool, call);
    };

    return (
        <div className="space-y-6">
            <div className="grid grid-cols-2 gap-6">
                <div className="space-y-4">
                    <h3 className="text-lg font-medium">Tool Definition</h3>
                    <div className="space-y-2">
                        <Label htmlFor="tool-name">Name</Label>
                        <Input
                            id="tool-name"
                            value={tool.name}
                            onChange={(e) => updateTool({ name: e.target.value })}
                            placeholder="get_weather"
                        />
                        <p className="text-xs text-muted-foreground">The name used by the AI to call this tool.</p>
                    </div>
                    <div className="space-y-2">
                        <Label htmlFor="tool-desc">Description</Label>
                        <Textarea
                            id="tool-desc"
                            value={tool.description}
                            onChange={(e) => updateTool({ description: e.target.value })}
                            placeholder="Returns the current weather for a given city."
                        />
                        <p className="text-xs text-muted-foreground">Explain what this tool does to the AI.</p>
                    </div>
                </div>

                <div className="space-y-4">
                    <h3 className="text-lg font-medium">HTTP Request</h3>
                    <div className="grid grid-cols-4 gap-4">
                        <div className="col-span-1 space-y-2">
                            <Label htmlFor="method">Method</Label>
                            <Select
                                value={call.method.toString()}
                                onValueChange={(v) => updateCall({ method: parseInt(v) })}
                            >
                                <SelectTrigger id="method">
                                    <SelectValue />
                                </SelectTrigger>
                                <SelectContent>
                                    <SelectItem value="1">GET</SelectItem>
                                    <SelectItem value="2">POST</SelectItem>
                                    <SelectItem value="3">PUT</SelectItem>
                                    <SelectItem value="4">DELETE</SelectItem>
                                    <SelectItem value="5">PATCH</SelectItem>
                                </SelectContent>
                            </Select>
                        </div>
                        <div className="col-span-3 space-y-2">
                            <Label htmlFor="path">Endpoint Path</Label>
                            <Input
                                id="path"
                                value={call.endpointPath}
                                onChange={(e) => updateCall({ endpointPath: e.target.value })}
                                placeholder="/weather"
                            />
                        </div>
                    </div>
                    <p className="text-xs text-muted-foreground">
                        Use <code>{"{{param_name}}"}</code> to inject parameters into the path or query string.
                        Example: <code>/users/{"{{id}}"}</code> or <code>/search?q={"{{query}}"}</code>
                    </p>
                </div>
            </div>

            <div className="space-y-4">
                <div className="flex justify-between items-center">
                    <h3 className="text-lg font-medium">Parameters</h3>
                    <Button onClick={addParameter} size="sm" variant="outline">
                        <Plus className="mr-2 h-4 w-4" /> Add Parameter
                    </Button>
                </div>

                <div className="rounded-md border">
                    <Table>
                        <TableHeader>
                            <TableRow>
                                <TableHead className="w-[200px]">Name</TableHead>
                                <TableHead>Description</TableHead>
                                <TableHead className="w-[120px]">Type</TableHead>
                                <TableHead className="w-[80px] text-center">Required</TableHead>
                                <TableHead className="w-[50px]"></TableHead>
                            </TableRow>
                        </TableHeader>
                        <TableBody>
                            {call.parameters.length === 0 ? (
                                <TableRow>
                                    <TableCell colSpan={5} className="h-24 text-center text-muted-foreground">
                                        No parameters defined.
                                    </TableCell>
                                </TableRow>
                            ) : (
                                call.parameters.map((param, index) => (
                                    <TableRow key={index}>
                                        <TableCell>
                                            <Input
                                                value={param.schema?.name || ""}
                                                onChange={(e) => updateParameter(index, { name: e.target.value })}
                                                placeholder="param_name"
                                                className="h-8"
                                            />
                                        </TableCell>
                                        <TableCell>
                                            <Input
                                                value={param.schema?.description || ""}
                                                onChange={(e) => updateParameter(index, { description: e.target.value })}
                                                placeholder="Description"
                                                className="h-8"
                                            />
                                        </TableCell>
                                        <TableCell>
                                            <Select
                                                value={param.schema?.type.toString() || "0"}
                                                onValueChange={(v) => updateParameter(index, { type: parseInt(v) })}
                                            >
                                                <SelectTrigger className="h-8">
                                                    <SelectValue />
                                                </SelectTrigger>
                                                <SelectContent>
                                                    <SelectItem value="0">String</SelectItem>
                                                    <SelectItem value="1">Number</SelectItem>
                                                    <SelectItem value="2">Integer</SelectItem>
                                                    <SelectItem value="3">Boolean</SelectItem>
                                                </SelectContent>
                                            </Select>
                                        </TableCell>
                                        <TableCell className="text-center">
                                            <Switch
                                                checked={param.schema?.isRequired || false}
                                                onCheckedChange={(checked) => updateParameter(index, { isRequired: checked })}
                                            />
                                        </TableCell>
                                        <TableCell>
                                            <Button
                                                variant="ghost"
                                                size="icon"
                                                onClick={() => removeParameter(index)}
                                                className="h-8 w-8 text-destructive hover:text-destructive"
                                            >
                                                <Trash2 className="h-4 w-4" />
                                            </Button>
                                        </TableCell>
                                    </TableRow>
                                ))
                            )}
                        </TableBody>
                    </Table>
                </div>
            </div>

            <div className="flex justify-end gap-2 pt-4 border-t">
                <Button variant="outline" onClick={onCancel}>Cancel</Button>
                <Button onClick={handleSave}>Save Tool</Button>
            </div>
        </div>
    );
}
