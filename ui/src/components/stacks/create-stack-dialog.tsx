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
import { Textarea } from "@/components/ui/textarea";
import { apiClient } from "@/lib/client";
import { useToast } from "@/hooks/use-toast";
import { Loader2 } from "lucide-react";

interface CreateStackDialogProps {
    open: boolean;
    onOpenChange: (open: boolean) => void;
    onStackCreated: () => void;
}

export function CreateStackDialog({ open, onOpenChange, onStackCreated }: CreateStackDialogProps) {
    const [name, setName] = useState("");
    const [description, setDescription] = useState("");
    const [isSubmitting, setIsSubmitting] = useState(false);
    const { toast } = useToast();

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();

        if (!name.trim()) {
            toast({
                variant: "destructive",
                title: "Validation Error",
                description: "Stack name is required.",
            });
            return;
        }

        setIsSubmitting(true);

        try {
            // Create a new empty collection
            await apiClient.saveCollection({
                name: name.trim(),
                description: description.trim(),
                services: [],
                version: "0.0.1",
                priority: 0
            });

            toast({
                title: "Stack Created",
                description: `Stack "${name}" has been created successfully.`,
            });

            // Reset and close
            setName("");
            setDescription("");
            onStackCreated();
            onOpenChange(false);
        } catch (error) {
            console.error("Failed to create stack", error);
            toast({
                variant: "destructive",
                title: "Creation Failed",
                description: error instanceof Error ? error.message : "An unknown error occurred",
            });
        } finally {
            setIsSubmitting(false);
        }
    };

    return (
        <Dialog open={open} onOpenChange={onOpenChange}>
            <DialogContent className="sm:max-w-[425px]">
                <form onSubmit={handleSubmit}>
                    <DialogHeader>
                        <DialogTitle>Create New Stack</DialogTitle>
                        <DialogDescription>
                            Create a new empty stack to organize your services.
                        </DialogDescription>
                    </DialogHeader>
                    <div className="grid gap-4 py-4">
                        <div className="grid grid-cols-4 items-center gap-4">
                            <Label htmlFor="name" className="text-right">
                                Name
                            </Label>
                            <Input
                                id="name"
                                value={name}
                                onChange={(e) => setName(e.target.value)}
                                className="col-span-3"
                                placeholder="my-stack"
                                disabled={isSubmitting}
                            />
                        </div>
                        <div className="grid grid-cols-4 items-center gap-4">
                            <Label htmlFor="description" className="text-right">
                                Description
                            </Label>
                            <Textarea
                                id="description"
                                value={description}
                                onChange={(e) => setDescription(e.target.value)}
                                className="col-span-3"
                                placeholder="Optional description..."
                                disabled={isSubmitting}
                            />
                        </div>
                    </div>
                    <DialogFooter>
                        <Button type="button" variant="outline" onClick={() => onOpenChange(false)} disabled={isSubmitting}>
                            Cancel
                        </Button>
                        <Button type="submit" disabled={isSubmitting}>
                            {isSubmitting && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                            Create Stack
                        </Button>
                    </DialogFooter>
                </form>
            </DialogContent>
        </Dialog>
    );
}
