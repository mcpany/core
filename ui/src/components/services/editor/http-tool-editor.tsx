/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect } from "react";
import { ToolDefinition } from "@proto/config/v1/tool";
import { HttpCallDefinition, HttpCallDefinition_HttpMethod, HttpParameterMapping, ParameterType } from "@proto/config/v1/call";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Switch } from "@/components/ui/switch";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Plus, Trash2 } from "lucide-react";

interface HttpToolEditorProps {
    initialTool?: ToolDefinition;
    initialCall?: HttpCallDefinition;
    onSave: (tool: ToolDefinition, call: HttpCallDefinition) => void;
    onCancel: () => void;
}

interface Parameter {
    name: string;
    description: string;
    type: ParameterType;
    required: boolean;
    defaultValue?: string;
}

/**
 * Editor form for a single HTTP tool.
 * Allows configuring the tool name, description, HTTP method, endpoint path, and parameters.
 *
 * @param props - The component props.
 * @param props.initialTool - The tool definition to edit (optional).
 * @param props.initialCall - The HTTP call definition to edit (optional).
 * @param props.onSave - Callback when the tool is saved.
 * @param props.onCancel - Callback when editing is cancelled.
 */
export function HttpToolEditor({ initialTool, initialCall, onSave, onCancel }: HttpToolEditorProps) {
    const [name, setName] = useState(initialTool?.name || "");
    const [description, setDescription] = useState(initialTool?.description || "");
    const [method, setMethod] = useState<HttpCallDefinition_HttpMethod>(initialCall?.method || HttpCallDefinition_HttpMethod.HTTP_METHOD_GET);
    const [endpointPath, setEndpointPath] = useState(initialCall?.endpointPath || "");

    // Parse initial parameters from Call Definition if available
    const [parameters, setParameters] = useState<Parameter[]>(() => {
        if (!initialCall?.parameters) return [];
        return initialCall.parameters.map(p => ({
            name: p.schema?.name || "",
            description: p.schema?.description || "",
            type: p.schema?.type || ParameterType.STRING,
            required: p.schema?.isRequired || false,
            defaultValue: p.schema?.defaultValue !== undefined ? String(p.schema.defaultValue) : undefined
        }));
    });

    const addParameter = () => {
        setParameters([...parameters, {
            name: "",
            description: "",
            type: ParameterType.STRING,
            required: true
        }]);
    };

    const updateParameter = (index: number, updates: Partial<Parameter>) => {
        const newParams = [...parameters];
        newParams[index] = { ...newParams[index], ...updates };
        setParameters(newParams);
    };

    const removeParameter = (index: number) => {
        const newParams = [...parameters];
        newParams.splice(index, 1);
        setParameters(newParams);
    };

    const handleSave = () => {
        // Generate UUID for call if new (client-side generation)
        const callId = initialCall?.id || crypto.randomUUID();

        // Construct ToolDefinition
        // Generate JSON Schema for input_schema
        const properties: Record<string, any> = {};
        const required: string[] = [];

        parameters.forEach(p => {
            properties[p.name] = {
                type: getJsonSchemaType(p.type),
                description: p.description
            };
            if (p.required) required.push(p.name);
        });

        const inputSchema = {
            type: "object",
            properties,
            required: required.length > 0 ? required : undefined
        };

        const tool: ToolDefinition = {
            name,
            description,
            inputSchema: inputSchema as any, // Cast to Struct compatible type
            callId: callId,
            serviceId: "", // Will be filled by backend or context
            isStream: false,
            title: name,
            readOnlyHint: method === HttpCallDefinition_HttpMethod.HTTP_METHOD_GET,
            destructiveHint: method !== HttpCallDefinition_HttpMethod.HTTP_METHOD_GET,
            idempotentHint: method === HttpCallDefinition_HttpMethod.HTTP_METHOD_GET || method === HttpCallDefinition_HttpMethod.HTTP_METHOD_PUT,
            openWorldHint: true,
            disable: false,
            profiles: [],
            mergeStrategy: 0,
            tags: [],
            integrity: undefined
        };

        // Construct HttpCallDefinition
        const httpParameters: HttpParameterMapping[] = parameters.map(p => ({
            schema: {
                name: p.name,
                description: p.description,
                type: p.type,
                isRequired: p.required,
                defaultValue: p.defaultValue // Simple handling for now
            },
            disableEscape: false,
            secret: undefined
        }));

        const call: HttpCallDefinition = {
            id: callId,
            endpointPath,
            method,
            parameters: httpParameters,
            inputTransformer: undefined,
            outputTransformer: undefined,
            cache: undefined,
            inputSchema: inputSchema as any, // Also store in call definition if needed
            outputSchema: undefined
        };

        onSave(tool, call);
    };

    const getJsonSchemaType = (type: ParameterType) => {
        switch (type) {
            case ParameterType.STRING: return "string";
            case ParameterType.NUMBER: return "number";
            case ParameterType.INTEGER: return "integer";
            case ParameterType.BOOLEAN: return "boolean";
            case ParameterType.ARRAY: return "array";
            case ParameterType.OBJECT: return "object";
            default: return "string";
        }
    };

    return (
        <div className="space-y-6">
            <div className="grid grid-cols-2 gap-4">
                <div className="space-y-2">
                    <Label htmlFor="tool-name">Tool Name</Label>
                    <Input
                        id="tool-name"
                        value={name}
                        onChange={(e) => setName(e.target.value)}
                        placeholder="e.g., get_user"
                    />
                </div>
                <div className="space-y-2">
                    <Label htmlFor="tool-method">HTTP Method</Label>
                    <Select
                        value={method.toString()}
                        onValueChange={(val) => setMethod(parseInt(val))}
                    >
                        <SelectTrigger id="tool-method">
                            <SelectValue />
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
            </div>

            <div className="space-y-2">
                <Label htmlFor="tool-desc">Description</Label>
                <Textarea
                    id="tool-desc"
                    value={description}
                    onChange={(e) => setDescription(e.target.value)}
                    placeholder="Describe what this tool does..."
                />
            </div>

            <div className="space-y-2">
                <Label htmlFor="endpoint-path">Endpoint Path</Label>
                <Input
                    id="endpoint-path"
                    value={endpointPath}
                    onChange={(e) => setEndpointPath(e.target.value)}
                    placeholder="/api/v1/resource/{id}"
                />
                <p className="text-xs text-muted-foreground">Use <code>{`{paramName}`}</code> for path parameters.</p>
            </div>

            <div className="space-y-2">
                <div className="flex items-center justify-between">
                    <Label>Parameters</Label>
                    <Button variant="outline" size="sm" onClick={addParameter}>
                        <Plus className="mr-2 h-3 w-3" /> Add Parameter
                    </Button>
                </div>

                <div className="border rounded-md">
                    <Table>
                        <TableHeader>
                            <TableRow>
                                <TableHead>Name</TableHead>
                                <TableHead>Type</TableHead>
                                <TableHead>Required</TableHead>
                                <TableHead>Description</TableHead>
                                <TableHead className="w-[50px]"></TableHead>
                            </TableRow>
                        </TableHeader>
                        <TableBody>
                            {parameters.map((param, index) => (
                                <TableRow key={index}>
                                    <TableCell>
                                        <Input
                                            value={param.name}
                                            onChange={(e) => updateParameter(index, { name: e.target.value })}
                                            placeholder="param_name"
                                            className="h-8"
                                        />
                                    </TableCell>
                                    <TableCell>
                                        <Select
                                            value={param.type.toString()}
                                            onValueChange={(val) => updateParameter(index, { type: parseInt(val) })}
                                        >
                                            <SelectTrigger className="h-8">
                                                <SelectValue />
                                            </SelectTrigger>
                                            <SelectContent>
                                                <SelectItem value={ParameterType.STRING.toString()}>String</SelectItem>
                                                <SelectItem value={ParameterType.NUMBER.toString()}>Number</SelectItem>
                                                <SelectItem value={ParameterType.INTEGER.toString()}>Integer</SelectItem>
                                                <SelectItem value={ParameterType.BOOLEAN.toString()}>Boolean</SelectItem>
                                            </SelectContent>
                                        </Select>
                                    </TableCell>
                                    <TableCell>
                                        <Switch
                                            checked={param.required}
                                            onCheckedChange={(checked) => updateParameter(index, { required: checked })}
                                        />
                                    </TableCell>
                                    <TableCell>
                                        <Input
                                            value={param.description}
                                            onChange={(e) => updateParameter(index, { description: e.target.value })}
                                            placeholder="Description"
                                            className="h-8"
                                        />
                                    </TableCell>
                                    <TableCell>
                                        <Button
                                            variant="ghost"
                                            size="icon"
                                            className="h-8 w-8 text-destructive hover:text-destructive"
                                            onClick={() => removeParameter(index)}
                                        >
                                            <Trash2 className="h-4 w-4" />
                                        </Button>
                                    </TableCell>
                                </TableRow>
                            ))}
                            {parameters.length === 0 && (
                                <TableRow>
                                    <TableCell colSpan={5} className="text-center text-muted-foreground py-4">
                                        No parameters defined.
                                    </TableCell>
                                </TableRow>
                            )}
                        </TableBody>
                    </Table>
                </div>
            </div>

            <div className="flex justify-end gap-2 pt-4 border-t">
                <Button variant="outline" onClick={onCancel}>Cancel</Button>
                <Button onClick={handleSave} disabled={!name || !endpointPath}>Save Tool</Button>
            </div>
        </div>
    );
}
