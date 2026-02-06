/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import * as z from "zod";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
    DialogTrigger,
} from "@/components/ui/dialog";
import {
    Form,
    FormControl,
    FormField,
    FormItem,
    FormLabel,
    FormMessage,
} from "@/components/ui/form";
import { Plus } from "lucide-react";
import { apiClient } from "@/lib/client";
import { toast } from "@/hooks/use-toast";

const formSchema = z.object({
    name: z.string()
        .min(2, "Name must be at least 2 characters")
        .regex(/^[a-zA-Z0-9_-]+$/, "Name can only contain letters, numbers, underscores, and dashes"),
});

interface CreateStackDialogProps {
    onStackCreated: () => void;
}

/**
 * CreateStackDialog displays a dialog to create a new stack.
 * @param props The component props.
 * @param props.onStackCreated Callback when a stack is successfully created.
 * @returns The rendered component.
 */
export function CreateStackDialog({ onStackCreated }: CreateStackDialogProps) {
    const [open, setOpen] = useState(false);
    const form = useForm<z.infer<typeof formSchema>>({
        resolver: zodResolver(formSchema),
        defaultValues: {
            name: "",
        },
    });

    const onSubmit = async (values: z.infer<typeof formSchema>) => {
        try {
            await apiClient.saveCollection({
                name: values.name,
                services: []
            });
            toast({
                title: "Stack created",
                description: `Stack "${values.name}" has been created successfully.`,
            });
            setOpen(false);
            form.reset();
            onStackCreated();
        } catch (error) {
            console.error(error);
            toast({
                variant: "destructive",
                title: "Error",
                description: "Failed to create stack. It might already exist.",
            });
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
                        Create a new empty stack configuration.
                    </DialogDescription>
                </DialogHeader>
                <Form {...form}>
                    <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
                        <FormField
                            control={form.control}
                            name="name"
                            render={({ field }) => (
                                <FormItem>
                                    <FormLabel>Stack Name</FormLabel>
                                    <FormControl>
                                        <Input placeholder="my-stack" {...field} />
                                    </FormControl>
                                    <FormMessage />
                                </FormItem>
                            )}
                        />
                        <DialogFooter>
                            <Button type="submit" disabled={form.formState.isSubmitting}>
                                {form.formState.isSubmitting ? "Creating..." : "Create Stack"}
                            </Button>
                        </DialogFooter>
                    </form>
                </Form>
            </DialogContent>
        </Dialog>
    );
}
