import React, { useState, useEffect } from "react";
import { apiClient, PromptDefinition } from "@/lib/client";
import { PromptList } from "./prompt-list";
import { PromptEditor } from "./prompt-editor";
import { PromptRunner } from "./prompt-runner";
import { useToast } from "@/hooks/use-toast";
import { MessageSquare } from "lucide-react";
import { Button } from "@/components/ui/button";

export function PromptWorkbench() {
    const [prompts, setPrompts] = useState<PromptDefinition[]>([]);
    const [selectedPrompt, setSelectedPrompt] = useState<PromptDefinition | null>(null);
    const [isEditing, setIsEditing] = useState(false);
    const [isCreating, setIsCreating] = useState(false);
    const { toast } = useToast();

    useEffect(() => {
        loadPrompts();
    }, []);

    const loadPrompts = async () => {
        try {
            const data = await apiClient.listPrompts();
            setPrompts(data?.prompts || []);
        } catch (e) {
            console.error("Failed to list prompts", e);
            toast({
                title: "Error",
                description: "Failed to load prompts.",
                variant: "destructive"
            });
        }
    };

    const handleCreate = () => {
        setSelectedPrompt(null);
        setIsEditing(true);
        setIsCreating(true);
    };

    const handleSelect = (prompt: PromptDefinition) => {
        setSelectedPrompt(prompt);
        setIsEditing(false);
        setIsCreating(false);
    };

    const handleSave = async (prompt: PromptDefinition) => {
        try {
            if (isCreating) {
                await apiClient.createPrompt(prompt);
                toast({ title: "Prompt Created", description: `Prompt ${prompt.name} created successfully.` });
            } else {
                if (!selectedPrompt) return;
                await apiClient.updatePrompt(selectedPrompt.name, prompt);
                toast({ title: "Prompt Updated", description: `Prompt ${prompt.name} updated successfully.` });
            }
            await loadPrompts();

            // Re-select logic
            // For simplicity, we just reload list. If we created, we might need to find it to select it.
            // But for now, let's just go back to edit/runner mode for that prompt?
            // Actually, after save, usually we go to view mode.

            setIsEditing(false);
            setIsCreating(false);

            // Optimistic select
            setSelectedPrompt(prompt);

        } catch (e) {
            console.error(e);
            toast({ title: "Error", description: (e as Error).message, variant: "destructive" });
        }
    };

    const handleDelete = async (name: string) => {
        if (!confirm(`Are you sure you want to delete prompt "${name}"?`)) return;
        try {
            await apiClient.deletePrompt(name);
            toast({ title: "Prompt Deleted", description: `Prompt ${name} deleted successfully.` });
            await loadPrompts();
            setSelectedPrompt(null);
            setIsEditing(false);
        } catch (e) {
            console.error(e);
            toast({ title: "Error", description: (e as Error).message, variant: "destructive" });
        }
    };

    return (
        <div className="flex h-full w-full rounded-md border bg-background shadow-sm overflow-hidden">
            <PromptList
                prompts={prompts}
                selectedPrompt={selectedPrompt}
                onSelect={handleSelect}
                onCreate={handleCreate}
            />
            <div className="flex-1 min-w-0 bg-background h-full border-l">
                {isEditing || isCreating ? (
                    <PromptEditor
                        prompt={isCreating ? null : selectedPrompt}
                        onSave={handleSave}
                        onDelete={handleDelete}
                        onCancel={() => {
                            setIsEditing(false);
                            setIsCreating(false);
                            if (isCreating) setSelectedPrompt(null);
                        }}
                    />
                ) : selectedPrompt ? (
                    <PromptRunner
                        prompt={selectedPrompt}
                        onEdit={() => setIsEditing(true)}
                    />
                ) : (
                    <div className="flex flex-col items-center justify-center h-full text-muted-foreground p-8">
                        <MessageSquare className="h-12 w-12 opacity-20 mb-4" />
                        <h3 className="text-lg font-medium">No Prompt Selected</h3>
                        <p className="max-w-xs text-center mt-2">
                            Select a prompt from the library to run it, or create a new one.
                        </p>
                        <Button className="mt-4" onClick={handleCreate}>Create Prompt</Button>
                    </div>
                )}
            </div>
        </div>
    );
}
