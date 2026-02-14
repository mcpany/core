/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useState, useEffect } from "react";
import { PromptDefinition } from "@/lib/client";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Save, Trash2, Plus, X, Loader2, Info } from "lucide-react";
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from "@/components/ui/select";
import { useToast } from "@/hooks/use-toast";

interface PromptEditorProps {
    prompt: PromptDefinition | null; // Null means new prompt
    onSave: (prompt: PromptDefinition) => Promise<void>;
    onDelete: (name: string) => Promise<void>;
    onCancel: () => void;
}

interface Argument {
    name: string;
    description: string;
    required: boolean;
}

interface Message {
    role: "user" | "assistant";
    text: string;
}

export function PromptEditor({ prompt, onSave, onDelete, onCancel }: PromptEditorProps) {
    const [name, setName] = useState("");
    const [description, setDescription] = useState("");
    const [args, setArgs] = useState<Argument[]>([]);
    const [messages, setMessages] = useState<Message[]>([]);
    const [saving, setSaving] = useState(false);
    const [deleting, setDeleting] = useState(false);
    const { toast } = useToast();

    useEffect(() => {
        if (prompt) {
            setName(prompt.name);
            setDescription(prompt.description || "");

            // Parse arguments from Schema
            const schema = prompt.inputSchema as any;
            const newArgs: Argument[] = [];
            if (schema && schema.properties) {
                const required = (schema.required as string[]) || [];
                Object.entries(schema.properties).forEach(([key, val]: [string, any]) => {
                    newArgs.push({
                        name: key,
                        description: val.description || "",
                        required: required.includes(key)
                    });
                });
            }
            setArgs(newArgs);

            // Parse messages
            const newMessages: Message[] = [];
            if (prompt.messages) {
                prompt.messages.forEach((msg: any) => {
                    // msg.role is enum number usually? Or string if JSON.
                    // Proto JSON mapping uses strings "USER", "ASSISTANT" or 0, 1?
                    // client.ts types says `messages: any[]` effectively.
                    // Let's assume standard MCP format or what we see in `TemplatedPrompt`.
                    // Actually configv1.PromptMessage has Role enum.
                    // JSON would be "USER" or "ASSISTANT".
                    let role: "user" | "assistant" = "user";
                    if (typeof msg.role === 'string') {
                        if (msg.role.toLowerCase() === "assistant") role = "assistant";
                    } else if (msg.role === 1) {
                        role = "assistant";
                    }

                    // Content
                    let text = "";
                    if (msg.text) text = msg.text.text; // TextContent inside
                    // Proto: content is oneof. JSON: { text: { text: "..." } } or just { text: "..." } if simplified?
                    // Let's handle { text: { text: "..." } } structure from protojson.
                    else if (msg.content?.text) text = msg.content.text.text;

                    newMessages.push({ role, text });
                });
            }
            if (newMessages.length === 0) {
                newMessages.push({ role: "user", text: "" });
            }
            setMessages(newMessages);
        } else {
            // New Prompt Defaults
            setName("");
            setDescription("");
            setArgs([]);
            setMessages([{ role: "user", text: "" }]);
        }
    }, [prompt]);

    const handleSave = async () => {
        if (!name.trim()) {
            toast({ title: "Validation Error", description: "Name is required.", variant: "destructive" });
            return;
        }

        setSaving(true);
        try {
            // Construct PromptDefinition
            const properties: Record<string, any> = {};
            const required: string[] = [];
            args.forEach(arg => {
                if (arg.name) {
                    properties[arg.name] = { type: "string", description: arg.description };
                    if (arg.required) required.push(arg.name);
                }
            });

            const inputSchema = {
                type: "object",
                properties,
                required
            };

            const protoMessages = messages.map(m => ({
                role: m.role.toUpperCase(), // "USER" or "ASSISTANT"
                text: { text: m.text } // TextContent
            }));

            const newPrompt: PromptDefinition = {
                name,
                description,
                inputSchema,
                messages: protoMessages,
                disable: false,
                profiles: []
            };

            await onSave(newPrompt);
        } catch (e) {
            console.error(e);
            toast({ title: "Error", description: "Failed to save prompt.", variant: "destructive" });
        } finally {
            setSaving(false);
        }
    };

    const handleDelete = async () => {
        if (!confirm("Are you sure you want to delete this prompt?")) return;
        setDeleting(true);
        try {
            await onDelete(name);
        } catch (e) {
            console.error(e);
            toast({ title: "Error", description: "Failed to delete prompt.", variant: "destructive" });
        } finally {
            setDeleting(false);
        }
    };

    return (
        <div className="flex flex-col h-full bg-background p-6 space-y-6 overflow-y-auto">
            <div className="flex items-center justify-between">
                <h2 className="text-2xl font-bold tracking-tight">{prompt ? "Edit Prompt" : "Create Prompt"}</h2>
                <div className="flex items-center gap-2">
                    {prompt && (
                        <Button variant="destructive" size="sm" onClick={handleDelete} disabled={deleting || saving}>
                            {deleting ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <Trash2 className="mr-2 h-4 w-4" />}
                            Delete
                        </Button>
                    )}
                    <Button variant="outline" size="sm" onClick={onCancel}>Cancel</Button>
                    <Button size="sm" onClick={handleSave} disabled={saving || deleting}>
                        {saving ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <Save className="mr-2 h-4 w-4" />}
                        Save
                    </Button>
                </div>
            </div>

            <div className="grid gap-6">
                {/* General Info */}
                <Card>
                    <CardHeader>
                        <CardTitle className="text-sm font-medium">General</CardTitle>
                    </CardHeader>
                    <CardContent className="space-y-4">
                        <div className="grid gap-2">
                            <Label htmlFor="name">Name</Label>
                            <Input
                                id="name"
                                value={name}
                                onChange={(e) => setName(e.target.value)}
                                placeholder="my_prompt"
                                disabled={!!prompt} // Disable renaming for simplicity
                            />
                            {prompt && <p className="text-[10px] text-muted-foreground">Prompt names cannot be changed once created.</p>}
                        </div>
                        <div className="grid gap-2">
                            <Label htmlFor="description">Description</Label>
                            <Input
                                id="description"
                                value={description}
                                onChange={(e) => setDescription(e.target.value)}
                                placeholder="What does this prompt do?"
                            />
                        </div>
                    </CardContent>
                </Card>

                {/* Arguments */}
                <Card>
                    <CardHeader className="flex flex-row items-center justify-between pb-2">
                        <CardTitle className="text-sm font-medium">Arguments</CardTitle>
                        <Button variant="ghost" size="sm" onClick={() => setArgs([...args, { name: "", description: "", required: true }])}>
                            <Plus className="h-4 w-4" /> Add
                        </Button>
                    </CardHeader>
                    <CardContent className="space-y-2">
                        {args.map((arg, idx) => (
                            <div key={idx} className="flex items-start gap-2">
                                <Input
                                    placeholder="Name"
                                    value={arg.name}
                                    onChange={(e) => {
                                        const newArgs = [...args];
                                        newArgs[idx].name = e.target.value;
                                        setArgs(newArgs);
                                    }}
                                    className="w-1/3"
                                />
                                <Input
                                    placeholder="Description"
                                    value={arg.description}
                                    onChange={(e) => {
                                        const newArgs = [...args];
                                        newArgs[idx].description = e.target.value;
                                        setArgs(newArgs);
                                    }}
                                    className="flex-1"
                                />
                                <div className="flex items-center gap-1 pt-2">
                                    <input
                                        type="checkbox"
                                        checked={arg.required}
                                        onChange={(e) => {
                                            const newArgs = [...args];
                                            newArgs[idx].required = e.target.checked;
                                            setArgs(newArgs);
                                        }}
                                        className="h-4 w-4"
                                        title="Required"
                                    />
                                </div>
                                <Button variant="ghost" size="icon" onClick={() => setArgs(args.filter((_, i) => i !== idx))}>
                                    <X className="h-4 w-4 text-muted-foreground hover:text-destructive" />
                                </Button>
                            </div>
                        ))}
                        {args.length === 0 && (
                            <div className="text-sm text-muted-foreground italic text-center py-4">
                                No arguments defined.
                            </div>
                        )}
                        <div className="flex items-center gap-2 text-xs text-muted-foreground bg-muted/30 p-2 rounded">
                            <Info className="h-3 w-3" />
                            Use arguments in messages with <code className="bg-muted px-1 rounded">{"{{arg_name}}"}</code> syntax.
                        </div>
                    </CardContent>
                </Card>

                {/* Messages */}
                <Card className="flex-1">
                    <CardHeader className="flex flex-row items-center justify-between pb-2">
                        <CardTitle className="text-sm font-medium">Template Messages</CardTitle>
                        <Button variant="ghost" size="sm" onClick={() => setMessages([...messages, { role: "user", text: "" }])}>
                            <Plus className="h-4 w-4" /> Add Message
                        </Button>
                    </CardHeader>
                    <CardContent className="space-y-4">
                        {messages.map((msg, idx) => (
                            <div key={idx} className="flex flex-col gap-2 p-3 border rounded-md relative group">
                                <div className="flex items-center justify-between">
                                    <Select
                                        value={msg.role}
                                        onValueChange={(val: "user" | "assistant") => {
                                            const newMessages = [...messages];
                                            newMessages[idx].role = val;
                                            setMessages(newMessages);
                                        }}
                                    >
                                        <SelectTrigger className="w-[120px] h-8 text-xs uppercase font-bold">
                                            <SelectValue />
                                        </SelectTrigger>
                                        <SelectContent>
                                            <SelectItem value="user">USER</SelectItem>
                                            <SelectItem value="assistant">ASSISTANT</SelectItem>
                                        </SelectContent>
                                    </Select>
                                    <Button
                                        variant="ghost"
                                        size="icon"
                                        className="h-6 w-6 opacity-0 group-hover:opacity-100 transition-opacity"
                                        onClick={() => setMessages(messages.filter((_, i) => i !== idx))}
                                    >
                                        <X className="h-4 w-4 text-muted-foreground hover:text-destructive" />
                                    </Button>
                                </div>
                                <Textarea
                                    value={msg.text}
                                    onChange={(e) => {
                                        const newMessages = [...messages];
                                        newMessages[idx].text = e.target.value;
                                        setMessages(newMessages);
                                    }}
                                    placeholder="Enter message content..."
                                    className="font-mono text-sm min-h-[100px]"
                                />
                            </div>
                        ))}
                    </CardContent>
                </Card>
            </div>
        </div>
    );
}
