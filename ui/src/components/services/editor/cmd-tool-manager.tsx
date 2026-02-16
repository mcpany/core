/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState } from "react";
import { UpstreamServiceConfig } from "@/lib/client";
import { ToolDefinition } from "@proto/config/v1/tool";
import { CommandLineCallDefinition } from "@proto/config/v1/call";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Plus, Trash2, Edit } from "lucide-react";
import {
    Sheet,
    SheetContent,
    SheetDescription,
    SheetHeader,
    SheetTitle,
} from "@/components/ui/sheet";
import { CmdToolEditor } from "./cmd-tool-editor";
import { Badge } from "@/components/ui/badge";

interface CmdToolManagerProps {
    service: UpstreamServiceConfig;
    onChange: (service: UpstreamServiceConfig) => void;
}

/**
 * Manager component for Command Line tools within a service.
 * Displays a list of tools and allows adding, editing, and deleting them.
 * @param props - The component props.
 * @returns The rendered tool manager.
 */
export function CmdToolManager({ service, onChange }: CmdToolManagerProps) {
    const [editingToolIndex, setEditingToolIndex] = useState<number | null>(null);
    const [isSheetOpen, setIsSheetOpen] = useState(false);

    // Helper to safely get tools/calls
    const tools = service.commandLineService?.tools || [];
    const calls = service.commandLineService?.calls || {};

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
            serviceId: "",
        };

        const newCall: CommandLineCallDefinition = {
            id: callId,
            args: [],
            parameters: [],
            cache: undefined,
            inputSchema: undefined,
            outputSchema: undefined,
        };

        const newTools = [...tools, newTool];
        const newCalls = { ...calls, [callId]: newCall };

        onChange({
            ...service,
            commandLineService: {
                ...service.commandLineService!,
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
            commandLineService: {
                ...service.commandLineService!,
                tools: newTools,
                calls: newCalls
            }
        });
    };

    const handleToolChange = (updatedTool: ToolDefinition, updatedCall: CommandLineCallDefinition) => {
        if (editingToolIndex === null) return;

        const newTools = [...tools];
        newTools[editingToolIndex] = updatedTool;

        const newCalls = { ...calls };
        if (updatedCall.id !== updatedTool.callId) {
             // Should not happen
        }
        newCalls[updatedCall.id] = updatedCall;

        onChange({
            ...service,
            commandLineService: {
                ...service.commandLineService!,
                tools: newTools,
                calls: newCalls
            }
        });
    };

    const getCallForTool = (tool: ToolDefinition): CommandLineCallDefinition => {
        return calls[tool.callId] || {
            id: tool.callId,
            args: [],
            parameters: [],
        } as CommandLineCallDefinition;
    };

    return (
        <div className="space-y-4">
            <div className="flex items-center justify-between">
                <div>
                    <h3 className="text-lg font-medium">Defined Tools</h3>
                    <p className="text-sm text-muted-foreground">
                        Define the tools exposed by this Command Line service.
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
                                    <Badge variant="outline" className="bg-slate-500/10 text-slate-500 border-slate-500/20">
                                        CLI
                                    </Badge>
                                </div>
                                <div className="text-sm text-muted-foreground font-mono truncate max-w-[400px]">
                                    {service.commandLineService?.command} {call.args ? call.args.join(" ") : ""}
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
                            Configure the tool definition and its command line arguments.
                        </SheetDescription>
                    </SheetHeader>
                    <div className="mt-6">
                        {editingToolIndex !== null && tools[editingToolIndex] && (
                            <CmdToolEditor
                                tool={tools[editingToolIndex]}
                                call={getCallForTool(tools[editingToolIndex])}
                                onChange={handleToolChange}
                            />
                        )}
                    </div>
                </SheetContent>
            </Sheet>
        </div>
    );
}
