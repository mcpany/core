/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useEffect, useState } from "react";
import { useForm, useFieldArray } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import * as z from "zod";
import { PromptDefinition, UpstreamServiceConfig } from "@/lib/client";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import {
    Sheet,
    SheetContent,
    SheetDescription,
    SheetHeader,
    SheetTitle,
    SheetFooter
} from "@/components/ui/sheet";
import {
    Form,
    FormControl,
    FormField,
    FormItem,
    FormLabel,
    FormMessage,
    FormDescription
} from "@/components/ui/form";
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from "@/components/ui/select";
import { Plus, Trash2, Code } from "lucide-react";
import Editor from "@monaco-editor/react";
import { useTheme } from "next-themes";

interface PromptEditorProps {
    open: boolean;
    onOpenChange: (open: boolean) => void;
    prompt: PromptDefinition | null;
    services: UpstreamServiceConfig[];
    onSave: (prompt: PromptDefinition, serviceId: string) => Promise<void>;
}

// Zod schema for validation
const promptSchema = z.object({
    name: z.string().min(1, "Name is required").regex(/^[a-zA-Z0-9_-]+$/, "Name must be alphanumeric with underscores or dashes"),
    description: z.string().optional(),
    serviceId: z.string().min(1, "Service is required"),
    inputSchema: z.string().optional(), // JSON string
    messages: z.array(z.object({
        role: z.enum(["user", "assistant"]),
        content: z.string().min(1, "Content is required")
    })).min(1, "At least one message is required")
});

type PromptValues = z.infer<typeof promptSchema>;

export function PromptEditor({ open, onOpenChange, prompt, services, onSave }: PromptEditorProps) {
    const { theme } = useTheme();
    const [isSubmitting, setIsSubmitting] = useState(false);

    // Identify which service the prompt belongs to if editing
    // prompt object doesn't strictly have serviceId in proto, but we might have injected it or need to find it.
    // In PromptWorkbench, we might need to pass the serviceId if known.
    // For now, let's assume we can infer it or default to the first one.
    const defaultServiceId = (prompt as any)?.serviceId || (services.length > 0 ? services[0].id : "");

    const form = useForm<PromptValues>({
        resolver: zodResolver(promptSchema),
        defaultValues: {
            name: "",
            description: "",
            serviceId: defaultServiceId,
            inputSchema: "{\n  \"type\": \"object\",\n  \"properties\": {\n    \"arg1\": { \"type\": \"string\" }\n  }\n}",
            messages: [{ role: "user", content: "" }]
        },
    });

    const { fields, append, remove } = useFieldArray({
        control: form.control,
        name: "messages"
    });

    useEffect(() => {
        if (open) {
            if (prompt) {
                form.reset({
                    name: prompt.name,
                    description: prompt.description || "",
                    serviceId: (prompt as any).serviceId || defaultServiceId,
                    inputSchema: JSON.stringify(prompt.inputSchema || {}, null, 2),
                    messages: prompt.messages?.map((m: any) => ({
                        role: m.role === "USER" || m.role === 0 ? "user" : "assistant", // Map enum to string
                        content: m.content?.text?.text || (typeof m.content === 'string' ? m.content : "") // Handle content structure
                    })) || [{ role: "user", content: "" }]
                });
            } else {
                form.reset({
                    name: "",
                    description: "",
                    serviceId: defaultServiceId,
                    inputSchema: "{\n  \"type\": \"object\",\n  \"properties\": {\n    \n  }\n}",
                    messages: [{ role: "user", content: "" }]
                });
            }
        }
    }, [open, prompt, services, form, defaultServiceId]);

    const onSubmit = async (data: PromptValues) => {
        setIsSubmitting(true);
        try {
            let parsedSchema = {};
            try {
                parsedSchema = JSON.parse(data.inputSchema || "{}");
            } catch (e) {
                form.setError("inputSchema", { message: "Invalid JSON" });
                setIsSubmitting(false);
                return;
            }

            const newPrompt: PromptDefinition = {
                name: data.name,
                description: data.description,
                inputSchema: parsedSchema,
                messages: data.messages.map(m => ({
                    role: m.role === "user" ? 0 : 1, // USER=0, ASSISTANT=1
                    content: {
                        text: { text: m.content }
                    }
                })),
                disable: false,
                profiles: [], // Default
                title: data.name // Default title to name
            } as any; // Type casting due to strict proto types vs simplified client types

            await onSave(newPrompt, data.serviceId);
            onOpenChange(false);
        } catch (e) {
            console.error(e);
        } finally {
            setIsSubmitting(false);
        }
    };

    return (
        <Sheet open={open} onOpenChange={onOpenChange}>
            <SheetContent className="sm:max-w-[600px] overflow-y-auto w-[600px]">
                <SheetHeader>
                    <SheetTitle>{prompt ? "Edit Prompt" : "New Prompt"}</SheetTitle>
                    <SheetDescription>
                        Configure the prompt template and arguments.
                    </SheetDescription>
                </SheetHeader>

                <Form {...form}>
                    <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-6 py-6">
                        <div className="grid grid-cols-2 gap-4">
                            <FormField
                                control={form.control}
                                name="name"
                                render={({ field }) => (
                                    <FormItem>
                                        <FormLabel>Name</FormLabel>
                                        <FormControl>
                                            <Input placeholder="summarize_text" {...field} disabled={!!prompt} />
                                        </FormControl>
                                        <FormMessage />
                                    </FormItem>
                                )}
                            />
                            <FormField
                                control={form.control}
                                name="serviceId"
                                render={({ field }) => (
                                    <FormItem>
                                        <FormLabel>Service</FormLabel>
                                        <Select onValueChange={field.onChange} defaultValue={field.value} value={field.value} disabled={!!prompt}>
                                            <FormControl>
                                                <SelectTrigger>
                                                    <SelectValue placeholder="Select Service" />
                                                </SelectTrigger>
                                            </FormControl>
                                            <SelectContent>
                                                {services.map(svc => (
                                                    <SelectItem key={svc.id} value={svc.id}>
                                                        {svc.name}
                                                    </SelectItem>
                                                ))}
                                            </SelectContent>
                                        </Select>
                                        <FormMessage />
                                    </FormItem>
                                )}
                            />
                        </div>

                        <FormField
                            control={form.control}
                            name="description"
                            render={({ field }) => (
                                <FormItem>
                                    <FormLabel>Description</FormLabel>
                                    <FormControl>
                                        <Input placeholder="Short description of what this prompt does" {...field} />
                                    </FormControl>
                                    <FormMessage />
                                </FormItem>
                            )}
                        />

                        <FormField
                            control={form.control}
                            name="inputSchema"
                            render={({ field }) => (
                                <FormItem>
                                    <FormLabel className="flex items-center gap-2">
                                        Arguments Schema (JSON Schema)
                                        <Code className="h-3 w-3 text-muted-foreground" />
                                    </FormLabel>
                                    <div className="h-[200px] border rounded-md overflow-hidden">
                                        <Editor
                                            height="100%"
                                            defaultLanguage="json"
                                            theme={theme === "dark" ? "vs-dark" : "light"}
                                            value={field.value}
                                            onChange={(val) => field.onChange(val || "")}
                                            options={{ minimap: { enabled: false }, fontSize: 12 }}
                                        />
                                    </div>
                                    <FormMessage />
                                </FormItem>
                            )}
                        />

                        <div className="space-y-4">
                            <div className="flex items-center justify-between">
                                <FormLabel>Messages Template</FormLabel>
                                <Button type="button" variant="outline" size="sm" onClick={() => append({ role: "user", content: "" })}>
                                    <Plus className="mr-2 h-3 w-3" /> Add Message
                                </Button>
                            </div>

                            <div className="space-y-4">
                                {fields.map((field, index) => (
                                    <div key={field.id} className="flex gap-2 items-start p-3 bg-muted/20 rounded-md border">
                                        <div className="w-[100px] shrink-0">
                                            <FormField
                                                control={form.control}
                                                name={`messages.${index}.role`}
                                                render={({ field: roleField }) => (
                                                    <FormItem>
                                                        <Select onValueChange={roleField.onChange} defaultValue={roleField.value} value={roleField.value}>
                                                            <FormControl>
                                                                <SelectTrigger className="h-8 text-xs">
                                                                    <SelectValue />
                                                                </SelectTrigger>
                                                            </FormControl>
                                                            <SelectContent>
                                                                <SelectItem value="user">User</SelectItem>
                                                                <SelectItem value="assistant">Assistant</SelectItem>
                                                            </SelectContent>
                                                        </Select>
                                                    </FormItem>
                                                )}
                                            />
                                        </div>
                                        <div className="flex-1">
                                            <FormField
                                                control={form.control}
                                                name={`messages.${index}.content`}
                                                render={({ field: contentField }) => (
                                                    <FormItem>
                                                        <FormControl>
                                                            <Textarea
                                                                placeholder="Enter prompt text (use {{arg}} for variables)"
                                                                className="min-h-[80px] text-sm font-mono"
                                                                {...contentField}
                                                            />
                                                        </FormControl>
                                                        <FormMessage />
                                                    </FormItem>
                                                )}
                                            />
                                        </div>
                                        <Button
                                            type="button"
                                            variant="ghost"
                                            size="icon"
                                            className="h-8 w-8 text-muted-foreground hover:text-destructive"
                                            onClick={() => remove(index)}
                                            disabled={fields.length === 1}
                                        >
                                            <Trash2 className="h-4 w-4" />
                                        </Button>
                                    </div>
                                ))}
                            </div>
                            {form.formState.errors.messages && (
                                <p className="text-sm font-medium text-destructive">{form.formState.errors.messages.message}</p>
                            )}
                        </div>

                        <SheetFooter className="pt-4">
                            <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>Cancel</Button>
                            <Button type="submit" disabled={isSubmitting}>
                                {isSubmitting ? "Saving..." : "Save Prompt"}
                            </Button>
                        </SheetFooter>
                    </form>
                </Form>
            </SheetContent>
        </Sheet>
    );
}
