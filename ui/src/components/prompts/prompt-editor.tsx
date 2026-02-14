/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect } from "react";
import { PromptDefinition } from "@/lib/client";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { Save, Trash2, Loader2, Code, MessageSquare } from "lucide-react";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import Editor from "@monaco-editor/react";
import { useTheme } from "next-themes";

interface PromptEditorProps {
    prompt: PromptDefinition | null;
    onSave: (prompt: PromptDefinition) => Promise<void>;
    onDelete: (name: string) => Promise<void>;
    isCreating: boolean;
}

export function PromptEditor({ prompt, onSave, onDelete, isCreating }: PromptEditorProps) {
    const [name, setName] = useState("");
    const [description, setDescription] = useState("");
    const [messagesJson, setMessagesJson] = useState("[]");
    const [schemaJson, setSchemaJson] = useState("{}");
    const [saving, setSaving] = useState(false);
    const { theme } = useTheme();

    useEffect(() => {
        if (prompt) {
            setName(prompt.name || "");
            setDescription(prompt.description || "");
            setMessagesJson(JSON.stringify(prompt.messages || [], null, 2));
            setSchemaJson(JSON.stringify(prompt.inputSchema || { type: "object", properties: {} }, null, 2));
        } else {
            // New Prompt Defaults
            setName("");
            setDescription("");
            setMessagesJson(JSON.stringify([
                {
                    role: "user",
                    content: {
                        text: "Hello {{name}}, how are you?"
                    }
                }
            ], null, 2));
            setSchemaJson(JSON.stringify({
                type: "object",
                properties: {
                    name: { type: "string", description: "The name of the user" }
                },
                required: ["name"]
            }, null, 2));
        }
    }, [prompt, isCreating]);

    const handleSave = async () => {
        setSaving(true);
        try {
            let messages = [];
            let inputSchema = {};
            try {
                messages = JSON.parse(messagesJson);
                inputSchema = JSON.parse(schemaJson);
            } catch (e) {
                alert("Invalid JSON in Messages or Schema");
                setSaving(false);
                return;
            }

            const newPrompt: PromptDefinition = {
                name,
                description,
                messages,
                inputSchema,
                disable: false,
                profiles: []
            };

            await onSave(newPrompt);
        } finally {
            setSaving(false);
        }
    };

    return (
        <div className="flex flex-col h-full bg-background">
            <div className="p-6 border-b pb-4 shrink-0">
                <div className="flex items-start justify-between">
                    <div>
                        <h2 className="text-2xl font-bold tracking-tight">
                            {isCreating ? "Create Prompt" : `Edit ${name}`}
                        </h2>
                        <p className="text-muted-foreground mt-1">
                            {isCreating ? "Define a new prompt template." : "Update prompt configuration."}
                        </p>
                    </div>
                    <div className="flex items-center gap-2">
                        {!isCreating && (
                            <Button variant="destructive" size="sm" onClick={() => onDelete(name)}>
                                <Trash2 className="mr-2 h-4 w-4" /> Delete
                            </Button>
                        )}
                        <Button onClick={handleSave} disabled={saving}>
                            {saving ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <Save className="mr-2 h-4 w-4" />}
                            Save
                        </Button>
                    </div>
                </div>
            </div>

            <div className="flex-1 overflow-hidden p-6 space-y-6 flex flex-col">
                <div className="grid grid-cols-2 gap-6 shrink-0">
                    <div className="space-y-2">
                        <Label htmlFor="name">Name</Label>
                        <Input
                            id="name"
                            value={name}
                            onChange={(e) => setName(e.target.value)}
                            placeholder="my_prompt"
                            disabled={!isCreating} // Cannot rename existing prompts easily without delete/create
                        />
                        <p className="text-[10px] text-muted-foreground">Unique identifier for the prompt.</p>
                    </div>
                    <div className="space-y-2">
                        <Label htmlFor="description">Description</Label>
                        <Input
                            id="description"
                            value={description}
                            onChange={(e) => setDescription(e.target.value)}
                            placeholder="What this prompt does..."
                        />
                    </div>
                </div>

                <Tabs defaultValue="messages" className="flex-1 flex flex-col min-h-0">
                    <TabsList>
                        <TabsTrigger value="messages" className="gap-2"><MessageSquare className="h-3 w-3" /> Messages (Template)</TabsTrigger>
                        <TabsTrigger value="schema" className="gap-2"><Code className="h-3 w-3" /> Arguments Schema</TabsTrigger>
                    </TabsList>
                    <TabsContent value="messages" className="flex-1 min-h-0 border rounded-md mt-2 relative overflow-hidden">
                        <Editor
                            height="100%"
                            defaultLanguage="json"
                            theme={theme === "dark" ? "vs-dark" : "light"}
                            value={messagesJson}
                            onChange={(v) => setMessagesJson(v || "")}
                            options={{ minimap: { enabled: false }, fontSize: 13 }}
                        />
                    </TabsContent>
                    <TabsContent value="schema" className="flex-1 min-h-0 border rounded-md mt-2 relative overflow-hidden">
                        <Editor
                            height="100%"
                            defaultLanguage="json"
                            theme={theme === "dark" ? "vs-dark" : "light"}
                            value={schemaJson}
                            onChange={(v) => setSchemaJson(v || "")}
                            options={{ minimap: { enabled: false }, fontSize: 13 }}
                        />
                    </TabsContent>
                </Tabs>
                <p className="text-xs text-muted-foreground shrink-0">
                    Use <code>{"{{variable}}"}</code> syntax in messages to insert arguments defined in the schema.
                </p>
            </div>
        </div>
    );
}
