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
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { EnvVarEditor } from "@/components/services/env-var-editor";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Info } from "lucide-react";

export interface BulkUpdate {
    tags?: string[];
    env?: Record<string, { plainText?: string; secretId?: string }>;
    timeout?: string;
    maxRetries?: number;
}

interface BulkServiceEditorProps {
    open: boolean;
    onOpenChange: (open: boolean) => void;
    selectedCount: number;
    onSave: (update: BulkUpdate) => void;
}

export function BulkServiceEditor({ open, onOpenChange, selectedCount, onSave }: BulkServiceEditorProps) {
    const [activeTab, setActiveTab] = useState("general");
    const [tags, setTags] = useState("");
    const [env, setEnv] = useState<Record<string, { plainText?: string; secretId?: string }>>({});
    const [timeout, setTimeout] = useState("");
    const [maxRetries, setMaxRetries] = useState("");

    const handleSave = () => {
        const update: BulkUpdate = {};
        if (tags.trim()) {
            update.tags = tags.split(",").map(t => t.trim()).filter(Boolean);
        }
        if (Object.keys(env).length > 0) {
            update.env = env;
        }
        if (timeout.trim()) {
            update.timeout = timeout.trim();
        }
        if (maxRetries.trim()) {
             const retries = parseInt(maxRetries);
             if (!isNaN(retries)) {
                 update.maxRetries = retries;
             }
        }
        onSave(update);
        onOpenChange(false);
        // Reset form
        setTags("");
        setEnv({});
        setTimeout("");
        setMaxRetries("");
    };

    return (
        <Dialog open={open} onOpenChange={onOpenChange}>
            <DialogContent className="sm:max-w-[600px] h-[80vh] flex flex-col p-0">
                <DialogHeader className="px-6 py-4 border-b">
                    <DialogTitle>Bulk Edit Services</DialogTitle>
                    <DialogDescription>
                        Apply changes to {selectedCount} selected services.
                    </DialogDescription>
                </DialogHeader>

                <Tabs value={activeTab} onValueChange={setActiveTab} className="flex-1 flex flex-col overflow-hidden">
                    <div className="px-6 border-b bg-muted/30">
                        <TabsList className="bg-transparent h-auto p-0">
                            <TabsTrigger value="general" className="rounded-none border-b-2 border-transparent data-[state=active]:border-primary data-[state=active]:bg-transparent px-4 py-3">General</TabsTrigger>
                            <TabsTrigger value="environment" className="rounded-none border-b-2 border-transparent data-[state=active]:border-primary data-[state=active]:bg-transparent px-4 py-3">Environment</TabsTrigger>
                            <TabsTrigger value="resilience" className="rounded-none border-b-2 border-transparent data-[state=active]:border-primary data-[state=active]:bg-transparent px-4 py-3">Resilience</TabsTrigger>
                        </TabsList>
                    </div>

                    <div className="flex-1 overflow-y-auto p-6 space-y-4">
                        <TabsContent value="general" className="mt-0 space-y-4">
                            <div className="space-y-2">
                                <Label htmlFor="bulk-tags">Add Tags (comma separated)</Label>
                                <Input
                                    id="bulk-tags"
                                    placeholder="production, web, internal"
                                    value={tags}
                                    onChange={(e) => setTags(e.target.value)}
                                />
                                <p className="text-xs text-muted-foreground">These tags will be added to the existing tags of selected services.</p>
                            </div>
                        </TabsContent>

                        <TabsContent value="environment" className="mt-0 space-y-4">
                             <Alert>
                                <Info className="h-4 w-4" />
                                <AlertTitle>Environment Variables</AlertTitle>
                                <AlertDescription>
                                    Variables added here will be merged into the configuration of supported services (Command Line, MCP Stdio).
                                </AlertDescription>
                            </Alert>
                            <EnvVarEditor
                                initialEnv={{}}
                                onChange={setEnv}
                            />
                        </TabsContent>

                        <TabsContent value="resilience" className="mt-0 space-y-4">
                             <div className="grid grid-cols-2 gap-4">
                                <div className="space-y-2">
                                    <Label htmlFor="timeout">Timeout</Label>
                                    <Input
                                        id="timeout"
                                        placeholder="e.g. 30s"
                                        value={timeout}
                                        onChange={(e) => setTimeout(e.target.value)}
                                    />
                                    <p className="text-xs text-muted-foreground">Overrides the request timeout.</p>
                                </div>
                                <div className="space-y-2">
                                    <Label htmlFor="retries">Max Retries</Label>
                                    <Input
                                        id="retries"
                                        type="number"
                                        placeholder="e.g. 3"
                                        value={maxRetries}
                                        onChange={(e) => setMaxRetries(e.target.value)}
                                    />
                                     <p className="text-xs text-muted-foreground">Overrides retry policy.</p>
                                </div>
                            </div>
                        </TabsContent>
                    </div>
                </Tabs>

                <DialogFooter className="px-6 py-4 border-t bg-muted/30">
                    <Button variant="outline" onClick={() => onOpenChange(false)}>Cancel</Button>
                    <Button onClick={handleSave}>Apply Changes</Button>
                </DialogFooter>
            </DialogContent>
        </Dialog>
    );
}
