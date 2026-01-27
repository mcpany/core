/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState } from "react";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Switch } from "@/components/ui/switch";
import { ResourceDefinition } from "@/lib/types";
import { apiClient } from "@/lib/client";
import { useToast } from "@/hooks/use-toast";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Database } from "lucide-react";

// Extend the base definition to include runtime status
type ResourceWithStatus = ResourceDefinition & { disable?: boolean };

interface ResourceSafetyTableProps {
    resources?: ResourceDefinition[];
    onUpdate?: () => void;
}

/**
 * ResourceSafetyTable displays a table of resources and allows toggling their enabled/disabled status.
 *
 * @param props - The component props.
 * @param props.resources - The list of resources to display.
 * @param props.onUpdate - Callback function called when a resource status is updated.
 * @returns A table component for managing resource safety.
 */
export function ResourceSafetyTable({ resources, onUpdate }: ResourceSafetyTableProps) {
    const { toast } = useToast();
    const [loading, setLoading] = useState<Record<string, boolean>>({});

    const handleToggle = async (uri: string, isCurrentlyEnabled: boolean) => {
        setLoading(prev => ({ ...prev, [uri]: true }));
        try {
            await apiClient.setResourceStatus(uri, isCurrentlyEnabled);
            toast({
                title: isCurrentlyEnabled ? "Resource Disabled" : "Resource Enabled",
                description: `Resource has been ${isCurrentlyEnabled ? 'disabled' : 'enabled'}.`,
            });
            onUpdate?.();
        } catch (error: any) {
            toast({
                variant: "destructive",
                title: "Failed to update status",
                description: error.message,
            });
        } finally {
            setLoading(prev => ({ ...prev, [uri]: false }));
        }
    };

    if (!resources || resources.length === 0) {
        return (
            <Card>
                <CardHeader>
                    <CardTitle className="flex items-center gap-2"><Database className="h-5 w-5" /> Resource Safety</CardTitle>
                    <CardDescription>No resources found for this service.</CardDescription>
                </CardHeader>
            </Card>
        );
    }

    return (
        <Card>
            <CardHeader>
                <CardTitle className="flex items-center gap-2"><Database className="h-5 w-5" /> Resource Safety</CardTitle>
                <CardDescription>
                    Control access to resources.
                </CardDescription>
            </CardHeader>
            <CardContent>
                <Table>
                    <TableHeader>
                        <TableRow>
                            <TableHead>Status</TableHead>
                            <TableHead>Name</TableHead>
                            <TableHead>URI</TableHead>
                            <TableHead>Description</TableHead>
                        </TableRow>
                    </TableHeader>
                    <TableBody>
                        {(resources as ResourceWithStatus[]).map((resource) => {
                            const isDisabled = resource.disable || false;
                            const isEnabled = !isDisabled;

                            return (
                                <TableRow key={resource.uri}>
                                    <TableCell className="w-[150px]">
                                        <div className="flex items-center space-x-2">
                                            <Switch
                                                checked={isEnabled}
                                                onCheckedChange={() => handleToggle(resource.uri, isEnabled)}
                                                disabled={loading[resource.uri]}
                                            />
                                            <span className="text-xs text-muted-foreground w-[50px]">
                                                {isEnabled ? "Enabled" : "Disabled"}
                                            </span>
                                        </div>
                                    </TableCell>
                                    <TableCell className="font-medium">
                                        {resource.name}
                                        {isDisabled && <Badge variant="destructive" className="ml-2 text-[10px]">Blocked</Badge>}
                                    </TableCell>
                                    <TableCell className="font-mono text-xs text-muted-foreground truncate max-w-[200px]" title={resource.uri}>
                                        {resource.uri}
                                    </TableCell>
                                    <TableCell className="text-muted-foreground text-sm">
                                        {resource.description || "No description"}
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
