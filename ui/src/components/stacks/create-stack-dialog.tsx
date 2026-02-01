/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import * as z from "zod";
import { Button } from "@/components/ui/button";
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
    DialogTrigger,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import {
    Form,
    FormControl,
    FormField,
    FormItem,
    FormLabel,
    FormMessage,
} from "@/components/ui/form";
import { Plus, Loader2 } from "lucide-react";
import { apiClient } from "@/lib/client";
import { useToast } from "@/hooks/use-toast";

const stackSchema = z.object({
    name: z.string().min(3, "Name must be at least 3 characters").regex(/^[a-z0-9-]+$/, "Name must be lowercase alphanumeric with dashes"),
    description: z.string().optional(),
    author: z.string().optional(),
});

type StackValues = z.infer<typeof stackSchema>;

interface CreateStackDialogProps {
    onStackCreated?: () => void;
}

/**
 * CreateStackDialog component.
 * @param props - The component props.
 * @returns The rendered component.
 */
export function CreateStackDialog({ onStackCreated }: CreateStackDialogProps) {
    const [open, setOpen] = useState(false);
    const [loading, setLoading] = useState(false);
    const { toast } = useToast();

    const form = useForm<StackValues>({
        resolver: zodResolver(stackSchema),
        defaultValues: {
            name: "",
            description: "",
            author: "User",
        },
    });

    const onSubmit = async (data: StackValues) => {
        setLoading(true);
        try {
            await apiClient.createCollection({
                name: data.name,
                description: data.description || "",
                author: data.author || "User",
                version: "0.0.1",
                services: [], // Empty stack
            });
            toast({
                title: "Stack Created",
                description: `Stack "${data.name}" has been created successfully.`,
            });
            setOpen(false);
            form.reset();
            if (onStackCreated) onStackCreated();
        } catch (error) {
            console.error(error);
            toast({
                variant: "destructive",
                title: "Error",
                description: "Failed to create stack. It might already exist.",
            });
        } finally {
            setLoading(false);
        }
    };

    return (
        <Dialog open={open} onOpenChange={setOpen}>
            <DialogTrigger asChild>
                <Button>
                    <Plus className="mr-2 h-4 w-4" /> Create Stack
                </Button>
            </DialogTrigger>
            <DialogContent className="sm:max-w-[425px]">
                <DialogHeader>
                    <DialogTitle>Create New Stack</DialogTitle>
                    <DialogDescription>
                        Create an empty stack to organize your services.
                    </DialogDescription>
                </DialogHeader>
                <Form {...form}>
                    <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
                        <FormField
                            control={form.control}
                            name="name"
                            render={({ field }) => (
                                <FormItem>
                                    <FormLabel>Name</FormLabel>
                                    <FormControl>
                                        <Input placeholder="my-stack-name" {...field} />
                                    </FormControl>
                                    <FormMessage />
                                </FormItem>
                            )}
                        />
                        <FormField
                            control={form.control}
                            name="description"
                            render={({ field }) => (
                                <FormItem>
                                    <FormLabel>Description</FormLabel>
                                    <FormControl>
                                        <Input placeholder="Purpose of this stack..." {...field} />
                                    </FormControl>
                                    <FormMessage />
                                </FormItem>
                            )}
                        />
                        <FormField
                            control={form.control}
                            name="author"
                            render={({ field }) => (
                                <FormItem>
                                    <FormLabel>Author</FormLabel>
                                    <FormControl>
                                        <Input placeholder="Your Name" {...field} />
                                    </FormControl>
                                    <FormMessage />
                                </FormItem>
                            )}
                        />
                        <DialogFooter>
                            <Button type="submit" disabled={loading}>
                                {loading && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                                Create
                            </Button>
                        </DialogFooter>
                    </form>
                </Form>
            </DialogContent>
        </Dialog>
    );
}
