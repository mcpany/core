/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { EnvVarEditor } from "@/components/services/env-var-editor";
import { SecretValue } from "@proto/config/v1/auth";
import { DialogFooter } from "@/components/ui/dialog";

/**
 * BulkUpdates interface.
 * Defines the structure for bulk updates to services.
 */
export interface BulkUpdates {
    tags?: string[];
    env?: Record<string, SecretValue>;
}

interface BulkServiceEditorProps {
    selectedCount: number;
    onApply: (updates: BulkUpdates) => void;
    onCancel: () => void;
}

/**
 * BulkServiceEditor component.
 * Allows editing multiple services at once.
 *
 * @param props - Component props.
 */
export function BulkServiceEditor({ selectedCount, onApply, onCancel }: BulkServiceEditorProps) {
    const [tags, setTags] = useState("");
    const [env, setEnv] = useState<Record<string, SecretValue>>({});

    const handleApply = () => {
        const updates: BulkUpdates = {};

        // Process Tags
        const tagList = tags.split(",").map(t => t.trim()).filter(Boolean);
        if (tagList.length > 0) {
            updates.tags = tagList;
        }

        // Process Env Vars
        if (Object.keys(env).length > 0) {
            updates.env = env;
        }

        onApply(updates);
    };

    return (
        <div className="flex flex-col h-full">
            <div className="mb-4">
                <p className="text-sm text-muted-foreground">
                    You are editing <strong>{selectedCount}</strong> services. Changes will be merged into the existing configuration.
                </p>
            </div>

            <Tabs defaultValue="tags" className="w-full flex-1">
                <TabsList className="grid w-full grid-cols-2">
                    <TabsTrigger value="tags">Tags</TabsTrigger>
                    <TabsTrigger value="env">Environment</TabsTrigger>
                </TabsList>

                <TabsContent value="tags" className="py-4 space-y-4">
                    <div className="space-y-2">
                        <Label htmlFor="bulk-tags">Add Tags (comma separated)</Label>
                        <Input
                            id="bulk-tags"
                            placeholder="production, web, internal"
                            value={tags}
                            onChange={(e) => setTags(e.target.value)}
                        />
                        <p className="text-xs text-muted-foreground">
                            These tags will be added to the selected services. Existing tags will be preserved.
                        </p>
                    </div>
                </TabsContent>

                <TabsContent value="env" className="py-4 space-y-4">
                     <div className="space-y-2">
                         <div className="flex items-center justify-between">
                            <Label>Add/Update Environment Variables</Label>
                         </div>
                         <div className="border rounded-md p-4 bg-muted/20">
                            <EnvVarEditor
                                initialEnv={{}}
                                onChange={setEnv}
                            />
                         </div>
                         <p className="text-xs text-muted-foreground mt-2">
                             Variables defined here will be added or updated on the selected services.
                             <br />
                             <strong>Note:</strong> Only applies to services that support environment variables (e.g., Command Line).
                         </p>
                     </div>
                </TabsContent>
            </Tabs>

            <DialogFooter className="mt-4">
                <Button variant="outline" onClick={onCancel}>Cancel</Button>
                <Button onClick={handleApply}>Apply Changes</Button>
            </DialogFooter>
        </div>
    );
}
