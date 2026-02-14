/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect } from "react";
import { apiClient, PromptDefinition } from "@/lib/client";
import { PromptList } from "./prompt-list";
import { PromptEditor } from "./prompt-editor";
import { PromptRunner } from "./prompt-runner";
import { useToast } from "@/hooks/use-toast";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { MessageSquare } from "lucide-react";

export function PromptWorkbench() {
    const [prompts, setPrompts] = useState<PromptDefinition[]>([]);
    const [selectedPrompt, setSelectedPrompt] = useState<PromptDefinition | null>(null);
    const [viewMode, setViewMode] = useState<"edit" | "run">("run");
    const [isCreating, setIsCreating] = useState(false);
    const [searchQuery, setSearchQuery] = useState("");
    const { toast } = useToast();

    useEffect(() => {
        loadPrompts();
    }, []);

    const loadPrompts = async () => {
        try {
            const res = await apiClient.listPrompts();
            setPrompts(res?.prompts || []);
        } catch (e) {
            console.error("Failed to load prompts", e);
            toast({
                variant: "destructive",
                title: "Error",
                description: "Failed to load prompts."
            });
        }
    };

    const handleSelect = (prompt: PromptDefinition) => {
        setSelectedPrompt(prompt);
        setIsCreating(false);
        // If it's a system prompt, default to run mode. User prompts can default to edit?
        // Let's stick to current viewMode or default to run
        if (viewMode === "edit" && !prompt.name.startsWith("user-library.")) {
             setViewMode("run");
        }
    };

    const handleNew = () => {
        setSelectedPrompt(null);
        setIsCreating(true);
        setViewMode("edit");
    };

    const handleSave = async (prompt: PromptDefinition) => {
        try {
            if (isCreating) {
                await apiClient.createPrompt(prompt);
                toast({ title: "Success", description: "Prompt created." });
            } else {
                await apiClient.updatePrompt(prompt);
                toast({ title: "Success", description: "Prompt updated." });
            }
            loadPrompts();
            setIsCreating(false);
            setSelectedPrompt(prompt);
        } catch (e: any) {
            console.error("Save failed", e);
            toast({
                variant: "destructive",
                title: "Error",
                description: e.message || "Failed to save prompt."
            });
        }
    };

    const handleDelete = async (name: string) => {
        if (!confirm(`Are you sure you want to delete prompt "${name}"?`)) return;
        try {
            await apiClient.deletePrompt(name);
            toast({ title: "Deleted", description: "Prompt deleted." });
            loadPrompts();
            if (selectedPrompt?.name === name) {
                setSelectedPrompt(null);
                setIsCreating(false);
            }
        } catch (e: any) {
            console.error("Delete failed", e);
            toast({
                variant: "destructive",
                title: "Error",
                description: e.message || "Failed to delete prompt."
            });
        }
    };

    const isEditable = isCreating || (selectedPrompt?.name?.startsWith("user-library.") ?? false);

    return (
        <div className="flex h-[calc(100vh-120px)] w-full rounded-md border bg-background shadow-sm overflow-hidden">
            <PromptList
                prompts={prompts}
                selectedPrompt={selectedPrompt}
                onSelect={handleSelect}
                onNew={handleNew}
                searchQuery={searchQuery}
                onSearchChange={setSearchQuery}
            />

            <div className="flex-1 flex flex-col min-w-0 bg-background">
                {selectedPrompt || isCreating ? (
                    <Tabs value={viewMode} onValueChange={(v: any) => setViewMode(v)} className="flex-1 flex flex-col">
                        <div className="border-b px-4 bg-muted/5">
                            <TabsList className="h-12 bg-transparent p-0">
                                <TabsTrigger
                                    value="run"
                                    className="data-[state=active]:bg-background data-[state=active]:border-b-2 data-[state=active]:border-primary rounded-none h-full px-6"
                                    disabled={isCreating}
                                >
                                    Run / Preview
                                </TabsTrigger>
                                <TabsTrigger
                                    value="edit"
                                    className="data-[state=active]:bg-background data-[state=active]:border-b-2 data-[state=active]:border-primary rounded-none h-full px-6"
                                    disabled={!isEditable}
                                >
                                    Edit Definition
                                </TabsTrigger>
                            </TabsList>
                        </div>

                        <div className="flex-1 overflow-hidden relative">
                            <TabsContent value="run" className="h-full m-0 data-[state=inactive]:hidden">
                                {selectedPrompt && <PromptRunner prompt={selectedPrompt} />}
                            </TabsContent>
                            <TabsContent value="edit" className="h-full m-0 data-[state=inactive]:hidden">
                                <PromptEditor
                                    prompt={selectedPrompt}
                                    onSave={handleSave}
                                    onDelete={handleDelete}
                                    isCreating={isCreating}
                                />
                            </TabsContent>
                        </div>
                    </Tabs>
                ) : (
                    <div className="flex flex-col items-center justify-center h-full text-muted-foreground p-8">
                        <MessageSquare className="h-12 w-12 opacity-20 mb-4" />
                        <h3 className="text-lg font-medium">No Prompt Selected</h3>
                        <p className="max-w-xs text-center mt-2">
                            Select a prompt from the library to view details or create a new one.
                        </p>
                    </div>
                )}
            </div>
        </div>
    );
}
