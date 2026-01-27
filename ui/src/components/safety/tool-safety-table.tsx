/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState } from "react";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Switch } from "@/components/ui/switch";
import { ToolDefinition } from "@/lib/types";
import { apiClient } from "@/lib/client";
import { useToast } from "@/hooks/use-toast";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Wrench } from "lucide-react";

// Extend the base definition to include runtime status
type ToolWithStatus = ToolDefinition & { disable?: boolean };

interface ToolSafetyTableProps {
    tools?: ToolDefinition[];
    onUpdate?: () => void;
}

/**
 * ToolSafetyTable component.
 * @param props - The component props.
 * @param props.tools - The tools property.
 * @param props.onUpdate - The onUpdate property.
 * @returns The rendered component.
 */
export function ToolSafetyTable({ tools, onUpdate }: ToolSafetyTableProps) {
    const { toast } = useToast();
    const [loading, setLoading] = useState<Record<string, boolean>>({});

    const handleToggle = async (toolName: string, isCurrentlyEnabled: boolean) => {
        setLoading(prev => ({ ...prev, [toolName]: true }));
        try {
            // If currently enabled, we want to disable (disable=true)
            // If currently disabled, we want to enable (disable=false)
            await apiClient.setToolStatus(toolName, isCurrentlyEnabled);

            toast({
                title: isCurrentlyEnabled ? "Tool Disabled" : "Tool Enabled",
                description: `Tool ${toolName} has been ${isCurrentlyEnabled ? 'disabled' : 'enabled'}.`,
            });
            onUpdate?.();
        } catch (error: any) {
            toast({
                variant: "destructive",
                title: "Failed to update status",
                description: error.message,
            });
        } finally {
            setLoading(prev => ({ ...prev, [toolName]: false }));
        }
    };

    if (!tools || tools.length === 0) {
        return (
            <Card>
                <CardHeader>
                    <CardTitle className="flex items-center gap-2"><Wrench className="h-5 w-5" /> Tool Safety</CardTitle>
                    <CardDescription>No tools found for this service.</CardDescription>
                </CardHeader>
            </Card>
        );
    }

    return (
        <Card>
            <CardHeader>
                <CardTitle className="flex items-center gap-2"><Wrench className="h-5 w-5" /> Tool Safety</CardTitle>
                <CardDescription>
                    Control which tools are exposed to the LLM. Disabling a tool prevents it from being called.
                </CardDescription>
            </CardHeader>
            <CardContent>
                <Table>
                    <TableHeader>
                        <TableRow>
                            <TableHead>Status</TableHead>
                            <TableHead>Name</TableHead>
                            <TableHead>Description</TableHead>
                        </TableRow>
                    </TableHeader>
                    <TableBody>
                        {(tools as ToolWithStatus[]).map((tool) => {
                            // Assuming tool has a 'disable' property. If not, default to false (enabled).
                            const isDisabled = tool.disable || false;
                            const isEnabled = !isDisabled;

                            return (
                                <TableRow key={tool.name}>
                                    <TableCell className="w-[150px]">
                                        <div className="flex items-center space-x-2">
                                            <Switch
                                                checked={isEnabled}
                                                onCheckedChange={() => handleToggle(tool.name, isEnabled)}
                                                disabled={loading[tool.name]}
                                            />
                                            <span className="text-xs text-muted-foreground w-[50px]">
                                                {isEnabled ? "Enabled" : "Disabled"}
                                            </span>
                                        </div>
                                    </TableCell>
                                    <TableCell className="font-medium">
                                        {tool.name}
                                        {isDisabled && <Badge variant="destructive" className="ml-2 text-[10px]">Blocked</Badge>}
                                    </TableCell>
                                    <TableCell className="text-muted-foreground text-sm">
                                        {tool.description || "No description"}
                                    </TableCell>
                                </TableRow>
                            );
                        })}
                    </TableBody>
                </Table>
            </CardContent>
        </Card>
    );
}
