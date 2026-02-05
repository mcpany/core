/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState } from "react";
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Plus, Trash2 } from "lucide-react";
import { Authentication } from "@/lib/client";

/**
 * Definition of updates to apply to multiple services.
 */
export interface BulkUpdates {
    /** Tags to add to the services. */
    tags?: string[];
    /** Timeout duration string (e.g. "30s") to set for resilience policy. */
    timeout?: string;
    /** Environment variables to add or update (for CLI/MCP services). */
    env?: Record<string, string>;
    /** Authentication configuration to set. */
    upstreamAuth?: Authentication;
}

interface BulkEditDialogProps {
    open: boolean;
    onOpenChange: (open: boolean) => void;
    selectedCount: number;
    onApply: (updates: BulkUpdates) => void;
}

/**
 * Dialog for editing multiple services at once.
 * Allows updating tags, timeouts, environment variables, and authentication.
 *
 * @param props - The component props.
 * @param props.open - Whether the dialog is open.
 * @param props.onOpenChange - Callback when open state changes.
 * @param props.selectedCount - Number of selected services.
 * @param props.onApply - Callback when changes are applied.
 * @returns The rendered component.
 */
export function BulkEditDialog({ open, onOpenChange, selectedCount, onApply }: BulkEditDialogProps) {
    const [tags, setTags] = useState("");
    const [timeoutSeconds, setTimeoutSeconds] = useState("");
    const [envVars, setEnvVars] = useState<{ key: string, value: string }[]>([]);

    // Auth State
    const [authType, setAuthType] = useState<string>("none");
    const [authValue, setAuthValue] = useState("");
    const [authHeader, setAuthHeader] = useState("X-API-Key");

    const addEnvVar = () => {
        setEnvVars([...envVars, { key: "", value: "" }]);
    };

    const removeEnvVar = (index: number) => {
        setEnvVars(envVars.filter((_, i) => i !== index));
    };

    const updateEnvVar = (index: number, field: "key" | "value", value: string) => {
        const newEnv = [...envVars];
        newEnv[index][field] = value;
        setEnvVars(newEnv);
    };

    const handleApply = () => {
        const updates: BulkUpdates = {};

        // Tags
        if (tags.trim()) {
            updates.tags = tags.split(",").map(t => t.trim()).filter(Boolean);
        }

        // Timeout
        if (timeoutSeconds) {
            updates.timeout = `${timeoutSeconds}s`;
        }

        // Env Vars
        if (envVars.length > 0) {
            const envMap: Record<string, string> = {};
            envVars.forEach(e => {
                if (e.key) envMap[e.key] = e.value;
            });
            if (Object.keys(envMap).length > 0) {
                updates.env = envMap;
            }
        }

        // Auth
        if (authType !== "none") {
            if (authType === "api_key") {
                updates.upstreamAuth = {
                    apiKey: {
                        in: 0, // HEADER
                        paramName: authHeader,
                        value: { plainText: authValue },
                        verificationValue: "" // Not needed for upstream
                    }
                };
            } else if (authType === "bearer") {
                updates.upstreamAuth = {
                    bearerToken: {
                        token: { plainText: authValue }
                    }
                };
            }
        }

        onApply(updates);
        onOpenChange(false);

        // Reset state? Maybe keep for next time or reset.
        // Let's reset to avoid confusion.
        setTags("");
        setTimeoutSeconds("");
        setEnvVars([]);
        setAuthType("none");
        setAuthValue("");
    };

    return (
        <Dialog open={open} onOpenChange={onOpenChange}>
            <DialogContent className="sm:max-w-[500px]">
                <DialogHeader>
                    <DialogTitle>Bulk Edit Services</DialogTitle>
                    <DialogDescription>
                        Apply changes to {selectedCount} selected services. Empty fields will remain unchanged.
                    </DialogDescription>
                </DialogHeader>

                <Tabs defaultValue="general" className="w-full">
                    <TabsList className="grid w-full grid-cols-3">
                        <TabsTrigger value="general">General</TabsTrigger>
                        <TabsTrigger value="environment">Environment</TabsTrigger>
                        <TabsTrigger value="auth">Authentication</TabsTrigger>
                    </TabsList>

                    <TabsContent value="general" className="space-y-4 py-4">
                        <div className="space-y-2">
                            <Label htmlFor="bulk-tags">Add Tags (comma separated)</Label>
                            <Input
                                id="bulk-tags"
                                placeholder="production, web, internal"
                                value={tags}
                                onChange={(e) => setTags(e.target.value)}
                            />
                        </div>
                        <div className="space-y-2">
                            <Label htmlFor="bulk-timeout">Timeout (seconds)</Label>
                            <Input
                                id="bulk-timeout"
                                type="number"
                                placeholder="30"
                                value={timeoutSeconds}
                                onChange={(e) => setTimeoutSeconds(e.target.value)}
                            />
                            <p className="text-xs text-muted-foreground">
                                Updates the resilience timeout policy.
                            </p>
                        </div>
                    </TabsContent>

                    <TabsContent value="environment" className="space-y-4 py-4">
                        <div className="space-y-2">
                             <div className="flex justify-between items-center">
                                <Label>Environment Variables</Label>
                                <Button variant="outline" size="sm" onClick={addEnvVar}>
                                    <Plus className="h-4 w-4 mr-2" /> Add
                                </Button>
                             </div>
                             <p className="text-xs text-muted-foreground">
                                Added to CLI and MCP Stdio services only.
                            </p>
                             <div className="space-y-2 max-h-[200px] overflow-y-auto">
                                {envVars.map((env, i) => (
                                    <div key={i} className="flex gap-2 items-center">
                                        <Input
                                            placeholder="Key"
                                            value={env.key}
                                            onChange={(e) => updateEnvVar(i, "key", e.target.value)}
                                            className="flex-1"
                                        />
                                        <Input
                                            placeholder="Value"
                                            value={env.value}
                                            onChange={(e) => updateEnvVar(i, "value", e.target.value)}
                                            className="flex-1"
                                        />
                                        <Button variant="ghost" size="icon" onClick={() => removeEnvVar(i)}>
                                            <Trash2 className="h-4 w-4 text-muted-foreground hover:text-destructive" />
                                        </Button>
                                    </div>
                                ))}
                                {envVars.length === 0 && (
                                    <div className="text-center text-sm text-muted-foreground py-4 border border-dashed rounded-md">
                                        No environment variables added.
                                    </div>
                                )}
                             </div>
                        </div>
                    </TabsContent>

                    <TabsContent value="auth" className="space-y-4 py-4">
                        <div className="space-y-2">
                            <Label>Authentication Method</Label>
                            <Select value={authType} onValueChange={setAuthType}>
                                <SelectTrigger>
                                    <SelectValue placeholder="Select Auth Type" />
                                </SelectTrigger>
                                <SelectContent>
                                    <SelectItem value="none">No Change</SelectItem>
                                    <SelectItem value="api_key">API Key</SelectItem>
                                    <SelectItem value="bearer">Bearer Token</SelectItem>
                                </SelectContent>
                            </Select>
                        </div>

                        {authType === "api_key" && (
                            <div className="space-y-4 animate-in fade-in zoom-in-95">
                                <div className="space-y-2">
                                    <Label>Header Name</Label>
                                    <Input
                                        value={authHeader}
                                        onChange={(e) => setAuthHeader(e.target.value)}
                                        placeholder="X-API-Key"
                                    />
                                </div>
                                <div className="space-y-2">
                                    <Label>API Key</Label>
                                    <Input
                                        type="password"
                                        value={authValue}
                                        onChange={(e) => setAuthValue(e.target.value)}
                                        placeholder="sk-..."
                                    />
                                </div>
                            </div>
                        )}

                        {authType === "bearer" && (
                            <div className="space-y-4 animate-in fade-in zoom-in-95">
                                <div className="space-y-2">
                                    <Label>Token</Label>
                                    <Input
                                        type="password"
                                        value={authValue}
                                        onChange={(e) => setAuthValue(e.target.value)}
                                        placeholder="eyJ..."
                                    />
                                </div>
                            </div>
                        )}
                        {authType !== "none" && (
                             <p className="text-xs text-muted-foreground text-amber-600">
                                Warning: This will overwrite existing authentication configuration for selected services.
                            </p>
                        )}
                    </TabsContent>
                </Tabs>

                <DialogFooter>
                    <Button variant="outline" onClick={() => onOpenChange(false)}>Cancel</Button>
                    <Button onClick={handleApply}>Apply Changes</Button>
                </DialogFooter>
            </DialogContent>
        </Dialog>
    );
}
