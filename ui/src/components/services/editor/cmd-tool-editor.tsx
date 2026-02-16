/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect } from "react";
import { ToolDefinition } from "@proto/config/v1/tool";
import { CommandLineCallDefinition, CommandLineParameterMapping, ParameterType } from "@proto/config/v1/call";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Button } from "@/components/ui/button";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Trash2, Plus, GripVertical, Sparkles } from "lucide-react";
import { Card, CardContent } from "@/components/ui/card";
import { Switch } from "@/components/ui/switch";
import { Badge } from "@/components/ui/badge";
import { DragDropContext, Droppable, Draggable } from "@hello-pangea/dnd";

interface CmdToolEditorProps {
    tool: ToolDefinition;
    call: CommandLineCallDefinition;
    onChange: (tool: ToolDefinition, call: CommandLineCallDefinition) => void;
}

/**
 * Editor for configuring a single Command Line tool.
 * Allows defining tool metadata, command arguments, and input parameters.
 * @param props - The component props.
 * @returns The rendered tool editor.
 */
export function CmdToolEditor({ tool, call, onChange }: CmdToolEditorProps) {
    const [localTool, setLocalTool] = useState<ToolDefinition>(tool);
    const [localCall, setLocalCall] = useState<CommandLineCallDefinition>(call);

    useEffect(() => {
        setLocalTool(tool);
        setLocalCall(call);
    }, [tool, call]);

    const updateTool = (updates: Partial<ToolDefinition>) => {
        const newTool = { ...localTool, ...updates };
        setLocalTool(newTool);
        onChange(newTool, localCall);
    };

    const updateCall = (updates: Partial<CommandLineCallDefinition>) => {
        const newCall = { ...localCall, ...updates };
        setLocalCall(newCall);
        onChange(localTool, newCall);
    };

    // --- Arguments Management ---

    const addArgument = () => {
        const newArgs = [...(localCall.args || []), ""];
        updateCall({ args: newArgs });
    };

    const updateArgument = (index: number, value: string) => {
        const newArgs = [...(localCall.args || [])];
        newArgs[index] = value;
        updateCall({ args: newArgs });
    };

    const removeArgument = (index: number) => {
        const newArgs = [...(localCall.args || [])];
        newArgs.splice(index, 1);
        updateCall({ args: newArgs });
    };

    const onDragEnd = (result: any) => {
        if (!result.destination) return;
        const newArgs = Array.from(localCall.args || []);
        const [reorderedItem] = newArgs.splice(result.source.index, 1);
        newArgs.splice(result.destination.index, 0, reorderedItem);
        updateCall({ args: newArgs });
    };

    // --- Parameters Management ---

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
                secret: undefined
            } as CommandLineParameterMapping
        ];
        updateCall({ parameters: newParams });
    };

    const updateParameterSchema = (index: number, updates: any) => {
        const newParams = [...(localCall.parameters || [])];
        const currentSchema = newParams[index].schema || {};
        newParams[index] = {
            ...newParams[index],
            schema: { ...currentSchema, ...updates }
        };
        updateCall({ parameters: newParams });
    };

    const removeParameter = (index: number) => {
        const newParams = [...(localCall.parameters || [])];
        newParams.splice(index, 1);
        updateCall({ parameters: newParams });
    };

    const scanArgumentsForParameters = () => {
        const args = localCall.args || [];
        const existingParams = new Set((localCall.parameters || []).map(p => p.schema?.name));
        const newParams = [...(localCall.parameters || [])];
        let addedCount = 0;

        const regex = /\{\{([^}]+)\}\}/g;
        args.forEach(arg => {
            let match;
            while ((match = regex.exec(arg)) !== null) {
                const paramName = match[1].trim();
                if (paramName && !existingParams.has(paramName)) {
                    newParams.push({
                        schema: {
                            name: paramName,
                            description: `Auto-generated from argument`,
                            type: ParameterType.STRING,
                            isRequired: true,
                        },
                        secret: undefined
                    } as CommandLineParameterMapping);
                    existingParams.add(paramName);
                    addedCount++;
                }
            }
        });

        if (addedCount > 0) {
            updateCall({ parameters: newParams });
        }
    };

    return (
        <div className="space-y-8">
            {/* Metadata Section */}
            <div className="space-y-4">
                <div className="grid grid-cols-2 gap-4">
                    <div className="space-y-2">
                        <Label htmlFor="tool-name">Tool Name</Label>
                        <Input
                            id="tool-name"
                            value={localTool.name}
                            onChange={(e) => updateTool({ name: e.target.value })}
                            placeholder="list_pods"
                        />
                         <p className="text-[10px] text-muted-foreground">The name used by the AI to invoke this tool.</p>
                    </div>
                     <div className="space-y-2">
                        <Label htmlFor="tool-description">Description</Label>
                        <Input
                            id="tool-description"
                            value={localTool.description}
                            onChange={(e) => updateTool({ description: e.target.value })}
                            placeholder="List all pods in the cluster"
                        />
                    </div>
                </div>
            </div>

            {/* Arguments Section */}
            <div className="space-y-4">
                 <div className="flex items-center justify-between">
                    <div>
                        <Label className="text-base">Command Arguments</Label>
                        <p className="text-sm text-muted-foreground">
                            Construct the command arguments. Use <code>{"{{ param }}"}</code> to insert parameters.
                        </p>
                    </div>
                    <Button variant="outline" size="sm" onClick={addArgument}>
                        <Plus className="mr-2 h-4 w-4" /> Add Argument
                    </Button>
                </div>

                <div className="bg-muted/30 rounded-md p-4">
                    <DragDropContext onDragEnd={onDragEnd}>
                        <Droppable droppableId="args-list">
                            {(provided) => (
                                <div
                                    {...provided.droppableProps}
                                    ref={provided.innerRef}
                                    className="space-y-2"
                                >
                                    {(localCall.args || []).map((arg, index) => (
                                        <Draggable key={index} draggableId={`arg-${index}`} index={index}>
                                            {(provided) => (
                                                <div
                                                    ref={provided.innerRef}
                                                    {...provided.draggableProps}
                                                    className="flex items-center gap-2"
                                                >
                                                     <div {...provided.dragHandleProps} className="text-muted-foreground cursor-grab">
                                                        <GripVertical className="h-4 w-4" />
                                                    </div>
                                                    <Input
                                                        value={arg}
                                                        onChange={(e) => updateArgument(index, e.target.value)}
                                                        placeholder={`Argument ${index + 1}`}
                                                        className="font-mono text-sm"
                                                    />
                                                     <Button
                                                        variant="ghost"
                                                        size="icon"
                                                        className="h-8 w-8 text-muted-foreground hover:text-destructive"
                                                        onClick={() => removeArgument(index)}
                                                    >
                                                        <Trash2 className="h-4 w-4" />
                                                    </Button>
                                                </div>
                                            )}
                                        </Draggable>
                                    ))}
                                    {provided.placeholder}
                                </div>
                            )}
                        </Droppable>
                    </DragDropContext>
                     {(!localCall.args || localCall.args.length === 0) && (
                        <div className="text-center py-4 text-muted-foreground text-sm">
                            No arguments defined. The base command will be executed as is.
                        </div>
                    )}
                </div>
            </div>

            {/* Parameters Section */}
            <div className="space-y-4">
                <div className="flex items-center justify-between">
                    <div>
                         <Label className="text-base">Input Parameters</Label>
                         <p className="text-sm text-muted-foreground">
                            Define the inputs accepted by this tool.
                        </p>
                    </div>
                    <div className="flex gap-2">
                        <Button variant="outline" size="sm" onClick={scanArgumentsForParameters}>
                            <Sparkles className="mr-2 h-4 w-4" /> Scan Args
                        </Button>
                        <Button variant="outline" size="sm" onClick={addParameter}>
                            <Plus className="mr-2 h-4 w-4" /> Add Parameter
                        </Button>
                    </div>
                </div>

                <div className="space-y-3">
                    {localCall.parameters?.map((param, index) => (
                        <Card key={index} className="relative group">
                             <Button
                                variant="ghost"
                                size="icon"
                                className="absolute right-2 top-2 h-6 w-6 text-muted-foreground hover:text-destructive opacity-0 group-hover:opacity-100 transition-opacity"
                                onClick={() => removeParameter(index)}
                            >
                                <Trash2 className="h-4 w-4" />
                            </Button>
                            <CardContent className="p-4 grid grid-cols-12 gap-4 items-start">
                                <div className="col-span-3 space-y-1">
                                    <Label htmlFor={`param-name-${index}`} className="text-xs">Name</Label>
                                    <Input
                                        id={`param-name-${index}`}
                                        value={param.schema?.name}
                                        onChange={(e) => updateParameterSchema(index, { name: e.target.value })}
                                        placeholder="param_name"
                                        className="h-8"
                                    />
                                </div>
                                <div className="col-span-3 space-y-1">
                                    <Label htmlFor={`param-type-${index}`} className="text-xs">Type</Label>
                                    <Select
                                        value={param.schema?.type?.toString()}
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
                                <div className="col-span-4 space-y-1">
                                    <Label htmlFor={`param-desc-${index}`} className="text-xs">Description</Label>
                                    <Input
                                        id={`param-desc-${index}`}
                                        value={param.schema?.description}
                                        onChange={(e) => updateParameterSchema(index, { description: e.target.value })}
                                        placeholder="Description"
                                        className="h-8"
                                    />
                                </div>
                                <div className="col-span-2 flex items-center justify-center pt-6 space-x-2">
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
                </div>
            </div>
        </div>
    );
}
